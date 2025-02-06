package scraper

import (
	"log/slog"
	"sync"
	"time"
)

type HasMetrics interface {
	GetMetrics() *SensorMetric
}

type Sensor[K comparable, T any] interface {
	Get(K) T
	GetAll() []*T
	GetAllRefs() []K
}

type BaseSensor[K comparable, T any] struct {
	Sensor[K, T]
	HasMetrics
	cache      map[K]*CacheItem[T]
	scraper    *VCenterScraper
	lock       sync.Mutex
	metricLock sync.Mutex
	metrics    *SensorMetric
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

func (s *BaseSensor[K, T]) setMetrics(m *SensorMetric) {
	s.metricLock.Lock()
	defer s.metricLock.Unlock()
	s.metrics = m
}

func (s *BaseSensor[K, T]) GetMetrics() *SensorMetric {
	s.metricLock.Lock()
	defer s.metricLock.Unlock()
	return s.metrics
}

func (s *BaseSensor[K, T]) Update(ref K, item *T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache[ref] = NewCacheItem(item)
}
