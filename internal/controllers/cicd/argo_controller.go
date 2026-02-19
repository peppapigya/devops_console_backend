package cicd

import (
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"fmt"
	"strings"
	"time"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	WorkDir = "/workspace"
)

type ArgoController struct {
	pipelineRunMapper  *mapper.PipelineRunMapper
	pipelineMapper     *mapper.PipelinesMapper
	pipelineStepMapper *mapper.PipelineStepsMapper
}

func NewArgoController(pipelineRunMapper *mapper.PipelineRunMapper, pipelineMapper *mapper.PipelinesMapper, pipelineStepMapper *mapper.PipelineStepsMapper) *ArgoController {
	return &ArgoController{
		pipelineRunMapper:  pipelineRunMapper,
		pipelineMapper:     pipelineMapper,
		pipelineStepMapper: pipelineStepMapper,
	}
}

func (c *ArgoController) ExecutePipeline(ctx *gin.Context) {
	var pipelineId uint32
	utils.GetParam(ctx, "pipelineId", &pipelineId, nil)
	helper := utils.NewResponseHelper(ctx)
	// 1. 通过 pipelineId 获取对应的步骤
	pipelineInfo, err := c.pipelineMapper.GetPipelineById(pipelineId)
	if err != nil {
		helper.DatabaseError("获取流水线信息失败")
		return
	}
	if pipelineInfo == nil {
		helper.NotFound("流水线不存在")
		return
	}
	steps, err := c.pipelineStepMapper.GetPipelineStepByPipelineId(pipelineId)
	// 2. 组装 Argo Workflow 模版
	var tasks []wfv1.DAGTask
	var templates []wfv1.Template
	for _, step := range steps {
		template := createArgoWorkflowTemplate(step)
		templates = append(templates, template)
		// 生成任务
		task := wfv1.DAGTask{
			Name:     strings.TrimSpace(step.Name),
			Template: template.Name,
		}
		if step.DependsOn != nil {
			task.Depends = *step.DependsOn
		}
		tasks = append(tasks, task)
	}
	// 3. 创建 Argo Workflow
	wf := createArgoWorkflow(pipelineInfo, tasks, templates)
	// 4. 提交到k8s中
	restConfig, exist := configs.GetK8sConfig(uint(pipelineInfo.K8sInstanceID))
	if !exist {
		helper.InternalError("获取k8s客户端失败")
		return
	}

	argoClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		helper.InternalError("创建 Argo 客户端失败")
		return
	}
	createWorkflow, err := argoClient.ArgoprojV1alpha1().Workflows("argo").Create(ctx, wf, metav1.CreateOptions{})
	if err != nil {
		helper.InternalError("创建 Argo Workflow 失败: " + err.Error())
		return
	}
	status := string(createWorkflow.Status.Phase)
	if status == "" {
		status = "Pending"
	}

	// Handle time fields to avoid zero date DB error
	var startTime *time.Time
	var endTime *time.Time
	var duration *uint32

	if !createWorkflow.Status.StartedAt.Time.IsZero() {
		t := createWorkflow.Status.StartedAt.Time
		startTime = &t
	}
	if !createWorkflow.Status.FinishedAt.Time.IsZero() {
		t := createWorkflow.Status.FinishedAt.Time
		endTime = &t
	}

	if startTime != nil && endTime != nil {
		d := uint32(endTime.Sub(*startTime).Seconds())
		duration = &d
	}

	// 记录记录表
	pipelineRun := &model.PipelineRun{
		PipelineID:   pipelineId,
		WorkflowName: createWorkflow.Name,
		Status:       &status,
		Operator:     utils.GetUserNameFromContext(ctx),
		Branch:       pipelineInfo.Branch,
		CommitID:     nil, // TODO: Get commit ID from git
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     duration,
	}
	err = c.pipelineRunMapper.CreatePipelineRun(pipelineRun)
	if err != nil {
		helper.DatabaseError("创建流水线运行记录失败: " + err.Error())
		return
	}
	helper.SuccessWithData("success", "data", pipelineRun)
}

func createArgoWorkflowTemplate(step *model.PipelineStep) wfv1.Template {
	templateName := fmt.Sprintf("%v%v", "temple-", strings.TrimSpace(step.Name))
	return wfv1.Template{
		Name: templateName,
		Container: &corev1.Container{
			Image:   strings.TrimSpace(step.Image),
			Command: []string{"sh", "-c"},
			Args:    []string{step.Commands},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "workdir",
					MountPath: WorkDir,
				},
			},
			WorkingDir: WorkDir,
		},
	}
}

func createArgoWorkflow(pipeline *model.Pipeline, tasks []wfv1.DAGTask, templates []wfv1.Template) *wfv1.Workflow {
	// 将后续步骤加入到主模版
	mainTemplate := wfv1.Template{
		Name: "main",
		DAG: &wfv1.DAGTemplate{
			Tasks: tasks,
		},
	}
	allTemplates := append([]wfv1.Template{mainTemplate}, templates...)

	return &wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%v-", pipeline.Name),
			Labels:       map[string]string{"app": pipeline.Name},
		},
		Spec: wfv1.WorkflowSpec{
			Entrypoint: "main",
			Templates:  allTemplates,
			Volumes: []corev1.Volume{
				{
					Name: "workdir",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
}
