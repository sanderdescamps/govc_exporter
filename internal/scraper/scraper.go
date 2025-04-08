package scraper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"reflect"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/pool"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type VCenterScraper struct {
	clientPool pool.VCenterPool
	config     Config
	metrics    []*SensorMetric
	runnables  []Runnable
	// logger     *slog.Logger

	Host             *HostSensor
	HostPerf         *HostPerfSensor
	Cluster          *ClusterSensor
	ComputeResources *ComputeResourceSensor
	VM               *VirtualMachineSensor
	VMPerf           *VMPerfSensor
	Datastore        *DatastoreSensor
	SPOD             *StoragePodSensor
	ResourcePool     *ResourcePoolSensor
	Tags             *TagsSensor
	Remain           *OnDemandSensor
}

func NewVCenterScraper(ctx context.Context, conf Config) (*VCenterScraper, error) {

	pool := pool.NewVCenterThrottlePool(
		conf.Endpoint,
		conf.Username,
		conf.Password,
		conf.ClientPoolSize,
	)

	err := pool.Init()
	if err != nil {
		return nil, err
	}
	// pool.StartAuthRefresher(ctx, 5*time.Minute)

	scraper := VCenterScraper{
		clientPool: pool,
		config:     conf,
		metrics:    []*SensorMetric{},
	}

	if conf.Cluster.Enabled {
		scraper.Cluster = NewClusterSensor(&scraper, conf.Cluster)
		scraper.runnables = append(scraper.runnables, scraper.Cluster)
	}

	if conf.ComputeResource.Enabled {
		scraper.ComputeResources = NewComputeResourceSensor(&scraper, conf.ComputeResource)
		scraper.runnables = append(scraper.runnables, scraper.ComputeResources)
	}

	if conf.Datastore.Enabled {
		scraper.Datastore = NewDatastoreSensor(&scraper, conf.Datastore)
		scraper.runnables = append(scraper.runnables, scraper.Datastore)
	}

	if conf.Host.Enabled {
		scraper.Host = NewHostSensor(&scraper, conf.Host)
		scraper.runnables = append(scraper.runnables, scraper.Host)
		if conf.HostPerf.Enabled {
			scraper.HostPerf = NewHostPerfSensor(&scraper, conf.HostPerf)
			scraper.runnables = append(scraper.runnables, scraper.HostPerf)
		}
	}

	if conf.ResourcePool.Enabled {
		scraper.ResourcePool = NewResourcePoolSensor(&scraper, conf.ResourcePool)
		scraper.runnables = append(scraper.runnables, scraper.ResourcePool)
	}

	if conf.Spod.Enabled {
		scraper.SPOD = NewStoragePodSensor(&scraper, conf.Spod)
		scraper.runnables = append(scraper.runnables, scraper.SPOD)
	}

	if conf.VirtualMachine.Enabled {
		scraper.VM = NewVirtualMachineSensor(&scraper, conf.VirtualMachine)
		scraper.runnables = append(scraper.runnables, scraper.VM)
		if conf.VirtualMachinePerf.Enabled {
			scraper.VMPerf = NewVMPerfSensor(&scraper, conf.VirtualMachinePerf)
			scraper.runnables = append(scraper.runnables, scraper.VMPerf)
		}
	}

	if conf.Tags.Enabled {
		if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
			logger.Info("Create TagsSensor", "TagsCategoryToCollect", conf.Tags.CategoryToCollect)
		}
		scraper.Tags = NewTagsSensorWithTaglist(&scraper, conf.Tags)
		scraper.runnables = append(scraper.runnables, scraper.Tags)
	}

	scraper.Remain = NewOnDemandSensor(&scraper, SensorConfig{
		Enabled:       true,
		CleanInterval: conf.OnDemand.CleanInterval,
		MaxAge:        conf.OnDemand.MaxAge,
	})
	scraper.runnables = append(scraper.runnables, scraper.Remain)

	return &scraper, nil
}

var (
	ErrVCenterURLInvalid  = errors.New("could not parse url")
	ErrVCenterConnectFail = errors.New("cannot connect to vcenter")
)

func (c *VCenterScraper) tcpConnectStatus() error {
	baseURL, err := c.config.URL()
	if err != nil {
		return ErrVCenterURLInvalid
	}

	_, err = tcpConnectionCheck(baseURL.Host)
	if err != nil {
		return ErrVCenterConnectFail
	}

	return nil
}

func (c *VCenterScraper) ScraperMetrics() []BaseSensorMetric {
	result := []BaseSensorMetric{}

	err := c.tcpConnectStatus()
	tcpConnectStatus := *NewSensorMetricStatus("scraper", "tcp_connect_status", false)
	if err == nil {
		tcpConnectStatus.Success()
	} else if err != nil && errors.Is(err, ErrVCenterURLInvalid) {
		tcpConnectStatus.Fail()
	} else if err != nil && errors.Is(err, ErrVCenterConnectFail) {
		tcpConnectStatus.Fail()
	} else {
		tcpConnectStatus.Fail()
	}

	for kind, enabled := range map[string]bool{
		"host":             c.config.Host.Enabled,
		"cluster":          c.config.Cluster.Enabled,
		"compute_resource": c.config.Cluster.Enabled,
		"datastore":        c.config.Datastore.Enabled,
		"vm":               c.config.VirtualMachine.Enabled,
		"spod":             c.config.Spod.Enabled,
		"repool":           c.config.ResourcePool.Enabled,
		"tags":             c.config.Tags.Enabled,
	} {
		sensor := NewSensorMetricStatus("scraper", fmt.Sprintf("sensor.%s.enabled", kind), false)
		sensor.Update(enabled)
		result = append(result, sensor.BaseSensorMetric)
	}

	for kind, available := range map[string]bool{
		"host":             c.Host != nil,
		"cluster":          c.Cluster != nil,
		"compute_resource": c.ComputeResources != nil,
		"datastore":        c.Datastore != nil,
		"vm":               c.VM != nil,
		"spod":             c.SPOD != nil,
		"repool":           c.ResourcePool != nil,
		"tags":             c.Tags != nil,
	} {
		sensor := NewSensorMetricStatus("scraper", fmt.Sprintf("sensor.%s.available", kind), false)
		sensor.Update(available)
		result = append(result, sensor.BaseSensorMetric)
	}

	for _, metric := range c.metrics {
		result = append(result, metric.BaseSensorMetric)
	}
	return result
}

func (c *VCenterScraper) RegisterSensorMetric(metric ...*SensorMetric) {
	c.metrics = append(c.metrics, metric...)
}

func (c *VCenterScraper) SensorList() []Sensor {
	sensors := []Sensor{}
	for _, s := range []Sensor{
		c.Cluster,
		c.ComputeResources,
		c.Datastore,
		c.Host,
		c.HostPerf,
		c.ResourcePool,
		c.SPOD,
		c.Tags,
		c.VM,
		c.VMPerf,
		c.Remain,
	} {
		if !reflect.ValueOf(s).IsNil() {
			sensors = append(sensors, s)
		}
	}

	return sensors
}

func (c *VCenterScraper) GetSensorRefreshByName(name string) Sensor {
	for _, s := range c.SensorList() {
		if s.Match(name) {
			return s
		}
	}

	return nil
}

func (c *VCenterScraper) RefreshSensor(ctx context.Context, names ...string) error {
	for _, s := range c.SensorList() {
		if helper.AnyMatch(s, names...) {
			err := s.TriggerRefresh(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *VCenterScraper) Start(ctx context.Context) error {
	err := c.tcpConnectStatus()
	if err != nil {
		return fmt.Errorf("cannot connect to vcenter: %w", err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	sensors := c.SensorList()
	startupDone := make(chan Sensor, len(sensors))
	// Start all sensors
	for _, r := range sensors {
		go func() {
			time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
			err := r.Start(ctx)
			if err != nil {
				if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
					logger.Error("failed to start sensor", "err", err)
				}
				return
			}
			r.WaitTillStartup()
			startupDone <- r
		}()
	}

	sensorRunning := []string{}
	for {
		select {
		case sensor := <-startupDone:
			sensorRunning = append(sensorRunning, sensor.Name())
			if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
				msg := fmt.Sprintf("Sensor started [%d/%d]", len(sensorRunning), len(sensors))
				logger.Info(msg, "sensor_name", sensor.Name(), "sensor_kind", sensor.Kind())
			}
			if len(sensorRunning) >= len(sensors) {
				if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
					logger.Info("All sensors started")
				}
				return nil
			}
		case <-ctxWithTimeout.Done():
			allSensors := []string{}
			for _, r := range sensors {
				allSensors = append(allSensors, r.Name())
			}

			sensorsNotStarted := helper.Subtract(allSensors, sensorRunning)

			if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
				logger.Error("Timeout waiting for sensors to start", "sensors_not_started", sensorsNotStarted)
			}
			return fmt.Errorf("timeout waiting for sensors to start: %v", sensorsNotStarted)
		}
	}
}

func (c *VCenterScraper) Stop(ctx context.Context) {

	// for _, r := range c.runnables {
	// 	// r.Stop(ctx)
	// }

	c.clientPool.Destroy()
	if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
		logger.Info("Close client pool")
	}
}

type ParentChain struct {
	DC           string
	Cluster      string
	ResourcePool string
	SPOD         string
	Chain        []string
}

func (c *VCenterScraper) GetParentChain(ref types.ManagedObjectReference) ParentChain {
	return c.walkParentChain(&ref, nil)
}

func (c *VCenterScraper) GetManagedEntity(ref *types.ManagedObjectReference) *mo.ManagedEntity {
	if ref == nil {
		return nil
	} else if c.Cluster != nil && ref.Type == string(types.ManagedObjectTypesClusterComputeResource) {
		return &(c.Cluster.Get(*ref).ManagedEntity)
	} else if c.ComputeResources != nil && ref.Type == string(types.ManagedObjectTypesComputeResource) {
		return &(c.ComputeResources.Get(*ref).ManagedEntity)
	} else if c.Datastore != nil && ref.Type == string(types.ManagedObjectTypesDatastore) {
		return &(c.Datastore.Get(*ref).ManagedEntity)
	} else if c.Host != nil && ref.Type == string(types.ManagedObjectTypesHostSystem) {
		return &(c.Host.Get(*ref).ManagedEntity)
	} else if c.SPOD != nil && ref.Type == string(types.ManagedObjectTypesStoragePod) {
		return &(c.SPOD.Get(*ref).ManagedEntity)
	} else if c.ResourcePool != nil && ref.Type == string(types.ManagedObjectTypesResourcePool) {
		return &(c.ResourcePool.Get(*ref).ManagedEntity)
	} else if c.VM != nil && ref.Type == string(types.ManagedObjectTypesVirtualMachine) {
		return &(c.VM.Get(*ref).ManagedEntity)
	} else {
		return c.Remain.Get(*ref)
	}
}

func (c *VCenterScraper) walkParentChain(ref *types.ManagedObjectReference, chain *ParentChain) ParentChain {

	if chain == nil {
		chain = &ParentChain{
			DC:           "",
			Cluster:      "",
			SPOD:         "",
			ResourcePool: "",
			Chain:        []string{},
		}
	}

	if ref == nil {
		return *chain
	}

	chain.Chain = append(chain.Chain, fmt.Sprintf("%s:%s", ref.Type, ref.Value))

	if c.Cluster != nil && ref.Type == string(types.ManagedObjectTypesClusterComputeResource) {
		entity := c.Cluster.Get(*ref)
		if entity != nil {
			chain.Cluster = entity.Name
			return c.walkParentChain(entity.Parent, chain)
		}
	} else if c.ComputeResources != nil && ref.Type == string(types.ManagedObjectTypesComputeResource) {
		entity := c.ComputeResources.Get(*ref)
		if entity != nil {
			return c.walkParentChain(entity.Parent, chain)
		}
	} else if c.SPOD != nil && ref.Type == string(types.ManagedObjectTypesStoragePod) {
		entity := c.SPOD.Get(*ref)
		if entity != nil {
			chain.SPOD = entity.Name
			return c.walkParentChain(entity.Parent, chain)
		}
	} else if c.ResourcePool != nil && ref.Type == string(types.ManagedObjectTypesResourcePool) {
		entity := c.ResourcePool.Get(*ref)
		if entity != nil {
			chain.ResourcePool = entity.Name
			return c.walkParentChain(entity.Parent, chain)
		}
	} else if c.Host != nil && ref.Type == string(types.ManagedObjectTypesHostSystem) {
		entity := c.Host.Get(*ref)
		if entity != nil {
			return c.walkParentChain(entity.Parent, chain)
		}
	} else if c.VM != nil && ref.Type == string(types.ManagedObjectTypesVirtualMachine) {
		entity := c.VM.Get(*ref)
		if entity != nil {
			if entity.ResourcePool != nil {
				return c.walkParentChain(entity.ResourcePool, chain)
			} else {
				return c.walkParentChain(entity.Parent, chain)
			}

		}
	} else if c.Datastore != nil && ref.Type == string(types.ManagedObjectTypesDatastore) {
		entity := c.Datastore.Get(*ref)
		if entity != nil {
			return c.walkParentChain(entity.Parent, chain)
		}
	} else {
		entity := c.Remain.Get(*ref)
		if entity != nil {
			if entity.Self.Type == string(types.ManagedObjectTypesDatacenter) {
				chain.DC = entity.Name
			} else if entity.Self.Type == string(types.ManagedObjectTypesClusterComputeResource) {
				chain.Cluster = entity.Name
			} else if entity.Self.Type == string(types.ManagedObjectTypesResourcePool) {
				chain.ResourcePool = entity.Name
			} else if entity.Self.Type == string(types.ManagedObjectTypesStoragePod) {
				chain.SPOD = entity.Name
			}
			return c.walkParentChain(entity.Parent, chain)
		}
	}
	return *chain
}
