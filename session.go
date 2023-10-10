package idempotency_manager

type Session[R any, C any] interface {
	Close()
	Ctx() *C
	SetResult(*R)
	Result() *R
	SetErr(error)
	Err() error
}
