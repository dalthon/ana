package idempotency_manager

import "fmt"

type ExpirationError struct {
	target string
	key    string
}

func newExpirationError(target string, key string) *ExpirationError {
	return &ExpirationError{target: target, key: key}
}

func (err *ExpirationError) Error() string {
	return fmt.Sprintf("Operation %v expired for key %v.", err.target, err.key)
}

type StillRunningError struct {
	target string
	key    string
}

func newStillRunningError(target string, key string) *StillRunningError {
	return &StillRunningError{target: target, key: key}
}

func (err *StillRunningError) Error() string {
	return fmt.Sprintf("Operation %v still running for key %v.", err.target, err.key)
}
