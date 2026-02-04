package database

import (
	"devops-console-backend/internal/common"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func InitRedis() *redis.Client {
	redisProperties := common.GetGlobalConfig().Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisProperties.Host, redisProperties.Port),
		Username: "",
		Password: redisProperties.Password,
		DB:       redisProperties.DB,
	})
	redisClient = client
	return redisClient
}

func CloseRedis() {
	err := redisClient.Close()
	if err != nil {
		fmt.Printf("关闭redis客户端失败: %v", err)
		return
	}
}
func GetRedisClient() *redis.Client {
	return redisClient
}
