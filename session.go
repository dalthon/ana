package idempotency_manager

import "time"

type SessionCtx[P any, R any] interface {
	Success(*TrackedOperation[P, R])
	Fail(*TrackedOperation[P, R])
}

// TODO: Add some tests at session_test.go
type Session[P any, R any, C SessionCtx[P, R]] struct {
	operation Operation[P, R, C]
	context   C
	startedAt time.Time
	result    *R
	err       error
	closed    bool
}

func NewSession[P any, R any, C SessionCtx[P, R]](operation Operation[P, R, C], context C) *Session[P, R, C] {
	return &Session[P, R, C]{
		operation: operation,
		context:   context,
		closed:    false,
	}
}

func (session *Session[P, R, C]) call(operation Operation[P, R, C]) {
	defer session.recover()
	session.startedAt = time.Now()
	session.result, session.err = operation.Call(session.context)
}

func (session *Session[P, R, C]) recover() {
	if recovery := recover(); recovery != nil {
		session.err = newPanicError(recovery)
	}
}

func (session *Session[P, R, C]) trackedOperation() *TrackedOperation[P, R] {
	timeout := time.Time{}
	expiration := time.Time{}

	if session.operation.Timeout() != time.Duration(0) {
		timeout = session.operation.ReferenceTime().Add(session.operation.Timeout())
	}

	if session.operation.Expiration() != time.Duration(0) {
		expiration = session.operation.ReferenceTime().Add(session.operation.Expiration())
	}

	status := Finished
	if session.err != nil {
		status = Failed
	}

	return NewTrackedOperation(
		status,
		session.operation.Key(),
		session.operation.Target(),
		session.operation.Payload(),
		session.operation.ReferenceTime(),
		session.startedAt,
		timeout,
		expiration,
		session.result,
		session.err,
	)
}

func (session *Session[P, R, C]) close() {
	if session.closed {
		return
	}

	if session.err == nil {
		session.context.Success(session.trackedOperation())
		session.closed = true
		return
	}

	session.context.Fail(session.trackedOperation())
	session.closed = true
}
