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
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
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
	namespace := ctx.Param("namespace")

	logData := map[string]interface{}{}
	logs.Debug(logData, "创建 CronJob")

	var req k8s.YAMLCronJobCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logs.Error(logData, "请求参数绑定失败: "+err.Error())
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

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	// 解析YAML
	cronJob, err := c.createCronJobFromYAML(req.YAML, namespace)
	if err != nil {
		logs.Error(logData, "YAML解析失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("YAML解析失败: " + err.Error())
		return
	}

	// 确保Namespace设置正确
	if cronJob.Namespace == "" {
		cronJob.Namespace = namespace
	}

	// 首先尝试使用 batch/v1 API (Kubernetes 1.21+)
	if groups, err := client.Discovery().ServerGroups(); err == nil {
		for _, group := range groups.Groups {
			if group.Name == "batch" {
				for _, version := range group.Versions {
					if version.Version == "v1" {
						// 转换为 batch/v1 格式
						v1CronJob := c.convertV1Beta1ToV1CronJob(cronJob)
						_, err = client.BatchV1().CronJobs(cronJob.Namespace).Create(context.TODO(), v1CronJob, metav1.CreateOptions{})
						break
					}
				}
				break
			}
		}
	}

	// 如果 batch/v1 不可用或失败，尝试使用 batch/v1beta1
	if err != nil {
		_, err = client.BatchV1beta1().CronJobs(cronJob.Namespace).Create(context.TODO(), cronJob, metav1.CreateOptions{})
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
	logs.Debug(logData, "删除 CronJob")

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

		// 提取资源限制
		resources := k8s.ResourceInfo{
			CPURequest:    "0",
			CPULimit:      "0",
			MemoryRequest: "0",
			MemoryLimit:   "0",
		}

		if len(cj.Spec.JobTemplate.Spec.Template.Spec.Containers) > 0 {
			c := cj.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
			if c.Resources.Requests != nil {
				resources.CPURequest = c.Resources.Requests.Cpu().String()
				resources.MemoryRequest = c.Resources.Requests.Memory().String()
			}
			if c.Resources.Limits != nil {
				resources.CPULimit = c.Resources.Limits.Cpu().String()
				resources.MemoryLimit = c.Resources.Limits.Memory().String()
			}
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
			LastScheduleTime: func() *string {
				if cj.Status.LastScheduleTime != nil {
					t := cj.Status.LastScheduleTime.Time.Format(time.RFC3339)
					return &t
				}
				return nil
			}(),
			Resources: resources,
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

// createCronJobFromYAML 从YAML创建CronJob对象
func (c *CronJobController) createCronJobFromYAML(yamlContent string, namespace string) (*batchv1beta1.CronJob, error) {
	// 创建scheme并注册batch/v1, batch/v1beta1 和 core/v1类型
	scheme := runtime.NewScheme()
	_ = batchv1.AddToScheme(scheme)
	_ = batchv1beta1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// 使用K8s的反序列化器解析YAML
	decode := serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode
	obj, gvk, err := decode([]byte(yamlContent), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("YAML解析失败: %v", err)
	}

	// 根据GVK类型进行转换
	if gvk.Group == "batch" {
		if gvk.Version == "v1" {
			if v1CronJob, ok := obj.(*batchv1.CronJob); ok {
				return c.convertV1ToV1Beta1CronJob(v1CronJob), nil
			}
		} else if gvk.Version == "v1beta1" {
			if v1beta1CronJob, ok := obj.(*batchv1beta1.CronJob); ok {
				return v1beta1CronJob, nil
			}
		}
	}

	return nil, fmt.Errorf("YAML 内容不是有效的CronJob资源")
}

// GetCronJobDetail 获取CronJob详情
func (c *CronJobController) GetCronJobDetail(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	cronJobName := ctx.Param("cronJobName")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	// 尝试获取 CronJob (v1 > v1beta1)
	var cronJob *batchv1.CronJob
	var err error

	// Try batch/v1
	if groups, err := client.Discovery().ServerGroups(); err == nil {
		for _, group := range groups.Groups {
			if group.Name == "batch" {
				for _, version := range group.Versions {
					if version.Version == "v1" {
						cronJob, err = client.BatchV1().CronJobs(namespace).Get(context.TODO(), cronJobName, metav1.GetOptions{})
						break
					}
				}
				break
			}
		}
	}

	// Try batch/v1beta1 if needed, converting to v1 for consistent response structure
	if cronJob == nil {
		v1beta1CronJob, errBeta := client.BatchV1beta1().CronJobs(namespace).Get(context.TODO(), cronJobName, metav1.GetOptions{})
		if errBeta == nil {
			cronJob = c.convertV1Beta1ToV1CronJob(v1beta1CronJob)
			err = nil
		} else {
			err = errBeta
		}
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("CronJob不存在")
		return
	}

	// 提取详情信息
	containerName := ""
	image := ""
	command := []string{}
	if len(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers) > 0 {
		container := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
		containerName = container.Name
		image = container.Image
		command = container.Command
	}

	status := "suspended"
	if cronJob.Spec.Suspend != nil && !*cronJob.Spec.Suspend {
		status = "running"
	}
	if len(cronJob.Status.Active) > 0 {
		status = "active"
	}

	// 转换标签
	labels := cronJob.Labels

	resp := k8s.CronJobDetail{
		Name:          cronJob.Name,
		Namespace:     cronJob.Namespace,
		ContainerName: containerName,
		Image:         image,
		Command:       command,
		Schedule:      cronJob.Spec.Schedule,
		Status:        status,
		Age:           cronJob.CreationTimestamp.Unix(),
		Labels:        labels,
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "cronJobDetail", resp)
}

// GetCronJobYAML 获取CronJob YAML
func (c *CronJobController) GetCronJobYAML(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	cronJobName := ctx.Param("cronJobName")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	// 获取对象 (优先v1)
	var obj runtime.Object
	var err error

	// Try batch/v1
	if groups, err := client.Discovery().ServerGroups(); err == nil {
		for _, group := range groups.Groups {
			if group.Name == "batch" {
				for _, version := range group.Versions {
					if version.Version == "v1" {
						cj, err1 := client.BatchV1().CronJobs(namespace).Get(context.TODO(), cronJobName, metav1.GetOptions{})
						if err1 == nil {
							cj.ManagedFields = nil
							cj.TypeMeta = metav1.TypeMeta{Kind: "CronJob", APIVersion: "batch/v1"}
							obj = cj
							err = nil
						} else {
							err = err1
						}
						break
					}
				}
				break
			}
		}
	}

	// Try batch/v1beta1
	if obj == nil {
		cj, err1 := client.BatchV1beta1().CronJobs(namespace).Get(context.TODO(), cronJobName, metav1.GetOptions{})
		if err1 == nil {
			cj.ManagedFields = nil
			cj.TypeMeta = metav1.TypeMeta{Kind: "CronJob", APIVersion: "batch/v1beta1"}
			obj = cj
			err = nil
		} else {
			err = err1
		}
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("CronJob不存在")
		return
	}

	// 序列化为 YAML
	serializer := k8sjson.NewYAMLSerializer(k8sjson.DefaultMetaFactory, nil, nil)
	var sb strings.Builder
	err = serializer.Encode(obj, &sb)
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("YAML序列化失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "yaml", sb.String())
}

// UpdateCronJobYAML 更新CronJob (YAML方式)
func (c *CronJobController) UpdateCronJobYAML(ctx *gin.Context) {
	namespace := ctx.Param("namespace")

	var req k8s.YAMLCronJobUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	// 解析YAML
	newCronJob, err := c.createCronJobFromYAML(req.YAML, namespace)
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("YAML解析失败: " + err.Error())
		return
	}

	// 更新操作
	var updateErr error

	isV1Supported := false
	if groups, err := client.Discovery().ServerGroups(); err == nil {
		for _, group := range groups.Groups {
			if group.Name == "batch" {
				for _, version := range group.Versions {
					if version.Version == "v1" {
						isV1Supported = true
						break
					}
				}
				break
			}
		}
	}

	if isV1Supported {
		// 获取现有对象以保留 ResourceVersion 和 UID
		oldObj, err := client.BatchV1().CronJobs(namespace).Get(context.TODO(), newCronJob.Name, metav1.GetOptions{})
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.NotFound("CronJob不存在: " + err.Error())
			return
		}

		v1CronJob := c.convertV1Beta1ToV1CronJob(newCronJob)
		// 关键：保留元数据
		v1CronJob.ResourceVersion = oldObj.ResourceVersion
		v1CronJob.UID = oldObj.UID

		_, updateErr = client.BatchV1().CronJobs(namespace).Update(context.TODO(), v1CronJob, metav1.UpdateOptions{})
	} else {
		// 获取现有对象以保留 ResourceVersion 和 UID
		oldObj, err := client.BatchV1beta1().CronJobs(namespace).Get(context.TODO(), newCronJob.Name, metav1.GetOptions{})
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.NotFound("CronJob不存在: " + err.Error())
			return
		}

		// 关键：保留元数据
		newCronJob.ResourceVersion = oldObj.ResourceVersion
		newCronJob.UID = oldObj.UID

		_, updateErr = client.BatchV1beta1().CronJobs(namespace).Update(context.TODO(), newCronJob, metav1.UpdateOptions{})
	}

	if updateErr != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("更新CronJob失败: " + updateErr.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("更新成功")
}
