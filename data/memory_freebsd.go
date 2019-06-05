package data

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/mathutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

var freebsdSysctlOutMap = map[string]string{
	`vm.stats.vm.v_page_size`:      `_pagesize`,
	`hw.realmem`:                   `memory.total`,
	`vm.swap_total`:                `memory.swap`,
	`vm.stats.vm.v_free_count`:     `memory.free`,
	`vm.stats.vm.v_wire_count`:     `memory.wired`,
	`vm.stats.vm.v_active_count`:   `memory.active`,
	`vm.stats.vm.v_inactive_count`: `memory.inactive`,
	`vm.stats.vm.v_cache_count`:    `memory.cached`,
}

type Memory struct {
}

func (self Memory) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	var pagesize int64

	keys := maputil.StringKeys(freebsdSysctlOutMap)
	sort.Slice(keys, func(i int, j int) bool {
		iV := freebsdSysctlOutMap[keys[i]]
		jV := freebsdSysctlOutMap[keys[j]]

		return (iV < jV)
	})

	fmt.Println(keys)
	values := lines(`sysctl -n ` + strings.Join(keys, ` `))

	for i, line := range values {
		value := typeutil.Int(strings.TrimSpace(line))

		if i < len(keys) {
			syskey := keys[i]

			if outkey, ok := freebsdSysctlOutMap[syskey]; ok {
				multiply := int64(1)

				if strings.HasSuffix(syskey, `_count`) {
					multiply = pagesize
				}

				switch outkey {
				case `_pagesize`:
					pagesize = value
				case `memory.wired`:
					out[outkey] = value
				default:
					out[outkey] = multiply * value
				}
			}
		}
	}

	total := typeutil.V(out[`memory.total`]).Int()
	used := typeutil.V(out[`memory.active`]).Int() + typeutil.V(out[`memory.wired`]).Int()

	out[`memory.available`] = total - used
	out[`memory.used`] = used
	out[`memory.percent_used`] = mathutil.RoundPlaces(float64(used)/float64(total)*100.0, 2)

	dmidecodeMemory(out)

	return out
}
