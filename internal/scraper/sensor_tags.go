package scraper

import (
	"context"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	metricshelper "github.com/sanderdescamps/govc_exporter/internal/scraper/metrics_helper"
	"github.com/vmware/govmomi/vapi/tags"
)

const TAGS_SENSOR_NAME = "TagsSensor"

type TagsSensor struct {
	logger.SensorLogger
	metricshelper.MetricHelperDefault
	started       *helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        TagsSensorConfig
}

func NewTagsSensor(scraper *VCenterScraper, config TagsSensorConfig, l *slog.Logger) *TagsSensor {
	return &TagsSensor{
		started:             helper.NewStartedCheck(),
		stopChan:            make(chan struct{}),
		manualRefresh:       make(chan struct{}),
		config:              config,
		SensorLogger:        logger.NewSLogLogger(l, logger.WithKind(TAGS_SENSOR_NAME)),
		MetricHelperDefault: *metricshelper.NewMetricHelperDefault(TAGS_SENSOR_NAME),
	}
}

func (s *TagsSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	s.MetricHelperDefault.Start()

	restclient, release, err := scraper.clientPool.AcquireRest()
	defer release()
	if err != nil {
		s.MetricHelperDefault.Fail()
		return ErrSensorCientFailed
	}
	s.MetricHelperDefault.Mark1()

	m := tags.NewManager(restclient)

	allCats, err := m.GetCategories(ctx)
	if err != nil {
		return NewSensorError("failed to get tag categories", "err", err)
	}

	catList := []tags.Category{}
	for _, cat := range allCats {
		if len(s.config.CategoryToCollect) == 0 || slices.Contains(s.config.CategoryToCollect, cat.Name) {
			catList = append(catList, cat)
		}
	}

	objectTags := map[string]objects.TagSet{}
	for _, cat := range catList {
		tags, err := m.GetTagsForCategory(ctx, cat.ID)
		if err != nil {
			s.MetricHelperDefault.Fail()
			return NewSensorError("failed to get tags for category", "category", cat, "err", err)
		}

		for _, tag := range tags {
			attachObjs, err := m.GetAttachedObjectsOnTags(ctx, []string{tag.ID})
			if err != nil {
				s.MetricHelperDefault.Fail()
				return NewSensorError("failed to get attached objects for tag", "tag", tag, "err", err)
			}

			for _, attachObj := range attachObjs {
				for _, elem := range attachObj.ObjectIDs {
					ref := objects.NewManagedObjectReferenceFromVMwareRef(elem.Reference())
					if _, ok := objectTags[ref.Hash()]; !ok {
						objectTags[ref.Hash()] = objects.TagSet{
							ObjectRef: ref,
							Tags:      map[string]string{cat.Name: tag.Name},
						}
					} else {
						objectTags[ref.Hash()].Tags[cat.Name] = tag.Name
					}
				}
			}
		}
	}

	s.MetricHelperDefault.Finish(true)

	for _, tag := range objectTags {
		err := scraper.DB.SetTags(ctx, tag, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *TagsSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
	if !s.started.IsStarted() {
		err := s.refresh(ctx, scraper)
		if err != nil {
			return err
		}
		s.started.Started()
	} else {
		return ErrSensorAlreadyStarted
	}
	return nil
}

func (s *TagsSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
	ticker := time.NewTicker(s.config.RefreshInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				go func() {
					err := s.refresh(ctx, scraper)
					if err == nil {
						s.SensorLogger.Info("refresh successful")
					} else {
						s.SensorLogger.Error("refresh failed", "err", err)
					}
				}()
			case <-s.manualRefresh:
				go func() {
					s.SensorLogger.Info("trigger manual refresh")
					err := s.refresh(ctx, scraper)
					if err == nil {
						s.SensorLogger.Info("manual refresh successful")
					} else {
						s.SensorLogger.Error("manual refresh failed", "err", err)
					}
				}()
			case <-s.stopChan:
				s.started.Stopped()
				ticker.Stop()
			case <-ctx.Done():
				s.started.Stopped()
				ticker.Stop()
			}
		}
	}()
	return nil
}

func (s *TagsSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *TagsSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *TagsSensor) Kind() string {
	return "TagsSensor"
}

func (s *TagsSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *TagsSensor) Match(name string) bool {
	return helper.NewMatcher("tags", "tag").Match(name)
}

func (s *TagsSensor) Enabled() bool {
	return true
}
