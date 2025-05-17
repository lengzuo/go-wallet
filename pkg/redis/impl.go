package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func New(redisURL string) *redis.Client {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(opts)

	parentCtx := context.Background()
	ctxWithTTL, cancel := context.WithTimeout(parentCtx, 1*time.Second)
	defer cancel()
	resp := client.Ping(ctxWithTTL)
	if resp.Val() != "PONG" {
		panic("faield to setup redis")
	}
	return client
}
