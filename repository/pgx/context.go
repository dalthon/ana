package pgx

import (
	"context"
	"errors"
	"time"

	a "github.com/dalthon/ana"
	pgx "github.com/jackc/pgx/v5"
)

var finishTrackedOperationQuery string = `
  UPDATE ana.tracked_operations
  SET
    payload       = @payload,
    result        = @result,
    finished_at   = NOW(),
    status        = 'finished',
    error_message = NULL
  WHERE
    key = @key AND target = @target;
`

var failTrackedOperationQuery string = `
  UPDATE ana.tracked_operations
  SET
    payload       = @payload,
    result        = NULL,
    finished_at   = NOW(),
    status        = 'failed',
    timeout       = NOW(),
    error_message = @error_message,
    error_count   = error_count + 1
  WHERE
    key = @key AND target = @target;
`

type PgxContext[P any, R any] struct {
	outerTx pgx.Tx
	Tx      pgx.Tx
	Context context.Context
}

func NewPgxContext[P any, R any](outerTx pgx.Tx, tx pgx.Tx, context context.Context) *PgxContext[P, R] {
	return &PgxContext[P, R]{outerTx: outerTx, Tx: tx, Context: context}
}

func (ctx *PgxContext[P, R]) Success(operation *a.TrackedOperation[P, R]) {
	if !operation.Expiration.IsZero() && time.Now().After(operation.Expiration) {
		operation.Err = errors.New("Operation expired")
		ctx.Fail(operation)
		return
	}

	if commitErr := ctx.Tx.Commit(ctx.Context); commitErr != nil {
		panic(commitErr)
	}

	_, err := ctx.outerTx.Exec(
		ctx.Context,
		finishTrackedOperationQuery,
		pgx.NamedArgs{
			"key":     operation.Key,
			"target":  operation.Target,
			"payload": serialize(operation.Payload),
			"result":  serialize(operation.Result),
		},
	)

	if err != nil {
		ctx.outerTx.Rollback(ctx.Context)
		panic(err)
	}

	if commitErr := ctx.outerTx.Commit(ctx.Context); commitErr != nil {
		panic(commitErr)
	}
}

func (ctx *PgxContext[P, R]) Fail(operation *a.TrackedOperation[P, R]) {
	if commitErr := ctx.Tx.Rollback(ctx.Context); commitErr != nil {
		panic(commitErr)
	}

	_, err := ctx.outerTx.Exec(
		ctx.Context,
		failTrackedOperationQuery,
		pgx.NamedArgs{
			"key":           operation.Key,
			"target":        operation.Target,
			"payload":       serialize(operation.Payload),
			"error_message": operation.Err.Error(),
		},
	)

	if err != nil {
		panic(err)
	}

	if commitErr := ctx.outerTx.Commit(ctx.Context); commitErr != nil {
		panic(commitErr)
	}
}
