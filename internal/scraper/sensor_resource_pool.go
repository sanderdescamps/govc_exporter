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
)

const RESOURCE_POOL_SENSOR_NAME = "ResourcePoolSensor"

type ResourcePoolSensor struct {
	BaseSensor
	logger.SensorLogger
	metricshelper.MetricHelperDefault
	started       *helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        SensorConfig
}

func NewResourcePoolSensor(scraper *VCenterScraper, config SensorConfig, l *slog.Logger) *ResourcePoolSensor {
	return &ResourcePoolSensor{
		BaseSensor: *NewBaseSensor(
			"ResourcePool", []string{
				"parent",
				"name",
				"summary",
			}),
		started:             helper.NewStartedCheck(),
		stopChan:            make(chan struct{}),
		manualRefresh:       make(chan struct{}),
		config:              config,
		SensorLogger:        logger.NewSLogLogger(l, logger.WithKind(RESOURCE_POOL_SENSOR_NAME)),
		MetricHelperDefault: *metricshelper.NewMetricHelperDefault(RESOURCE_POOL_SENSOR_NAME),
	}
}

func (s *ResourcePoolSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	var resourcePools []mo.ResourcePool
	stats, err := s.baseRefresh(ctx, scraper, &resourcePools)
	s.MetricHelperDefault.LoadStats(stats)
	if err != nil {
		return err
	}

	for _, rp := range resourcePools {
		oRPool := ConvertToResourcePool(ctx, scraper, rp, time.Now())
		err := scraper.DB.SetResourcePool(ctx, oRPool, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ResourcePoolSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *ResourcePoolSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *ResourcePoolSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *ResourcePoolSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *ResourcePoolSensor) Kind() string {
	return "ResourcePoolSensor"
}

func (s *ResourcePoolSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *ResourcePoolSensor) Match(name string) bool {
	return helper.NewMatcher("resource_pool", "resourcepool", "repool").Match(name)
}

func (s *ResourcePoolSensor) Enabled() bool {
	return true
}

func ConvertToResourcePool(ctx context.Context, scraper *VCenterScraper, p mo.ResourcePool, t time.Time) objects.ResourcePool {
	self := objects.NewManagedObjectReferenceFromVMwareRef(p.Self)

	var parent *objects.ManagedObjectReference
	if p.Parent != nil {
		p := objects.NewManagedObjectReferenceFromVMwareRef(*p.Parent)
		parent = &p
	}

	pool := objects.ResourcePool{
		Timestamp: t,
		Name:      p.Name,
		Self:      self,
		Parent:    parent,
	}

	if pool.Parent != nil {
		parentChain := scraper.DB.GetParentChain(ctx, *pool.Parent)
		pool.Datacenter = parentChain.DC
	}

	mb := int64(1024 * 1024)
	if summary := p.Summary.GetResourcePoolSummary(); summary != nil {
		pool.OverallCPUUsage = float64(summary.QuickStats.OverallCpuUsage)
		pool.OverallCPUDemand = float64(summary.QuickStats.OverallCpuDemand)
		pool.GuestMemoryUsage = float64(summary.QuickStats.GuestMemoryUsage * mb)
		pool.HostMemoryUsage = float64(summary.QuickStats.HostMemoryUsage * mb)
		pool.DistributedCPUEntitlement = float64(summary.QuickStats.DistributedCpuEntitlement)
		pool.DistributedMemoryEntitlement = float64(summary.QuickStats.DistributedMemoryEntitlement * mb)
		pool.StaticCPUEntitlement = float64(summary.QuickStats.StaticCpuEntitlement)
		pool.PrivateMemory = float64(summary.QuickStats.PrivateMemory * mb)
		pool.SwappedMemory = float64(summary.QuickStats.SwappedMemory * mb)
		pool.BalloonedMemory = float64(summary.QuickStats.BalloonedMemory * mb)
		pool.OverheadMemory = float64(summary.QuickStats.OverheadMemory * mb)
		pool.ConsumedOverheadMemory = float64(summary.QuickStats.ConsumedOverheadMemory * mb)
		if limit := summary.Config.MemoryAllocation.Limit; limit != nil {
			pool.MemoryAllocationLimit = float64((*limit) * mb)
		}
		if limit := summary.Config.CpuAllocation.Limit; limit != nil {
			pool.CPUAllocationLimit = float64((*limit) * mb)
		}
		pool.OverallStatus = string(summary.Runtime.OverallStatus)
	}

	return pool
}
