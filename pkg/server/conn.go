package server

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
)

type UClient struct {
	Conn       *net.UDPConn
	RAddr      *net.UDPAddr
	Iface      *water.Interface
	NetToIface chan *packet.VLPkt
	Once       sync.Once
	User       string
	IP         string
	Login      string
}

var UPool map[string]*UClient

func init() {
	UPool = make(map[string]*UClient)
}

func (client *UClient) HandleOnce() {
	client.Once.Do(client.Handle)
}

func (client *UClient) Close() {
	// Remove it from UPool and delete tap interface
	log.Info("close client ", client.RAddr.String())
	if err := utils.DelLinkByName(client.Iface.Name()); err != nil {
		log.Error(err)
	}

	delete(UPool, client.RAddr.String())
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
		var buf [65535]byte

		n, err := client.Iface.Read(buf[:])
		if err != nil {
			log.Errorf("read from tap %s %s\n", client.Iface.Name(), err.Error())
			// If tap has been deleted, break it
			goto EXIT
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

EXIT:
	return
}

func (svc *Server) SendResponse(conn *net.UDPConn, code uint16, raddr *net.UDPAddr) {
	pkt := packet.NewResponsePkt(code)

	stream, err := pkt.Encode()
	if err != nil {
		log.Errorf("encode response packet %s", err.Error())
		return
	}

	_, err = conn.WriteToUDP(stream, raddr)
	if err != nil {
		log.Errorf("send udp stream to %s %s\n", raddr.String(), err.Error())
		os.Exit(1)
	}
}

func (svc *Server) CreateClientForAddr(addr *net.UDPAddr, conn *net.UDPConn) (*UClient, error) {
	iface, err := utils.NewTap(svc.Bridge)
	if err != nil {
		return nil, err
	}

	client := new(UClient)
	client.Iface = iface
	client.RAddr = addr
	client.Conn = conn
	client.NetToIface = make(chan *packet.VLPkt, 1024)
	client.Login = time.Now().Format("2006-01-02 15:04:05")
	client.Once = sync.Once{}
	UPool[addr.String()] = client
	return client, nil
}
