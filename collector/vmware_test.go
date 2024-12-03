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
