package server

import (
	"net"

	"github.com/urfave/cli/v2"
)

func Run(cCtx *cli.Context) error {
	ln, err := net.Listen("tcp4", "0.0.0.0:6123")
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go HandleConn(conn)
	}
}
