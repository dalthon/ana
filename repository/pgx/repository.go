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

type PgxRepository struct {
	pool *pgxpool.Pool
}

func NewPgxRepository(pool *pgxpool.Pool) *PgxRepository {
	return &PgxRepository{pool: pool}
}

func (repo *PgxRepository) FetchOrStart(operation im.Operation[*PgxPayload, PgxResult, *PgxContext]) *im.TrackedOperation[*PgxPayload, PgxResult] {
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

	return rowsToTrackedOperation(rows)
}

func (repo *PgxRepository) NewSession(operation im.Operation[*PgxPayload, PgxResult, *PgxContext]) *im.Session[*PgxPayload, PgxResult, *PgxContext] {
	context := context.Background()
	outerTx, _ := repo.pool.Begin(context)
	tx, _ := outerTx.Begin(context)

	tx.Exec(context, lockTrackOperationQuery, pgx.NamedArgs{
		"key":            operation.Key(),
		"target":         operation.Target(),
		"reference_time": operation.ReferenceTime(),
	})

	return im.NewSession(operation, NewPgxContext(outerTx, tx, context))
}
