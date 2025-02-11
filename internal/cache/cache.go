package cache

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/yaghoubi-mn/voter/internal/config"
)

type Cache interface {
	Get(model string, id uint64) (string, error)
	Set(model string, id uint64, data string) error
	FlushDB() error
	Del(models string, id uint64) error
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

func (c *cache) Set(model string, id uint64, data string) error {

	err := c.redisClient.Set(c.ctx, fmt.Sprintf("%v:%v", model, id), data, config.RedisExpiration).Err()

	return err
}

func (c *cache) Get(model string, id uint64) (string, error) {

	data, err := c.redisClient.Get(c.ctx, fmt.Sprintf("%v:%v", model, id)).Result()

	return data, err
}

func (c *cache) FlushDB() error {
	return c.redisClient.FlushDB(c.ctx).Err()
}

func (c *cache) Del(model string, id uint64) error {
	return c.redisClient.Del(c.ctx, fmt.Sprintf("%v:%v", model, id)).Err()
}
