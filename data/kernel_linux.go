package data

import (
	"strings"
	"time"

	"github.com/ghetzel/go-stockutil/fileutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type Kernel struct {
}

func (self Kernel) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	out[`arch`] = shell(`uname -i`).String()

	if line, err := fileutil.ReadFirstLine(`/proc/uptime`); err == nil {
		seconds, _ := stringutil.SplitPair(strings.TrimSpace(line), ` `)
		sec := (time.Second * time.Duration(typeutil.Float(seconds))).Round(time.Second)

		bootedAt := time.Now().Add(-1 * sec).Round(time.Second)
		out[`booted_at`] = bootedAt
		out[`uptime`] = sec
	}

	out[`kernel.version`] = shell(`uname -r`).String()
	out[`kernel.hostname`] = shell(`uname -n`).String()

	return out
}
