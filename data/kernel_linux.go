package data

import (
	"strings"
	"time"

	"go.gary.cool/go-stockutil/fileutil"
	"go.gary.cool/go-stockutil/stringutil"
	"go.gary.cool/go-stockutil/typeutil"
)

type Kernel struct {
}

func (self Kernel) Collect() map[string]any {
	out := make(map[string]any)

	out[`arch`] = shell(`uname -i`).String()

	if line, err := fileutil.ReadFirstLine(`/proc/uptime`); err == nil {
		seconds, _ := stringutil.SplitPair(strings.TrimSpace(line), ` `)
		sec := (time.Second * time.Duration(typeutil.Float(seconds))).Round(time.Second)

		bootedAt := time.Now().Add(-1 * sec).Round(time.Second)
		out[`booted_at`] = bootedAt
		out[`uptime`] = (sec / time.Second)
		out[`uptime_readable`] = sec.String()
	}

	out[`kernel.version`] = shell(`uname -r`).String()
	out[`kernel.hostname`] = shell(`uname -n`).String()

	return out
}
