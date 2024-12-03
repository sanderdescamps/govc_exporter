package collector

import (
	"context"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

func NewVMRefreshFunc(ctx context.Context, sh *SensorHub) func() ([]Metric, error) {
	return func() ([]Metric, error) {

		var items []mo.Datastore
		c := sh.GetClient()
		m := view.NewManager(c.Client)
		v, err := m.CreateContainerView(
			ctx,
			c.ServiceContent.RootFolder,
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
			labels := Labels{}
			labels = labels.Add("name", summary.Name)
			metrics = append(metrics, NewBasicMetric("total_capacity_bytes", float64(summary.Capacity), labels))
			metrics = append(metrics, NewBasicMetric("free_space_bytes", float64(summary.FreeSpace), labels))
			metrics = append(metrics, NewBasicMetric("accessible", b2f(summary.Accessible), labels))
			metrics = append(metrics, NewBasicMetric("maintenance_mode", ConvertDatastoreMaintenanceModeStateToValue(summary.MaintenanceMode), labels))
		}

		return metrics, nil
	}
}
