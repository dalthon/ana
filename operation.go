package idempotency_manager

import (
	"time"
)

type Operation[P any, R any, C SessionCtx[P, R]] interface {
	Key() string
	Target() string
	Payload() *P
	ReferenceTime() time.Time
	Timeout() time.Duration
	Expiration() time.Duration
	Call(C) (*R, error)
}
