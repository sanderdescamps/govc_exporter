package scraper

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/intrinsec/govc_exporter/collector/pool"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type SensorType int

const (
	HostSensor SensorType = iota
	ClusterSensor
	DatastoreSensor
	VMSensor
)

func newHostRefreshFunc(ctx context.Context, scraper *VCenterScraper) func(pool.Pool[govmomi.Client], *slog.Logger) ([]mo.HostSystem, error) {
	return func(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) ([]mo.HostSystem, error) {
		t1 := time.Now()
		client, clientID := clientPool.Acquire()
		defer clientPool.Release(clientID)
		t2 := time.Now()
		m := view.NewManager(client.Client)
		v, err := m.CreateContainerView(
			ctx,
			client.ServiceContent.RootFolder,
			[]string{"HostSystem"},
			true,
		)
		if err != nil {
			return nil, err
		}
		defer v.Destroy(ctx)

		var items []mo.HostSystem
		err = v.Retrieve(
			context.Background(),
			[]string{"HostSystem"},
			[]string{
				"name",
				"parent",
				"summary",
				"runtime",
				// "network",
			},
			&items,
		)
		t3 := time.Now()
		scraper.dispatchMetrics(SensorMetric{
			Name:           "host",
			QueryTime:      t3.Sub(t2),
			ClientWaitTime: t2.Sub(t1),
			Status:         true,
		})
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}

func newClusterRefreshFunc(ctx context.Context, scraper *VCenterScraper) func(pool.Pool[govmomi.Client], *slog.Logger) ([]mo.ComputeResource, error) {
	return func(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) ([]mo.ComputeResource, error) {
		t1 := time.Now()
		client, clientID := clientPool.Acquire()
		defer clientPool.Release(clientID)
		t2 := time.Now()
		m := view.NewManager(client.Client)
		v, err := m.CreateContainerView(
			ctx,
			client.ServiceContent.RootFolder,
			[]string{"ClusterComputeResource"},
			true,
		)
		if err != nil {
			return nil, err
		}
		defer v.Destroy(ctx)

		var items []mo.ClusterComputeResource
		err = v.Retrieve(
			context.Background(),
			[]string{"ClusterComputeResource"},
			[]string{
				"parent",
				"name",
				"summary",
			},
			&items,
		)
		t3 := time.Now()
		scraper.dispatchMetrics(SensorMetric{
			Name:           "cluster",
			QueryTime:      t3.Sub(t2),
			ClientWaitTime: t2.Sub(t1),
			Status:         true,
		})
		if err != nil {
			return nil, err
		}

		computeResources := []mo.ComputeResource{}
		for _, c := range items {
			computeResources = append(computeResources, c.ComputeResource)
		}

		return computeResources, nil
	}
}

func newComputeResourceRefreshFunc(ctx context.Context, scraper *VCenterScraper) func(pool.Pool[govmomi.Client], *slog.Logger) ([]mo.ComputeResource, error) {
	return func(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) ([]mo.ComputeResource, error) {
		t1 := time.Now()
		client, clientID := clientPool.Acquire()
		defer clientPool.Release(clientID)
		t2 := time.Now()
		m := view.NewManager(client.Client)
		v, err := m.CreateContainerView(
			ctx,
			client.ServiceContent.RootFolder,
			[]string{"ComputeResource"},
			true,
		)
		if err != nil {
			return nil, err
		}
		defer v.Destroy(ctx)

		var items []mo.ComputeResource
		err = v.Retrieve(
			context.Background(),
			[]string{"ComputeResource"},
			[]string{
				"parent",
				"summary",
				"name",
			},
			&items,
		)
		t3 := time.Now()
		scraper.dispatchMetrics(SensorMetric{
			Name:           "compute_resource",
			QueryTime:      t3.Sub(t2),
			ClientWaitTime: t2.Sub(t1),
			Status:         true,
		})
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}

func newResourcePoolRefreshFunc(ctx context.Context, scraper *VCenterScraper) func(pool.Pool[govmomi.Client], *slog.Logger) ([]mo.ResourcePool, error) {
	return func(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) ([]mo.ResourcePool, error) {
		t1 := time.Now()
		client, clientID := clientPool.Acquire()
		defer clientPool.Release(clientID)
		t2 := time.Now()
		m := view.NewManager(client.Client)
		v, err := m.CreateContainerView(
			ctx,
			client.ServiceContent.RootFolder,
			[]string{"ResourcePool"},
			true,
		)
		if err != nil {
			return nil, err
		}
		defer v.Destroy(ctx)

		var items []mo.ResourcePool
		err = v.Retrieve(
			context.Background(),
			[]string{"ResourcePool"},
			[]string{
				"parent",
				"summary",
				"name",
			},
			&items,
		)
		t3 := time.Now()
		scraper.dispatchMetrics(SensorMetric{
			Name:           "resource_pool",
			QueryTime:      t3.Sub(t2),
			ClientWaitTime: t2.Sub(t1),
			Status:         true,
		})
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}

func newVMRefreshFunc(ctx context.Context, scraper *VCenterScraper) func(pool.Pool[govmomi.Client], *slog.Logger) ([]mo.VirtualMachine, error) {
	return func(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) ([]mo.VirtualMachine, error) {

		var hostRefs []types.ManagedObjectReference
		func() {
			result := make(chan bool)
			go func() {
				for {
					hostRefs = scraper.Host.GetAllRefs()
					if hostRefs != nil {
						break
					}
					logger.Info("VM sensor is waiting for hosts")
					time.Sleep(1 * time.Second)
				}
				result <- true
			}()

			select {
			case <-time.After(5 * time.Second):
				logger.Warn("Scraper can not request virtual machines as it needs to find hosts first.")
			case <-result:
				return
			}
		}()

		var wg sync.WaitGroup
		wg.Add(len(hostRefs))
		resultChan := make(chan *[]mo.VirtualMachine, len(hostRefs))
		errChan := make(chan error, len(hostRefs))
		metricChan := make(chan SensorMetric, len(hostRefs))
		for _, host := range hostRefs {
			go func() {
				t1 := time.Now()
				client, clientID := clientPool.Acquire()
				defer clientPool.Release(clientID)
				defer wg.Done()
				t2 := time.Now()
				m := view.NewManager(client.Client)
				v, err := m.CreateContainerView(
					ctx,
					host,
					[]string{"VirtualMachine"},
					true,
				)
				if err != nil {
					errChan <- err
					return
				}
				defer v.Destroy(ctx)

				var items []mo.VirtualMachine
				err = v.Retrieve(
					context.Background(),
					[]string{"VirtualMachine"},
					[]string{
						"name",
						"config",
						//"datatore",
						"guest",
						"guestHeartbeatStatus",
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
				t3 := time.Now()
				if err != nil {
					errChan <- err
				} else {
					resultChan <- &items
					metricChan <- SensorMetric{
						Name:           "vm",
						QueryTime:      t3.Sub(t2),
						ClientWaitTime: t2.Sub(t1),
						Status:         true,
					}
				}
			}()
		}
		vms := []mo.VirtualMachine{}
		// for i := 0; i < len(hostRefs); i++ {
		// 	result := <-resultChan
		// 	if result != nil {
		// 		vms = append(vms, *result...)
		// 	}
		// }

		for len(resultChan) > 0 {
			<-resultChan
		}

		readyChan := make(chan bool)
		go func() {
			//Wait for all goroutines to finish and trigger the readyChan
			wg.Wait()
			readyChan <- true
		}()

		failed := false
		var allClientWaitTimes []time.Duration
		var allQueryTimes []time.Duration
		go func() {
			for {
				select {
				case r := <-resultChan:
					vms = append(vms, *r...)
				case m := <-metricChan:
					allClientWaitTimes = append(allClientWaitTimes, m.ClientWaitTime)
					allQueryTimes = append(allQueryTimes, m.QueryTime)
				case err := <-errChan:
					logger.Error("failed to get vm's from host", "err", err)
					failed = true
				case <-readyChan:
					break
				}
			}
		}()

		scraper.dispatchMetrics(SensorMetric{
			Name:           "vm",
			ClientWaitTime: avgDuration(allClientWaitTimes),
			QueryTime:      avgDuration(allQueryTimes),
			Status:         failed,
		})

		return vms, nil
	}
}

func newSpodRefreshFunc(ctx context.Context, scraper *VCenterScraper) func(pool.Pool[govmomi.Client], *slog.Logger) ([]mo.StoragePod, error) {
	return func(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) ([]mo.StoragePod, error) {
		t1 := time.Now()
		client, clientID := clientPool.Acquire()
		defer clientPool.Release(clientID)
		t2 := time.Now()
		m := view.NewManager(client.Client)
		v, err := m.CreateContainerView(
			ctx,
			client.ServiceContent.RootFolder,
			[]string{"StoragePod"},
			true,
		)
		if err != nil {
			return nil, err
		}
		defer v.Destroy(ctx)

		var items []mo.StoragePod
		err = v.Retrieve(
			ctx,
			[]string{"StoragePod"},
			[]string{
				"name",
				"parent",
				"summary",
			},
			&items,
		)
		t3 := time.Now()
		scraper.dispatchMetrics(SensorMetric{
			Name:           "spod",
			QueryTime:      t3.Sub(t2),
			ClientWaitTime: t2.Sub(t1),
			Status:         true,
		})
		if err != nil {
			return nil, err
		}
		return items, nil
	}
}

func newDatastoreRefreshFunc(ctx context.Context, scraper *VCenterScraper) func(pool.Pool[govmomi.Client], *slog.Logger) ([]mo.Datastore, error) {
	return func(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) ([]mo.Datastore, error) {
		t1 := time.Now()
		client, clientID := clientPool.Acquire()
		defer clientPool.Release(clientID)
		t2 := time.Now()
		m := view.NewManager(client.Client)
		v, err := m.CreateContainerView(
			ctx,
			client.ServiceContent.RootFolder,
			[]string{"Datastore"},
			true,
		)
		if err != nil {
			return nil, err
		}
		defer v.Destroy(ctx)

		var items []mo.Datastore
		err = v.Retrieve(
			ctx,
			[]string{"Datastore"},
			[]string{
				"name",
				"parent",
				"summary",
				"info",
			},
			&items,
		)
		t3 := time.Now()
		scraper.dispatchMetrics(SensorMetric{
			Name:           "datastore",
			QueryTime:      t3.Sub(t2),
			ClientWaitTime: t2.Sub(t1),
			Status:         true,
		})
		if err != nil {
			return nil, err
		}
		return items, nil
	}
}

func NewManagedEntityGetFunc(ctx context.Context, scraper *VCenterScraper) func(types.ManagedObjectReference, pool.Pool[govmomi.Client], *slog.Logger) (*mo.ManagedEntity, error) {
	return func(ref types.ManagedObjectReference, clientPool pool.Pool[govmomi.Client], logger *slog.Logger) (*mo.ManagedEntity, error) {
		t1 := time.Now()
		client, clientID := clientPool.Acquire()
		defer clientPool.Release(clientID)
		t2 := time.Now()
		var entity mo.ManagedEntity
		pc := property.DefaultCollector(client.Client)
		err := pc.RetrieveOne(ctx, ref, []string{"name", "parent"}, &entity)
		t3 := time.Now()
		scraper.dispatchMetrics(SensorMetric{
			Name:           "on_demand",
			QueryTime:      t3.Sub(t2),
			ClientWaitTime: t2.Sub(t1),
			Status:         true,
		})
		if err != nil {
			return nil, err
		}
		return &entity, nil
	}
}
