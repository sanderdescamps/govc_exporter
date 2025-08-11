package scraper

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	sensormetrics "github.com/sanderdescamps/govc_exporter/internal/scraper/sensor_metrics"
	"github.com/vmware/govmomi/vim25/mo"
)

const CLUSTER_SENSOR_NAME = "ClusterSensor"

type ClusterSensor struct {
	BaseSensor
	logger.SensorLogger
	metricsCollector *sensormetrics.SensorMetricsCollector
	statusMonitor    *sensormetrics.StatusMonitor
	started          *helper.StartedCheck
	sensorLock       sync.Mutex
	manualRefresh    chan struct{}
	stopChan         chan struct{}
	config           config.SensorConfig
}

func NewClusterSensor(scraper *VCenterScraper, config config.SensorConfig, l *slog.Logger) *ClusterSensor {
	var mc *sensormetrics.SensorMetricsCollector = sensormetrics.NewLastSensorMetricsCollector()
	var sm *sensormetrics.StatusMonitor = sensormetrics.NewStatusMonitor()

	var sensor ClusterSensor = ClusterSensor{
		BaseSensor: *NewBaseSensor(
			"ClusterComputeResource", []string{
				"parent",
				"name",
				"summary",
			},
			mc, sm),
		started:          helper.NewStartedCheck(),
		stopChan:         make(chan struct{}),
		manualRefresh:    make(chan struct{}),
		config:           config,
		SensorLogger:     logger.NewSLogLogger(l, logger.WithKind(CLUSTER_SENSOR_NAME)),
		metricsCollector: mc,
		statusMonitor:    sm,
	}

	return &sensor
}

func (s *ClusterSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	var clusters []mo.ClusterComputeResource
	err := s.baseRefresh(ctx, scraper, &clusters)
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		oCluster := ConvertToCluster(ctx, scraper, cluster, time.Now())
		err := scraper.DB.SetCluster(ctx, oCluster, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ClusterSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *ClusterSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *ClusterSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *ClusterSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *ClusterSensor) Kind() string {
	return "ClusterSensor"
}

func (s *ClusterSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *ClusterSensor) Match(name string) bool {
	return helper.NewMatcher("cluster").Match(name)
}

func (s *ClusterSensor) Enabled() bool {
	return true
}

func (s *ClusterSensor) GetLatestMetrics() []sensormetrics.SensorMetric {
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

func ConvertToCluster(ctx context.Context, scraper *VCenterScraper, c mo.ClusterComputeResource, t time.Time) objects.Cluster {
	self := objects.NewManagedObjectReferenceFromVMwareRef(c.Self)

	var parent *objects.ManagedObjectReference
	if c.Parent != nil {
		p := objects.NewManagedObjectReferenceFromVMwareRef(*c.Parent)
		parent = &p
	}

	cluster := objects.Cluster{
		Timestamp: t,
		Name:      c.Name,
		Self:      self,
		Parent:    parent,
	}

	if cluster.Parent != nil {
		parentChain := scraper.DB.GetParentChain(ctx, *cluster.Parent)
		cluster.Datacenter = parentChain.DC
	}

	if summary := c.Summary.GetComputeResourceSummary(); summary != nil {
		cluster.TotalCPU = float64(summary.TotalCpu)
		cluster.EffectiveCPU = float64(summary.EffectiveCpu)
		cluster.TotalMemory = float64(summary.TotalMemory)
		cluster.EffectiveMemory = float64(summary.EffectiveMemory)
		cluster.NumCPUCores = float64(summary.NumCpuCores)
		cluster.NumCPUThreads = float64(summary.NumCpuThreads)
		cluster.NumEffectiveHosts = float64(summary.NumEffectiveHosts)
		cluster.NumHosts = float64(summary.NumHosts)
		cluster.OverallStatus = string(summary.OverallStatus)
	}

	return cluster
}
