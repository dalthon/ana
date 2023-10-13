package pgx

import (
	"context"
	"errors"
	"os"
	"time"

	a "github.com/dalthon/ana"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"testing"
)

func TestPgxRepositoryFetchOrStart(t *testing.T) {
	pool := newPool()
	clearDatabase(pool)

	repo := NewPgxRepository[debugPayload, debugResult](pool)
	operation := newMockedOperation("key", "target", "payload", "result", true)

	trackedOperation := repo.FetchOrStart(operation)
	assertEqual(t, trackedOperation.Status, a.Ready)
	assertEqual(t, trackedOperation.Key, operation.Key())
	assertEqual(t, trackedOperation.Target, operation.Target())
	assertEqual(t, trackedOperation.Payload.Value, operation.Payload().Value)
	assertTimeEqual(t, trackedOperation.ReferenceTime, operation.ReferenceTime())
	assertNil(t, trackedOperation.Result)
	assertErrorNil(t, trackedOperation.Err)

	anotherTrackedOperation := repo.FetchOrStart(operation)
	assertEqual(t, anotherTrackedOperation.Status, a.Running)
	assertEqual(t, anotherTrackedOperation.Key, operation.Key())
	assertEqual(t, anotherTrackedOperation.Target, operation.Target())
	assertEqual(t, anotherTrackedOperation.Payload.Value, operation.Payload().Value)
	assertTimeEqual(t, anotherTrackedOperation.ReferenceTime, operation.ReferenceTime())
	assertNil(t, anotherTrackedOperation.Result)
	assertErrorNil(t, anotherTrackedOperation.Err)

	assertEqual(t, trackedOperation.StartedAt, anotherTrackedOperation.StartedAt)
	assertEqual(t, trackedOperation.Timeout, anotherTrackedOperation.Timeout)
	assertEqual(t, trackedOperation.Expiration, anotherTrackedOperation.Expiration)
}

func TestPgxContextSuccess(t *testing.T) {
	pool := newPool()
	clearDatabase(pool)

	repo := NewPgxRepository[debugPayload, debugResult](pool)
	operation := newMockedOperation("key", "target", "payload", "result", true)

	trackedOperation := repo.FetchOrStart(operation)
	assertEqual(t, trackedOperation.Status, a.Ready)

	trackedOperation.Result = &debugResult{"result"}
	session := repo.NewSession(operation)
	session.Context.Success(trackedOperation)

	refreshedOperation := repo.FetchOrStart(operation)
	assertEqual(t, refreshedOperation.Status, a.Finished)
	assertEqual(t, refreshedOperation.Result.Value, "result")
	assertErrorNil(t, refreshedOperation.Err)
}

func TestPgxContextFail(t *testing.T) {
	pool := newPool()
	clearDatabase(pool)

	repo := NewPgxRepository[debugPayload, debugResult](pool)
	operation := newMockedOperation("key", "target", "payload", "result", true)

	trackedOperation := repo.FetchOrStart(operation)
	assertEqual(t, trackedOperation.Status, a.Ready)

	trackedOperation.Err = errors.New("Something went wrong")
	session := repo.NewSession(operation)
	session.Context.Fail(trackedOperation)

	refreshedOperation := repo.FetchOrStart(operation)
	assertEqual(t, refreshedOperation.Status, a.Failed)
	assertNil(t, refreshedOperation.Result)
	assertEqual(t, refreshedOperation.Err.Error(), trackedOperation.Err.Error())
}

func TestPgxContextFailOnSuccess(t *testing.T) {
	pool := newPool()
	clearDatabase(pool)

	repo := NewPgxRepository[debugPayload, debugResult](pool)
	operation := newMockedOperation("key", "target", "payload", "result", true)

	trackedOperation := repo.FetchOrStart(operation)
	assertEqual(t, trackedOperation.Status, a.Ready)

	trackedOperation.Expiration = time.Now().Add(-10 * time.Second)
	session := repo.NewSession(operation)
	session.Context.Success(trackedOperation)

	refreshedOperation := repo.FetchOrStart(operation)
	assertEqual(t, refreshedOperation.Status, a.Failed)
	assertNil(t, refreshedOperation.Result)
	assertEqual(t, refreshedOperation.Err.Error(), "Operation expired")
}

func TestRepositoryDeleteExpired(t *testing.T) {
	pool := newPool()
	clearDatabase(pool)

	pool.Exec(
		context.Background(),
		`
    INSERT INTO ana.tracked_operations (
      status,       key,            target,           payload, reference_time,                timeout,                       expiration,                    started_at
    ) VALUES
      ('finished', 'expired_key',  'finished_target', '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'expired_key',  'failed_target',   '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'expired_key',  'running_target',  '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('finished', 'finished_key', 'finished_target', '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'failed_key',   'failed_target',   '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'running_key',  'running_target',  '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('finished', 'expired_key',  'finished_again',  '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'expired_key',  'failed_again',    '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'expired_key',  'running_again',   '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('finished', 'finished_key', 'finished_again',  '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'failed_key',   'failed_again',    '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'running_key',  'running_again',   '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval)
    `,
		pgx.NamedArgs{},
	)

	repo := NewPgxRepository[debugPayload, debugResult](pool)
	assertEqual(t, int64(2), repo.DeleteExpired(a.Finished, 10))
	assertEqual(t, int64(0), repo.DeleteExpired(a.Finished, 10))
	assertEqual(t, int64(1), repo.DeleteExpired(a.Running, 1))
	assertEqual(t, int64(1), repo.DeleteExpired(a.Running, 1))
	assertEqual(t, int64(0), repo.DeleteExpired(a.Running, 1))
	assertEqual(t, int64(2), repo.DeleteExpired(a.Failed, 2))
	assertEqual(t, int64(0), repo.DeleteExpired(a.Failed, 2))
}

func TestRepositoryFailExpiredStillRunning(t *testing.T) {
	pool := newPool()
	clearDatabase(pool)

	repo := NewPgxRepository[debugPayload, debugResult](pool)

	pool.Exec(
		context.Background(),
		`
    INSERT INTO ana.tracked_operations (
       status,     key,            target,            payload, reference_time,                timeout,                       expiration,                    started_at
    ) VALUES
      ('finished', 'expired_key',  'finished_target', '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'expired_key',  'failed_target',   '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'expired_key',  'running_target',  '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('finished', 'finished_key', 'finished_target', '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'failed_key',   'failed_target',   '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'running_key',  'running_target',  '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('finished', 'expired_key',  'finished_again',  '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'expired_key',  'failed_again',    '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'expired_key',  'running_again',   '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('finished', 'finished_key', 'finished_again',  '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'failed_key',   'failed_again',    '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'running_key',  'running_again',   '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval)
    `,
		pgx.NamedArgs{},
	)
	assertEqual(t, int64(2), repo.FailExpiredStillRunning(10))
	assertEqual(t, int64(0), repo.FailExpiredStillRunning(10))

	pool.Exec(
		context.Background(),
		`
    INSERT INTO ana.tracked_operations (
       status,    key,           target,         payload, reference_time,                timeout,                       expiration,                    started_at
    ) VALUES
      ('running', 'expired_key', 'other_target', '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'running_key', 'other_target', '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'expired_key', 'other_again',  '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'running_key', 'other_again',  '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval)
    `,
		pgx.NamedArgs{},
	)
	assertEqual(t, int64(1), repo.FailExpiredStillRunning(1))
	assertEqual(t, int64(1), repo.FailExpiredStillRunning(1))
	assertEqual(t, int64(0), repo.FailExpiredStillRunning(1))

	pool.Exec(
		context.Background(),
		`
    INSERT INTO ana.tracked_operations (
       status,    key,           target,        payload, reference_time,                timeout,                       expiration,                    started_at
    ) VALUES
      ('running', 'expired_key', 'more_target', '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'running_key', 'more_target', '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'expired_key', 'more_again',  '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'running_key', 'more_again',  '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval)
    `,
		pgx.NamedArgs{},
	)
	assertEqual(t, int64(2), repo.FailExpiredStillRunning(2))
	assertEqual(t, int64(0), repo.FailExpiredStillRunning(2))
}

func TestRepositoryFailTimedOutStillRunning(t *testing.T) {
	pool := newPool()
	clearDatabase(pool)

	repo := NewPgxRepository[debugPayload, debugResult](pool)

	pool.Exec(
		context.Background(),
		`
    INSERT INTO ana.tracked_operations (
       status,     key,            target,            payload, reference_time,                timeout,                       expiration,                    started_at
    ) VALUES
      ('finished', 'expired_key',  'finished_target', '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'expired_key',  'failed_target',   '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'expired_key',  'running_target',  '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('finished', 'finished_key', 'finished_target', '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'failed_key',   'failed_target',   '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'running_key',  'running_target',  '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('finished', 'expired_key',  'finished_again',  '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'expired_key',  'failed_again',    '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'expired_key',  'running_again',   '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('finished', 'finished_key', 'finished_again',  '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('failed',   'failed_key',   'failed_again',    '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running',  'running_key',  'running_again',   '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval)
    `,
		pgx.NamedArgs{},
	)
	assertEqual(t, int64(2), repo.FailTimedOutStillRunning(10))
	assertEqual(t, int64(0), repo.FailTimedOutStillRunning(10))

	pool.Exec(
		context.Background(),
		`
    INSERT INTO ana.tracked_operations (
       status,    key,           target,         payload, reference_time,                timeout,                       expiration,                    started_at
    ) VALUES
      ('running', 'expired_key', 'other_target', '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'running_key', 'other_target', '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'expired_key', 'other_again',  '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'running_key', 'other_again',  '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval)
    `,
		pgx.NamedArgs{},
	)
	assertEqual(t, int64(1), repo.FailTimedOutStillRunning(1))
	assertEqual(t, int64(1), repo.FailTimedOutStillRunning(1))
	assertEqual(t, int64(0), repo.FailTimedOutStillRunning(1))

	pool.Exec(
		context.Background(),
		`
    INSERT INTO ana.tracked_operations (
       status,    key,           target,        payload, reference_time,                timeout,                       expiration,                    started_at
    ) VALUES
      ('running', 'expired_key', 'more_target', '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'running_key', 'more_target', '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'expired_key', 'more_again',  '',      NOW() - '3 minutes'::interval, NOW() - '2 minutes'::interval, NOW() - '1 minutes'::interval, NOW() - '2 minutes'::interval),
      ('running', 'running_key', 'more_again',  '',      NOW() - '3 minutes'::interval, NOW() + '2 minutes'::interval, NOW() + '1 minutes'::interval, NOW() - '2 minutes'::interval)
    `,
		pgx.NamedArgs{},
	)
	assertEqual(t, int64(2), repo.FailTimedOutStillRunning(2))
	assertEqual(t, int64(0), repo.FailTimedOutStillRunning(2))
}

func newPool() *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	return pool
}

func clearDatabase(pool *pgxpool.Pool) {
	if _, err := pool.Exec(context.Background(), "TRUNCATE ana.tracked_operations;"); err != nil {
		panic(err)
	}
}
