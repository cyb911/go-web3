package redis

import (
	"context"
	"go-web3/internal/config"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Rdb *redis.Client
	Ctx = context.Background()
)

// Init 初始化 Redis 客户端
func InitRedis() {
	cfg := config.Get()
	Rdb = redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     20,
		MinIdleConns: 5,

		DialTimeout:  3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	if err := Rdb.Ping(Ctx).Err(); err != nil {
		log.Fatalf("Redis 连接失败: %v", err)
	} else {
		log.Println("Redis 连接成功")
	}
}
