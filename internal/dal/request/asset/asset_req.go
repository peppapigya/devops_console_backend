package asset

// ==================== 主机分组 ====================

type HostGroupCreateReq struct {
	ParentID uint64 `json:"parentId"`
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Remark   string `json:"remark"`
}

type HostGroupUpdateReq struct {
	ParentID uint64 `json:"parentId"`
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Remark   string `json:"remark"`
}

// ==================== 主机管理 ====================

type HostPageReq struct {
	Page          int    `form:"page" binding:"required,min=1"`
	PageSize      int    `form:"pageSize" binding:"required,min=1,max=200"`
	GroupID       uint64 `form:"groupId"`
	Name          string `form:"name"`
	IP            string `form:"ip"`
	Status        string `form:"status"` // online/offline/unknown
	CloudProvider string `form:"cloudProvider"`
}

type HostCreateReq struct {
	GroupID       uint64   `json:"groupId"`
	Name          string   `json:"name" binding:"required,min=1,max=200"`
	IP            string   `json:"ip" binding:"required"`
	Port          uint16   `json:"port" binding:"required,min=1,max=65535"`
	OsType        string   `json:"osType"`
	CloudProvider string   `json:"cloudProvider"`
	Username      string   `json:"username" binding:"required"`
	AuthType      string   `json:"authType" binding:"required,oneof=password key"`
	Password      string   `json:"password"`
	PrivateKey    string   `json:"privateKey"`
	Tags          []string `json:"tags"`
	Remark        string   `json:"remark"`
}

type HostUpdateReq struct {
	GroupID       uint64   `json:"groupId"`
	Name          string   `json:"name" binding:"required,min=1,max=200"`
	IP            string   `json:"ip" binding:"required"`
	Port          uint16   `json:"port" binding:"required,min=1,max=65535"`
	OsType        string   `json:"osType"`
	CloudProvider string   `json:"cloudProvider"`
	Username      string   `json:"username" binding:"required"`
	AuthType      string   `json:"authType"`
	Password      string   `json:"password"`
	PrivateKey    string   `json:"privateKey"`
	Tags          []string `json:"tags"`
	Remark        string   `json:"remark"`
}

type BatchDeleteReq struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}
