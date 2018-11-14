package data

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghetzel/go-stockutil/fileutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type BlockDevices struct {
	BlockDeviceRoot string
}

func (self BlockDevices) Collect() map[string]interface{} {
	out := make(map[string]interface{})
	devid := 0
	root := `/sys/block`

	if r := self.BlockDeviceRoot; r != `` {
		root = r
	}

	if items, err := ioutil.ReadDir(root); err == nil {
		for _, entry := range items {
			devroot := filepath.Join(root, entry.Name())

			if fileutil.Exists(filepath.Join(devroot, `device`)) {
				for k, v := range self.collectDevice(devroot) {
					out[fmt.Sprintf("disk.block.%d.%s", devid, k)] = v
				}

				devid += 1
			}
		}
	}

	return out
}

func (self BlockDevices) collectDevice(blockpath string) map[string]interface{} {
	physical := typeutil.Int(readvalue(blockpath, `queue`, `physical_block_size`))
	logical := typeutil.Int(readvalue(blockpath, `queue`, `logical_block_size`))

	return map[string]interface{}{
		`name`:               filepath.Base(blockpath),
		`device`:             fmt.Sprintf("/dev/%s", filepath.Base(blockpath)),
		`size`:               typeutil.Int(readvalue(blockpath, `size`)) * physical,
		`removable`:          typeutil.Bool(readvalue(blockpath, `removable`)),
		`ssd`:                !typeutil.Bool(readvalue(blockpath, `queue`, `rotational`)),
		`vendor`:             typeutil.String(readvalue(blockpath, `device`, `vendor`)),
		`model`:              typeutil.String(readvalue(blockpath, `device`, `model`)),
		`revision`:           typeutil.String(readvalue(blockpath, `device`, `rev`)),
		`blocksize.physical`: physical,
		`blocksize.logical`:  logical,
	}
}
