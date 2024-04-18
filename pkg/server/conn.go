package server

import (
	"net"
	"os"
	"sync"

	"github.com/lucheng0127/virtuallan/pkg/packet"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
)

type UClient struct {
	Conn       *net.UDPConn
	RAddr      *net.UDPAddr
	Iface      *water.Interface
	NetToIface chan *packet.VLPkt
	Once       sync.Once
}

var UPool map[string]*UClient

func init() {
	UPool = make(map[string]*UClient)
}

func (client *UClient) HandleOnce() {
	client.Once.Do(client.Handle)
}

func (client *UClient) Close() {
	// When don't get keepalive for several times, close it
	// TODO(shawnlu): When read from conn failed, remove it from UPool and delete tap interface
}

func (client *UClient) Handle() {
	go func() {
		for {
			pkt := <-client.NetToIface
			if pkt.Type != packet.P_RAW {
				continue
			}

			stream, err := pkt.VLBody.Encode()
			if err != nil {
				log.Warn("encode raw vlpkt body failed: ", err)
				continue
			}

			_, err = client.Iface.Write(stream)
			if err != nil {
				log.Errorf("write to tap %s %s\n", client.Iface.Name(), err.Error())
				continue
			}
		}
	}()

	for {
		var buf [1500]byte

		n, err := client.Iface.Read(buf[:])
		if err != nil {
			log.Errorf("read from tap %s %s\n", client.Iface.Name(), err.Error())
			continue
		}

		pkt := packet.NewRawPkt(buf[:n])
		stream, err := pkt.Encode()
		if err != nil {
			log.Warn("encode raw vlpkt failed: ", err)
			continue
		}

		_, err = client.Conn.WriteToUDP(stream, client.RAddr)
		if err != nil {
			// If send failed it means udp server got something wrong, exit
			log.Errorf("send udp stream to %s %s\n", client.Conn.RemoteAddr().String(), err.Error())
			os.Exit(1)
		}
	}
}
