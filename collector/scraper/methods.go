package scraper

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/intrinsec/govc_exporter/collector/pool"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type SensorType int

// const (
// 	HostSensor SensorType = iota
// 	ClusterSensor
// 	DatastoreSensor
// 	VMSensor
// )

func newHostRefreshFunc(ctx context.Context, scraper *VCenterScraper) func(pool.Pool[govmomi.Client], *slog.Logger) ([]mo.HostSystem, error) {
	return func(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) ([]mo.HostSystem, error) {
		t1 := time.Now()
		client, release, err := clientPool.Acquire()
		if err != nil {
			return nil, err
		}
		defer release()
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
				"config.storageDevice",
				"config.fileSystemVolume",
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

func newVMTagsRefreshFunc(ctx context.Context, scraper *VCenterScraper) func(*pool.VCenterClientPool, *slog.Logger) (map[types.ManagedObjectReference][]*tags.Tag, error) {
	return func(clientPool *pool.VCenterClientPool, logger *slog.Logger) (map[types.ManagedObjectReference][]*tags.Tag, error) {
		t1 := time.Now()
		restclient, release, err := clientPool.AcquireRest()
		if err != nil {
			return nil, err
		}
		defer release()
		t2 := time.Now()

		m := tags.NewManager(restclient)

		tagList, err := m.ListTags(ctx)
		if err != nil {
			log.Print(err)
			return nil, err
		}

		r := make(map[types.ManagedObjectReference][]*tags.Tag)
		for _, tag := range tagList {
			attachObjs, err := m.GetAttachedObjectsOnTags(ctx, []string{tag})
			if err != nil {
				return nil, err
			}

			for _, attachObj := range attachObjs {
				for _, elem := range attachObj.ObjectIDs {
					r[elem.Reference()] = append(r[elem.Reference()], attachObj.Tag)
				}
			}
		}
		t3 := time.Now()
		scraper.dispatchMetrics(SensorMetric{
			Name:           "tags",
			QueryTime:      t3.Sub(t2),
			ClientWaitTime: t2.Sub(t1),
			Status:         true,
		})
		return r, nil
	}
}

func newClusterRefreshFunc(ctx context.Context, scraper *VCenterScraper) func(pool.Pool[govmomi.Client], *slog.Logger) ([]mo.ComputeResource, error) {
	return func(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) ([]mo.ComputeResource, error) {
		t1 := time.Now()
		client, release, err := clientPool.Acquire()
		if err != nil {
			return nil, err
		}
		defer release()
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
		client, release, err := clientPool.Acquire()
		if err != nil {
			return nil, err
		}
		defer release()
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
		client, release, err := clientPool.Acquire()
		if err != nil {
			return nil, err
		}
		defer release()
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
		hostRefs, err := func() ([]types.ManagedObjectReference, error) {
			resultChan := make(chan *[]types.ManagedObjectReference)
			quitChan := make(chan bool)
			go func() {
				ticker := time.NewTicker(1 * time.Second)
				for {
					select {
					case <-ticker.C:
						refs := scraper.Host.GetAllRefs()
						if refs != nil {
							resultChan <- &refs
							break
						}
						logger.Info("VM sensor is waiting for hosts")
					case <-quitChan:
						return
					}
				}
			}()

			select {
			case <-time.After(10 * time.Second):
				quitChan <- true
				logger.Warn("Scraper can not request virtual machines without hosts")
				return nil, fmt.Errorf("waiting for hosts timeout")
			case refs := <-resultChan:
				return *refs, nil
			}
		}()

		if err != nil {
			scraper.dispatchMetrics(SensorMetric{
				Name:   "vm",
				Status: false,
			})
			return nil, err
		}

		var wg sync.WaitGroup
		wg.Add(len(hostRefs))
		resultChan := make(chan *[]mo.VirtualMachine, len(hostRefs))
		errChan := make(chan error, len(hostRefs))
		metricChan := make(chan SensorMetric, len(hostRefs))
		for _, host := range hostRefs {
			go func() {
				t1 := time.Now()
				client, release, err := clientPool.Acquire()
				if err != nil {
					errChan <- err
					return
				}
				defer release()
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

		readyChan := make(chan bool)
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
					return
				}
			}
		}()

		wg.Wait()
		readyChan <- true

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
		client, release, err := clientPool.Acquire()
		if err != nil {
			return nil, err
		}
		defer release()
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
		client, release, err := clientPool.Acquire()
		if err != nil {
			return nil, err
		}
		defer release()
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

func NewManagedEntityGetFunc(ctx context.Context, scraper *VCenterScraper, logger *slog.Logger) func(types.ManagedObjectReference, pool.Pool[govmomi.Client]) (*mo.ManagedEntity, error) {
	return func(ref types.ManagedObjectReference, clientPool pool.Pool[govmomi.Client]) (*mo.ManagedEntity, error) {
		t1 := time.Now()
		client, release, err := clientPool.Acquire()
		if err != nil {
			return nil, err
		}
		defer release()
		t2 := time.Now()
		var entity mo.ManagedEntity
		pc := property.DefaultCollector(client.Client)
		err = pc.RetrieveOne(ctx, ref, []string{"name", "parent"}, &entity)
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
