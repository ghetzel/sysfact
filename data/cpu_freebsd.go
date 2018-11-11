package data

import (
	"fmt"
	"strings"

	"github.com/ghetzel/go-stockutil/mathutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type CPU struct {
}

func (self CPU) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	ncpu := int(shell(`sysctl -n hw.ncpu`).Int())
	model, speed := stringutil.SplitPair(shell(`sysctl -n hw.model`).String(), `@`)
	model = strings.TrimSpace(model)
	model = strings.TrimSuffix(model, ` CPU`)

	speed = strings.TrimSpace(speed)
	speed = strings.ToLower(speed)
	speed = strings.TrimSuffix(speed, `ghz`)

	out[`cpu.count`] = ncpu

	for i := 0; i < ncpu; i++ {
		out[fmt.Sprintf("cpu.cores.%d.model", i)] = model
		out[fmt.Sprintf("cpu.cores.%d.speed", i)] = mathutil.Round(typeutil.Float(speed) * 1000)
		out[fmt.Sprintf("cpu.cores.%d.temperature", i)] = typeutil.Float(strings.TrimSuffix(
			shell("sysctl -n dev.cpu.%d.temperature", i).String(),
			`C`,
		))

		out[fmt.Sprintf("cpu.cores.%d.max_temperature", i)] = typeutil.Float(strings.TrimSuffix(
			shell("sysctl -n dev.cpu.%d.coretemp.tjmax", i).String(),
			`C`,
		))
	}

	return out
}
