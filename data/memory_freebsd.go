package data

import (
	"github.com/ghetzel/go-stockutil/mathutil"
)

type Memory struct {
}

func (self Memory) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	var pgsz, total, swap, wired, active, inactive, free, cache int64

	pgsz = shellfl(`sysctl -n vm.stats.vm.v_page_size`).Int()
	total = shellfl(`sysctl -n hw.realmem`).Int()
	swap = shellfl(`sysctl -n vm.swap_total`).Int()
	wired = shellfl(`sysctl -n vm.stats.vm.v_wire_count`).Int()
	active = pgsz * shellfl(`sysctl -n vm.stats.vm.v_active_count`).Int()
	inactive = pgsz * shellfl(`sysctl -n vm.stats.vm.v_inactive_count`).Int()
	free = pgsz * shellfl(`sysctl -n vm.stats.vm.v_free_count`).Int()
	cache = pgsz * shellfl(`sysctl -n vm.stats.vm.v_cache_count`).Int()
	used := active + wired

	out[`memory.total`] = total
	out[`memory.free`] = free
	out[`memory.available`] = total - used
	out[`memory.active`] = active
	out[`memory.inactive`] = inactive
	out[`memory.swap`] = swap
	out[`memory.used`] = used
	out[`memory.percent_used`] = mathutil.RoundPlaces(float64(used)/float64(total)*100.0, 2)

	dmidecodeMemory(out)

	return out
}
