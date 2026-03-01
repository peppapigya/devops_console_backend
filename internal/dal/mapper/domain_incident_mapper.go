package mapper

import (
	"devops-console-backend/internal/dal/model"

	"gorm.io/gorm"
)

// ==================== DomainMapper ====================

type DomainMapper struct {
	DB *gorm.DB
}

func NewDomainMapper(db *gorm.DB) *DomainMapper {
	return &DomainMapper{DB: db}
}

func (m *DomainMapper) Stats() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"total": int64(0), "normal": int64(0), "abnormal": int64(0), "expiringSoon": int64(0), "expired": int64(0),
	}
	var total int64
	m.DB.Model(&model.MonitorDomain{}).Where("deleted_at IS NULL").Count(&total)
	result["total"] = total

	var normal, abnormal, expiringSoon, expired int64
	m.DB.Model(&model.MonitorDomain{}).Where("deleted_at IS NULL AND status = 'normal'").Count(&normal)
	m.DB.Model(&model.MonitorDomain{}).Where("deleted_at IS NULL AND status = 'abnormal'").Count(&abnormal)
	m.DB.Model(&model.MonitorDomain{}).Where("deleted_at IS NULL AND ssl_days_left BETWEEN 1 AND 30").Count(&expiringSoon)
	m.DB.Model(&model.MonitorDomain{}).Where("deleted_at IS NULL AND ssl_days_left <= 0").Count(&expired)
	result["normal"] = normal
	result["abnormal"] = abnormal
	result["expiringSoon"] = expiringSoon
	result["expired"] = expired
	return result, nil
}

func (m *DomainMapper) ListPage(page, pageSize int, domain, status string, expireWithin int) (int64, []model.MonitorDomain, error) {
	var total int64
	var list []model.MonitorDomain
	db := m.DB.Model(&model.MonitorDomain{}).Where("deleted_at IS NULL")
	if domain != "" {
		db = db.Where("domain LIKE ?", "%"+domain+"%")
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if expireWithin > 0 {
		db = db.Where("ssl_days_left BETWEEN 1 AND ?", expireWithin)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return total, list, err
}

func (m *DomainMapper) GetByID(id uint64) (*model.MonitorDomain, error) {
	var d model.MonitorDomain
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&d).Error
	return &d, err
}

func (m *DomainMapper) Create(d *model.MonitorDomain) error {
	return m.DB.Create(d).Error
}

func (m *DomainMapper) Update(d *model.MonitorDomain) error {
	return m.DB.Save(d).Error
}

func (m *DomainMapper) UpdateFields(id uint64, fields map[string]interface{}) error {
	return m.DB.Model(&model.MonitorDomain{}).Where("id = ?", id).Updates(fields).Error
}

func (m *DomainMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.MonitorDomain{}).Error
}

// ==================== SslCertMapper ====================

type SslCertMapper struct {
	DB *gorm.DB
}

func NewSslCertMapper(db *gorm.DB) *SslCertMapper {
	return &SslCertMapper{DB: db}
}

func (m *SslCertMapper) ListPage(page, pageSize int, domain string, status *int8) (int64, []model.MonitorSslCert, error) {
	var total int64
	var list []model.MonitorSslCert
	db := m.DB.Model(&model.MonitorSslCert{}).Where("deleted_at IS NULL")
	if domain != "" {
		db = db.Where("domain LIKE ?", "%"+domain+"%")
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return total, list, err
}

func (m *SslCertMapper) GetByID(id uint64) (*model.MonitorSslCert, error) {
	var c model.MonitorSslCert
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&c).Error
	return &c, err
}

func (m *SslCertMapper) Create(c *model.MonitorSslCert) error {
	return m.DB.Create(c).Error
}

func (m *SslCertMapper) Update(c *model.MonitorSslCert) error {
	return m.DB.Save(c).Error
}

func (m *SslCertMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.MonitorSslCert{}).Error
}

// ==================== DnsProviderMapper ====================

type DnsProviderMapper struct {
	DB *gorm.DB
}

func NewDnsProviderMapper(db *gorm.DB) *DnsProviderMapper {
	return &DnsProviderMapper{DB: db}
}

func (m *DnsProviderMapper) ListPage(page, pageSize int, name, status string) (int64, []model.MonitorDnsProvider, error) {
	var total int64
	var list []model.MonitorDnsProvider
	db := m.DB.Model(&model.MonitorDnsProvider{}).Where("deleted_at IS NULL")
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
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

func (m *DnsProviderMapper) GetByID(id uint64) (*model.MonitorDnsProvider, error) {
	var p model.MonitorDnsProvider
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&p).Error
	return &p, err
}

func (m *DnsProviderMapper) Create(p *model.MonitorDnsProvider) error {
	return m.DB.Create(p).Error
}

func (m *DnsProviderMapper) Update(p *model.MonitorDnsProvider) error {
	return m.DB.Save(p).Error
}

func (m *DnsProviderMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.MonitorDnsProvider{}).Error
}

// ==================== IncidentMapper ====================

type IncidentMapper struct {
	DB *gorm.DB
}

func NewIncidentMapper(db *gorm.DB) *IncidentMapper {
	return &IncidentMapper{DB: db}
}

func (m *IncidentMapper) Stats() (map[string]interface{}, error) {
	result := map[string]interface{}{"total": int64(0), "pending": int64(0), "processing": int64(0), "done": int64(0)}
	var total, pending, processing, done int64
	m.DB.Model(&model.MonitorIncident{}).Where("deleted_at IS NULL").Count(&total)
	m.DB.Model(&model.MonitorIncident{}).Where("deleted_at IS NULL AND status = 'pending'").Count(&pending)
	m.DB.Model(&model.MonitorIncident{}).Where("deleted_at IS NULL AND status = 'processing'").Count(&processing)
	m.DB.Model(&model.MonitorIncident{}).Where("deleted_at IS NULL AND status = 'done'").Count(&done)
	result["total"] = total
	result["pending"] = pending
	result["processing"] = processing
	result["done"] = done
	return result, nil
}

func (m *IncidentMapper) ListPage(page, pageSize int, businessLine, level, status, dept string) (int64, []model.MonitorIncident, error) {
	var total int64
	var list []model.MonitorIncident
	db := m.DB.Model(&model.MonitorIncident{}).Where("deleted_at IS NULL")
	if businessLine != "" {
		db = db.Where("business_line = ?", businessLine)
	}
	if level != "" {
		db = db.Where("level = ?", level)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if dept != "" {
		db = db.Where("dept = ?", dept)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("alert_time DESC").Find(&list).Error
	return total, list, err
}

func (m *IncidentMapper) GetByID(id uint64) (*model.MonitorIncident, error) {
	var inc model.MonitorIncident
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&inc).Error
	return &inc, err
}

func (m *IncidentMapper) Create(inc *model.MonitorIncident) error {
	return m.DB.Create(inc).Error
}

func (m *IncidentMapper) Update(inc *model.MonitorIncident) error {
	return m.DB.Save(inc).Error
}

func (m *IncidentMapper) UpdateFields(id uint64, fields map[string]interface{}) error {
	return m.DB.Model(&model.MonitorIncident{}).Where("id = ?", id).Updates(fields).Error
}

func (m *IncidentMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.MonitorIncident{}).Error
}

func (m *IncidentMapper) ListBusinessLines() ([]string, error) {
	var lines []string
	err := m.DB.Model(&model.MonitorIncident{}).
		Where("deleted_at IS NULL").
		Select("DISTINCT business_line").
		Order("business_line ASC").
		Pluck("business_line", &lines).Error
	return lines, err
}
