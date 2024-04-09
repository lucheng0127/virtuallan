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
				Name:     "cfg-dir",
				Usage:    "config directory that contain config.yaml, server.crt, server.key for server run",
				Required: true,
			},
		},
	}
}
