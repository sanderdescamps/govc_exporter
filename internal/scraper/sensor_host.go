package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"math/rand"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	sensormetrics "github.com/sanderdescamps/govc_exporter/internal/scraper/sensor_metrics"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const HOST_SENSOR_NAME = "HostSensor"

type HostSensor struct {
	metricsCollector *sensormetrics.SensorMetricsCollector
	statusMonitor    *sensormetrics.StatusMonitor
	logger.SensorLogger
	started       *helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        config.SensorConfig
}

func NewHostSensor(scraper *VCenterScraper, config config.SensorConfig, l *slog.Logger) *HostSensor {
	var mc *sensormetrics.SensorMetricsCollector = sensormetrics.NewAvgSensorMetricsCollector(50)
	var sm *sensormetrics.StatusMonitor = sensormetrics.NewStatusMonitor()
	return &HostSensor{
		config:           config,
		started:          helper.NewStartedCheck(),
		stopChan:         make(chan struct{}),
		manualRefresh:    make(chan struct{}),
		SensorLogger:     logger.NewSLogLogger(l, logger.WithKind(HOST_SENSOR_NAME)),
		metricsCollector: mc,
		statusMonitor:    sm,
	}
}

func (s *HostSensor) querryAllHosts(ctx context.Context, scraper *VCenterScraper) ([]objects.Host, error) {
	if scraper.Cluster == nil {
		s.SensorLogger.Error("Can't query for hosts if cluster sensor is not defined")
		return nil, fmt.Errorf("no cluster sensor found")
	}
	(scraper.Cluster).(*ClusterSensor).WaitTillStartup()

	clusterRefs := scraper.DB.GetAllClusterRefs(ctx)

	var wg sync.WaitGroup
	wg.Add(len(clusterRefs))
	resultChan := make(chan *[]objects.Host, len(clusterRefs))
	for _, clusterRef := range clusterRefs {
		go func() {
			defer wg.Done()

			hosts, err := s.queryHostsForCluster(ctx, scraper, clusterRef.ToVMwareRef())
			if err != nil {
				s.SensorLogger.Error("Failed to get hosts for cluster", "cluster", clusterRef.Value, "err", err)
				s.statusMonitor.Fail()
				return
			}
			s.statusMonitor.Success()
			resultChan <- &hosts
		}()
	}

	readyChan := make(chan bool)
	allHosts := map[objects.ManagedObjectReference]objects.Host{}

	go func() {
		for {
			select {
			case r := <-resultChan:
				for _, host := range *r {
					if _, exist := allHosts[host.Self]; !exist {
						allHosts[host.Self] = host
					} else {
						s.SensorLogger.Error("host exist on multiple clusters", "host", host.Name, "ref", host.Self)
					}
				}
			case <-readyChan:
				close(readyChan)
				close(resultChan)
				return
			case <-ctx.Done():
				close(resultChan)
				return
			}
		}
	}()

	wg.Wait()
	readyChan <- true

	return slices.Collect(maps.Values(allHosts)), nil
}

func (s *HostSensor) queryHostsForCluster(ctx context.Context, scraper *VCenterScraper, clusterRef types.ManagedObjectReference) ([]objects.Host, error) {
	sensorStopwatch := sensormetrics.NewSensorStopwatch()
	sensorStopwatch.Start()
	client, release, err := scraper.clientPool.AcquireWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	sensorStopwatch.Mark1()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		clusterRef,
		[]string{"HostSystem"},
		true,
	)
	if err != nil {
		return nil, NewSensorError("failed to create container", "err", err)
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
			"vm",
		},
		&entities,
	)
	sensorStopwatch.Finish()
	s.metricsCollector.UploadStats(sensorStopwatch.GetStats())

	if err != nil {
		return nil, err
	}

	oHosts := []objects.Host{}
	for _, item := range entities {
		oHosts = append(oHosts, ConvertToHost(ctx, scraper, item, time.Now()))
	}

	return oHosts, nil
}

func (s *HostSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	hosts, err := s.querryAllHosts(ctx, scraper)
	if err != nil {
		return err
	}

	for _, host := range hosts {
		err := scraper.DB.SetHost(ctx, host, s.config.MaxAge)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *HostSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *HostSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *HostSensor) GetLatestMetrics() []sensormetrics.SensorMetric {
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
				if path.Transport != nil {
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
