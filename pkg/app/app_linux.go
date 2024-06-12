package app

import (
	vcli "github.com/lucheng0127/virtuallan/pkg/cli"
	"github.com/urfave/cli/v2"
)

func GetApp() *cli.App {
	return &cli.App{
		Commands: []*cli.Command{
			vcli.NewServerCmd(),
			vcli.NewClientCmd(),
			vcli.NewUserCmd(),
		},
	}
}
