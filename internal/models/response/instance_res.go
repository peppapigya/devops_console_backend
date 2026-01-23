package response

import (
	"devops-console-backend/internal/models"
	"time"
)

// InstanceListResponse 集群列表响应结构体
type InstanceListResponse struct {
	Page       int                    `json:"page"`        // 当前页码
	PageSize   int                    `json:"page_size"`   // 每页记录数
	Total      int64                  `json:"total"`       // 总记录数
	TotalPages int                    `json:"total_pages"` // 总页数
	Data       []InstanceItemResponse `json:"data"`        // 数据列表
}

// InstanceItemResponse 集群项目响应结构体
type InstanceItemResponse struct {
	ID           uint      `json:"id"`            // 实例ID（使用uint匹配UNSIGNED INT）
	InstanceType string    `json:"instance_type"` // 实例类型名称
	Name         string    `json:"name"`          // 实例名称
	Address      string    `json:"address"`       // 实例地址
	HttpsEnabled bool      `json:"https_enabled"` // 是否启用HTTPS
	Status       string    `json:"status"`        // 实例状态
	AuthConfigs  string    `json:"auth_configs"`  // 认证配置（键值对拼接的字符串）
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
}

// NewInstanceListResponse 创建集群列表响应
func NewInstanceListResponse(page, pageSize int, total int64, instances []models.Instance, instanceTypes map[uint64]string) *InstanceListResponse {
	// 处理边界情况：如果pageSize为0，避免除零错误
	if pageSize <= 0 {
		pageSize = 10
	}

	// 计算总页数，使用更高效的方式
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// 预分配切片容量以提高性能
	data := make([]InstanceItemResponse, 0, len(instances))

	// 使用范围遍历构建响应数据
	for _, instance := range instances {
		// 安全获取实例类型，如果不存在则使用默认值
		instanceType, ok := instanceTypes[uint64(instance.InstanceTypeID)]
		if !ok {
			instanceType = "unknown"
		}

		data = append(data, InstanceItemResponse{
			ID:           instance.ID,
			InstanceType: instanceType,
			Name:         instance.Name,
			Address:      instance.Address,
			HttpsEnabled: instance.HttpsEnabled,
			Status:       instance.Status,
			AuthConfigs:  "", // 可根据需要从数据库查询认证配置
			CreatedAt:    instance.CreatedAt,
			UpdatedAt:    instance.UpdatedAt,
		})
	}

	return &InstanceListResponse{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		Data:       data,
	}
}
