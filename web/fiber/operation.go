package fiber

import (
	"fmt"
	"time"

	a "github.com/dalthon/ana"
	f "github.com/gofiber/fiber/v2"
)

type HttpOperation[C a.SessionCtx[HttpPayload, HttpResponse]] struct {
	fiberCtx *f.Ctx
	handler  func(*f.Ctx, C) (*HttpResponse, error)

	key           string
	target        string
	payload       *HttpPayload
	referenceTime time.Time
	timeout       time.Duration
	expiration    time.Duration
}

func newHttpOperation[C a.SessionCtx[HttpPayload, HttpResponse]](
	fiberCtx *f.Ctx,
	handler func(*f.Ctx, C) (*HttpResponse, error),
	specificConfig, sharedConfig *Config,
) *HttpOperation[C] {
	return &HttpOperation[C]{
		fiberCtx: fiberCtx,
		handler:  handler,

		key:           coalesceConfigCall(specificConfig.Key, sharedConfig.Key, DefaultKey, fiberCtx),
		target:        coalesceConfigCall(specificConfig.Target, sharedConfig.Target, DefaultTarget, fiberCtx),
		payload:       coalesceConfigCall(specificConfig.Payload, sharedConfig.Payload, DefaultPayload, fiberCtx),
		referenceTime: coalesceConfigCall(specificConfig.ReferenceTime, sharedConfig.ReferenceTime, DefaultReferenceTime, fiberCtx),
		timeout:       coalesceConfigCall(specificConfig.Timeout, sharedConfig.Timeout, DefaultTimeout, fiberCtx),
		expiration:    coalesceConfigCall(specificConfig.Expiration, sharedConfig.Expiration, DefaultExpiration, fiberCtx),
	}
}

func (o *HttpOperation[C]) Key() string {
	return o.key
}

func (o *HttpOperation[C]) Target() string {
	return o.target
}

func (o *HttpOperation[C]) Payload() *HttpPayload {
	return o.payload
}

func (o *HttpOperation[C]) ReferenceTime() time.Time {
	return o.referenceTime
}

func (o *HttpOperation[C]) Timeout() time.Duration {
	return o.timeout
}

func (o *HttpOperation[C]) Expiration() time.Duration {
	return o.expiration
}

func (o *HttpOperation[C]) Call(ctx C) (*HttpResponse, error) {
	return o.handler(o.fiberCtx, ctx)
}

func DefaultKey(fiberCtx *f.Ctx) string {
	return fiberCtx.Get("X-Idempotency-Key")
}

func DefaultTarget(fiberCtx *f.Ctx) string {
	return fmt.Sprintf("[%s]%s", fiberCtx.Method(), fiberCtx.Path())
}

func DefaultPayload(fiberCtx *f.Ctx) *HttpPayload {
	return &HttpPayload{
		fiberCtx.Method(),
		fiberCtx.OriginalURL(),
		fiberCtx.BodyRaw(),
	}
}

func DefaultReferenceTime(fiberCtx *f.Ctx) time.Time {
	referenceString := fiberCtx.Get("X-Idempotency-Reference-Time")

	referenceTime, err := time.Parse(time.RFC3339, referenceString)
	if err != nil {
		panic("Invalid X-Idempotency-Reference-Time")
	}

	return referenceTime
}

func DefaultTimeout(fiberCtx *f.Ctx) time.Duration {
	timeoutString := fiberCtx.Get("X-Idempotency-Timeout")

	timeoutDuration, err := time.ParseDuration(timeoutString)
	if err != nil {
		panic("Invalid X-Idempotency-Timeout")
	}

	return timeoutDuration
}

func DefaultExpiration(fiberCtx *f.Ctx) time.Duration {
	expirationString := fiberCtx.Get("X-Idempotency-Expiration")

	expirationDuration, err := time.ParseDuration(expirationString)
	if err != nil {
		panic("Invalid X-Idempotency-Expiration")
	}

	return expirationDuration
}

func coalesceConfigCall[R any, F func(*f.Ctx) R](specificFn, sharedFn, defaultFn F, fiberCtx *f.Ctx) R {
	if specificFn != nil {
		return specificFn(fiberCtx)
	}

	if sharedFn != nil {
		return sharedFn(fiberCtx)
	}

	return defaultFn(fiberCtx)
}
