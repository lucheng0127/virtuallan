package client

import (
	"io"
	"net"

	"github.com/lucheng0127/virtuallan/pkg/utils"
	"github.com/songgao/water"
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

	config := new(water.Config)
	config.DeviceType = water.TAP
	config.Name = "tap0"

	iface, err := water.New(*config)
	if err != nil {
		return err
	}

	if err = utils.AsignAddrToLink("tap0", "192.168.123.2/24", true); err != nil {
		return err
	}

	for {
		go io.Copy(iface, conn)
		io.Copy(conn, iface)
	}
}
