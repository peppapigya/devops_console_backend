package deployment

import (
	"devops-console-backend/config"
	"devops-console-backend/models/k8s"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/logs"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentController Deployment控制器
type DeploymentController struct{}

// NewDeploymentController 创建Deployment控制器实例
func NewDeploymentController() *DeploymentController {
	return &DeploymentController{}
}

// GetDeploymentDetail 获取Deployment详情
func (c *DeploymentController) GetDeploymentDetail(ctx *gin.Context) {
	logData := map[string]interface{}{
		"namespace":      ctx.Param("namespace"),
		"deploymentName": ctx.Param("deploymentName"),
	}
	logs.Debug(logData, "获取Deployment详情")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	deploymentDetail, err := client.AppsV1().Deployments(logData["namespace"].(string)).Get(ctx, logData["deploymentName"].(string), metav1.GetOptions{})
	if err != nil {
		logs.Error(logData, "获取Deployment失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("Deployment不存在")
		return
	}

	// 转换Conditions为interface{}
	conditions := make([]interface{}, len(deploymentDetail.Status.Conditions))
	for i, condition := range deploymentDetail.Status.Conditions {
		conditions[i] = condition
	}

	logs.Info(logData, "获取Deployment详情成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "deploymentDetail", k8s.DeploymentDetail{
		Name:       deploymentDetail.Name,
		Namespace:  deploymentDetail.Namespace,
		Replicas:   deploymentDetail.Status.Replicas,
		Ready:      deploymentDetail.Status.ReadyReplicas,
		Available:  deploymentDetail.Status.AvailableReplicas,
		Conditions: conditions,
		Labels:     deploymentDetail.Labels,
		Age:        deploymentDetail.CreationTimestamp.Unix(),
	})
}

// GetDeploymentList 获取Deployment列表
func (c *DeploymentController) GetDeploymentList(ctx *gin.Context) {
	namespace := ctx.Param("namespace")

	// 如果 namespace 为 "all"，则使用空字符串获取所有命名空间的资源
	if namespace == "all" {
		namespace = ""
	}

	logData := map[string]interface{}{"namespace": namespace}
	logs.Debug(logData, "获取Deployment列表")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	deploymentList, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logs.Error(logData, "获取Deployment列表失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Deployment列表失败")
		return
	}

	// 简化返回数据，只返回关键信息
	var simplifiedList []k8s.DeploymentListItem
	for _, deployment := range deploymentList.Items {
		simplifiedList = append(simplifiedList, k8s.DeploymentListItem{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			Replicas:  deployment.Status.Replicas,
			Ready:     deployment.Status.ReadyReplicas,
			Available: deployment.Status.AvailableReplicas,
			Created:   deployment.CreationTimestamp.Time,
		})
	}

	logs.Info(map[string]interface{}{"count": len(simplifiedList), "data": logData}, "获取Deployment列表成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "deploymentList", simplifiedList)
}

// CreateDeployment 创建Deployment
func (c *DeploymentController) CreateDeployment(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	logData := map[string]interface{}{"namespace": namespace}
	logs.Debug(logData, "创建Deployment")

	var deploymentReq k8s.DeploymentCreateRequest
	if err := ctx.ShouldBindJSON(&deploymentReq); err != nil {
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

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	deployment := c.convertCreateRequestToK8sDeployment(namespace, deploymentReq)
	_, err := client.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		logs.Error(map[string]interface{}{"deploymentName": deploymentReq.Name, "error": err.Error(), "data": logData}, "创建Deployment失败")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建Deployment失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{"deploymentName": deploymentReq.Name, "data": logData}, "创建Deployment成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("Deployment创建成功", "data", map[string]interface{}{
		"name":      deploymentReq.Name,
		"namespace": namespace,
	})
}

// UpdateDeployment 更新Deployment
func (c *DeploymentController) UpdateDeployment(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	deploymentName := ctx.Param("deploymentName")
	logData := map[string]interface{}{
		"namespace":      namespace,
		"deploymentName": deploymentName,
	}
	logs.Debug(logData, "更新Deployment")

	var updateReq k8s.DeploymentUpdateRequest
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
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

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	// 获取现有的Deployment
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		logs.Error(logData, "获取Deployment失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("Deployment不存在")
		return
	}

	// 更新镜像版本
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Image = updateReq.Image
	}

	// 添加更新注释
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	deployment.Annotations["updated-by"] = "devops-console"
	deployment.Annotations["updated-at"] = metav1.Now().String()

	_, err = client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		logs.Error(map[string]interface{}{"image": updateReq.Image, "error": err.Error(), "data": logData}, "更新Deployment失败")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("更新Deployment失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{"image": updateReq.Image, "data": logData}, "更新Deployment成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("Deployment更新成功", "data", map[string]interface{}{
		"name":      deployment.Name,
		"namespace": deployment.Namespace,
		"image":     updateReq.Image,
	})
}

// ScaleDeployment 扩缩容Deployment
func (c *DeploymentController) ScaleDeployment(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	deploymentName := ctx.Param("deploymentName")
	replicasStr := ctx.Param("replicas")
	logData := map[string]interface{}{
		"namespace":      namespace,
		"deploymentName": deploymentName,
		"replicasStr":    replicasStr,
	}
	logs.Debug(logData, "扩缩容Deployment")

	replicas, err := strconv.ParseInt(replicasStr, 10, 32)
	if err != nil {
		logs.Error(logData, "副本数参数转换失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("副本数参数错误")
		return
	}

	replicas32 := int32(replicas)
	if replicas32 < 0 {
		logs.Error(map[string]interface{}{"replicas": replicas32, "data": logData}, "副本数不能为负数")
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("副本数不能为负数")
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
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	// 获取现有的Deployment
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		logs.Error(logData, "获取Deployment失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("Deployment不存在")
		return
	}

	// 记录原始副本数
	originalReplicas := int32(0)
	if deployment.Spec.Replicas != nil {
		originalReplicas = *deployment.Spec.Replicas
	}

	// 更新副本数
	deployment.Spec.Replicas = &replicas32

	// 添加扩缩容注释
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	deployment.Annotations["scaled-by"] = "devops-console"
	deployment.Annotations["scaled-at"] = metav1.Now().String()
	deployment.Annotations["previous-replicas"] = fmt.Sprintf("%d", originalReplicas)

	_, err = client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		logs.Error(map[string]interface{}{"replicas": replicas32, "error": err.Error(), "data": logData}, "扩缩容Deployment失败")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("扩缩容Deployment失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{"originalReplicas": originalReplicas, "newReplicas": replicas32, "data": logData}, "扩缩容Deployment成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("Deployment扩缩容成功", "data", map[string]interface{}{
		"name":      deployment.Name,
		"namespace": deployment.Namespace,
		"replicas":  replicas32,
	})
}

// DeleteDeployment 删除Deployment
func (c *DeploymentController) DeleteDeployment(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	deploymentName := ctx.Param("deploymentName")
	logData := map[string]interface{}{
		"namespace":      namespace,
		"deploymentName": deploymentName,
	}
	logs.Debug(logData, "删除Deployment")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	deletePolicy := metav1.DeletePropagationForeground
	err := client.AppsV1().Deployments(namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		logs.Error(logData, "删除Deployment失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除Deployment失败: " + err.Error())
		return
	}

	logs.Info(logData, "删除Deployment成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("Deployment删除成功", "data", map[string]interface{}{
		"name":      deploymentName,
		"namespace": namespace,
	})
}

// convertCreateRequestToK8sDeployment 转换创建请求为K8s Deployment
func (c *DeploymentController) convertCreateRequestToK8sDeployment(namespace string, req k8s.DeploymentCreateRequest) *appsv1.Deployment {
	if req.Labels == nil {
		req.Labels = map[string]string{
			"app": req.Name,
		}
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
			Labels:    req.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &req.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: req.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: req.Labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  req.Name,
							Image: req.Image,
							Ports: []corev1.ContainerPort{},
						},
					},
				},
			},
		},
	}

	// 如果指定了端口，添加端口配置
	if req.Port > 0 {
		deployment.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: req.Port,
				Protocol:      corev1.ProtocolTCP,
			},
		}
	}

	return deployment
}
