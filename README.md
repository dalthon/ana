<h1 align="center">
  <img src="https://raw.githubusercontent.com/dalthon/ana/master/doc/images/a_power_n_equals_a.svg" alt="A^n=A"/>
</h1>

[API Reference][api-reference]

This project is supposed to be a generic and very customizable
idempotency utility.

Right now we support only [fiber web framework][fiber] with Postgres as
persistence using [pgx v5][pgx].

I will be very pleased to receive pull requests to support other persistences
and frameworks.

## Install

```sh
go get github.com/dalthon/ana
```

## Usage

The simplest way of using it is shown by [examples/02-fiber.go][example] which
you can run with `make example-02`:

```go
package main

import (
	"context"
	"fmt"
	"os"

	a "github.com/dalthon/ana"
	r "github.com/dalthon/ana/repository/pgx"
	af "github.com/dalthon/ana/web/fiber"

	f "github.com/gofiber/fiber/v2"
	fr "github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/jackc/pgx/v5/pgxpool"
)

type idCtx = r.PgxContext[af.HttpPayload, af.HttpResponse]

func main() {
	pool := newPool()
	repo := r.NewPgxRepository[af.HttpPayload, af.HttpResponse](pool)
	ana := a.New(repo)
	middleware := af.New(ana, &af.Config{})

	app := f.New()
	app.Use(fr.New())

	app.Get("/*", middleware.Call(idempotentHandler, nil))

	app.Listen(":3000")
}

func idempotentHandler(c *f.Ctx, i *idCtx) (*af.HttpResponse, error) {
	fmt.Println("Not persisted!")

	return &af.HttpResponse{
		f.StatusOK,
		fmt.Sprintf("Hello %s!\n", c.Get("X-Idempotency-Key")),
	}, nil
}

func newPool() *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	return pool
}
```

In that example we apply idempotency on any `GET` route with default config.

Default config expects to get idempotency key from HTTP header
`X-Idempotency-Key` and also expects three more HTTP headers:
`X-Idempotency-Reference-Time`, `X-Idempotency-Timeout` and
`X-Idempotency-Expiration`.

All expected headers means:
* `X-Idempotency-Key`: Idempotency key provided by client
* `X-Idempotency-Reference-Time`: Reference time that will be considered from
first attempt to process a given operation (format [RFC 3339][rfc-time])
* `X-Idempotency-Timeout`: Timeout, in duration, that will not allow more than
one simultaneous attempt of operation (format from [time.ParseDuration][duration])
* `X-Idempotency-Expiration`: Timeout, in duration, that will not try to
execute a given operation again (format from [time.ParseDuration][duration])

```sh
  curl \
    -H "X-Idempotency-Key: some-key-as-string" \
    -H "X-Idempotency-Reference-Time: 1835-09-20T09:00:00Z" \
    -H "X-Idempotency-Timeout: 10s" \
    -H "X-Idempotency-Expiration: 24h" \
    http://localhost:3000/whatever-you-may-like
```

On `middleware.Call`, we can use an `*af.Config` to override given config at
middleware initialization.

Config is defined as:

```go
type Config struct {
	Key           func(*fiber.Ctx) string
	Target        func(*fiber.Ctx) string
	Payload       func(*fiber.Ctx) *HttpPayload
	ReferenceTime func(*fiber.Ctx) time.Time
	Timeout       func(*fiber.Ctx) time.Duration
	Expiration    func(*fiber.Ctx) time.Duration
}
```

Which are functions that returns value:
* `Key`: used as idempotency key
* `Target`: used to identify target of operation 
* `Payload`: that stores HTTP path and body of a given request
* `ReferenceTime`: that is considered the first possible time that this
operation should have started
* `Timeout`: that ensures that no more than one operation of same `Key` and
`Target` will run simultaneously
* `Expiration`: used as duration which after `ReferenceTime` + `Duration` we
assume this execution should not execute again.

Also there is a nice helper which in that example could be invoked as
`af.Value` to set configs as this:

```go
&af.Config{
  Timeout: af.Value(10 * time.Second)
}
```

So, this helper is just a function that returns a function that always returns
the same value. This seems silly, but is quite useful to have fixed configs for
`Timeout` and `Expiration`.

## TODOs

* Chores:
  * Add a few more tests on postgres repository showing that transactions are
  really rolling back everything done by user in case of failure.
  * Add fiber's middleware tests.
* Features:
  * Add `FinishedAt` and `ErrorCount` on `TrackedOperation`.
  * Add `net/http` middleware.
  * Add Redis persistence.
  * On Postgres repository, add config to store Response in Redis instead
  of Postgres.
  * Consider timeout to add statement timeout on session.

## Contributing

Pull requests and issues are welcome! I'll try to review them as soon as I can.

This project is quite simple and its [Makefile][makefile] is quite useful to do
whatever you may need. Run `make help` for more info.

To run tests, run `make test`.

To run test with coverage, run `make cover`.

To run a full featured example available at [examples/02-fiber.go][example], run
`make example-02`.

## License

This project is released under the [MIT License][license]

[api-reference]: https://pkg.go.dev/github.com/dalthon/ana
[duration]:      https://pkg.go.dev/time#ParseDuration
[example]:       examples/02-fiber.go
[fiber]:         https://gofiber.io/
[license]:       https://opensource.org/licenses/MIT
[makefile]:      Makefile
[rfc-time]:      https://www.rfc-editor.org/rfc/rfc3339.html
[pgx]:           https://github.com/jackc/pgx
