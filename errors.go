package ana

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

type PanicError struct {
	err interface{}
}

func newPanicError(err interface{}) *PanicError {
	return &PanicError{err: err}
}

func (err *PanicError) Error() string {
	return fmt.Sprintf("Got panic \"%v\"", err.err)
}
