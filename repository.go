package idempotency_manager

type IdempotencyRepository[P any, R any, C SessionCtx[R]] interface {
	FetchOrStart(Operation[P, R, C]) *TrackedOperation[P, R]
	NewSession(Operation[P, R, C]) *Session[P, R, C]
}
