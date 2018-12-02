package sysfact

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/sysfact/plugins"
)

type ReporterKeyFormat int

const (
	FormatUnderscore ReporterKeyFormat = iota
	FormatPascalize
	FormatCamelize
)

type Reporter struct {
	Plugins     []plugins.Plugin
	FieldPrefix string
	KeyFormat   ReporterKeyFormat
}

func NewReporter(paths ...string) *Reporter {
	shellExecPath := append([]string{
		`~/.sysfact/shell.d`,
		`/usr/local/lib/sysfact/shell.d`,
		`/var/lib/sysfact/shell.d`,
	}, paths...)

	return &Reporter{
		Plugins: []plugins.Plugin{
			// add built-in system data collection plugin
			plugins.EmbeddedPlugin{},

			// add shell script plugin
			plugins.ShellPlugin{
				ExecPath:         shellExecPath,
				PerPluginTimeout: (30 * time.Second),
				MaxTimeout:       (60 * time.Second),
			},
		},
	}
}

// Generate and return the full report from all discovered plugins.
//
func (self *Reporter) Report() (map[string]interface{}, error) {
	outputData := make(map[string]interface{})

	//  collected_at is ALWAYS set
	outputData[self.keyformat("collected_at")] = time.Now()

	//  for each plugin
	for _, plugin := range self.Plugins {
		observations, _ := plugin.Collect()

		//  save all collected observations into an output map
		for _, observation := range observations {
			key := self.keyformat(observation.Name)

			if _, exists := outputData[key]; !exists {
				outputData[self.FieldPrefix+key] = observation.Value
			}
		}
	}

	return outputData, nil
}

// Generates a report and retrieves the values of the given fields.
//
func (self *Reporter) GetReportValues(fields []string, skipFields []string) (map[string]interface{}, error) {
	filteredValues := make(map[string]interface{})
	patterns := make([]*regexp.Regexp, 0)
	antipatterns := make([]*regexp.Regexp, 0)

	// build regexp patterns from input fields
	for _, pattern := range fields {
		var rx *regexp.Regexp

		if strings.ContainsAny(pattern, "[]()^$*") {
			rx = regexp.MustCompile(pattern)
		} else {
			rx = regexp.MustCompile("^" + self.FieldPrefix + pattern + "(?:\\..*)?$")
		}

		patterns = append(patterns, rx)
	}

	// build regexp patterns for fields to skip
	for _, antipattern := range skipFields {
		var rx *regexp.Regexp

		if strings.ContainsAny(antipattern, "[]()^$*") {
			rx = regexp.MustCompile(antipattern)
		} else {
			rx = regexp.MustCompile("^" + self.FieldPrefix + antipattern + "(?:\\..*)?$")
		}

		antipatterns = append(antipatterns, rx)
	}

	// generate report
	if report, err := self.Report(); err == nil {
		for field, value := range report {
			skip := false

			for _, antipattern := range antipatterns {
				if antipattern.MatchString(field) {
					skip = true
					break
				}
			}

			if skip {
				continue
			}

			if len(patterns) == 0 {
				filteredValues[field] = value
			} else {
				for _, pattern := range patterns {
					if pattern.MatchString(field) {
						filteredValues[field] = value
						break
					}
				}
			}
		}
	} else {
		return filteredValues, err
	}

	return filteredValues, nil
}

func (self *Reporter) keyformat(key string) string {
	parts := strings.Split(key, `.`)

	for i, part := range parts {
		switch self.KeyFormat {
		case FormatPascalize:
			// NOTE: stringutil.Camelize is wrong and actually returns PascalCase
			parts[i] = stringutil.Camelize(part)
		case FormatCamelize:
			for i, v := range stringutil.Camelize(part) {
				parts[i] = string(unicode.ToLower(v)) + part[i+1:]
				break
			}
		default:
			parts[i] = stringutil.Underscore(part)
		}
	}

	return strings.Join(parts, `.`)
}
