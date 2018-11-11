package data

import (
	"time"

	"github.com/ghetzel/go-stockutil/rxutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type Kernel struct {
}

func (self Kernel) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	if m := rxutil.Match(
		`sec\s*=\s*(?P<epoch>\d+)`,
		shell(`sysctl -n kern.boottime`).String(),
	); m != nil {
		uptime := time.Since(
			time.Unix(typeutil.Int(m.Group(`epoch`)), 0),
		).Round(time.Second)

		out[`uptime`] = uptime
		out[`booted_at`] = time.Now().Add(-1 * uptime).Round(time.Second)
	}

	out[`kernel.version`] = shell(`uname -K`).String()
	out[`kernel.hostname`] = shell(`uname -n`).String()
	out[`arch`] = shell(`uname -m`).String()

	return out
}
