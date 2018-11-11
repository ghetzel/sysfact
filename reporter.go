package sysfact

import (
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/sysfact/plugins"
)

type Reporter struct {
	Plugins     []plugins.Plugin
	FieldPrefix string
}

func NewReporter(paths ...string) *Reporter {
	reporter := &Reporter{}

	// shellExecPath := append([]string{
	// 	`~/.sysfact/shell.d`,
	// 	`/usr/local/lib/sysfact/shell.d`,
	// 	`/var/lib/sysfact/shell.d`,
	// }, paths...)

	// reporter.Plugins = append(reporter.Plugins, plugins.ShellPlugin{
	// 	ExecPath:         shellExecPath,
	// 	PerPluginTimeout: (30 * time.Second),
	// 	MaxTimeout:       (60 * time.Second),
	// })

	reporter.Plugins = append(reporter.Plugins, plugins.EmbeddedPlugin{})

	return reporter
}

// Generate and return the full report from all discovered plugins.
//
func (self *Reporter) Report() (map[string]interface{}, error) {
	log.Info("Gathering report data...")
	outputData := make(map[string]interface{})

	//  collected_at is ALWAYS set
	outputData["collected_at"] = time.Now()

	//  for each plugin
	for _, plugin := range self.Plugins {
		log.Infof("Collecting data for %s", reflect.TypeOf(plugin))
		observations, _ := plugin.Collect()

		//  save all collected observations into an output map
		for _, observation := range observations {
			if _, exists := outputData[observation.Name]; !exists {
				outputData[self.FieldPrefix+observation.Name] = observation.Value
			} else {
				log.Warningf("Cannot set value for field '%s', another plugin has already set a value for this field", observation.Name)
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
