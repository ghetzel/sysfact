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

	var pgsz, total, swap, wired, active, inactive, free int64

	pgsz = shellfl(`sysctl -n hw.pagesize`).Int()
	total = shellfl(`sysctl -n hw.memsize`).Int()

	// $ vm_stat
	// Pages free:                                4928.
	// Pages active:                            220643.
	// Pages inactive:                          218767.
	// Pages speculative:                         1269.
	// Pages throttled:                              0.
	// Pages wired down:                        123604.
	// Pages purgeable:                            431.
	// "Translation faults":               10066059740.
	// Pages copy-on-write:                 1630129245.
	// Pages zero filled:                   2184826075.
	// Pages reactivated:                     68128443.
	// Pages purged:                          15533628.
	// File-backed pages:                       197475.
	// Anonymous pages:                         243204.
	// Pages stored in compressor:             1001112.
	// Pages occupied by compressor:            445946.
	// Decompressions:                        73242998.
	// Compressions:                          82751998.
	// Pageins:                              262496770.
	// Pageouts:                                195429.
	// Swapins:                                1184380.
	// Swapouts:                               1308281.
	//
	for _, line := range lines(`vm_stat`) {
		line = strings.TrimSuffix(line, `.`)

		if strings.TrimSpace(line) == `` {
			continue
		}

		var k, v = stringutil.SplitPairTrimSpace(line, `: `)

		k = strings.ToLower(k)
		k = strings.ReplaceAll(k, `"`, ``)
		k = stringutil.Hyphenate(k)

		var value int64 = typeutil.Int(v)

		switch k {
		case `pages-free`:
			free += pgsz * value
		case `pages-active`:
			active = pgsz * value
		case `pages-inactive`:
			inactive = pgsz * value
		case `pages-speculative`:
			free += (pgsz * value)
		case `pages-wired-down`:
			wired = pgsz * value
		}
	}

	var used = active + wired

	out[`memory.total`] = total
	out[`memory.free`] = free
	out[`memory.available`] = total - used
	out[`memory.active`] = active
	out[`memory.inactive`] = inactive
	out[`memory.swap`] = swap
	// out[`memory.cached`] = cache
	out[`memory.used`] = used
	out[`memory.percent_used`] = mathutil.RoundPlaces(float64(used)/float64(total)*100.0, 2)

	return out
}
