package utils

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestExtractRetryDelay(t *testing.T) {

	type test struct {
		name          string
		err           error
		expectedDelay time.Duration
		expectedOk    bool
	}

	makeGRPCError := func(delay time.Duration) error {
		st := status.New(codes.ResourceExhausted, "rate limited")
		retryInfo := &errdetails.RetryInfo{
			RetryDelay: durationpb.New(delay),
		}
		st, _ = st.WithDetails(retryInfo)
		return st.Err()
	}

	tests := []test{
		{
			name:          "nil error",
			err:           nil,
			expectedDelay: 0,
			expectedOk:    false,
		},
		{
			name:          "non-gRPC error",
			err:           errors.New("regular error"),
			expectedDelay: 0,
			expectedOk:    false,
		},
		{
			name:          "gRPC non-RESOURCE_EXHAUSTED error",
			err:           status.Error(codes.Internal, "internal error"),
			expectedDelay: 0,
			expectedOk:    false,
		},
		{
			name:          "gRPC RESOURCE_EXHAUSTED without RetryInfo",
			err:           status.Error(codes.ResourceExhausted, "rate limited"),
			expectedDelay: 0,
			expectedOk:    false,
		},
		{
			name:          "gRPC RESOURCE_EXHAUSTED with RetryInfo",
			err:           makeGRPCError(5 * time.Second),
			expectedDelay: 5 * time.Second,
			expectedOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay, ok := extractRetryDelay(tt.err)

			if ok != tt.expectedOk {
				t.Errorf("got ok = %t, want %t", ok, tt.expectedOk)
			}

			if delay != tt.expectedDelay {
				t.Errorf("got delay = %v, want %v", delay, tt.expectedDelay)
			}
		})
	}
}

func TestRetry(t *testing.T) {

	type test struct {
		name         string
		ctx          context.Context
		retryConfig  *RetryConfig
		callable     func() (string, error)
		expectedData string
		wantErr      bool
	}

	makeGRPCError := func(delay time.Duration) error {
		st := status.New(codes.ResourceExhausted, "rate limited")
		retryInfo := &errdetails.RetryInfo{
			RetryDelay: durationpb.New(delay),
		}
		st, _ = st.WithDetails(retryInfo)
		return st.Err()
	}

	ctx := context.TODO()
	noContext, cancel := context.WithCancel(ctx)
	cancel()

	tests := []test{
		{
			name: "success (0 retries)",
			ctx:  ctx,
			retryConfig: &RetryConfig{
				MaxRetries: 0,
				MaxJitter:  time.Nanosecond,
				Delay:      time.Nanosecond,
			},
			callable:     func() (string, error) { return "foo", nil },
			expectedData: "foo",
			wantErr:      false,
		},
		{
			name: "success (1+ retries)",
			ctx:  ctx,
			retryConfig: &RetryConfig{
				MaxRetries: 3,
				MaxJitter:  time.Nanosecond,
				Delay:      time.Nanosecond,
			},
			callable:     func() (string, error) { return "foo", nil },
			expectedData: "foo",
			wantErr:      false,
		},
		{
			name: "error",
			ctx:  ctx,
			retryConfig: &RetryConfig{
				MaxRetries: 3,
				MaxJitter:  time.Nanosecond,
				Delay:      time.Nanosecond,
			},
			callable:     func() (string, error) { return "", errors.New("error") },
			expectedData: "",
			wantErr:      true,
		},
		{
			name: "error (no context)",
			ctx:  noContext,
			retryConfig: &RetryConfig{
				MaxRetries: 3,
				MaxJitter:  time.Nanosecond,
				Delay:      time.Nanosecond,
			},
			callable:     func() (string, error) { return "", errors.New("error") },
			expectedData: "",
			wantErr:      true,
		},
		{
			name: "gRPC error",
			ctx:  ctx,
			retryConfig: &RetryConfig{
				MaxRetries: 3,
				MaxJitter:  time.Nanosecond,
				Delay:      time.Nanosecond,
			},
			callable:     func() (string, error) { return "", makeGRPCError(time.Nanosecond) },
			expectedData: "",
			wantErr:      true,
		},
		{
			name: "gRPC error (gRPC delay > sleep time)",
			ctx:  ctx,
			retryConfig: &RetryConfig{
				MaxRetries: 3,
				MaxJitter:  time.Nanosecond,
				Delay:      time.Nanosecond,
			},
			callable:     func() (string, error) { return "", makeGRPCError(2 * time.Nanosecond) },
			expectedData: "",
			wantErr:      true,
		},
		{
			name: "gRPC error (no context)",
			ctx:  noContext,
			retryConfig: &RetryConfig{
				MaxRetries: 3,
				MaxJitter:  time.Nanosecond,
				Delay:      time.Nanosecond,
			},
			callable:     func() (string, error) { return "", makeGRPCError(time.Nanosecond) },
			expectedData: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Retry(
				tt.ctx,
				tt.retryConfig,
				tt.callable,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %v", err, tt.wantErr)
			}

			if data != tt.expectedData {
				t.Errorf("got data = %v, want data = %v", data, tt.expectedData)
			}
		})
	}
}
