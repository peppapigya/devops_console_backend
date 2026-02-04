// 项目的总入口
// @title DevOps Console API
// @version 1.0
// @description DevOps Console后端API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	_ "devops-console-backend/docs" // swagger docs
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/controllers/monitor"
	"devops-console-backend/internal/middlewares"
	"devops-console-backend/internal/routes"
	"devops-console-backend/internal/websocket"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/database"
	"devops-console-backend/pkg/utils/logs"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 1. 加载程序的配置
	// 2. 配置gin
	r := gin.Default()
	err := configs.LoadConfig()
	if err != nil {
		logs.Error(nil, "加载配置文件失败")
		panic(err)
	}
	globalConfig := common.GetGlobalConfig()
	setMiddleware(r, globalConfig)
	// 初始化数据库
	database.InitRedis()
	defer database.CloseRedis()
	configs.NewDB()
	defer configs.CloseDB()
	// 跨域配置 todo 待迁移
	r.Use(cors.New(cors.Config{
		//AllowOrigins:     []string{"http://127.0.0.1:5174", "http://localhost:5174"}, // 前端地址
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-ES-Host", "X-ES-Username", "X-ES-Password", "X-Requested-With", "Accept", "X-HTTP-Method-Override"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// 初始化 prometheus monitor
	monitor.InitPrometheus()
	configs.InitConfig()
	// 3. 日志配置
	logs.Info(nil, "程序启动成功")

	// Swagger API文档 - 初始化已移至 config 包
	configs.InitSwagger(r)

	// 添加健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// 注册路由
	routers.RegisterRouters(r, configs.GORMDB)
	// 注册WebSocket路由
	websocket.RegisterWebSocketRoutes(r)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(":9090", nil)
	}()
	err = r.Run(configs.Port)
	if err != nil {
		return
	}
}

// 设置中间件
func setMiddleware(router *gin.Engine, globalConfig *common.GlobalConfig) {
	// 认证
	router.Use(middlewares.Authenticate(globalConfig.Jwt.ExcludePaths...))
	router.Use(middlewares.Metrics())
	router.Use(middlewares.IPRateLimit())
}
