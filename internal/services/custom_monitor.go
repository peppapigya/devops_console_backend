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

	// 如果查询成功且没有任何数据，则根据 targetType 自动初始化默认的图表模板
	if err == nil && len(monitors) == 0 {
		if targetType == "node" {
			seedNodeMonitors(accountID)
			query.Find(&monitors) // 重新查询
		} else if targetType == "pod" {
			seedPodMonitors(accountID)
			query.Find(&monitors) // 重新查询
		}
	}

	return monitors, err
}

func seedNodeMonitors(accountID int64) {
	defaults := []dal.CustomMonitor{
		{AccountID: uint(accountID), TargetType: "node", Title: "CPU 使用率", PromQLTemplate: `min(100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle", instance="{{nodeName}}"}[5m])) * 100))`, ChartType: "line", UnitSuffix: "%", ColorTheme: "#f56c6c"},
		{AccountID: uint(accountID), TargetType: "node", Title: "内存使用率", PromQLTemplate: `(1 - (node_memory_MemAvailable_bytes{instance="{{nodeName}}"} / node_memory_MemTotal_bytes{instance="{{nodeName}}"})) * 100`, ChartType: "line", UnitSuffix: "%", ColorTheme: "#409EFF"},
		{AccountID: uint(accountID), TargetType: "node", Title: "网络接收速率", PromQLTemplate: `sum(rate(node_network_receive_bytes_total{instance="{{nodeName}}"}[5m]))`, ChartType: "line", UnitSuffix: "B/s", ColorTheme: "#67c23a"},
		{AccountID: uint(accountID), TargetType: "node", Title: "根目录磁盘使用率", PromQLTemplate: `100 - (node_filesystem_avail_bytes{mountpoint="/", instance="{{nodeName}}"} / node_filesystem_size_bytes{mountpoint="/", instance="{{nodeName}}"} * 100)`, ChartType: "line", UnitSuffix: "%", ColorTheme: "#e6a23c"},
	}
	for _, m := range defaults {
		configs.GORMDB.Create(&m)
	}
}

func seedPodMonitors(accountID int64) {
	defaults := []dal.CustomMonitor{
		{AccountID: uint(accountID), TargetType: "pod", Title: "CPU 使用量", PromQLTemplate: `sum(rate(container_cpu_usage_seconds_total{pod="{{podName}}", container!="", container!="POD"}[5m])) by (pod)`, ChartType: "line", UnitSuffix: "Core", ColorTheme: "#f56c6c"},
		{AccountID: uint(accountID), TargetType: "pod", Title: "内存使用量", PromQLTemplate: `sum(container_memory_working_set_bytes{pod="{{podName}}", container!="", container!="POD"}) by (pod) / 1024 / 1024`, ChartType: "line", UnitSuffix: "MiB", ColorTheme: "#409EFF"},
		{AccountID: uint(accountID), TargetType: "pod", Title: "网络接收速率", PromQLTemplate: `sum(rate(container_network_receive_bytes_total{pod="{{podName}}"}[5m])) by (pod)`, ChartType: "line", UnitSuffix: "B/s", ColorTheme: "#67c23a"},
		{AccountID: uint(accountID), TargetType: "pod", Title: "网络发送速率", PromQLTemplate: `sum(rate(container_network_transmit_bytes_total{pod="{{podName}}"}[5m])) by (pod)`, ChartType: "line", UnitSuffix: "B/s", ColorTheme: "#e6a23c"},
	}
	for _, m := range defaults {
		configs.GORMDB.Create(&m)
	}
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
