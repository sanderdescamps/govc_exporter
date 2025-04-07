package pool

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/soap"
)

func NewVCenterClient(endpoint string, username string, password string) (*govmomi.Client, error) {
	u, err := soap.ParseURL(endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %s", endpoint)
	}
	ctx := context.Background()
	u.User = url.UserPassword(username, password)
	client, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		return nil, fmt.Errorf("failed to login")
	}
	return client, nil
}
