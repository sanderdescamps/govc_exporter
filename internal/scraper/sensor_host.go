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
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const HOST_SENSOR_NAME = "HostSensor"

type HostSensor struct {
	metricshelper.MetricHelperDefault
	logger.SensorLogger
	started       *helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        SensorConfig
}

func NewHostSensor(scraper *VCenterScraper, config SensorConfig, l *slog.Logger) *HostSensor {
	return &HostSensor{
		config:              config,
		started:             helper.NewStartedCheck(),
		stopChan:            make(chan struct{}),
		manualRefresh:       make(chan struct{}),
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
		ctx,
		[]string{"HostSystem"},
		[]string{
			"name",
			"parent",
			"summary",
			"runtime",
			"config.storageDevice",
			"config.fileSystemVolume",
			// "network",
			"hardware",
		},
		&entities,
	)
	s.MetricHelperDefault.Finish(err == nil)
	if err != nil {
		return NewSensorError("failed to retrieve data", "err", err)
	}

	for _, entity := range entities {
		oHost := ConvertToHost(ctx, scraper, entity, time.Now())
		scraper.DB.SetHost(ctx, oHost, s.config.MaxAge)
	}
	return nil
}

func (s *HostSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *HostSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
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
	return helper.NewMatcher("host", "esx").Match(name)
}

func (s *HostSensor) Enabled() bool {
	return true
}

func ConvertToHost(ctx context.Context, scraper *VCenterScraper, h mo.HostSystem, t time.Time) objects.Host {
	self := objects.NewManagedObjectReferenceFromVMwareRef(h.Self)

	var parent *objects.ManagedObjectReference
	if h.Parent != nil {
		p := objects.NewManagedObjectReferenceFromVMwareRef(*h.Parent)
		parent = &p
	}

	host := objects.Host{
		Timestamp: t,
		Name:      h.Name,
		Self:      self,
		Parent:    parent,
	}

	if host.Parent != nil {
		parentChain := scraper.DB.GetParentChain(ctx, *host.Parent)
		host.Cluster = parentChain.Cluster
		host.Datacenter = parentChain.DC
	}

	summary := h.Summary
	host.UptimeSeconds = float64(summary.QuickStats.Uptime)
	host.RebootRequired = summary.RebootRequired
	host.UsedCPUMhz = float64(summary.QuickStats.OverallCpuUsage)
	host.UsedMemBytes = float64(int64(summary.QuickStats.OverallMemoryUsage) * int64(1024*1024))
	host.NumberOfVMs = float64(len(h.Vm))
	host.NumberOfDatastores = float64(len(h.Datastore))
	host.OverallStatus = string(summary.OverallStatus)

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
				host.ServiceTag = i.IdentifierValue
			}
		}
		host.Vendor = hardware.Vendor
		host.Model = hardware.Model
	}
	if h.Hardware != nil && h.Hardware.BiosInfo != nil {
		host.BiosVersion = h.Hardware.BiosInfo.BiosVersion
	}

	runtime := h.Runtime
	host.PowerState = string(runtime.PowerState)
	host.ConnectionState = string(runtime.ConnectionState)
	host.Maintenance = runtime.InMaintenanceMode

	if healthSystemRuntime := runtime.HealthSystemRuntime; healthSystemRuntime != nil {
		if systemHealthInfo := healthSystemRuntime.SystemHealthInfo; systemHealthInfo != nil {
			for _, info := range systemHealthInfo.NumericSensorInfo {
				host.SystemHealthNumericSensors = append(host.SystemHealthNumericSensors,
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

		if hardwareStatusInfo := healthSystemRuntime.HardwareStatusInfo; hardwareStatusInfo != nil {
			for _, baseInfo := range hardwareStatusInfo.CpuStatusInfo {
				info := baseInfo.GetHostHardwareElementInfo()
				if info != nil {
					host.HardwareStatus = append(host.HardwareStatus,
						objects.HardwareStatus{
							Name:    info.Name,
							Type:    "cpu",
							Status:  string(info.Status.GetElementDescription().Key),
							Summary: string(info.Status.GetElementDescription().Summary),
						},
					)
				}
			}
			for _, baseInfo := range hardwareStatusInfo.MemoryStatusInfo {
				info := baseInfo.GetHostHardwareElementInfo()
				if info != nil {
					host.HardwareStatus = append(host.HardwareStatus,
						objects.HardwareStatus{
							Name:    info.Name,
							Type:    "memory",
							Status:  string(info.Status.GetElementDescription().Key),
							Summary: string(info.Status.GetElementDescription().Summary),
						},
					)
				}
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
						Name:     hba.Device,
						Model:    cleanString(hba.Model),
						Protocol: hba.StorageProtocol,
						Driver:   hba.Driver,
						State:    hba.Status,
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
					Name:     baseHba.Device,
					Model:    cleanString(baseHba.Model),
					Driver:   baseHba.Driver,
					State:    baseHba.Status,
					Protocol: baseHba.StorageProtocol,
				})
			}
		}
	}

	host.IscsiMultiPathPaths = getHostMultiPathPaths(h)
	host.SCSILuns = getAllScsiLuns(h)

	return host
}

func getHostMultiPathPaths(host mo.HostSystem) []objects.IscsiPath {
	hbaDeviceNameFunc := getHostHBADeviceNameLookupFunc(host)
	lunCanonnicalNameLookup := getHostLunCanonicalNameLookupFunc(host)
	scsiLunLookupFunc := getHostScsiLunLookupFunc(host)

	iscsiPaths := []objects.IscsiPath{}
	if host.Config != nil && host.Config.StorageDevice != nil {
		for _, logicalUnit := range host.Config.StorageDevice.MultipathInfo.Lun {
			uuid := logicalUnit.Id
			naa := lunCanonnicalNameLookup(uuid)
			scsiLun := scsiLunLookupFunc(naa)
			for _, path := range logicalUnit.Path {
				device := hbaDeviceNameFunc(path.Adapter)

				transportInterface := reflect.ValueOf(path.Transport).Elem().Interface()
				switch transport := transportInterface.(type) {
				case types.HostInternetScsiTargetTransport:
					for _, address := range transport.Address {
						iscsiPaths = append(iscsiPaths, objects.IscsiPath{
							Device:        device,
							NAA:           naa,
							TargetIQN:     transport.IScsiName,
							TargetAddress: address,
							DatastoreName: scsiLun.Datastore,
							State:         path.State,
						})
					}
					// default:
					// 	iscsiPaths = append(iscsiPaths, iscsiPath{
					// 		Device:        device,
					// 		NAA:           naa,
					// 		TargetIQN:     "",
					// 		TargetAddress: "",
					// 		DatastoreName: scsiLun.Datastore,
					// 		State:         path.State,
					// 	})
				}
			}
		}
	}

	return iscsiPaths
}

func getHostHBADeviceNameLookupFunc(host mo.HostSystem) func(hbaKey string) string {
	hbaDevices := map[string]string{}
	if host.Config != nil && host.Config.StorageDevice != nil {
		for _, adapter := range host.Config.StorageDevice.HostBusAdapter {
			hbaDevices[adapter.GetHostHostBusAdapter().Key] = adapter.GetHostHostBusAdapter().Device
		}
	}
	return func(hbaKey string) string {
		if val, ok := hbaDevices[hbaKey]; ok {
			return val
		}
		return ""
	}
}

func getHostLunCanonicalNameLookupFunc(host mo.HostSystem) func(lunID string) string {
	naas := map[string]string{}
	if host.Config != nil && host.Config.StorageDevice != nil {
		for _, lun := range host.Config.StorageDevice.ScsiLun {
			naas[lun.GetScsiLun().Uuid] = lun.GetScsiLun().CanonicalName
		}
	}

	return func(lunID string) string {
		if val, ok := naas[lunID]; ok {
			return val
		}
		return ""
	}
}

func getHostScsiLunLookupFunc(host mo.HostSystem) func(canonicalName string) objects.ScsiLun {
	scsiLuns := map[string]objects.ScsiLun{}

	for _, l := range getAllScsiLuns(host) {
		scsiLuns[l.CanonicalName] = l
	}

	return func(canonicalName string) objects.ScsiLun {
		if val, ok := scsiLuns[canonicalName]; ok {
			return val
		}
		return objects.ScsiLun{}
	}
}

func getAllScsiLuns(host mo.HostSystem) []objects.ScsiLun {
	mountInfoMap := map[string]types.HostFileSystemMountInfo{}
	if host.Config != nil && host.Config.FileSystemVolume != nil {
		for _, info := range host.Config.FileSystemVolume.MountInfo {
			volumeInterface := reflect.ValueOf(info.Volume).Elem().Interface()
			switch volume := volumeInterface.(type) {
			case types.HostVmfsVolume:
				for _, extend := range volume.Extent {
					mountInfoMap[extend.DiskName] = info
				}
			default:
				continue
			}
		}
	}

	scsiLuns := map[string]types.ScsiLun{}
	if host.Config != nil && host.Config.StorageDevice != nil {
		for _, l := range host.Config.StorageDevice.ScsiLun {
			lun := l.GetScsiLun()
			if lun != nil {
				scsiLuns[lun.CanonicalName] = *lun
			}
		}
	}

	result := []objects.ScsiLun{}
	if host.Config != nil && host.Config.StorageDevice != nil {
		for _, l := range host.Config.StorageDevice.ScsiLun {
			lun := l.GetScsiLun()
			if mountInfo, ok := mountInfoMap[lun.CanonicalName]; ok {
				volume := mountInfo.Volume.GetHostFileSystemVolume()
				result = append(result, objects.ScsiLun{
					CanonicalName: lun.CanonicalName,
					Model:         cleanString(lun.Model),
					Vendor:        cleanString(lun.Vendor),
					Datastore: func() string {
						if volume != nil {
							return volume.Name
						}
						return "unknown"
					}(),
					Mounted: func() bool {
						if mountInfo.MountInfo.Mounted != nil {
							return *mountInfo.MountInfo.Mounted
						}
						return false
					}(),
					Accessible: func() bool {
						if mountInfo.MountInfo.Accessible != nil {
							return *mountInfo.MountInfo.Accessible
						}
						return false
					}(),
				})
			}

		}
	}
	return result
}
