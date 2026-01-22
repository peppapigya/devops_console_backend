package job

import (
	"devops-console-backend/config"
	"devops-console-backend/models/k8s"
	"devops-console-backend/utils"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// JobController Job控制器
type JobController struct{}

// NewJobController 创建Job控制器实例
func NewJobController() *JobController {
	return &JobController{}
}

// GetJobDetail 获取Job详情
func (c *JobController) GetJobDetail(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	jobName := ctx.Param("jobName")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	jobDetail, err := client.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("Job不存在")
		return
	}

	// 处理容器信息（取第一个容器）
	containerName := ""
	containerImage := ""
	commandArgs := ""
	if len(jobDetail.Spec.Template.Spec.Containers) > 0 {
		container := jobDetail.Spec.Template.Spec.Containers[0]
		containerName = container.Name
		containerImage = container.Image

		// 拼接命令和参数为字符串
		var cmdArgs []string
		cmdArgs = append(cmdArgs, container.Command...) // 命令
		cmdArgs = append(cmdArgs, container.Args...)    // 参数
		commandArgs = strings.Join(cmdArgs, " ")        // 用空格分隔
	}

	// 处理标签（转换为JSON字符串）
	labelsJson, _ := json.Marshal(jobDetail.Labels)
	labels := string(labelsJson)

	// 处理时间（转换为RFC3339格式字符串）
	startTime := ""
	endTime := ""
	if jobDetail.Status.StartTime != nil {
		startTime = jobDetail.Status.StartTime.Format(time.RFC3339)
	}
	if jobDetail.Status.CompletionTime != nil {
		endTime = jobDetail.Status.CompletionTime.Format(time.RFC3339)
	}

	// 处理关联Pod状态
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName), // 通过job-name标签关联Pod
	})
	podsStatuses := ""
	if err != nil {
		podsStatuses = "get pods error: " + err.Error()
	} else {
		var podStatusList []string
		for _, pod := range pods.Items {
			podStatusList = append(podStatusList,
				fmt.Sprintf("pod[%s]:%s", pod.Name, pod.Status.Phase))
		}
		podsStatuses = strings.Join(podStatusList, "; ")
	}

	jobDetailResponse := k8s.JobDetail{
		JobName:        jobDetail.Name,
		NameSpace:      jobDetail.Namespace,
		ContainerName:  containerName,
		ContainerImage: containerImage,
		CommandArgs:    commandArgs,
		Labels:         labels,
		StartTime:      startTime,
		EndTime:        endTime,
		PodsStatuses:   podsStatuses,
		Age:            jobDetail.CreationTimestamp.Unix(),
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "jobDetail", jobDetailResponse)
}

// GetJobList 获取Job列表
func (c *JobController) GetJobList(ctx *gin.Context) {
	namespace := ctx.Param("namespace")

	// 如果 namespace 为 "all"，则使用空字符串获取所有命名空间的资源
	if namespace == "all" {
		namespace = ""
	}

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	// 核心逻辑：空namespace时获取所有命名空间的Job，否则获取指定命名空间的Job
	jobList, err := client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Job列表失败")
		return
	}

	var rspList []k8s.JobListItem
	for _, job := range jobList.Items {
		// 提取容器信息（默认取第一个容器）
		var containerName, containerImage string
		var commandArgs []string
		if len(job.Spec.Template.Spec.Containers) > 0 {
			container := job.Spec.Template.Spec.Containers[0]
			containerName = container.Name
			containerImage = container.Image
			// 合并命令和参数（更完整的命令展示）
			commandArgs = append(commandArgs, container.Command...)
			commandArgs = append(commandArgs, container.Args...)
		}

		// 格式化Pod状态
		podStatus := fmt.Sprintf(
			"活跃: %d, 成功: %d, 失败: %d",
			job.Status.Active,
			job.Status.Succeeded,
			job.Status.Failed,
		)

		// 转换标签为字符串（保留原始格式）
		labels := fmt.Sprintf("%v", job.Labels)

		// 处理时间格式（空值保护）
		startTime := ""
		if job.Status.StartTime != nil {
			startTime = job.Status.StartTime.String()
		}
		endTime := ""
		if job.Status.CompletionTime != nil {
			endTime = job.Status.CompletionTime.String()
		}

		// 构建响应对象
		rsp := k8s.JobListItem{
			JobName:        job.Name,
			NameSpace:      job.Namespace,
			ContainerName:  containerName,
			ContainerImage: containerImage,
			CommandArgs:    strings.Join(commandArgs, " "),
			Labels:         labels,
			StartTime:      startTime,
			EndTime:        endTime,
			PodsStatuses:   podStatus,
		}

		rspList = append(rspList, rsp)
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "jobList", rspList)
}

// CreateJob 创建Job
func (c *JobController) CreateJob(ctx *gin.Context) {
	var req k8s.JobCreateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	k8sJob := c.convertCreateRequestToK8sJob(&req)
	if k8sJob == nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("命令不能为空")
		return
	}

	_, err := client.BatchV1().Jobs(k8sJob.Namespace).Create(ctx, k8sJob, metav1.CreateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建Job失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Job创建成功")
}

// DeleteJob 删除Job
func (c *JobController) DeleteJob(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	jobName := ctx.Param("jobName")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	err := client.BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除Job失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Job删除成功")
}

// convertCreateRequestToK8sJob 转换创建请求为K8s Job
func (c *JobController) convertCreateRequestToK8sJob(req *k8s.JobCreateRequest) *batchv1.Job {
	// 将命令字符串按空格拆分为切片（处理任意数量空格）
	commandSlice := strings.Fields(req.Command)
	if len(commandSlice) == 0 {
		return nil
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.JobName,
			Namespace: req.NameSpace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: pointer.Int32Ptr(3), // 失败重试3次（默认值）
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    req.ContainerName,
							Image:   req.ContainerImage,
							Command: commandSlice, // 使用拆分后的命令切片
						},
					},
				},
			},
		},
	}
}
