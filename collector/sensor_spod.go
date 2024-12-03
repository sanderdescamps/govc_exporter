package collector

import (
	"context"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

func NewSpodRefreshFunc(ctx context.Context, sh *SensorHub) func() ([]Metric, error) {
	return func() ([]Metric, error) {
		c := sh.GetClient()

		var items []mo.StoragePod

		m := view.NewManager(c.Client)
		v, err := m.CreateContainerView(
			ctx,
			c.ServiceContent.RootFolder,
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
				"parent",
				"summary",
			},
			&items,
		)
		if err != nil {
			return nil, err
		}

		metrics := []Metric{}
		for _, item := range items {
			summary := item.Summary
			// name := summary.Name
			// tmp := getParents(c.ctx, c.logger, c.client.Client, item.ManagedEntity)

			labels := Labels{}
			labels = labels.Add("name", item.Summary.Name)

			metrics = append(metrics, NewBasicMetric("total_capacity_bytes", float64(summary.Capacity), labels))
			metrics = append(metrics, NewBasicMetric("free_space_bytes", float64(summary.FreeSpace), labels))
		}
		return metrics, nil
	}
}
