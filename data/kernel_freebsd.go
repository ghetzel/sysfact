package data

import (
	"time"

	"go.gary.cool/go-stockutil/rxutil"
	"go.gary.cool/go-stockutil/typeutil"
)

type Kernel struct {
}

func (self Kernel) Collect() map[string]any {
	out := make(map[string]any)

	if m := rxutil.Match(
		`sec\s*=\s*(?P<epoch>\d+)`,
		shell(`sysctl -n kern.boottime`).String(),
	); m != nil {
		uptime := time.Since(
			time.Unix(typeutil.Int(m.Group(`epoch`)), 0),
		).Round(time.Second)

		out[`uptime`] = (uptime / time.Second)
		out[`uptime_readable`] = uptime.String()
		out[`booted_at`] = time.Now().Add(-1 * uptime).Round(time.Second)
	}

	out[`kernel.version`] = shell(`uname -K`).String()
	out[`kernel.hostname`] = shell(`uname -n`).String()
	out[`arch`] = shell(`uname -m`).String()

	return out
}
