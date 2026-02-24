package services

import (
	"devops-console-backend/internal/dal"
	"devops-console-backend/internal/models/request"
	"devops-console-backend/pkg/configs"
	"errors"
)

// ListCustomMonitors 获取属于当前用户的自定义监控列表
func ListCustomMonitors(accountID int64, targetType string) ([]dal.CustomMonitor, error) {
	var monitors []dal.CustomMonitor
	query := configs.GORMDB.Where("account_id = ?", accountID)
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	err := query.Find(&monitors).Error
	return monitors, err
}

// CreateCustomMonitor 创建自定义图表
func CreateCustomMonitor(accountID int64, req request.CreateCustomMonitorRequest) (*dal.CustomMonitor, error) {
	monitor := &dal.CustomMonitor{
		AccountID:      uint(accountID),
		TargetType:     req.TargetType,
		Title:          req.Title,
		PromQLTemplate: req.PromQLTemplate,
		ChartType:      req.ChartType,
		UnitSuffix:     req.UnitSuffix,
		ColorTheme:     req.ColorTheme,
	}
	err := configs.GORMDB.Create(monitor).Error
	if err != nil {
		return nil, err
	}
	return monitor, nil
}

// UpdateCustomMonitor 修改自定义图表
func UpdateCustomMonitor(accountID uint, id uint, req request.UpdateCustomMonitorRequest) (*dal.CustomMonitor, error) {
	var monitor dal.CustomMonitor
	if err := configs.GORMDB.Where("id = ? AND account_id = ?", id, accountID).First(&monitor).Error; err != nil {
		return nil, errors.New("未找到记录或拒绝访问")
	}

	if req.Title != nil {
		monitor.Title = *req.Title
	}
	if req.PromQLTemplate != nil {
		monitor.PromQLTemplate = *req.PromQLTemplate
	}
	if req.ChartType != nil {
		monitor.ChartType = *req.ChartType
	}
	if req.UnitSuffix != nil {
		monitor.UnitSuffix = *req.UnitSuffix
	}
	if req.ColorTheme != nil {
		monitor.ColorTheme = *req.ColorTheme
	}

	err := configs.GORMDB.Save(&monitor).Error
	return &monitor, err
}

// DeleteCustomMonitor 删除自定义图表
func DeleteCustomMonitor(accountID uint, id uint) error {
	result := configs.GORMDB.Where("id = ? AND account_id = ?", id, accountID).Delete(&dal.CustomMonitor{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("未找到记录或拒绝访问")
	}
	return nil
}
