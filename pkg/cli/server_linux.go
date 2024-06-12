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
				Name:     "config-dir",
				Aliases:  []string{"d"},
				Usage:    "config directory to launch virtuallan server, conf.yaml as config file, users as user storage",
				Required: true,
			},
		},
	}
}
