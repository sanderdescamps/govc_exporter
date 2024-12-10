package collector

import (
	"context"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

type SensorType int

const (
	HostSensor SensorType = iota
	ClusterSensor
	DatastoreSensor
	VMSensor
)

func NewHostRefreshFunc(ctx context.Context, sh *SensorHub) (SensorType, func() ([]mo.HostSystem, error)) {
	return HostSensor, func() ([]mo.HostSystem, error) {
		c := sh.GetClient()

		m := view.NewManager(c.Client)
		v, err := m.CreateContainerView(
			ctx,
			c.ServiceContent.RootFolder,
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

func NewClusterRefreshFunc(ctx context.Context, sh *SensorHub) (SensorType, func() ([]mo.ComputeResource, error)) {
	return ClusterSensor, func() ([]mo.ComputeResource, error) {
		c := sh.GetClient()

		m := view.NewManager(c.Client)
		v, err := m.CreateContainerView(
			ctx,
			c.ServiceContent.RootFolder,
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
				// "parent",
				// "summary",
			},
			&items,
		)
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}

func NewHostPerClusterRefreshFunc(ctx context.Context, sh *SensorHub, cluster string) func() ([]mo.HostSystem, error) {
	return func() ([]mo.HostSystem, error) {
		c := sh.GetClient()

		m := view.NewManager(c.Client)
		v, err := m.CreateContainerView(
			ctx,
			c.ServiceContent.RootFolder,
			[]string{"HostSystem"},
			true,
		)
		if err != nil {
			return nil, err
		}
		defer v.Destroy(ctx)

		// filter := property.Filter{}
		// filter.MatchAnyPropertyList([]types.DynamicProperty{})

		var items []mo.HostSystem
		err = v.RetrieveWithFilter(
			context.Background(),
			[]string{"HostSystem"},
			[]string{
				"parent",
				"summary",
			},
			&items, property.Filter{},
		)
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}

func NewVMPerHostRefreshFunc(ctx context.Context, sh *SensorHub, cluster string) func() ([]mo.VirtualMachine, error) {
	return func() ([]mo.VirtualMachine, error) {
		c := sh.GetClient()

		m := view.NewManager(c.Client)
		v, err := m.CreateContainerView(
			ctx,
			c.ServiceContent.RootFolder,
			[]string{"VirtualMachine"},
			true,
		)
		if err != nil {
			return nil, err
		}
		defer v.Destroy(ctx)

		// filter := property.Filter{}
		// filter.MatchAnyPropertyList([]types.DynamicProperty{})

		var items []mo.VirtualMachine
		err = v.RetrieveWithFilter(
			context.Background(),
			[]string{"VirtualMachine"},
			[]string{
				"parent",
				"summary",
			},
			&items, property.Filter{},
		)
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}
