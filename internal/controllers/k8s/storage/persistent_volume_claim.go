package storage

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeClaimController PVC控制器
type PersistentVolumeClaimController struct{}

// NewPersistentVolumeClaimController 创建PVC控制器实例
func NewPersistentVolumeClaimController() *PersistentVolumeClaimController {
	return &PersistentVolumeClaimController{}
}

// GetPersistentVolumeClaimList 获取PVC列表
func (c *PersistentVolumeClaimController) GetPersistentVolumeClaimList(ctx *gin.Context) {
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

	var list *corev1.PersistentVolumeClaimList
	var err error

	if namespace == "all" {
		list, err = client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取PVC列表失败: " + err.Error())
		return
	}

	pvcList := make([]k8s.PersistentVolumeClaimListItem, 0)
	for _, item := range list.Items {
		pvcList = append(pvcList, k8s.PersistentVolumeClaimListItem{
			Name:         item.Name,
			Namespace:    item.Namespace,
			Status:       string(item.Status.Phase),
			Volume:       item.Spec.VolumeName,
			Capacity:     item.Status.Capacity.Storage().String(),
			AccessModes:  item.Spec.AccessModes,
			StorageClass: item.Spec.StorageClassName,
			Age:          item.CreationTimestamp.Unix(),
		})
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "pvcList", pvcList)
}

// GetPersistentVolumeClaimDetail 获取PVC详情
func (c *PersistentVolumeClaimController) GetPersistentVolumeClaimDetail(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	pvcName := ctx.Param("pvcname")
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

	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("PVC 不存在")
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "pvcDetail", pvc)
}

// CreatePersistentVolumeClaim 创建PVC
func (c *PersistentVolumeClaimController) CreatePersistentVolumeClaim(ctx *gin.Context) {
	var req k8s.PersistentVolumeClaimCreateRequest
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

	var pvc *corev1.PersistentVolumeClaim
	var err error

	if req.YAML != "" {
		pvc, err = c.parseYAMLToPVC(req.YAML)
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		pvc = c.convertCreateRequestToK8sPVC(&req)
	}

	_, err = client.CoreV1().PersistentVolumeClaims(req.Namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建PVC失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("PVC创建成功")
}

// DeletePersistentVolumeClaim 删除PVC
func (c *PersistentVolumeClaimController) DeletePersistentVolumeClaim(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	pvcName := ctx.Param("pvcname")
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

	err := client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除PVC失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("PVC删除成功")
}

func (c *PersistentVolumeClaimController) parseYAMLToPVC(yamlContent string) (*corev1.PersistentVolumeClaim, error) {
	var pvc corev1.PersistentVolumeClaim
	err := yaml.Unmarshal([]byte(yamlContent), &pvc)
	if err != nil {
		return nil, err
	}
	return &pvc, nil
}

func (c *PersistentVolumeClaimController) convertCreateRequestToK8sPVC(req *k8s.PersistentVolumeClaimCreateRequest) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &req.StorageClassName,
			AccessModes:      []corev1.PersistentVolumeAccessMode{},
		},
	}

	// Capacity
	if req.Capacity != "" {
		if capacity, err := resource.ParseQuantity(req.Capacity); err == nil {
			pvc.Spec.Resources = corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: capacity,
				},
			}
		}
	}

	// AccessModes
	for _, mode := range req.AccessModes {
		pvc.Spec.AccessModes = append(pvc.Spec.AccessModes, corev1.PersistentVolumeAccessMode(mode))
	}

	if len(req.Labels) > 0 {
		labels := make(map[string]string)
		for _, label := range req.Labels {
			if label.Key != "" && label.Value != "" {
				labels[label.Key] = label.Value
			}
		}
		pvc.ObjectMeta.Labels = labels
	}

	return pvc
}
