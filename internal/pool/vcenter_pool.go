package pool

import (
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vapi/rest"
)

type VCenterPool interface {
	Init() error
	Reauthenticate() error
	Acquire() (client *govmomi.Client, release func(), err error)
	AcquireRest() (client *rest.Client, release func(), err error)
	// StartAuthRefresher(context.Context, time.Duration) (stop func())
	Destroy()
}
