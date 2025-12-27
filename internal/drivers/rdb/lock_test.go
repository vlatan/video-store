package rdb

import (
	"testing"
	"time"
)

func TestNewRedisLock(t *testing.T) {

	rdb, err := New(testCfg)
	if err != nil {
		t.Fatalf("failed to create Redis client; %v", err)
	}

	t.Cleanup(func() { rdb.Client.Close() })

	ttl := time.Nanosecond
	key, value := "foo", "bar"
	lock := rdb.NewRedisLock(key, value, ttl)

	if lock.rdb != rdb {
		t.Error("lock should reference the service")
	}
	if lock.key != key {
		t.Errorf("got key = %s, want %s", lock.key, key)
	}
	if lock.value != value {
		t.Errorf("got value = %s, want %s", lock.value, value)
	}
	if lock.ttl != ttl {
		t.Errorf("got ttl = %s, want %s", lock.ttl, ttl)
	}
}
