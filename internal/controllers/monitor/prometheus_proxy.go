package monitor

import (
	"devops-console-backend/internal/dal/query"
	"devops-console-backend/pkg/utils"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PrometheusController struct {
	db *gorm.DB
}

var (
	HttpProxyPrefix  = "http://"
	HttpsProxyPrefix = "https://"
)

func NewPrometheusController(db *gorm.DB) *PrometheusController {
	return &PrometheusController{
		db: db,
	}
}

// Proxy 处理转发到 Prometheus 实例的请求
func (p *PrometheusController) Proxy(ctx *gin.Context) {
	instanceIDStr := ctx.Param("instanceId")
	instanceID, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("无效的实例 ID")
		return
	}

	q := query.Use(p.db)

	// 1. 获取 Prometheus 实例地址
	instance, err := q.Instance.WithContext(ctx).Where(q.Instance.ID.Eq(uint32(instanceID))).First()
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("找不到指定的 Prometheus 实例")
		return
	}

	prometheusUrl := fmt.Sprintf("%v%v", HttpProxyPrefix, instance.Address)
	if instance.HTTPSEnabled != nil && *instance.HTTPSEnabled {
		prometheusUrl = fmt.Sprintf("%v%v", HttpsProxyPrefix, instance.Address)
	}
	// 解析目标 URL
	targetURL, err := url.Parse(prometheusUrl)
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("Prometheus 实例地址格式不正确" + err.Error())
		return
	}

	// 获取前端请求的路径后缀 (例如 /api/v1/query_range)
	proxyPath := ctx.Param("path")
	if !strings.HasPrefix(proxyPath, "/") {
		proxyPath = "/" + proxyPath
	}

	// 2. 获取可能的认证配置 (Basic Auth / Token 等)
	authConfig, err := q.AuthConfig.WithContext(ctx).
		Where(q.AuthConfig.ResourceType.Eq("instances"), q.AuthConfig.ResourceID.Eq(uint64(instanceID))).
		First()

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 重写 Director (修改请求往目标服务器发的样子)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 重新设置被代理请求的 Path
		req.URL.Path = targetURL.Path + proxyPath

		// 重新设置 Host 头，因为某些服务器可能会校验
		req.Host = targetURL.Host

		// 处理认证信息注入
		if err == nil && authConfig != nil && authConfig.ConfigValue != nil {
			if authConfig.AuthType == "token" {
				req.Header.Set("Authorization", "Bearer "+*authConfig.ConfigValue)
			} else if authConfig.AuthType == "basic" {
				// 对于 Basic Auth, ConfigValue 可能是以 username:password 形式存的 base64 或明文
				// 这里我们需要根据实际的 authConfig 逻辑做适配
				// 如果是明文存储的 username:password:
				parts := strings.SplitN(*authConfig.ConfigValue, ":", 2)
				if len(parts) == 2 {
					req.SetBasicAuth(parts[0], parts[1])
				}
			}
		}
	}

	// 设置错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"code":502, "msg": "代理请求到 Prometheus 失败", "data": "` + err.Error() + `"}`))
	}

	// 执行代理
	proxy.ServeHTTP(ctx.Writer, ctx.Request)
}
