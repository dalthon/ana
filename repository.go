package ana

type IdempotencyRepository[P any, R any, C SessionCtx[P, R]] interface {
	FetchOrStart(Operation[P, R, C]) *TrackedOperation[P, R]
	NewSession(Operation[P, R, C]) *Session[P, R, C]
}
