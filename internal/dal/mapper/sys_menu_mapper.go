package mapper

import (
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/dal/response"

	"gorm.io/gorm"
)

// SysMenuMapper 菜单数据访问
type SysMenuMapper struct {
	DB *gorm.DB
}

func NewSysMenuMapper(db *gorm.DB) *SysMenuMapper {
	return &SysMenuMapper{DB: db}
}

// ListAll 查询所有菜单（平铺，树形构建由调用方完成）
func (m *SysMenuMapper) ListAll() ([]model.SysMenu, error) {
	var list []model.SysMenu
	err := m.DB.Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

// ListByRoleID 查询角色拥有的所有菜单
func (m *SysMenuMapper) ListByRoleID(roleID uint64) ([]model.SysMenu, error) {
	var list []model.SysMenu
	err := m.DB.Table("sys_menu mn").
		Joins("JOIN sys_role_menu rm ON rm.menu_id = mn.id").
		Where("rm.role_id = ? AND mn.status = 1", roleID).
		Order("mn.sort ASC, mn.id ASC").
		Find(&list).Error
	return list, err
}

// ListByUserID 查询用户拥有的所有可见菜单（用于动态路由，只返回类型1和2）
func (m *SysMenuMapper) ListByUserID(userID uint64) ([]model.SysMenu, error) {
	var list []model.SysMenu
	err := m.DB.Table("sys_menu mn").
		Joins("JOIN sys_role_menu rm ON rm.menu_id = mn.id").
		Joins("JOIN sys_user_role ur ON ur.role_id = rm.role_id").
		Where("ur.user_id = ? AND mn.status = 1 AND mn.visible = 1 AND mn.type IN (1, 2)", userID).
		Group("mn.id").
		Order("mn.sort ASC, mn.id ASC").
		Find(&list).Error
	return list, err
}

// GetByID 根据 ID 查询
func (m *SysMenuMapper) GetByID(id uint64) (*model.SysMenu, error) {
	var menu model.SysMenu
	err := m.DB.Where("id = ?", id).First(&menu).Error
	return &menu, err
}

// Create 新建菜单
func (m *SysMenuMapper) Create(menu *model.SysMenu) error {
	return m.DB.Create(menu).Error
}

// Update 更新菜单
func (m *SysMenuMapper) Update(menu *model.SysMenu) error {
	return m.DB.Save(menu).Error
}

// Delete 删除菜单（硬删除，无软删除字段）
func (m *SysMenuMapper) Delete(id uint64) error {
	// 先检查是否有子菜单
	var childCount int64
	m.DB.Model(&model.SysMenu{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		return gorm.ErrRecordNotFound // 用自定义 error 更好，这里简化
	}
	return m.DB.Delete(&model.SysMenu{}, id).Error
}

// BuildMenuTree 将菜单列表转为树形 VO
func BuildMenuTree(menus []model.SysMenu, parentID uint64) []*response.MenuVO {
	result := make([]*response.MenuVO, 0)
	for _, m := range menus {
		if m.ParentID == parentID {
			vo := menuToVO(m)
			vo.Children = BuildMenuTree(menus, m.ID)
			result = append(result, vo)
		}
	}
	return result
}

func menuToVO(m model.SysMenu) *response.MenuVO {
	vo := &response.MenuVO{
		ID:       m.ID,
		ParentID: m.ParentID,
		Name:     m.Name,
		Type:     m.Type,
		Sort:     m.Sort,
		Visible:  m.Visible,
		Status:   m.Status,
	}
	if m.Path != nil {
		vo.Path = *m.Path
	}
	if m.Component != nil {
		vo.Component = *m.Component
	}
	if m.Icon != nil {
		vo.Icon = *m.Icon
	}
	if m.Perm != nil {
		vo.Perm = *m.Perm
	}
	return vo
}
