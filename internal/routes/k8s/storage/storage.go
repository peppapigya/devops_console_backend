package storage

import (
	"devops-console-backend/internal/controllers/k8s/storage"

	"github.com/gin-gonic/gin"
)

type StorageRoute struct {
	pvController  *storage.PersistentVolumeController
	pvcController *storage.PersistentVolumeClaimController
	scController  *storage.StorageClassController
}

func NewStorageRoute() *StorageRoute {
	return &StorageRoute{
		pvController:  storage.NewPersistentVolumeController(),
		pvcController: storage.NewPersistentVolumeClaimController(),
		scController:  storage.NewStorageClassController(),
	}
}

func (r *StorageRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	// PV
	pvGroup := apiGroup.Group("/k8s/pv")
	{
		pvGroup.GET("/list", r.pvController.GetPersistentVolumeList)
		pvGroup.GET("/detail/:pvname", r.pvController.GetPersistentVolumeDetail)
		pvGroup.POST("/create", r.pvController.CreatePersistentVolume)
		pvGroup.DELETE("/delete/:pvname", r.pvController.DeletePersistentVolume)
	}

	// PVC
	pvcGroup := apiGroup.Group("/k8s/pvc")
	{
		pvcGroup.GET("/list/:namespace", r.pvcController.GetPersistentVolumeClaimList)
		pvcGroup.GET("/list/all", r.pvcController.GetPersistentVolumeClaimList) // If supported, though logic might need all override
		pvcGroup.GET("/detail/:namespace/:pvcname", r.pvcController.GetPersistentVolumeClaimDetail)
		pvcGroup.POST("/create", r.pvcController.CreatePersistentVolumeClaim)
		pvcGroup.DELETE("/delete/:namespace/:pvcname", r.pvcController.DeletePersistentVolumeClaim)
	}

	// StorageClass
	scGroup := apiGroup.Group("/k8s/storageclass")
	{
		scGroup.GET("/list", r.scController.GetStorageClassList)
		scGroup.GET("/detail/:scname", r.scController.GetStorageClassDetail)
		scGroup.POST("/create", r.scController.CreateStorageClass)
		scGroup.DELETE("/delete/:scname", r.scController.DeleteStorageClass)
	}
}
