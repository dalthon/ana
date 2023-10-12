package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	im "github.com/dalthon/idempotency_manager"
	r "github.com/dalthon/idempotency_manager/repository/pgx"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool := newPool()
	repo := r.NewPgxRepository(pool)

	manager := im.New(repo)
	operation := newOperation()
	result, err := manager.Call(operation)

	fmt.Println("OPERATION")
	fmt.Println("Key:", operation.Key())
	fmt.Println("Target:", operation.Target())
	fmt.Println("Payload:", operation.Payload())
	fmt.Println("ReferenceTime:", operation.ReferenceTime())
	fmt.Println("Timeout:", operation.Timeout())
	fmt.Println("Expiration:", operation.Expiration())

	fmt.Println("\nResult:", result)
	fmt.Println("Error:", err)
}

func newPool() *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	return pool
}

type operation struct {
	key           string
	target        string
	payload       *r.PgxPayload
	referenceTime time.Time
	timeout       time.Duration
	expiration    time.Duration
}

func newOperation() *operation {
	now := time.Now()

	key := "default key"
	if len(os.Args) > 1 {
		key = os.Args[1]
	}

	target := "default target"
	if len(os.Args) > 2 {
		target = os.Args[2]
	}

	payloadValue := "default payload value"
	if len(os.Args) > 3 {
		payloadValue = os.Args[3]
	}

	return &operation{
		key:        key,
		target:     target,
		payload:    &r.PgxPayload{payloadValue, 1},
		timeout:    1 * time.Minute,
		expiration: 1 * time.Minute,
		referenceTime: time.Date(
			now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC,
		),
	}
}

func (o *operation) Key() string {
	return o.key
}

func (o *operation) Target() string {
	return o.target
}

func (o *operation) Payload() *r.PgxPayload {
	return o.payload
}

func (o *operation) ReferenceTime() time.Time {
	return o.referenceTime
}

func (o *operation) Timeout() time.Duration {
	return o.timeout
}

func (o *operation) Expiration() time.Duration {
	return o.expiration
}

func (o *operation) Call(ctx *r.PgxContext) (*r.PgxResult, error) {
	fmt.Println("Not already processed")

	if len(os.Args) > 4 {
		return nil, errors.New(os.Args[4])
	}

	return &r.PgxResult{o.Payload().Str, 42}, nil
}
