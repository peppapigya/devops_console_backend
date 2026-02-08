package config

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"encoding/base64"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretController Secret控制器
type SecretController struct{}

// NewSecretController 创建Secret控制器实例
func NewSecretController() *SecretController {
	return &SecretController{}
}

// GetSecretList 获取Secret列表
func (c *SecretController) GetSecretList(ctx *gin.Context) {
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
	var list *corev1.SecretList
	var err error

	if namespace == "all" || namespace == "" {
		list, err = client.CoreV1().Secrets("").List(ctx, listOptions)
	} else {
		list, err = client.CoreV1().Secrets(namespace).List(ctx, listOptions)
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Secret列表失败: " + err.Error())
		return
	}

	secretList := make([]k8s.SecretListItem, 0)
	for _, item := range list.Items {
		secretList = append(secretList, k8s.SecretListItem{
			Name:      item.Name,
			Namespace: item.Namespace,
			Type:      string(item.Type),
			DataCount: len(item.Data),
			Age:       item.CreationTimestamp.Unix(),
		})
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "secretList", secretList)
}

// GetSecretDetail 获取Secret详情
func (c *SecretController) GetSecretDetail(ctx *gin.Context) {
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

	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("Secret 不存在")
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "secretDetail", secret)
}

// CreateSecret 创建Secret
func (c *SecretController) CreateSecret(ctx *gin.Context) {
	var req k8s.SecretCreateRequest
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

	var secret *corev1.Secret
	var err error

	if req.YAML != "" {
		secret, err = c.parseYAMLToSecret(req.YAML)
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		// Convert string data to byte data
		data := make(map[string][]byte)
		for k, v := range req.Data {
			// Decode base64 if already encoded, otherwise encode it
			decoded, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				// Not base64, treat as plain text
				data[k] = []byte(v)
			} else {
				data[k] = decoded
			}
		}

		secretType := corev1.SecretTypeOpaque
		if req.Type != "" {
			secretType = corev1.SecretType(req.Type)
		}

		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
			Type: secretType,
			Data: data,
		}
	}

	_, err = client.CoreV1().Secrets(req.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建Secret失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Secret创建成功")
}

// UpdateSecret 更新Secret
func (c *SecretController) UpdateSecret(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	var req k8s.SecretUpdateRequest
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

	var secret *corev1.Secret
	var err error

	if req.YAML != "" {
		secret, err = c.parseYAMLToSecret(req.YAML)
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		// Get existing secret first
		existingSecret, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.NotFound("Secret 不存在")
			return
		}

		// Convert string data to byte data
		data := make(map[string][]byte)
		for k, v := range req.Data {
			decoded, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				data[k] = []byte(v)
			} else {
				data[k] = decoded
			}
		}

		existingSecret.Data = data
		secret = existingSecret
	}

	_, err = client.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("更新Secret失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Secret 更新成功")
}

// DeleteSecret 删除Secret
func (c *SecretController) DeleteSecret(ctx *gin.Context) {
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

	err := client.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除Secret失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Secret 删除成功")
}

func (c *SecretController) parseYAMLToSecret(yamlContent string) (*corev1.Secret, error) {
	var secret corev1.Secret
	err := yaml.Unmarshal([]byte(yamlContent), &secret)
	if err != nil {
		return nil, err
	}
	return &secret, nil
}
