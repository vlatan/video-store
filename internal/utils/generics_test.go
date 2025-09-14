package utils

import (
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
			name:          "non-grpc error",
			err:           errors.New("regular error"),
			expectedDelay: 0,
			expectedOk:    false,
		},
		{
			name:          "RESOURCE_EXHAUSTED without RetryInfo",
			err:           status.Error(codes.ResourceExhausted, "rate limited"),
			expectedDelay: 0,
			expectedOk:    false,
		},
		{
			name:          "RESOURCE_EXHAUSTED with RetryInfo",
			err:           makeGRPCError(5 * time.Second),
			expectedDelay: 5 * time.Second,
			expectedOk:    true,
		},
		{
			name:          "RESOURCE_EXHAUSTED with RetryInfo (zero delay)",
			err:           makeGRPCError(0),
			expectedDelay: 0,
			expectedOk:    true,
		},
		{
			name:          "RESOURCE_EXHAUSTED with RetryInfo (large delay)",
			err:           makeGRPCError(30 * time.Minute),
			expectedDelay: 30 * time.Minute,
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
