package pool

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sync"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25/soap"
)

var ErrReInitAlreadyRunning = errors.New("Client reinitialisation on going")

type VCenterClientPool struct {
	MultiAccessPool[govmomi.Client]
	clientGenerator func() (*govmomi.Client, error)
	reInitLock      sync.Mutex
	blockLock       sync.Mutex
	userInfo        func() *url.Userinfo
}

func NewVCenterClient(ctx context.Context, endpoint string, username string, password string) (*govmomi.Client, error) {
	u, err := soap.ParseURL(endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %s", endpoint)
	}
	u.User = url.UserPassword(username, password)
	client, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		return nil, fmt.Errorf("failed to login")
	}
	return client, nil
}

func NewVCenterClientFuncWithLogger(ctx context.Context, endpoint string, username string, password string, logger *slog.Logger) func() (*govmomi.Client, error) {
	return func() (*govmomi.Client, error) {
		client, err := NewVCenterClient(ctx, endpoint, username, password)
		if err != nil {
			if logger != nil {
				logger.Error("failed to create client", "err", err)
			}
			return nil, err
		}
		return client, nil
	}
}

func NewVCenterClientFunc(ctx context.Context, endpoint string, username string, password string) func() (*govmomi.Client, error) {
	return func() (*govmomi.Client, error) {
		client, err := NewVCenterClient(ctx, endpoint, username, password)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}

func NewVCenterClientPoolWithLogger(ctx context.Context, endpoint string, username string, password string, size int, logger *slog.Logger) (*VCenterClientPool, error) {
	userInfo := func() *url.Userinfo {
		return url.UserPassword(username, password)
	}

	clientFunc := NewVCenterClientFuncWithLogger(ctx, endpoint, username, password, logger)
	client, err := clientFunc()
	if err != nil {
		return nil, err
	}

	pool := &VCenterClientPool{
		MultiAccessPool: *NewMultiAccessPool(client, size, nil),
		userInfo:        userInfo,
		clientGenerator: clientFunc,
	}

	return pool, nil
}

func (p *VCenterClientPool) WaitForBlock() (unblock func()) {
	p.blockLock.Lock()
	defer p.blockLock.Unlock()
	releaseFunc := []func(){}
	for i := 0; i < p.max; i++ {
		_, release, _ := p.MultiAccessPool.Acquire()
		releaseFunc = append(releaseFunc, release)
	}
	return func() {
		for _, release := range releaseFunc {
			release()
		}
	}
}

func (p *VCenterClientPool) WaitForUnblock() {
	p.blockLock.Lock()
	defer p.blockLock.Unlock()
}

func (p *VCenterClientPool) reInit() error {
	if ok := p.reInitLock.TryLock(); ok {
		defer p.reInitLock.Unlock()

		release := p.WaitForBlock()
		defer release()

		ctx := context.Background()
		err := p.poolObject.Login(ctx, p.userInfo())
		if err != nil {
			return err
		}
		return nil
	}
	return ErrReInitAlreadyRunning
}

func (p *VCenterClientPool) Acquire() (*govmomi.Client, func(), error) {
	client, release, err := p.MultiAccessPool.Acquire()
	ctx := context.Background()
	if session, _ := client.SessionManager.UserSession(ctx); session == nil {
		fmt.Println("client not valid")
		release()
		err := p.reInit()
		if err != nil {
			return nil, nil, err
		}
		return p.Acquire()
	}
	return client, release, err
}

func (p *VCenterClientPool) AcquireRest() (*rest.Client, func(), error) {
	client, release, err := p.Acquire()
	if err != nil {
		return nil, release, err
	}

	restClient := rest.NewClient(client.Client)
	ctx := context.Background()
	if s, err := restClient.Session(ctx); s == nil {
		restClient.Login(ctx, p.userInfo())
	} else if err != nil {
		return nil, release, err
	}

	return restClient, release, nil
}
