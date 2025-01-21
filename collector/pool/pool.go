package pool

type Pool[T any] interface {
	Acquire() (*T, func(), error)
	Destroy() error
}
