package rdb

import "fmt"

type LockError struct {
	Msg string
	Err error // Original error if any
}

func (e *LockError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

// Allows errors.Is and errors.As to inspect the underlying error
func (e *LockError) Unwrap() error {
	return e.Err
}
