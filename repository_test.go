package idempotency_manager

type emptyRepository struct {
}

func newEmptyRepository() *emptyRepository {
	return &emptyRepository{}
}

func (repo *emptyRepository) FetchOrStart(Operation[mockedPayload, mockedResult, *mockedCtx]) *TrackedOperation[mockedPayload, mockedResult] {
	return nil
}

func (repo *emptyRepository) NewSession(operation Operation[mockedPayload, mockedResult, *mockedCtx]) *Session[mockedPayload, mockedResult, *mockedCtx] {
	return NewSession(operation, newMockedCtx())
}

type trackedOperationRepository struct {
	trackedOperation *TrackedOperation[mockedPayload, mockedResult]
}

func newTrackedOperationRepository(trackedOperation *TrackedOperation[mockedPayload, mockedResult]) *trackedOperationRepository {
	return &trackedOperationRepository{
		trackedOperation: trackedOperation,
	}
}

func (repo *trackedOperationRepository) FetchOrStart(operation Operation[mockedPayload, mockedResult, *mockedCtx]) *TrackedOperation[mockedPayload, mockedResult] {
	return repo.trackedOperation
}

func (repo *trackedOperationRepository) NewSession(operation Operation[mockedPayload, mockedResult, *mockedCtx]) *Session[mockedPayload, mockedResult, *mockedCtx] {
	return NewSession(operation, newMockedCtx())
}
