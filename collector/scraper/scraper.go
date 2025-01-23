package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/intrinsec/govc_exporter/collector/pool"
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
	cleanupInterval := time.Duration(conf.CleanIntervalSec) * time.Second

	if conf.ClusterCollectorEnabled {
		scraper.Cluster = NewClusterSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.Cluster, time.Duration(conf.ClusterRefreshIntervalSec)*time.Second))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.Cluster, cleanupInterval, time.Duration(conf.ClusterMaxAgeSec)))
	}

	if true {
		scraper.ComputeResources = NewComputeResourceSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.ComputeResources, time.Duration(conf.ClusterRefreshIntervalSec)*time.Second))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.ComputeResources, cleanupInterval, time.Duration(conf.ClusterMaxAgeSec)*time.Second))
	}

	if conf.DatastoreCollectorEnabled {
		scraper.Datastore = NewDatastoreSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.Datastore, time.Duration(conf.DatastoreRefreshIntervalSec)*time.Second))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.Datastore, cleanupInterval, time.Duration(conf.DatastoreMaxAgeSec)*time.Second))
	}

	if true {
		scraper.Host = NewHostSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.Host, time.Duration(conf.HostRefreshIntervalSec)*time.Second))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.Host, cleanupInterval, time.Duration(conf.HostMaxAgeSec)*time.Second))
	}

	if conf.ResourcePoolCollectorEnabled {
		scraper.ResourcePool = NewResourcePoolSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.ResourcePool, time.Duration(conf.ResourcePoolRefreshIntervalSec)*time.Second))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.ResourcePool, cleanupInterval, time.Duration(conf.ResourcePoolMaxAgeSec)*time.Second))
	}

	if conf.SpodCollectorEnabled {
		scraper.SPOD = NewStoragePodSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.SPOD, time.Duration(conf.SpodRefreshIntervalSec)*time.Second))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.SPOD, cleanupInterval, time.Duration(conf.SpodMaxAgeSec)*time.Second))
	}

	if conf.VirtualMachineCollectorEnabled {
		scraper.VM = NewVirtualMachineSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.VM, time.Duration(conf.VirtualMachineRefreshIntervalSec)*time.Second))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.VM, cleanupInterval, time.Duration(conf.VirtualMachineMaxAgeSec)*time.Second))
	}

	if conf.SpodCollectorEnabled {
		scraper.SPOD = NewStoragePodSensor(&scraper)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.SPOD, time.Duration(conf.SpodRefreshIntervalSec)*time.Second))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.SPOD, cleanupInterval, time.Duration(conf.SpodMaxAgeSec)*time.Second))
	}

	if conf.TagsCollectorEnbled {
		logger.Info("Create TagsSensor", "TagsCategoryToCollect", conf.TagsCategoryToCollect)
		scraper.Tags = NewTagsSensorWithTaglist(&scraper, conf.TagsCategoryToCollect)
		scraper.refreshers = append(scraper.refreshers, NewAutoRefresh(scraper.Tags, time.Duration(conf.TagsRefreshIntervalSec)*time.Second))
		scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.Tags, cleanupInterval, time.Duration(conf.TagsMaxAgeSec)*time.Second))
	}

	scraper.Remain = NewOnDemandSensor(&scraper, logger)
	scraper.cleaners = append(scraper.cleaners, NewAutoClean(scraper.Remain, cleanupInterval, time.Duration(conf.OnDemandCacheMaxAge)*time.Second))

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
		"cluster":          c.config.ClusterCollectorEnabled,
		"compute_resource": c.config.ClusterCollectorEnabled,
		"datastore":        c.config.DatastoreCollectorEnabled,
		"vm":               c.config.VirtualMachineCollectorEnabled,
		"spod":             c.config.SpodCollectorEnabled,
		"repool":           c.config.ResourcePoolCollectorEnabled,
		"tags":             c.config.TagsCollectorEnbled,
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
	DC      string
	Cluster string
	SPOD    string
	Chain   []string
}

func (c *VCenterScraper) GetParentChain(ref types.ManagedObjectReference) ParentChain {
	return c.walkParentChain(&ref, nil)
}

func (c *VCenterScraper) walkParentChain(ref *types.ManagedObjectReference, chain *ParentChain) ParentChain {

	if chain == nil {
		chain = &ParentChain{}
	}

	if ref == nil {
		return *chain
	} else if ref.Type == "StoragePod" && c.SPOD != nil {
		entity := c.SPOD.Get(*ref)
		if entity != nil {
			chain.SPOD = entity.Name
			chain.Chain = append(chain.Chain, fmt.Sprintf("%s:%s", ref.Type, ref.Value))
			return c.walkParentChain(entity.Parent, chain)
		}
	} else if ref.Type == "HostSystem" && c.Host != nil {
		entity := c.Host.Get(*ref)
		if entity != nil {
			chain.Chain = append(chain.Chain, fmt.Sprintf("%s:%s", ref.Type, ref.Value))
			return c.walkParentChain(entity.Parent, chain)
		}
	} else if ref.Type == "ClusterComputeResource" && c.Cluster != nil {
		entity := c.Cluster.Get(*ref)
		if entity != nil {
			chain.Cluster = entity.Name
			chain.Chain = append(chain.Chain, fmt.Sprintf("%s:%s", ref.Type, ref.Value))
			return c.walkParentChain(entity.Parent, chain)
		}
	} else if ref.Type == "ComputeResource" && c.ComputeResources != nil {
		entity := c.ComputeResources.Get(*ref)
		if entity != nil {
			chain.Chain = append(chain.Chain, fmt.Sprintf("%s:%s", ref.Type, ref.Value))
			return c.walkParentChain(entity.Parent, chain)
		}
	} else if ref.Type == "VirtualMachine" && c.VM != nil {
		entity := c.VM.Get(*ref)
		if entity != nil {
			chain.Chain = append(chain.Chain, fmt.Sprintf("%s:%s", ref.Type, ref.Value))
			return c.walkParentChain(entity.Parent, chain)
		}
	} else if ref.Type == "Datastore" && c.Datastore != nil {
		entity := c.Datastore.Get(*ref)
		if entity != nil && c.SPOD != nil {
			chain.SPOD = func() string {
				for _, spod := range c.SPOD.GetAll() {
					for _, child := range spod.ChildEntity {
						if child == *ref {
							return spod.Name
						}
					}
				}
				return "NONE"
			}()

			chain.Chain = append(chain.Chain, fmt.Sprintf("%s:%s", ref.Type, ref.Value))
			return c.walkParentChain(entity.Parent, chain)
		}
	} else {
		entity := c.Remain.Get(*ref)
		if entity != nil {
			if entity.Self.Type == "Datacenter" {
				chain.DC = entity.Name
			} else if entity.Self.Type == "ClusterComputeResource" {
				chain.Cluster = entity.Name
			} else if entity.Self.Type == "StoragePod" {
				chain.SPOD = entity.Name
			}
			chain.Chain = append(chain.Chain, fmt.Sprintf("%s:%s", ref.Type, ref.Value))
			return c.walkParentChain(entity.Parent, chain)
		}
	}
	return *chain
}
