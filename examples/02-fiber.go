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
