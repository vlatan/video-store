package rdb

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLock struct {
	rdb   *Service
	key   string // should be unique to the resource being locked
	value string // should be unique to the worker doing the lock
	ttl   time.Duration
}

func (s *Service) NewRedisLock(key, value string, ttl time.Duration) *RedisLock {
	return &RedisLock{
		rdb:   s,
		key:   key,
		value: value,
		ttl:   ttl,
	}
}

// Lock tries to aquire a lock in an infinite loop.
// It sets key-value ONLY if the key doesn't exist.
// Therefore it's a blocking method until it can acquire the lock.
func (l *RedisLock) Lock(ctx context.Context) error {
	for {
		success, err := l.rdb.Client.SetNX(ctx, l.key, l.value, l.ttl).Result()

		if err != nil && err != redis.Nil {
			return fmt.Errorf("unexpected error during lock acquire; %w", err)
		}

		if success {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// TryLock only tries to aquire a lock,
// and informs the caller if it was successful or not.
// It sets key-value ONLY if the key doesn't exist.
func (l *RedisLock) TryLock(ctx context.Context) (bool, error) {
	return l.rdb.Client.SetNX(ctx, l.key, l.value, l.ttl).Result()
}

// CheckLock checks if the caller still owns the lock.
func (l *RedisLock) CheckLock(ctx context.Context) error {

	// Get the lock value
	value, err := l.rdb.Client.Get(ctx, l.key).Result()

	if err == redis.Nil {
		return fmt.Errorf("lock expired or deleted; %w", err)
	}

	if err != nil {
		return fmt.Errorf("unexpected error during lock check; %w", err)
	}

	if value != l.value {
		return fmt.Errorf(
			"caller does not own this lock (expected %s, got %s)",
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
