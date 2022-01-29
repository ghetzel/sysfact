package data

import (
	"time"

	"github.com/ghetzel/go-stockutil/mathutil"
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
