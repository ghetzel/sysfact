package data

import (
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/ghetzel/go-stockutil/fileutil"
	"github.com/ghetzel/go-stockutil/mathutil"
	"github.com/ghetzel/go-stockutil/stringutil"
)

type Memory struct {
}

func (self Memory) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	// basic OS memory info
	if meminfo, err := fileutil.ReadAllLines(`/proc/meminfo`); err == nil {
		var total, free, buffers, cache uint64

		for _, line := range meminfo {
			line = strings.TrimSpace(line)
			k, v := stringutil.SplitPair(line, ` `)
			k = strings.TrimSpace(k)
			k = strings.ToLower(k)
			v = strings.TrimSpace(v)

			if bytes, err := humanize.ParseBytes(v); err == nil {
				switch k {
				case `memtotal:`:
					out[`memory.total`] = bytes
					total = bytes
				case `memfree:`:
					out[`memory.free`] = bytes
					free = bytes
				case `memavailable:`:
					out[`memory.available`] = bytes
				case `buffers:`:
					out[`memory.buffers`] = bytes
					buffers = bytes
				case `cached:`:
					out[`memory.cached`] = bytes
					cache = bytes
				case `active:`:
					out[`memory.active`] = bytes
				case `inactive:`:
					out[`memory.inactive`] = bytes
				case `swaptotal:`:
					out[`memory.swap`] = bytes
				}
			}
		}

		if total > 0 {
			// as described in free(1)
			used := total - free - buffers - cache

			out[`memory.used`] = used
			out[`memory.percent_used`] = mathutil.RoundPlaces(float64(used)/float64(total)*100.0, 2)
		}
	}

	dmidecodeMemory(out)

	return out
}
