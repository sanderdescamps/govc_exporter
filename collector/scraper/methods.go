package scraper

import (
	"context"
	"log/slog"
	"time"

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

func newHostRefreshFunc(ctx context.Context) func(*govmomi.Client, *slog.Logger) ([]mo.HostSystem, error) {
	return func(client *govmomi.Client, logger *slog.Logger) ([]mo.HostSystem, error) {
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
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}

func newClusterRefreshFunc(ctx context.Context) func(*govmomi.Client, *slog.Logger) ([]mo.ComputeResource, error) {
	return func(client *govmomi.Client, logger *slog.Logger) ([]mo.ComputeResource, error) {
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

func newComputeResourceRefreshFunc(ctx context.Context) func(*govmomi.Client, *slog.Logger) ([]mo.ComputeResource, error) {
	return func(client *govmomi.Client, logger *slog.Logger) ([]mo.ComputeResource, error) {
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
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}

func newResourcePoolRefreshFunc(ctx context.Context) func(*govmomi.Client, *slog.Logger) ([]mo.ResourcePool, error) {
	return func(client *govmomi.Client, logger *slog.Logger) ([]mo.ResourcePool, error) {
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
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}

func newVMRefreshFunc(ctx context.Context, scraper *VCenterScraper) func(*govmomi.Client, *slog.Logger) ([]mo.VirtualMachine, error) {
	return func(client *govmomi.Client, logger *slog.Logger) ([]mo.VirtualMachine, error) {

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

		resultChan := make(chan *[]mo.VirtualMachine, len(hostRefs))
		errChan := make(chan error, 1)
		for _, host := range hostRefs {
			go func(errChan chan error) {
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
				if err != nil {
					errChan <- err
					resultChan <- nil
				} else {
					resultChan <- &items
				}
			}(errChan)
		}
		vms := []mo.VirtualMachine{}
		for i := 0; i < len(hostRefs); i++ {
			result := <-resultChan
			if result != nil {
				vms = append(vms, *result...)
			}
		}
		return vms, nil
	}
}

func newSpodRefreshFunc(ctx context.Context) func(*govmomi.Client, *slog.Logger) ([]mo.StoragePod, error) {
	return func(client *govmomi.Client, logger *slog.Logger) ([]mo.StoragePod, error) {
		var items []mo.StoragePod

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
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}

func newDatastoreRefreshFunc(ctx context.Context) func(*govmomi.Client, *slog.Logger) ([]mo.Datastore, error) {
	return func(client *govmomi.Client, logger *slog.Logger) ([]mo.Datastore, error) {
		var items []mo.Datastore
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
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}

func NewManagedEntityGetFunc(ctx context.Context) func(types.ManagedObjectReference, *govmomi.Client, *slog.Logger) (*mo.ManagedEntity, error) {
	return func(ref types.ManagedObjectReference, client *govmomi.Client, logger *slog.Logger) (*mo.ManagedEntity, error) {
		var entity mo.ManagedEntity
		pc := property.DefaultCollector(client.Client)
		err := pc.RetrieveOne(ctx, ref, []string{"name", "parent"}, &entity)
		if err != nil {
			return nil, err
		}
		return &entity, nil
	}
}
