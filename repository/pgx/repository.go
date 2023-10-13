package pgx

import (
	"context"

	a "github.com/dalthon/ana"
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
  FROM ana.fetch_or_start(
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
  FROM ana.tracked_operations
  WHERE
    key            = @key    AND
    target         = @target AND
    reference_time = @reference_time
  FOR UPDATE;
`

var failTimedOutStillRunningQuery string = `
  UPDATE ana.tracked_operations AS operation
  SET
    status        = 'failed',
    finished_at   = NOW(),
    error_count   = error_count + 1,
    error_message = 'Operation timed out'
  FROM (
    SELECT key, target
    FROM ana.tracked_operations AS t
    WHERE
      t.status = 'running' AND
      t.timeout < NOW()
    ORDER BY t.timeout ASC
    LIMIT @count
    FOR UPDATE SKIP LOCKED
  ) AS expired
  WHERE operation.key = expired.key AND operation.target = expired.target;
`

var failExpiredStillRunningQuery string = `
  UPDATE ana.tracked_operations AS operation
  SET
    status        = 'failed',
    finished_at   = NOW(),
    error_count   = error_count + 1,
    error_message = 'Operation expired'
  FROM (
    SELECT key, target
    FROM ana.tracked_operations AS t
    WHERE
      t.status = 'running' AND
      t.expiration < NOW()
    ORDER BY t.expiration ASC
    LIMIT @count
    FOR UPDATE SKIP LOCKED
  ) AS expired
  WHERE operation.key = expired.key AND operation.target = expired.target;
`

var deleteExpiredQuery string = `
  DELETE FROM ana.tracked_operations AS operation
  USING (
    SELECT key, target
    FROM ana.tracked_operations AS t
    WHERE
      t.status = @status AND
      t.expiration < NOW()
    ORDER BY t.expiration ASC
    LIMIT @count
    FOR UPDATE SKIP LOCKED
  ) AS expired
  WHERE operation.key = expired.key AND operation.target = expired.target;
`

type PgxRepository[P any, R any] struct {
	pool *pgxpool.Pool
}

func NewPgxRepository[P any, R any](pool *pgxpool.Pool) *PgxRepository[P, R] {
	return &PgxRepository[P, R]{pool: pool}
}

func (repo *PgxRepository[P, R]) FetchOrStart(operation a.Operation[P, R, *PgxContext[P, R]]) *a.TrackedOperation[P, R] {
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

func (repo *PgxRepository[P, R]) NewSession(operation a.Operation[P, R, *PgxContext[P, R]]) *a.Session[P, R, *PgxContext[P, R]] {
	context := context.Background()
	outerTx, _ := repo.pool.Begin(context)
	tx, _ := outerTx.Begin(context)

	tx.Exec(context, lockTrackOperationQuery, pgx.NamedArgs{
		"key":            operation.Key(),
		"target":         operation.Target(),
		"reference_time": operation.ReferenceTime(),
	})

	return a.NewSession(operation, NewPgxContext[P, R](outerTx, tx, context))
}

func (repo *PgxRepository[P, R]) FailTimedOutStillRunning(count int) int64 {
	info, err := repo.pool.Exec(
		context.Background(),
		failTimedOutStillRunningQuery,
		pgx.NamedArgs{"count": count},
	)

	if err != nil {
		panic(err)
	}

	return info.RowsAffected()
}

func (repo *PgxRepository[P, R]) FailExpiredStillRunning(count int) int64 {
	info, err := repo.pool.Exec(
		context.Background(),
		failExpiredStillRunningQuery,
		pgx.NamedArgs{"count": count},
	)

	if err != nil {
		panic(err)
	}

	return info.RowsAffected()
}

func (repo *PgxRepository[P, R]) DeleteExpired(status a.TrackedOperationStatus, count int) int64 {
	info, err := repo.pool.Exec(
		context.Background(),
		deleteExpiredQuery,
		pgx.NamedArgs{
			"status": trackedStatusToPgStatus(status),
			"count":  count,
		},
	)

	if err != nil {
		panic(err)
	}

	return info.RowsAffected()
}

func trackedStatusToPgStatus(status a.TrackedOperationStatus) string {
	switch status {
	case a.Ready:
		return "ready"
	case a.Running:
		return "running"
	case a.Finished:
		return "finished"
	case a.Failed:
		return "failed"
	default:
		panic("Dead code")
	}
}
