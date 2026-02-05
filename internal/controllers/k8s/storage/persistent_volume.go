package storage

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeController PV控制器
type PersistentVolumeController struct{}

// NewPersistentVolumeController 创建PV控制器实例
func NewPersistentVolumeController() *PersistentVolumeController {
	return &PersistentVolumeController{}
}

// GetPersistentVolumeList 获取PV列表
func (c *PersistentVolumeController) GetPersistentVolumeList(ctx *gin.Context) {
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

	list, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取PV列表失败: " + err.Error())
		return
	}

	pvList := make([]k8s.PersistentVolumeListItem, 0)
	for _, item := range list.Items {
		claim := ""
		if item.Spec.ClaimRef != nil {
			claim = fmt.Sprintf("%s/%s", item.Spec.ClaimRef.Namespace, item.Spec.ClaimRef.Name)
		}
		reclaimPolicy := item.Spec.PersistentVolumeReclaimPolicy

		pvList = append(pvList, k8s.PersistentVolumeListItem{
			Name:          item.Name,
			Status:        string(item.Status.Phase),
			Claim:         claim,
			AccessModes:   item.Spec.AccessModes,
			ReclaimPolicy: string(reclaimPolicy),
			Capacity:      item.Spec.Capacity.Storage().String(),
			StorageClass:  item.Spec.StorageClassName,
			Reason:        item.Status.Reason,
			Age:           item.CreationTimestamp.Unix(),
		})
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "pvList", pvList)
}

// GetPersistentVolumeDetail 获取PV详情
func (c *PersistentVolumeController) GetPersistentVolumeDetail(ctx *gin.Context) {
	pvName := ctx.Param("pvname")
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

	pv, err := client.CoreV1().PersistentVolumes().Get(ctx, pvName, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("PV 不存在")
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "pvDetail", pv)
}

// CreatePersistentVolume 创建PV
func (c *PersistentVolumeController) CreatePersistentVolume(ctx *gin.Context) {
	var req k8s.PersistentVolumeCreateRequest
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

	var pv *corev1.PersistentVolume
	var err error

	if req.YAML != "" {
		pv, err = c.parseYAMLToPV(req.YAML)
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		pv = c.convertCreateRequestToK8sPV(&req)
	}

	_, err = client.CoreV1().PersistentVolumes().Create(ctx, pv, metav1.CreateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建PV失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("PV创建成功")
}

// DeletePersistentVolume 删除PV
func (c *PersistentVolumeController) DeletePersistentVolume(ctx *gin.Context) {
	pvName := ctx.Param("pvname")
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

	err := client.CoreV1().PersistentVolumes().Delete(ctx, pvName, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除PV失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("PV 删除成功")
}

func (c *PersistentVolumeController) parseYAMLToPV(yamlContent string) (*corev1.PersistentVolume, error) {
	var pv corev1.PersistentVolume
	err := yaml.Unmarshal([]byte(yamlContent), &pv)
	if err != nil {
		return nil, err
	}
	return &pv, nil
}

func (c *PersistentVolumeController) convertCreateRequestToK8sPV(req *k8s.PersistentVolumeCreateRequest) *corev1.PersistentVolume {
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Name,
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: req.StorageClassName,
			AccessModes:      []corev1.PersistentVolumeAccessMode{},
		},
	}

	// Capacity
	if req.Capacity != "" {
		if capacity, err := resource.ParseQuantity(req.Capacity); err == nil {
			pv.Spec.Capacity = corev1.ResourceList{
				corev1.ResourceStorage: capacity,
			}
		}
	}

	// AccessModes
	for _, mode := range req.AccessModes {
		pv.Spec.AccessModes = append(pv.Spec.AccessModes, corev1.PersistentVolumeAccessMode(mode))
	}

	// HostPath as a simple example, in real world might need more types
	if req.HostPath != "" {
		pv.Spec.HostPath = &corev1.HostPathVolumeSource{
			Path: req.HostPath,
		}
	}

	// NFS
	if req.NFS != nil && req.NFS.Server != "" && req.NFS.Path != "" {
		pv.Spec.NFS = &corev1.NFSVolumeSource{
			Server: req.NFS.Server,
			Path:   req.NFS.Path,
		}
	}

	if len(req.Labels) > 0 {
		labels := make(map[string]string)
		for _, label := range req.Labels {
			if label.Key != "" && label.Value != "" {
				labels[label.Key] = label.Value
			}
		}
		pv.ObjectMeta.Labels = labels
	}

	return pv
}
