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
		helper.BadRequest("无效的 ID或记录不存在")
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

	// 1. 先查 run 记录，获取 workflowName 和 pipelineId
	run, err := c.mapper.GetPipelineRunById(id)
	if err != nil || run == nil {
		helper.NotFound("运行记录不存在")
		return
	}

	// 2. 查 pipeline 获取 k8sInstanceId，用于构造 Argo 客户端
	if run.WorkflowName != "" {
		pipeline, pErr := c.pipelinesMapper.GetPipelineById(run.PipelineID)
		if pErr == nil && pipeline != nil {
			restConfig, exist := configs.GetK8sConfig(uint(pipeline.K8sInstanceID))
			if exist {
				argoClient, aErr := versioned.NewForConfig(restConfig)
				if aErr == nil {
					// 删除 Argo Workflow，如果不存在则忽略
					_ = argoClient.ArgoprojV1alpha1().Workflows("argo").Delete(
						ctx, run.WorkflowName, metav1.DeleteOptions{},
					)
				}
			}
		}
	}

	// 3. 删除 DB 记录
	err = c.mapper.DeletePipelineRun(id)
	if err != nil {
		helper.DatabaseError(err.Error())
		return
	}
	helper.Success("success")
}

func (c *PipelineRunController) GetPagePipelineRuns(ctx *gin.Context) {
	var pageNum int
	var pageSize int
	var pipelineId uint64
	helper := utils.NewResponseHelper(ctx)
	utils.GetParam(ctx, "pageNum", &pageNum, nil)
	utils.GetParam(ctx, "pageSize", &pageSize, nil)
	utils.GetParam(ctx, "pipelineId", &pipelineId, nil)
	pipelineRuns, total, err := c.mapper.GetPagePipelineRuns(pageNum, pageSize, pipelineId)
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

	// 4. 将前端传入的原始步骤名转换为 Argo task 名（DisplayName），用于节点匹配
	stepName = strings.TrimSpace(stepName)
	argoStepName := stepName // git-clone 直接使用原名
	if stepName != "git-clone" {
		// 从 DB 取 definedSteps，重建与 argo_controller 相同的 nameMap
		definedSteps, dErr := c.pipelinesMapper.GetPipelineStepsByPipelineId(run.PipelineID)
		if dErr == nil {
			for i, s := range definedSteps {
				safe := sanitizeArgoName(s.Name)
				var mapped string
				if safe == "" || safe == "step" {
					mapped = fmt.Sprintf("s%d", i)
				} else {
					mapped = fmt.Sprintf("s%d-%s", i, safe)
				}
				if s.Name == stepName {
					argoStepName = mapped
					break
				}
			}
		}
	}

	// 5. Find Node Name by matching Argo DisplayName
	var targetNodeName string
	for _, node := range wf.Status.Nodes {
		if node.Type == "Pod" && node.DisplayName == argoStepName {
			targetNodeName = node.Name
			break
		}
	}

	if targetNodeName == "" {
		helper.NotFound(fmt.Sprintf("未找到对应步骤的 Node. StepName: %s (argoName: %s)", stepName, argoStepName))
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

	// 3. 构建 nodeStatusMap: key = Argo 节点的 DisplayName（即 DAGTask.Name）
	type nodeInfo struct {
		Status     string
		StartedAt  time.Time
		FinishedAt time.Time
		PodName    string
	}
	nodeStatusMap := make(map[string]nodeInfo)
	for _, node := range wf.Status.Nodes {
		if node.Type != "Pod" {
			continue
		}
		key := node.DisplayName
		if key == "" {
			key = node.TemplateName
		}
		nodeStatusMap[key] = nodeInfo{
			Status:     string(node.Phase),
			StartedAt:  node.StartedAt.Time,
			FinishedAt: node.FinishedAt.Time,
			PodName:    node.Name,
		}
	}

	// 4. 与 argo_controller.go 完全相同的 nameMap 逻辑，确保能正确反查 Argo task 名
	stepNameMap := make(map[string]string, len(definedSteps))
	for i, step := range definedSteps {
		safe := sanitizeArgoName(step.Name)
		var argoTaskName string
		if safe == "" || safe == "step" {
			argoTaskName = fmt.Sprintf("s%d", i)
		} else {
			argoTaskName = fmt.Sprintf("s%d-%s", i, safe)
		}
		stepNameMap[step.Name] = argoTaskName
	}

	type StepStatus struct {
		Name      string `json:"name"`
		Status    string `json:"status"`
		Duration  string `json:"duration"`
		StartedAt string `json:"startedAt"`
		PodName   string `json:"podName"`
		Image     string `json:"image"`
	}
	var steps []StepStatus

	// 5. 首先插入 git-clone 步骤（始终存在）
	{
		ni := nodeStatusMap["git-clone"]
		startedAtStr := ""
		durationStr := ""
		if !ni.StartedAt.IsZero() {
			startedAtStr = ni.StartedAt.Local().Format("2006-01-02 15:04:05")
			if !ni.FinishedAt.IsZero() {
				durationStr = ni.FinishedAt.Sub(ni.StartedAt).String()
			} else {
				durationStr = time.Since(ni.StartedAt).Round(time.Second).String()
			}
		}
		gitCloneStatus := ni.Status
		if gitCloneStatus == "" {
			gitCloneStatus = "Pending"
		}
		steps = append(steps, StepStatus{
			Name:      "git-clone",
			Status:    gitCloneStatus,
			Duration:  durationStr,
			StartedAt: startedAtStr,
			PodName:   ni.PodName,
			Image:     "alpine/git:latest",
		})
	}

	// 6. 用户定义步骤，通过 stepNameMap 查 Argo task 名
	for _, defStep := range definedSteps {
		argoTaskName := stepNameMap[defStep.Name]
		ni := nodeStatusMap[argoTaskName]

		status := ni.Status
		if status == "" {
			status = "Pending"
		}
		startedAtStr := ""
		durationStr := ""
		if !ni.StartedAt.IsZero() {
			startedAtStr = ni.StartedAt.Local().Format("2006-01-02 15:04:05")
			if !ni.FinishedAt.IsZero() {
				durationStr = ni.FinishedAt.Sub(ni.StartedAt).String()
			} else {
				durationStr = time.Since(ni.StartedAt).Round(time.Second).String()
			}
		}
		steps = append(steps, StepStatus{
			Name:      defStep.Name,
			Status:    status,
			Duration:  durationStr,
			StartedAt: startedAtStr,
			PodName:   ni.PodName,
			Image:     defStep.Image,
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
