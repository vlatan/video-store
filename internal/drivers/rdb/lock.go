package rdb

import (
	"context"
	"time"
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

// TryLock tries to aquire a lock,
// by setting key-value ONLY if the key doesn't exist.
// The key should be unique to the resource being locked,
// and the value should be unique to the client doing the lock.
func (l *RedisLock) TryLock(ctx context.Context) (bool, error) {
	return l.rdb.Client.SetNX(ctx, l.key, l.value, l.expiry).Result()
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
