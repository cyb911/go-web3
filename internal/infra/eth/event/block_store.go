package event

import (
	"context"
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type BlockStore interface {
	GetLastBlock(ctx context.Context, chain string, contract string) (uint64, error)
	SetLastBlock(ctx context.Context, chain string, contract string, v uint64) error
}

const BlockKeyPrefix = "event:lastBlock1:"

//存 lastProcessedBlock

type RedisBlockStore struct {
	Client *redis.Client
	v      uint64
}

func NewRedisBlockStore(client *redis.Client, v uint64) *RedisBlockStore {
	return &RedisBlockStore{client, v}
}

func key(chain, contract string) string {
	return BlockKeyPrefix + chain + ":" + contract
}

func (rs *RedisBlockStore) GetLastBlock(ctx context.Context, chain string, contract string) (uint64, error) {
	str, err := rs.Client.Get(ctx, key(chain, contract)).Result()
	if err == redis.Nil {
		return rs.v, nil
	}
	if err != nil {
		return rs.v, err
	}

	val, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		log.Printf("no last block found for %s/%s", chain, contract)
		return rs.v, err
	}
	return val, nil
}

func (rs *RedisBlockStore) SetLastBlock(ctx context.Context, chain string, contract string, v uint64) error {
	return rs.Client.Set(ctx, key(chain, contract), strconv.FormatUint(v, 10), 0).Err()
}
