package request

// InstanceRequest 集群请求参数结构体
type InstanceRequest struct {
	ID             uint              `json:"id,omitempty"`                        // 更新时需要（使用uint匹配UNSIGNED INT）
	InstanceTypeID uint              `json:"instance_type_id" binding:"required"` // 实例类型ID（使用uint匹配UNSIGNED INT）
	Name           string            `json:"name" binding:"required"`             // 实例名称
	Address        string            `json:"address" binding:"required"`          // 实例地址
	HttpsEnabled   bool              `json:"https_enabled,omitempty"`             // 是否启用HTTPS
	SkipSslVerify  bool              `json:"skip_ssl_verify,omitempty"`           // 是否跳过SSL证书验证
	Status         string            `json:"status,omitempty"`                    // 实例状态
	AuthConfig     AuthConfigRequest `json:"auth_configs,omitempty"`              // 认证配置
}

// AuthConfigRequest 认证配置请求参数结构体
type AuthConfigRequest struct {
	AuthType    string `json:"auth_type" binding:"required"`  // 认证类型
	ConfigKey   string `json:"config_key" binding:"required"` // 配置键名
	ConfigValue string `json:"config_value,omitempty"`        // 配置值
	IsEncrypted bool   `json:"is_encrypted,omitempty"`        // 是否加密
}

// InstanceListRequest 集群列表查询请求参数结构体
type InstanceListRequest struct {
	Page     int    `form:"page" json:"page" binding:"min=1"`                                     // 页码，从1开始
	PageSize int    `form:"page_size" json:"page_size" binding:"min=1,max=100"`                   // 每页记录数，最大100
	Status   string `form:"status" json:"status" binding:"omitempty,oneof=active inactive error"` // 按状态筛选
	TypeName string `form:"type_name" json:"type_name" binding:"omitempty"`                       // 按实例类型筛选
	Keyword  string `form:"keyword" json:"keyword" binding:"omitempty,max=100"`                   // 关键词搜索（名称或地址），最大100字符
}

// ConnectionTestRequest 连接测试请求参数结构体
type ConnectionTestRequest struct {
	InstanceID uint `json:"instance_id" binding:"required"` // 实例ID（使用uint匹配UNSIGNED INT）
}
