package client

import (
	"fmt"
	"net"
	"time"

	"github.com/lucheng0127/virtuallan/pkg/packet"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func (c *Client) HandleConn(netToIface chan *packet.VLPkt) error {
	g := new(errgroup.Group)

	g.Go(func() error {
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

			_, err = c.Iface.Write(stream)
			if err != nil {
				return fmt.Errorf("write to tap %s %s", c.Iface.Name(), err.Error())
			}
		}
	})

	g.Go(func() error {
		for {
			var buf [65535]byte

			n, err := c.Iface.Read(buf[:])
			if err != nil {
				log.Errorf("read from tap %s %s\n", c.Iface.Name(), err.Error())
				continue
			}

			pkt := packet.NewRawPkt(buf[:n])
			stream, err := pkt.Encode()
			if err != nil {
				log.Warn("encode raw vlpkt failed: ", err)
				continue
			}

			_, err = c.Conn.Write(stream)
			if err != nil {
				return fmt.Errorf("send udp stream to %s %s", c.Conn.RemoteAddr().String(), err.Error())
			}
		}
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
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

func (c *Client) DoKeepalive(interval int) error {
	ticker := time.NewTicker(time.Second * time.Duration(interval))

	for {
		err := SendKeepalive(c.Conn, c.IPAddr)
		if err != nil {
			return fmt.Errorf("send keepalive to %s %s", c.Conn.RemoteAddr(), err.Error())
		}
		<-ticker.C
	}
}
