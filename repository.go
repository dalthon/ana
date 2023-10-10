package idempotency_manager

type IdempotencyRepository[P any, R any, C any] interface {
	FetchOrStart(Operation[P, R, C]) *TrackedOperation[P, R]
	Call(Operation[P, R, C]) (*R, error)
}
