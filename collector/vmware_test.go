package collector_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
)

func GetClient(t *testing.T) *govmomi.Client {
	endpoint := "https://localhost:8989"
	username := "testuser"
	password := "testpass"

	u, err := soap.ParseURL(endpoint)
	if err != nil {
		panic(err)
	}
	u.User = url.UserPassword(username, password)
	client, err := govmomi.NewClient(context.Background(), u, true)
	if err != nil {
		t.Fatalf("Client creation failed: %v", err)
	}
	return client
}

func TestVMwareHost(t *testing.T) {

	var hss []mo.HostSystem

	client := GetClient(t)
	ctx := context.Background()
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

}

func TestVMwareComputeResource(t *testing.T) {

	var hss []mo.ComputeResource

	client := GetClient(t)
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"ComputeResource"},
		true,
	)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"ComputeResource"},
		[]string{
			// "parent",
			// "summary",
		},
		&hss,
	)
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}
	for _, h := range hss {
		t.Logf("host:  %s", h.Name)
	}

	t.Logf("%v", hss)

}

func TestVMwareCluster(t *testing.T) {

	var hss []mo.ClusterComputeResource

	client := GetClient(t)
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"ClusterComputeResource"},
		true,
	)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"ClusterComputeResource"},
		[]string{
			// "parent",
			// "summary",
		},
		&hss,
	)
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}
	for _, h := range hss {
		t.Logf("host:  %s", h.Name)
	}

	t.Logf("%v", hss)

}

func TestVMwareVM(t *testing.T) {
	var items []mo.VirtualMachine
	client := GetClient(t)
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"VirtualMachine"},
		true,
	)
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"VirtualMachine"},
		[]string{
			"config",
			//"datatore",
			"guest",
			"guestHeartbeatStatus",
			"network",
			"parent",
			"resourceConfig",
			"resourcePool",
			"runtime",
			"snapshot",
			"summary",
		},
		&items,
	)

	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}
	for _, vm := range items {
		t.Logf("host:  %s", vm.Config.Name)
	}
}

// func TestVMwareCache(t *testing.T) {
// 	activeCache := collector.NewVMwareActiveCache(collector.VMwareConfig{
// 		RefreshPeriod: 5,
// 		Endpoint:      "https://127.0.0.1:8989",
// 		Username:      "testuser",
// 		Password:      "testpass",
// 	})

// 	var items []mo.VirtualMachine
// 	err := activeCache.GetAllWithKindFromCache("HostSystem", items)
// 	if err != nil {
// 		t.Errorf("%v", err)
// 		t.Fail()
// 	}

// 	ctx := context.Background()
// 	m := view.NewManager(client.Client)
// 	v, err := m.CreateContainerView(
// 		ctx,
// 		client.ServiceContent.RootFolder,
// 		[]string{"HostSystem"},
// 		true,
// 	)
// 	if err != nil {
// 		t.Errorf("%v", err)
// 		t.Fail()
// 	}
// 	defer v.Destroy(ctx)

// 	err = v.Retrieve(
// 		ctx,
// 		[]string{"VirtualMachine"},
// 		[]string{
// 			"config",
// 			//"datatore",
// 			"guest",
// 			"guestHeartbeatStatus",
// 			"network",
// 			"parent",
// 			"resourceConfig",
// 			"resourcePool",
// 			"runtime",
// 			"snapshot",
// 			"summary",
// 		},
// 		&items,
// 	)

// 	if err != nil {
// 		t.Errorf("%v", err)
// 		t.Fail()
// 	}
// 	for _, vm := range items {
// 		t.Logf("host:  %s", vm.Config.Name)
// 	}
// }
