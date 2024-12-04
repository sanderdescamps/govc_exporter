package collector

import (
	"context"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

func NewHostRefreshFunc(ctx context.Context, sh *SensorHub) func() ([]mo.HostSystem, error) {
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

func NewClusterRefreshFunc(ctx context.Context, sh *SensorHub) func() ([]mo.HostSystem, error) {
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
