package data

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghetzel/go-stockutil/fileutil"
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
	physical := readvalue(blockpath, `queue`, `physical_block_size`).Int()
	logical := readvalue(blockpath, `queue`, `logical_block_size`).Int()

	return map[string]interface{}{
		`name`:               filepath.Base(blockpath),
		`device`:             fmt.Sprintf("/dev/%s", filepath.Base(blockpath)),
		`size`:               readvalue(blockpath, `size`).Int() * physical,
		`removable`:          readvalue(blockpath, `removable`).Bool(),
		`ssd`:                !readvalue(blockpath, `queue`, `rotational`).Bool(),
		`vendor`:             readvalue(blockpath, `device`, `vendor`).String(),
		`model`:              readvalue(blockpath, `device`, `model`).String(),
		`revision`:           readvalue(blockpath, `device`, `rev`).String(),
		`blocksize.physical`: physical,
		`blocksize.logical`:  logical,
	}
}
