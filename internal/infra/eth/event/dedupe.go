package event

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/redis/go-redis/v9"
)

type DedupeStore interface {
	AlreadyHandled(lg types.Log) bool
	MarkHandled(lg types.Log)
}

type RedisDedupeStore struct {
	Client *redis.Client
	Ctx    context.Context
}

func NewRedisDedupeStore(ctx context.Context, client *redis.Client) *RedisDedupeStore {
	return &RedisDedupeStore{client, ctx}
}

func logKey(lg types.Log) string {
	return fmt.Sprintf("event:handled:%s:%s:%d",
		lg.BlockHash.Hex(),
		lg.TxHash.Hex(),
		lg.Index,
	)
}

func (rs *RedisDedupeStore) AlreadyHandled(lg types.Log) bool {
	key := logKey(lg)
	_, err := rs.Client.Get(rs.Ctx, key).Result()
	return err == nil // key 存在 = 已处理
}

func (rs *RedisDedupeStore) MarkHandled(lg types.Log) {
	key := logKey(lg)

	// 存 30 天即可，不要永久占存储
	rs.Client.Set(rs.Ctx, key, 1, 30*24*time.Hour)
}
