package collector

import (
	"fmt"

	"github.com/vmware/govmomi/vim25/mo"
)

func b2f(val bool) float64 {
	if val {
		return 1.0
	}
	return 0.0
}

func me2id(me mo.ManagedEntity) string {
	return fmt.Sprintf("%s", me.Self.Value)
}
