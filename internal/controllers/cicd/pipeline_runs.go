package cicd

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"fmt"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PipelineRunController struct {
	mapper          *mapper.PipelineRunMapper
	pipelinesMapper *mapper.PipelinesMapper
}

func NewPipelineRunController(mapper *mapper.PipelineRunMapper, pipelinesMapper *mapper.PipelinesMapper) *PipelineRunController {
	return &PipelineRunController{
		mapper:          mapper,
		pipelinesMapper: pipelinesMapper,
	}
}

func (c *PipelineRunController) resolveRunId(ctx *gin.Context) (uint64, error) {
	idStr := ctx.Param("id")
	if strings.HasSuffix(idStr, "-latest") {
		pipelineIdStr := strings.TrimSuffix(idStr, "-latest")
		var pipelineId uint64
		_, err := fmt.Sscanf(pipelineIdStr, "%d", &pipelineId)
		if err != nil {
			return 0, err
		}
		var lastRun *model.PipelineRun
		lastRun, err = c.mapper.GetLastPipelineRunByPipelineId(pipelineId)
		if err != nil {
			return 0, err
		}
		return uint64(lastRun.ID), nil
	}
	var id uint64
	// Try parsing directly
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		// Fallback to GetParam if strict int parsing failed (though Sscanf is usually enough for uint64)
		// Or assume invalid
		return 0, err
	}
	return id, nil
}

func (c *PipelineRunController) GetPipelineRunById(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	id, err := c.resolveRunId(ctx)
	if err != nil {
		helper.BadRequest("无效的ID或记录不存在")
		return
	}

	// Sync status first
	c.syncPipelineRunStatus(ctx, id)

	pipelineRun, err := c.mapper.GetPipelineRunById(id)
	if err != nil {
		helper.DatabaseError(err.Error())
		return
	}
	helper.SuccessWithData("success", "pipelineRun", pipelineRun)
}

func (c *PipelineRunController) UpdatePipelineRun(ctx *gin.Context) {
	var pipelineRun model.PipelineRun
	helper := utils.NewResponseHelper(ctx)
	if !utils.BindAndValidate(ctx, &pipelineRun) {
		return
	}
	err := c.mapper.UpdatePipelineRun(&pipelineRun)
	if err != nil {
		helper.DatabaseError(err.Error())
	}
	helper.Success("success")
}

func (c *PipelineRunController) CreatePipelineRun(ctx *gin.Context) {
	var pipelineRun model.PipelineRun
	helper := utils.NewResponseHelper(ctx)
	if !utils.BindAndValidate(ctx, &pipelineRun) {
		return
	}
	err := c.mapper.CreatePipelineRun(&pipelineRun)
	if err != nil {
		helper.DatabaseError(err.Error())
	}
	helper.SuccessWithData("success", "data", pipelineRun)
}

func (c *PipelineRunController) DeletePipelineRun(ctx *gin.Context) {
	var id uint64
	helper := utils.NewResponseHelper(ctx)
	utils.GetParam(ctx, "id", &id, nil)
	err := c.mapper.DeletePipelineRun(id)
	if err != nil {
		helper.DatabaseError(err.Error())
	}
	helper.Success("success")
}

func (c *PipelineRunController) GetPagePipelineRuns(ctx *gin.Context) {
	var pageNum int
	var pageSize int
	helper := utils.NewResponseHelper(ctx)
	utils.GetParam(ctx, "pageNum", &pageNum, nil)
	utils.GetParam(ctx, "pageSize", &pageSize, nil)
	pipelineRuns, total, err := c.mapper.GetPagePipelineRuns(pageNum, pageSize)
	response := &common.PageInfoResponse[*model.PipelineRun]{
		Data:     pipelineRuns,
		PageNum:  pageNum,
		PageSize: pageSize,
		Total:    total,
	}
	if err != nil {
		helper.DatabaseError(err.Error())
	}
	helper.SuccessWithData("success", "data", response)
}

func (c *PipelineRunController) GetPipelineRunLogs(ctx *gin.Context) {
	stepName := ctx.Query("step_name")
	helper := utils.NewResponseHelper(ctx)

	id, err := c.resolveRunId(ctx)
	if err != nil {
		helper.BadRequest("无效的 ID或记录不存在")
		return
	}

	fmt.Printf("[Debug] GetLogs: runId=%v, stepName=%s\n", id, stepName)

	// 1. Get Run Logic
	run, err := c.mapper.GetPipelineRunById(id)
	if err != nil {
		helper.DatabaseError("获取运行记录失败")
		return
	}
	pipeline, err := c.pipelinesMapper.GetPipelineById(run.PipelineID)
	if err != nil {
		helper.DatabaseError("获取流水线信息失败")
		return
	}

	// 2. K8s Config
	restConfig, exist := configs.GetK8sConfig(uint(pipeline.K8sInstanceID))
	if !exist {
		helper.InternalError("获取K8s配置失败")
		return
	}

	// 3. Get Workflow
	argoClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		helper.InternalError("创建Argo客户端失败")
		return
	}

	fmt.Printf("[Debug] Fetching workflow logs for: %s\n", run.WorkflowName)
	wf, err := argoClient.ArgoprojV1alpha1().Workflows("argo").Get(ctx, run.WorkflowName, metav1.GetOptions{})
	if err != nil {
		helper.InternalError("获取Workflow失败: " + err.Error())
		return
	}

	// 4. Find Node Name first
	var targetNodeName string
	stepName = strings.TrimSpace(stepName)

	// We need to find the node in the workflow status that corresponds to our step
	for _, node := range wf.Status.Nodes {
		if node.Type == "Pod" {
			// Match by DisplayName (e.g. "build") or TemplateName (e.g. "temple-build")
			// Argo often names templates like "temple-pipeline-name..." but simpler checks first
			if node.DisplayName == stepName || node.TemplateName == stepName || node.TemplateName == "temple-"+stepName {
				targetNodeName = node.Name
				break
			}
		}
	}

	if targetNodeName == "" {
		helper.NotFound(fmt.Sprintf("未找到对应步骤的 Node. StepName: %s", stepName))
		return
	}

	// 5. Find Pod by listing pods with label selector
	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		helper.InternalError("创建K8s客户端失败")
		return
	}

	// List pods belonging to this workflow
	labelSelector := fmt.Sprintf("workflows.argoproj.io/workflow=%s", run.WorkflowName)
	pods, err := k8sClient.CoreV1().Pods("argo").List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		helper.InternalError("查询Pod失败: " + err.Error())
		return
	}

	var podName string
	var foundPodNames []string
	for _, pod := range pods.Items {
		foundPodNames = append(foundPodNames, fmt.Sprintf("%s(Node:%s)", pod.Name, pod.Annotations["workflows.argoproj.io/node-name"]))
		// Match by annotation which links Pod to Node
		if val, ok := pod.Annotations["workflows.argoproj.io/node-name"]; ok && val == targetNodeName {
			podName = pod.Name
			break
		}
	}

	if podName == "" {
		errorMsg := fmt.Sprintf("未找到对应步骤的 Pod. NodeName: %s. Found Pods: %v", targetNodeName, foundPodNames)
		fmt.Printf("[Debug] %s\n", errorMsg)
		helper.NotFound(errorMsg)
		return
	}
	fmt.Printf("[Debug] Resolved Pod Name via K8s List: %s\n", podName)

	// 6. Get Logs
	// Argo pods usually have "main" and "wait" containers. We need "main".
	req := k8sClient.CoreV1().Pods("argo").GetLogs(podName, &corev1.PodLogOptions{
		Container: "main",
	})
	logs, err := req.DoRaw(ctx)
	if err != nil {
		helper.InternalError("获取日志失败: " + err.Error())
		return
	}

	helper.SuccessWithData("success", "logs", string(logs))
}

func (c *PipelineRunController) GetPipelineRunSteps(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	id, err := c.resolveRunId(ctx)
	if err != nil {
		helper.BadRequest("无效的 ID或记录不存在")
		return
	}

	// Sync first
	c.syncPipelineRunStatus(ctx, id)

	run, err := c.mapper.GetPipelineRunById(id)
	if err != nil {
		helper.DatabaseError("获取运行记录失败")
		return
	}
	pipeline, err := c.pipelinesMapper.GetPipelineById(run.PipelineID)
	if err != nil {
		helper.DatabaseError("获取流水线信息失败")
		return
	}

	// 1. Fetch defined steps from DB to ensure correct order
	definedSteps, err := c.pipelinesMapper.GetPipelineStepsByPipelineId(run.PipelineID)
	if err != nil {
		helper.DatabaseError("获取流水线步骤定义失败")
		return
	}

	// 2. Fetch runtime status from Argo
	restConfig, exist := configs.GetK8sConfig(uint(pipeline.K8sInstanceID))
	if !exist {
		helper.InternalError("获取K8s配置失败")
		return
	}

	argoClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		helper.InternalError("创建Argo客户端失败")
		return
	}

	wf, err := argoClient.ArgoprojV1alpha1().Workflows("argo").Get(ctx, run.WorkflowName, metav1.GetOptions{})
	if err != nil {
		helper.InternalError("获取Workflow失败")
		return
	}

	// 3. Map Argo Node Status to Defined Steps
	type StepStatus struct {
		Name     string `json:"name"`
		Status   string `json:"status"`
		Duration string `json:"duration"`
		PodName  string `json:"podName"` // Add Pod Name for details
		Image    string `json:"image"`   // Add Image for details
	}
	var steps []StepStatus

	nodeStatusMap := make(map[string]struct {
		Status     string
		StartedAt  time.Time
		FinishedAt time.Time
		PodName    string
	})

	for _, node := range wf.Status.Nodes {
		if node.Type == "Pod" {
			// Key by DisplayName (step name) or clean TemplateName
			name := node.DisplayName
			if name == "" {
				name = strings.TrimPrefix(node.TemplateName, "temple-")
			}
			nodeStatusMap[name] = struct {
				Status     string
				StartedAt  time.Time
				FinishedAt time.Time
				PodName    string
			}{
				Status:     string(node.Phase),
				StartedAt:  node.StartedAt.Time,
				FinishedAt: node.FinishedAt.Time,
				PodName:    node.Name, // This is the actual Pod Name
			}
		}
	}

	// 4. Construct Result in Defined Order
	for _, defStep := range definedSteps {
		status := "Pending"
		duration := ""
		podName := ""

		if nodeInfo, ok := nodeStatusMap[defStep.Name]; ok {
			status = nodeInfo.Status
			if !nodeInfo.StartedAt.IsZero() && !nodeInfo.FinishedAt.IsZero() {
				d := nodeInfo.FinishedAt.Sub(nodeInfo.StartedAt)
				duration = d.String()
			} else if !nodeInfo.StartedAt.IsZero() {
				// Running duration
				d := time.Since(nodeInfo.StartedAt).Round(time.Second)
				duration = d.String()
			}
			podName = nodeInfo.PodName
		}

		steps = append(steps, StepStatus{
			Name:     defStep.Name,
			Status:   status,
			Duration: duration,
			PodName:  podName,
			Image:    defStep.Image,
		})
	}

	helper.SuccessWithData("success", "items", steps)
}

func (c *PipelineRunController) syncPipelineRunStatus(ctx *gin.Context, runId uint64) {
	fmt.Printf("[Debug] Syncing run status for id=%v\n", runId)
	run, err := c.mapper.GetPipelineRunById(runId)
	if err != nil {
		fmt.Printf("[Debug] Sync: GetPipelineRunById failed: %v\n", err)
		return
	}
	// If already validated final status, skip (optional optimization)
	if run.Status != nil && (*run.Status == "Succeeded" || *run.Status == "Failed" || *run.Status == "Error") {
		// return // Force update might be better if user wants to see latest logs/times
	}

	pipeline, err := c.pipelinesMapper.GetPipelineById(run.PipelineID)
	if err != nil || pipeline == nil {
		fmt.Printf("[Debug] Sync: GetPipelineById failed\n")
		return
	}

	restConfig, exist := configs.GetK8sConfig(uint(pipeline.K8sInstanceID))
	if !exist {
		fmt.Printf("[Debug] Sync: GetK8sConfig failed\n")
		return
	}

	argoClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		fmt.Printf("[Debug] Sync: NewForConfig failed\n")
		return
	}

	wf, err := argoClient.ArgoprojV1alpha1().Workflows("argo").Get(ctx, run.WorkflowName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("[Debug] Sync: Get Workflow failed: %v\n", err)
		return
	}

	// Update Fields
	status := string(wf.Status.Phase)
	run.Status = &status

	if !wf.Status.StartedAt.Time.IsZero() {
		t := wf.Status.StartedAt.Time
		run.StartTime = &t
	}
	if !wf.Status.FinishedAt.Time.IsZero() {
		t := wf.Status.FinishedAt.Time
		run.EndTime = &t
	}
	if run.StartTime != nil && run.EndTime != nil {
		d := uint32(run.EndTime.Sub(*run.StartTime).Seconds())
		run.Duration = &d
	}

	err = c.mapper.UpdatePipelineRun(run)
	if err != nil {
		fmt.Printf("[Debug] Sync: UpdatePipelineRun failed: %v\n", err)
	} else {
		fmt.Printf("[Debug] Sync: Status updated to %s\n", status)
	}
}
