package cache

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/yaghoubi-mn/voter/internal/config"
)

type Cache interface {
	Get(key string) (string, error)
	Set(key string, data string) error
	FlushDB() error
	Del(key string) error
}

type cache struct {
	redisClient *redis.Client
	ctx         context.Context
}

func NewCache(redisClient *redis.Client, ctx context.Context) Cache {
	return &cache{
		redisClient: redisClient,
		ctx:         ctx,
	}
}

// env variables should be loaded
func Setup() (*redis.Client, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return redisClient, nil
}

func (c *cache) Set(key string, data string) error {

	err := c.redisClient.Set(c.ctx, key, data, config.RedisExpiration).Err()
	return err
}

func (c *cache) Get(key string) (string, error) {

	data, err := c.redisClient.Get(c.ctx, key).Result()
	return data, err
}

func (c *cache) FlushDB() error {
	return c.redisClient.FlushDB(c.ctx).Err()
}

func (c *cache) Del(key string) error {
	return c.redisClient.Del(c.ctx, key).Err()
}
