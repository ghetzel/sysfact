package main

import (
	"encoding/json"
	"fmt"
	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghodss/yaml"
	"github.com/op/go-logging"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"
)

var log = logging.MustGetLogger(`main`)

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
			Usage: `How the output should be formatted (one of "flat", "json", "yaml", "graphite", "tsdb", or "influxdb")`,
			Value: `flat`,
		},
		cli.StringSliceFlag{
			Name:   `additional-paths, A`,
			Usage:  `Zero or more additional paths to search for plugins in.`,
			EnvVar: `SYSFACT_PATH`,
		},
		cli.StringSliceFlag{
			Name:  `skip-field, x`,
			Usage: `Zero of more patterns of fields to explicitly omit from the output.`,
		},
		cli.StringSliceFlag{
			Name:  `tag, t`,
			Usage: `Zero or key=value pairs to include in output formats that support additional data.`,
		},
		cli.StringFlag{
			Name:  `prefix, p`,
			Usage: `A prefix to prepend to all facts. For output types the represent nested data structures, all output will be nested under this dot-separated path.`,
		},
		cli.BoolFlag{
			Name:  `extract-tags, T`,
			Usage: `Automatically extract related tag values from arrays of items.`,
		},
	}

	var reporter *Reporter

	app.Before = func(c *cli.Context) error {
		logging.SetFormatter(logging.MustStringFormatter(`%{color}%{level:.4s}%{color:reset}[%{id:04d}] %{message}`))

		if level, err := logging.LogLevel(c.String(`log-level`)); err == nil {
			logging.SetLevel(level, ``)
		}

		if level, err := logging.LogLevel(c.String(`plugin-log-level`)); err == nil {
			logging.SetLevel(level, `plugins`)
		}

		reporter = NewReporter(c.StringSlice(`additional-paths`)...)

		reporter.FieldPrefix = c.String(`prefix`)

		return nil
	}

	app.Action = func(c *cli.Context) {
		var report map[string]interface{}
		tags := make(map[string]interface{})

		for _, kv := range c.StringSlice(`tag`) {
			parts := strings.SplitN(kv, `=`, 2)

			if len(parts) == 2 {
				tags[parts[0]] = parts[1]
			}
		}

		if v, err := reporter.GetReportValues(c.Args(), c.StringSlice(`skip-field`)); err == nil {
			report = v
		} else {
			log.Fatal(err)
			return
		}

		tagsets := make(map[string]map[string]interface{})
		tuples := make(TupleSet, 0)
		modifiedTuples := make(map[string]TupleSet)

		if err := maputil.Walk(report, func(value interface{}, path []string, leaf bool) error {
			if leaf && len(path) > 0 {
				path = strings.Split(path[0], `.`)
				key := strings.Join(path, `.`)
				normalizedPath := make([]string, 0)
				tagsetKey := ``
				firstIndexAfterTagsetKey := 0

				for i, part := range path {
					if v, err := stringutil.ConvertToInteger(part); err == nil {
						tagsetKey = strings.Join(path[0:(i+1)], `.`)
						firstIndexAfterTagsetKey = (i + 1)

						if tags, ok := tagsets[tagsetKey]; !ok {
							tagsets[tagsetKey] = map[string]interface{}{
								`index`: v,
							}
						} else {
							tags[`index`] = v
						}

						normalizedPath = append(path[0:i])
						break
					}
				}

				if !c.Bool(`extract-tags`) || tagsetKey == `` {
					tuples = append(tuples, Tuple{
						Key:           key,
						Value:         value,
						NormalizedKey: key,
					})
				} else {
					if stringutil.IsInteger(value) || stringutil.IsFloat(value) {
						if v, err := stringutil.ConvertToFloat(value); err == nil {
							var currentTupleSet TupleSet

							if ts, ok := modifiedTuples[tagsetKey]; ok {
								currentTupleSet = ts
							} else {
								currentTupleSet = make(TupleSet, 0)
							}

							currentTupleSet = append(currentTupleSet, Tuple{
								Key:   key,
								Value: v,
								NormalizedKey: strings.Join(
									append(normalizedPath, path[firstIndexAfterTagsetKey:]...),
									`.`,
								),
							})

							modifiedTuples[tagsetKey] = currentTupleSet
						} else {
							log.Warningf("Failed to convert value for %s: %v", key, err)
						}
					} else if tags, ok := tagsets[tagsetKey]; ok {
						// non-numeric values become tags
						tagKey := strings.Join(path[firstIndexAfterTagsetKey:], `_`)

						tags[tagKey] = value
					}
				}

			}

			return nil
		}); err == nil {
			for tagsetKey, tupleset := range modifiedTuples {
				for _, tuple := range tupleset {
					if tags, ok := tagsets[tagsetKey]; ok {
						tuple.Tags = tags
					}

					tuples = append(tuples, tuple)
				}
			}

			sort.Sort(tuples)

			printWithFormat(c.String(`format`), tuples, tags)
		} else {
			log.Fatal(err)
		}

		// for i, fieldName := range keys {
		// 	if value, ok := values[fieldName]; ok {
		// 		tuples[i] = Tuple{
		// 			Key:           fieldName,
		// 			Value:         value,
		// 		}
		// 	}
		// }

	}

	app.Run(os.Args)
}

func printWithFormat(format string, tuples TupleSet, tags map[string]interface{}) {
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
		tags := strings.TrimSpace(` ` + maputil.Join(tags, `=`, ` `))

		for _, tuple := range tuples {
			if value, err := toFloat(tuple.Value); err == nil {
				fmt.Printf("put %s %d %f%s\n", tuple.Key, epochMs, value, tags)
			} else {
				log.Notice(err)
			}
		}

	case `influxdb`:
		var writer InfluxdbPayload

		if out, err := writer.Generate(tuples, tags, &now); err == nil {
			fmt.Println(out)
		} else {
			log.Error(err)
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
