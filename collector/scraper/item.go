package scraper

import (
	"time"
)

type VMwareCacheItem[t VMwareResource] struct {
	Item       *t
	expireTime time.Time
}

func NewVMwareCacheItem[T VMwareResource](item *T, maxAge time.Duration) *VMwareCacheItem[T] {
	return &VMwareCacheItem[T]{
		Item:       item,
		expireTime: time.Now().Add(maxAge),
	}
}

func (s *VMwareCacheItem[T]) Expired() bool {
	return time.Now().After(s.expireTime)
}
