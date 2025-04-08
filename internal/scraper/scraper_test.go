package scraper_test

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/prometheus/common/promslog"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

func GetClient(ctx context.Context, t *testing.T) *govmomi.Client {
	endpoint := "https://localhost:8989"
	username := "testuser"
	password := "testpass"

	u, err := soap.ParseURL(endpoint)
	if err != nil {
		panic(err)
	}
	u.User = url.UserPassword(username, password)
	client, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		t.Fatalf("Client creation failed: %v", err)
	}
	return client
}

func TestVCenterScraper(t *testing.T) {

	conf := scraper.DefaultConfig()
	conf.VCenter = "https://localhost:8989"
	conf.Username = "testuser"
	conf.Password = "testpass"
	conf.Tags.CategoryToCollect = []string{"tenants"}

	err := conf.Validate()
	if err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	promlogConfig := &promslog.Config{
		// Level:
	}
	logger := promslog.New(promlogConfig)
	ctx := context.Background()
	ctxScraper := context.WithValue(ctx, scraper.ContextKeyScraperLogger{}, logger)

	aCache, _ := scraper.NewVCenterScraper(ctxScraper, conf)

	t1 := time.Now()
	aCache.Host.Refresh(ctxScraper)
	t2 := time.Now()
	fmt.Printf("fetching all hosts took %dms\n", t2.Sub(t1).Milliseconds())

	aCache.Datastore.Refresh(ctxScraper)
	t3 := time.Now()
	fmt.Printf("fetching all datastores took %dms\n", t3.Sub(t2).Milliseconds())

	aCache.SPOD.Refresh(ctxScraper)
	t4 := time.Now()
	fmt.Printf("fetching all spod took %dms\n", t4.Sub(t3).Milliseconds())

	aCache.VM.Refresh(ctxScraper)
	t5 := time.Now()
	fmt.Printf("fetching all VMs took %dms\n", t5.Sub(t4).Milliseconds())

	aCache.Cluster.Refresh(ctxScraper)
	t6 := time.Now()
	fmt.Printf("fetching all clusters took %dms\n", t6.Sub(t5).Milliseconds())

	aCache.Tags.Refresh(ctxScraper)
	t7 := time.Now()
	fmt.Printf("fetching all tags took %dms\n", t7.Sub(t6).Milliseconds())
	// for _, cat := range conf.TagsCategoryToCollect {
	// 	catID := aCache.Tags.GetCategoryID(cat)
	// 	for _, ref := range aCache.Tags.GetAllRefs() {
	// 		tags := aCache.Tags.Get(ref)
	// 		tagString := []string{}
	// 		for _, tag := range tags {
	// 			tagString = append(tagString)
	// 		}
	// 		fmt.Printf("tags for %s\n", ref)
	// 	}
	// }

	fmt.Print("end")
}

func TestVCenterScraperStart(t *testing.T) {

	conf := scraper.DefaultConfig()

	conf.VCenter = "https://localhost:8989"
	conf.Username = "testuser"
	conf.Password = "testpass"
	conf.Tags.CategoryToCollect = []string{"tenants"}

	err := conf.Validate()
	if err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	promlogConfig := &promslog.Config{
		// Level:
	}
	logger := promslog.New(promlogConfig)
	ctx := context.Background()
	ctxScraper := context.WithValue(ctx, scraper.ContextKeyScraperLogger{}, logger)

	aCache, _ := scraper.NewVCenterScraper(ctxScraper, conf)

	sensors := aCache.SensorList()
	for _, sensor := range sensors {
		logger.Info("Sensor", "name", sensor.Name(), "kind", sensor.Kind())
	}
	aCache.Start(ctxScraper)

}

func TestVMwareHost(t *testing.T) {

	var hss []mo.HostSystem

	ctx := context.Background()
	client := GetClient(ctx, t)
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"InventoryServiceCategory"},
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
			"config.storageDevice",
			"config.network.vswitch",
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

func TestVMwareVMTags(t *testing.T) {

	catsToObserve := []string{"tenants"}
	ctx := context.Background()
	client := GetClient(ctx, t)

	re := rest.NewClient(client.Client)
	_, err := re.Session(ctx)
	if err != nil {
		log.Print(err)
		t.Fatal(err.Error())
	}
	username := "testuser"
	password := "testpass"
	userPass := url.UserPassword(username, password)
	re.Login(ctx, userPass)
	m := tags.NewManager(re)

	catList := []tags.Category{}
	allCats, err := m.GetCategories(ctx)
	if err != nil {
		log.Print(err)
		t.Fatal(err.Error())
	}
	for _, cat := range allCats {
		if len(catsToObserve) == 0 || slices.Contains(catsToObserve, cat.Name) {
			catList = append(catList, cat)
			func() {
				//store cats in cache
			}()
		}
	}

	tagList := []tags.Tag{}
	for _, cat := range catList {
		tags, err := m.GetTagsForCategory(ctx, cat.ID)
		if err != nil {
			log.Print(err)
			t.Fatal(err.Error())
		}
		tagList = append(tagList, tags...)
	}

	objectTags := make(map[types.ManagedObjectReference][]*tags.Tag)
	for _, tag := range tagList {
		attachObjs, err := m.GetAttachedObjectsOnTags(ctx, []string{tag.ID})
		if err != nil {
			t.Fatal(err.Error())
		}

		// tag2, err := m.GetTag(ctx, tag)
		for _, attachObj := range attachObjs {
			for _, elem := range attachObj.ObjectIDs {
				objectTags[elem.Reference()] = append(objectTags[elem.Reference()], attachObj.Tag)
			}
		}
	}

	for name, tags := range objectTags {
		fmt.Printf(" %s", name)
		for _, tag := range tags {
			fmt.Printf("   name: %s", tag.Name)
			fmt.Printf("   Description: %s", tag.Description)
		}
	}
}

func TestVMwareNetwork(t *testing.T) {

	var items []mo.Network

	ctx := context.Background()
	client := GetClient(ctx, t)
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

	ctx := context.Background()
	client := GetClient(ctx, t)
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

	ctx := context.Background()
	client := GetClient(ctx, t)
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

	ctx := context.Background()
	client := GetClient(ctx, t)
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

	ctx := context.Background()
	client := GetClient(ctx, t)
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

	ctx := context.Background()
	client := GetClient(ctx, t)
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

	ctx := context.Background()
	client := GetClient(ctx, t)
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
	ctx := context.Background()
	client := GetClient(ctx, t)
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
	ctx := context.Background()
	client := GetClient(ctx, t)
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
	ctx := context.Background()
	client := GetClient(ctx, t)
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
	ctx := context.Background()
	client := GetClient(ctx, t)
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

func TestHostPerf(t *testing.T) {

	conf := scraper.DefaultConfig()
	conf.VCenter = "https://localhost:8989"
	conf.Username = "testuser"
	conf.Password = "testpass"
	conf.Tags.CategoryToCollect = []string{"tenants"}

	err := conf.Validate()
	if err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	promlogConfig := &promslog.Config{
		// Level:
	}
	logger := promslog.New(promlogConfig)
	ctx := context.Background()
	ctxScraper := context.WithValue(ctx, scraper.ContextKeyScraperLogger{}, logger)

	aCache, _ := scraper.NewVCenterScraper(ctxScraper, conf)
	aCache.Host.Refresh(ctxScraper)

	t1 := time.Now()
	err = aCache.HostPerf.Refresh(ctxScraper)
	t2 := time.Now()

	fmt.Printf("fetching all hosts took %dms\n", t2.Sub(t1).Milliseconds())
	if err != nil {
		logger.Error("error fetching metrics", "err", err)
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
