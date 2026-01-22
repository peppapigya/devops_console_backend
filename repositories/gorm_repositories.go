package repositories

import (
	"devops-console-backend/database"
	"devops-console-backend/models"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// InstanceTypeRepository 实例类型GORM操作
type InstanceTypeRepository struct{}

// NewInstanceTypeRepository 创建实例类型GORM操作实例
func NewInstanceTypeRepository() *InstanceTypeRepository {
	return &InstanceTypeRepository{}
}

// GetAll 获取所有实例类型
func (r *InstanceTypeRepository) GetAll() ([]models.InstanceType, error) {
	var types []models.InstanceType
	err := database.GORMDB.Order("type_name").Find(&types).Error
	return types, err
}

// GetByID 根据ID获取实例类型
func (r *InstanceTypeRepository) GetByID(id uint) (*models.InstanceType, error) {
	var instanceType models.InstanceType
	err := database.GORMDB.Where("id = ?", id).First(&instanceType).Error
	if err != nil {
		return nil, err
	}
	return &instanceType, nil
}

// GetByName 根据名称获取实例类型
func (r *InstanceTypeRepository) GetByName(typeName string) (*models.InstanceType, error) {
	var instanceType models.InstanceType
	err := database.GORMDB.Where("type_name = ?", typeName).First(&instanceType).Error
	if err != nil {
		return nil, err
	}
	return &instanceType, nil
}

// Create 创建实例类型
func (r *InstanceTypeRepository) Create(instanceType *models.InstanceType) error {
	return database.GORMDB.Create(instanceType).Error
}

// Update 更新实例类型
func (r *InstanceTypeRepository) Update(instanceType *models.InstanceType) error {
	return database.GORMDB.Save(instanceType).Error
}

// Delete 根据ID删除实例类型
func (r *InstanceTypeRepository) Delete(id uint) error {
	return database.GORMDB.Delete(&models.InstanceType{}, id).Error
}

// InstanceRepository 实例GORM操作
type InstanceRepository struct{}

// NewInstanceRepository 创建实例GORM操作实例
func NewInstanceRepository() *InstanceRepository {
	return &InstanceRepository{}
}

// GetAll 获取所有实例
func (r *InstanceRepository) GetAll() ([]models.Instance, error) {
	var instances []models.Instance
	err := database.GORMDB.Find(&instances).Error
	return instances, err
}

// GetByID 根据ID获取实例
func (r *InstanceRepository) GetByID(id uint) (*models.Instance, error) {
	var instance models.Instance
	err := database.GORMDB.Where("id = ?", id).First(&instance).Error
	return &instance, err
}

// GetByName 根据名称获取实例
func (r *InstanceRepository) GetByName(name string) (*models.Instance, error) {
	var instance models.Instance
	err := database.GORMDB.Where("name = ?", name).First(&instance).Error
	return &instance, err
}

// Create 创建实例
func (r *InstanceRepository) Create(instance *models.Instance) error {
	return database.GORMDB.Create(instance).Error
}

// Update 更新实例
func (r *InstanceRepository) Update(instance *models.Instance) error {
	return database.GORMDB.Save(instance).Error
}

// Delete 根据ID删除实例
func (r *InstanceRepository) Delete(id uint) error {
	return database.GORMDB.Transaction(func(tx *gorm.DB) error {
		// 删除关联的认证配置
		if err := tx.Where("resource_type = ? AND resource_id = ?", "instance", id).Delete(&models.AuthConfig{}).Error; err != nil {
			return err
		}
		// 删除关联的连接测试记录
		if err := tx.Where("resource_type = ? AND resource_id = ?", "instance", id).Delete(&models.ConnectionTest{}).Error; err != nil {
			return err
		}
		// 删除实例
		return tx.Delete(&models.Instance{}, id).Error
	})
}

// GetByTypeID 根据类型ID获取实例列表
func (r *InstanceRepository) GetByTypeID(typeID uint) ([]models.Instance, error) {
	var instances []models.Instance
	err := database.GORMDB.Where("instance_type_id = ?", typeID).Find(&instances).Error
	return instances, err
}

// GetByStatus 根据状态获取实例列表
func (r *InstanceRepository) GetByStatus(status string) ([]models.Instance, error) {
	var instances []models.Instance
	err := database.GORMDB.Where("status = ?", status).Find(&instances).Error
	return instances, err
}

// GetWithPagination 分页获取实例
func (r *InstanceRepository) GetWithPagination(offset, limit int, filters map[string]interface{}) ([]models.Instance, int64, error) {
	var instances []models.Instance
	var total int64

	query := database.GORMDB.Model(&models.Instance{})

	// 应用过滤条件
	if name, ok := filters["name"].(string); ok && name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if typeName, ok := filters["type_name"].(string); ok && typeName != "" {
		query = query.Joins("JOIN instance_types ON instances.instance_type_id = instance_types.id").
			Where("instance_types.type_name = ?", typeName)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset(offset).Limit(limit).Find(&instances).Error
	return instances, total, err
}

// ResourceRepository 统一资源GORM操作
type ResourceRepository struct{}

// NewResourceRepository 创建资源GORM操作实例
func NewResourceRepository() *ResourceRepository {
	return &ResourceRepository{}
}

// GetResourceDetails 获取资源详情列表
func (r *ResourceRepository) GetResourceDetails(filter *models.ResourceFilter) ([]models.ResourceDetail, error) {
	var details []models.ResourceDetail
	query := database.GORMDB.Table("resource_details")

	// 应用过滤条件
	if filter != nil {
		if filter.ResourceType != nil {
			query = query.Where("resource_type = ?", *filter.ResourceType)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.TypeName != nil {
			query = query.Where("type_name = ?", *filter.TypeName)
		}
	}

	err := query.Find(&details).Error
	return details, err
}

// GetResourceDetailByID 根据资源ID和类型获取详情
func (r *ResourceRepository) GetResourceDetailByID(resourceType string, resourceID int) (*models.ResourceDetail, error) {
	var detail models.ResourceDetail
	err := database.GORMDB.Table("resource_details").
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		First(&detail).Error
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// GetInstanceDetailByID 根据实例ID获取详情（兼容性方法）
func (r *ResourceRepository) GetInstanceDetailByID(instanceID int) (*models.ResourceDetail, error) {
	return r.GetResourceDetailByID(models.ResourceTypeInstance, instanceID)
}

// GetClusterDetailByID 根据集群ID获取详情（兼容性方法）
func (r *ResourceRepository) GetClusterDetailByID(clusterID int) (*models.ResourceDetail, error) {
	return r.GetResourceDetailByID(models.ResourceTypeCluster, clusterID)
}

// GetResourceList 获取资源列表（带分页）
func (r *ResourceRepository) GetResourceList(filter *models.ResourceFilter, offset, limit int) (*models.ResourceList, error) {
	query := database.GORMDB.Table("resource_details")

	// 应用过滤条件
	if filter != nil {
		if filter.ResourceType != nil {
			query = query.Where("resource_type = ?", *filter.ResourceType)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.TypeName != nil {
			query = query.Where("type_name = ?", *filter.TypeName)
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 获取分页数据
	var details []models.ResourceDetail
	err := query.Offset(offset).Limit(limit).Find(&details).Error
	if err != nil {
		return nil, err
	}

	return &models.ResourceList{
		Resources: details,
		Total:     int(total),
	}, nil
}

// AuthConfigRepository 统一认证配置GORM操作（兼容性保留）
type AuthConfigRepository struct{}

// NewAuthConfigRepository 创建认证配置GORM操作实例
func NewAuthConfigRepository() *AuthConfigRepository {
	return &AuthConfigRepository{}
}

// GetByResourceID 根据资源ID和类型获取认证配置
func (r *AuthConfigRepository) GetByResourceID(resourceType string, resourceID int) ([]models.AuthConfig, error) {
	var configs []models.AuthConfig
	err := database.GORMDB.Where("resource_type = ? AND resource_id = ? AND status = ?", resourceType, resourceID, models.AuthConfigStatusActive).Find(&configs).Error
	return configs, err
}

// GetByInstanceID 根据实例ID获取认证配置（兼容性方法）
func (r *AuthConfigRepository) GetByInstanceID(instanceID uint) ([]models.AuthConfig, error) {
	return r.GetByResourceID(models.ResourceTypeInstance, int(instanceID))
}

// GetByClusterID 根据集群ID获取认证配置
func (r *AuthConfigRepository) GetByClusterID(clusterID int) ([]models.AuthConfig, error) {
	return r.GetByResourceID(models.ResourceTypeCluster, clusterID)
}

// Create 创建认证配置
func (r *AuthConfigRepository) Create(authConfig *models.AuthConfig) error {
	return database.GORMDB.Create(authConfig).Error
}

// Update 更新认证配置
func (r *AuthConfigRepository) Update(authConfig *models.AuthConfig) error {
	return database.GORMDB.Save(authConfig).Error
}

// DeleteByResourceID 根据资源类型和ID删除认证配置
func (r *AuthConfigRepository) DeleteByResourceID(resourceType string, resourceID uint) error {
	return database.GORMDB.Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).Delete(&models.AuthConfig{}).Error
}

// DeleteByInstanceID 根据实例ID删除认证配置
func (r *AuthConfigRepository) DeleteByInstanceID(instanceID uint) error {
	return r.DeleteByResourceID(models.ResourceTypeInstance, instanceID)
}

// DeleteByClusterID 根据集群ID删除所有认证配置
func (r *AuthConfigRepository) DeleteByClusterID(clusterID uint) error {
	return r.DeleteByResourceID(models.ResourceTypeCluster, clusterID)
}

// GetByKey 根据资源ID、类型和配置键获取认证配置
func (r *AuthConfigRepository) GetByKey(resourceType string, resourceID uint, configKey string) (*models.AuthConfig, error) {
	var authConfig models.AuthConfig
	err := database.GORMDB.Where("resource_type = ? AND resource_id = ? AND config_key = ? AND status = ?", resourceType, resourceID, configKey, models.AuthConfigStatusActive).First(&authConfig).Error
	if err != nil {
		return nil, err
	}
	return &authConfig, nil
}

// GetClusterList 获取集群列表（兼容性方法）
func (r *AuthConfigRepository) GetClusterList() ([]models.ClusterSimpleResult, error) {
	var clusters []models.ClusterSimpleResult
	err := database.GORMDB.Table("resource_details").
		Select("resource_id as id, resource_name as cluster_name").
		Where("resource_type = ? AND status = ?", models.ResourceTypeCluster, models.AuthConfigStatusActive).
		Find(&clusters).Error
	return clusters, err
}

// GetResourceCount 获取资源数量统计
func (r *ResourceRepository) GetResourceCount(filter *models.ResourceFilter) (map[string]int64, error) {
	query := database.GORMDB.Table("resource_details")

	// 应用过滤条件
	if filter != nil {
		if filter.ResourceType != nil {
			query = query.Where("resource_type = ?", *filter.ResourceType)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.TypeName != nil {
			query = query.Where("type_name = ?", *filter.TypeName)
		}
	}

	var results []struct {
		ResourceType string `json:"resource_type"`
		Count        int64  `json:"count"`
	}

	err := query.Select("resource_type, COUNT(*) as count").
		Group("resource_type").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	countMap := make(map[string]int64)
	for _, result := range results {
		countMap[result.ResourceType] = result.Count
	}

	return countMap, nil
}

// ConnectionTestRepository 连接测试GORM操作
type ConnectionTestRepository struct{}

// NewConnectionTestRepository 创建连接测试GORM操作实例
func NewConnectionTestRepository() *ConnectionTestRepository {
	return &ConnectionTestRepository{}
}

// Create 创建连接测试记录
func (r *ConnectionTestRepository) Create(test *models.ConnectionTest) error {
	return database.GORMDB.Create(test).Error
}

// GetByInstanceID 根据实例ID获取连接测试记录
func (r *ConnectionTestRepository) GetByInstanceID(instanceID uint, limit int) ([]models.ConnectionTest, error) {
	var tests []models.ConnectionTest
	query := database.GORMDB.Where("resource_type = ? AND resource_id = ?", "instance", instanceID).Order("tested_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&tests).Error
	return tests, err
}

// GetByTimeRange 根据时间范围获取连接测试记录
func (r *ConnectionTestRepository) GetByTimeRange(startTime, endTime time.Time) ([]models.ConnectionTest, error) {
	var tests []models.ConnectionTest
	err := database.GORMDB.Where("tested_at >= ? AND tested_at < ?", startTime, endTime).
		Order("tested_at DESC").Find(&tests).Error
	return tests, err
}

// GetTodayStats 获取今日统计信息
func (r *ConnectionTestRepository) GetTodayStats() (map[string]interface{}, error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var totalTests int64
	if err := database.GORMDB.Model(&models.ConnectionTest{}).
		Where("tested_at >= ? AND tested_at < ?", today, tomorrow).
		Count(&totalTests).Error; err != nil {
		return nil, err
	}

	var instanceTests []struct {
		InstanceID int    `json:"instance_id"`
		Name       string `json:"name"`
		Count      int64  `json:"count"`
	}

	err := database.GORMDB.Table("connection_tests").
		Select("connection_tests.resource_id as instance_id, instances.name, COUNT(*) as count").
		Joins("JOIN instances ON connection_tests.resource_id = instances.id").
		Where("connection_tests.resource_type = ? AND connection_tests.tested_at >= ? AND connection_tests.tested_at < ?", "instance", today, tomorrow).
		Group("connection_tests.resource_id, instances.name").
		Scan(&instanceTests).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_tests":    totalTests,
		"instance_tests": instanceTests,
	}, nil
}

// DeleteOldRecords 删除旧的测试记录（保留最近30天）
func (r *ConnectionTestRepository) DeleteOldRecords() error {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return database.GORMDB.Where("tested_at < ?", thirtyDaysAgo).Delete(&models.ConnectionTest{}).Error
}

// InstanceDetailRepository 实例详情GORM操作
type InstanceDetailRepository struct{}

// NewInstanceDetailRepository 创建实例详情GORM操作实例
func NewInstanceDetailRepository() *InstanceDetailRepository {
	return &InstanceDetailRepository{}
}

// GetAll 获取所有实例详情
func (r *InstanceDetailRepository) GetAll() ([]models.ResourceDetail, error) {
	var details []models.ResourceDetail
	err := database.GORMDB.Find(&details).Error
	return details, err
}

// GetByID 根据ID获取资源详情
func (r *InstanceDetailRepository) GetByID(id uint) (*models.ResourceDetail, error) {
	var instance models.Instance
	err := database.GORMDB.Where("id = ?", id).First(&instance).Error
	if err != nil {
		return nil, err
	}

	// 手动查询实例类型
	var instanceType models.InstanceType
	database.GORMDB.First(&instanceType, instance.InstanceTypeID)

	// 转换为ResourceDetail格式
	detail := &models.ResourceDetail{
		ResourceType:    "instance",
		TypeName:        instanceType.TypeName,
		TypeDescription: instanceType.Description,
		ResourceID:      instance.ID,
		ResourceName:    instance.Name,
		Status:          instance.Status,
		CreatedAt:       instance.CreatedAt,
		UpdatedAt:       instance.UpdatedAt,
		AuthConfigs:     "", // 将在下面设置
		AuthTypeDesc:    "", // 将在下面设置
		Address:         &instance.Address,
		HttpsEnabled:    &instance.HttpsEnabled,
		SkipSslVerify:   &instance.SkipSslVerify,
	}

	// 构建认证配置JSON - 手动查询认证配置
	var authConfigs []models.AuthConfig
	if err := database.GORMDB.Where("resource_type = ? AND resource_id = ?", "instance", instance.ID).Find(&authConfigs).Error; err == nil && len(authConfigs) > 0 {
		authConfigMaps := make([]map[string]interface{}, 0)
		for _, authConfig := range authConfigs {
			config := map[string]interface{}{
				"config_key":   authConfig.ConfigKey,
				"config_value": authConfig.ConfigValue,
				"auth_type":    authConfig.AuthType,
				"is_encrypted": authConfig.IsEncrypted,
			}
			authConfigMaps = append(authConfigMaps, config)

			// 设置主要认证类型描述
			if detail.AuthTypeDesc == "" {
				switch authConfig.AuthType {
				case "kubeconfig":
					detail.AuthTypeDesc = "配置文件认证"
				case "token":
					detail.AuthTypeDesc = "令牌认证"
				case "basic":
					detail.AuthTypeDesc = "基础认证"
				case "api_key":
					detail.AuthTypeDesc = "API密钥认证"
				case "certificate":
					detail.AuthTypeDesc = "证书认证"
				case "aws_iam":
					detail.AuthTypeDesc = "AWS IAM认证"
				default:
					detail.AuthTypeDesc = "无认证"
				}
			}
		}

		// 转换为JSON字符串
		if authConfigJSON, err := json.Marshal(authConfigMaps); err == nil {
			detail.AuthConfigs = string(authConfigJSON)
		}
	}

	return detail, nil
}

// GetByTypeName 根据类型名称获取实例详情
func (r *InstanceDetailRepository) GetByTypeName(typeName string) ([]models.ResourceDetail, error) {
	var details []models.ResourceDetail
	err := database.GORMDB.Where("type_name = ?", typeName).Find(&details).Error
	return details, err
}

// AccountRepository 用户账号GORM操作
type AccountRepository struct{}

// NewAccountRepository 创建用户账号GORM操作实例
func NewAccountRepository() *AccountRepository {
	return &AccountRepository{}
}

// GetAll 获取所有用户账号
func (r *AccountRepository) GetAll() ([]models.Account, error) {
	var accounts []models.Account
	err := database.GORMDB.Find(&accounts).Error
	return accounts, err
}

// GetByID 根据ID获取账户
func (r *AccountRepository) GetByID(id uint) (*models.Account, error) {
	var account models.Account
	err := database.GORMDB.Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetByUserID 根据用户ID获取用户账号
func (r *AccountRepository) GetByUserID(userID string) (*models.Account, error) {
	var account models.Account
	err := database.GORMDB.Where("user_id = ?", userID).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// Create 创建用户账号
func (r *AccountRepository) Create(account *models.Account) error {
	return database.GORMDB.Create(account).Error
}

// Update 更新用户账号
func (r *AccountRepository) Update(account *models.Account) error {
	return database.GORMDB.Save(account).Error
}

// Delete 根据ID删除账户
func (r *AccountRepository) Delete(id uint) error {
	return database.GORMDB.Delete(&models.Account{}, id).Error
}
