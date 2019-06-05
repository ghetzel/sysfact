package data

import (
	"strings"

	"github.com/ghetzel/go-stockutil/mathutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type Memory struct {
}

func (self Memory) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	var total, swap, active, inactive, free, pgsz int64

	total = shellfl(`sysctl -n hw.realmem`).Int()
	swap = shellfl(`sysctl -n vm.swap_total`).Int()

	for _, line := range shell(`vmstat -s`).Split("\n") {
		line = strings.TrimSpace(line)

		v, k := stringutil.SplitPair(line, ` `)
		k = stringutil.SqueezeSpace(k)
		vi := typeutil.Int(v)

		switch k {
		case `pages free`:
			free = vi
		case `pages active`:
			active = vi
		case `pages inactive`:
			inactive = vi
		case `bytes per page`:
			pgsz = vi
		}
	}

	total *= pgsz
	active *= pgsz
	inactive *= pgsz
	free *= pgsz
	used := total - active

	out[`memory.total`] = total
	out[`memory.free`] = free
	out[`memory.available`] = total - free
	out[`memory.active`] = active
	out[`memory.inactive`] = inactive
	out[`memory.swap`] = swap
	out[`memory.used`] = used
	out[`memory.percent_used`] = mathutil.RoundPlaces(float64(used)/float64(total)*100.0, 2)

	dmidecodeMemory(out)

	return out
}
