package cli

import (
	"github.com/lucheng0127/virtuallan/pkg/client"
	"github.com/urfave/cli/v2"
)

func NewClientCmd() *cli.Command {
	return &cli.Command{
		Name:   "client",
		Usage:  "connect to virtuallan server",
		Action: client.Run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "target",
				Usage:    "socket virtuallan server listened on",
				Required: true,
			},
		},
	}
}
