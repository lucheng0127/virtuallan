package server

import (
	"fmt"
	"io"
	"net"

	"github.com/lucheng0127/virtuallan/pkg/utils"
	"github.com/songgao/water"
)

func HandleConn(conn net.Conn) {
	// Create tap
	config := new(water.Config)
	config.DeviceType = water.TAP
	config.Name = "tap0"

	iface, err := water.New(*config)
	if err != nil {
		fmt.Println(err)
	}

	if err = utils.AsignAddrToLink("tap0", "192.168.123.1/24", true); err != nil {
		fmt.Println(err)
	}

	go io.Copy(iface, conn)
	io.Copy(conn, iface)
}
