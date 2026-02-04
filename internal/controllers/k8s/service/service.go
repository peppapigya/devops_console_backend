package service

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ServiceController Service控制器
type ServiceController struct{}

// NewServiceController 创建Service控制器实例
func NewServiceController() *ServiceController {
	return &ServiceController{}
}

// GetServiceDetail 获取Service详情
func (c *ServiceController) GetServiceDetail(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
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

	service, err := client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("Service不存在")
		return
	}

	// 获取端点信息
	endpoints, err := c.getEndpoints(ctx, namespace, name)
	var endpointsInterface interface{}
	if err != nil {
		endpointsInterface = map[string]interface{}{"message": "获取端点信息失败: " + err.Error()}
	} else {
		endpointsInterface = endpoints
	}

	serviceDetail := k8s.ServiceDetail{
		Name:                  service.Name,
		Namespace:             service.Namespace,
		Status:                service.Status,
		Labels:                service.Labels,
		Annotations:           service.Annotations,
		Selector:              service.Spec.Selector,
		Type:                  service.Spec.Type,
		IP:                    service.Spec.ClusterIP,
		IPs:                   service.Spec.ClusterIPs,
		Ports:                 service.Spec.Ports,
		SessionAffinity:       service.Spec.SessionAffinity,
		ExternalTrafficPolicy: service.Spec.ExternalTrafficPolicy,
		Events:                service.Status.LoadBalancer.Ingress,
		Endpoints:             endpointsInterface,
		Age:                   service.CreationTimestamp.Unix(),
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "serviceDetail", serviceDetail)
}

// GetServiceList 获取Service列表
func (c *ServiceController) GetServiceList(ctx *gin.Context) {
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

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	servicesList, err := client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Service列表失败")
		return
	}

	// 构建服务列表
	services := make([]k8s.ServiceListItem, 0, len(servicesList.Items))
	for _, svc := range servicesList.Items {
		services = append(services, k8s.ServiceListItem{
			Namespace:  svc.Namespace,
			Name:       svc.Name,
			Labels:     svc.Labels,
			Type:       svc.Spec.Type,
			ClusterIP:  svc.Spec.ClusterIP,
			ExternalIP: svc.Spec.ExternalIPs,
			Selector:   svc.Spec.Selector,
			Ports:      svc.Spec.Ports,
			CreatedAt:  svc.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Age:        svc.CreationTimestamp.Unix(),
		})
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "services", services)
}

// CreateService 创建Service
func (c *ServiceController) CreateService(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
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

	service, err := c.parseServiceRequest(ctx, name, namespace)
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	_, err = client.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建Service失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Service创建成功")
}

// UpdateService 更新Service
func (c *ServiceController) UpdateService(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
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

	// 获取现有服务
	existingService, err := client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("Service不存在")
		return
	}

	// 解析更新请求
	var request k8s.ServiceUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	// 更新服务配置
	updatedService := existingService.DeepCopy()
	if request.Labels != nil {
		updatedService.Labels = request.Labels
	}
	if request.Selector != nil {
		updatedService.Spec.Selector = request.Selector
	}
	if request.Ports != nil {
		updatedService.Spec.Ports = c.convertToServicePorts(request.Ports)
	}

	_, err = client.CoreV1().Services(namespace).Update(ctx, updatedService, metav1.UpdateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("更新Service失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Service更新成功")
}

// DeleteService 删除Service
func (c *ServiceController) DeleteService(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
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

	err := client.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除Service失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Service删除成功")
}

// DeleteMultipleServices 删除多个Service
func (c *ServiceController) DeleteMultipleServices(ctx *gin.Context) {
	namespace := ctx.Param("namespace")

	var request k8s.MultipleDeleteRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	if len(request.Services) == 0 {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("services不能为空")
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
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	// 用于记录成功和失败的服务名
	success := make([]string, 0)
	failures := make([]string, 0)

	for _, serviceName := range request.Services {
		err := client.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
		if err != nil {
			failures = append(failures, serviceName)
		} else {
			success = append(success, serviceName)
		}
	}

	// 根据是否有失败情况返回不同的消息
	if len(failures) == 0 {
		helper := utils.NewResponseHelper(ctx)
		helper.SuccessWithData("所有服务删除成功", "success", success)
	} else {
		helper := utils.NewResponseHelper(ctx)
		helper.SuccessWithData("部分服务删除失败", "data", map[string]interface{}{
			"success": success,
			"failed":  failures,
		})
	}
}

// DeleteServiceByLabel 根据标签删除Service
func (c *ServiceController) DeleteServiceByLabel(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	labelSelector := ctx.Query("labelSelector")

	if labelSelector == "" {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("labelSelector不能为空")
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
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	services, err := client.CoreV1().Services(namespace).List(ctx, listOptions)
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("查询服务失败: " + err.Error())
		return
	}

	if len(services.Items) == 0 {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("未找到匹配标签选择器的服务")
		return
	}

	// 批量删除匹配的服务
	success := make([]string, 0)
	failures := make([]string, 0)

	for _, service := range services.Items {
		err := client.CoreV1().Services(namespace).Delete(ctx, service.Name, metav1.DeleteOptions{})
		if err != nil {
			failures = append(failures, service.Name)
		} else {
			success = append(success, service.Name)
		}
	}

	// 根据是否有失败情况返回不同的消息
	if len(failures) == 0 {
		helper := utils.NewResponseHelper(ctx)
		helper.SuccessWithData("所有服务删除成功", "success", success)
	} else {
		helper := utils.NewResponseHelper(ctx)
		helper.SuccessWithData("部分服务删除失败", "data", map[string]interface{}{
			"success": success,
			"failed":  failures,
		})
	}
}

// 辅助方法

// getEndpoints 获取端点信息
func (c *ServiceController) getEndpoints(ctx *gin.Context, namespace, name string) (interface{}, error) {
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		return nil, nil
	}

	endpoints, err := client.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return map[string]interface{}{"message": "获取端点信息失败: " + err.Error()}, err
	}

	return map[string]interface{}{
		"namespace": namespace,
		"name":      name,
		"subsets":   endpoints.Subsets,
	}, nil
}

// parseServiceRequest 解析服务请求
func (c *ServiceController) parseServiceRequest(ctx *gin.Context, name, namespace string) (*corev1.Service, error) {
	var request k8s.ServiceCreateRequest

	// 设置默认值
	request.Type = corev1.ServiceTypeClusterIP
	request.Selector = map[string]string{"app": name}
	request.Ports = []k8s.ServicePortRequest{
		{
			Name:       "http",
			Port:       80,
			TargetPort: intstr.FromInt(8080),
			Protocol:   "TCP",
		},
	}

	// 尝试绑定JSON，如果失败则使用默认值
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 这里可以选择记录日志，但继续使用默认值
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    request.Labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     request.Type,
			Selector: request.Selector,
			Ports:    c.convertToServicePorts(request.Ports),
		},
	}, nil
}

// convertToServicePorts 转换端口配置
func (c *ServiceController) convertToServicePorts(portRequests []k8s.ServicePortRequest) []corev1.ServicePort {
	servicePorts := make([]corev1.ServicePort, len(portRequests))

	for i, port := range portRequests {
		protocol := corev1.ProtocolTCP
		if port.Protocol == "UDP" {
			protocol = corev1.ProtocolUDP
		}

		servicePorts[i] = corev1.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: port.TargetPort,
			NodePort:   port.NodePort,
			Protocol:   protocol,
		}
	}

	return servicePorts
}
