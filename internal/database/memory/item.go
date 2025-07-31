package memory_db

import "time"

type Item struct {
	expire time.Time
	value  interface{}
}

func NewItem(value interface{}) *Item {
	return &Item{
		value:  value,
		expire: time.Time{},
	}
}

func NewItemWithExpire(value interface{}, expire time.Time) *Item {
	return &Item{
		value:  value,
		expire: expire,
	}
}

func NewItemWithTTL(value interface{}, ttl time.Duration) *Item {
	return &Item{
		value:  value,
		expire: time.Now().Add(ttl),
	}
}

func (i *Item) Expired() bool {
	if !i.expire.IsZero() {
		return time.Now().After(i.expire)
	}
	return false
}
