package scraper

import (
	"context"
	"log/slog"
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	sensormetrics "github.com/sanderdescamps/govc_exporter/internal/scraper/sensor_metrics"
	"github.com/vmware/govmomi/vapi/tags"
)

const TAGS_SENSOR_NAME = "TagsSensor"

type TagsSensor struct {
	logger.SensorLogger
	metricsCollector *sensormetrics.SensorMetricsCollector
	statusMonitor    *sensormetrics.StatusMonitor
	started          *helper.StartedCheck
	sensorLock       sync.Mutex
	manualRefresh    chan struct{}
	stopChan         chan struct{}
	config           config.TagsSensorConfig
}

func NewTagsSensor(scraper *VCenterScraper, config config.TagsSensorConfig, l *slog.Logger) *TagsSensor {
	var mc *sensormetrics.SensorMetricsCollector = sensormetrics.NewLastSensorMetricsCollector()
	var sm *sensormetrics.StatusMonitor = sensormetrics.NewStatusMonitor()
	return &TagsSensor{
		started:          helper.NewStartedCheck(),
		stopChan:         make(chan struct{}),
		manualRefresh:    make(chan struct{}),
		config:           config,
		SensorLogger:     logger.NewSLogLogger(l, logger.WithKind(TAGS_SENSOR_NAME)),
		metricsCollector: mc,
		statusMonitor:    sm,
	}
}

func (s *TagsSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	sensorStopwatch := sensormetrics.NewSensorStopwatch()
	sensorStopwatch.Start()

	restclient, release, err := scraper.clientPool.AcquireRest()
	defer release()
	if err != nil {
		return ErrSensorCientFailed
	}
	sensorStopwatch.Mark1()

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
			return NewSensorError("failed to get tags for category", "category", cat, "err", err)
		}

		for _, tag := range tags {
			attachObjs, err := m.GetAttachedObjectsOnTags(ctx, []string{tag.ID})
			if err != nil {
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

	sensorStopwatch.Finish()
	s.metricsCollector.UploadStats(sensorStopwatch.GetStats())

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
			s.statusMonitor.Fail()
			return err
		}
		s.statusMonitor.Success()
		s.started.Started()
	} else {
		return ErrSensorAlreadyStarted
	}
	return nil
}

func (s *TagsSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
	ticker := time.NewTicker(s.config.RefreshInterval)
	go func() {
		time.Sleep(time.Duration(rand.Intn(20000)) * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				go func() {
					err := s.refresh(ctx, scraper)
					if err == nil {
						s.SensorLogger.Debug("refresh successful")
						s.statusMonitor.Success()
					} else {
						s.SensorLogger.Error("refresh failed", "err", err)
						s.statusMonitor.Fail()
					}
				}()
			case <-s.manualRefresh:
				go func() {
					s.SensorLogger.Info("trigger manual refresh")
					err := s.refresh(ctx, scraper)
					if err == nil {
						s.SensorLogger.Info("manual refresh successful")
						s.statusMonitor.Success()
					} else {
						s.SensorLogger.Error("manual refresh failed", "err", err)
						s.statusMonitor.Fail()
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

func (s *TagsSensor) GetLatestMetrics() []sensormetrics.SensorMetric {
	return append(
		s.metricsCollector.ComposeMetrics(s.Kind()),
		sensormetrics.SensorMetric{
			Sensor:     s.Kind(),
			MetricName: "failed",
			Value:      s.statusMonitor.StatusFailedFloat64(),
			Unit:       "boolean",
		}, sensormetrics.SensorMetric{
			Sensor:     s.Kind(),
			MetricName: "fail_rate",
			Value:      s.statusMonitor.FailRate(),
			Unit:       "boolean",
		}, sensormetrics.SensorMetric{
			Sensor:     s.Kind(),
			MetricName: "enabled",
			Value:      1.0,
			Unit:       "boolean",
		},
	)
}
