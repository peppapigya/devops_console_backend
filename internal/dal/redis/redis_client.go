package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// redis操作封装

type RedisClient struct {
	client *redis.Client
}

func (c *RedisClient) Set(id string, value string) error {
	return c.client.Set(context.Background(), id, value, time.Second*60).Err()
}

func (c *RedisClient) Get(id string, clear bool) string {
	val := c.client.Get(context.Background(), id).Val()
	if clear {
		c.client.Del(context.Background(), id)
	}
	return val
}

// Exists 判断 key 是否存在
func (c *RedisClient) Exists(id string) bool {
	return c.client.Exists(context.Background(), id).Val() > 0
}
func (c *RedisClient) Verify(id, answer string, clear bool) bool {
	value := c.Get(id, clear)
	return value == answer
}

func NewClient(client *redis.Client) *RedisClient {
	return &RedisClient{
		client: client,
	}
}

// SetWithExpiration 设置缓存
func (c *RedisClient) SetWithExpiration(cxt context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(cxt, key, value, expiration).Err()
}

// Delete 删除缓存
func (c *RedisClient) Delete(key string) error {
	_, err := c.client.Del(context.Background(), key).Result()
	return err
}

// IsExpired 判断缓存是否过期
func (c *RedisClient) IsExpired(key string) bool {
	result, err := c.client.TTL(context.Background(), key).Result()
	if err != nil {
		return true
	}
	return result < 0
}
