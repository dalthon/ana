package idempotency_manager

import (
	"time"
)

type Manager[P any, R any, C any] struct {
	repository IdempotencyRepository[P, R, C]
}

func New[P any, R any, C any](repository IdempotencyRepository[P, R, C]) *Manager[P, R, C] {
	return &Manager[P, R, C]{repository: repository}
}

func (manager *Manager[P, R, C]) Call(operation Operation[P, R, C]) (*R, error) {
	if manager.isExpiredOperation(operation) {
		return nil, newExpirationError(operation.Target(), operation.Key())
	}

	trackedOperation := manager.repository.FetchOrStart(operation)

	if trackedOperation.isFinished() {
		return trackedOperation.Result, nil
	}

	if trackedOperation.isExpired() {
		return nil, newExpirationError(trackedOperation.Target, trackedOperation.Key)
	}

	if trackedOperation.stillRunning() {
		return nil, newStillRunningError(trackedOperation.Target, trackedOperation.Key)
	}

	return manager.callOperation(operation)
}

func (manager *Manager[P, R, C]) callOperation(operation Operation[P, R, C]) (*R, error) {
	session := manager.repository.NewSession(operation)
	defer session.Close()

	callWithinSession(session, operation)

	return session.Result(), session.Err()
}

func (manager *Manager[P, R, C]) isExpiredOperation(operation Operation[P, R, C]) bool {
	if operation.Expiration() == time.Duration(0) {
		return false
	}

	return time.Now().After(
		operation.ReferenceTime().Add(operation.Expiration()),
	)
}

func callWithinSession[P any, R any, C any](session Session[R, C], operation Operation[P, R, C]) {
	defer recoverSession(session)

	result, err := operation.Call(session.Ctx())
	session.SetResult(result)
	session.SetErr(err)
}

func recoverSession[R any, C any](session Session[R, C]) {
	if recovery := recover(); recovery != nil {
		session.SetErr(newPanicError(recovery))
	}
}
