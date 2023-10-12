package pgx

import (
	"errors"
	"time"
)

type debugPayload struct {
	Value string
}

type debugResult struct {
	Value string
}

type mockedOperation struct {
	key           string
	target        string
	payload       *debugPayload
	referenceTime time.Time
	timeout       time.Duration
	expiration    time.Duration
	result        string
	success       bool
}

func newMockedOperation(key, target, payload, result string, success bool) *mockedOperation {
	now := time.Now()

	return &mockedOperation{
		key:        key,
		target:     target,
		payload:    &debugPayload{payload},
		result:     result,
		success:    success,
		timeout:    1 * time.Minute,
		expiration: 1 * time.Minute,
		referenceTime: time.Date(
			now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC,
		),
	}
}

func (o *mockedOperation) Key() string {
	return o.key
}

func (o *mockedOperation) Target() string {
	return o.target
}

func (o *mockedOperation) Payload() *debugPayload {
	return o.payload
}

func (o *mockedOperation) ReferenceTime() time.Time {
	return o.referenceTime
}

func (o *mockedOperation) Timeout() time.Duration {
	return o.timeout
}

func (o *mockedOperation) Expiration() time.Duration {
	return o.expiration
}

func (o *mockedOperation) Call(ctx *PgxContext[debugPayload, debugResult]) (*debugResult, error) {
	if o.success {
		return &debugResult{o.result}, nil
	}

	return nil, errors.New(o.result)
}
