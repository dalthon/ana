package pgx

import (
	"context"
	"errors"
	"os"
	"time"

	a "github.com/dalthon/ana"

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
