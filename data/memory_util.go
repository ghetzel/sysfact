package data

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

func dmidecodeMemory(out map[string]interface{}) {
	// Memory Banks
	banks := shell(`dmidecode -t 16`).String()

	for i, bank := range strings.Split(banks, `Physical Memory Array`) {
		if i == 0 {
			continue
		} else {
			prefix := fmt.Sprintf("memory.hw.banks.%d.", i-1)

			for _, line := range strings.Split(bank, "\n") {
				k, v := stringutil.SplitPair(line, `:`)
				k = strings.TrimSpace(k)
				k = strings.ToLower(k)
				k = stringutil.Underscore(k)
				v = strings.TrimSpace(v)

				switch k {
				case `maximum_capacity`:
					out[prefix+`capacity`], _ = humanize.ParseBytes(v)

				case `error_correction_type`:
					if strings.Contains(strings.ToLower(v), `none`) {
						out[prefix+`ecc`] = false
					} else {
						out[prefix+`ecc`] = true
						out[prefix+`ecc_type`] = stringutil.Underscore(v)
					}

				case `number_of_devices`:
					out[prefix+`slot_count`] = typeutil.Int(v)
				}
			}
		}
	}

	devices := shell(`dmidecode -t 17`).String()
	dimmTotal := 0
	empties := make(map[int]bool)

	for i, dimm := range strings.Split(devices, `Memory Device`) {
		if i == 0 {
			continue
		} else {
			dimmTotal += 1
			prefix := fmt.Sprintf("memory.hw.slots.%d.", i-1)
			isEmpty := false
			dimmLines := strings.Split(dimm, "\n")

			for _, line := range dimmLines {
				k, v := stringutil.SplitPair(line, `:`)
				k = strings.TrimSpace(k)
				k = strings.ToLower(k)
				k = stringutil.Underscore(k)
				v = strings.TrimSpace(v)

				switch k {
				case `size`:
					if sz, err := humanize.ParseBytes(v); err != nil || sz == 0 {
						empties[i] = true
					}
				}
			}

			if empty, ok := empties[i]; ok && empty {
				isEmpty = true
			}

			out[prefix+`empty`] = isEmpty

			for _, line := range dimmLines {
				k, v := stringutil.SplitPair(line, `:`)
				k = strings.TrimSpace(k)
				k = strings.ToLower(k)
				k = stringutil.Underscore(k)
				v = strings.TrimSpace(v)

				// always output, even when the slot is empty
				switch k {
				case `locator`:
					out[prefix+`name`] = v
				case `bank_locator`:
					out[prefix+`bank_name`] = v
				case `data_width`:
					bits, _ := stringutil.SplitPair(v, ` `)
					out[prefix+`bits`] = typeutil.Int(bits)
				case `form_factor`:
					out[prefix+`form_factor`] = stringutil.Underscore(strings.ToLower(v))
				}

				if !isEmpty {
					switch k {
					case `type`:
						out[prefix+`type`] = stringutil.Underscore(strings.ToLower(v))
					case `size`:
						out[prefix+`size`], _ = humanize.ParseBytes(v)
					case `manufacturer`:
						out[prefix+`make`] = v
					case `part_number`:
						out[prefix+`model`] = v
					case `serial_number`:
						out[prefix+`serial`] = v
					case `asset_tag`:
						out[prefix+`asset_tag`] = v
					case `rank`:
						out[prefix+`rank`] = typeutil.Int(v)
					case `speed`:
						mts, _ := stringutil.SplitPair(v, ` `)
						out[prefix+`speed_mhz`] = typeutil.Int(mts)
					}
				}
			}
		}
	}

	if dimmTotal > 0 {
		out[`memory.hw.slots_count`] = dimmTotal
		out[`memory.hw.slots_empty`] = len(empties)
		out[`memory.hw.slots_populated`] = dimmTotal - len(empties)
	}
}
