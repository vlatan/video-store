package rdb

import (
	"context"
	"testing"
	"time"
)

func TestNewLock(t *testing.T) {

	ttl := time.Nanosecond
	key, value := "foo", "bar"
	lock := testRdb.NewLock(key, value, ttl)

	if lock.rdb != testRdb {
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

func TestUnlock(t *testing.T) {

	tests := []struct {
		name    string
		ctx     context.Context
		lock    *RedisLock
		wantErr bool
	}{
		{
			"no context", noCtx,
			testRdb.NewLock("random_unlock_key", "worker", time.Millisecond),
			true,
		},
		{
			"success unlock", baseCtx,
			testRdb.NewLock("new_unlock_key", "worker", time.Second),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			err := tt.lock.Unlock(tt.ctx)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf(
					"got error = %v, want error = %t",
					err, tt.wantErr,
				)
			}

			tt.lock.Unlock(baseCtx)
		})
	}
}

func TestLock(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping timing-dependent test")
	}

	tests := []struct {
		name         string
		ctx          context.Context
		lock1, lock2 *RedisLock
		wantErr      bool
	}{
		{
			"no context",
			noCtx,
			testRdb.NewLock("lock_key_1", "worker1", time.Millisecond),
			testRdb.NewLock("lock_key_1", "worker2", time.Millisecond),
			true,
		},
		{
			"success lock",
			baseCtx,
			testRdb.NewLock("lock_key_2", "worker1", 200*time.Millisecond),
			testRdb.NewLock("lock_key_2", "worker2", 5*time.Second),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			err := tt.lock1.Lock(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}

			t.Cleanup(func() { tt.lock1.Unlock(baseCtx) })

			err = tt.lock2.Lock(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}

			tt.lock2.Unlock(baseCtx)
		})
	}
}

func TestTryLock(t *testing.T) {

	const existingLockKey = "try_lock_key"
	existingLock := testRdb.NewLock(existingLockKey, "existing_try_worker", 5*time.Second)
	if err := existingLock.Lock(baseCtx); err != nil {
		t.Fatalf("failed to create Redis lock; %v", err)
	}
	t.Cleanup(func() { existingLock.Unlock(baseCtx) })

	tests := []struct {
		name             string
		ctx              context.Context
		lock             *RedisLock
		success, wantErr bool
	}{
		{
			"no context", noCtx,
			testRdb.NewLock("random_try_lock_key", "worker", time.Millisecond),
			false, true,
		},
		{
			"success lock", baseCtx,
			testRdb.NewLock("new_try_lock_key", "worker", time.Millisecond),
			true, false,
		},
		{
			"failed lock", baseCtx,
			testRdb.NewLock(existingLockKey, "worker", time.Millisecond),
			false, false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			success, err := tt.lock.TryLock(tt.ctx)
			if success != tt.success {
				t.Errorf(
					"got success = %t, want success = %t",
					success, tt.success,
				)
			}

			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf(
					"got error = %v, want error = %t",
					err, tt.wantErr,
				)
			}

			tt.lock.Unlock(baseCtx)
		})
	}
}

func TestCheckLock(t *testing.T) {

	const existingLockKey = "check_lock_key"
	const existingLockValue = "existing_check_worker"
	existingLock := testRdb.NewLock(existingLockKey, existingLockValue, 5*time.Second)
	if err := existingLock.Lock(baseCtx); err != nil {
		t.Fatalf("failed to create Redis lock; %v", err)
	}

	t.Cleanup(func() { existingLock.Unlock(baseCtx) })

	tests := []struct {
		name    string
		ctx     context.Context
		lock    *RedisLock
		wantErr bool
	}{
		{
			"no context", noCtx,
			testRdb.NewLock("random_check_lock_key", "worker", time.Millisecond),
			true,
		},
		{
			"no lock", baseCtx,
			testRdb.NewLock("new_check_lock_key", "worker", time.Millisecond),
			true,
		},
		{
			"lock exists", baseCtx,
			testRdb.NewLock(existingLockKey, "worker", time.Millisecond),
			true,
		},
		{
			"lock doesn't exist", baseCtx,
			testRdb.NewLock(existingLockKey, existingLockValue, time.Millisecond),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			err := tt.lock.CheckLock(tt.ctx)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf(
					"got error = %v, want error = %t",
					err, tt.wantErr,
				)
			}

			tt.lock.Unlock(baseCtx)
		})
	}
}
