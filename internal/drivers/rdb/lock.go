package rdb

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLock struct {
	rdb    *Service
	key    string
	value  string
	expiry time.Duration
}

func (s *Service) NewRedisLock(key, value string, expiry time.Duration) *RedisLock {
	return &RedisLock{
		rdb:    s,
		key:    key,
		value:  value,
		expiry: expiry,
	}
}

// Lock tries to aquire a lock in an infinite loop,
// by setting key-value ONLY if the key doesn't exist.
// For Lock to work the workers should use the same key.
func (l *RedisLock) Lock(ctx context.Context) error {
	for {
		success, err := l.rdb.Client.SetNX(ctx, l.key, l.value, l.expiry).Result()

		if err != nil {
			return err
		}

		if success {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// CheckLock checks the ownership of the lock
func (l *RedisLock) CheckLock(ctx context.Context) error {

	// Get the lock value
	value, err := l.rdb.Client.Get(ctx, l.key).Result()

	if err == redis.Nil {
		return fmt.Errorf("lock expired or deleted")
	}

	if err != nil {
		return fmt.Errorf("connectivity error during lock check: %w", err)
	}

	if value != l.value {
		return fmt.Errorf(
			"lock ownership hijacked by another worker (expected %s, got %s)",
			l.value, value,
		)
	}

	return nil
}

// Unlock deletes the key-value from Redis
// ONLY if the value is the correct value using LUA atomic script.
func (l *RedisLock) Unlock(ctx context.Context) error {
	script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `
	_, err := l.rdb.Client.Eval(ctx, script, []string{l.key}, l.value).Result()
	return err
}
