package scraper_test

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
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

// func TestVCenterScraper(t *testing.T) {

// 	conf := scraper.NewDefaultScraperConfig()
// 	conf.Endpoint = "https://localhost:8989"
// 	conf.Username = "testuser"
// 	conf.Password = "testpass"

// 	aCache, _ := scraper.NewVCenterScraper(conf)

// 	promlogConfig := &promslog.Config{
// 		// Level:
// 	}
// 	logger := promslog.New(promlogConfig)

// 	t1 := time.Now()
// 	aCache.Host.Refresh(logger)
// 	t2 := time.Now()
// 	fmt.Printf("fetching all hosts took %dms\n", t2.Sub(t1).Milliseconds())

// 	aCache.Datastore.Refresh(logger)
// 	t3 := time.Now()
// 	fmt.Printf("fetching all datastores took %dms\n", t3.Sub(t2).Milliseconds())

// 	aCache.SPOD.Refresh(logger)
// 	t4 := time.Now()
// 	fmt.Printf("fetching all spod took %dms\n", t4.Sub(t3).Milliseconds())

// 	aCache.VM.Refresh(logger)
// 	t5 := time.Now()
// 	fmt.Printf("fetching all VMs took %dms\n", t5.Sub(t4).Milliseconds())

// 	aCache.Cluster.Refresh(logger)
// 	t6 := time.Now()
// 	fmt.Printf("fetching all clusters took %dms\n", t6.Sub(t5).Milliseconds())

// 	aCache.Start(logger)

// 	ticker := time.NewTicker(10 * time.Second)

// 	for t := range ticker.C {
// 		fmt.Printf("Metrics at: %s\n", t.String())
// 		// clusters := aCache.Cluster.GetAll()
// 		// for _, c := range clusters {
// 		// 	fmt.Printf("cluster: %s\n", c.ManagedEntity.Name)
// 		// }

// 		// hosts := aCache.Host.GetAll()
// 		// for _, h := range hosts {
// 		// 	fmt.Printf("host: Name=%s Self.Type=%s Self.Value=%s\n", h.Name, h.Self.Type, h.Self.Value)
// 		// 	pc := aCache.GetParentChain(h.Self)
// 		// 	fmt.Printf("host: Name=%s chain=%s\n", h.Name, strings.Join(pc.Chain, ","))
// 		// 	fmt.Printf("host: Name=%s Self.Type=%s Self.Value=%s Cluster=%s Datacenter=%s\n", h.Name, h.Self.Type, h.Self.Value, pc.Cluster, pc.DC)
// 		// }

// 		// // vms := aCache.VM.GetAll()
// 		// // for _, vm := range vms {
// 		// // 	fmt.Printf("virtual-machine: Name=%s Self.Type=%s Self.Value=%s\n", vm.Name, vm.Self.Type, vm.Self.Value)
// 		// // }

// 		// datastores := aCache.Datastore.GetAll()
// 		// for _, d := range datastores {
// 		// 	pc := aCache.GetParentChain(d.Self)
// 		// 	fmt.Printf("datastore: Name=%s chain=%s\n", d.Name, strings.Join(pc.Chain, ","))
// 		// 	fmt.Printf("datastore: Name=%s Self.Type=%s Self.Value=%s SPOD=%s Datacenter=%s\n", d.Name, d.Self.Type, d.Self.Value, pc.SPOD, pc.DC)
// 		// }

// 		// spods := aCache.SPOD.GetAll()
// 		// for _, s := range spods {
// 		// 	fmt.Printf("storage-pods: Name=%s Self.Type=%s Self.Value=%s\n", s.Name, s.Self.Type, s.Self.Value)
// 		// }

// 		fmt.Print()
// 	}

// 	// time.Sleep(60 * time.Second)
// 	// ticker.Stop()

// 	fmt.Print("end")
// }

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
			"runtime",
			"name",
			// "vm",
			"network",
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

func TestVMwareNetwork(t *testing.T) {

	var items []mo.Network

	client := GetClient(t)
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"Network"},
		true,
	)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"Network"},
		[]string{
			// "parent",
			// "summary",
			// "runtime",
			// "name",
			// // "vm",
			// "network",
		},
		&items,
	)
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}
	for _, item := range items {
		t.Logf("Network:  %s", item.Name)
	}

}

func TestVMwareCluster(t *testing.T) {

	var clusters []mo.ClusterComputeResource

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
		&clusters,
	)
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}
	for _, c := range clusters {
		cr := c.ComputeResource
		t.Logf("cluster:  Name=%s Self.Type=%s Self.Value=%s", cr.Name, cr.Self.Type, cr.Self.Value)
	}
}

func TestVMwareComputeResource(t *testing.T) {

	var compResources []mo.ComputeResource

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
			"parent",
			"summary",
			"name",
		},
		&compResources,
	)
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}

}

func TestVMwareFolder(t *testing.T) {

	var folders []mo.Folder

	client := GetClient(t)
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"Folder"},
		true,
	)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"Folder"},
		[]string{
			// "parent",
			// "summary",
			// "name",
		},
		&folders,
	)
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}

}

func TestVMwareDatastore(t *testing.T) {

	var items []mo.Datastore

	client := GetClient(t)
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"Datastore"},
		true,
	)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"Datastore"},
		[]string{
			"parent",
			"summary",
			"name",
			"info",
		},
		&items,
	)
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}

	for _, i := range items {
		var kind string
		func() {
			iInfo := reflect.ValueOf(i.Info).Elem().Interface()
			switch pInfo := iInfo.(type) {
			case types.LocalDatastoreInfo:
				kind = "local"
			case types.VmfsDatastoreInfo:
				if pInfo.Vmfs != nil {
					var ssd string
					if *pInfo.Vmfs.Ssd {
						ssd = "ssd"
					} else {
						ssd = "hdd"
					}
					var local string
					if *pInfo.Vmfs.Local {
						local = "local"
					} else {
						local = "non-local"
					}
					kind = fmt.Sprintf("vmfs-%s-%s", local, ssd)
				}
			case types.NasDatastoreInfo:
				kind = "nas"
			case types.PMemDatastoreInfo:
				kind = "pmem"
			case types.VsanDatastoreInfo:
				kind = "vsan"
			case types.VvolDatastoreInfo:
				kind = "vvol"
			}
		}()
		fmt.Printf("kind: %s", kind)
	}

}

func TestVMwareDatacenter(t *testing.T) {

	var datacenters []mo.Datacenter

	client := GetClient(t)
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"Datacenter"},
		true,
	)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"Datacenter"},
		[]string{
			// "parent",
			// "summary",
			// "name",
		},
		&datacenters,
	)
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}

}

func TestVMwareResourcePool(t *testing.T) {

	var items []mo.ResourcePool

	client := GetClient(t)
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"ResourcePool"},
		true,
	)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"ResourcePool"},
		[]string{
			"parent",
			"summary",
			"name",
		},
		&items,
	)
	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}

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
		t.Logf("VM:  %s", vm.Config.Name)
	}
}

func TestVMwareVMperHost(t *testing.T) {
	host := types.ManagedObjectReference{
		Type:  "HostSystem",
		Value: "host-78",
	}

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

	err = v.RetrieveWithFilter(
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
		&items, property.Match{"Runtime.Host": host},
	)

	if err != nil {
		t.Errorf("%v", err)
		t.Fail()
	}
	for _, vm := range items {
		t.Logf("host:  %s", vm.Config.Name)
	}
}

func TestVMwareVMperCluster(t *testing.T) {
	cluster := types.ManagedObjectReference{
		Type:  "ClusterComputeResource",
		Value: "domain-c82",
	}

	var items []mo.VirtualMachine
	client := GetClient(t)
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		cluster,
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

func TestVMwareVMperHost2(t *testing.T) {
	host := types.ManagedObjectReference{
		Type:  "HostSystem",
		Value: "host-78",
	}
	// host := types.ManagedObjectReference{
	// 	Type:  "HostSystem",
	// 	Value: "host-101",
	// }

	var items []mo.VirtualMachine
	client := GetClient(t)
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		host,
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
// 	activeCache := collector.NewVCenterScraper(collector.VMwareConfig{
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
