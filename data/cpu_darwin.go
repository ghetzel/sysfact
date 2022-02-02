package data

import (
	"strings"

	"github.com/ghetzel/go-stockutil/mathutil"
	"github.com/ghetzel/go-stockutil/rxutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type CPU struct {
	lastStats *CPUStats
}

func (self CPU) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	var ncpu = shell(`sysctl -n hw.logicalcpu`).NInt()
	var pcpu = shell(`sysctl -n hw.packages`).NInt()

	var model = shell(`sysctl -n machdep.cpu.brand_string`).String()
	model = strings.TrimSpace(model)
	// model = strings.TrimSuffix(model, ` CPU`)
	// speed = strings.TrimSpace(speed)
	// speed = strings.ToLower(speed)
	// speed = strings.TrimSuffix(speed, `ghz`)

	for _, line := range lines("top -f -l 1 -ncols 3") {
		var linelc = strings.ToLower(line)

		if strings.HasPrefix(linelc, "cpu usage: ") {
			var _, tail = stringutil.SplitPairTrimSpace(line, `: `)

			if tail != `` {
				var user float64
				var sys float64
				var idle float64

				for _, part := range rxutil.Split(",\\s+", tail) {
					var val, key = stringutil.SplitPairTrimSpace(part, "%")
					val = strings.TrimSpace(val)
					var num = mathutil.RoundPlaces(typeutil.Float(val), 1)

					switch key {
					case `user`:
						user = num
						out[`cpu.usage.user`] = user
					case `sys`, `system`:
						sys = num
						out[`cpu.usage.system`] = sys
					case `idle`:
						idle = num
						out[`cpu.usage.idle`] = idle
					}
				}

				out[`cpu.usage.active`] = mathutil.RoundPlaces(100-idle, 1)
			}
		}
	}

	out[`cpu.count`] = ncpu
	out[`cpu.physical`] = pcpu
	out[`cpu.model`] = model

	return out
}
