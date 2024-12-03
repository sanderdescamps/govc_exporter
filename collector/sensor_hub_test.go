package collector_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/intrinsec/govc_exporter/collector"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

func TestSensorHub(t *testing.T) {
	ctx := context.Background()
	sensHub, err := collector.NexVCenterHub(ctx, collector.SensorHubConf{
		SensorRefreshPeriod: 3600,
		SensorHubJobPeriod:  2,
		Username:            "testuser",
		Password:            "testpass",
		Endpoint:            "https://127.0.0.1:8989",
	})

	if err != nil {
		t.Errorf("Failed to start SensorHub: %v", err)
	}

	sensHub.Login()
	sensHub.Start()
	time.Sleep(10 * time.Second)

	fmt.Println("print logs")
	for _, m := range sensHub.GetMetrics() {
		fmt.Printf("%s{%s} %f\n", m.GetName(), strings.Join(func(labels collector.Labels) []string {
			result := []string{}
			for _, l := range labels {
				result = append(result, fmt.Sprintf("%s=\"%s\"", l.Key, l.Value))
			}
			return result
		}(m.GetLabels()), ", "), m.GetValue())
	}
	sensHub.Stop()
	sensHub.Logout()
}

func TestClient(t *testing.T) {
	ctx := context.Background()
	sensHub, err := collector.NexVCenterHub(ctx, collector.SensorHubConf{
		SensorRefreshPeriod: 3600,
		SensorHubJobPeriod:  2,
		Username:            "testuser",
		Password:            "testpass",
		Endpoint:            "https://127.0.0.1:8989",
	})

	if err != nil {
		t.Errorf("Failed to start SensorHub: %v", err)
	}

	sensHub.Login()

	client := sensHub.GetClient()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"HostSystem"},
		true,
	)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer v.Destroy(ctx)

	var hss []mo.HostSystem
	err = v.Retrieve(
		ctx,
		[]string{"HostSystem"},
		[]string{
			"parent",
			"summary",
		},
		&hss,
	)
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}
	for _, h := range hss {
		t.Logf("host:  %s", h.Summary.Config.Name)
	}

	t.Logf("%v", hss)
	sensHub.Logout()
}
