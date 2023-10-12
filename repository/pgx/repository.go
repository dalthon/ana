package pgx

import (
	"context"

	im "github.com/dalthon/idempotency_manager"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var fetchOrStartQuery string = `
  SELECT
    status,
    key,
    target,
    payload,
    reference_time,
    started_at,
    timeout,
    expiration,
    result,
    error_message
  FROM idempotency.fetch_or_start(
    @key,
    @target,
    @payload,
    @reference_time,
    @timeout,
    @expiration
  );
`

var lockTrackOperationQuery string = `
  SELECT *
  FROM idempotency.tracked_operations
  WHERE
    key            = @key    AND
    target         = @target AND
    reference_time = @reference_time
  FOR UPDATE;
`

type PgxRepository[P any, R any] struct {
	pool *pgxpool.Pool
}

func NewPgxRepository[P any, R any](pool *pgxpool.Pool) *PgxRepository[P, R] {
	return &PgxRepository[P, R]{pool: pool}
}

func (repo *PgxRepository[P, R]) FetchOrStart(operation im.Operation[P, R, *PgxContext[P, R]]) *im.TrackedOperation[P, R] {
	rows, err := repo.pool.Query(
		context.Background(),
		fetchOrStartQuery,
		pgx.NamedArgs{
			"key":            operation.Key(),
			"target":         operation.Target(),
			"payload":        serialize(operation.Payload()),
			"reference_time": operation.ReferenceTime(),
			"timeout":        operation.Timeout(),
			"expiration":     operation.Expiration(),
		},
	)

	if err != nil {
		panic(err)
	}

	return rowsToTrackedOperation[P, R](rows)
}

func (repo *PgxRepository[P, R]) NewSession(operation im.Operation[P, R, *PgxContext[P, R]]) *im.Session[P, R, *PgxContext[P, R]] {
	context := context.Background()
	outerTx, _ := repo.pool.Begin(context)
	tx, _ := outerTx.Begin(context)

	tx.Exec(context, lockTrackOperationQuery, pgx.NamedArgs{
		"key":            operation.Key(),
		"target":         operation.Target(),
		"reference_time": operation.ReferenceTime(),
	})

	return im.NewSession(operation, NewPgxContext[P, R](outerTx, tx, context))
}
