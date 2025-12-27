package rdb

import (
	"context"
	"testing"
	"time"
)

func TestNewLock(t *testing.T) {

	rdb, err := New(testCfg)
	if err != nil {
		t.Fatalf("failed to create Redis client; %v", err)
	}

	t.Cleanup(func() { rdb.Client.Close() })

	ttl := time.Nanosecond
	key, value := "foo", "bar"
	lock := rdb.NewLock(key, value, ttl)

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

func TestLock(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping timing-dependent test")
	}

	// Main context
	ctx := context.TODO()

	// Cancelled context
	noContext, cancel := context.WithCancel(ctx)
	cancel()

	rdb, err := New(testCfg)
	if err != nil {
		t.Fatalf("failed to create Redis client; %v", err)
	}

	t.Cleanup(func() { rdb.Client.Close() })

	tests := []struct {
		name         string
		ctx          context.Context
		lock1, lock2 *RedisLock
		wantErr      bool
	}{
		{
			"no context",
			noContext,
			rdb.NewLock("lock_key_1", "worker1", time.Nanosecond),
			rdb.NewLock("lock_key_1", "worker2", time.Nanosecond),
			true,
		},
		{
			"success lock",
			ctx,
			rdb.NewLock("lock_key_2", "worker1", 200*time.Millisecond),
			rdb.NewLock("lock_key_2", "worker2", 5*time.Second),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			err := tt.lock1.Lock(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}

			t.Cleanup(func() { tt.lock1.Unlock(ctx) })

			err = tt.lock2.Lock(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}

			tt.lock2.Unlock(ctx)
		})
	}
}
