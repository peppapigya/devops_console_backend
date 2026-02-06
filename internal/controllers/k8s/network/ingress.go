package network

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressController Ingress控制器
type IngressController struct{}

// NewIngressController 创建Ingress控制器实例
func NewIngressController() *IngressController {
	return &IngressController{}
}

// GetIngressList 获取Ingress列表
func (c *IngressController) GetIngressList(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
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

	var listOptions metav1.ListOptions
	var list *networkingv1.IngressList
	var err error

	if namespace == "all" || namespace == "" {
		list, err = client.NetworkingV1().Ingresses("").List(ctx, listOptions)
	} else {
		list, err = client.NetworkingV1().Ingresses(namespace).List(ctx, listOptions)
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Ingress列表失败: " + err.Error())
		return
	}

	ingressList := make([]k8s.IngressListItem, 0)
	for _, item := range list.Items {
		var hosts []string
		for _, rule := range item.Spec.Rules {
			if rule.Host != "" {
				hosts = append(hosts, rule.Host)
			}
		}

		address := ""
		if len(item.Status.LoadBalancer.Ingress) > 0 {
			if item.Status.LoadBalancer.Ingress[0].IP != "" {
				address = item.Status.LoadBalancer.Ingress[0].IP
			} else if item.Status.LoadBalancer.Ingress[0].Hostname != "" {
				address = item.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		ingressList = append(ingressList, k8s.IngressListItem{
			Name:      item.Name,
			Namespace: item.Namespace,
			ClassName: item.Spec.IngressClassName,
			Hosts:     hosts,
			Address:   address,
			Age:       item.CreationTimestamp.Unix(),
		})
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "ingressList", ingressList)
}

// GetIngressDetail 获取Ingress详情
func (c *IngressController) GetIngressDetail(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
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

	ingress, err := client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("Ingress 不存在")
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "ingressDetail", ingress)
}

// CreateIngress 创建Ingress
func (c *IngressController) CreateIngress(ctx *gin.Context) {
	var req k8s.IngressCreateRequest
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

	var ingress *networkingv1.Ingress
	var err error

	if req.YAML != "" {
		ingress, err = c.parseYAMLToIngress(req.YAML)
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		ingress = c.convertCreateRequestToK8sIngress(&req)
	}

	_, err = client.NetworkingV1().Ingresses(req.Namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建Ingress失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Ingress创建成功")
}

// UpdateIngress 更新Ingress
func (c *IngressController) UpdateIngress(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	var req k8s.IngressUpdateRequest
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

	var ingress *networkingv1.Ingress
	var err error

	if req.YAML != "" {
		ingress, err = c.parseYAMLToIngress(req.YAML)
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		// Get existing ingress first
		existingIngress, err := client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.NotFound("Ingress 不存在")
			return
		}

		ingress = c.applyUpdateToIngress(existingIngress, &req)
	}

	_, err = client.NetworkingV1().Ingresses(namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("更新Ingress失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Ingress 更新成功")
}

// DeleteIngress 删除Ingress
func (c *IngressController) DeleteIngress(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
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

	err := client.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除Ingress失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Ingress 删除成功")
}

func (c *IngressController) parseYAMLToIngress(yamlContent string) (*networkingv1.Ingress, error) {
	var ingress networkingv1.Ingress
	err := yaml.Unmarshal([]byte(yamlContent), &ingress)
	if err != nil {
		return nil, err
	}
	return &ingress, nil
}

func (c *IngressController) convertCreateRequestToK8sIngress(req *k8s.IngressCreateRequest) *networkingv1.Ingress {
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Annotations: req.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: req.ClassName,
		},
	}

	// Convert rules
	for _, rule := range req.Rules {
		k8sRule := networkingv1.IngressRule{
			Host: rule.Host,
		}

		if len(rule.Paths) > 0 {
			k8sRule.IngressRuleValue = networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{},
				},
			}

			for _, path := range rule.Paths {
				pathType := networkingv1.PathType(path.PathType)
				k8sPath := networkingv1.HTTPIngressPath{
					Path:     path.Path,
					PathType: &pathType,
					Backend: networkingv1.IngressBackend{
						Service: &networkingv1.IngressServiceBackend{
							Name: path.ServiceName,
							Port: networkingv1.ServiceBackendPort{
								Number: path.ServicePort,
							},
						},
					},
				}
				k8sRule.HTTP.Paths = append(k8sRule.HTTP.Paths, k8sPath)
			}
		}

		ingress.Spec.Rules = append(ingress.Spec.Rules, k8sRule)
	}

	// Convert TLS
	for _, tls := range req.TLS {
		k8sTLS := networkingv1.IngressTLS{
			Hosts:      tls.Hosts,
			SecretName: tls.SecretName,
		}
		ingress.Spec.TLS = append(ingress.Spec.TLS, k8sTLS)
	}

	return ingress
}

func (c *IngressController) applyUpdateToIngress(existing *networkingv1.Ingress, req *k8s.IngressUpdateRequest) *networkingv1.Ingress {
	existing.Spec.IngressClassName = req.ClassName

	if req.Annotations != nil {
		existing.Annotations = req.Annotations
	}

	// Update rules
	existing.Spec.Rules = []networkingv1.IngressRule{}
	for _, rule := range req.Rules {
		k8sRule := networkingv1.IngressRule{
			Host: rule.Host,
		}

		if len(rule.Paths) > 0 {
			k8sRule.IngressRuleValue = networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{},
				},
			}

			for _, path := range rule.Paths {
				pathType := networkingv1.PathType(path.PathType)
				k8sPath := networkingv1.HTTPIngressPath{
					Path:     path.Path,
					PathType: &pathType,
					Backend: networkingv1.IngressBackend{
						Service: &networkingv1.IngressServiceBackend{
							Name: path.ServiceName,
							Port: networkingv1.ServiceBackendPort{
								Number: path.ServicePort,
							},
						},
					},
				}
				k8sRule.HTTP.Paths = append(k8sRule.HTTP.Paths, k8sPath)
			}
		}

		existing.Spec.Rules = append(existing.Spec.Rules, k8sRule)
	}

	// Update TLS
	existing.Spec.TLS = []networkingv1.IngressTLS{}
	for _, tls := range req.TLS {
		k8sTLS := networkingv1.IngressTLS{
			Hosts:      tls.Hosts,
			SecretName: tls.SecretName,
		}
		existing.Spec.TLS = append(existing.Spec.TLS, k8sTLS)
	}

	return existing
}
