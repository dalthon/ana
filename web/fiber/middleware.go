package fiber

import (
	"time"

	a "github.com/dalthon/ana"
	f "github.com/gofiber/fiber/v2"
)

type HttpPayload struct {
	Method string
	Url    string
	Body   []byte
}

type HttpResponse struct {
	Status int
	Body   string
}

type Config struct {
	Key           func(*f.Ctx) string
	Target        func(*f.Ctx) string
	Payload       func(*f.Ctx) *HttpPayload
	ReferenceTime func(*f.Ctx) time.Time
	Timeout       func(*f.Ctx) time.Duration
	Expiration    func(*f.Ctx) time.Duration
}

func Value[V any](value V) func(*f.Ctx) V {
	return func(*f.Ctx) V { return value }
}

type Middleware[C a.SessionCtx[HttpPayload, HttpResponse]] struct {
	config *Config
	ana    *a.Manager[HttpPayload, HttpResponse, C]
}

func New[C a.SessionCtx[HttpPayload, HttpResponse]](
	ana *a.Manager[HttpPayload, HttpResponse, C],
	config *Config,
) *Middleware[C] {
	return &Middleware[C]{config, ana}
}

func (middleware *Middleware[C]) Call(idempotentHandler func(*f.Ctx, C) (*HttpResponse, error), config *Config) func(*f.Ctx) error {
	if config == nil {
		config = &Config{}
	}

	return func(c *f.Ctx) error {
		operation := newHttpOperation(c, idempotentHandler, config, middleware.config)

		result, err := middleware.ana.Call(operation)
		if err != nil {
			return err
		}

		c.Status(result.Status).SendString(result.Body)
		return nil
	}
}
