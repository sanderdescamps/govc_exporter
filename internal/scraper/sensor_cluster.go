package scraper

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	metricshelper "github.com/sanderdescamps/govc_exporter/internal/scraper/metrics_helper"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const CLUSTER_SENSOR_NAME = "ClusterSensor"

type ClusterSensor struct {
	BaseSensor
	logger.SensorLogger
	metricshelper.MetricHelperDefault
	started       helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        SensorConfig
}

func NewClusterSensor(scraper *VCenterScraper, config SensorConfig, l *slog.Logger) *ClusterSensor {
	var sensor ClusterSensor
	sensor = ClusterSensor{
		BaseSensor: *NewBaseSensor(
			"ClusterComputeResource", []string{
				"parent",
				"name",
				"summary",
			}),
		stopChan:            make(chan struct{}),
		config:              config,
		SensorLogger:        logger.NewSLogLogger(l, logger.WithKind(CLUSTER_SENSOR_NAME)),
		MetricHelperDefault: *metricshelper.NewMetricHelperDefault(CLUSTER_SENSOR_NAME),
	}

	return &sensor
}

func (s *ClusterSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	var clusters []mo.ClusterComputeResource
	stats, err := s.baseRefresh(ctx, scraper, &clusters)
	s.MetricHelperDefault.LoadStats(stats)
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		oCluster := ConvertToCluster(ctx, scraper, cluster, time.Now())
		err := scraper.DB.SetCluster(ctx, &oCluster, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ClusterSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) {
	ticker := time.NewTicker(s.config.RefreshInterval)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				s.refresh(ctx, scraper)
				err := s.refresh(ctx, scraper)
				if err == nil {
					s.SensorLogger.Info("refresh successful")
				} else {
					s.SensorLogger.Error("refresh failed", "err", err)
				}
				s.started.Started()
			case <-s.manualRefresh:
				s.SensorLogger.Info("trigger manual refresh")
				err := s.refresh(ctx, scraper)
				if err == nil {
					s.SensorLogger.Info("manual refresh successful")
				} else {
					s.SensorLogger.Error("manual refresh failed", "err", err)
				}
			case <-s.stopChan:
				s.started.Stopped()
				return
			}
		}
	}()
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
	return helper.NewMatcher("keyword").Match(name)
}

func (s *ClusterSensor) Enabled() bool {
	return true
}

func ConvertToCluster(ctx context.Context, scraper *VCenterScraper, c mo.ClusterComputeResource, t time.Time) objects.Cluster {
	self := objects.NewManagedObjectReference(c.Self.Type, c.Self.Value)

	var parent *objects.ManagedObjectReference
	if c.Parent != nil {
		p := objects.NewManagedObjectReference(parent.Type, parent.Value)
		parent = &p
	}

	cluster := objects.Cluster{
		Timestamp: t,
		Name:      c.Name,
		Self:      self,
		Parent:    parent,
	}

	if cluster.Parent != nil {
		parentChain := scraper.GetParentChain(*cluster.Parent)
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
		cluster.OverallStatus = ConvertManagedEntityStatusToValue(summary.OverallStatus)
	}
	return cluster
}

func ConvertManagedEntityStatusToValue(s types.ManagedEntityStatus) float64 {
	if s == types.ManagedEntityStatusRed {
		return 1.0
	} else if s == types.ManagedEntityStatusYellow {
		return 2.0
	} else if s == types.ManagedEntityStatusGreen {
		return 3.0
	}
	return 0
}
