// 中间件层
package middlewares

import (
	"context"
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/controllers/monitor"
	"devops-console-backend/internal/dal/redis"
	"devops-console-backend/pkg/database"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/jwt"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Authenticate  认证中间件
func Authenticate(excludePaths ...string) gin.HandlerFunc {
	excludePathRegex := make([]*regexp.Regexp, 0)
	for _, path := range excludePaths {
		str := strings.ReplaceAll(path, "*", ".*")
		excludePathRegex = append(excludePathRegex, regexp.MustCompile(str))
	}
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// 如果当前请求路径在排除列表中，则不进行权限验证
		for _, regexPath := range excludePathRegex {
			if regexPath.MatchString(path) {
				c.Next()
				return
			}
		}
		token := c.GetHeader(common.TokenKey)
		token, found := strings.CutPrefix(token, "Bearer ")
		if token == "" && !found {
			log.Print("token not found")
			common.Fail(c, common.UNAUTHORIZED)
			c.Abort()
			return
		}
		claims, err := jwt.ParseToken(token)
		if err != nil {
			common.Fail(c, common.UNAUTHORIZED)
			c.Abort()
			return
		}
		// 如果redis中有数据，执行用户下线操作
		redisOperator(claims, c, token)
		// 将解析的用户信息设置到上下文中
		c.Set(common.UserInfoKey, claims)
		c.Next()
	}
}

// Metrics 相关中间件
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		monitor.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(c.Writer.Status()),
		).Inc()

		monitor.HttpDuration.WithLabelValues(
			c.FullPath(),
		).Observe(time.Since(start).Seconds())
	}
}

// IPRateLimit IP限流中间件
func IPRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := utils.GetClientIP(c.Request)
		key := fmt.Sprintf("api:ip:req:%s", time.Now().Format("202602021504"))

		res, err := database.IPLimitScript.Run(
			context.Background(),
			database.GetRedisClient(),
			[]string{key},
			ip,
			100,
			120,
		).Int()

		if err != nil || res == 1 {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(429, gin.H{"message": "请求次数过多"})
	}
}

// redis 相关操作
func redisOperator(claim *jwt.Claims, c *gin.Context, token string) {
	client := database.GetRedisClient()
	if client == nil {
		panic("redis 客户端未初始化")
		return
	}
	key := fmt.Sprintf("%v:%v:%v", common.BlockedTokenPrefix, claim.GetUserId(), token)
	redisClient := redis.NewClient(client)
	helper := utils.NewResponseHelper(c)
	value := redisClient.Get(key, false)
	if value != "" {
		_ = redisClient.Delete(key)
		helper.Error(401, "账号已在其他地方登录")
		c.Abort()
		return
	}
}

// InstanceAuth 实例认证中间件
func InstanceAuth() gin.HandlerFunc {
	return func(r *gin.Context) {
		// 从请求中获取实例ID，这里简化处理，实际应该从token或其他认证信息中获取
		// 首先尝试从URL参数中获取
		instanceIDStr := r.Query("instance_id")
		if instanceIDStr == "" {
			// 如果URL参数中没有，尝试从header中获取
			instanceIDStr = r.GetHeader("X-Instance-ID")
		}

		var instanceID uint
		if instanceIDStr != "" {
			// 如果找到了实例ID，转换为uint
			if id, err := strconv.ParseUint(instanceIDStr, 10, 32); err == nil {
				instanceID = uint(id)
			} else {
				// 转换失败，使用默认值
				instanceID = 1
			}
		} else {
			// 如果没有找到实例ID，使用默认值
			instanceID = 1
		}

		r.Set("instance_id", instanceID)
		r.Next()
	}
}
