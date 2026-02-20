package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client
var Ctx = context.Background()

const (
	TransformQueueKey = "zennews:queue:transform"
	DeadLetterKey     = "zennews:queue:failed"
)

func ConnectRedis() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		fmt.Println("REDIS_URL environment variable is not set")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		opt = &redis.Options{Addr: redisURL}
	}

	Redis = redis.NewClient(opt)

	_, err = Redis.Ping(Ctx).Result()
	return err
}

func CloseRedis() {
	if Redis != nil {
		Redis.Close()
	}
}

func PushToQueue(queueKey string, data string) error {
	return Redis.LPush(Ctx, queueKey, data).Err()
}

func PopFromQueue(queueKey string, timeout time.Duration) (string, error) {
	result, err := Redis.BRPop(Ctx, timeout, queueKey).Result()
	if err != nil {
		return "", err
	}
	return result[1], nil
}

func GetqueueLength(queueKey string) (int64, error) {
	return Redis.LLen(Ctx, queueKey).Result()
}
