package sysfact

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/stringutil"
)

type InfluxdbPayload []string

var InfluxdbTagValueCharFilter = regexp.MustCompile(`\s{2,}`)
var InfluxdbTagValueCharCompress = regexp.MustCompile(`[\s,]+`)

func (self InfluxdbPayload) Generate(series TupleSet, tags map[string]interface{}, now *time.Time) (string, error) {
	var epochUs int64

	if now == nil {
		epochUs = time.Now().UnixNano()
	} else {
		epochUs = now.UnixNano()
	}

	for _, metric := range series {
		metricValue := metric.Value

		// convert times and durations to epoch milliseconds
		if tm, ok := metricValue.(time.Time); ok {
			metricValue = tm.UnixNano() / 1e6
		} else if td, ok := metricValue.(time.Duration); ok {
			metricValue = td.Nanoseconds() / 1e6
		} else if tb, ok := metricValue.(bool); ok {
			if tb {
				metricValue = 1
			} else {
				metricValue = 0
			}
		}

		if stringutil.IsInteger(metricValue) || stringutil.IsFloat(metricValue) {
			finalTags := maputil.Append(tags, metric.Tags)

			for key, value := range finalTags {
				vStr := fmt.Sprintf("%v", value)
				vStr = strings.TrimSpace(vStr)
				vStr = InfluxdbTagValueCharFilter.ReplaceAllString(vStr, ``)
				vStr = InfluxdbTagValueCharCompress.ReplaceAllString(vStr, `_`)
				vStr = strings.TrimPrefix(vStr, `_`)
				vStr = strings.TrimSuffix(vStr, `_`)

				if vStr != `` {
					finalTags[key] = vStr
				}
			}

			joinedTags := ``

			if len(finalTags) > 0 {
				joinedTags += `,` + maputil.Join(finalTags, `=`, `,`)
			}

			var value string

			if v, err := stringutil.ToString(metricValue); err == nil {
				value = v
			} else {
				return ``, err
			}

			self = append(self, fmt.Sprintf("%s%s value=%s %d",
				metric.NormalizedKey,
				joinedTags,
				value,
				epochUs,
			))
		}
	}

	return strings.Join(self, "\n"), nil
}
