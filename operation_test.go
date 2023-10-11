package idempotency_manager

import (
	"errors"
	"time"
)

type mockedOperation struct {
	key           string
	target        string
	payload       string
	referenceTime time.Time
	timeout       time.Duration
	expiration    time.Duration
	result        func() (*mockedResult, error)
}

func newMockedOperation(
	key string,
	target string,
	payload string,
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
	return operation.result()
}

type mockedResult struct {
	result string
}

func newMockedResult(result string) *mockedResult {
	return &mockedResult{result: result}
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

type mockedCtx struct{}

func newMockedCtx() *mockedCtx {
	return &mockedCtx{}
}

func (ctx *mockedCtx) Success(*TrackedOperation[string, mockedResult]) {
}

func (ctx *mockedCtx) Fail(*TrackedOperation[string, mockedResult]) {
}
