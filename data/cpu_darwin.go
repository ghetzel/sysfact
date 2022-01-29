package data

import (
	"strings"
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

	out[`cpu.count`] = ncpu
	out[`cpu.physical`] = pcpu
	out[`cpu.model`] = model

	return out
}
