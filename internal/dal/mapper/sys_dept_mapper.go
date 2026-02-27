package mapper

import (
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/dal/response"
	"errors"

	"gorm.io/gorm"
)

// SysDeptMapper 部门数据访问
type SysDeptMapper struct {
	DB *gorm.DB
}

func NewSysDeptMapper(db *gorm.DB) *SysDeptMapper {
	return &SysDeptMapper{DB: db}
}

// ListAll 查询所有部门（构建树形在 Service/Controller 层做）
func (m *SysDeptMapper) ListAll(name string, status *uint8) ([]model.SysDepartment, error) {
	var depts []model.SysDepartment
	db := m.DB.Where("deleted_at IS NULL")
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	err := db.Order("sort ASC, id ASC").Find(&depts).Error
	return depts, err
}

// GetByID 根据 ID 查询
func (m *SysDeptMapper) GetByID(id uint64) (*model.SysDepartment, error) {
	var dept model.SysDepartment
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&dept).Error
	return &dept, err
}

// Create 新建部门
func (m *SysDeptMapper) Create(dept *model.SysDepartment) error {
	return m.DB.Create(dept).Error
}

// Update 更新部门
func (m *SysDeptMapper) Update(dept *model.SysDepartment) error {
	return m.DB.Save(dept).Error
}

// SoftDelete 软删除（先检查是否有子部门或用户）
func (m *SysDeptMapper) SoftDelete(id uint64) error {
	var childCount int64
	m.DB.Model(&model.SysDepartment{}).Where("parent_id = ? AND deleted_at IS NULL", id).Count(&childCount)
	if childCount > 0 {
		return errors.New("该部门下存在子部门，请先删除子部门")
	}
	var userCount int64
	m.DB.Table("system_users").Where("dept_id = ? AND deleted_at IS NULL", id).Count(&userCount)
	if userCount > 0 {
		return errors.New("该部门下存在用户，请先移除用户")
	}
	return m.DB.Where("id = ?", id).Delete(&model.SysDepartment{}).Error
}

// BuildTree 将部门列表转为树形结构
func BuildDeptTree(depts []model.SysDepartment, parentID uint64) []*response.DeptVO {
	result := make([]*response.DeptVO, 0)
	for _, d := range depts {
		if d.ParentID == parentID {
			dept := d
			vo := &response.DeptVO{
				ID:       d.ID,
				ParentID: d.ParentID,
				Name:     d.Name,
				Sort:     d.Sort,
				Status:   d.Status,
			}
			if d.Leader != nil {
				vo.Leader = *d.Leader
			}
			if d.Phone != nil {
				vo.Phone = *d.Phone
			}
			if d.Email != nil {
				vo.Email = *d.Email
			}
			if d.Remark != nil {
				vo.Remark = *d.Remark
			}
			_ = dept
			vo.Children = BuildDeptTree(depts, d.ID)
			result = append(result, vo)
		}
	}
	return result
}
