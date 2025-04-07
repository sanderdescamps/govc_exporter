package pool

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vapi/rest"
)

type VCenterThrottlePool struct {
	VCenterPool
	// VCenterStatusChecker
	ThrottlerPool[govmomi.Client]

	endpoint         string
	username         string
	password         string
	clientReAuthLock sync.Mutex
}

func NewVCenterThrottlePool(endpoint string, username string, password string, poolSize int) *VCenterThrottlePool {
	return &VCenterThrottlePool{
		endpoint:      endpoint,
		username:      username,
		password:      password,
		ThrottlerPool: *NewThrottlerPool[govmomi.Client](nil, poolSize),
		// VCenterStatusChecker: *NewVCenterStatusChecker(endpoint, username, password),
	}
}

func (p *VCenterThrottlePool) Init() error {
	return p.Reauthenticate()
}

var (
	ErrReauthenticateAlreadyInProgress = fmt.Errorf("reauthentication already in progress")
)

func (p *VCenterThrottlePool) Reauthenticate() error {
	if hasLock := p.clientReAuthLock.TryLock(); !hasLock {
		return ErrReauthenticateAlreadyInProgress
	}
	defer p.clientReAuthLock.Unlock()

	client, release, err := p.ThrottlerPool.Drain()
	if err != nil {
		return fmt.Errorf("failed to drain client pool: %w", err)
	}
	defer release()

	if client == nil || client.Client == nil {
		client, err := NewVCenterClient(p.endpoint, p.username, p.password)
		if err != nil {
			return err
		}
		p.poolObject = client
	} else {
		ctx := context.Background()
		session, err := p.poolObject.SessionManager.UserSession(ctx)
		if err != nil {
			return err
		} else if session == nil {
			userInfo := url.UserPassword(p.username, p.password)
			err := p.poolObject.Login(ctx, userInfo)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *VCenterThrottlePool) Acquire() (*govmomi.Client, func(), error) {
	var authErr error
	for retry := 0; retry < 2; retry++ {
		client, release, err := p.ThrottlerPool.Acquire()
		if err != nil {
			return nil, release, err
		}
		ctx := context.Background()
		if session, sessionErr := client.SessionManager.UserSession(ctx); session == nil {
			release()
			authErr = p.Reauthenticate()
			if errors.Is(authErr, ErrReauthenticateAlreadyInProgress) {
				time.Sleep(time.Second * 5)
				continue
			} else if authErr != nil {
				break
			}
		} else {
			return client, release, sessionErr
		}
	}
	return nil, func() {}, fmt.Errorf("failed to reauthenticate: %w", authErr)
}

func (p *VCenterThrottlePool) AcquireRest() (*rest.Client, func(), error) {
	client, release, err := p.Acquire()
	if err != nil {
		return nil, release, err
	}

	restClient := rest.NewClient(client.Client)
	ctx := context.Background()
	userInfo := url.UserPassword(p.username, p.password)
	if err := restClient.Login(ctx, userInfo); err != nil {
		return nil, release, fmt.Errorf("failed to login rest client: %w", err)
	}
	return restClient, release, nil
}

func (p *VCenterThrottlePool) Destroy() {
	if p.poolObject != nil {
		p.poolObject.Logout(context.Background())
	}
}
