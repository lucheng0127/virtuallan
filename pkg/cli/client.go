package cli

import (
	"os"

	"github.com/lucheng0127/virtuallan/pkg/client"
	"github.com/urfave/cli/v2"
)

type ClientOpts func(*ClientFlags)

type ClientFlags struct {
	target   string
	user     string
	passwd   string
	key      string
	logLevel string
}

func SetClientTarget(t string) ClientOpts {
	return func(c *ClientFlags) {
		c.target = t
	}
}

func SetClientUser(u string) ClientOpts {
	return func(c *ClientFlags) {
		c.user = u
	}
}

func SetClientPasswd(p string) ClientOpts {
	return func(c *ClientFlags) {
		c.passwd = p
	}
}

func SetClientKey(k string) ClientOpts {
	return func(c *ClientFlags) {
		c.key = k
	}
}

func SetClientLogLevel(l string) ClientOpts {
	return func(c *ClientFlags) {
		c.logLevel = l
	}
}

func RunClient(opts ...ClientOpts) error {
	clientFlags := new(ClientFlags)
	for _, opt := range opts {
		opt(clientFlags)
	}

	app := &cli.App{
		Commands: []*cli.Command{
			NewClientCmd(),
		},
	}

	return app.Run([]string{os.Args[0], "client", "-t", clientFlags.target, "-u", clientFlags.user, "-p", clientFlags.passwd, "-k", clientFlags.key, "-l", clientFlags.logLevel})
}

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
			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				Usage:       "log level",
				DefaultText: "info",
			},
		},
	}
}
