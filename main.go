package main

import (
	"fmt"
	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/op/go-logging"
	"os"
	"sort"
	"strings"
)

var log = logging.MustGetLogger(`main`)

func main() {
	app := cli.NewApp()
	app.Name = `sysfact`
	app.Version = `0.0.1`
	app.EnableBashCompletion = false
	app.Flags = []cli.Flag{}

	var reporter *Reporter

	app.Before = func(c *cli.Context) error {
		logging.SetFormatter(logging.MustStringFormatter(`%{color}%{level:.4s}%{color:reset}[%{id:04d}] %{message}`))
		logging.SetLevel(logging.WARNING, `main`)
		logging.SetLevel(logging.CRITICAL, `plugins`)
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
			Action: func(c *cli.Context) {
				if c.NArg() > 0 {
					if values, err := reporter.GetReportValues(c.Args()); err == nil {
						keys := maputil.StringKeys(values)
						sort.Strings(keys)
						fields := make([]string, len(keys))

						for i, fieldName := range keys {
							if value, ok := values[fieldName]; ok {
								if str, err := stringutil.ToString(value); err == nil {
									fields[i] = str
								} else {
									fields[i] = fmt.Sprintf("!ERR<%v>!", err)
								}
							} else {

							}
						}

						fmt.Println(strings.Join(fields, "\t"))
					} else {
						log.Fatal(err)
					}
				}
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
