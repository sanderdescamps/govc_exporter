package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"math/rand"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
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
	(scraper.Datacenter).(*DatacenterSensor).WaitTillStartup()

	dcRefs := scraper.DB.GetAllDatacenterRefs(ctx)

	var wg sync.WaitGroup
	wg.Add(len(dcRefs))
	resultChan := make(chan *[]objects.Host, len(dcRefs))

	//query hosts by datacenter
	for _, dcRef := range dcRefs {
		go func() {
			defer wg.Done()

			ref := dcRef.ToVMwareRef()
			hosts, err := s.queryHostsInContainer(ctx, scraper, &ref, true)
			if err != nil {
				s.SensorLogger.Error("Failed to get hosts for datacenter", "datacenter", dcRef.Value, "err", err)
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

// Queries for all hosts in a container. If container==nil the root will be used.
func (s *HostSensor) queryHostsInContainer(ctx context.Context, scraper *VCenterScraper, containerRef *types.ManagedObjectReference, recursive bool) ([]objects.Host, error) {
	sensorStopwatch := sensormetrics.NewSensorStopwatch()
	sensorStopwatch.Start()
	client, release, err := scraper.clientPool.AcquireWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	if containerRef == nil {
		containerRef = &client.ServiceContent.RootFolder
	}

	sensorStopwatch.Mark1()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		*containerRef,
		[]string{"HostSystem"},
		recursive,
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
				host.AssetTag = cleanString(i.IdentifierValue)
			}
			if i.IdentifierType.GetElementDescription().Key == "ServiceTag" {
				host.ServiceTag = cleanString(i.IdentifierValue)
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

	host.Volumes = getVolumes(h)
	host.HBA = getHBAs(h)
	host.Luns = getSCSILuns(h)
	host.MultipathPathInfo = getMultipathInfo(h)

	return host
}

func getVolumes(host mo.HostSystem) []objects.Volume {
	volumes := []objects.Volume{}
	if host.Config != nil && host.Config.FileSystemVolume != nil {
		for _, info := range host.Config.FileSystemVolume.MountInfo {
			volumeInterface := reflect.ValueOf(info.Volume).Elem().Interface()
			switch volInfo := volumeInterface.(type) {
			case types.HostVmfsVolume:
				volumes = append(volumes, objects.Volume{
					AccessMode: info.MountInfo.AccessMode,
					Accessible: info.MountInfo.Accessible != nil && *info.MountInfo.Accessible,
					Mounted:    info.MountInfo.Mounted != nil && *info.MountInfo.Mounted,
					Path:       info.MountInfo.Path,
					Capacity:   volInfo.Capacity,
					Name:       volInfo.Name,
					Type:       volInfo.Type,
					UUID:       volInfo.Uuid,
					DiskName: func() string {
						for _, extend := range volInfo.Extent {
							return extend.DiskName
						}
						return ""
					}(),
					SSD:   volInfo.Ssd != nil && *volInfo.Ssd,
					Local: volInfo.Local == nil || *volInfo.Local,
				})
			}
		}
	}
	return volumes
}

func multipathPathCounter(host mo.HostSystem, lunKey string) (active int64, total int64) {
	if config := host.Config; config != nil {
		if storDev := config.StorageDevice; storDev != nil {
			if mpInfo := storDev.MultipathInfo; mpInfo != nil {
				for _, lun := range mpInfo.Lun {
					if lun.Lun == lunKey {
						for _, path := range lun.Path {
							if strings.EqualFold(path.State, "active") {
								active += 1
							}
							total += 1
						}
						return
					}

				}
			}
		}
	}
	return 0, 0
}

func canonicalNameLookup(host mo.HostSystem, lunKey string) string {
	if config := host.Config; config != nil {
		if storDev := config.StorageDevice; storDev != nil {
			for _, scsiLun := range storDev.ScsiLun {
				switch disk := (reflect.ValueOf(scsiLun).Elem().Interface()).(type) {
				case types.HostScsiDisk:
					return disk.CanonicalName
				default:
					continue
				}
			}
		}
	}
	return ""
}

func getSCSILuns(host mo.HostSystem) []objects.SCSILun {
	res := []objects.SCSILun{}
	if host.Config != nil && host.Config.StorageDevice != nil {
		for _, sl := range host.Config.StorageDevice.ScsiLun {

			switch disk := (reflect.ValueOf(sl).Elem().Interface()).(type) {
			case types.HostScsiDisk:
				activePaths, totalPaths := multipathPathCounter(host, disk.Key)
				res = append(res, objects.SCSILun{
					CanonicalName:     disk.CanonicalName,
					Vendor:            cleanString(disk.Vendor),
					Model:             cleanString(disk.Model),
					ActiveNumberPaths: activePaths,
					TotalNumberPaths:  totalPaths,
					SSD:               disk.Ssd != nil && *disk.Ssd,
					Local:             disk.LocalDisk == nil || *disk.LocalDisk,
				})
			default:
				continue
			}
		}
	}
	return res
}

func getHBAs(host mo.HostSystem) []objects.HostBusAdapter {
	res := []objects.HostBusAdapter{}
	if config := host.Config; config != nil {
		if storDev := config.StorageDevice; storDev != nil {
			for _, adapter := range storDev.HostBusAdapter {
				ihba := reflect.ValueOf(adapter).Elem().Interface()
				switch hba := ihba.(type) {
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

					res = append(res, objects.HostBusAdapter{
						Type:                 "iscsi",
						Device:               hba.Device,
						Model:                cleanString(hba.Model),
						Driver:               hba.Driver,
						Status:               hba.Status,
						Protocol:             hba.StorageProtocol,
						IscsiInitiatorIQN:    hba.IScsiName,
						IscsiDiscoveryTarget: iscsiDiscoveryTarget,
						IscsiStaticTarget:    iscsiStaticTarget,
					})
				// case types.HostTcpHba:
				// case types.HostBlockHba:
				// case types.HostFibreChannelHba:
				// case types.HostParallelScsiHba:
				// case types.HostPcieHba:
				// case types.HostRdmaDevice:
				// case types.HostSerialAttachedHba:
				default:
					baseHba := adapter.GetHostHostBusAdapter()
					res = append(res, objects.HostBusAdapter{
						Type:     "generic",
						Device:   baseHba.Device,
						Model:    cleanString(baseHba.Model),
						Driver:   baseHba.Driver,
						Status:   baseHba.Status,
						Protocol: baseHba.StorageProtocol,
					})
				}
			}
		}
	}
	return res
}

func getMultipathInfo(host mo.HostSystem) []objects.MultipathPathInfo {
	res := []objects.MultipathPathInfo{}
	if config := host.Config; config != nil {
		if storDev := config.StorageDevice; storDev != nil {
			if mpInfo := storDev.MultipathInfo; mpInfo != nil {
				for _, lun := range mpInfo.Lun {
					canonicalName := canonicalNameLookup(host, lun.Key)
					for _, path := range lun.Path {
						mpPathInfo := objects.MultipathPathInfo{
							Name:          path.Name,
							State:         path.State,
							Adapter:       "",
							CanonicalName: canonicalName,
						}

						if r := regexp.MustCompile(`(\w+)-([\w\.]+)-(\w+)`); r.MatchString(path.Adapter) {
							mpPathInfo.Adapter = r.FindStringSubmatch(path.Adapter)[3]
						}

						if r := regexp.MustCompile(`(\w+):(\w+):(\w+):L(\d+)`); r.MatchString(path.Name) {
							lunNr, _ := strconv.Atoi(r.FindStringSubmatch(path.Name)[4])
							mpPathInfo.LUN = lunNr
						}

						iTransport := reflect.ValueOf(path.Transport).Elem().Interface()
						switch transport := iTransport.(type) {
						case types.HostInternetScsiTargetTransport:
							mpPathInfo.Type = "iscsi"
							if len(transport.Address) == 1 {
								mpPathInfo.IscsiTargetAddress = transport.Address[0]
							}
							mpPathInfo.IscsiTargetIQN = transport.IScsiName
						default:
							mpPathInfo.Type = "generic"
						}

						res = append(res, mpPathInfo)
					}
				}
			}
		}
	}
	return res
}
