package configs

import (
	"devops-console-backend/internal/dal/model"
	"fmt"
	"time"

	"github.com/emicklei/go-restful/v3/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	GORMDB *gorm.DB
)

func NewDB() *gorm.DB {
	databaseConfig := Config.Database.MySQL
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		databaseConfig.Username, databaseConfig.Password, databaseConfig.Host, databaseConfig.Port, databaseConfig.Database)
	GORMDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		log.Printf("database init failed: %v", err)
		return nil
	}

	sqlDB, _ := GORMDB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if Config != nil && Config.Database.AutoMigrate {
		if err := migrateRepairSchema(GORMDB); err != nil {
			log.Printf("repair schema migrate failed: %v", err)
		}
	}
	return GORMDB
}

func CloseDB() {
	sqlDB, _ := GORMDB.DB()
	if err := sqlDB.Close(); err != nil {
		log.Printf("database close failed: %v", err)
	}
}

func migrateRepairSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.RepairSession{},
		&model.SessionMessage{},
		&model.RepairAction{},
		&model.RepairSessionEvent{},
		&model.AlertDispatch{},
	); err != nil {
		return err
	}

	statusUpdates := map[string]string{
		"pending":         "created",
		"running":         "executing",
		"waiting_confirm": "waiting_approval",
		"success":         "completed",
		"partial":         "completed",
	}
	for oldStatus, newStatus := range statusUpdates {
		if err := db.Model(&model.RepairSession{}).Where("status = ?", oldStatus).Update("status", newStatus).Error; err != nil {
			return err
		}
	}
	return nil
}
