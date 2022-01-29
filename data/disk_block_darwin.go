package data

type BlockDevices struct {
	BlockDeviceRoot string
}

func (self BlockDevices) Collect() map[string]interface{} {
	var out = make(map[string]interface{})

	return out
}

// func (self BlockDevices) collectDevice(blockpath string) map[string]interface{} {
// 	physical := readvalue(blockpath, `queue`, `physical_block_size`).Int()
// 	logical := readvalue(blockpath, `queue`, `logical_block_size`).Int()

// 	return map[string]interface{}{
// 		`name`:               filepath.Base(blockpath),
// 		`device`:             fmt.Sprintf("/dev/%s", filepath.Base(blockpath)),
// 		`size`:               readvalue(blockpath, `size`).Int() * physical,
// 		`removable`:          readvalue(blockpath, `removable`).Bool(),
// 		`ssd`:                !readvalue(blockpath, `queue`, `rotational`).Bool(),
// 		`vendor`:             readvalue(blockpath, `device`, `vendor`).String(),
// 		`model`:              readvalue(blockpath, `device`, `model`).String(),
// 		`revision`:           readvalue(blockpath, `device`, `rev`).String(),
// 		`blocksize.physical`: physical,
// 		`blocksize.logical`:  logical,
// 	}
// }
