package scraper

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/soap"
)

type ClientConf struct {
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewVMwareClient(ctx context.Context, config ClientConf) (*govmomi.Client, error) {
	u, err := soap.ParseURL(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %s", config.Endpoint)
	}
	u.User = url.UserPassword(config.Username, config.Password)
	client, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		return nil, fmt.Errorf("failed to login")
	}
	return client, nil
}
