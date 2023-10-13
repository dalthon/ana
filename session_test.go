package ana

import (
	"time"

	"testing"
)

func TestSuccessfullSession(t *testing.T) {
	operation := newMockedOperation(
		"key",
		"target",
		newMockedPayload("payload"),
		time.Now().Add(-20*time.Second),
		5*time.Second,
		10*time.Second,
		newMockedResultFn("result"),
	)

	ctx := newMockedCtx()
	session := NewSession(operation, ctx)
	session.call()
	session.close()
	session.close()

	if ctx.SuccessCount != 1 {
		t.Fatalf("Expected to have called ctx.Success once, but called %d times", ctx.SuccessCount)
	}

	if ctx.FailCount != 0 {
		t.Fatalf("Expected to not have called ctx.Fail once, but called %d times", ctx.FailCount)
	}
}
