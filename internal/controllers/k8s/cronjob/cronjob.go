package cronjob

import (
	"context"
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/logs"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// CronJobController CronJob控制器
type CronJobController struct{}

// NewCronJobController 创建CronJob控制器实例
func NewCronJobController() *CronJobController {
	return &CronJobController{}
}

// convertV1ToV1Beta1CronJob 将 batch/v1 CronJob 转换为 batch/v1beta1 CronJob
func (c *CronJobController) convertV1ToV1Beta1CronJob(v1CronJob *batchv1.CronJob) *batchv1beta1.CronJob {
	if v1CronJob == nil {
		return nil
	}

	return &batchv1beta1.CronJob{
		ObjectMeta: v1CronJob.ObjectMeta,
		Spec: batchv1beta1.CronJobSpec{
			Schedule:                v1CronJob.Spec.Schedule,
			StartingDeadlineSeconds: v1CronJob.Spec.StartingDeadlineSeconds,
			ConcurrencyPolicy:       batchv1beta1.ConcurrencyPolicy(v1CronJob.Spec.ConcurrencyPolicy),
			Suspend:                 v1CronJob.Spec.Suspend,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				ObjectMeta: v1CronJob.Spec.JobTemplate.ObjectMeta,
				Spec: batchv1.JobSpec{
					Parallelism:             v1CronJob.Spec.JobTemplate.Spec.Parallelism,
					Completions:             v1CronJob.Spec.JobTemplate.Spec.Completions,
					ActiveDeadlineSeconds:   v1CronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds,
					BackoffLimit:            v1CronJob.Spec.JobTemplate.Spec.BackoffLimit,
					TTLSecondsAfterFinished: v1CronJob.Spec.JobTemplate.Spec.TTLSecondsAfterFinished,
					Template:                v1CronJob.Spec.JobTemplate.Spec.Template,
				},
			},
			SuccessfulJobsHistoryLimit: v1CronJob.Spec.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     v1CronJob.Spec.FailedJobsHistoryLimit,
		},
		Status: batchv1beta1.CronJobStatus{
			Active:           v1CronJob.Status.Active,
			LastScheduleTime: v1CronJob.Status.LastScheduleTime,
			LastSuccessfulTime: func() *metav1.Time {
				if v1CronJob.Status.LastSuccessfulTime != nil {
					t := *v1CronJob.Status.LastSuccessfulTime
					return &t
				}
				return nil
			}(),
		},
	}
}

// convertV1Beta1ToV1CronJob 将 batch/v1beta1 CronJob 转换为 batch/v1 CronJob
func (c *CronJobController) convertV1Beta1ToV1CronJob(v1beta1CronJob *batchv1beta1.CronJob) *batchv1.CronJob {
	if v1beta1CronJob == nil {
		return nil
	}

	return &batchv1.CronJob{
		ObjectMeta: v1beta1CronJob.ObjectMeta,
		Spec: batchv1.CronJobSpec{
			Schedule:                v1beta1CronJob.Spec.Schedule,
			StartingDeadlineSeconds: v1beta1CronJob.Spec.StartingDeadlineSeconds,
			ConcurrencyPolicy:       batchv1.ConcurrencyPolicy(v1beta1CronJob.Spec.ConcurrencyPolicy),
			Suspend:                 v1beta1CronJob.Spec.Suspend,
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: v1beta1CronJob.Spec.JobTemplate.ObjectMeta,
				Spec: batchv1.JobSpec{
					Parallelism:             v1beta1CronJob.Spec.JobTemplate.Spec.Parallelism,
					Completions:             v1beta1CronJob.Spec.JobTemplate.Spec.Completions,
					ActiveDeadlineSeconds:   v1beta1CronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds,
					BackoffLimit:            v1beta1CronJob.Spec.JobTemplate.Spec.BackoffLimit,
					TTLSecondsAfterFinished: v1beta1CronJob.Spec.JobTemplate.Spec.TTLSecondsAfterFinished,
					Template:                v1beta1CronJob.Spec.JobTemplate.Spec.Template,
				},
			},
			SuccessfulJobsHistoryLimit: v1beta1CronJob.Spec.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     v1beta1CronJob.Spec.FailedJobsHistoryLimit,
		},
		Status: batchv1.CronJobStatus{
			Active:           v1beta1CronJob.Status.Active,
			LastScheduleTime: v1beta1CronJob.Status.LastScheduleTime,
			LastSuccessfulTime: func() *metav1.Time {
				if v1beta1CronJob.Status.LastSuccessfulTime != nil {
					t := *v1beta1CronJob.Status.LastSuccessfulTime
					return &t
				}
				return nil
			}(),
		},
	}
}

// CreateCronJob 创建CronJob
func (c *CronJobController) CreateCronJob(ctx *gin.Context) {
	logData := map[string]interface{}{}
	logs.Debug(logData, "创建CronJob")

	var req k8s.CronJobCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logs.Error(logData, "请求参数绑定失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	logData = map[string]interface{}{
		"name":      req.Name,
		"namespace": req.Namespace,
		"schedule":  req.Schedule,
	}

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	cronJob := c.convertCreateRequestToK8sCronJob(&req)
	var err error

	// 首先尝试使用 batch/v1 API (Kubernetes 1.21+)
	if groups, err := client.Discovery().ServerGroups(); err == nil {
		for _, group := range groups.Groups {
			if group.Name == "batch" {
				for _, version := range group.Versions {
					if version.Version == "v1" {
						// 转换为 batch/v1 格式
						v1CronJob := c.convertV1Beta1ToV1CronJob(cronJob)
						_, err = client.BatchV1().CronJobs(req.Namespace).Create(context.TODO(), v1CronJob, metav1.CreateOptions{})
						break
					}
				}
				break
			}
		}
	}

	// 如果 batch/v1 不可用或失败，尝试使用 batch/v1beta1
	if err != nil {
		_, err = client.BatchV1beta1().CronJobs(req.Namespace).Create(context.TODO(), cronJob, metav1.CreateOptions{})
	}

	if err != nil {
		logs.Error(logData, "创建CronJob失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建CronJob失败: " + err.Error())
		return
	}

	logs.Info(logData, "创建CronJob成功")
	helper := utils.NewResponseHelper(ctx)
	helper.Success("CronJob创建成功")
}

// DeleteCronJob 删除CronJob
func (c *CronJobController) DeleteCronJob(ctx *gin.Context) {
	logData := map[string]interface{}{}
	logs.Debug(logData, "删除CronJob")

	var req k8s.CronJobDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logs.Error(logData, "请求参数绑定失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	logData = map[string]interface{}{
		"name":      req.Name,
		"namespace": req.Namespace,
	}

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	var err error

	// 首先尝试使用 batch/v1 API (Kubernetes 1.21+)
	if groups, err := client.Discovery().ServerGroups(); err == nil {
		for _, group := range groups.Groups {
			if group.Name == "batch" {
				for _, version := range group.Versions {
					if version.Version == "v1" {
						err = client.BatchV1().CronJobs(req.Namespace).Delete(context.TODO(), req.Name, metav1.DeleteOptions{})
						break
					}
				}
				break
			}
		}
	}

	// 如果 batch/v1 不可用或失败，尝试使用 batch/v1beta1
	if err != nil {
		err = client.BatchV1beta1().CronJobs(req.Namespace).Delete(context.TODO(), req.Name, metav1.DeleteOptions{})
	}
	if err != nil {
		logs.Error(logData, "删除CronJob失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除CronJob失败: " + err.Error())
		return
	}

	logs.Info(logData, "删除CronJob成功")
	helper := utils.NewResponseHelper(ctx)
	helper.Success("CronJob删除成功")
}

// GetCronJobList 获取CronJob列表
func (c *CronJobController) GetCronJobList(ctx *gin.Context) {
	namespace := ctx.Param("namespace")

	// 如果 namespace 为 "all"，则使用空字符串获取所有命名空间的资源
	if namespace == "all" {
		namespace = ""
	} else if namespace == "" {
		namespace = "default"
	}

	logData := map[string]interface{}{"namespace": namespace}
	logs.Debug(logData, "获取CronJob列表")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	var cronJobList *batchv1beta1.CronJobList
	var err error

	// 首先尝试使用 batch/v1 API (Kubernetes 1.21+)
	if groups, err := client.Discovery().ServerGroups(); err == nil {
		// 检查是否支持 batch/v1
		for _, group := range groups.Groups {
			if group.Name == "batch" {
				for _, version := range group.Versions {
					if version.Version == "v1" {
						// 使用 batch/v1 API
						v1CronJobs, err1 := client.BatchV1().CronJobs(namespace).List(context.TODO(), metav1.ListOptions{})
						if err1 == nil {
							// 转换为 v1beta1 格式以保持兼容性
							cronJobList = &batchv1beta1.CronJobList{
								ListMeta: v1CronJobs.ListMeta,
								Items:    make([]batchv1beta1.CronJob, len(v1CronJobs.Items)),
							}
							for i, item := range v1CronJobs.Items {
								converted := c.convertV1ToV1Beta1CronJob(&item)
								if converted != nil {
									cronJobList.Items[i] = *converted
								}
							}
							err = nil
							break
						}
					}
				}
				break
			}
		}
	}

	// 如果 batch/v1 不可用或失败，尝试使用 batch/v1beta1
	if cronJobList == nil {
		cronJobList, err = client.BatchV1beta1().CronJobs(namespace).List(context.TODO(), metav1.ListOptions{})
	}

	if err != nil {
		logs.Error(logData, "获取CronJob列表失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取CronJob列表失败: " + err.Error())
		return
	}

	resp := make([]k8s.CronJobListItem, 0, len(cronJobList.Items))
	for _, cj := range cronJobList.Items {
		status := ""
		if strings.Contains(cj.Status.String(), "Active") {
			status = "active"
		}

		containerName := ""
		image := ""
		command := []string{}
		if len(cj.Spec.JobTemplate.Spec.Template.Spec.Containers) > 0 {
			container := cj.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
			containerName = container.Name
			image = container.Image
			command = container.Command
		}

		resp = append(resp, k8s.CronJobListItem{
			Name:          cj.Name,
			Namespace:     cj.Namespace,
			ContainerName: containerName,
			Image:         image,
			Command:       command,
			Schedule:      cj.Spec.Schedule,
			Status:        status,
			Age:           cj.CreationTimestamp.Unix(),
		})
	}

	logs.Info(map[string]interface{}{"count": len(resp), "data": logData}, "获取CronJob列表成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("查询成功", "cronJobList", resp)
}

// UpdateCronJob 更新CronJob
func (c *CronJobController) UpdateCronJob(ctx *gin.Context) {
	logData := map[string]interface{}{}
	logs.Debug(logData, "更新CronJob")

	var req k8s.CronJobUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logs.Error(logData, "请求参数绑定失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	logData = map[string]interface{}{
		"name":      req.Name,
		"namespace": req.Namespace,
	}

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	// 获取现有CronJob
	var oldCJ *batchv1beta1.CronJob
	var err error

	// 首先尝试使用 batch/v1 API (Kubernetes 1.21+)
	if groups, err := client.Discovery().ServerGroups(); err == nil {
		for _, group := range groups.Groups {
			if group.Name == "batch" {
				for _, version := range group.Versions {
					if version.Version == "v1" {
						v1CronJob, err1 := client.BatchV1().CronJobs(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
						if err1 == nil {
							// 转换为 v1beta1 格式以保持兼容性
							oldCJ = c.convertV1ToV1Beta1CronJob(v1CronJob)
							err = nil
							break
						}
					}
				}
				break
			}
		}
	}

	// 如果 batch/v1 不可用或失败，尝试使用 batch/v1beta1
	if oldCJ == nil {
		oldCJ, err = client.BatchV1beta1().CronJobs(req.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
	}
	if err != nil {
		logs.Error(logData, "获取CronJob失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("CronJob不存在")
		return
	}

	// 构造patch数据
	patch := make(map[string]interface{})
	updateFields := make(map[string]interface{})

	if req.Schedule != nil {
		patch["spec"] = make(map[string]interface{})
		patch["spec"].(map[string]interface{})["schedule"] = *req.Schedule
		updateFields["schedule"] = *req.Schedule
	}

	if req.Image != nil {
		containerName := ""
		if len(oldCJ.Spec.JobTemplate.Spec.Template.Spec.Containers) > 0 {
			containerName = oldCJ.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Name
		}

		jobTpl := map[string]interface{}{
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"restartPolicy": "OnFailure",
						"containers": []map[string]interface{}{{
							"name":  containerName,
							"image": *req.Image,
						}},
					},
				},
			},
		}

		if _, ok := patch["spec"]; !ok {
			patch["spec"] = make(map[string]interface{})
		}
		patch["spec"].(map[string]interface{})["jobTemplate"] = jobTpl
		updateFields["image"] = *req.Image
	}

	// 发送patch请求
	payload, _ := json.Marshal(patch)

	// 首先尝试使用 batch/v1 API (Kubernetes 1.21+)
	if groups, err := client.Discovery().ServerGroups(); err == nil {
		for _, group := range groups.Groups {
			if group.Name == "batch" {
				for _, version := range group.Versions {
					if version.Version == "v1" {
						_, err = client.BatchV1().CronJobs(req.Namespace).Patch(context.TODO(), req.Name, types.MergePatchType, payload, metav1.PatchOptions{})
						break
					}
				}
				break
			}
		}
	}

	// 如果 batch/v1 不可用或失败，尝试使用 batch/v1beta1
	if err != nil {
		_, err = client.BatchV1beta1().CronJobs(req.Namespace).Patch(context.TODO(), req.Name, types.MergePatchType, payload, metav1.PatchOptions{})
	}
	if err != nil {
		logs.Error(map[string]interface{}{"updateFields": updateFields, "error": err.Error(), "data": logData}, "更新CronJob失败")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("更新CronJob失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{"updateFields": updateFields, "data": logData}, "更新CronJob成功")
	helper := utils.NewResponseHelper(ctx)
	helper.Success("CronJob更新成功")
}

// convertCreateRequestToK8sCronJob 转换创建请求为K8s CronJob
func (c *CronJobController) convertCreateRequestToK8sCronJob(req *k8s.CronJobCreateRequest) *batchv1beta1.CronJob {
	return &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule: req.Schedule,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyOnFailure,
							Containers: []corev1.Container{{
								Name:    req.ContainerName,
								Image:   req.Image,
								Command: req.Command,
							}},
						},
					},
				},
			},
		},
	}
}
