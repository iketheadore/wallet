package main

import (
	"github.com/kittycash/wallet/legacy/ex24/store"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
)

var app = cli.NewApp()

func init() {
	app.Name = "ikotools"
	app.Description = "tools for setting up initial kitty offering"
	app.Commands = cli.Commands{
		cli.Command{
			Name:  "dir",
			Usage: "tools for managing initial directory",
			Subcommands: cli.Commands{
				cli.Command{
					Name:  "init",
					Usage: "initialize dir",
					Flags: cli.FlagsByName{
						cli.StringFlag{
							Name:  "dir, d",
							Usage: "directory to initialize",
							Value: "iko_dir",
						},
						cli.IntFlag{
							Name:  "count, c",
							Usage: "kitty count for initial kitty offering",
							Value: 20,
						},
					},
					Action: func(ctx *cli.Context) error {
						return store.InitDir(
							ctx.String("dir"),
							ctx.Int("count"),
						)
					},
				},
			},
		},
	}
}

func main() {
	if e := app.Run(os.Args); e != nil {
		log.Println(e)
	}
}
