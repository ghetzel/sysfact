package data

import (
	"fmt"
	"strings"

	"go.gary.cool/go-stockutil/stringutil"
	"go.gary.cool/go-stockutil/typeutil"
)

type IPMI struct {
}

func (self IPMI) Collect() map[string]any {
	out := make(map[string]any)
	i := 0

	for _, line := range lines(`ipmitool sensor list`) {
		if parts := strings.Split(line, `|`); len(parts) >= 10 {
			prefix := fmt.Sprintf("ipmi.sensors.%d.", i)

			if name := strings.TrimSpace(parts[0]); name != `` {
				out[prefix+`name`] = stringutil.Underscore(name)

				if value := strings.TrimSpace(parts[1]); typeutil.IsNumeric(value) {
					out[prefix+`value`] = typeutil.Float(value)
				} else if strings.HasPrefix(value, `0x`) {
					out[prefix+`value`] = typeutil.Int(value)
				} else {
					continue
				}

				unit := strings.TrimSpace(parts[2])
				unit = strings.ToLower(unit)
				unit = stringutil.Underscore(unit)
				out[prefix+`unit`] = unit

				if low := strings.TrimSpace(parts[8]); typeutil.IsNumeric(low) {
					out[prefix+`warning_value`] = typeutil.Float(low)
				}

				if high := strings.TrimSpace(parts[9]); typeutil.IsNumeric(high) {
					out[prefix+`critical_value`] = typeutil.Float(high)
				}

				i += 1
			}
		}

	}

	return out
}
