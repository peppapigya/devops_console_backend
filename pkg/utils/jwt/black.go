package jwt

import (
	"context"
	"devops-console-backend/internal/dal/redis"
	"time"
)

// 黑名单机制

type BlackListManager struct {
	redisCli *redis.RedisClient
}

func NewBlackListManager(redisCli *redis.RedisClient) *BlackListManager {
	return &BlackListManager{redisCli: redisCli}
}

// Add 加入黑名单
func (m *BlackListManager) Add(ctx context.Context, key string, expireTime time.Duration) error {
	//key := fmt.Sprintf("%v:%v", common.BlockedTokenPrefix, token)
	return m.redisCli.SetWithExpiration(ctx, key, "1", expireTime)
}

func (m *BlackListManager) Exists(key string) bool {
	//key := fmt.Sprintf("%v:%v", common.BlockedTokenPrefix, )
	return m.redisCli.Exists(key)
}
