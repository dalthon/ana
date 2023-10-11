package idempotency_manager

import "time"

type mockedOperation struct {
	key           string
	target        string
	payload       string
	referenceTime time.Time
	timeout       time.Duration
	expiration    time.Duration
	result        *mockedResult
}

func newMockedOperation(
	key string,
	target string,
	payload string,
	referenceTime time.Time,
	timeout time.Duration,
	expiration time.Duration,
	result *mockedResult,
) *mockedOperation {
	return &mockedOperation{
		key:           key,
		target:        target,
		payload:       payload,
		referenceTime: referenceTime,
		timeout:       timeout,
		expiration:    expiration,
		result:        result,
	}
}

func (operation *mockedOperation) Key() string {
	return operation.key
}

func (operation *mockedOperation) Target() string {
	return operation.target
}

func (operation *mockedOperation) Payload() string {
	return operation.payload
}

func (operation *mockedOperation) ReferenceTime() time.Time {
	return operation.referenceTime
}

func (operation *mockedOperation) Timeout() time.Duration {
	return operation.timeout
}

func (operation *mockedOperation) Expiration() time.Duration {
	return operation.expiration
}

func (operation *mockedOperation) Call(ctx *mockedCtx) (*mockedResult, error) {
	return operation.result, nil
}

type mockedResult struct {
	result string
}

func newMockedResult(result string) *mockedResult {
	return &mockedResult{result: result}
}

type mockedCtx struct{}

func newMockedCtx() *mockedCtx {
	return &mockedCtx{}
}

func (ctx *mockedCtx) Success(*TrackedOperation[string, mockedResult]) {
}

func (ctx *mockedCtx) Fail(*TrackedOperation[string, mockedResult]) {
}
