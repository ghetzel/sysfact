package data

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ghetzel/go-stockutil/mathutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type CPU struct {
}

func (self CPU) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	if d, err := ioutil.ReadFile(`/proc/cpuinfo`); err == nil {
		logical := 0
		physical := 0
		physIdSeen := make(map[string]bool)

		for _, line := range strings.Split(strings.TrimSpace(string(d)), "\n") {
			key, valueS := stringutil.SplitPair(line, `:`)
			key = strings.ToLower(strings.TrimSpace(key))
			valueS = strings.TrimSpace(valueS)
			value := typeutil.V(normalize(valueS))

			switch key {
			case `processor`:
				logical += 1
			case `physical id`:
				if _, ok := physIdSeen[valueS]; !ok {
					physical += 1
					physIdSeen[valueS] = true
				}
			}

			if index := (logical - 1); index >= 0 {
				switch key {
				case `cpu mhz`:
					out[fmt.Sprintf("cpu.cores.%d.current_speed", index)] = mathutil.Round(value.Float())
				case `model name`:
					model, speed := stringutil.SplitPair(value.String(), `@`)
					model = strings.TrimSpace(model)
					model = strings.TrimSuffix(model, ` CPU`)

					speed = strings.TrimSpace(speed)
					speed = strings.ToLower(speed)
					speed = strings.TrimSuffix(speed, `ghz`)

					out[fmt.Sprintf("cpu.cores.%d.model", index)] = model

					// NOTE: assumes the bit after @ is in GHz
					out[fmt.Sprintf("cpu.cores.%d.speed", index)] = (typeutil.V(speed).Float() * 1000)
				case `flags`:
					for i, flag := range strings.Split(value.String(), ` `) {
						out[fmt.Sprintf("cpu.cores.%d.flags.%d", index, i)] = flag
					}
				case `bugs`:
					for i, flag := range strings.Split(value.String(), ` `) {
						out[fmt.Sprintf("cpu.cores.%d.bugs.%d", index, i)] = flag
					}
				}
			}
		}

		out[`cpu.count`] = logical
		out[`cpu.physical`] = physical
	}

	return out
}
