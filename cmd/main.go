// Package cmd is the CLI entrypoint
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/carderne/gobean/api"
	"github.com/carderne/gobean/bean"
	"github.com/urfave/cli/v2"
)

var debug bool

func init() {
	log.SetFlags(0)
	log.SetOutput(io.Discard)
	if os.Getenv("DEBUG") == "1" {
		log.SetOutput(os.Stderr)
		debug = true
	}
}

// Cmd creates the CLI interface
func Cmd() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "api",
				Aliases: []string{"a"},
				Usage:   "Run the API for a beancount file",
				Action: func(cCtx *cli.Context) error {
					path := cCtx.Args().First()
					if len(path) == 0 {
						fmt.Println("Must provide a filepath as the first arg")
						return nil
					}
					api.API(path)
					return nil
				},
			},
			{
				Name:    "balances",
				Aliases: []string{"b"},
				Usage:   "Print all account balances",
				Action: func(cCtx *cli.Context) error {
					defer func() {
						if v := recover(); v != nil {
							if debug {
								panic(v)
							} else {
								fmt.Println("gobean crashed: ", v)
							}
						}
					}()
					path := cCtx.Args().First()
					file, err := os.Open(path)
					if err != nil {
						panic(err)
					}
					defer file.Close()
					ledger := bean.NewLedger(debug)
					_, err = ledger.Load(file)
					if err != nil {
						panic(err)
					}
					date := time.Now()
					bals, err := ledger.GetBalances(date)
					if err != nil {
						panic(err)
					}
					bean.PrintAccBalances(bals)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
