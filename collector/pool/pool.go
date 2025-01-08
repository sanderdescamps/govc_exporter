package pool

type Pool[T any] interface {
	Acquire() (*T, int)
	Release(id int)
	Close() error
}
