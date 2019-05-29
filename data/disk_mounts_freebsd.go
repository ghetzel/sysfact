package data

type Mounts struct {
}

func (self Mounts) Collect() map[string]interface{} {
	out := make(map[string]interface{})
	// mnt := 0

	// if mounts, err := fileutil.ReadAllLines(`/proc/mounts`); err == nil {
	// 	for _, line := range mounts {
	// 		if columns := strings.Split(line, ` `); len(columns) >= 3 {
	// 			device := columns[0]
	// 			mountpoint := columns[1]
	// 			filesystem := columns[2]

	// 			out[fmt.Sprintf("disk.mounts.%d.mount", mnt)] = mountpoint
	// 			out[fmt.Sprintf("disk.mounts.%d.device", mnt)] = device
	// 			out[fmt.Sprintf("disk.mounts.%d.filesystem", mnt)] = filesystem

	// 			for i, flag := range strings.Split(columns[3], `,`) {
	// 				out[fmt.Sprintf("disk.mounts.%d.flags.%d", mnt, i)] = flag
	// 			}

	// 			for i, line := range lines(fmt.Sprintf("df -P --block-size 1 %s", mountpoint)) {
	// 				if i == 0 {
	// 					continue
	// 				} else if parts := strings.Split(line, ` `); len(parts) >= 5 {
	// 					used := typeutil.Int(parts[3])
	// 					total := typeutil.Int(parts[1])

	// 					out[fmt.Sprintf("disk.mounts.%d.total", mnt)] = total
	// 					out[fmt.Sprintf("disk.mounts.%d.available", mnt)] = typeutil.Int(parts[2])
	// 					out[fmt.Sprintf("disk.mounts.%d.used", mnt)] = used

	// 					if total > 0 {
	// 						out[fmt.Sprintf("disk.mounts.%d.percent_used", mnt)] = (float64(used) / float64(total)) * 100.0
	// 					} else {
	// 						out[fmt.Sprintf("disk.mounts.%d.percent_used", mnt)] = 0
	// 					}
	// 				}
	// 			}

	// 			mnt += 1
	// 		}
	// 	}
	// }

	return out
}
