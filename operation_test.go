package idempotency_manager

import (
	"errors"
	"time"
)

type mockedOperation struct {
	key           string
	target        string
	payload       *mockedPayload
	referenceTime time.Time
	timeout       time.Duration
	expiration    time.Duration
	result        func() (*mockedResult, error)
}

func newMockedOperation(
	key string,
	target string,
	payload *mockedPayload,
	referenceTime time.Time,
	timeout time.Duration,
	expiration time.Duration,
	result func() (*mockedResult, error),
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

func (operation *mockedOperation) Payload() *mockedPayload {
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
	return operation.result()
}

type mockedPayload struct {
	payload string
}

func newMockedPayload(payload string) *mockedPayload {
	return &mockedPayload{payload}
}

type mockedResult struct {
	result string
}

func newMockedResult(result string) *mockedResult {
	return &mockedResult{result}
}

func newMockedResultFn(result string) func() (*mockedResult, error) {
	return func() (*mockedResult, error) {
		return newMockedResult(result), nil
	}
}

func newMockedErrorFn(message string) func() (*mockedResult, error) {
	return func() (*mockedResult, error) {
		return nil, errors.New(message)
	}
}

func newMockedPanicFn(message string) func() (*mockedResult, error) {
	return func() (*mockedResult, error) {
		panic(message)
	}
}

type mockedCtx struct {
	SuccessCount uint
	FailCount    uint
}

func newMockedCtx() *mockedCtx {
	return &mockedCtx{0, 0}
}

func (ctx *mockedCtx) Success(*TrackedOperation[mockedPayload, mockedResult]) {
	ctx.SuccessCount += 1
}

func (ctx *mockedCtx) Fail(*TrackedOperation[mockedPayload, mockedResult]) {
	ctx.FailCount += 1
}
