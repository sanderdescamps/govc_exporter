package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"math/rand"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	sensormetrics "github.com/sanderdescamps/govc_exporter/internal/scraper/sensor_metrics"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const VM_SENSOR_NAME = "VirtualMachineSensor"

var regexPatternIPv4 = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])$`)

type VirtualMachineSensor struct {
	logger.SensorLogger
	metricsCollector *sensormetrics.SensorMetricsCollector
	statusMonitor    *sensormetrics.StatusMonitor
	started          *helper.StartedCheck
	sensorLock       sync.Mutex
	manualRefresh    chan struct{}
	stopChan         chan struct{}
	config           config.SensorConfig

	// moType       string
	// moProperties []string
}

func NewVirtualMachineSensor(scraper *VCenterScraper, config config.SensorConfig, l *slog.Logger) *VirtualMachineSensor {
	var mc *sensormetrics.SensorMetricsCollector = sensormetrics.NewAvgSensorMetricsCollector(100)
	var sm *sensormetrics.StatusMonitor = sensormetrics.NewStatusMonitor()
	var sensor VirtualMachineSensor = VirtualMachineSensor{
		started:          helper.NewStartedCheck(),
		stopChan:         make(chan struct{}),
		manualRefresh:    make(chan struct{}),
		config:           config,
		SensorLogger:     logger.NewSLogLogger(l, logger.WithKind(VM_SENSOR_NAME)),
		metricsCollector: mc,
		statusMonitor:    sm,
	}

	return &sensor
}

func (s *VirtualMachineSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	vms, err := s.querryAllVMs(ctx, scraper)
	if err != nil {
		return err
	}

	for _, vm := range vms {
		err := scraper.DB.SetVM(ctx, vm, s.config.MaxAge)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *VirtualMachineSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *VirtualMachineSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *VirtualMachineSensor) querryAllVMs(ctx context.Context, scraper *VCenterScraper) ([]objects.VirtualMachine, error) {
	if scraper.Host == nil {
		s.SensorLogger.Error("Can't query for vm's if host sensor is not defined")
		return nil, fmt.Errorf("no host sensor found")
	}
	(scraper.Host).(*HostSensor).WaitTillStartup()

	hostRefs := scraper.DB.GetAllHostRefs(ctx)

	var wg sync.WaitGroup
	wg.Add(len(hostRefs))
	resultChan := make(chan *[]objects.VirtualMachine, len(hostRefs))
	for _, hostRef := range hostRefs {
		go func() {
			defer wg.Done()

			vms, err := s.queryVmsForHost(ctx, scraper, hostRef.ToVMwareRef())
			if err != nil {
				s.SensorLogger.Error("Failed to get vm's for host", "host", hostRef.Value, "err", err)
				s.statusMonitor.Fail()
				return
			}
			s.statusMonitor.Success()
			resultChan <- &vms
		}()
	}

	readyChan := make(chan bool)
	allVMs := map[objects.ManagedObjectReference]objects.VirtualMachine{}

	go func() {
		for {
			select {
			case r := <-resultChan:
				for _, vm := range *r {
					if other, exist := allVMs[vm.Self]; exist {
						if vm.TimeConfigChanged.After(other.TimeConfigChanged) {
							allVMs[vm.Self] = vm
						}
					} else {
						allVMs[vm.Self] = vm
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

	return slices.Collect(maps.Values(allVMs)), nil
}

func (s *VirtualMachineSensor) queryVmsForHost(ctx context.Context, scraper *VCenterScraper, hostRef types.ManagedObjectReference) ([]objects.VirtualMachine, error) {
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
		hostRef,
		[]string{"VirtualMachine"},
		true,
	)
	if err != nil {
		return nil, err
	}
	defer v.Destroy(ctx)

	var items []mo.VirtualMachine
	err = v.Retrieve(
		ctx,
		[]string{"VirtualMachine"},
		[]string{
			"name",
			"config",
			//"datatore",
			"guest", //(only for advanced network)
			// "guestHeartbeatStatus", //(not sure)
			// "network",
			"parent",
			// "resourceConfig",
			"resourcePool",
			"runtime",
			// "snapshot",
			"summary",
		},
		&items,
	)
	sensorStopwatch.Finish()
	if err != nil {
		return nil, err
	}

	oVMs := []objects.VirtualMachine{}
	for _, item := range items {
		oVMs = append(oVMs, ConvertToVirtualMachine(ctx, scraper, item, time.Now()))
	}

	s.metricsCollector.UploadStats(sensorStopwatch.GetStats())

	return oVMs, nil
}

func (s *VirtualMachineSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *VirtualMachineSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *VirtualMachineSensor) Kind() string {
	return "VirtualMachineSensor"
}

func (s *VirtualMachineSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *VirtualMachineSensor) Match(name string) bool {
	return helper.NewMatcher("vm", "virtual_machine", "virtualmachine").Match(name)
}

func (s *VirtualMachineSensor) Enabled() bool {
	return true
}

func (s *VirtualMachineSensor) GetLatestMetrics() []sensormetrics.SensorMetric {
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

func ConvertToVirtualMachine(ctx context.Context, scraper *VCenterScraper, vm mo.VirtualMachine, t time.Time) objects.VirtualMachine {
	self := objects.NewManagedObjectReferenceFromVMwareRef(vm.Self)

	var parent *objects.ManagedObjectReference
	if vm.Parent != nil {
		p := objects.NewManagedObjectReferenceFromVMwareRef(*vm.Parent)
		parent = &p
	}

	virtualMachine := objects.VirtualMachine{
		Timestamp: t,
		Name:      vm.Name,
		Self:      self,
		Parent:    parent,
	}

	if virtualMachine.Parent != nil {
		parentChain := scraper.DB.GetParentChain(ctx, *virtualMachine.Parent)
		virtualMachine.Datacenter = parentChain.DC
		// virtualMachine.Cluster = parentChain.Cluster
	}
	mb := int64(1024 * 1024)
	summary := vm.Summary
	virtualMachine.OverallStatus = string(summary.OverallStatus)

	if config := vm.Config; config != nil {

		virtualMachine.UUID = config.Uuid
		virtualMachine.GuestID = config.GuestId
		virtualMachine.Template = config.Template
		timeConfChanged, err := time.Parse(time.RFC3339Nano, config.ChangeVersion)
		if err == nil {
			virtualMachine.TimeConfigChanged = timeConfChanged
		}

		if timeCreated := config.CreateDate; timeCreated != nil {
			virtualMachine.TimeConfigChanged = *timeCreated
		}

		if config.CpuAllocation != nil {
			if config.CpuAllocation.Shares != nil {
				virtualMachine.CPUAllocationShares = float64(vm.Config.CpuAllocation.Shares.Shares)
			}
			if config.CpuAllocation.Reservation != nil {
				virtualMachine.CPUAllocationShares = float64(*vm.Config.CpuAllocation.Reservation)
			}
		}
		if config.MemoryAllocation != nil {
			if config.MemoryAllocation.Shares != nil {
				virtualMachine.MemoryAllocationShares = float64(int64(vm.Config.MemoryAllocation.Shares.Shares) * mb)
			}
			if config.MemoryAllocation.Reservation != nil {
				virtualMachine.MemoryAllocationReservation = float64(int64(*vm.Config.MemoryAllocation.Reservation) * mb)
			}
		}

		if tools := config.Tools; tools != nil {
			virtualMachine.GuestToolsVersion = strconv.Itoa(int(tools.ToolsVersion))
		}

		hardware := config.Hardware
		virtualMachine.MemoryBytes = float64(int64(hardware.MemoryMB) * mb)
		virtualMachine.NumCPU = float64(hardware.NumCPU)
		virtualMachine.NumCoresPerSocket = float64(hardware.NumCoresPerSocket)
	}

	runtime := summary.Runtime
	virtualMachine.MaxCPUUsage = float64(runtime.MaxCpuUsage)
	virtualMachine.PowerState = string(runtime.PowerState)
	if hostRef := runtime.Host; hostRef != nil {
		oRef := objects.NewManagedObjectReferenceFromVMwareRef(*hostRef)
		if host := scraper.DB.GetHost(ctx, oRef); host != nil {
			virtualMachine.HostInfo = objects.VirtualMachineHostInfo{
				Host:       host.Name,
				Datacenter: host.Datacenter,
				Cluster:    host.Cluster,
			}
		}
	}

	qs := summary.QuickStats
	virtualMachine.OverallCPUUsage = float64(qs.OverallCpuUsage)
	virtualMachine.OverallCPUDemand = float64(qs.OverallCpuDemand)
	virtualMachine.GuestMemoryUsage = float64(int64(qs.GuestMemoryUsage) * mb)
	virtualMachine.HostMemoryUsage = float64(int64(qs.HostMemoryUsage) * mb)
	virtualMachine.UptimeSeconds = float64(qs.UptimeSeconds)
	virtualMachine.GuestHeartbeatStatus = string(qs.GuestHeartbeatStatus)
	virtualMachine.DistributedCPUEntitlement = float64(qs.DistributedCpuEntitlement)
	virtualMachine.DistributedMemoryEntitlement = float64(int64(qs.DistributedMemoryEntitlement) * mb)
	virtualMachine.StaticCPUEntitlement = float64(qs.StaticCpuEntitlement)
	virtualMachine.StaticMemoryEntitlement = float64(int64(qs.StaticMemoryEntitlement) * mb)
	virtualMachine.PrivateMemory = float64(int64(qs.PrivateMemory) * mb)
	virtualMachine.SharedMemory = float64(int64(qs.SharedMemory) * mb)
	virtualMachine.SwappedMemory = float64(int64(qs.SwappedMemory) * mb)
	virtualMachine.BalloonedMemory = float64(int64(qs.BalloonedMemory) * mb)
	virtualMachine.ConsumedOverheadMemory = float64(int64(qs.ConsumedOverheadMemory) * mb)
	virtualMachine.FtLogBandwidth = float64(qs.FtLogBandwidth)
	virtualMachine.FtSecondaryLatency = float64(qs.FtSecondaryLatency)
	virtualMachine.CompressedMemory = float64(int64(qs.CompressedMemory) * mb)
	virtualMachine.SsdSwappedMemory = float64(int64(qs.SsdSwappedMemory) * mb)

	if guest := summary.Guest; guest != nil {
		virtualMachine.GuestToolsStatus = string(guest.ToolsStatus)
	}

	if guest := vm.Guest; guest != nil {
		guestNets := []objects.VirtualMachineGuestNet{}
		for _, net := range guest.Net {
			if ipConfig := net.IpConfig; ipConfig != nil {
				for _, address := range ipConfig.IpAddress {
					if match := regexPatternIPv4.MatchString(address.IpAddress); match {
						virtualMachine.GuestNetwork = append(virtualMachine.GuestNetwork, objects.VirtualMachineGuestNet{
							MacAddress: net.MacAddress,
							IpAddress:  address.IpAddress,
							Connected:  net.Connected,
						})
					}
				}
			}
		}
		virtualMachine.GuestNetwork = helper.DedupFunc(guestNets, func(i1, i2 objects.VirtualMachineGuestNet) bool {
			return i1.IpAddress != i2.IpAddress && i1.MacAddress != i2.MacAddress
		})
	}

	virtualMachine.Disk = ExtractDisksFromVM(vm)

	if snapshots := vm.Snapshot; snapshots != nil {
		virtualMachine.Snapshot = append(virtualMachine.Snapshot, walkSnapshotTree(snapshots.RootSnapshotList)...)
	}

	if rPool := vm.ResourcePool; rPool != nil {
		rPoolRef := objects.NewManagedObjectReferenceFromVMwareRef(*rPool)

		if rp := scraper.DB.GetResourcePool(ctx, rPoolRef); rp != nil {
			virtualMachine.ResourcePool = rp.Name
		}
	}

	virtualMachine.IsecAnnotation = GetIsecAnnotation(vm)

	return virtualMachine
}

func walkSnapshotTree(snaps []types.VirtualMachineSnapshotTree) []objects.VirtualMachineSnapshot {
	result := []objects.VirtualMachineSnapshot{}
	for _, snap := range snaps {
		result = append(result, objects.VirtualMachineSnapshot{
			Name:         snap.Name,
			CreationTime: snap.CreateTime,
		})
		result = append(result, walkSnapshotTree(snap.ChildSnapshotList)...)
	}
	return result
}

func ExtractDisksFromVM(vm mo.VirtualMachine) []objects.VirtualMachineDisk {
	result := []objects.VirtualMachineDisk{}
	if vm.Config != nil {
		disks := object.VirtualDeviceList(vm.Config.Hardware.Device).SelectByType((*types.VirtualDisk)(nil))
		for _, d := range disks {
			disk := d.(*types.VirtualDisk)
			info := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo)
			result = append(result, objects.VirtualMachineDisk{
				UUID:            info.Uuid,
				VMDKFile:        info.FileName,
				Capacity:        float64(disk.CapacityInBytes),
				ThinProvisioned: *info.ThinProvisioned,
			})
		}
	}
	return result
}

func GetIsecAnnotation(vm mo.VirtualMachine) *objects.IsecAnnotation {
	tmp := objects.IsecAnnotation{
		Service:     "not defined",
		Responsable: "not defined",
		Criticality: "not defined",
	}
	if config := vm.Config; config != nil {
		_ = json.Unmarshal([]byte(config.Annotation), &tmp)
	}
	return &tmp
}
