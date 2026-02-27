package response

import "time"

// ==================== 用户响应 ====================

// UserVO 用户视图对象
type UserVO struct {
	ID        uint64       `json:"id"`
	Username  string       `json:"username"`
	Nickname  string       `json:"nickname"`
	Email     string       `json:"email"`
	Phone     string       `json:"phone"`
	Avatar    string       `json:"avatar"`
	DeptID    uint64       `json:"deptId"`
	DeptName  string       `json:"deptName"`
	Status    uint8        `json:"status"`
	Remark    string       `json:"remark"`
	Roles     []RoleSimple `json:"roles"`
	Positions []PosSimple  `json:"positions"`
	CreatedAt *time.Time   `json:"createdAt"`
}

// RoleSimple 角色简要信息
type RoleSimple struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// PosSimple 岗位简要信息
type PosSimple struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// UserPageVO 用户分页响应
type UserPageVO struct {
	Total int64    `json:"total"`
	List  []UserVO `json:"list"`
}

// ==================== 部门响应 ====================

// DeptVO 部门视图对象（树形）
type DeptVO struct {
	ID       uint64    `json:"id"`
	ParentID uint64    `json:"parentId"`
	Name     string    `json:"name"`
	Sort     int       `json:"sort"`
	Leader   string    `json:"leader"`
	Phone    string    `json:"phone"`
	Email    string    `json:"email"`
	Status   uint8     `json:"status"`
	Remark   string    `json:"remark"`
	Children []*DeptVO `json:"children,omitempty"`
}

// ==================== 岗位响应 ====================

// PositionVO 岗位视图对象
type PositionVO struct {
	ID        uint64     `json:"id"`
	Name      string     `json:"name"`
	Code      string     `json:"code"`
	Sort      int        `json:"sort"`
	Status    uint8      `json:"status"`
	Remark    string     `json:"remark"`
	CreatedAt *time.Time `json:"createdAt"`
}

// PositionPageVO 岗位分页响应
type PositionPageVO struct {
	Total int64        `json:"total"`
	List  []PositionVO `json:"list"`
}

// ==================== 角色响应 ====================

// RoleVO 角色视图对象
type RoleVO struct {
	ID        uint64     `json:"id"`
	Name      string     `json:"name"`
	Code      string     `json:"code"`
	Sort      int        `json:"sort"`
	Status    uint8      `json:"status"`
	Remark    string     `json:"remark"`
	CreatedAt *time.Time `json:"createdAt"`
}

// RolePageVO 角色分页响应
type RolePageVO struct {
	Total int64    `json:"total"`
	List  []RoleVO `json:"list"`
}

// ==================== 菜单响应 ====================

// MenuVO 菜单视图对象（树形）
type MenuVO struct {
	ID        uint64    `json:"id"`
	ParentID  uint64    `json:"parentId"`
	Name      string    `json:"name"`
	Type      int8      `json:"type"`
	Path      string    `json:"path"`
	Component string    `json:"component"`
	Icon      string    `json:"icon"`
	Perm      string    `json:"perm"`
	Sort      int       `json:"sort"`
	Visible   int8      `json:"visible"`
	Status    int8      `json:"status"`
	Children  []*MenuVO `json:"children,omitempty"`
}

// ==================== 当前用户信息（供前端动态路由使用）====================

// AuthInfoVO 当前登录用户信息
type AuthInfoVO struct {
	UserID   uint64    `json:"userId"`
	Username string    `json:"username"`
	Nickname string    `json:"nickname"`
	Avatar   string    `json:"avatar"`
	Roles    []string  `json:"roles"` // 角色编码列表，如 ["admin","operator"]
	Perms    []string  `json:"perms"` // 权限标识列表，如 ["system:user:list"]
	Menus    []*MenuVO `json:"menus"` // 菜单树（type=1或2，用于动态路由）
}
