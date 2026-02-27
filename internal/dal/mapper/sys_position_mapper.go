package mapper

import (
	"devops-console-backend/internal/dal/model"

	"gorm.io/gorm"
)

// SysPositionMapper 岗位数据访问
type SysPositionMapper struct {
	DB *gorm.DB
}

func NewSysPositionMapper(db *gorm.DB) *SysPositionMapper {
	return &SysPositionMapper{DB: db}
}

// ListPage 分页查询岗位
func (m *SysPositionMapper) ListPage(page, pageSize int, name, code string, status *uint8) (int64, []model.SysPosition, error) {
	var total int64
	var list []model.SysPosition
	db := m.DB.Model(&model.SysPosition{}).Where("deleted_at IS NULL")
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

// ListAll 查询全部启用岗位（供用户编辑下拉使用）
func (m *SysPositionMapper) ListAll() ([]model.SysPosition, error) {
	var list []model.SysPosition
	err := m.DB.Where("deleted_at IS NULL AND status = 1").Order("sort ASC").Find(&list).Error
	return list, err
}

// GetByID 根据 ID 查询
func (m *SysPositionMapper) GetByID(id uint64) (*model.SysPosition, error) {
	var pos model.SysPosition
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&pos).Error
	return &pos, err
}

// Create 新建岗位
func (m *SysPositionMapper) Create(pos *model.SysPosition) error {
	return m.DB.Create(pos).Error
}

// Update 更新岗位
func (m *SysPositionMapper) Update(pos *model.SysPosition) error {
	return m.DB.Save(pos).Error
}

// SoftDelete 软删除
func (m *SysPositionMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.SysPosition{}).Error
}

// ExistsByCode 检查编码是否重复
func (m *SysPositionMapper) ExistsByCode(code string, excludeID uint64) (bool, error) {
	var count int64
	db := m.DB.Model(&model.SysPosition{}).Where("code = ? AND deleted_at IS NULL", code)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	err := db.Count(&count).Error
	return count > 0, err
}
