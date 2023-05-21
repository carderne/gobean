package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/carderne/gobean/api"
	"github.com/carderne/gobean/bean"
	"github.com/urfave/cli/v2"
)

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
							if bean.Debug {
								panic(v)
							} else {
								fmt.Println("gobean crashed: ", v)
							}
						}
					}()
					path := cCtx.Args().First()
					bean.GetBalances(path)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
