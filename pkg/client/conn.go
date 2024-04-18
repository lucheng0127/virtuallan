package client

import (
	"net"
	"os"

	"github.com/lucheng0127/virtuallan/pkg/packet"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
)

func HandleConn(iface *water.Interface, netToIface chan *packet.VLPkt, conn *net.UDPConn) {
	go func() {
		for {
			pkt := <-netToIface
			if pkt.Type != packet.P_RAW {
				continue
			}

			stream, err := pkt.VLBody.Encode()
			if err != nil {
				log.Warn("encode raw vlpkt body failed: ", err)
				continue
			}

			_, err = iface.Write(stream)
			if err != nil {
				log.Errorf("write to tap %s %s\n", iface.Name(), err.Error())
				continue
			}
		}
	}()

	for {
		var buf [1500]byte

		n, err := iface.Read(buf[:])
		if err != nil {
			log.Errorf("read from tap %s %s\n", iface.Name(), err.Error())
			continue
		}

		pkt := packet.NewRawPkt(buf[:n])
		stream, err := pkt.Encode()
		if err != nil {
			log.Warn("encode raw vlpkt failed: ", err)
			continue
		}

		_, err = conn.Write(stream)
		if err != nil {
			log.Errorf("send udp stream to %s %s\n", conn.RemoteAddr().String(), err.Error())
			os.Exit(1)
		}
	}
}
