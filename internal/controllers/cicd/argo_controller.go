package cicd

import (
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"fmt"

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
			Name:     step.Name,
			Template: template.Name,
		}
		if step.DependsOn != nil {
			task.Depends = *step.DependsOn
		}
		tasks = append(tasks, task)
	}
	// 3. 创建 Argo Workflow
	wf := createArgoWorkflow(pipelineInfo, tasks)
	// 4. 提交到k8s中
	restConfig, exist := configs.GetK8sConfig(uint(pipelineInfo.K8sInstanceID))
	if !exist {
		helper.InternalError("获取k8s客户端失败")
	}

	argoClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		helper.InternalError("创建 Argo 客户端失败")
	}
	createWorkflow, err := argoClient.ArgoprojV1alpha1().Workflows("argo").Create(ctx, wf, metav1.CreateOptions{})
	if err != nil {
		helper.InternalError("创建 Argo Workflow 失败")
	}
	status := string(createWorkflow.Status.Phase)
	if status == "" {
		status = "UNKNOWN"
	}
	startTime := createWorkflow.Status.StartedAt.Time
	endTime := createWorkflow.Status.FinishedAt.Time
	var duration uint32
	if !startTime.IsZero() && !endTime.IsZero() {
		duration = uint32(endTime.Sub(startTime).Seconds())
	}
	// 记录记录表
	pipelineRun := &model.PipelineRun{
		PipelineID:   pipelineId,
		WorkflowName: createWorkflow.Name,
		Status:       &status,
		Operator:     utils.GetUserNameFromContext(ctx),
		Branch:       pipelineInfo.Branch,
		CommitID:     nil,
		StartTime:    &startTime,
		EndTime:      &endTime,
		Duration:     &duration,
	}
	err = c.pipelineRunMapper.CreatePipelineRun(pipelineRun)
	if err != nil {
		helper.DatabaseError("创建流水线运行记录失败")
	}
	helper.SuccessWithData("success", "data", pipelineRun)
}

func createArgoWorkflowTemplate(step *model.PipelineStep) wfv1.Template {
	templateName := fmt.Sprintf("%v:%v", "templ-", step.Name)
	return wfv1.Template{
		Name: templateName,
		Container: &corev1.Container{
			Image:   step.Image,
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

func createArgoWorkflow(pipeline *model.Pipeline, tasks []wfv1.DAGTask) *wfv1.Workflow {
	return &wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%v-", pipeline.Name),
			Labels:       map[string]string{"app": pipeline.Name},
		},
		Spec: wfv1.WorkflowSpec{
			Templates: []wfv1.Template{
				{
					Name: "main",
					DAG: &wfv1.DAGTemplate{
						Tasks: tasks,
					},
				},
			},
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
