package idempotency_manager

import (
	"time"

	"testing"
)

func TestExpiredOperation(t *testing.T) {
	manager := New(newEmptyRepository())
	operation := newMockedOperation(
		"key",
		"target",
		"payload",
		time.Now().Add(-20*time.Second),
		5*time.Second,
		10*time.Second,
		newMockedResult("result"),
	)
	result, err := manager.Call(operation)

	exptectedErr := newExpirationError("target", "key")
	if err == nil || err.Error() != exptectedErr.Error() {
		t.Fatalf("Expected to have \"%v\" error, but got \"%v\"", exptectedErr, err)
	}

	if result != nil {
		t.Fatalf("Expected to have no result, but got \"%s\"", result.result)
	}
}

func TestAlreadyFinishedOperation(t *testing.T) {
	trackedOperation := NewTrackedOperation(
		Finished,
		"key",
		"target",
		"payload",
		time.Now().Add(-10*time.Second),
		time.Now().Add(-5*time.Second),
		time.Now().Add(5*time.Second),
		time.Now().Add(10*time.Second),
		newMockedResult("tracked result"),
	)
	manager := New(newTrackedOperationRepository(trackedOperation))
	operation := newMockedOperation(
		"key",
		"target",
		"payload",
		time.Now(),
		5*time.Second,
		10*time.Second,
		newMockedResult("result"),
	)
	result, err := manager.Call(operation)

	if err != nil {
		t.Fatalf("Expected to have no error, but got \"%v\"", err)
	}

	if result == nil || result.result != "tracked result" {
		t.Fatalf("Expected to have \"tracked result\" as result, but got \"%s\"", result.result)
	}
}

func TestAlreadyExpiredAndFinishedOperation(t *testing.T) {
	trackedOperation := NewTrackedOperation(
		Finished,
		"key",
		"target",
		"payload",
		time.Now().Add(-10*time.Second),
		time.Now().Add(-7*time.Second),
		time.Now().Add(-10*time.Second),
		time.Now().Add(-5*time.Second),
		newMockedResult("tracked result"),
	)
	manager := New(newTrackedOperationRepository(trackedOperation))
	operation := newMockedOperation(
		"key",
		"target",
		"payload",
		time.Now(),
		5*time.Second,
		10*time.Second,
		newMockedResult("result"),
	)
	result, err := manager.Call(operation)

	if err != nil {
		t.Fatalf("Expected to have no error, but got \"%v\"", err)
	}

	if result == nil || result.result != "tracked result" {
		t.Fatalf("Expected to have \"tracked result\" as result, but got \"%s\"", result.result)
	}
}

func TestAlreadyExpiredAndReadyOperation(t *testing.T) {
	trackedOperation := NewTrackedOperation(
		Ready,
		"key",
		"target",
		"payload",
		time.Now().Add(-10*time.Second),
		time.Now().Add(-7*time.Second),
		time.Now().Add(-10*time.Second),
		time.Now().Add(-5*time.Second),
		newMockedResult("tracked result"),
	)
	manager := New(newTrackedOperationRepository(trackedOperation))
	operation := newMockedOperation(
		"key",
		"target",
		"payload",
		time.Now(),
		5*time.Second,
		10*time.Second,
		newMockedResult("result"),
	)
	result, err := manager.Call(operation)

	exptectedErr := newExpirationError("target", "key")
	if err == nil || err.Error() != exptectedErr.Error() {
		t.Fatalf("Expected to have \"%v\" error, but got \"%v\"", exptectedErr, err)
	}

	if result != nil {
		t.Fatalf("Expected to have no result, but got \"%s\"", result.result)
	}
}

func TestAlreadyExpiredAndRunningOperation(t *testing.T) {
	trackedOperation := NewTrackedOperation(
		Running,
		"key",
		"target",
		"payload",
		time.Now().Add(-10*time.Second),
		time.Now().Add(-7*time.Second),
		time.Now().Add(-10*time.Second),
		time.Now().Add(-5*time.Second),
		newMockedResult("tracked result"),
	)
	manager := New(newTrackedOperationRepository(trackedOperation))
	operation := newMockedOperation(
		"key",
		"target",
		"payload",
		time.Now(),
		5*time.Second,
		10*time.Second,
		newMockedResult("result"),
	)
	result, err := manager.Call(operation)

	exptectedErr := newExpirationError("target", "key")
	if err == nil || err.Error() != exptectedErr.Error() {
		t.Fatalf("Expected to have \"%v\" error, but got \"%v\"", exptectedErr, err)
	}

	if result != nil {
		t.Fatalf("Expected to have no result, but got \"%s\"", result.result)
	}
}

func TestAlreadyExpiredAndFailedOperation(t *testing.T) {
	trackedOperation := NewTrackedOperation(
		Failed,
		"key",
		"target",
		"payload",
		time.Now().Add(-10*time.Second),
		time.Now().Add(-7*time.Second),
		time.Now().Add(-10*time.Second),
		time.Now().Add(-5*time.Second),
		newMockedResult("tracked result"),
	)
	manager := New(newTrackedOperationRepository(trackedOperation))
	operation := newMockedOperation(
		"key",
		"target",
		"payload",
		time.Now(),
		5*time.Second,
		10*time.Second,
		newMockedResult("result"),
	)
	result, err := manager.Call(operation)

	exptectedErr := newExpirationError("target", "key")
	if err == nil || err.Error() != exptectedErr.Error() {
		t.Fatalf("Expected to have \"%v\" error, but got \"%v\"", exptectedErr, err)
	}

	if result != nil {
		t.Fatalf("Expected to have no result, but got \"%s\"", result.result)
	}
}

func TestAlreadyStillRunningOperation(t *testing.T) {
	trackedOperation := NewTrackedOperation(
		Running,
		"key",
		"target",
		"payload",
		time.Now().Add(-10*time.Second),
		time.Now().Add(-7*time.Second),
		time.Now().Add(5*time.Second),
		time.Now().Add(7*time.Second),
		newMockedResult("tracked result"),
	)
	manager := New(newTrackedOperationRepository(trackedOperation))
	operation := newMockedOperation(
		"key",
		"target",
		"payload",
		time.Now(),
		5*time.Second,
		10*time.Second,
		newMockedResult("result"),
	)
	result, err := manager.Call(operation)

	exptectedErr := newStillRunningError("target", "key")
	if err == nil || err.Error() != exptectedErr.Error() {
		t.Fatalf("Expected to have \"%v\" error, but got \"%v\"", exptectedErr, err)
	}

	if result != nil {
		t.Fatalf("Expected to have no result, but got \"%s\"", result.result)
	}
}

func TestNeverTimeoutRunningOperation(t *testing.T) {
	trackedOperation := NewTrackedOperation(
		Running,
		"key",
		"target",
		"payload",
		time.Now().Add(-10*time.Second),
		time.Now().Add(-7*time.Second),
		time.Time{},
		time.Now().Add(5*time.Second),
		newMockedResult("tracked result"),
	)
	manager := New(newTrackedOperationRepository(trackedOperation))
	operation := newMockedOperation(
		"key",
		"target",
		"payload",
		time.Now(),
		time.Duration(0),
		5*time.Second,
		newMockedResult("result"),
	)
	result, err := manager.Call(operation)

	exptectedErr := newStillRunningError("target", "key")
	if err == nil || err.Error() != exptectedErr.Error() {
		t.Fatalf("Exptected to have \"%v\" error, but got \"%v\"", exptectedErr, err)
	}

	if result != nil {
		t.Fatalf("Exptected to have no result, but got \"%s\"", result.result)
	}
}
