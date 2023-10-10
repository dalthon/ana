package idempotency_manager

type SessionCtx[R any] interface {
	Success(*R)
	Fail(error)
}

type Session[P any, R any, C SessionCtx[R]] struct {
	operation Operation[P, R, C]
	context   C
	result    *R
	err       error
	closed    bool
}

func NewSession[P any, R any, C SessionCtx[R]](operation Operation[P, R, C], context C) *Session[P, R, C] {
	return &Session[P, R, C]{
		operation: operation,
		context:   context,
		closed:    false,
	}
}

func (session *Session[P, R, C]) call(operation Operation[P, R, C]) {
	defer session.recover()
	session.result, session.err = operation.Call(session.context)
}

func (session *Session[P, R, C]) recover() {
	if recovery := recover(); recovery != nil {
		session.err = newPanicError(recovery)
	}
}

func (session *Session[P, R, C]) close() {
	if session.closed {
		return
	}

	if session.err == nil {
		session.context.Success(session.result)
		session.closed = true
		return
	}

	session.context.Fail(session.err)
	session.closed = true
}
