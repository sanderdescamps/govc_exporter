package scraper

import (
	"context"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/types"
)

type TagsSensor struct {
	BaseSensor[types.ManagedObjectReference, []*tags.Tag]
	Refreshable
	Cleanable
	TagCatsToObserve []string
	catCacheLock     sync.Mutex
	catCache         map[string]tags.Category
}

func NewTagsSensor(scraper *VCenterScraper) *TagsSensor {
	return &TagsSensor{
		BaseSensor: BaseSensor[types.ManagedObjectReference, []*tags.Tag]{
			cache:   make(map[types.ManagedObjectReference]*CacheItem[[]*tags.Tag]),
			scraper: scraper,
			metrics: nil,
		},
	}
}

func NewTagsSensorWithTaglist(scraper *VCenterScraper, t []string) *TagsSensor {
	return &TagsSensor{
		BaseSensor: BaseSensor[types.ManagedObjectReference, []*tags.Tag]{
			cache:   make(map[types.ManagedObjectReference]*CacheItem[[]*tags.Tag]),
			scraper: scraper,
			metrics: nil,
		},
		TagCatsToObserve: t,
		// categoryCache: nil,
	}
}

func (s *TagsSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	t1 := time.Now()

	restclient, release, err := s.scraper.clientPool.AcquireRest()
	defer release()
	if err != nil {
		return err
	}
	t2 := time.Now()

	m := tags.NewManager(restclient)

	allCats, err := m.GetCategories(ctx)
	if err != nil {
		logger.Error("failed to get tag categories", "err", err)
		return err
	}

	catList := []tags.Category{}
	for _, cat := range allCats {
		if len(s.TagCatsToObserve) == 0 || slices.Contains(s.TagCatsToObserve, cat.Name) {
			catList = append(catList, cat)
		}
	}
	s.UpdateCatCache(catList)

	tagList := []tags.Tag{}
	for _, cat := range catList {
		tags, err := m.GetTagsForCategory(ctx, cat.ID)
		if err != nil {
			logger.Error("failed to get tags for category", "category", cat, "err", err)
			return err
		}
		tagList = append(tagList, tags...)
	}

	objectTags := make(map[types.ManagedObjectReference][]*tags.Tag)
	for _, tag := range tagList {
		attachObjs, err := m.GetAttachedObjectsOnTags(ctx, []string{tag.ID})
		if err != nil {
			logger.Error("failed to get attached objects for tag", "tag", tag, "err", err)
			return err
		}

		for _, attachObj := range attachObjs {
			for _, elem := range attachObj.ObjectIDs {
				objectTags[elem.Reference()] = append(objectTags[elem.Reference()], attachObj.Tag)
			}
		}
	}
	t3 := time.Now()
	s.setMetrics(&SensorMetric{
		Name:           "tags",
		QueryTime:      t3.Sub(t2),
		ClientWaitTime: t2.Sub(t1),
		Status:         true,
	})

	for ref, tags := range objectTags {
		s.Update(ref, &tags)
	}

	return nil
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

func (s *TagsSensor) Clean(maxAge time.Duration, logger *slog.Logger) {
	s.BaseSensor.Clean(maxAge, logger)
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
