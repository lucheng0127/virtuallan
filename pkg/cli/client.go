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
				Aliases:  []string{"t"},
				Usage:    "socket virtuallan server listened on",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "user",
				Aliases:  []string{"u"},
				Usage:    "username of virtuallan endpoint",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "passwd",
				Aliases:  []string{"p"},
				Usage:    "password of virtuallan endpoint user",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "key",
				Aliases:  []string{"k"},
				Usage:    "encryption key of virtuallan",
				Required: true,
			},
		},
	}
}
