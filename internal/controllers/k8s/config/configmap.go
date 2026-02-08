package config

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapController ConfigMap控制器
type ConfigMapController struct{}

// NewConfigMapController 创建ConfigMap控制器实例
func NewConfigMapController() *ConfigMapController {
	return &ConfigMapController{}
}

// GetConfigMapList 获取ConfigMap列表
func (c *ConfigMapController) GetConfigMapList(ctx *gin.Context) {
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
	var list *corev1.ConfigMapList
	var err error

	if namespace == "all" || namespace == "" {
		list, err = client.CoreV1().ConfigMaps("").List(ctx, listOptions)
	} else {
		list, err = client.CoreV1().ConfigMaps(namespace).List(ctx, listOptions)
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取ConfigMap列表失败: " + err.Error())
		return
	}

	configMapList := make([]k8s.ConfigMapListItem, 0)
	for _, item := range list.Items {
		configMapList = append(configMapList, k8s.ConfigMapListItem{
			Name:      item.Name,
			Namespace: item.Namespace,
			DataCount: len(item.Data),
			Age:       item.CreationTimestamp.Unix(),
		})
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "configMapList", configMapList)
}

// GetConfigMapDetail 获取ConfigMap详情
func (c *ConfigMapController) GetConfigMapDetail(ctx *gin.Context) {
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

	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("ConfigMap 不存在")
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "configMapDetail", configMap)
}

// CreateConfigMap 创建ConfigMap
func (c *ConfigMapController) CreateConfigMap(ctx *gin.Context) {
	var req k8s.ConfigMapCreateRequest
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

	var configMap *corev1.ConfigMap
	var err error

	if req.YAML != "" {
		configMap, err = c.parseYAMLToConfigMap(req.YAML)
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
			Data: req.Data,
		}
	}

	_, err = client.CoreV1().ConfigMaps(req.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建ConfigMap失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("ConfigMap创建成功")
}

// UpdateConfigMap 更新ConfigMap
func (c *ConfigMapController) UpdateConfigMap(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	var req k8s.ConfigMapUpdateRequest
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

	var configMap *corev1.ConfigMap
	var err error

	if req.YAML != "" {
		configMap, err = c.parseYAMLToConfigMap(req.YAML)
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		// Get existing configmap first
		existingConfigMap, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.NotFound("ConfigMap 不存在")
			return
		}

		existingConfigMap.Data = req.Data
		configMap = existingConfigMap
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("更新ConfigMap失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("ConfigMap 更新成功")
}

// DeleteConfigMap 删除ConfigMap
func (c *ConfigMapController) DeleteConfigMap(ctx *gin.Context) {
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

	err := client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除ConfigMap失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("ConfigMap 删除成功")
}

func (c *ConfigMapController) parseYAMLToConfigMap(yamlContent string) (*corev1.ConfigMap, error) {
	var configMap corev1.ConfigMap
	err := yaml.Unmarshal([]byte(yamlContent), &configMap)
	if err != nil {
		return nil, err
	}
	return &configMap, nil
}
