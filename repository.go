package idempotency_manager

type IdempotencyRepository[P any, R any, C any] interface {
	FetchOrStart(Operation[P, R, C]) *TrackedOperation[P, R]
	NewSession(Operation[P, R, C]) Session[R, C]
}
