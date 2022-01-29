package data

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/ghetzel/go-stockutil/fileutil"
	"github.com/ghetzel/go-stockutil/mathutil"
	"github.com/ghetzel/go-stockutil/rxutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type CPU struct {
	lastStats *CPUStats
}

func (self CPU) Collect() map[string]interface{} {
	var out = make(map[string]interface{})

	if d, err := ioutil.ReadFile(`/proc/cpuinfo`); err == nil {
		var logical int
		var physical int
		var physIdSeen = make(map[string]bool)

		for _, line := range strings.Split(strings.TrimSpace(string(d)), "\n") {
			var key, valueS = stringutil.SplitPair(line, `:`)
			key = strings.ToLower(strings.TrimSpace(key))
			valueS = strings.TrimSpace(valueS)
			var value = typeutil.V(normalize(valueS))

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
					var model, speed = stringutil.SplitPair(value.String(), `@`)
					model = strings.TrimSpace(model)
					model = strings.TrimSuffix(model, ` CPU`)

					speed = strings.TrimSpace(speed)
					speed = strings.ToLower(speed)
					speed = strings.TrimSuffix(speed, `ghz`)

					out[`cpu.model`] = model
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

		var ok bool
		var i int

		self.lastStats = nil

		for iter := 0; iter < 2; iter++ {
			var newStats = new(CPUStats)

			newStats.CollectedAt = time.Now()
			newStats.User = make(map[int]int64)
			newStats.Nice = make(map[int]int64)
			newStats.System = make(map[int]int64)
			newStats.Idle = make(map[int]int64)
			newStats.IOWait = make(map[int]int64)
			newStats.IRQ = make(map[int]int64)
			newStats.SoftIRQ = make(map[int]int64)
			newStats.Steal = make(map[int]int64)
			newStats.Guest = make(map[int]int64)
			newStats.GuestNice = make(map[int]int64)

			for _, line := range fileutil.ShouldReadAllLines(`/proc/stat`) {
				if cols := rxutil.Whitespace.Split(line, -1); len(cols) >= 9 {
					if cols[0] == `cpu` {
						ok = true
						i = -1
					} else if strings.HasPrefix(cols[0], `cpu`) {
						i = int(typeutil.Int(cols[0][3:]))
						ok = true
					} else {
						ok = false
					}

					if ok {
						newStats.User[i] = typeutil.Int(cols[1])
						newStats.Nice[i] = typeutil.Int(cols[2])
						newStats.System[i] = typeutil.Int(cols[3])
						newStats.Idle[i] = typeutil.Int(cols[4])
						newStats.IOWait[i] = typeutil.Int(cols[5])
						newStats.IRQ[i] = typeutil.Int(cols[6])
						newStats.SoftIRQ[i] = typeutil.Int(cols[7])
						newStats.Steal[i] = typeutil.Int(cols[8])

						if cols := rxutil.Whitespace.Split(line, -1); len(cols) >= 11 {
							newStats.Guest[i] = typeutil.Int(cols[9])
							newStats.GuestNice[i] = typeutil.Int(cols[10])
						}
					}
				}
			}

			newStats.deltaFrom(self.lastStats)

			for k, v := range newStats.data(logical) {
				out[k] = v
			}

			self.lastStats = newStats
			time.Sleep(100 * time.Millisecond)
		}

		out[`cpu.count`] = logical
		out[`cpu.physical`] = physical
	}

	return out
}
