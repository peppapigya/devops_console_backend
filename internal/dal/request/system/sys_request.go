package system

// ==================== 用户管理 ====================

// UserPageRequest 用户分页查询
type UserPageRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
	Username string `form:"username"`
	Nickname string `form:"nickname"`
	DeptID   uint64 `form:"deptId"`
	Status   *uint8 `form:"status"`
}

// UserCreateRequest 新建用户
type UserCreateRequest struct {
	Username    string   `json:"username" binding:"required,min=3,max=50"`
	Password    string   `json:"password" binding:"required,min=6,max=100"`
	Nickname    string   `json:"nickname"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	DeptID      uint64   `json:"deptId"`
	PositionIDs []uint64 `json:"positionIds"`
	RoleIDs     []uint64 `json:"roleIds"`
	Status      uint8    `json:"status"`
	Remark      string   `json:"remark"`
}

// UserUpdateRequest 更新用户
type UserUpdateRequest struct {
	Nickname    string   `json:"nickname"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	DeptID      uint64   `json:"deptId"`
	PositionIDs []uint64 `json:"positionIds"`
	RoleIDs     []uint64 `json:"roleIds"`
	Status      uint8    `json:"status"`
	Remark      string   `json:"remark"`
}

// UserStatusRequest 修改用户状态
type UserStatusRequest struct {
	Status uint8 `json:"status" binding:"oneof=0 1"`
}

// UserResetPwdRequest 重置密码
type UserResetPwdRequest struct {
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// ==================== 部门管理 ====================

// DeptCreateRequest 新建部门
type DeptCreateRequest struct {
	ParentID uint64 `json:"parentId"`
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Sort     int    `json:"sort"`
	Leader   string `json:"leader"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Status   uint8  `json:"status"`
	Remark   string `json:"remark"`
}

// DeptUpdateRequest 更新部门
type DeptUpdateRequest struct {
	ParentID uint64 `json:"parentId"`
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Sort     int    `json:"sort"`
	Leader   string `json:"leader"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Status   uint8  `json:"status"`
	Remark   string `json:"remark"`
}

// ==================== 岗位管理 ====================

// PositionPageRequest 岗位分页查询
type PositionPageRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
	Name     string `form:"name"`
	Code     string `form:"code"`
	Status   *uint8 `form:"status"`
}

// PositionCreateRequest 新建岗位
type PositionCreateRequest struct {
	Name   string `json:"name" binding:"required,min=1,max=100"`
	Code   string `json:"code" binding:"required,min=1,max=100"`
	Sort   int    `json:"sort"`
	Status uint8  `json:"status"`
	Remark string `json:"remark"`
}

// PositionUpdateRequest 更新岗位
type PositionUpdateRequest struct {
	Name   string `json:"name" binding:"required,min=1,max=100"`
	Code   string `json:"code" binding:"required,min=1,max=100"`
	Sort   int    `json:"sort"`
	Status uint8  `json:"status"`
	Remark string `json:"remark"`
}

// ==================== 角色管理 ====================

// RolePageRequest 角色分页查询
type RolePageRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
	Name     string `form:"name"`
	Code     string `form:"code"`
	Status   *uint8 `form:"status"`
}

// RoleCreateRequest 新建角色
type RoleCreateRequest struct {
	Name   string `json:"name" binding:"required,min=1,max=100"`
	Code   string `json:"code" binding:"required,min=1,max=100"`
	Sort   int    `json:"sort"`
	Status uint8  `json:"status"`
	Remark string `json:"remark"`
}

// RoleUpdateRequest 更新角色
type RoleUpdateRequest struct {
	Name   string `json:"name" binding:"required,min=1,max=100"`
	Code   string `json:"code" binding:"required,min=1,max=100"`
	Sort   int    `json:"sort"`
	Status uint8  `json:"status"`
	Remark string `json:"remark"`
}

// RoleMenuAssignRequest 给角色分配菜单
type RoleMenuAssignRequest struct {
	MenuIDs []uint64 `json:"menuIds" binding:"required"`
}

// ==================== 菜单管理 ====================

// MenuCreateRequest 新建菜单
type MenuCreateRequest struct {
	ParentID  uint64 `json:"parentId"`
	Name      string `json:"name" binding:"required,min=1,max=100"`
	Type      int8   `json:"type" binding:"required,oneof=1 2 3"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Icon      string `json:"icon"`
	Perm      string `json:"perm"`
	Sort      int    `json:"sort"`
	Visible   int8   `json:"visible"`
	Status    int8   `json:"status"`
}

// MenuUpdateRequest 更新菜单
type MenuUpdateRequest struct {
	ParentID  uint64 `json:"parentId"`
	Name      string `json:"name" binding:"required,min=1,max=100"`
	Type      int8   `json:"type" binding:"required,oneof=1 2 3"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Icon      string `json:"icon"`
	Perm      string `json:"perm"`
	Sort      int    `json:"sort"`
	Visible   int8   `json:"visible"`
	Status    int8   `json:"status"`
}

// ==================== 个人中心 ====================

// ProfileUpdateRequest 更新个人信息
type ProfileUpdateRequest struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
}

// ChangePasswordRequest 修改密码
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=100"`
}
