package idempotency_manager

type mockedSession struct {
	operation Operation[string, mockedResult, mockedCtx]
	ctx       mockedCtx
	result    *mockedResult
	err       error
}

func newMockedSession(operation Operation[string, mockedResult, mockedCtx]) *mockedSession {
	return &mockedSession{operation: operation}
}

func (session *mockedSession) Close() {
}

func (session *mockedSession) Ctx() *mockedCtx {
	return &session.ctx
}

func (session *mockedSession) SetResult(result *mockedResult) {
	session.result = result
}

func (session *mockedSession) Result() *mockedResult {
	return session.result
}

func (session *mockedSession) SetErr(err error) {
	session.err = err
}

func (session *mockedSession) Err() error {
	return session.err
}
