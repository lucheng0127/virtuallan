package client

import (
	"io"
	"net"

	"github.com/lucheng0127/virtuallan/pkg/utils"
	"github.com/urfave/cli/v2"
)

func Run(cCtx *cli.Context) error {
	udpAddr, err := net.ResolveUDPAddr("udp4", cCtx.String("target"))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return err
	}

	iface, err := utils.NewTap("")
	if err != nil {
		return err
	}

	if err = utils.AsignAddrToLink(iface.Name(), cCtx.String("addr"), true); err != nil {
		return err
	}

	for {
		go io.Copy(iface, conn)
		io.Copy(conn, iface)
	}
}
