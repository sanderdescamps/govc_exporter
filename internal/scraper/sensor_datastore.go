package scraper

import (
	"context"
	"log/slog"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	sensormetrics "github.com/sanderdescamps/govc_exporter/internal/scraper/sensor_metrics"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const DATASTORE_SENSOR_NAME = "DatastoreSensor"

type DatastoreSensor struct {
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

func NewDatastoreSensor(scraper *VCenterScraper, config config.SensorConfig, l *slog.Logger) *DatastoreSensor {
	var mc *sensormetrics.SensorMetricsCollector = sensormetrics.NewLastSensorMetricsCollector()
	var sm *sensormetrics.StatusMonitor = sensormetrics.NewStatusMonitor()
	return &DatastoreSensor{
		BaseSensor: *NewBaseSensor(
			"Datastore", []string{
				"name",
				"parent",
				"summary",
				"info",
			}, mc, sm),
		started:          helper.NewStartedCheck(),
		stopChan:         make(chan struct{}),
		manualRefresh:    make(chan struct{}),
		config:           config,
		SensorLogger:     logger.NewSLogLogger(l, logger.WithKind(DATASTORE_SENSOR_NAME)),
		metricsCollector: mc,
		statusMonitor:    sm,
	}
}

func (s *DatastoreSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	var datastores []mo.Datastore
	err := s.baseRefresh(ctx, scraper, &datastores)
	if err != nil {
		return err
	}

	for _, ds := range datastores {
		oDS := ConvertToDatastore(ctx, scraper, ds, time.Now())
		err := scraper.DB.SetDatastore(ctx, oDS, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DatastoreSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *DatastoreSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *DatastoreSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *DatastoreSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *DatastoreSensor) Kind() string {
	return "DatastoreSensor"
}

func (s *DatastoreSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *DatastoreSensor) Match(name string) bool {
	return helper.NewMatcher("datastore", "ds").Match(name)
}

func (s *DatastoreSensor) Enabled() bool {
	return true
}

func (s *DatastoreSensor) GetLatestMetrics() []sensormetrics.SensorMetric {
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

func ConvertToDatastore(ctx context.Context, scraper *VCenterScraper, d mo.Datastore, t time.Time) objects.Datastore {
	self := objects.NewManagedObjectReferenceFromVMwareRef(d.Self)

	var parent *objects.ManagedObjectReference
	if d.Parent != nil {
		p := objects.NewManagedObjectReferenceFromVMwareRef(*d.Parent)
		parent = &p
	}

	kind := ExtractKindFromDatastore(d)
	var vmfsInfo *objects.DatastoreVmfsInfo
	if kind == "vmfs" {
		vmfsInfo = ExtractVmfsInfoFromDatastore(d)
	}

	datastore := objects.Datastore{
		Timestamp: t,
		Name:      d.Name,
		Kind:      kind,
		Self:      self,
		Parent:    parent,
		VmfsInfo:  vmfsInfo,
	}

	if datastore.Parent != nil {
		parentChain := scraper.DB.GetParentChain(ctx, *datastore.Parent)
		datastore.DatastoreCluster = parentChain.SPOD
	}

	summary := d.Summary
	datastore.Accessible = summary.Accessible
	datastore.Capacity = float64(summary.Capacity)
	datastore.FreeSpace = float64(summary.FreeSpace)
	datastore.Maintenance = summary.MaintenanceMode

	for _, hostMountInfo := range d.Host {
		hostMountInfoRef := objects.NewManagedObjectReferenceFromVMwareRef(hostMountInfo.Key)
		host := scraper.DB.GetHost(ctx, hostMountInfoRef)
		datastore.HostMountInfo = append(datastore.HostMountInfo, objects.DatastoreHostMountInfo{
			Host:            host.Name,
			HostID:          hostMountInfo.Key.Value,
			Accessible:      hostMountInfo.MountInfo.Accessible != nil && *hostMountInfo.MountInfo.Accessible,
			Mounted:         hostMountInfo.MountInfo.Mounted != nil && *hostMountInfo.MountInfo.Mounted,
			VmknicActiveNic: hostMountInfo.MountInfo.VmknicActive != nil && *hostMountInfo.MountInfo.VmknicActive,
		})
	}

	datastore.OverallStatus = string(d.OverallStatus)

	return datastore
}

func ExtractKindFromDatastore(d mo.Datastore) string {
	iInfo := reflect.ValueOf(d.Info).Elem().Interface()
	switch iInfo.(type) {
	case types.LocalDatastoreInfo:
		return "local"
	case types.VmfsDatastoreInfo:
		return "vmfs"
	case types.NasDatastoreInfo:
		return "nas"
	case types.PMemDatastoreInfo:
		return "pmem"
	case types.VsanDatastoreInfo:
		return "vsan"
	case types.VvolDatastoreInfo:
		return "vvol"
	}
	return ""
}

func ExtractVmfsInfoFromDatastore(d mo.Datastore) *objects.DatastoreVmfsInfo {
	iInfo := reflect.ValueOf(d.Info).Elem().Interface()
	switch parsedInfo := iInfo.(type) {
	case types.VmfsDatastoreInfo:
		if vmfs := parsedInfo.Vmfs; vmfs != nil {
			return &objects.DatastoreVmfsInfo{
				Name:  parsedInfo.Name,
				UUID:  vmfs.Uuid,
				SSD:   vmfs.Ssd != nil && *vmfs.Ssd,
				Local: vmfs.Local == nil || *vmfs.Local,
				NAA: func() string {
					for _, extent := range vmfs.Extent {
						return extent.DiskName
					}
					return ""
				}(),
			}
		}
	}
	return nil
}
