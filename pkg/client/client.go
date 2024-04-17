package client

import (
	"fmt"
	"net"

	"github.com/lucheng0127/virtuallan/pkg/packet"
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

	netToIface := make(chan *packet.VLPkt, 1024)
	// TODO(shawnlu): Add keepalive

	go func() {
		go func() {
			for {
				pkt := <-netToIface
				if pkt.Type != packet.P_RAW {
					continue
				}

				stream, err := pkt.VLBody.Encode()
				if err != nil {
					fmt.Println(err)
					continue
				}

				_, err = iface.Write(stream)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
		}()

		for {
			var buf [1500]byte

			n, err := iface.Read(buf[:])
			if err != nil {
				fmt.Println(err)
				continue
			}

			pkt := packet.NewRawPkt(buf[:n])
			stream, err := pkt.Encode()
			if err != nil {
				fmt.Println(err)
				continue
			}

			_, err = conn.Write(stream)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}()

	for {
		var buf [1502]byte
		n, _, err := conn.ReadFromUDP(buf[:])
		if err != nil {
			return err
		}

		pkt := packet.NewRawPkt(buf[2:n])

		netToIface <- pkt
	}
}
