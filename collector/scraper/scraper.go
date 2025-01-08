package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/intrinsec/govc_exporter/collector/pool"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ScraperStatus struct {
	TCPStatusCheck         bool
	TCPStatusCheckMgs      string
	TCPStatusCheckEndpoint string

	SensorEnabled   map[string]bool
	SensorAvailable map[string]bool
	SensorMetric    map[string]SensorMetrics
}

type VCenterScraper struct {
	clientPool       pool.Pool[govmomi.Client]
	config           ScraperConfig
	Host             *AutoRefreshSensor[mo.HostSystem]
	Cluster          *AutoRefreshSensor[mo.ComputeResource]
	ComputeResources *AutoRefreshSensor[mo.ComputeResource]
	VM               *AutoRefreshSensor[mo.VirtualMachine]
	Datastore        *AutoRefreshSensor[mo.Datastore]
	SPOD             *AutoRefreshSensor[mo.StoragePod]
	ResourcePool     *AutoRefreshSensor[mo.ResourcePool]
	Remain           *OnDemandSensor
}

func NewVCenterScraper(conf ScraperConfig) (*VCenterScraper, error) {
	cache := VCenterScraper{
		// clientPool: pool,
		config: conf,
	}

	return &cache, nil
}

func (c *VCenterScraper) Status() ScraperStatus {

	tcpConnect := false
	var tcpErrMsg string
	baeURL, err := c.config.URL()
	if err == nil {
		tcpConnect, err = tcpConnectionCheck(baeURL.Host)
		if err != nil {
			tcpErrMsg = err.Error()
		}
	} else {
		tcpConnect = false
		tcpErrMsg = err.Error()
	}

	sensorEnabled := make(map[string]bool)
	sensorAvailable := make(map[string]bool)
	sensorMetrics := map[string]SensorMetrics{}

	sensorAvailable["host"] = (c.Host != nil)
	sensorEnabled["host"] = true
	if c.Host != nil {
		sensorMetrics["host"] = c.Host.GetMetrics()
	}

	sensorAvailable["cluster"] = (c.Cluster != nil)
	sensorEnabled["cluster"] = c.config.ClusterCollectorEnabled
	if c.Cluster != nil {
		sensorMetrics["cluster"] = c.Cluster.GetMetrics()
	}

	sensorAvailable["datastore"] = (c.Datastore != nil)
	sensorEnabled["datastore"] = c.config.DatastoreCollectorEnabled
	if c.Datastore != nil {
		sensorMetrics["datastore"] = c.Datastore.GetMetrics()
	}

	sensorAvailable["vm"] = (c.VM != nil)
	sensorEnabled["vm"] = c.config.VirtualMachineCollectorEnabled
	if c.VM != nil {
		sensorMetrics["vm"] = c.VM.GetMetrics()
	}

	sensorAvailable["spod"] = (c.SPOD != nil)
	sensorEnabled["spod"] = c.config.SpodCollectorEnabled
	if c.SPOD != nil {
		sensorMetrics["spod"] = c.SPOD.GetMetrics()
	}

	sensorAvailable["repool"] = (c.ResourcePool != nil)
	sensorEnabled["repool"] = c.config.ResourcePoolCollectorEnabled
	if c.ResourcePool != nil {
		sensorMetrics["repool"] = c.ResourcePool.GetMetrics()
	}

	sensorAvailable["compute_resource"] = (c.ComputeResources != nil)
	sensorEnabled["compute_resource"] = c.config.ResourcePoolCollectorEnabled
	if c.ComputeResources != nil {
		sensorMetrics["compute_resource"] = c.ComputeResources.GetMetrics()
	}

	return ScraperStatus{
		TCPStatusCheck:         tcpConnect,
		TCPStatusCheckMgs:      tcpErrMsg,
		TCPStatusCheckEndpoint: baeURL.Host,
		SensorEnabled:          sensorEnabled,
		SensorAvailable:        sensorAvailable,
		SensorMetric:           sensorMetrics,
	}
}

func (c *VCenterScraper) startClientPool(ctx context.Context) error {
	client, err := NewVMwareClient(ctx, ClientConf{
		Endpoint: c.config.Endpoint,
		Username: c.config.Username,
		Password: c.config.Password,
	})
	if err != nil {
		return err
	}

	atExitFunc := func() error {
		return client.Logout(ctx)
	}
	c.clientPool = pool.NewMultiAccessPool(client, c.config.ClientPoolSize, atExitFunc)
	return nil
}

func (c *VCenterScraper) Start(logger *slog.Logger) {
	ctx := context.Background()
	for {
		err := c.startClientPool(ctx)
		if err != nil {
			logger.Error("failed to start client pool", "err", err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	c.Host = NewAutoRefreshSensor(newHostRefreshFunc(ctx), cacheConfig{
		MaxAge:             c.config.HostMaxAgeSec,
		RefreshInterval:    c.config.HostRefreshIntervalSec,
		CleanCacheInterval: c.config.CleanIntervalSec,
	})
	c.Host.Start(c.clientPool, logger)
	logger.Info("Host sensor started")

	if c.config.ClusterCollectorEnabled {
		c.Cluster = NewAutoRefreshSensor(newClusterRefreshFunc(ctx), cacheConfig{
			MaxAge:             c.config.ClusterMaxAgeSec,
			RefreshInterval:    c.config.ClusterRefreshIntervalSec,
			CleanCacheInterval: c.config.CleanIntervalSec,
		})
		c.Cluster.Start(c.clientPool, logger)
		logger.Info("Cluster sensor started")
	}

	if c.config.VirtualMachineCollectorEnabled {
		c.VM = NewAutoRefreshSensor(newVMRefreshFunc(ctx, c), cacheConfig{
			MaxAge:             c.config.VirtualMachineMaxAgeSec,
			RefreshInterval:    c.config.VirtualMachineRefreshIntervalSec,
			CleanCacheInterval: c.config.CleanIntervalSec,
		})
		c.VM.Start(c.clientPool, logger)
		logger.Info("VirtualMachine sensor started")
	}

	if c.config.DatastoreCollectorEnabled {
		c.Datastore = NewAutoRefreshSensor(newDatastoreRefreshFunc(ctx), cacheConfig{
			MaxAge:             c.config.DatastoreMaxAgeSec,
			RefreshInterval:    c.config.DatastoreRefreshIntervalSec,
			CleanCacheInterval: c.config.CleanIntervalSec,
		})
		c.Datastore.Start(c.clientPool, logger)
		logger.Info("Datastore sensor started")
	}

	if c.config.SpodCollectorEnabled {
		c.SPOD = NewAutoRefreshSensor(newSpodRefreshFunc(ctx), cacheConfig{
			MaxAge:             c.config.SpodMaxAgeSec,
			RefreshInterval:    c.config.SpodRefreshIntervalSec,
			CleanCacheInterval: c.config.CleanIntervalSec,
		})
		c.SPOD.Start(c.clientPool, logger)
		logger.Info("StoragePod sensor started")
	}

	if c.config.ResourcePoolCollectorEnabled {
		c.ResourcePool = NewAutoRefreshSensor(newResourcePoolRefreshFunc(ctx), cacheConfig{
			MaxAge:             c.config.SpodMaxAgeSec,
			RefreshInterval:    c.config.SpodRefreshIntervalSec,
			CleanCacheInterval: c.config.CleanIntervalSec,
		})
		c.ResourcePool.Start(c.clientPool, logger)
		logger.Info("ResourcePool sensor started")
	}

	c.Remain = NewOnDemandSensor(NewManagedEntityGetFunc(ctx), cacheConfig{
		MaxAge:             c.config.OnDemandCacheMaxAge,
		RefreshInterval:    0,
		CleanCacheInterval: c.config.CleanIntervalSec,
	})
	c.Remain.Start(c.clientPool, logger)
	logger.Info("OnDemand sensor started")
}

func (c *VCenterScraper) Stop(logger *slog.Logger) {

	c.Host.Stop(logger)
	logger.Info("Host sensor stopped")

	if c.config.ClusterCollectorEnabled {
		c.Cluster.Stop(logger)
		logger.Info("Cluster sensor stopped")
	}

	if c.config.VirtualMachineCollectorEnabled {
		c.VM.Stop(logger)
		logger.Info("VirtualMachine sensor stopped")
	}

	if c.config.DatastoreCollectorEnabled {
		c.Datastore.Stop(logger)
		logger.Info("Datastore sensor stopped")
	}

	if c.config.SpodCollectorEnabled {
		c.SPOD.Stop(logger)
		logger.Info("StoragePod sensor stopped")
	}

	if c.config.ResourcePoolCollectorEnabled {
		c.ResourcePool.Stop(logger)
		logger.Info("ResourcePool sensor stopped")
	}

	c.Remain.Stop(logger)
	logger.Info("OnDemand sensor stopped")

	c.clientPool.Close()
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
