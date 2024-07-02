package main

import (
	"os"

	vcli "github.com/lucheng0127/virtuallan/pkg/cli"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			vcli.NewClientCmd(),
			vcli.NewUserCmd(),
		},
	}

	serverCmd := vcli.NewServerCmd()
	if serverCmd != nil {
		app.Commands = append(app.Commands, serverCmd)
	}

	if err := app.Run(os.Args); err != nil {
		log.Panic(err)
		os.Exit(1)
	}
}
