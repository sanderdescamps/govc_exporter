package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	metricshelper "github.com/sanderdescamps/govc_exporter/internal/scraper/metrics_helper"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const HOST_SENSOR_NAME = "HostSensor"

type HostSensor struct {
	metricshelper.MetricHelperDefault
	logger.SensorLogger
	started       helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        SensorConfig
}

func NewHostSensor(scraper *VCenterScraper, config SensorConfig, l *slog.Logger) *HostSensor {
	return &HostSensor{
		config:              config,
		stopChan:            make(chan struct{}),
		SensorLogger:        logger.NewSLogLogger(l, logger.WithKind(HOST_SENSOR_NAME)),
		MetricHelperDefault: *metricshelper.NewMetricHelperDefault(HOST_SENSOR_NAME),
	}
}

func (s *HostSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	s.MetricHelperDefault.Start()
	client, release, err := scraper.clientPool.Acquire()
	if err != nil {
		return ErrSensorCientFailed
	}
	defer release()

	s.MetricHelperDefault.Mark1()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"HostSystem"},
		true,
	)
	if err != nil {
		return NewSensorError("failed to create container", "err", err)
	}
	defer v.Destroy(ctx)

	var entities []mo.HostSystem
	err = v.Retrieve(
		context.Background(),
		[]string{"HostSystem"},
		[]string{
			"name",
			"parent",
			"summary",
			"runtime",
			"config.storageDevice",
			"config.fileSystemVolume",
			// "network",
		},
		&entities,
	)
	s.MetricHelperDefault.Finish(err == nil)
	if err != nil {
		return NewSensorError("failed to retrieve data", "err", err)
	}

	for _, entity := range entities {
		entity.Summary.Hardware.OtherIdentifyingInfo = nil
		entity.Summary.Runtime = nil
		entity.Config = nil

		oHost := ConvertToHost(ctx, scraper, entity, time.Now())
		scraper.DB.SetHost(ctx, &oHost, s.config.MaxAge)
	}
	return nil
}

func (s *HostSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) {
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

func (s *HostSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *HostSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *HostSensor) Kind() string {
	return "HostSensor"
}

func (s *HostSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *HostSensor) Match(name string) bool {
	return helper.NewMatcher("keyword").Match(name)
}

func (s *HostSensor) Enabled() bool {
	return true
}

func ConvertToHost(ctx context.Context, scraper *VCenterScraper, h mo.HostSystem, t time.Time) objects.Host {
	self := objects.NewManagedObjectReference(h.Self.Type, h.Self.Value)

	var parent *objects.ManagedObjectReference
	if h.Parent != nil {
		p := objects.NewManagedObjectReference(parent.Type, parent.Value)
		parent = &p
	}

	host := objects.Host{
		Timestamp: t,
		Name:      h.Name,
		Self:      self,
		Parent:    parent,
	}

	if host.Parent != nil {
		parentChain := scraper.GetParentChain(*host.Parent)
		host.Cluster = parentChain.Cluster
		host.Datacenter = parentChain.DC
	}

	summary := h.Summary
	host.UptimeSeconds = float64(summary.QuickStats.Uptime)
	host.RebootRequired = summary.RebootRequired
	host.OverallStatus = ConvertManagedEntityStatusToValue(summary.OverallStatus)
	host.UsedCPUMhz = float64(summary.QuickStats.OverallCpuUsage)
	host.UsedMemBytes = float64(int64(summary.QuickStats.OverallMemoryUsage) * int64(1024*1024))
	host.NumberOfVMs = float64(len(h.Vm))
	host.NumberOfDatastores = float64(len(h.Datastore))

	if product := summary.Config.Product; product != nil {
		host.OSVersion = product.FullName
	}

	if hardware := summary.Hardware; hardware != nil {
		host.CPUCoresTotal = float64(hardware.NumCpuCores)
		host.CPUThreadsTotal = float64(hardware.NumCpuThreads)
		host.AvailCPUMhz = float64(int64(hardware.NumCpuCores) * int64(hardware.CpuMhz))
		host.AvailMemBytes = float64(hardware.MemorySize)

		for _, i := range hardware.OtherIdentifyingInfo {
			if i.IdentifierType.GetElementDescription().Key == "AssetTag" {
				host.AssetTag = i.IdentifierValue
			}
			if i.IdentifierType.GetElementDescription().Key == "ServiceTag" {
				host.AssetTag = i.IdentifierValue
			}
		}
		host.Vendor = hardware.Vendor
		host.Model = hardware.Model
	}

	runtime := summary.Runtime
	host.PowerState = fmt.Sprintf("%s", runtime.PowerState)
	host.ConnectionState = fmt.Sprintf("%s", runtime.ConnectionState)
	host.Maintenance = runtime.InMaintenanceMode

	if healthSystemRuntime := runtime.HealthSystemRuntime; healthSystemRuntime != nil {
		if systemHealthInfo := healthSystemRuntime.SystemHealthInfo; systemHealthInfo != nil {
			for _, info := range systemHealthInfo.NumericSensorInfo {
				host.SystemHealthNumerivSensor = append(host.SystemHealthNumerivSensor,
					objects.HostNumericSensorHealth{
						Name:  info.Name,
						Type:  info.SensorType,
						ID:    info.Id,
						Unit:  info.BaseUnits,
						Value: float64(info.CurrentReading),
						State: func() string {
							if state := info.HealthState.GetElementDescription(); state != nil {
								return state.Key
							}
							return ""
						}(),
					},
				)
			}
		}
	}

	if config := h.Config; config != nil {
		for _, adapter := range h.Config.StorageDevice.HostBusAdapter {

			hbaInterface := reflect.ValueOf(adapter).Elem().Interface()
			switch hba := hbaInterface.(type) {
			case types.HostInternetScsiHba:
				iscsiDiscoveryTarget := []objects.IscsiDiscoveryTarget{}
				for _, target := range hba.ConfiguredSendTarget {
					iscsiDiscoveryTarget = append(iscsiDiscoveryTarget, objects.IscsiDiscoveryTarget{
						Address: target.Address,
						Port:    target.Port,
					})
				}

				iscsiStaticTarget := []objects.IscsiStaticTarget{}
				for _, target := range hba.ConfiguredStaticTarget {
					iscsiStaticTarget = append(iscsiStaticTarget, objects.IscsiStaticTarget{
						Address:         target.Address,
						Port:            target.Port,
						IQN:             target.IScsiName,
						DiscoveryMethod: target.DiscoveryMethod,
					})
				}

				host.IscsiHBA = append(host.IscsiHBA, objects.IscsiHostBusAdapter{
					HostBusAdapter: objects.HostBusAdapter{
						Name:   hba.Device,
						Model:  cleanString(hba.Model),
						Driver: hba.Driver,
						State:  hba.Status,
					},
					IQN:             hba.IScsiName,
					DiscoveryTarget: iscsiDiscoveryTarget,
					StaticTarget:    iscsiStaticTarget,
				})
			// case types.HostBlockHba:
			// case types.HostFibreChannelHba:
			// case types.HostParallelScsiHba:
			// case types.HostPcieHba:
			// case types.HostRdmaDevice:
			// case types.HostSerialAttachedHba:
			// case types.HostTcpHba:
			default:
				baseHba := adapter.GetHostHostBusAdapter()
				host.GenericHBA = append(host.GenericHBA, objects.HostBusAdapter{
					Name:   baseHba.Device,
					Model:  cleanString(baseHba.Model),
					Driver: baseHba.Driver,
					State:  baseHba.Status,
				})
			}
		}
	}

	return host
}
