package collector

import (
	"context"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func NewDatastoreRefreshFunc(ctx context.Context, sh *SensorHub) func() ([]mo.Datastore, error) {
	return func() ([]mo.Datastore, error) {

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

		return items, nil
	}
}

// func NewDatastoreRefreshFunc(ctx context.Context, sh *SensorHub) func() ([]Metric, error) {
// 	return func() ([]Metric, error) {

// 		var items []mo.Datastore
// 		c := sh.GetClient()
// 		m := view.NewManager(c.Client)
// 		v, err := m.CreateContainerView(
// 			ctx,
// 			c.ServiceContent.RootFolder,
// 			[]string{"Datastore"},
// 			true,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer v.Destroy(ctx)

// 		err = v.Retrieve(
// 			ctx,
// 			[]string{"Datastore"},
// 			[]string{
// 				"parent",
// 				"summary",
// 			},
// 			&items,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}

// 		metrics := []Metric{}
// 		for _, item := range items {
// 			summary := item.Summary
// 			labels := Labels{}
// 			labels = labels.Add("name", summary.Name)
// 			metrics = append(metrics, NewBasicMetric("total_capacity_bytes", float64(summary.Capacity), labels))
// 			metrics = append(metrics, NewBasicMetric("free_space_bytes", float64(summary.FreeSpace), labels))
// 			metrics = append(metrics, NewBasicMetric("accessible", b2f(summary.Accessible), labels))
// 			metrics = append(metrics, NewBasicMetric("maintenance_mode", ConvertDatastoreMaintenanceModeStateToValue(summary.MaintenanceMode), labels))
// 		}

// 		return metrics, nil
// 	}
// }

func ConvertDatastoreMaintenanceModeStateToValue(d string) float64 {
	dTyped := types.DatastoreSummaryMaintenanceModeState(d)
	if dTyped == types.DatastoreSummaryMaintenanceModeStateEnteringMaintenance {
		return 1.0
	} else if dTyped == types.DatastoreSummaryMaintenanceModeStateInMaintenance {
		return 2.0
	}
	return 0
}
