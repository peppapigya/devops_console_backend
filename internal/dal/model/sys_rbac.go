package model

import (
	"time"

	"gorm.io/gorm"
)

// ===================== 部门 =====================

const TableNameSysDepartment = "sys_department"

// SysDepartment 系统部门
type SysDepartment struct {
	ID        uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	ParentID  uint64         `gorm:"column:parent_id;type:bigint unsigned;not null;default:0;comment:父部门ID" json:"parentId"`
	Name      string         `gorm:"column:name;type:varchar(100);not null;comment:部门名称" json:"name"`
	Sort      int            `gorm:"column:sort;not null;default:0;comment:显示顺序" json:"sort"`
	Leader    *string        `gorm:"column:leader;type:varchar(50);comment:负责人" json:"leader"`
	Phone     *string        `gorm:"column:phone;type:varchar(20);comment:联系电话" json:"phone"`
	Email     *string        `gorm:"column:email;type:varchar(100);comment:联系邮箱" json:"email"`
	Status    uint8          `gorm:"column:status;type:tinyint unsigned;not null;default:1;comment:状态1启用0停用" json:"status"`
	Remark    *string        `gorm:"column:remark;type:text;comment:备注" json:"remark"`
	CreatedAt *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"deletedAt"`

	// 非数据库字段，用于树形结构构建
	Children []*SysDepartment `gorm:"-" json:"children,omitempty"`
}

func (*SysDepartment) TableName() string { return TableNameSysDepartment }

// ===================== 岗位 =====================

const TableNameSysPosition = "sys_position"

// SysPosition 系统岗位
type SysPosition struct {
	ID        uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"column:name;type:varchar(100);not null;comment:岗位名称" json:"name"`
	Code      string         `gorm:"column:code;type:varchar(100);not null;uniqueIndex;comment:岗位编码" json:"code"`
	Sort      int            `gorm:"column:sort;not null;default:0;comment:显示顺序" json:"sort"`
	Status    uint8          `gorm:"column:status;type:tinyint unsigned;not null;default:1;comment:状态1启用0停用" json:"status"`
	Remark    *string        `gorm:"column:remark;type:text;comment:备注" json:"remark"`
	CreatedAt *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"deletedAt"`
}

func (*SysPosition) TableName() string { return TableNameSysPosition }

// SysUserPosition 用户-岗位关联
type SysUserPosition struct {
	UserID     uint64 `gorm:"column:user_id;primaryKey" json:"userId"`
	PositionID uint64 `gorm:"column:position_id;primaryKey" json:"positionId"`
}

func (*SysUserPosition) TableName() string { return "sys_user_position" }

// ===================== 角色 =====================

const TableNameSysRole = "sys_role"

// SysRole 系统角色
type SysRole struct {
	ID        uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"column:name;type:varchar(100);not null;comment:角色名称" json:"name"`
	Code      string         `gorm:"column:code;type:varchar(100);not null;uniqueIndex;comment:角色编码" json:"code"`
	Sort      int            `gorm:"column:sort;not null;default:0;comment:显示顺序" json:"sort"`
	Status    uint8          `gorm:"column:status;type:tinyint unsigned;not null;default:1;comment:状态1启用0停用" json:"status"`
	Remark    *string        `gorm:"column:remark;type:text;comment:备注" json:"remark"`
	CreatedAt *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"deletedAt"`
}

func (*SysRole) TableName() string { return TableNameSysRole }

// SysUserRole 用户-角色关联
type SysUserRole struct {
	UserID uint64 `gorm:"column:user_id;primaryKey" json:"userId"`
	RoleID uint64 `gorm:"column:role_id;primaryKey" json:"roleId"`
}

func (*SysUserRole) TableName() string { return "sys_user_role" }

// ===================== 菜单 =====================

const TableNameSysMenu = "sys_menu"

// SysMenu 系统菜单
type SysMenu struct {
	ID        uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	ParentID  uint64         `gorm:"column:parent_id;type:bigint unsigned;not null;default:0;index;comment:父菜单ID" json:"parentId"`
	Name      string         `gorm:"column:name;type:varchar(100);not null;comment:菜单名称" json:"name"`
	Type      int8           `gorm:"column:type;not null;default:1;comment:类型1目录2菜单3按钮" json:"type"`
	Path      *string        `gorm:"column:path;type:varchar(200);comment:前端路由" json:"path"`
	Component *string        `gorm:"column:component;type:varchar(200);comment:组件路径" json:"component"`
	Icon      *string        `gorm:"column:icon;type:varchar(50);comment:图标" json:"icon"`
	Perm      *string        `gorm:"column:perm;type:varchar(200);comment:权限标识" json:"perm"`
	Sort      int            `gorm:"column:sort;not null;default:0;comment:显示顺序" json:"sort"`
	Visible   int8           `gorm:"column:visible;not null;default:1;comment:是否显示" json:"visible"`
	Status    int8           `gorm:"column:status;not null;default:1;comment:状态" json:"status"`
	CreatedAt *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"-"`

	// 非数据库字段，用于树形结构构建
	Children []*SysMenu `gorm:"-" json:"children,omitempty"`
}

func (*SysMenu) TableName() string { return TableNameSysMenu }

// SysRoleMenu 角色-菜单关联
type SysRoleMenu struct {
	RoleID uint64 `gorm:"column:role_id;primaryKey" json:"roleId"`
	MenuID uint64 `gorm:"column:menu_id;primaryKey" json:"menuId"`
}

func (*SysRoleMenu) TableName() string { return "sys_role_menu" }
