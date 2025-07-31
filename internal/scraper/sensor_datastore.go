package scraper

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	metricshelper "github.com/sanderdescamps/govc_exporter/internal/scraper/metrics_helper"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const DATASTORE_SENSOR_NAME = "DatastoreSensor"

type DatastoreSensor struct {
	BaseSensor
	logger.SensorLogger
	metricshelper.MetricHelperDefault
	started       helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        SensorConfig
}

func NewDatastoreSensor(scraper *VCenterScraper, config SensorConfig, l *slog.Logger) *DatastoreSensor {
	return &DatastoreSensor{
		BaseSensor: *NewBaseSensor(
			"Datastore", []string{
				"name",
				"parent",
				"summary",
				"info",
			}),
		stopChan:            make(chan struct{}),
		config:              config,
		SensorLogger:        logger.NewSLogLogger(l, logger.WithKind(DATASTORE_SENSOR_NAME)),
		MetricHelperDefault: *metricshelper.NewMetricHelperDefault(DATASTORE_SENSOR_NAME),
	}
}

func (s *DatastoreSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	var datastores []mo.Datastore
	stats, err := s.baseRefresh(ctx, scraper, &datastores)
	s.MetricHelperDefault.LoadStats(stats)
	if err != nil {
		return err
	}

	for _, ds := range datastores {
		oDS := ConvertToDatastore(ctx, scraper, ds, time.Now())
		err := scraper.DB.SetDatastore(ctx, &oDS, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DatastoreSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) {
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
	return helper.NewMatcher("keyword").Match(name)
}

func (s *DatastoreSensor) Enabled() bool {
	return true
}

func ConvertToDatastore(ctx context.Context, scraper *VCenterScraper, d mo.Datastore, t time.Time) objects.Datastore {
	self := objects.NewManagedObjectReference(d.Self.Type, d.Self.Value)

	var parent *objects.ManagedObjectReference
	if d.Parent != nil {
		p := objects.NewManagedObjectReference(parent.Type, parent.Value)
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
		parentChain := scraper.GetParentChain(*datastore.Parent)
		datastore.Cluster = parentChain.SPOD
	}

	summary := d.Summary
	datastore.Accessible = summary.Accessible
	datastore.Capacity = float64(summary.Capacity)
	datastore.FreeSpace = float64(summary.FreeSpace)
	datastore.Maintenance = summary.MaintenanceMode
	datastore.OverallStatus = ConvertManagedEntityStatusToValue(d.OverallStatus)

	for _, hostMountInfo := range d.Host {
		hostMountInfoRef := objects.NewManagedObjectReference(hostMountInfo.Key.Type, hostMountInfo.Key.Value)
		host := scraper.DB.GetHost(ctx, hostMountInfoRef)
		datastore.HostMountInfo = append(datastore.HostMountInfo, objects.DatastoreHostMountInfo{
			Host:            host.Name,
			HostID:          hostMountInfo.Key.Value,
			Accessable:      hostMountInfo.MountInfo.Accessible != nil && *hostMountInfo.MountInfo.Accessible,
			Mounted:         hostMountInfo.MountInfo.Mounted != nil && *hostMountInfo.MountInfo.Mounted,
			VmknicActiveNic: hostMountInfo.MountInfo.VmknicActive != nil && *hostMountInfo.MountInfo.VmknicActive,
		})
	}

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
