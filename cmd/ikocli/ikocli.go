package main

import (
	"fmt"
	"os"

	"gopkg.in/sirupsen/logrus.v1"
	ikocli "gopkg.in/urfave/cli.v1"
)

const (
	Version = "0.1"
)

var (
	app = ikocli.NewApp()
	log = logrus.New()
)

func init() {
	app.Name = "ikocli"
	app.Usage = "KittyCash CLI is a help tool for iko-chain and kitty-api"
	app.Description = "KittyCash IKO CLI is a tool to submit kitties to iko-chain and kitty-api"
	app.Version = Version
	commands := ikocli.CommandsByName{
		initCommand(),
		editCommand(),
	}
	app.Commands = commands
	app.EnableBashCompletion = true
	app.OnUsageError = func(context *ikocli.Context, err error, isSubcommand bool) error {
		fmt.Fprintf(context.App.Writer, "error: %v\n\n", err)
		ikocli.ShowAppHelp(context)
		return nil
	}
	app.CommandNotFound = func(context *ikocli.Context, command string) {
		tmp := fmt.Sprintf("{{.HelpName}}: '%s' is not a {{.HelpName}} "+
			"command. See '{{.HelpName}} --help'. \n", command)
		ikocli.HelpPrinter(app.Writer, tmp, app)
	}
}

func initCommand() ikocli.Command {
	return ikocli.Command{
		Name:  "init",
		Usage: "Creates a file containing a list of kitties of the given index range",
		Flags: []ikocli.Flag{
			ikocli.StringFlag{
				Name:  "index-range",
				Usage: "Range of kitty indexes to generate",
			},
			ikocli.StringFlag{
				Name:  "file",
				Usage: "Output generated data to `FILE`",
			},
		},
	}
}

func editCommand() ikocli.Command {
	return ikocli.Command{
		Name:  "edit",
		Usage: "Mass edit a file containing kitty data",
		Flags: []ikocli.Flag{
			ikocli.StringFlag{
				Name:  "field",
				Usage: "Breed of the kitty to edit",
			},
			ikocli.StringFlag{
				Name:  "value",
				Usage: "Value to be set",
			},
			ikocli.StringFlag{
				Name:  "index-range",
				Usage: "Range of kitty index to edit",
			},
			ikocli.StringFlag{
				Name:  "file",
				Usage: "Read data from `FILE`",
			},
		},
	}
}

func main() {
	if e := app.Run(os.Args); e != nil {
		log.Println(e)
	}
}
