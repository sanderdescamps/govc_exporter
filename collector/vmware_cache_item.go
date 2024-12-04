package collector

import (
	"time"

	"github.com/vmware/govmomi/vim25/mo"
)

type VMwareCacheItem struct {
	Entity    mo.ManagedEntity
	timeStamp time.Time
}

func (s *VMwareCacheItem) Expired(maxAge time.Duration) bool {
	return time.Now().After(s.timeStamp.Add(maxAge))
}
