package collector

import (
	"strings"

	"github.com/vmware/govmomi/vim25/mo"
)

func b2f(val bool) float64 {
	if val {
		return 1.0
	}
	return 0.0
}

func me2id(me mo.ManagedEntity) string {
	return me.Self.Value
}

func cleanString(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}
