package scraper

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sanderdescamps/govc_exporter/collector/pool"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ScraperStatus struct {
	TCPStatusCheck         bool
	TCPStatusCheckMgs      string
	TCPStatusCheckEndpoint string

	SensorEnabled   map[string]bool
	SensorAvailable map[string]bool
	SensorMetric    []SensorMetric
}

type VCenterScraper struct {
	clientPool *pool.VCenterClientPool
	config     ScraperConfig
	metrics    map[string]*SensorMetric
	refreshers []*AutoRefresh
	cleaners   []*AutoClean

	Host             *HostSensor
	Cluster          *ClusterSensor
	ComputeResources *ComputeResourceSensor
	VM               *VirtualMachineSensor
	Datastore        *DatastoreSensor
	SPOD             *StoragePodSensor
	ResourcePool     *ResourcePoolSensor
	Tags             *TagsSensor
	Remain           *OnDemandSensor
}

func NewVCenterScraper(conf ScraperConfig, logger *slog.Logger) (*VCenterScraper, error) {

	ctx := context.Background()
	pool, err := pool.NewVCenterClientPoolWithLogger(
		ctx,
		conf.Endpoint,
		conf.Username,
		conf.Password,
		conf.ClientPoolSize,
		logger,
	)
	if err != nil {
		return nil, err
	}

	scraper := VCenterScraper{
		clientPool: pool,
		config:     conf,
		metrics:    make(map[string]*SensorMetric),
		refreshers: []*AutoRefresh{},
		cleaners:   []*AutoClean{},
	}

	if conf.ClusterScraperEnabled {
		scraper.Cluster = NewClusterSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.Cluster, conf.ClusterRefreshInterval))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.Cluster, conf.CleanInterval, conf.ClusterMaxAge))
	}

	if conf.ComputeResourceScraperEnabled {
		scraper.ComputeResources = NewComputeResourceSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.ComputeResources, conf.ComputeResourceRefreshInterval))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.ComputeResources, conf.CleanInterval, conf.ComputeResourceMaxAge))
	}

	if conf.DatastoreScraperEnabled {
		scraper.Datastore = NewDatastoreSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.Datastore, conf.DatastoreRefreshInterval))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.Datastore, conf.CleanInterval, conf.DatastoreMaxAge))
	}

	if conf.HostScraperEnabled {
		scraper.Host = NewHostSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.Host, conf.HostRefreshInterval))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.Host, conf.CleanInterval, conf.HostMaxAge))
	}

	if conf.ResourcePoolScraperEnabled {
		scraper.ResourcePool = NewResourcePoolSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.ResourcePool, conf.ResourcePoolRefreshInterval))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.ResourcePool, conf.CleanInterval, conf.ResourcePoolMaxAge))
	}

	if conf.SpodScraperEnabled {
		scraper.SPOD = NewStoragePodSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.SPOD, conf.SpodRefreshInterval))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.SPOD, conf.CleanInterval, conf.SpodMaxAge))
	}

	if conf.VirtualMachineScraperEnabled {
		scraper.VM = NewVirtualMachineSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.VM, conf.VirtualMachineRefreshInterval))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.VM, conf.CleanInterval, conf.VirtualMachineMaxAge))
	}

	if conf.SpodScraperEnabled {
		scraper.SPOD = NewStoragePodSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.SPOD, conf.SpodRefreshInterval))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.SPOD, conf.CleanInterval, conf.SpodMaxAge))
	}

	if conf.TagsScraperEnabled {
		logger.Info("Create TagsSensor", "TagsCategoryToCollect", conf.TagsCategoryToCollect)
		scraper.Tags = NewTagsSensorWithTaglist(&scraper, conf.TagsCategoryToCollect)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.Tags, conf.TagsRefreshInterval))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.Tags, conf.CleanInterval, conf.TagsMaxAge))
	}

	scraper.Remain = NewOnDemandSensor(&scraper, logger)
	scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.Remain, conf.CleanInterval, conf.OnDemandCacheMaxAge))

	return &scraper, nil
}

// func (c *VCenterScraper) dispatchMetrics(m SensorMetric) {
// 	if m.Name == "on_demand" {
// 		if current, ok := c.metrics[m.Name]; ok {
// 			newMean := current.RollingMean(m, 10)
// 			c.metricsLock.Lock()
// 			defer c.metricsLock.Unlock()
// 			c.metrics[m.Name] = &newMean
// 			return
// 		}
// 	}
// 	c.metricsLock.Lock()
// 	defer c.metricsLock.Unlock()
// 	c.metrics[m.Name] = &m
// }

func (c *VCenterScraper) Status() ScraperStatus {
	tcpConnect := false
	tcpErrMsg := ""
	tcpEndpoint := ""
	baseURL, err := c.config.URL()
	if err == nil {
		tcpEndpoint = baseURL.Host
		tcpConnect, err = tcpConnectionCheck(tcpEndpoint)

		if err != nil {
			tcpErrMsg = err.Error()
		}
	} else {
		tcpConnect = false
		tcpErrMsg = err.Error()
	}

	sensorEnabled := map[string]bool{
		"host":             true,
		"cluster":          c.config.ClusterScraperEnabled,
		"compute_resource": c.config.ClusterScraperEnabled,
		"datastore":        c.config.DatastoreScraperEnabled,
		"vm":               c.config.VirtualMachineScraperEnabled,
		"spod":             c.config.SpodScraperEnabled,
		"repool":           c.config.ResourcePoolScraperEnabled,
		"tags":             c.config.TagsScraperEnabled,
	}

	sensorAvailable := map[string]bool{
		"host":             c.Host != nil,
		"cluster":          c.Cluster != nil,
		"compute_resource": c.ComputeResources != nil,
		"datastore":        c.Datastore != nil,
		"vm":               c.VM != nil,
		"spod":             c.SPOD != nil,
		"repool":           c.ResourcePool != nil,
		"tags":             c.Tags != nil,
	}

	sensors := []HasMetrics{
		c.Host,
		c.Cluster,
		c.ComputeResources,
		c.VM,
		c.Datastore,
		c.SPOD,
		c.ResourcePool,
		c.Tags,
		c.Remain,
	}

	sensorMetrics := []SensorMetric{}
	for _, s := range sensors {
		metrics := s.GetMetrics()
		if metrics != nil {
			sensorMetrics = append(sensorMetrics, *metrics)
		}

	}

	return ScraperStatus{
		TCPStatusCheck:         tcpConnect,
		TCPStatusCheckMgs:      tcpErrMsg,
		TCPStatusCheckEndpoint: tcpEndpoint,
		SensorEnabled:          sensorEnabled,
		SensorAvailable:        sensorAvailable,
		SensorMetric:           sensorMetrics,
	}
}

func (c *VCenterScraper) Start(logger *slog.Logger) error {
	ctx := context.Background()

	for _, r := range c.refreshers {
		r.Start(ctx, logger)
	}

	for _, c := range c.cleaners {
		c.Start(logger)
	}

	return nil
}

func (c *VCenterScraper) Stop(logger *slog.Logger) {
	logger.Info("stopping all refreshers...")
	for _, r := range c.refreshers {
		r.Stop(logger)
	}
	logger.Info("refreshers all stopped")

	logger.Info("stopping all cleaners...")
	for _, c := range c.cleaners {
		c.Stop(logger)
	}
	logger.Info("cleaners all stopped")

	c.clientPool.Destroy()
	logger.Info("Close client pool")
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
