package pool

import (
	"context"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vapi/rest"
)

type VCenterPool interface {
	Init() error
	Reauthenticate() error
	Acquire() (client *govmomi.Client, release func(), err error)
	AcquireWithContext(ctx context.Context) (*govmomi.Client, func(), error)
	AcquireRest() (client *rest.Client, release func(), err error)
	// StartAuthRefresher(context.Context, time.Duration) (stop func())
	Destroy(ctx context.Context)
}
