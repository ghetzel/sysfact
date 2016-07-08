package main

import (
	"fmt"
	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-stockutil/maputil"
	// "github.com/ghetzel/go-stockutil/stringutil"
	"github.com/op/go-logging"
	"os"
	"sort"
)

var log = logging.MustGetLogger(`main`)

type Tuple struct {
	Key   string
	Value interface{}
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
			Value: `warning`,
		},
		cli.StringFlag{
			Name:  `plugin-log-level, P`,
			Usage: `The level of logging verbosity to output when executing plugins.`,
			Value: `critical`,
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

		reporter = NewReporter()

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  `json`,
			Usage: `Retrieve one or more facts`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  `flat, f`,
					Usage: `Output the report as a single-level object instead of a deeply nested one.`,
				},
			},
			Action: func(c *cli.Context) {
				var err error

				if c.NArg() > 0 {
					err = reporter.PrintReportValues(c.Args(), c.Bool(`flat`))
				} else {
					err = reporter.PrintReport(c.Bool(`flat`))
				}

				if err != nil {
					log.Fatal(err)
				}
			},
		}, {
			Name:  `get`,
			Usage: `Retrieve one or more facts output as a tab-separated table of values`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  `format, f`,
					Usage: `How the output should be formatted.`,
					Value: `flat`,
				},
			},
			Action: func(c *cli.Context) {
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
				tuples := make([]Tuple, len(keys))

				for i, fieldName := range keys {
					if value, ok := values[fieldName]; ok {
						tuples[i] = Tuple{
							Key:   fieldName,
							Value: value,
						}
					}
				}

				printWithFormat(c.String(`format`), tuples)
			},
		}, {
			Name:  `version`,
			Usage: `Output only the version string and exit`,
			Action: func(c *cli.Context) {
				fmt.Println(c.App.Version)
			},
		},
	}

	app.Run(os.Args)
}

func printWithFormat(format string, tuples []Tuple) {
	switch format {
	case `flat`:
		for _, tuple := range tuples {
			fmt.Printf("%s=%v\n", tuple.Key, tuple.Value)
		}
	}
}
