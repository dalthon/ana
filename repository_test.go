package idempotency_manager

type emptyRepository struct {
}

func newEmptyRepository() *emptyRepository {
	return &emptyRepository{}
}

func (repo *emptyRepository) FetchOrStart(Operation[string, mockedResult, mockedCtx]) *TrackedOperation[string, mockedResult] {
	return nil
}

func (repo *emptyRepository) NewSession(operation Operation[string, mockedResult, mockedCtx]) Session[mockedResult, mockedCtx] {
	return newMockedSession(operation)
}

type trackedOperationRepository struct {
	trackedOperation *TrackedOperation[string, mockedResult]
}

func newTrackedOperationRepository(trackedOperation *TrackedOperation[string, mockedResult]) *trackedOperationRepository {
	return &trackedOperationRepository{
		trackedOperation: trackedOperation,
	}
}

func (repo *trackedOperationRepository) FetchOrStart(operation Operation[string, mockedResult, mockedCtx]) *TrackedOperation[string, mockedResult] {
	return repo.trackedOperation
}

func (repo *trackedOperationRepository) NewSession(operation Operation[string, mockedResult, mockedCtx]) Session[mockedResult, mockedCtx] {
	return newMockedSession(operation)
}
