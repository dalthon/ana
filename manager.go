package idempotency_manager

import "time"

type Manager[P any, R any, C SessionCtx[P, R]] struct {
	repository IdempotencyRepository[P, R, C]
}

func New[P any, R any, C SessionCtx[P, R]](repository IdempotencyRepository[P, R, C]) *Manager[P, R, C] {
	return &Manager[P, R, C]{repository: repository}
}

func (manager *Manager[P, R, C]) Call(operation Operation[P, R, C]) (*R, error) {
	if manager.isExpiredOperation(operation) {
		return nil, newExpirationError(operation.Target(), operation.Key())
	}

	trackedOperation := manager.repository.FetchOrStart(operation)

	if trackedOperation != nil {
		if trackedOperation.isFinished() {
			return trackedOperation.Result, nil
		}

		if trackedOperation.isExpired() {
			return nil, newExpirationError(trackedOperation.Target, trackedOperation.Key)
		}

		if trackedOperation.stillRunning() {
			return nil, newStillRunningError(trackedOperation.Target, trackedOperation.Key)
		}
	}

	return manager.callOperation(operation)
}

func (manager *Manager[P, R, C]) callOperation(operation Operation[P, R, C]) (*R, error) {
	session := manager.repository.NewSession(operation)
	defer session.close()

	session.call(operation)

	return session.result, session.err
}

func (manager *Manager[P, R, C]) isExpiredOperation(operation Operation[P, R, C]) bool {
	if operation.Expiration() == time.Duration(0) {
		return false
	}

	return time.Now().After(
		operation.ReferenceTime().Add(operation.Expiration()),
	)
}
