package eth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/redis/go-redis/v9"
)

type NonceManager struct {
	redis  *redis.Client
	client *ethclient.Client
}

func NewNonceManager(redis *redis.Client, client *ethclient.Client) *NonceManager {
	return &NonceManager{
		redis:  redis,
		client: client,
	}
}

// redis key
func (nm *NonceManager) nonceKey(addr common.Address) string {
	return fmt.Sprintf("nonce_%s", addr.Hex())
}

// redis 锁 key
func (nm *NonceManager) lockKey(addr common.Address) string {
	return fmt.Sprintf("nonce_lock_%s", addr.Hex())
}

// 尝试加锁
func (nm *NonceManager) acquireLock(ctx context.Context, key string) error {
	for {
		ok, err := nm.redis.SetNX(ctx, key, "1", 3*time.Second).Result()
		if err != nil {
			return err
		}
		if ok {
			return nil // 获得锁
		}

		// 锁被占用，稍后重试
		time.Sleep(30 * time.Millisecond)
	}
}

// 释放分布式锁
func (nm *NonceManager) releaseLock(ctx context.Context, key string) {
	nm.redis.Del(ctx, key)
}

// 获取 nonce
func (nm *NonceManager) GetNextNonce(ctx context.Context, addr common.Address) (uint64, error) {
	lock := nm.lockKey(addr)

	if err := nm.acquireLock(ctx, lock); err != nil {
		return 0, err
	}
	defer nm.releaseLock(ctx, lock)

	key := nm.nonceKey(addr)

	// 1. 尝试从 Redis 取 nonce
	val, err := nm.redis.Get(ctx, key).Result()

	if err == redis.Nil {
		// Redis 中没有 → 从链上初始化
		pending, err := nm.client.PendingNonceAt(ctx, addr)
		if err != nil {
			return 0, err
		}

		// 保存下一次要用的 nonce（pending+1）
		err = nm.redis.Set(ctx, key, pending+1, time.Minute*5).Err()
		if err != nil {
			return 0, err
		}

		return pending, nil
	}

	if err != nil {
		return 0, err
	}

	// 2. Redis 已有 nonce
	nonce, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, err
	}

	// 存入下一次的 nonce
	err = nm.redis.Set(ctx, key, nonce+1, 0).Err()
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

// 混滚 nonce
func (nm *NonceManager) RevertNonce(ctx context.Context, addr common.Address) error {
	key := nm.nonceKey(addr)

	val, err := nm.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	nonce, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return err
	}

	if nonce == 0 {
		return nil
	}

	return nm.redis.Set(ctx, key, nonce-1, 0).Err()
}
