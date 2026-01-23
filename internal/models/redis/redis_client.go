package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// redis操作封装

type Client struct {
	client *redis.Client
}

func (c *Client) Set(id string, value string) error {
	return c.client.Set(context.Background(), id, value, time.Second*60).Err()
}

func (c *Client) Get(id string, clear bool) string {
	val := c.client.Get(context.Background(), id).Val()
	if clear {
		c.client.Del(context.Background(), id)
	}
	return val
}

func (c *Client) Verify(id, answer string, clear bool) bool {
	value := c.Get(id, clear)
	return value == answer
}

func NewClient(client *redis.Client) *Client {
	return &Client{
		client: client,
	}
}

// SetWithExpiration 设置缓存
func (c *Client) SetWithExpiration(cxt context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(cxt, key, value, expiration).Err()
}

// Delete 删除缓存
func (c *Client) Delete(key string) error {
	_, err := c.client.Del(context.Background(), key).Result()
	return err
}

// IsExpired 判断缓存是否过期
func (c *Client) IsExpired(key string) bool {
	result, err := c.client.TTL(context.Background(), key).Result()
	if err != nil {
		return true
	}
	return result < 0
}
