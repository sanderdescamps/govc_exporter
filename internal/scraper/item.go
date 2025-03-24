package scraper

import (
	"time"
)

type CacheItem[T any] struct {
	Item     *T
	creation time.Time
}

type Snapshot[T any] struct {
	Timestamp time.Time
	Item      T
}

func NewCacheItem[T any](item *T) *CacheItem[T] {
	return &CacheItem[T]{
		Item:     item,
		creation: time.Now(),
	}
}

func (s CacheItem[T]) Expired(maxAge time.Duration) bool {
	expireTime := s.creation.Add(maxAge)
	return time.Now().After(expireTime)
}

func (s CacheItem[T]) ToSnapshot() Snapshot[T] {
	return Snapshot[T]{
		Timestamp: s.creation,
		Item:      *s.Item,
	}
}
