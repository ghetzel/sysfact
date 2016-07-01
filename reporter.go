package main

import (
	"encoding/json"
	"fmt"
	"github.com/ghetzel/sysfact/plugins"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type Reporter struct {
	Plugins     []plugins.Plugin
	FieldPrefix string
}

func NewReporter() *Reporter {
	reporter := &Reporter{}

	shellExecPath := []string{
		`./shell.d`,
		`/usr/local/lib/sysfact/shell.d`,
		`/var/lib/sysfact/shell.d`,
	}

	reporter.Plugins = append(reporter.Plugins, plugins.ShellPlugin{
		ExecPath:         shellExecPath,
		PerPluginTimeout: (30 * time.Second),
		MaxTimeout:       (60 * time.Second),
	})

	return reporter
}

func (self *Reporter) Report() (map[string]interface{}, error) {
	log.Info("Gathering report data...")
	outputData := make(map[string]interface{})

	//  collected_at is ALWAYS set
	outputData["collected_at"] = "now"

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

func (self *Reporter) GetReportValues(fields []string) (map[string]interface{}, error) {
	filteredValues := make(map[string]interface{})
	patterns := make([]*regexp.Regexp, 0)

	// built regexp patterns from input fields
	for _, pattern := range fields {
		if strings.ContainsAny(pattern, "[]()^$*") {
			patterns = append(patterns, regexp.MustCompile(pattern))
		} else {
			patterns = append(patterns, regexp.MustCompile("^"+self.FieldPrefix+pattern+"$"))
		}
	}

	// short circuit if none provided
	if len(patterns) == 0 {
		return filteredValues, nil
	}

	// generate report
	if report, err := self.Report(); err == nil {
		for field, value := range report {
			for _, pattern := range patterns {
				if pattern.MatchString(field) {
					filteredValues[field] = value
				}
			}
		}
	} else {
		return filteredValues, err
	}

	return filteredValues, nil
}

func (self *Reporter) PrintReportValues(fields []string) error {
	if filteredValues, err := self.GetReportValues(fields); err == nil {
		// print selected values
		if output, err := json.MarshalIndent(filteredValues, "", "    "); err == nil {
			fmt.Println(string(output))
		} else {
			return err
		}
	} else {
		return err
	}

	return nil
}

func (self *Reporter) PrintReport() error {
	if report, err := self.Report(); err == nil {
		if output, err := json.MarshalIndent(report, "", "    "); err == nil {
			fmt.Println(string(output))
		} else {
			return err
		}
	} else {
		return err
	}

	return nil
}
