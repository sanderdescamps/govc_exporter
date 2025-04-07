package scraper

import (
	"context"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/types"
)

type TagsSensor struct {
	BaseSensor[types.ManagedObjectReference, []*tags.Tag]
	AutoRunSensor
	Refreshable
	TagCatsToObserve []string
	catCacheLock     sync.Mutex
	catCache         map[string]tags.Category
}

func NewTagsSensor(scraper *VCenterScraper, config TagsSensorConfig) *TagsSensor {
	return NewTagsSensorWithTaglist(scraper, config)
}

func NewTagsSensorWithTaglist(scraper *VCenterScraper, config TagsSensorConfig) *TagsSensor {
	var sensor TagsSensor
	sensor = TagsSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, []*tags.Tag](
			"tags",
			"TagsSensor",
			helper.NewMatcher("tags"),
			scraper,
		),
		TagCatsToObserve: config.CategoryToCollect,
		AutoRunSensor:    *NewAutoRunSensor(&sensor, config.SensorConfig),
	}
	sensor.metrics.ClientWaitTime = NewSensorMetricDuration(sensor.Kind(), "client_wait_time", 0)
	sensor.metrics.QueryTime = NewSensorMetricDuration(sensor.Kind(), "query_time", 0)
	sensor.metrics.Status = NewSensorMetricStatus(sensor.Kind(), "status", true)
	scraper.RegisterSensorMetric(
		&sensor.metrics.ClientWaitTime.SensorMetric,
		&sensor.metrics.QueryTime.SensorMetric,
		&sensor.metrics.Status.SensorMetric,
	)
	return &sensor
}

func (s *TagsSensor) Refresh(ctx context.Context) error {
	tagMap, err := s.tagRefresh(ctx)
	if err != nil {
		return err
	}

	for ref, tags := range tagMap {
		s.Update(ref, &tags)
	}

	return nil
}

func (s *TagsSensor) tagRefresh(ctx context.Context) (map[types.ManagedObjectReference][]*tags.Tag, error) {
	t1 := time.Now()

	restclient, release, err := s.scraper.clientPool.AcquireRest()
	defer release()
	if err != nil {
		return nil, err
	}
	t2 := time.Now()

	m := tags.NewManager(restclient)

	allCats, err := m.GetCategories(ctx)
	if err != nil {
		if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
			logger.Error("failed to get tag categories", "err", err)
		}
		return nil, err
	}

	catList := []tags.Category{}
	for _, cat := range allCats {
		if len(s.TagCatsToObserve) == 0 || slices.Contains(s.TagCatsToObserve, cat.Name) {
			catList = append(catList, cat)
		}
	}
	s.UpdateCatCache(catList)

	tagList := []tags.Tag{}
	s.metrics.Status.Reset()
	for _, cat := range catList {
		tags, err := m.GetTagsForCategory(ctx, cat.ID)
		if err != nil {
			if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
				logger.Error("failed to get tags for category", "category", cat, "err", err)
			}
			s.metrics.Status.Fail()
			return nil, err
		} else {
			s.metrics.Status.Success()
		}
		tagList = append(tagList, tags...)
	}

	objectTags := make(map[types.ManagedObjectReference][]*tags.Tag)
	for _, tag := range tagList {
		attachObjs, err := m.GetAttachedObjectsOnTags(ctx, []string{tag.ID})
		if err != nil {
			if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
				logger.Error("failed to get attached objects for tag", "tag", tag, "err", err)
			}
			return nil, err
		}

		for _, attachObj := range attachObjs {
			for _, elem := range attachObj.ObjectIDs {
				objectTags[elem.Reference()] = append(objectTags[elem.Reference()], attachObj.Tag)
			}
		}
	}
	t3 := time.Now()
	s.metrics.ClientWaitTime.Update(t2.Sub(t1))
	s.metrics.QueryTime.Update(t3.Sub(t2))

	return objectTags, nil
}

func (s *TagsSensor) UpdateCatCache(cats []tags.Category) {
	newCatCache := map[string]tags.Category{}
	for _, cat := range cats {
		newCatCache[cat.Name] = cat
	}
	s.catCacheLock.Lock()
	defer s.catCacheLock.Unlock()
	s.catCache = newCatCache
}

func (s *TagsSensor) GetCategoryID(cat string) string {
	s.catCacheLock.Lock()
	defer s.catCacheLock.Unlock()

	if s.catCache != nil {
		if tagVal, ok := s.catCache[cat]; ok {
			return tagVal.ID
		}
	}
	return ""
}

func (s *TagsSensor) GetTag(ref types.ManagedObjectReference, cat string) *tags.Tag {
	tags := s.Get(ref)
	catId := s.GetCategoryID(cat)
	if tags != nil && catId != "" {
		for _, tag := range *tags {
			if tag.CategoryID == catId {
				return tag
			}
		}
	}
	return nil
}
