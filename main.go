package main

import (
	"fmt"
	"os"

	vcli "github.com/lucheng0127/virtuallan/pkg/cli"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			vcli.NewServerCmd(),
			vcli.NewClientCmd(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
