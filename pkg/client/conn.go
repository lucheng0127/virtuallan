package client

import (
	"net"
	"os"
	"time"

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

func SendKeepalive(conn *net.UDPConn, addr string) error {
	pkt, err := packet.NewKeepalivePkt(addr)
	if err != nil {
		return err
	}

	stream, err := pkt.Encode()
	if err != nil {
		return err
	}

	_, err = conn.Write(stream)
	if err != nil {
		return err
	}

	return nil
}

func DoKeepalive(conn *net.UDPConn, addr string, interval int) {
	ticker := time.NewTicker(time.Second * time.Duration(interval))

	for {
		err := SendKeepalive(conn, addr)
		if err != nil {
			log.Errorf("send keepalive to %s %s", conn.RemoteAddr(), err.Error())
		}
		<-ticker.C
	}
}
