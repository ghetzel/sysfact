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
		sec := (time.Second * time.Duration(typeutil.Float(seconds)))

		bootedAt := time.Now().Add(-1 * sec)
		out[`booted_at`] = bootedAt.Format(time.RFC1123Z)
		out[`uptime`] = sec
	}

	out[`kernel.version`] = shell(`uname -r`).String()
	out[`kernel.hostname`] = shell(`uname -n`).String()

	return out
}
