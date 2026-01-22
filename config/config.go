// 配置层 存放程序的配置信息
package config

import (
	"devops-console-backend/database"
	"devops-console-backend/models"
	"devops-console-backend/utils/logs"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/go-sql-driver/mysql"
)

var (
	Port   string
	Config *AppConfig
)

// MySQL配置
type MySQLConfig struct {
	Host         string `mapstructure:"host" yaml:"host"`
	Port         int    `mapstructure:"port" yaml:"port"`
	Username     string `mapstructure:"username" yaml:"username"`
	Password     string `mapstructure:"password" yaml:"password"`
	Database     string `mapstructure:"database" yaml:"database"`
	Charset      string `mapstructure:"charset" yaml:"charset"`
	ParseTime    bool   `mapstructure:"parse_time" yaml:"parse_time"`
	MaxOpenConns int    `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
}

// 数据库配置
type DatabaseConfig struct {
	Type        string      `mapstructure:"type" yaml:"type"`
	AutoMigrate bool        `mapstructure:"auto_migrate" yaml:"auto_migrate"`
	MySQL       MySQLConfig `mapstructure:"mysql" yaml:"mysql"`
}

// 服务器配置
type ServerConfig struct {
	Port     string `mapstructure:"port" yaml:"port"`
	LogLevel string `mapstructure:"log_level" yaml:"log_level"`
}

// 日志配置
type LoggingConfig struct {
	Format       string `mapstructure:"format" yaml:"format"`
	TimeFormat   string `mapstructure:"time_format" yaml:"time_format"`
	ReportCaller bool   `mapstructure:"report_caller" yaml:"report_caller"`
}

// Elasticsearch配置
type ElasticsearchConfig struct {
	Timeout             int `mapstructure:"timeout" yaml:"timeout"`
	Retry               int `mapstructure:"retry" yaml:"retry"`
	HealthCheckInterval int `mapstructure:"health_check_interval" yaml:"health_check_interval"`
}

// Kubernetes配置
type KubernetesConfig struct {
	ConfigPath string `mapstructure:"config_path" yaml:"config_path"`
	Timeout    int    `mapstructure:"timeout" yaml:"timeout"`
	Retry      int    `mapstructure:"retry" yaml:"retry"`
}

// Swagger配置
type SwaggerConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	Host     string `mapstructure:"host" yaml:"host"`
	BasePath string `mapstructure:"base_path" yaml:"base_path"`
}

// 健康检查配置
type HealthConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	Endpoint string `mapstructure:"endpoint" yaml:"endpoint"`
	Interval int    `mapstructure:"interval" yaml:"interval"`
}

// 应用配置
type AppConfig struct {
	Server        ServerConfig        `mapstructure:"server" yaml:"server"`
	Database      DatabaseConfig      `mapstructure:"database" yaml:"database"`
	Logging       LoggingConfig       `mapstructure:"logging" yaml:"logging"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch" yaml:"elasticsearch"`
	Kubernetes    KubernetesConfig    `mapstructure:"kubernetes" yaml:"kubernetes"`
	Swagger       SwaggerConfig       `mapstructure:"swagger" yaml:"swagger"`
	Health        HealthConfig        `mapstructure:"health" yaml:"health"`
}

// initLogConfig 初始化日志配置
func initLogConfig(config *AppConfig) {
	// 设置日志级别
	switch strings.ToLower(config.Server.LogLevel) {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	// 设置调用者信息
	logrus.SetReportCaller(config.Logging.ReportCaller)

	// 设置日志格式
	if config.Logging.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: config.Logging.TimeFormat,
			CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
				fileName := filepath.Base(f.File)
				return f.Function, fileName
			},
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: config.Logging.TimeFormat,
		})
	}
}

// InitMySQL 初始化MySQL数据库
func InitMySQL(config *MySQLConfig) error {
	// 首先连接到MySQL服务器创建数据库
	baseDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=%t&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Charset, config.ParseTime)

	baseDB, err := gorm.Open(mysql.Open(baseDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("连接MySQL服务器失败: %v", err)
	}

	// 创建数据库
	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", config.Database)
	if err := baseDB.Exec(createDBSQL).Error; err != nil {
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	// 关闭临时连接
	if sqlBaseDB, err := baseDB.DB(); err == nil {
		sqlBaseDB.Close()
	}

	// 连接到指定数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database, config.Charset, config.ParseTime)

	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:  logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time { return time.Now().Local() },
	})
	if err != nil {
		return fmt.Errorf("连接MySQL数据库失败: %v", err)
	}

	// 设置连接池参数
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	database.GORMDB = gormDB
	logs.Info(map[string]interface{}{
		"host":     config.Host,
		"port":     config.Port,
		"database": config.Database,
	}, "MySQL数据库连接成功")

	return initDatabase("mysql")
}

// initDatabase 通用数据库初始化
func initDatabase(dbType string) error {
	// 初始化数据库表结构
	if err := initDatabaseSchema(dbType); err != nil {
		return fmt.Errorf("初始化数据库表结构失败: %v", err)
	}

	// 自动迁移模型
	if err := autoMigrateModels(); err != nil {
		logs.Warning(map[string]interface{}{"error": err.Error()}, "GORM模型迁移失败，但继续使用数据库")
	}

	return nil
}

// InitDatabase 根据配置初始化数据库
func InitDatabase() error {
	return InitDatabaseWithType(Config.Database.Type)
}

// InitDatabaseWithType 根据指定的数据库类型初始化数据库
func InitDatabaseWithType(dbType string) error {
	switch strings.ToLower(dbType) {
	case "mysql":
		return InitMySQL(&Config.Database.MySQL)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
}

// initDatabaseWithFallback 数据库初始化
func initDatabaseWithFallback() error {
	dbType := strings.ToLower(Config.Database.Type)

	logs.Info(map[string]interface{}{
		"dbType":      dbType,
		"autoMigrate": Config.Database.AutoMigrate,
	}, "开始初始化数据库")

	if err := InitDatabaseWithType(dbType); err != nil {
		return fmt.Errorf("数据库(%s)初始化失败: %v", dbType, err)
	}

	return nil
}

// GetConfig 获取当前配置
func GetConfig() *AppConfig {
	return Config
}

// GetServerConfig 获取服务器配置
func GetServerConfig() ServerConfig {
	return Config.Server
}

// GetDatabaseConfig 获取数据库配置
func GetDatabaseConfig() DatabaseConfig {
	return Config.Database
}

// GetElasticsearchConfig 获取Elasticsearch配置
func GetElasticsearchConfig() ElasticsearchConfig {
	return Config.Elasticsearch
}

// GetKubernetesConfig 获取Kubernetes配置
func GetKubernetesConfig() KubernetesConfig {
	return Config.Kubernetes
}

// GetSwaggerConfig 获取Swagger配置
func GetSwaggerConfig() SwaggerConfig {
	return Config.Swagger
}

// GetHealthConfig 获取健康检查配置
func GetHealthConfig() HealthConfig {
	return Config.Health
}

// IsDebugMode 判断是否为调试模式
func IsDebugMode() bool {
	return strings.ToLower(Config.Server.LogLevel) == "debug"
}

// IsProductionMode 判断是否为生产模式
func IsProductionMode() bool {
	return strings.ToLower(Config.Server.LogLevel) == "error" ||
		strings.ToLower(Config.Server.LogLevel) == "warn"
}

// initDatabaseSchema 初始化数据库表结构
func initDatabaseSchema(dbType string) error {
	// 获取SQL脚本路径
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))

	sqlPath := filepath.Join(rootDir, "sql", "mysql.sql")

	// 读取SQL脚本
	sqlBytes, err := ioutil.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("读取SQL脚本失败: %v", err)
	}

	sqlContent := string(sqlBytes)

	// 执行MySQL脚本
	return executeMySQLScript(sqlContent)
}

// executeMySQLScript 分批执行MySQL脚本
func executeMySQLScript(sqlContent string) error {
	// 按分号分割SQL语句，但要注意忽略注释中的分号
	statements := splitSQLStatements(sqlContent)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") || strings.HasPrefix(stmt, "/*") {
			continue
		}

		err := database.GORMDB.Exec(stmt).Error
		if err != nil {
			// 检查是否是表已存在的错误
			if isTableExistsError(err) {
				logs.Info(map[string]interface{}{"statement": stmt}, "表已存在，跳过")
				continue
			}
			return fmt.Errorf("执行SQL语句失败: %v\n语句: %s", err, stmt)
		}
	}

	logs.Info(nil, "MySQL数据库表结构初始化完成")
	return nil
}

// splitSQLStatements 分割SQL语句
func splitSQLStatements(sqlContent string) []string {
	// 使用正则表达式分割SQL语句，保留换行符
	re := regexp.MustCompile(`;[\r\n]*`)
	matches := re.Split(sqlContent, -1)

	var statements []string
	for _, stmt := range matches {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" && !strings.HasPrefix(stmt, "--") && !strings.HasPrefix(stmt, "/*") {
			statements = append(statements, stmt)
		}
	}

	return statements
}

// isTableExistsError 检查是否是表已存在的错误
func isTableExistsError(err error) bool {
	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "already exists") ||
		strings.Contains(errorMsg, "table") && strings.Contains(errorMsg, "already exists") ||
		strings.Contains(errorMsg, "base table or view already exists") ||
		strings.Contains(errorMsg, "duplicate table name") ||
		strings.Contains(errorMsg, "table") && strings.Contains(errorMsg, "exists") ||
		strings.Contains(errorMsg, "already exists: table") || // SQLite错误格式
		strings.Contains(errorMsg, "table") && strings.Contains(errorMsg, "already") || // SQLite通用格式
		strings.Contains(err.Error(), "already exists") // 原始错误消息检查
}

// autoMigrateModels 自动迁移GORM模型
func autoMigrateModels() error {
	// 禁用外键检查，避免迁移时的约束问题
	if err := database.GORMDB.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		logs.Warning(map[string]interface{}{"error": err.Error()}, "禁用外键检查失败")
	}
	defer func() {
		if err := database.GORMDB.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
			logs.Warning(map[string]interface{}{"error": err.Error()}, "恢复外键检查失败")
		}
	}()

	// 先删除可能导致冲突的外键约束
	dropProblematicForeignKeys()

	// 按顺序迁移模型，确保依赖关系正确
	err := database.GORMDB.AutoMigrate(
		&models.InstanceType{},
		&models.Account{},
		&models.Instance{},
		&models.AuthConfig{},
		&models.ConnectionTest{},
	)

	if err != nil {
		return fmt.Errorf("GORM模型迁移失败: %v", err)
	}

	// 创建外键约束
	if err := createForeignKeys(); err != nil {
		logs.Warning(map[string]interface{}{"error": err.Error()}, "创建外键约束失败，但应用程序仍可正常运行")
	}

	logs.Info(nil, "GORM模型迁移完成")
	return nil
}

// dropProblematicForeignKeys 删除可能导致冲突的外键约束
func dropProblematicForeignKeys() {
	// 查询并删除instances表的外键约束
	var constraintName string
	database.GORMDB.Raw("SELECT constraint_name FROM information_schema.table_constraints WHERE constraint_schema = DATABASE() AND table_name = 'instances' AND constraint_type = 'FOREIGN KEY'").Scan(&constraintName)
	if constraintName != "" {
		if err := database.GORMDB.Exec(fmt.Sprintf("ALTER TABLE instances DROP FOREIGN KEY %s", constraintName)).Error; err != nil {
			logs.Debug(map[string]interface{}{"error": err.Error(), "constraint": constraintName}, "删除instances表外键约束失败")
		} else {
			logs.Info(map[string]interface{}{"constraint": constraintName}, "成功删除instances表外键约束")
		}
	}

	// 查询并删除connection_tests表的外键约束
	constraintName = ""
	database.GORMDB.Raw("SELECT constraint_name FROM information_schema.table_constraints WHERE constraint_schema = DATABASE() AND table_name = 'connection_tests' AND constraint_type = 'FOREIGN KEY'").Scan(&constraintName)
	if constraintName != "" {
		if err := database.GORMDB.Exec(fmt.Sprintf("ALTER TABLE connection_tests DROP FOREIGN KEY %s", constraintName)).Error; err != nil {
			logs.Debug(map[string]interface{}{"error": err.Error(), "constraint": constraintName}, "删除connection_tests表外键约束失败")
		} else {
			logs.Info(map[string]interface{}{"constraint": constraintName}, "成功删除connection_tests表外键约束")
		}
	}
}

// dropExistingForeignKeys 删除现有的外键约束
func dropExistingForeignKeys() error {
	// 获取所有外键约束
	var constraints []struct {
		TableName      string `gorm:"column:table_name"`
		ConstraintName string `gorm:"column:constraint_name"`
	}

	database.GORMDB.Raw("SELECT table_name, constraint_name FROM information_schema.table_constraints WHERE constraint_schema = DATABASE() AND constraint_type = 'FOREIGN KEY'").Scan(&constraints)

	// 删除每个外键约束
	for _, c := range constraints {
		sql := fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s", c.TableName, c.ConstraintName)
		if err := database.GORMDB.Exec(sql).Error; err != nil {
			logs.Debug(map[string]interface{}{"table": c.TableName, "fk": c.ConstraintName, "error": err.Error()}, "删除外键约束失败")
		}
	}

	return nil
}

// createForeignKeys 创建外键约束
func createForeignKeys() error {
	// 创建instances表的外键约束
	if err := database.GORMDB.Exec("ALTER TABLE instances ADD CONSTRAINT fk_instances_instance_type FOREIGN KEY (instance_type_id) REFERENCES instance_types (id) ON DELETE RESTRICT ON UPDATE CASCADE").Error; err != nil {
		logs.Warning(map[string]interface{}{"error": err.Error()}, "创建instances表外键约束失败")
		return err
	}

	// 创建connection_tests表的外键约束
	if err := database.GORMDB.Exec("ALTER TABLE connection_tests ADD CONSTRAINT fk_connection_tests_auth FOREIGN KEY (resource_type, resource_id) REFERENCES auth_configs (resource_type, resource_id) ON DELETE CASCADE").Error; err != nil {
		logs.Warning(map[string]interface{}{"error": err.Error()}, "创建connection_tests表外键约束失败")
		return err
	}

	logs.Info(nil, "外键约束创建完成")
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if database.GORMDB != nil {
		sqlDB, err := database.GORMDB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// InitSwagger 初始化Swagger API文档
func InitSwagger(r *gin.Engine) {
	if !Config.Swagger.Enabled {
		logs.Info(nil, "Swagger API文档已禁用")
		return
	}

	// Swagger API文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logs.Info(map[string]interface{}{
		"host":      Config.Swagger.Host,
		"base_path": Config.Swagger.BasePath,
	}, "Swagger API文档初始化完成")
}

// LoadConfig 加载配置文件
func LoadConfig() error {
	// 设置配置文件路径和名称
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置环境变量前缀
	viper.SetEnvPrefix("DEVOPS")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 手动绑定Docker Compose环境变量（无前缀）
	viper.BindEnv("database.mysql.host", "DB_HOST")
	viper.BindEnv("database.mysql.port", "DB_PORT")
	viper.BindEnv("database.mysql.username", "DB_USER")
	viper.BindEnv("database.mysql.password", "DB_PASSWORD")
	viper.BindEnv("database.mysql.database", "DB_NAME")

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logs.Info(nil, "配置文件未找到，使用默认配置和环境变量")
		} else {
			return fmt.Errorf("读取配置文件失败: %v", err)
		}
	}

	// 解析配置到结构体
	Config = &AppConfig{}
	if err := viper.Unmarshal(Config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置全局端口变量
	Port = Config.Server.Port

	// 初始化日志配置
	initLogConfig(Config)

	logs.Info(map[string]interface{}{
		"config_file": viper.ConfigFileUsed(),
		"log_level":   Config.Server.LogLevel,
	}, "配置加载完成")

	return nil
}

// setDefaults 设置最小必要的默认配置值
func setDefaults() {
	// 只设置最基本的默认值，其他配置从YAML文件读取
	viper.SetDefault("server.port", ":8081")
	viper.SetDefault("server.log_level", "info")

	// 如果没有配置文件，提供基本的生产可用默认值
	viper.SetDefault("database.type", "mysql")
	viper.SetDefault("database.auto_migrate", true)
	viper.SetDefault("database.mysql.host", "mysql") // Docker服务名
	viper.SetDefault("database.mysql.port", 3306)
	viper.SetDefault("database.mysql.username", "devops")       // Docker用户
	viper.SetDefault("database.mysql.password", "devops123456") // Docker密码
	viper.SetDefault("database.mysql.database", "devops_console")
	viper.SetDefault("database.mysql.charset", "utf8mb4")
	viper.SetDefault("database.mysql.parse_time", true)
	viper.SetDefault("database.mysql.max_open_conns", 10)
	viper.SetDefault("database.mysql.max_idle_conns", 5)
}

// Initialize 初始化应用配置
func Initialize() error {
	// 加载配置
	if err := LoadConfig(); err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 直接初始化用户选择的数据库类型
	if err := InitDatabase(); err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "数据库初始化失败")
		return handleDatabaseError(err)
	}

	// 初始化其他组件
	initComponents()

	return nil
}

// handleDatabaseError 处理数据库初始化错误
func handleDatabaseError(err error) error {
	logs.Error(nil, "请确保MySQL服务器正在运行且配置正确")
	logs.Info(map[string]interface{}{
		"host":     Config.Database.MySQL.Host,
		"port":     Config.Database.MySQL.Port,
		"username": Config.Database.MySQL.Username,
		"database": Config.Database.MySQL.Database,
	}, "MySQL连接配置信息")
	logs.Error(nil, "生产环境数据库初始化失败，程序退出")
	return err
}

// initComponents 初始化其他组件
func initComponents() {
	// 延迟初始化EsClient，改为按需加载
	InitEsClients()

	// 初始化K8s客户端
	if err := InitK8sClients(); err != nil {
		logs.Warning(map[string]interface{}{
			"error":       err.Error(),
			"config_path": Config.Kubernetes.ConfigPath,
			"timeout":     Config.Kubernetes.Timeout,
		}, "K8s客户端初始化失败")
	}
}

func init() {
	if err := Initialize(); err != nil {
		panic(err)
	}
}
