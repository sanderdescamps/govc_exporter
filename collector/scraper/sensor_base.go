package scraper

import (
	"encoding/json"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/collector/helper"
	"github.com/vmware/govmomi/vim25/types"
)

type Sensor interface {
	helper.Matchable
	Refreshable
	GetAllJsons() (map[string][]byte, error)
}

type BaseSensor[K comparable, T any] struct {
	cache       map[K]*CacheItem[T]
	scraper     *VCenterScraper
	lock        sync.Mutex
	refreshLock sync.Mutex
	metrics     struct {
		QueryTime      *SensorMetricDuration
		ClientWaitTime *SensorMetricDuration
		Status         *SensorMetricStatus
	}
}

func NewBaseSensor[K comparable, T any](scraper *VCenterScraper) *BaseSensor[K, T] {
	return &BaseSensor[K, T]{
		cache:   make(map[K]*CacheItem[T]),
		scraper: scraper,
	}
}

func (s *BaseSensor[K, T]) Clean(maxAge time.Duration, logger *slog.Logger) {
	s.lock.Lock()
	defer s.lock.Unlock()
	cleanCache := map[K]*CacheItem[T]{}
	for k, v := range s.cache {
		if !v.Expired(maxAge) {
			cleanCache[k] = v
		} else {
			logger.Debug("Cleanup metric from sensor", "ref", k)
		}
	}
	s.cache = cleanCache
}

func (s *BaseSensor[K, T]) Get(ref K) *T {
	s.lock.Lock()
	defer s.lock.Unlock()
	if o, ok := s.cache[ref]; ok {
		return o.Item
	}
	return nil
}

func (s *BaseSensor[K, T]) GetAll() []*T {
	result := []*T{}
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, v := range s.cache {
		result = append(result, v.Item)
	}
	return result
}

type TimeData[T any] struct {
	Time time.Time
	Data T
}

func (s *BaseSensor[K, T]) GetAllSnapshots() []Snapshot[T] {
	result := []Snapshot[T]{}
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, v := range s.cache {
		result = append(result, v.ToSnapshot())
	}
	return result
}

func (s *BaseSensor[K, T]) GetAllRefs() []K {
	result := []K{}
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.cache) == 0 {
		return nil
	}

	for k := range s.cache {
		result = append(result, k)
	}
	return result
}

func (s *BaseSensor[K, T]) Update(ref K, item *T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache[ref] = NewCacheItem(item)
}

func (s *BaseSensor[K, T]) UpdateCacheItem(ref K, item *CacheItem[T]) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache[ref] = item
}

func (s *BaseSensor[K, T]) GetKind() string {
	return reflect.ValueOf(s).String()
}

func (s *BaseSensor[K, T]) GetAllJsons() (map[string][]byte, error) {
	result := map[string][]byte{}
	s.lock.Lock()
	defer s.lock.Unlock()
	for i, cacheItem := range s.cache {
		name := ""
		key := reflect.ValueOf(i).Interface()
		if ref, ok := key.(types.ManagedObjectReference); ok {
			name = ref.Value
		} else {
			name = helper.RandomString(8)
		}

		jsonBytes, err := json.MarshalIndent(map[string]any{
			"time":   cacheItem.creation,
			"object": cacheItem.Item,
		}, "", "  ")
		if err != nil {
			return nil, err
		}
		result[name] = jsonBytes
	}
	return result, nil
}
