package storage

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	v2 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageClassController struct{}

func NewStorageClassController() *StorageClassController {
	return &StorageClassController{}
}

func (c *StorageClassController) GetStorageClassList(ctx *gin.Context) {
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

	list, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取StorageClass列表失败: " + err.Error())
		return
	}

	scList := make([]k8s.StorageClassListItem, 0)
	for _, item := range list.Items {
		volumeBindingMode := item.VolumeBindingMode
		scList = append(scList, k8s.StorageClassListItem{
			Name:              item.Name,
			Provisioner:       item.Provisioner,
			ReclaimPolicy:     item.ReclaimPolicy,
			VolumeBindingMode: string(*volumeBindingMode),
			Age:               item.CreationTimestamp.Unix(),
		})
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "scList", scList)
}

func (c *StorageClassController) GetStorageClassDetail(ctx *gin.Context) {
	scName := ctx.Param("scname")
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

	sc, err := client.StorageV1().StorageClasses().Get(ctx, scName, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("StorageClass不存在")
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "scDetail", sc)
}

func (c *StorageClassController) CreateStorageClass(ctx *gin.Context) {
	var req k8s.StorageClassCreateRequest
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

	var sc *v1.StorageClass
	var err error

	if req.YAML != "" {
		sc, err = c.parseYAMLToSC(req.YAML)
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		sc = c.convertCreateRequestToK8sSC(&req)
	}

	_, err = client.StorageV1().StorageClasses().Create(ctx, sc, metav1.CreateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建StorageClass失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("StorageClass创建成功")
}

func (c *StorageClassController) DeleteStorageClass(ctx *gin.Context) {
	scName := ctx.Param("scname")
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

	err := client.StorageV1().StorageClasses().Delete(ctx, scName, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除StorageClass失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("StorageClass删除成功")
}

func (c *StorageClassController) parseYAMLToSC(yamlContent string) (*v1.StorageClass, error) {
	var sc v1.StorageClass
	err := yaml.Unmarshal([]byte(yamlContent), &sc)
	if err != nil {
		return nil, err
	}
	return &sc, nil
}

func (c *StorageClassController) convertCreateRequestToK8sSC(req *k8s.StorageClassCreateRequest) *v1.StorageClass {
	sc := &v1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Name,
		},
		Provisioner: req.Provisioner,
	}
	if req.ReclaimPolicy != "" {
		reclaimPolicy := v2.PersistentVolumeReclaimPolicy(req.ReclaimPolicy)
		sc.ReclaimPolicy = &reclaimPolicy
	}

	if len(req.Parameters) > 0 {
		params := make(map[string]string)
		for _, param := range req.Parameters {
			if param.Key != "" && param.Value != "" {
				params[param.Key] = param.Value
			}
		}
		sc.Parameters = params
	}

	return sc
}
