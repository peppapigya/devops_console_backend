package instance

import (
	"devops-console-backend/config"
	"devops-console-backend/database"
	"devops-console-backend/models"
	"devops-console-backend/models/request"
	"devops-console-backend/repositories"
	"devops-console-backend/utils"
	"devops-console-backend/utils/logs"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// addOrUpdateK8sClient 添加或更新k8s客户端
func addOrUpdateK8sClient(instance *models.Instance, authConfig *models.AuthConfig) error {
	if authConfig.AuthType != "kubeconfig" {
		return nil
	}

	return config.AddK8sClient(instance, authConfig)
}

// removeK8sClient 移除k8s客户端
func removeK8sClient(instanceID uint) {
	config.RemoveK8sClient(instanceID)
}

// handleInstanceOperation 处理实例操作（添加/更新）
func handleInstanceOperation(r *gin.Context, isAdd bool) {
	helper := utils.NewResponseHelper(r)
	var req request.InstanceRequest
	operation := "添加"
	if !isAdd {
		operation = "修改"
	}

	// 绑定请求参数
	if err := r.ShouldBindJSON(&req); err != nil {
		helper.LogAndBadRequest("请求参数绑定失败", map[string]interface{}{"error": err.Error()})
		return
	}

	// 设置默认状态
	if req.Status == "" {
		req.Status = "active"
	}

	// 使用GORM处理事务
	instanceRepo := repositories.NewInstanceRepository()
	authConfigRepo := repositories.NewAuthConfigRepository()

	var instanceID uint
	var err error

	if isAdd {
		id, err := addInstance(instanceRepo, req)
		if err == nil {
			instanceID = uint(id)
		}
	} else {
		id, err := updateInstance(instanceRepo, req)
		if err == nil {
			instanceID = uint(id)
		}
	}

	if err != nil {
		helper.TransactionError(operation, err.Error())
		return
	}

	// 处理认证配置
	if err = handleAuthConfig(authConfigRepo, instanceID, &req.AuthConfig, isAdd); err != nil {
		helper.TransactionError(operation, "处理认证配置失败: "+err.Error())
		return
	}

	logs.Info(map[string]interface{}{
		"operation":   operation,
		"instance_id": instanceID,
		"name":        req.Name,
	}, "集群操作成功")

	helper.SuccessWithData(operation+"成功", "instance_id", instanceID)
}

// addInstance 添加实例
func addInstance(repo *repositories.InstanceRepository, req request.InstanceRequest) (int, error) {
	// 验证实例类型存在性
	instanceTypeRepo := repositories.NewInstanceTypeRepository()
	if _, err := instanceTypeRepo.GetByID(req.InstanceTypeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.Warning(map[string]interface{}{"instance_type_id": req.InstanceTypeID}, "实例类型不存在")
			return 0, errors.New("实例类型不存在")
		}
		return 0, errors.New("查询实例类型失败: " + err.Error())
	}

	// 检查实例名称唯一性
	if _, err := repo.GetByName(req.Name); err == nil {
		logs.Warning(map[string]interface{}{"name": req.Name}, "实例名称已存在")
		return 0, errors.New("实例名称已存在")
	}

	// 创建实例
	instance := &models.Instance{
		InstanceTypeID: req.InstanceTypeID,
		Name:           req.Name,
		Address:        req.Address,
		HttpsEnabled:   req.HttpsEnabled,
		SkipSslVerify:  req.SkipSslVerify,
		Status:         req.Status,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := repo.Create(instance); err != nil {
		return 0, errors.New("添加实例失败: " + err.Error())
	}

	return int(instance.ID), nil
}

// updateInstance 更新实例
func updateInstance(repo *repositories.InstanceRepository, req request.InstanceRequest) (int, error) {
	if req.ID == 0 {
		return 0, errors.New("更新操作需要提供实例ID")
	}

	// 验证实例类型存在性
	instanceTypeRepo := repositories.NewInstanceTypeRepository()
	if _, err := instanceTypeRepo.GetByID(req.InstanceTypeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.Warning(map[string]interface{}{"instance_type_id": req.InstanceTypeID}, "实例类型不存在")
			return 0, errors.New("实例类型不存在")
		}
		return 0, errors.New("查询实例类型失败: " + err.Error())
	}

	// 检查实例存在性
	instance, err := repo.GetByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("实例不存在")
		}
		return 0, errors.New("检查实例存在性失败: " + err.Error())
	}

	// 检查实例名称唯一性（排除自己）
	if existingInstance, err := repo.GetByName(req.Name); err == nil && existingInstance.ID != req.ID {
		return 0, errors.New("实例名称已存在")
	}

	// 更新实例
	instance.InstanceTypeID = req.InstanceTypeID
	instance.Name = req.Name
	instance.Address = req.Address
	instance.HttpsEnabled = req.HttpsEnabled
	instance.SkipSslVerify = req.SkipSslVerify
	instance.Status = req.Status
	instance.UpdatedAt = time.Now()

	if err := repo.Update(instance); err != nil {
		return 0, errors.New("更新实例失败: " + err.Error())
	}

	return int(instance.ID), nil
}

// handleAuthConfig 处理认证配置
func handleAuthConfig(repo *repositories.AuthConfigRepository, instanceID uint, authConfig *request.AuthConfigRequest, isAdd bool) error {
	// 无认证配置时删除现有配置
	if authConfig.AuthType == "" || authConfig.AuthType == "none" {
		if !isAdd {
			if err := repo.DeleteByInstanceID(instanceID); err != nil {
				return errors.New("删除认证配置失败: " + err.Error())
			}
			logs.Info(map[string]interface{}{"instance_id": instanceID}, "已删除认证配置")
			// 移除k8s客户端
			removeK8sClient(instanceID)
		}
		return nil
	}

	// 验证认证类型
	validTypes := map[string]bool{
		"basic":       true,
		"api_key":     true,
		"aws_iam":     true,
		"token":       true,
		"certificate": true,
		"kubeconfig":  true,
	}
	if !validTypes[authConfig.AuthType] {
		return errors.New("无效的认证类型")
	}

	// 检查现有配置
	existingConfigs, err := repo.GetByInstanceID(instanceID)
	if err != nil {
		return errors.New("查询认证配置失败: " + err.Error())
	}

	now := time.Time{}
	var instance *models.Instance
	var newAuthConfig *models.AuthConfig

	if len(existingConfigs) > 0 {
		// 更新现有配置
		existingConfig := &existingConfigs[0]
		existingConfig.AuthType = authConfig.AuthType
		existingConfig.ConfigKey = authConfig.ConfigKey
		existingConfig.ConfigValue = authConfig.ConfigValue
		existingConfig.IsEncrypted = authConfig.IsEncrypted
		existingConfig.UpdatedAt = now

		if err := repo.Update(existingConfig); err != nil {
			return errors.New("更新认证配置失败: " + err.Error())
		}
		logs.Info(map[string]interface{}{"instance_id": instanceID, "auth_type": authConfig.AuthType}, "认证配置更新成功")
		newAuthConfig = existingConfig
	} else {
		// 插入新配置
		// 获取实例名称
		instanceRepo := repositories.NewInstanceRepository()
		instance, err := instanceRepo.GetByID(instanceID)
		if err != nil {
			return errors.New("获取实例信息失败: " + err.Error())
		}

		newAuthConfig = &models.AuthConfig{
			ResourceType: "instance",
			ResourceID:   instanceID,
			ResourceName: instance.Name,
			AuthType:     authConfig.AuthType,
			ConfigKey:    authConfig.ConfigKey,
			ConfigValue:  authConfig.ConfigValue,
			IsEncrypted:  authConfig.IsEncrypted,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := repo.Create(newAuthConfig); err != nil {
			return errors.New("添加认证配置失败: " + err.Error())
		}
		logs.Info(map[string]interface{}{"instance_id": instanceID, "auth_type": authConfig.AuthType}, "认证配置添加成功")
	}

	// 如果是kubernetes实例，更新k8s客户端
	if instance == nil {
		instanceRepo := repositories.NewInstanceRepository()
		instance, err = instanceRepo.GetByID(instanceID)
		if err != nil {
			return errors.New("获取实例信息失败: " + err.Error())
		}
	}

	// 检查是否是kubernetes实例 - 手动查询实例类型
	var instanceType models.InstanceType
	if err := database.GORMDB.First(&instanceType, instance.InstanceTypeID).Error; err == nil && instanceType.TypeName == "kubernetes" {
		if authConfig.AuthType == "kubeconfig" {
			// 添加或更新k8s客户端
			if err := addOrUpdateK8sClient(instance, newAuthConfig); err != nil {
				logs.Warning(map[string]interface{}{
					"instance_id": instanceID,
					"error":       err.Error(),
				}, "添加或更新k8s客户端失败")
			}
		} else {
			// 移除k8s客户端
			removeK8sClient(instanceID)
		}
	}

	return nil
}
