package dal

// 认证类型常量
const (
	AuthTypeNone        = "none"
	AuthTypeBasic       = "basic"
	AuthTypeAPIKey      = "api_key"
	AuthTypeAWSIAM      = "aws_iam"
	AuthTypeToken       = "token"
	AuthTypeCertificate = "certificate"
	AuthTypeKubeconfig  = "kubeconfig"
)

// 资源类型常量
const (
	ResourceTypeInstance = "instance"
	ResourceTypeCluster  = "cluster"
)

// 认证配置状态常量
const (
	AuthConfigStatusActive   = "active"
	AuthConfigStatusInactive = "inactive"
)
