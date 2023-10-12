package pgx

import (
	"time"

	"testing"
)

func assertEqual(t *testing.T, expected, value any) {
	if expected != value {
		t.Fatalf("Expected \"%v\" to be equal to \"%v\", but wasn't.", expected, value)
	}
}

func assertTimeEqual(t *testing.T, expected, value time.Time) {
	if !expected.Equal(value) {
		t.Fatalf("Expected \"%v\" to be equal to \"%v\", but wasn't.", expected, value)
	}
}

func assertNil[R any](t *testing.T, expected *R) {
	if expected != nil {
		t.Fatalf("Expected \"%v\" to be nil, but wasn't.", expected)
	}
}

func assertErrorNil(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Expected \"%v\" to be nil, but wasn't.", err)
	}
}
