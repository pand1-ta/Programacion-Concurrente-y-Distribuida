package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(host, port string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
		DB:   0,
	})
	return &RedisClient{Client: rdb}
}

func (r *RedisClient) GetCached(key string, dest any) (bool, error) {
	ctx := context.Background()
	s, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(s), dest); err != nil {
		return false, err
	}
	return true, nil
}

func (r *RedisClient) SetCached(key string, value any, ttl time.Duration) error {
	ctx := context.Background()
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Client.Set(ctx, key, b, ttl).Err()
}
