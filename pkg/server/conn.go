package server

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/lucheng0127/virtuallan/pkg/packet"
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

func (client *UClient) Handle() {
	go func() {
		for {
			pkt := <-client.NetToIface
			if pkt.Type != packet.P_RAW {
				continue
			}

			stream, err := pkt.VLBody.Encode()
			if err != nil {
				fmt.Println(err)
				continue
			}

			_, err = client.Iface.Write(stream)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}()

	for {
		var buf [1500]byte

		n, err := client.Iface.Read(buf[:])
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

		_, err = client.Conn.WriteToUDP(stream, client.RAddr)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
