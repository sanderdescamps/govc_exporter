package collector

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/soap"
)

const getClientRetryTime = 3 * time.Second

type Pool chan *govmomi.Client

// type VMwareClientPool struct {
// 	config  map[string]string
// 	clients map[string]*govmomi.Client
// }

type ClientConf struct {
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewVMwareClientPool(ctx context.Context, size int, conf ClientConf) Pool {
	p := make(Pool, size)
	wg := new(sync.WaitGroup)
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func(clientId int) {
			client, err := NewVMwareClient(ctx, conf)
			if err != nil {
				slog.Error("Failed to start client", "err", err)
			}
			p <- client
			wg.Done()
		}(i)
	}
	wg.Wait()
	return p
}

func (p Pool) GetClient() (r *govmomi.Client) {
	select {
	case r := <-p:
		return r
	case <-time.After(5 * time.Minute):
		panic("Failed to get govmomi.Client from pool")
	}
}

func NewVMwareClient(ctx context.Context, config ClientConf) (*govmomi.Client, error) {
	// logger := sh.ctx.Value(SensorLoggerCtxKey).(*slog.Logger)
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

// func Logout() {
// 	logger := sh.ctx.Value(SensorLoggerCtxKey).(*slog.Logger)
// 	err := sh.client.Logout(sh.ctx)
// 	if err != nil {
// 		logger.LogAttrs(sh.ctx, slog.LevelError, "logout error", slog.String("err", err.Error()))
// 	}
// 	sh.ctx.Done()
// }
