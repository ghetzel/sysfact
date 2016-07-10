package main

import (
	"encoding/json"
	"fmt"
	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghodss/yaml"
	"github.com/op/go-logging"
	"os"
	"reflect"
	"sort"
	"time"
)

var log = logging.MustGetLogger(`main`)

type Tuple struct {
	Key   string
	Value interface{}
}

type TupleSet []Tuple

func (self TupleSet) ToMap(flat bool) map[string]interface{} {
	output := make(map[string]interface{})

	for _, tuple := range self {
		output[tuple.Key] = tuple.Value
	}

	if !flat {
		if nestedOutput, err := maputil.DiffuseMap(output, `.`); err == nil {
			output = nestedOutput
		}
	}

	return output
}

func main() {
	app := cli.NewApp()
	app.Name = `sysfact`
	app.Version = `0.0.1`
	app.EnableBashCompletion = false
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  `log-level, L`,
			Usage: `The level of logging verbosity to output.`,
			Value: `error`,
		},
		cli.StringFlag{
			Name:  `plugin-log-level, P`,
			Usage: `The level of logging verbosity to output when executing plugins.`,
			Value: `critical`,
		},
		cli.StringFlag{
			Name:  `format, f`,
			Usage: `How the output should be formatted (one of "flat", "json", "yaml", "graphite", or "tsdb")`,
			Value: `flat`,
		},
		cli.StringSliceFlag{
			Name:   `additional-paths, p`,
			Usage:  `Zero or more additional paths to search for plugins in.`,
			EnvVar: `SYSFACT_PATH`,
		},
	}

	var reporter *Reporter

	app.Before = func(c *cli.Context) error {
		logging.SetFormatter(logging.MustStringFormatter(`%{color}%{level:.4s}%{color:reset}[%{id:04d}] %{message}`))

		if level, err := logging.LogLevel(c.String(`log-level`)); err == nil {
			logging.SetLevel(level, `main`)
		}

		if level, err := logging.LogLevel(c.String(`plugin-log-level`)); err == nil {
			logging.SetLevel(level, `plugins`)
		}

		reporter = NewReporter(c.StringSlice(`additional-paths`)...)

		return nil
	}

	app.Action = func(c *cli.Context) {
		var values map[string]interface{}

		if c.NArg() > 0 {
			if v, err := reporter.GetReportValues(c.Args()); err == nil {
				values = v
			} else {
				log.Fatal(err)
				return
			}
		} else {
			if v, err := reporter.Report(); err == nil {
				values = v
			} else {
				log.Fatal(err)
				return
			}
		}

		keys := maputil.StringKeys(values)
		sort.Strings(keys)
		tuples := make(TupleSet, len(keys))

		for i, fieldName := range keys {
			if value, ok := values[fieldName]; ok {
				tuples[i] = Tuple{
					Key:   fieldName,
					Value: value,
				}
			}
		}

		printWithFormat(c.String(`format`), tuples)
	}

	app.Run(os.Args)
}

func printWithFormat(format string, tuples TupleSet) {
	now := time.Now()

	switch format {
	case `flat`:
		for _, tuple := range tuples {
			fmt.Printf("%s=%v\n", tuple.Key, tuple.Value)
		}

	case `yaml`:
		if data, err := yaml.Marshal(tuples.ToMap(false)); err == nil {
			fmt.Println(string(data[:]))
		}

	case `json`:
		if data, err := json.MarshalIndent(tuples.ToMap(false), ``, `  `); err == nil {
			fmt.Println(string(data[:]))
		}

	case `graphite`:
		epoch := now.Unix()

		for _, tuple := range tuples {
			if value, err := toFloat(tuple.Value); err == nil {
				fmt.Printf("%s %f %d\n", tuple.Key, value, epoch)
			} else {
				log.Notice(err)
			}
		}

	case `tsdb`:
		epochMs := int64(now.UnixNano() / 1000000)

		for _, tuple := range tuples {
			if value, err := toFloat(tuple.Value); err == nil {
				fmt.Printf("put %s %d %f\n", tuple.Key, epochMs, value)
			} else {
				log.Notice(err)
			}
		}
	}
}

func toFloat(in interface{}) (float64, error) {
	floatT := reflect.TypeOf(float64(0))

	if reflect.TypeOf(in).ConvertibleTo(floatT) {
		floatV := reflect.ValueOf(in).Convert(floatT)
		return floatV.Float(), nil
	}

	return 0.0, fmt.Errorf("Cannot convert %T to float64", in)
}
