package data

import (
	"fmt"
	"runtime"

	"github.com/diegomagdaleno/whatmac"
)

type OS struct {
}

func (self OS) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	out[`os.platform`] = runtime.GOOS
	out[`os.family`] = runtime.GOOS
	out[`os.distribution`] = whatmac.GetProductName()
	out[`os.version`] = whatmac.GetProductUserVisibleVersion()
	out[`os.description`] = fmt.Sprintf(
		"%v %v",
		out[`os.distribution`],
		out[`os.version`],
	)
	out[`os.product_build_version`] = whatmac.GetProductBuildVersion()
	out[`os.hardware_platform`] = shellfl(`uname -i`).String()

	return out
}
