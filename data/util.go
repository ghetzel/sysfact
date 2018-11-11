package data

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/mattn/go-shellwords"
)

func normalize(in interface{}) interface{} {
	if inS, ok := in.(string); ok {
		inS = strings.TrimSpace(inS)

		switch strings.ToLower(inS) {
		case `yes`, `on`:
			return true
		case `no`, `off`:
			return false
		case `to be filled by o.e.m.`, `not specified`:
			return nil
		default:
			inS = strings.Replace(inS, `(R)`, ``, -1)
			inS = strings.Replace(inS, `(TM)`, ``, -1)

			return typeutil.Auto(inS)
		}
	} else {
		return in
	}
}

func shell(cmdline string) typeutil.Variant {
	if words, err := shellwords.Parse(cmdline); err == nil {
		cmd := exec.Command(words[0], words[1:]...)

		if data, err := cmd.Output(); err == nil {
			return typeutil.V(strings.TrimSpace(string(data)))
		}
	}
	return typeutil.V(nil)
}

func lines(cmdline string) []string {
	return strings.Split(shell(cmdline).String(), "\n")
}

type Collector interface {
	Collect() map[string]interface{}
}

func Collect(only ...string) map[string]interface{} {
	var wg sync.WaitGroup
	var mergelock sync.Mutex
	output := make(map[string]interface{})

	collect := func(wg *sync.WaitGroup, want string, collector Collector) {
		wg.Add(1)

		go func() {
			if len(only) == 0 || sliceutil.ContainsString(only, want) {
				mergelock.Lock()
				output, _ = maputil.Merge(output, collector.Collect())
				mergelock.Unlock()
			}

			wg.Done()
		}()
	}

	collect(&wg, `cpu`, CPU{})
	collect(&wg, `kernel`, Kernel{})
	collect(&wg, `network`, Network{})
	collect(&wg, `os`, OS{})
	collect(&wg, `system`, System{})

	wg.Wait()

	return output
}
