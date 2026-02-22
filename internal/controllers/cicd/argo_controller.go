package cicd

import (
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/gitutil"
	"fmt"
	"regexp"
	"strings"
	"time"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	WorkDir              = "/workspace"
	GitCloneTemplateName = "git-clone"
	PVCSize              = "1Gi"
	GitRepoPath          = "./"
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

	// 2.1. 自动插入 git-clone 作为第一步
	branch := "master"
	if pipelineInfo.Branch != nil && *pipelineInfo.Branch != "" {
		branch = *pipelineInfo.Branch
	}
	var gitToken string
	if pipelineInfo.GitToken != nil {
		gitToken = *pipelineInfo.GitToken
	}
	gitCloneTemplate := createGitCloneTemplate(pipelineInfo.GitURL, branch, gitToken)
	templates = append(templates, gitCloneTemplate)
	gitCloneTask := wfv1.DAGTask{
		Name:     GitCloneTemplateName,
		Template: gitCloneTemplate.Name,
	}
	tasks = append(tasks, gitCloneTask)

	// 2.2. 先建立 原始步骤名 → Argo 合法 taskName 的映射，用索引保证唯一
	nameMap := make(map[string]string, len(steps))
	for i, step := range steps {
		safe := sanitizeArgoName(step.Name)
		var argoTaskName string
		if safe == "" || safe == "step" {
			argoTaskName = fmt.Sprintf("s%d", i)
		} else {
			argoTaskName = fmt.Sprintf("s%d-%s", i, safe)
		}
		nameMap[step.Name] = argoTaskName
	}

	// 2.3. 用户定义步骤
	for _, step := range steps {
		argoTaskName := nameMap[step.Name]
		templateName := "tpl-" + argoTaskName
		template := createArgoWorkflowTemplateNamed(step, templateName)
		templates = append(templates, template)

		task := wfv1.DAGTask{
			Name:     argoTaskName,
			Template: templateName,
		}
		if step.DependsOn != nil && *step.DependsOn != "" {
			// 将 dependsOn 中每个原始名称查映射表，找到对应 argoTaskName
			var depNames []string
			for _, dep := range strings.Split(*step.DependsOn, ",") {
				dep = strings.TrimSpace(dep)
				if dep == "" {
					continue
				}
				if mappedName, ok := nameMap[dep]; ok {
					depNames = append(depNames, mappedName)
				}
			}
			if len(depNames) > 0 {
				task.Depends = GitCloneTemplateName + " && (" + strings.Join(depNames, " && ") + ")"
			} else {
				task.Depends = GitCloneTemplateName
			}
		} else {
			task.Depends = GitCloneTemplateName
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

	// 处理时间字段以避免零日期数据库错误
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
	commitId, err := gitutil.GetCommitID(GitRepoPath, *pipelineInfo.Branch)
	if err != nil {
		helper.InternalError("获取Git提交ID失败: " + err.Error())
		return
	}

	// 记录记录表
	pipelineRun := &model.PipelineRun{
		PipelineID:   pipelineId,
		WorkflowName: createWorkflow.Name,
		Status:       &status,
		Operator:     utils.GetUserNameFromContext(ctx),
		Branch:       pipelineInfo.Branch,
		CommitID:     &commitId,
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
func createGitCloneTemplate(gitURL, branch, gitToken string) wfv1.Template {
	// 如果有 token，将其注入到 HTTPS URL 中，格式: https://token@github.com/...
	cloneURL := gitURL
	if gitToken != "" && strings.HasPrefix(gitURL, "https://") {
		cloneURL = "https://" + gitToken + "@" + strings.TrimPrefix(gitURL, "https://")
	}
	cloneCmd := fmt.Sprintf("git -c http.version=HTTP/1.1 clone --depth 1 --branch %s '%s' %s && echo 'clone done'", branch, cloneURL, WorkDir)
	return wfv1.Template{
		Name: GitCloneTemplateName,
		Container: &corev1.Container{
			Image:   "alpine/git:latest",
			Command: []string{"sh", "-c"},
			Args:    []string{cloneCmd},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "workdir",
					MountPath: WorkDir,
				},
			},
		},
	}
}

func createArgoWorkflowTemplateNamed(step *model.PipelineStep, templateName string) wfv1.Template {
	image := strings.TrimSpace(step.Image)

	// 共用 workspace 挂载
	workdirMount := corev1.VolumeMount{
		Name:      "workdir",
		MountPath: WorkDir,
	}

	// Kaniko 镜像构建（默认没有 shell，直接传参数
	// 用户在命令框中每行写一个参数，如 --destination=xxx
	if strings.Contains(image, "kaniko-project/executor") || strings.Contains(image, "kaniko/executor") {
		var args []string
		for _, line := range strings.Split(step.Commands, "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				args = append(args, line)
			}
		}
		return wfv1.Template{
			Name: templateName,
			Container: &corev1.Container{
				Image:   image,
				Command: []string{"/kaniko/executor"},
				Args:    args,
				VolumeMounts: []corev1.VolumeMount{
					workdirMount,
					{
						Name:      "docker-config",
						MountPath: "/kaniko/.docker",
					},
				},
			},
		}
	}

	// ── 默认：sh -c 模式
	return wfv1.Template{
		Name: templateName,
		Container: &corev1.Container{
			Image:        image,
			Command:      []string{"sh", "-c"},
			Args:         []string{step.Commands},
			WorkingDir:   WorkDir,
			VolumeMounts: []corev1.VolumeMount{workdirMount},
		},
	}
}

func createArgoWorkflow(pipeline *model.Pipeline, tasks []wfv1.DAGTask, templates []wfv1.Template) *wfv1.Workflow {
	mainTemplate := wfv1.Template{
		Name: "main",
		DAG: &wfv1.DAGTemplate{
			Tasks: tasks,
		},
	}
	allTemplates := append([]wfv1.Template{mainTemplate}, templates...)

	// Argo 为每次 Workflow 运行创建一个独立 PVC，所有 Pod 共享，git-clone 的代码才能传递给后续步骤
	storageRequest, _ := resource.ParseQuantity(PVCSize)

	return &wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			// GenerateName 也要净化，避免中文 pipeline 名称导致 Workflow 创建失败
			GenerateName: sanitizeArgoName(pipeline.Name) + "-",
			Labels:       map[string]string{"app": sanitizeArgoName(pipeline.Name)},
		},
		Spec: wfv1.WorkflowSpec{
			Entrypoint: "main",
			Templates:  allTemplates,
			// ServiceAccountName: argo 服务账号通常拥有足够权限操作集群资源
			ServiceAccountName: "argo",
			// 全局额外卷：docker-config 用于 Kaniko 镜像推送
			Volumes: []corev1.Volume{
				{
					Name: "docker-config",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "registry-creds",
							Items: []corev1.KeyToPath{
								{
									Key:  ".dockerconfigjson",
									Path: "config.json",
								},
							},
						},
					},
				},
			},
			// 每次 Workflow 自动创建 1Gi PVC，跨 Pod 共享
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "workdir",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: storageRequest,
							},
						},
					},
				},
			},
		},
	}
}

var argoInvalidRe = regexp.MustCompile(`[^a-z0-9-]`)
var argoMultiHyphenRe = regexp.MustCompile(`-{2,}`)

func sanitizeArgoName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	// 非法字符替换为连字符
	name = argoInvalidRe.ReplaceAllString(name, "-")
	// 折叠多个连续连字符
	name = argoMultiHyphenRe.ReplaceAllString(name, "-")
	// 去除首尾连字符
	name = strings.Trim(name, "-")
	if name == "" {
		name = "step"
	}
	return name
}
