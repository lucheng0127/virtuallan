package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lucheng0127/virtuallan/pkg/client"
	"github.com/lucheng0127/virtuallan/pkg/users"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	"github.com/urfave/cli/v2"
)

func listUser(cCtx *cli.Context) error {
	users, err := users.ListUser(cCtx.String("db"))
	if err != nil {
		return err
	}

	fmt.Println(strings.Join(users, ","))
	return nil
}

func addUser(cCtx *cli.Context) error {
	var user, passwd string
	if cCtx.String("passwd") == "" || cCtx.String("user") == "" {
		u, p, err := client.GetLoginInfo()
		if err != nil {
			return err
		}

		user = u
		passwd = p
	} else {
		user = cCtx.String("user")
		passwd = cCtx.String("passwd")
	}

	// TODO(shawnlu): Check user exist
	if !utils.ValidateUsername(user) || !utils.ValidatePasswd(passwd) {
		return errors.New("invalidate username or passwd")
	}

	err := users.AddUser(cCtx.String("db"), user, passwd)
	if err != nil {
		return err
	}

	fmt.Println("Add user succeed")

	return nil
}

func NewUserCmd() *cli.Command {
	return &cli.Command{
		Name:  "user",
		Usage: "user management tools",
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "list users",
				Action: listUser,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db",
						Aliases:  []string{"d"},
						Usage:    "user db file loaction",
						Required: true,
					},
				},
			},
			{
				Name:   "add",
				Usage:  "add user",
				Action: addUser,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db",
						Aliases:  []string{"d"},
						Usage:    "user db file loaction",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "user",
						Aliases:  []string{"u"},
						Usage:    "username of user",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "passwd",
						Aliases:  []string{"p"},
						Usage:    "password of user",
						Required: false,
					},
				},
			},
		},
	}
}
