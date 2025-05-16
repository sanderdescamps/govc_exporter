package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type Cleanable interface {
	Clean(ctx context.Context, maxAge time.Duration)
}

type Refreshable interface {
	Refresh(context.Context) error
}

type Sensor interface {
	// helper.Matchable
	Refreshable
	Cleanable
	Runnable
	Name() string
	Match(string) bool
	Kind() string
	GetAllJsons() (map[string][]byte, error)
	TriggerInstantRefresh(context.Context) error
}

type CleanOnlySensor interface {
	// helper.Matchable
	Cleanable
	Name() string
	Match(string) bool
	Kind() string
	GetAllJsons() (map[string][]byte, error)
}

type BaseSensor[K comparable, T any | mo.ManagedEntity] struct {
	sensorName    string
	sensorKind    string
	sensorMatcher helper.Matchable
	cache         map[K]*CacheItem[T]
	scraper       *VCenterScraper

	lock       sync.RWMutex
	sensorLock sync.Mutex
	metrics    struct {
		QueryTime      *SensorMetricDuration
		ClientWaitTime *SensorMetricDuration
		Status         *SensorMetricStatus
	}
}

func NewBaseSensor[K comparable, T any | mo.ManagedEntity](name string, kind string, matcher helper.Matchable, scraper *VCenterScraper) *BaseSensor[K, T] {
	return &BaseSensor[K, T]{
		cache:         make(map[K]*CacheItem[T]),
		sensorName:    name,
		sensorKind:    kind,
		sensorMatcher: matcher,
		scraper:       scraper,
	}
}

func (s *BaseSensor[K, T]) Clean(ctx context.Context, maxAge time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.cache) < 1 {
		return
	}
	cleanupCount := 0
	for k, v := range s.cache {
		if v.Expired(maxAge) {
			delete(s.cache, k)
			cleanupCount += 1
		}
	}

	if cleanupCount > 0 {
		if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
			msg := fmt.Sprintf("Clean %d objects from sensor cache", cleanupCount)
			logger.Debug(msg, "kind", s.Kind())
		}
	}

}

func (s *BaseSensor[K, T]) Get(ref K) *T {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if o, ok := s.cache[ref]; ok {
		return o.Item
	}
	return nil
}

func (s *BaseSensor[K, T]) GetAll() []*T {
	result := []*T{}
	s.lock.RLock()
	defer s.lock.RUnlock()
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
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, v := range s.cache {
		result = append(result, v.ToSnapshot())
	}
	return result
}

func (s *BaseSensor[K, T]) GetAllRefs() []K {
	result := []K{}
	s.lock.RLock()
	defer s.lock.RUnlock()
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
	if item == nil {
		return
	}
	s.cache[ref] = NewCacheItem(item)
}

func (s *BaseSensor[K, T]) GetAllJsons() (map[string][]byte, error) {
	result := map[string][]byte{}
	s.lock.RLock()
	defer s.lock.RUnlock()
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

func (s *BaseSensor[K, T]) Kind() string {
	return s.sensorKind
}

func (s *BaseSensor[K, T]) Name() string {
	return s.sensorName
}

func (s *BaseSensor[K, T]) Match(name string) bool {
	return s.sensorMatcher.Match(name)
}

var ErrSensorAlreadyRunning = fmt.Errorf("Sensor already running")

func (s *BaseSensor[K, T]) baseRefresh(ctx context.Context, moType string, moProperties []string) ([]T, error) {
	if ok := s.sensorLock.TryLock(); !ok {
		if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
			logger.Info("Sensor Refresh already running", "sensor_type", s.Kind())
		}
		return nil, ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	t1 := time.Now()
	client, release, err := s.scraper.clientPool.Acquire()
	if err != nil {
		return nil, err
	}
	defer release()
	t2 := time.Now()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{moType},
		true,
	)
	if err != nil {
		return nil, err
	}
	defer v.Destroy(ctx)

	var entities []T
	err = v.Retrieve(
		context.Background(),
		[]string{moType},
		moProperties,
		&entities,
	)
	t3 := time.Now()
	s.metrics.ClientWaitTime.Update(t2.Sub(t1))
	s.metrics.QueryTime.Update(t3.Sub(t2))
	if err == nil {
		s.metrics.Status.Update(true)
	} else {
		s.metrics.Status.Update(false)
		return nil, err
	}

	return entities, nil
}
