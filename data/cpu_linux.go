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

type CPUStats struct {
	User        map[int]int64
	Nice        map[int]int64
	System      map[int]int64
	Idle        map[int]int64
	IOWait      map[int]int64
	IRQ         map[int]int64
	SoftIRQ     map[int]int64
	Steal       map[int]int64
	Guest       map[int]int64
	GuestNice   map[int]int64
	CollectedAt time.Time
}

func (self *CPUStats) MaxCore() int {
	var i int

	for k, _ := range self.User {
		if k > i {
			i = k
		}
	}

	return i
}

func (self *CPUStats) Active(core int) int64 {
	return self.Total(core) - self.Idle[core]
}

func (self *CPUStats) Total(core int) (sum int64) {
	sum += self.User[core]
	sum += self.Nice[core]
	sum += self.System[core]
	sum += self.Idle[core]
	sum += self.IOWait[core]
	sum += self.IRQ[core]
	sum += self.SoftIRQ[core]
	sum += self.Steal[core]
	sum += self.Guest[core]
	sum += self.GuestNice[core]

	return
}

func (self *CPUStats) deltaFrom(other *CPUStats) {
	if other != nil {
		for i := -1; i < self.MaxCore(); i++ {
			self.User[i] -= other.User[i]
			self.Nice[i] -= other.Nice[i]
			self.System[i] -= other.System[i]
			self.Idle[i] -= other.Idle[i]
			self.IOWait[i] -= other.IOWait[i]
			self.IRQ[i] -= other.IRQ[i]
			self.SoftIRQ[i] -= other.SoftIRQ[i]
			self.Steal[i] -= other.Steal[i]
			self.Guest[i] -= other.Guest[i]
			self.GuestNice[i] -= other.GuestNice[i]
		}
	}
}

func (self CPUStats) data(logicalCoreCount int) map[string]interface{} {
	var out = make(map[string]interface{})

	for i := -1; i < logicalCoreCount; i++ {
		var total = float64(self.Total(i))
		var key = `cpu.usage.`

		if i >= 0 {
			key = `cpu.cores.` + typeutil.String(i) + `.usage.`
		}

		out[key+`idle`] = mathutil.RoundPlaces(100*(float64(self.Idle[i])/total), 1)
		out[key+`active`] = mathutil.RoundPlaces(100*(float64(self.Active(i))/total), 1)
		out[key+`user`] = mathutil.RoundPlaces(100*(float64(self.User[i])/total), 1)
		out[key+`nice`] = mathutil.RoundPlaces(100*(float64(self.Nice[i])/total), 1)
		out[key+`system`] = mathutil.RoundPlaces(100*(float64(self.System[i])/total), 1)
		out[key+`iowait`] = mathutil.RoundPlaces(100*(float64(self.IOWait[i])/total), 1)
		out[key+`irq`] = mathutil.RoundPlaces(100*(float64(self.IRQ[i])/total), 1)
		out[key+`softirq`] = mathutil.RoundPlaces(100*(float64(self.SoftIRQ[i])/total), 1)
		out[key+`steal`] = mathutil.RoundPlaces(100*(float64(self.Steal[i])/total), 1)
		out[key+`guest`] = mathutil.RoundPlaces(100*(float64(self.Guest[i])/total), 1)
		out[key+`guest_nice`] = mathutil.RoundPlaces(100*(float64(self.GuestNice[i])/total), 1)
	}

	return out
}

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
