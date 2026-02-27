package mapper

import (
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/dal/response"

	"gorm.io/gorm"
)

// SysUserMapper 系统用户数据访问
type SysUserMapper struct {
	DB *gorm.DB
}

func NewSysUserMapper(db *gorm.DB) *SysUserMapper {
	return &SysUserMapper{DB: db}
}

// PageResult 分页结果
type PageResult struct {
	Total int64
	List  interface{}
}

// ListPage 分页查询用户列表
func (m *SysUserMapper) ListPage(page, pageSize int, username, nickname string, deptID uint64, status *uint8) (int64, []response.UserVO, error) {
	var total int64
	var users []struct {
		model.SystemUser
		DeptName string `gorm:"column:dept_name"`
	}

	db := m.DB.Table("system_users u").
		Select("u.*, d.name as dept_name").
		Joins("LEFT JOIN sys_department d ON u.dept_id = d.id AND d.deleted_at IS NULL").
		Where("u.deleted_at IS NULL")

	if username != "" {
		db = db.Where("u.username LIKE ?", "%"+username+"%")
	}
	if nickname != "" {
		db = db.Where("u.nickname LIKE ?", "%"+nickname+"%")
	}
	if deptID > 0 {
		db = db.Where("u.dept_id = ?", deptID)
	}
	if status != nil {
		db = db.Where("u.status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("u.id DESC").Scan(&users).Error; err != nil {
		return 0, nil, err
	}

	vos := make([]response.UserVO, 0, len(users))
	for _, u := range users {
		nickname := ""
		if u.Nickname != nil {
			nickname = *u.Nickname
		}
		vo := response.UserVO{
			ID:       uint64(u.ID),
			Username: u.Username,
			Nickname: nickname,
			Status:   uint8(u.Status),
			DeptName: u.DeptName,
		}
		if u.DeptID != nil {
			vo.DeptID = *u.DeptID
		}
		if u.Email != nil {
			vo.Email = *u.Email
		}
		if u.Phone != nil {
			vo.Phone = *u.Phone
		}
		if u.Avatar != nil {
			vo.Avatar = *u.Avatar
		}
		vo.CreatedAt = u.CreatedAt
		vos = append(vos, vo)
	}
	return total, vos, nil
}

// GetByID 根据 ID 查询
func (m *SysUserMapper) GetByID(id uint64) (*model.SystemUser, error) {
	var u model.SystemUser
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&u).Error
	return &u, err
}

// GetUserRoles 获取用户角色列表
func (m *SysUserMapper) GetUserRoles(userID uint64) ([]response.RoleSimple, error) {
	var roles []response.RoleSimple
	err := m.DB.Table("sys_role r").
		Select("r.id, r.name, r.code").
		Joins("JOIN sys_user_role ur ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.deleted_at IS NULL AND r.status = 1", userID).
		Scan(&roles).Error
	return roles, err
}

// GetUserPositions 获取用户岗位列表
func (m *SysUserMapper) GetUserPositions(userID uint64) ([]response.PosSimple, error) {
	var positions []response.PosSimple
	err := m.DB.Table("sys_position p").
		Select("p.id, p.name, p.code").
		Joins("JOIN sys_user_position up ON up.position_id = p.id").
		Where("up.user_id = ? AND p.deleted_at IS NULL AND p.status = 1", userID).
		Scan(&positions).Error
	return positions, err
}

// Create 新建用户
func (m *SysUserMapper) Create(u *model.SystemUser) error {
	return m.DB.Create(u).Error
}

// Update 更新用户
func (m *SysUserMapper) Update(u *model.SystemUser) error {
	return m.DB.Save(u).Error
}

// UpdateFields 更新指定字段
func (m *SysUserMapper) UpdateFields(id uint64, fields map[string]interface{}) error {
	return m.DB.Model(&model.SystemUser{}).Where("id = ?", id).Updates(fields).Error
}

// SoftDelete 软删除
func (m *SysUserMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.SystemUser{}).Error
}

// ExistsByUsername 检查用户名是否存在
func (m *SysUserMapper) ExistsByUsername(username string, excludeID uint64) (bool, error) {
	var count int64
	db := m.DB.Model(&model.SystemUser{}).Where("username = ? AND deleted_at IS NULL", username)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	err := db.Count(&count).Error
	return count > 0, err
}

// AssignRoles 设置用户角色（先删后插）
func (m *SysUserMapper) AssignRoles(userID uint64, roleIDs []uint64) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.SysUserRole{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			if err := tx.Create(&model.SysUserRole{UserID: userID, RoleID: rid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AssignPositions 设置用户岗位（先删后插）
func (m *SysUserMapper) AssignPositions(userID uint64, positionIDs []uint64) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.SysUserPosition{}).Error; err != nil {
			return err
		}
		for _, pid := range positionIDs {
			if err := tx.Create(&model.SysUserPosition{UserID: userID, PositionID: pid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetUserPerms 获取用户权限标识列表（通过角色→菜单）
func (m *SysUserMapper) GetUserPerms(userID uint64) ([]string, error) {
	var perms []string
	err := m.DB.Table("sys_menu mn").
		Select("DISTINCT mn.perm").
		Joins("JOIN sys_role_menu rm ON rm.menu_id = mn.id").
		Joins("JOIN sys_user_role ur ON ur.role_id = rm.role_id").
		Where("ur.user_id = ? AND mn.perm IS NOT NULL AND mn.perm != '' AND mn.status = 1", userID).
		Pluck("mn.perm", &perms).Error
	return perms, err
}

// GetUserRoleCodes 获取用户角色编码列表
func (m *SysUserMapper) GetUserRoleCodes(userID uint64) ([]string, error) {
	var codes []string
	err := m.DB.Table("sys_role r").
		Select("r.code").
		Joins("JOIN sys_user_role ur ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.deleted_at IS NULL AND r.status = 1", userID).
		Pluck("r.code", &codes).Error
	return codes, err
}
