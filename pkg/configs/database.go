package configs

import (
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

// NewDB 连接数据库
func NewDB() *gorm.DB {
	databaseConfig := Config.Database.MySQL
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		databaseConfig.Username, databaseConfig.Password, databaseConfig.Host, databaseConfig.Port, databaseConfig.Database)
	GORMDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Printf("数据库初始化失败: %v", err)
		return nil
	}
	sqlDb, _ := GORMDB.DB()
	// 设置最大空闲连接数
	sqlDb.SetMaxIdleConns(10)
	// 设置最大打开连接数
	sqlDb.SetMaxOpenConns(100)
	// 设置每个连接的过期时间
	sqlDb.SetConnMaxLifetime(time.Hour)
	return GORMDB
}

func CloseDB() {
	sqlDb, _ := GORMDB.DB()
	if err := sqlDb.Close(); err != nil {
		log.Printf("数据库关闭连接失败: %v", err)
		return
	}
}
