package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.gary.cool/go-stockutil/fileutil"
	"go.gary.cool/go-stockutil/typeutil"
)

type BlockDevices struct {
	BlockDeviceRoot string
}

func (self BlockDevices) Collect() map[string]any {
	out := make(map[string]any)
	devid := 0
	root := `/sys/block`

	if r := self.BlockDeviceRoot; r != `` {
		root = r
	}

	if items, err := os.ReadDir(root); err == nil {
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

func (self BlockDevices) collectDevice(blockpath string) map[string]any {
	physical := readvalue(blockpath, `queue`, `physical_block_size`).Int()
	logical := readvalue(blockpath, `queue`, `logical_block_size`).Int()

	return map[string]any{
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

func readvalue(pathparts ...string) typeutil.Variant {
	if value, err := fileutil.ReadAllString(filepath.Join(pathparts...)); err == nil {
		return typeutil.V(strings.TrimSpace(value))
	} else {
		return typeutil.Nil()
	}
}
