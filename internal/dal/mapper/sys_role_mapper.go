package mapper

import (
	"devops-console-backend/internal/dal/model"

	"gorm.io/gorm"
)

// SysRoleMapper 角色数据访问
type SysRoleMapper struct {
	DB *gorm.DB
}

func NewSysRoleMapper(db *gorm.DB) *SysRoleMapper {
	return &SysRoleMapper{DB: db}
}

// ListPage 分页查询角色
func (m *SysRoleMapper) ListPage(page, pageSize int, name, code string, status *uint8) (int64, []model.SysRole, error) {
	var total int64
	var list []model.SysRole
	db := m.DB.Model(&model.SysRole{}).Where("deleted_at IS NULL")
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		db = db.Where("code LIKE ?", "%"+code+"%")
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Order("sort ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

// ListAll 查询全部启用角色（供用户编辑下拉使用）
func (m *SysRoleMapper) ListAll() ([]model.SysRole, error) {
	var list []model.SysRole
	err := m.DB.Where("deleted_at IS NULL AND status = 1").Order("sort ASC").Find(&list).Error
	return list, err
}

// GetByID 根据 ID 查询
func (m *SysRoleMapper) GetByID(id uint64) (*model.SysRole, error) {
	var role model.SysRole
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error
	return &role, err
}

// Create 新建角色
func (m *SysRoleMapper) Create(role *model.SysRole) error {
	return m.DB.Create(role).Error
}

// Update 更新角色
func (m *SysRoleMapper) Update(role *model.SysRole) error {
	return m.DB.Save(role).Error
}

// SoftDelete 软删除
func (m *SysRoleMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.SysRole{}).Error
}

// GetRoleMenuIDs 获取角色已分配的菜单 ID 列表
func (m *SysRoleMapper) GetRoleMenuIDs(roleID uint64) ([]uint64, error) {
	var ids []uint64
	err := m.DB.Table("sys_role_menu").
		Where("role_id = ?", roleID).
		Pluck("menu_id", &ids).Error
	return ids, err
}

// AssignMenus 给角色分配菜单（先删后插）
func (m *SysRoleMapper) AssignMenus(roleID uint64, menuIDs []uint64) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.SysRoleMenu{}).Error; err != nil {
			return err
		}
		for _, mid := range menuIDs {
			if err := tx.Create(&model.SysRoleMenu{RoleID: roleID, MenuID: mid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ExistsByCode 检查编码是否重复
func (m *SysRoleMapper) ExistsByCode(code string, excludeID uint64) (bool, error) {
	var count int64
	db := m.DB.Model(&model.SysRole{}).Where("code = ? AND deleted_at IS NULL", code)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	err := db.Count(&count).Error
	return count > 0, err
}
