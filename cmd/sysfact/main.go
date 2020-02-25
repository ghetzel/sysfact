package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-stockutil/executil"
	"github.com/ghetzel/go-stockutil/fileutil"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/sysfact"
	"github.com/ghodss/yaml"
)

func main() {
	app := cli.NewApp()
	app.Name = `sysfact`
	app.Usage = `A utility for collecting and formatting system information.`
	app.Version = sysfact.Version
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
			Usage: `How the output should be formatted (one of "flat", "env", "values", "json", "yaml", "graphite", "tsdb", or "influxdb")`,
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
		cli.StringFlag{
			Name:  `key-format`,
			Usage: `How keys in the output data should be formatted.  One of: "underscore", "camel", "pascal"`,
		},
		cli.BoolFlag{
			Name:  `extract-tags, T`,
			Usage: `Automatically extract related tag values from arrays of items.`,
		},
		cli.StringFlag{
			Name:  `http-url, u`,
			Usage: `Submit the formatted output to the given URL via HTTP`,
		},
		cli.StringFlag{
			Name:  `http-method, m`,
			Usage: `The HTTP method to use when submitting output.`,
			Value: `post`,
		},
		cli.StringSliceFlag{
			Name:  `http-header, H`,
			Usage: `Zero or more HTTP headers to include in the request (specfied as 'Header-Name=value')`,
		},
		cli.StringSliceFlag{
			Name:  `http-query, q`,
			Usage: `Zero or more query string parameters to include in the request (specfied as 'key=value')`,
		},
		cli.DurationFlag{
			Name:  `http-timeout`,
			Usage: `The time to wait for the HTTP request to complete.`,
			Value: (10 * time.Second),
		},
	}

	var reporter *sysfact.Reporter

	app.Before = func(c *cli.Context) error {
		log.SetLevelString(c.String(`log-level`))

		reporter = sysfact.NewReporter(c.StringSlice(`additional-paths`)...)
		reporter.FieldPrefix = c.String(`prefix`)

		switch c.String(`key-format`) {
		case `camel`:
			reporter.KeyFormat = sysfact.FormatCamelize
		case `pascal`:
			reporter.KeyFormat = sysfact.FormatPascalize
		case `underscore`:
			reporter.KeyFormat = sysfact.FormatUnderscore
		}

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
		tuples := make(sysfact.TupleSet, 0)
		modifiedTuples := make(map[string]sysfact.TupleSet)

		if err := maputil.Walk(report, func(value interface{}, path []string, leaf bool) error {
			if leaf && len(path) > 0 {
				path = strings.Split(path[0], `.`)
				key := strings.Join(path, `.`)
				normalizedPath := make([]string, 0)
				tagsetKey := ``

				for i, part := range path {
					if v, err := stringutil.ConvertToInteger(part); err == nil {
						tagsetKey = strings.Join(path[0:(i+1)], `.`)
						index_key := `index`

						if (i - 1) > 0 {
							index_key = string(path[i-1]) + `_` + index_key
						}

						if tags, ok := tagsets[tagsetKey]; !ok {
							tagsets[tagsetKey] = map[string]interface{}{
								index_key: v,
							}
						} else {
							tags[index_key] = v
						}
					} else {
						normalizedPath = append(normalizedPath, part)
					}
				}

				if !c.Bool(`extract-tags`) || tagsetKey == `` {
					tuples = append(tuples, sysfact.Tuple{
						Key:           key,
						Value:         value,
						NormalizedKey: key,
					})
				} else {
					if stringutil.IsInteger(value) || stringutil.IsFloat(value) {
						if v, err := stringutil.ConvertToFloat(value); err == nil {
							var currentTupleSet sysfact.TupleSet

							if ts, ok := modifiedTuples[tagsetKey]; ok {
								currentTupleSet = ts
							} else {
								currentTupleSet = make(sysfact.TupleSet, 0)
							}

							currentTupleSet = append(currentTupleSet, sysfact.Tuple{
								Key:           key,
								Value:         v,
								NormalizedKey: strings.Join(normalizedPath, `.`),
							})

							modifiedTuples[tagsetKey] = currentTupleSet
						} else {
							log.Warningf("Failed to convert value for %s: %v", key, err)
						}
					} else if tags, ok := tagsets[tagsetKey]; ok {
						// non-numeric values become tags
						tagKey := strings.Join(normalizedPath[len(normalizedPath)-2:len(normalizedPath)], `_`)

						tags[tagKey] = value
					}
				}

			}

			return nil
		}); err == nil {
			// spew.Dump(tagsets)

			for tagsetKey, tupleset := range modifiedTuples {
				for _, tuple := range tupleset {
					for key, tags := range tagsets {
						if strings.HasPrefix(tagsetKey, key) {
							tuple.Tags = maputil.Append(tuple.Tags, tags)
						}
					}

					tuples = append(tuples, tuple)
				}
			}

			sort.Sort(tuples)

			if url := c.String(`http-url`); url == `` {
				writeWithFormat(os.Stdout, c.String(`format`), tuples, tags)
			} else {
				timeout := c.Duration(`http-timeout`)
				client := http.Client{
					Timeout: timeout,
				}

				var buffer bytes.Buffer

				writeWithFormat(&buffer, c.String(`format`), tuples, tags)

				if req, err := http.NewRequest(
					strings.ToUpper(c.String(`http-method`)),
					url,
					&buffer,
				); err == nil {
					for _, kv := range c.StringSlice(`http-header`) {
						parts := strings.SplitN(kv, `=`, 2)

						if len(parts) == 2 {
							req.Header.Set(parts[0], parts[1])
						}
					}

					qs := req.URL.Query()

					for _, kv := range c.StringSlice(`http-query`) {
						parts := strings.SplitN(kv, `=`, 2)

						if len(parts) == 2 {
							qs.Set(parts[0], parts[1])
						}
					}

					req.URL.RawQuery = qs.Encode()

					log.Infof("Performing request: %v %v", req.Method, req.URL)

					if response, err := client.Do(req); err == nil {
						if response.StatusCode < 400 {
							log.Infof("Request completed successfully (%s)", response.Status)
						} else {
							log.Fatalf("Request failed: %s", response.Status)
						}
					} else {
						log.Fatal(err)
					}
				} else {
					log.Fatal(err)
				}
			}
		} else {
			log.Fatal(err)
		}
	}

	app.Commands = []cli.Command{
		{
			Name:  `interpolate`,
			Usage: `Interpolate the string the given arguments or standard input using the system report.`,
			Action: func(c *cli.Context) {
				var input string

				if fileutil.IsTerminal() {
					input = strings.Join(c.Args(), ` `)
				} else if in, err := ioutil.ReadAll(os.Stdin); err == nil {
					input = string(in)
				} else {
					log.Fatal(err)
				}

				if input == `` {
					return
				}

				if report, err := reporter.Report(); err == nil {
					fmt.Println(maputil.Sprintf(input, report))
				} else {
					log.Fatal(err)
				}
			},
		}, {
			Name:      `render`,
			Usage:     `Render the file or standard input as a template.`,
			ArgsUsage: `[FILENAME]`,
			Action: func(c *cli.Context) {
				var template string

				if filename := c.Args().First(); filename != `` {
					template = fileutil.MustReadAllString(filename)
				} else if in, err := ioutil.ReadAll(os.Stdin); err == nil {
					template = string(in)
				} else {
					log.Fatal(err)
				}

				if template == `` {
					return
				}

				if report, err := reporter.Report(); err == nil {
					report, _ = maputil.DiffuseMap(report, `.`)

					if rendered, err := sysfact.RenderString(report, template); err == nil {
						fmt.Print(rendered)
					} else {
						log.Fatal(err)
					}
				} else {
					log.Fatal(err)
				}
			},
		}, {
			Name:  `apply`,
			Usage: `Recursively copy a given source directory over top of a destination directory, selectively treating filenames starting with "@" as text templates that are given a Sysfact report as input.`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  `srcdir, s`,
					Usage: `The source directory tree containing the files to render.`,
					Value: executil.RootOrString(
						`/etc/sysfact/apply`,
						`~/.config/sysfact/apply`,
					),
				},
				cli.StringFlag{
					Name:  `destdir, d`,
					Usage: `The destination directory tree where files will be copied/rendered to.`,
					Value: executil.RootOrString(
						`/`,
						`~`,
					),
				},
				cli.BoolFlag{
					Name:  `dry-run, n`,
					Usage: `Don't actually modify the destination directory, just report what would be done.`,
				},
				cli.BoolFlag{
					Name:  `follow-links, L`,
					Usage: `Follow symlinks and copy the contents of the linked file instead of re-creating symlinks in the destination.`,
				},
				cli.StringSliceFlag{
					Name:  `srcdir-pattern, P`,
					Usage: `Provide an additional pattern string to use when searching for files to copy from srcdir.`,
				},
			},
			Action: func(c *cli.Context) {
				var opts sysfact.RenderOptions

				opts.DestDir = c.String(`destdir`)
				opts.DryRun = c.Bool(`dry-run`)
				opts.AdditionalPatterns = c.StringSlice(`srcdir-pattern`)
				opts.FollowSymlinks = c.Bool(`follow-links`)

				log.FatalIf(sysfact.Render(c.String(`srcdir`), &opts))
			},
		},
	}

	app.Run(os.Args)
}

func writeWithFormat(w io.Writer, format string, tuples sysfact.TupleSet, tags map[string]interface{}) {
	now := time.Now()

	switch format {
	case `flat`:
		for _, tuple := range tuples {
			fmt.Fprintf(w, "%s=%v\n", tuple.Key, tuple.Value)
		}

	case `env`:
		for _, tuple := range tuples {
			k := stringutil.Underscore(tuple.Key)
			k = strings.ToUpper(k)

			fmt.Fprintf(w, "%s=%v\n", k, tuple.Value)
		}

	case `values`:
		for _, tuple := range tuples {
			fmt.Fprintf(w, "%v\n", tuple.Value)
		}

	case `yaml`:
		if data, err := yaml.Marshal(tuples.ToMap(false)); err == nil {
			fmt.Fprintln(w, (string(data[:])))
		}

	case `json`:
		if data, err := json.MarshalIndent(tuples.ToMap(false), ``, `  `); err == nil {
			fmt.Fprintln(w, string(data[:]))
		}

	case `graphite`:
		epoch := now.Unix()

		for _, tuple := range tuples {
			if value, err := toFloat(tuple.Value); err == nil {
				fmt.Fprintf(w, "%s %f %d\n", tuple.Key, value, epoch)
			} else {
				log.Notice(err)
			}
		}

	case `tsdb`:
		epochMs := int64(now.UnixNano() / 1000000)
		tags := strings.TrimSpace(` ` + maputil.Join(tags, `=`, ` `))

		for _, tuple := range tuples {
			if value, err := toFloat(tuple.Value); err == nil {
				fmt.Fprintf(w, "put %s %d %f%s\n", tuple.Key, epochMs, value, tags)
			} else {
				log.Notice(err)
			}
		}

	case `influxdb`:
		var writer sysfact.InfluxdbPayload

		if out, err := writer.Generate(tuples, tags, &now); err == nil {
			fmt.Fprintln(w, out)
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
