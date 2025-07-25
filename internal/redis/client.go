package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/saxenaaman628/redis-voting-system/config"
)

var Rdb *redis.Client
var Ctx = context.Background()

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     config.GetEnv("REDIS_URI", "localhost:6379"),
		Password: config.GetEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})
	pong, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}

	fmt.Printf("✅ Redis connected: %s\n", pong)
}
