package idempotency_manager

type emptyRepository struct {
}

func newEmptyRepository() *emptyRepository {
	return &emptyRepository{}
}

func (repo *emptyRepository) FetchOrStart(Operation[string, mockedResult, mockedCtx]) *TrackedOperation[string, mockedResult] {
	return nil
}

func (repo *emptyRepository) Call(Operation[string, mockedResult, mockedCtx]) (*mockedResult, error) {
	return nil, nil
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

func (repo *trackedOperationRepository) Call(operation Operation[string, mockedResult, mockedCtx]) (*mockedResult, error) {
	return operation.Call(nil)
}
