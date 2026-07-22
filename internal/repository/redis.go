package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
	ttl    time.Duration
}

type URLCache struct {
	LongURL   string `json:"long_url"`
	ShortCode string `json:"short_code"`
}

func NewRedisRepo(addr, password string, db int, ttl time.Duration) (*RedisRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	return &RedisRepo{client: client, ttl: ttl}, nil
}

func (r *RedisRepo) GetLongLink(ctx context.Context, shortCode string) (string, error) {
	key := fmt.Sprintf("short:%s", shortCode)

	data, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", errors.New("not found in cache")
	}

	var cache URLCache
	if err := json.Unmarshal([]byte(data), &cache); err != nil {
		return "", err
	}

	return cache.LongURL, nil

}

func (r *RedisRepo) SetLongLink(ctx context.Context, shortCode, longURL string) error {
	key := fmt.Sprintf("short:%s", shortCode)

	cache := URLCache{
		LongURL:   longURL,
		ShortCode: shortCode,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, r.ttl).Err()
}
