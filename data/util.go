package data

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ghetzel/go-stockutil/fileutil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/mattn/go-shellwords"
)

func normalize(in interface{}) interface{} {
	if inS, ok := in.(string); ok {
		inS = strings.TrimSpace(inS)
		inS = strings.TrimSuffix(inS, `-`)
		inS = stringutil.SqueezeSpace(inS)

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

			return inS
		}
	} else {
		return in
	}
}

// shell first non-empty
func shellfne(cmdlines ...string) typeutil.Variant {
	for _, cmdline := range cmdlines {
		if v := shell(cmdline); strings.TrimSpace(v.String()) != `` {
			return v
		}
	}

	return typeutil.Nil()
}

func shell(cmdline string, values ...interface{}) typeutil.Variant {
	if words, err := shellwords.Parse(fmt.Sprintf(cmdline, values...)); err == nil {
		cmd := exec.Command(words[0], words[1:]...)

		if data, err := cmd.Output(); err == nil {
			return typeutil.V(strings.TrimSpace(string(data)))
		}
	}
	return typeutil.Nil()
}

func lines(cmdline string) []string {
	return strings.Split(shell(cmdline).String(), "\n")
}

func shellfl(cmdline string) typeutil.Variant {
	if value := strings.TrimSpace(lines(cmdline)[0]); value != `` {
		return typeutil.V(value)
	} else {
		return typeutil.Nil()
	}
}

func readvalue(path ...string) typeutil.Variant {
	if line, err := fileutil.ReadFirstLine(filepath.Join(path...)); err == nil {
		return typeutil.V(normalize(line))
	}

	return typeutil.Nil()
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

				for k, v := range collector.Collect() {
					output[k] = normalize(v)
				}

				mergelock.Unlock()
			}

			wg.Done()
		}()
	}

	collect(&wg, `cpu`, CPU{})
	collect(&wg, `memory`, Memory{})
	collect(&wg, `kernel`, Kernel{})
	collect(&wg, `network`, Network{})
	collect(&wg, `os`, OS{})
	collect(&wg, `system`, System{})
	collect(&wg, `ipmi`, IPMI{})
	collect(&wg, `disk.block`, BlockDevices{})
	collect(&wg, `disk.mounts`, Mounts{})

	wg.Wait()

	return output
}
