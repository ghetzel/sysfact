package main

import (
	"fmt"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"regexp"
	"strings"
	"time"
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
		if stringutil.IsInteger(metric.Value) || stringutil.IsFloat(metric.Value) {
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

			self = append(self, fmt.Sprintf("%s%s value=%v %d",
				metric.NormalizedKey,
				joinedTags,
				metric.Value,
				epochUs,
			))
		}
	}

	return strings.Join(self, "\n"), nil
}
