package collector

import (
	"github.com/vmware/govmomi/vim25/types"
)

// func NewHostRefreshFunc(ctx context.Context, sh *SensorHub) func() ([]Metric, error) {
// 	return func() ([]Metric, error) {
// 		c := sh.GetClient()

// 		m := view.NewManager(c.Client)
// 		v, err := m.CreateContainerView(
// 			ctx,
// 			c.ServiceContent.RootFolder,
// 			[]string{"HostSystem"},
// 			true,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer v.Destroy(ctx)

// 		var items []mo.HostSystem
// 		err = v.Retrieve(
// 			context.Background(),
// 			[]string{"HostSystem"},
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
// 			labels = labels.Add("name", summary.Config.Name)
// 			// labels = labels.Add("cluster", ????)
// 			// labels = labels.Add("dc", ????)
// 			powerState := ConvertHostSystemPowerStateToValue(summary.Runtime.PowerState)
// 			connState := ConvertHostSystemConnectionStateToValue(summary.Runtime.ConnectionState)
// 			maintenance := b2f(item.Runtime.InMaintenanceMode)
// 			metrics = append(metrics, NewBasicMetric("power_state", powerState, labels))
// 			metrics = append(metrics, NewBasicMetric("connection_state", connState, labels))
// 			metrics = append(metrics, NewBasicMetric("maintenance", maintenance, labels))
// 			metrics = append(metrics, NewBasicMetric("uptime", float64(summary.QuickStats.Uptime), labels))
// 			metrics = append(metrics, NewBasicMetric("reboot_required", b2f(summary.RebootRequired), labels))
// 			metrics = append(metrics, NewBasicMetric("num_cpu_cores", float64(summary.Hardware.NumCpuCores), labels))
// 			metrics = append(metrics, NewBasicMetric("avail_cpu_mhz", float64(int64(summary.Hardware.NumCpuCores)*int64(summary.Hardware.CpuMhz)), labels))
// 			metrics = append(metrics, NewBasicMetric("used_cpu_mhz", float64(summary.QuickStats.OverallCpuUsage), labels))
// 			metrics = append(metrics, NewBasicMetric("avail_mem_bytes", float64(summary.Hardware.MemorySize), labels))
// 			metrics = append(metrics, NewBasicMetric("used_mem_bytes", float64(int64(summary.QuickStats.OverallMemoryUsage)*int64(1024*1024)), labels))

// 			//Version Info
// 			labels = labels.Add("esxi_version", summary.Config.Product.Version)
// 			labels = labels.Add("cpu_model", summary.Hardware.CpuModel)
// 			labels = labels.Add("vendor", summary.Hardware.Vendor)
// 			labels = labels.Add("model", summary.Hardware.Model)
// 			metrics = append(metrics, NewBasicMetric("host_info", b2f(summary.RebootRequired), labels))
// 		}

// 		return metrics, nil
// 	}
// }

func ConvertHostSystemPowerStateToValue(s types.HostSystemPowerState) float64 {
	if s == types.HostSystemPowerStateStandBy {
		return 1.0
	} else if s == types.HostSystemPowerStatePoweredOn {
		return 2.0
	}
	return 0
}

func ConvertHostSystemConnectionStateToValue(s types.HostSystemConnectionState) float64 {
	if s == types.HostSystemConnectionStateNotResponding {
		return 1.0
	} else if s == types.HostSystemConnectionStateConnected {
		return 2.0
	}
	return 0
}

func ConvertHostSystemStandbyModeToValue(s types.HostStandbyMode) float64 {
	if s == types.HostStandbyModeExiting {
		return 1.0
	} else if s == types.HostStandbyModeEntering {
		return 2.0
	} else if s == types.HostStandbyModeIn {
		return 2.0
	}
	return 0
}
