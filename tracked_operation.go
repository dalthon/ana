package idempotency_manager

import (
	"time"
)

type TrackedOperationStatus uint64

const (
	Ready TrackedOperationStatus = iota
	Running
	Finished
	Failed
)

type TrackedOperation[P any, R any] struct {
	Status        TrackedOperationStatus
	Key           string
	Target        string
	Payload       P
	ReferenceTime time.Time
	StartedAt     time.Time
	Timeout       time.Time
	Expiration    time.Time
	Result        *R
	Err           error
}

func NewTrackedOperation[P any, R any](
	status TrackedOperationStatus,
	key string,
	target string,
	payload P,
	referenceTime time.Time,
	startedAt time.Time,
	timeout time.Time,
	expiration time.Time,
	result *R,
	err error,
) *TrackedOperation[P, R] {
	return &TrackedOperation[P, R]{
		Status:        status,
		Key:           key,
		Target:        target,
		Payload:       payload,
		ReferenceTime: referenceTime,
		StartedAt:     startedAt,
		Timeout:       timeout,
		Expiration:    expiration,
		Result:        result,
		Err:           err,
	}
}

func (operation *TrackedOperation[P, R]) isFinished() bool {
	return operation.Status == Finished
}

func (operation *TrackedOperation[P, R]) isExpired() bool {
	return operation.Expiration != time.Time{} && time.Now().After(operation.Expiration)
}

func (operation *TrackedOperation[P, R]) stillRunning() bool {
	return operation.Status == Running && (operation.Timeout == time.Time{} || operation.Timeout.After(time.Now()))
}
