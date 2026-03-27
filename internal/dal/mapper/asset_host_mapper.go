package mapper

import (
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils/aes"
	"errors"

	"gorm.io/gorm"
)

// ==================== AssetHostGroupMapper ====================

type AssetHostGroupMapper struct {
	DB *gorm.DB
}

func NewAssetHostGroupMapper(db *gorm.DB) *AssetHostGroupMapper {
	return &AssetHostGroupMapper{DB: db}
}

func (m *AssetHostGroupMapper) ListAll() ([]model.AssetHostGroup, error) {
	var list []model.AssetHostGroup
	err := m.DB.Where("deleted_at IS NULL").Order("id ASC").Find(&list).Error
	return list, err
}

func (m *AssetHostGroupMapper) GetByID(id uint64) (*model.AssetHostGroup, error) {
	var g model.AssetHostGroup
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&g).Error
	return &g, err
}

func (m *AssetHostGroupMapper) Create(g *model.AssetHostGroup) error {
	return m.DB.Create(g).Error
}

func (m *AssetHostGroupMapper) Update(g *model.AssetHostGroup) error {
	return m.DB.Save(g).Error
}

func (m *AssetHostGroupMapper) SoftDelete(id uint64) error {
	var count int64
	m.DB.Model(&model.AssetHostGroup{}).Where("parent_id = ? AND deleted_at IS NULL", id).Count(&count)
	if count > 0 {
		return errors.New("该分组下存在子分组，请先删除子分组")
	}
	return m.DB.Where("id = ?", id).Delete(&model.AssetHostGroup{}).Error
}

// BuildHostGroupTree 构建分组树
func BuildHostGroupTree(groups []model.AssetHostGroup, parentID uint64) []*model.AssetHostGroup {
	result := make([]*model.AssetHostGroup, 0)
	for i := range groups {
		if groups[i].ParentID == parentID {
			g := groups[i]
			g.Children = BuildHostGroupTree(groups, g.ID)
			result = append(result, &g)
		}
	}
	return result
}

// ==================== AssetHostMapper ====================

type AssetHostMapper struct {
	DB *gorm.DB
}

func NewAssetHostMapper(db *gorm.DB) *AssetHostMapper {
	return &AssetHostMapper{DB: db}
}

func (m *AssetHostMapper) ListPage(page, pageSize int, groupID uint64, name, ip, status string) (int64, []model.AssetHost, error) {
	var total int64
	var list []model.AssetHost
	db := m.DB.Model(&model.AssetHost{}).Where("deleted_at IS NULL")
	if groupID > 0 {
		db = db.Where("group_id = ?", groupID)
	}
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if ip != "" {
		db = db.Where("ip LIKE ?", "%"+ip+"%")
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return total, list, err
}

func (m *AssetHostMapper) GetByID(id uint64) (*model.AssetHost, error) {
	var h model.AssetHost
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&h).Error
	return &h, err
}

func (m *AssetHostMapper) Create(h *model.AssetHost) error {
	return m.DB.Create(h).Error
}

func (m *AssetHostMapper) Update(h *model.AssetHost) error {
	return m.DB.Save(h).Error
}

func (m *AssetHostMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.AssetHost{}).Error
}

func (m *AssetHostMapper) BatchDelete(ids []uint64) error {
	return m.DB.Where("id IN ?", ids).Delete(&model.AssetHost{}).Error
}

// Stats 统计各类主机数量
func (m *AssetHostMapper) Stats(groupID uint64) (map[string]interface{}, error) {
	db := m.DB.Model(&model.AssetHost{}).Where("deleted_at IS NULL")
	if groupID > 0 {
		db = db.Where("group_id = ?", groupID)
	}
	result := map[string]interface{}{
		"total": int64(0), "linux": int64(0), "windows": int64(0),
		"offline": int64(0), "aliyun": int64(0), "huawei": int64(0), "aws": int64(0),
	}
	var total int64
	db.Count(&total)
	result["total"] = total

	type countRow struct {
		Key   string
		Count int64
	}
	var osRows []countRow
	m.DB.Model(&model.AssetHost{}).
		Where("deleted_at IS NULL").
		Select("os_type as `key`, count(*) as `count`").
		Group("os_type").Scan(&osRows)
	for _, r := range osRows {
		result[r.Key] = r.Count
	}

	var statusRows []countRow
	m.DB.Model(&model.AssetHost{}).
		Where("deleted_at IS NULL AND status = 'offline'").
		Select("'offline' as `key`, count(*) as `count`").Scan(&statusRows)
	if len(statusRows) > 0 {
		result["offline"] = statusRows[0].Count
	}

	var cloudRows []countRow
	m.DB.Model(&model.AssetHost{}).
		Where("deleted_at IS NULL AND cloud_provider IN ('aliyun','huawei','aws')").
		Select("cloud_provider as `key`, count(*) as `count`").
		Group("cloud_provider").Scan(&cloudRows)
	for _, r := range cloudRows {
		result[r.Key] = r.Count
	}
	return result, nil
}

// GetDecryptedPassword 获取解密后的密码
func (m *AssetHostMapper) GetDecryptedPassword(id uint64) (string, error) {
	host, err := m.GetByID(id)
	if err != nil {
		return "", err
	}
	if host.Password == nil || *host.Password == "" {
		return "", nil
	}
	return decryptPassword(*host.Password)
}

func decryptPassword(encryptedPassword string) (string, error) {
	key, err := configs.GetEncryptionKey()
	if err != nil {
		return "", err
	}
	if key == nil {
		return encryptedPassword, nil
	}
	return aes.AESDecrypt(key, encryptedPassword)
}
