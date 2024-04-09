package cli

import (
	"github.com/lucheng0127/virtuallan/pkg/server"
	"github.com/urfave/cli/v2"
)

func NewServerCmd() *cli.Command {
	return &cli.Command{
		Name:   "server",
		Usage:  "run virtuallan server",
		Action: server.Run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config-file",
				Aliases:  []string{"c"},
				Usage:    "config yaml to launch virtuallan server",
				Required: true,
			},
		},
	}
}
