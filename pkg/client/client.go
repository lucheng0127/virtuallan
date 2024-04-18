package client

import (
	"encoding/binary"
	"net"
	"strings"

	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	log "github.com/sirupsen/logrus"
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

	netToIface := make(chan *packet.VLPkt, 1024)

	// Send keepalive
	go DoKeepalive(conn, strings.Split(cCtx.String("addr"), "/")[0], 10)

	// Switch io between udp net and tap interface
	go HandleConn(iface, netToIface, conn)

	// Handle udp packet
	for {
		var buf [1502]byte
		n, _, err := conn.ReadFromUDP(buf[:])

		if err != nil {
			return err
		}

		if n < 2 {
			continue
		}

		headerType := binary.BigEndian.Uint16(buf[:2])

		switch headerType {
		case packet.P_RAW:
			pkt := packet.NewRawPkt(buf[2:n])
			netToIface <- pkt
		default:
			log.Debug("unknow stream, do nothing")
			continue
		}
	}
}
