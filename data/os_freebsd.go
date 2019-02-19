package data

import (
	"fmt"
	"runtime"
	"strings"
)

type OS struct {
}

func (self OS) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	out[`os.platform`] = runtime.GOOS
	out[`os.family`] = runtime.GOOS
	out[`os.distribution`] = strings.ToLower(shell(`uname -s`).String())
	out[`os.version`] = shell(`uname -r`).String()
	out[`os.description`] = fmt.Sprintf(
		"%v %v %v",
		out[`os.distribution`],
		out[`os.version`],
		shell(`uname -i`).String(),
	)

	return out
}
