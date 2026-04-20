// 配置层 存放程序的配置信息
package configs

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/pkg/utils/logs"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

// ai相关配置文件
type AiConfig struct {
	MCPConfig MCPConfig `mapstructure:"mcp"`
}

type MCPConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Url        string `mapstructure:"url"`   // mcp-agent 工具服务器地址
	Token      string `mapstructure:"token"` // mcp-agent 认证 token
	MaxRetries int    `mapstructure:"max_retries"`
	// LLM 直连配置（Backend 作为 Host 直接调用通义千问）
	LLMApiKey  string `mapstructure:"llm_api_key"`
	LLMBaseUrl string `mapstructure:"llm_base_url"` // 如: https://dashscope.aliyuncs.com/compatible-mode/v1
	LLMModel   string `mapstructure:"llm_model"`    // 如: qwen-plus
	MaxRounds  int    `mapstructure:"max_rounds"`   // Agent 最大循环轮数，0 表示默认 10
}

// 应用配置
type EncryptionConfig struct {
	Key string `mapstructure:"key" yaml:"key"`
}

// 飞书相关配置
type FeiShuConfig struct {
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled"`
	WebHookUrl string `mapstructure:"webhook"`
	Secret     string `mapstructure:"secret"`
	AppId      string `mapstructure:"appId"`
	AppSecret  string `mapstructure:"appSecret"`
}

type AlertingConfig struct {
	Alertmanager   AlertmanagerConfig     `mapstructure:"alertmanager" yaml:"alertmanager"`
	Analysis       AlertingAnalysisConfig `mapstructure:"analysis" yaml:"analysis"`
	FeiShu         AlertingFeiShuConfig   `mapstructure:"feishu" yaml:"feishu"`
	AssetMapping   AssetMappingConfig     `mapstructure:"asset_mapping" yaml:"asset_mapping"`
	ConsoleBaseURL string                 `mapstructure:"console_base_url" yaml:"console_base_url"`
}

type AlertmanagerConfig struct {
	Enabled       bool   `mapstructure:"enabled" yaml:"enabled"`
	WebhookSecret string `mapstructure:"webhook_secret" yaml:"webhook_secret"`
	AuthHeader    string `mapstructure:"auth_header" yaml:"auth_header"`
}

type AlertingAnalysisConfig struct {
	AutoCreateSession   bool `mapstructure:"auto_create_session" yaml:"auto_create_session"`
	PollIntervalSeconds int  `mapstructure:"poll_interval_seconds" yaml:"poll_interval_seconds"`
	PollTimeoutSeconds  int  `mapstructure:"poll_timeout_seconds" yaml:"poll_timeout_seconds"`
}

type AlertingFeiShuConfig struct {
	NotifyOnFiring   bool `mapstructure:"notify_on_firing" yaml:"notify_on_firing"`
	NotifyOnResult   bool `mapstructure:"notify_on_result" yaml:"notify_on_result"`
	NotifyOnResolved bool `mapstructure:"notify_on_resolved" yaml:"notify_on_resolved"`
}

type AssetMappingConfig struct {
	Enabled           bool `mapstructure:"enabled" yaml:"enabled"`
	AllowNoCredential bool `mapstructure:"allow_no_credential" yaml:"allow_no_credential"`
}

type AppConfig struct {
	Server        ServerConfig        `mapstructure:"server" yaml:"server"`
	Database      DatabaseConfig      `mapstructure:"database" yaml:"database"`
	Logging       LoggingConfig       `mapstructure:"logging" yaml:"logging"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch" yaml:"elasticsearch"`
	Kubernetes    KubernetesConfig    `mapstructure:"kubernetes" yaml:"kubernetes"`
	Swagger       SwaggerConfig       `mapstructure:"swagger" yaml:"swagger"`
	Health        HealthConfig        `mapstructure:"health" yaml:"health"`
	AiConfig      AiConfig            `mapstructure:"ai"`
	Encryption    EncryptionConfig    `mapstructure:"encryption" yaml:"encryption"`
	FeiShuConfig  FeiShuConfig        `mapstructure:"feishu" yaml:"feishu"`
	Alerting      AlertingConfig      `mapstructure:"alerting" yaml:"alerting"`
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

func GetEncryptionConfig() EncryptionConfig {
	return Config.Encryption
}

func GetEncryptionKey() ([]byte, error) {
	keyStr := Config.Encryption.Key
	if keyStr == "" {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(keyStr)
}

// GetAiConfig 获取ai配置文件
func GetAiConfig() AiConfig {
	return Config.AiConfig
}

func GetAlertingConfig() AlertingConfig {
	return Config.Alerting
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
	_ = viper.BindEnv("database.mysql.host", "DB_HOST")
	_ = viper.BindEnv("database.mysql.port", "DB_PORT")
	_ = viper.BindEnv("database.mysql.username", "DB_USER")
	_ = viper.BindEnv("database.mysql.password", "DB_PASSWORD")
	_ = viper.BindEnv("database.mysql.database", "DB_NAME")

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			logs.Info(nil, "配置文件未找到，使用默认配置和环境变量")
		}
	}
	err := common.LoadConfig()
	if err != nil {
		return err
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
	viper.SetDefault("alerting.alertmanager.enabled", false)
	viper.SetDefault("alerting.alertmanager.auth_header", "X-Alertmanager-Token")
	viper.SetDefault("alerting.analysis.auto_create_session", true)
	viper.SetDefault("alerting.analysis.poll_interval_seconds", 5)
	viper.SetDefault("alerting.analysis.poll_timeout_seconds", 900)
	viper.SetDefault("alerting.feishu.notify_on_firing", true)
	viper.SetDefault("alerting.feishu.notify_on_result", true)
	viper.SetDefault("alerting.feishu.notify_on_resolved", true)
	viper.SetDefault("alerting.asset_mapping.enabled", true)
	viper.SetDefault("alerting.asset_mapping.allow_no_credential", true)
}

// Initialize 初始化应用配置
func Initialize() error {
	// 初始化其他组件
	initComponents()

	return nil
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

func InitConfig() {
	if err := Initialize(); err != nil {
		panic(err)
	}
}
