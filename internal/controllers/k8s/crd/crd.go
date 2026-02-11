package crd

import (
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CRDController struct{}

func NewCRDController() *CRDController {
	return &CRDController{}
}

func (c *CRDController) GetCRDList(ctx *gin.Context) {
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetApiExtensionsClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s ApiExtensions客户端未初始化")
		return
	}

	list, err := client.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取CRD列表失败: " + err.Error())
		return
	}

	crdList := make([]gin.H, 0)
	for _, item := range list.Items {
		crdList = append(crdList, gin.H{
			"name":    item.Name,
			"group":   item.Spec.Group,
			"version": getCrdVersions(item.Spec.Versions),
			"scope":   item.Spec.Scope,
			"age":     item.CreationTimestamp.Unix(),
		})
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "crdList", crdList)
}

func (c *CRDController) GetCRDDetail(ctx *gin.Context) {
	name := ctx.Param("name")
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetApiExtensionsClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s ApiExtensions客户端未初始化")
		return
	}

	crd, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取CRD详情失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "crdDetail", crd)
}

func (c *CRDController) DeleteCRD(ctx *gin.Context) {
	name := ctx.Param("name")
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetApiExtensionsClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s ApiExtensions客户端未初始化")
		return
	}

	err := client.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除CRD失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("CRD删除成功")
}
func getCrdVersions(versions []apiextensionsv1.CustomResourceDefinitionVersion) []string {
	v := make([]string, 0, len(versions))
	for _, version := range versions {
		v = append(v, version.Name)
	}
	return v
}
