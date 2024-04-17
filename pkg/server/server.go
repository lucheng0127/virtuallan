package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/lucheng0127/virtuallan/pkg/config"
	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	"github.com/urfave/cli/v2"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type Server struct {
	*config.ServerConfig
}

func (svc *Server) SetupLan() error {
	// Create bridge
	la := netlink.NewLinkAttrs()
	la.Name = svc.Bridge
	br := &netlink.Bridge{LinkAttrs: la}

	err := netlink.LinkAdd(br)
	if err != nil {
		return err
	}

	// Add ip and set up
	err = utils.AsignAddrToLink(la.Name, svc.IP, true)
	if err != nil {
		return err
	}

	return nil
}

func (svc *Server) GetClientForAddr(addr *net.UDPAddr, conn *net.UDPConn) (*UClient, error) {
	client, ok := UPool[addr.String()]
	if ok {
		return client, nil
	}

	iface, err := utils.NewTap(svc.Bridge)
	if err != nil {
		return nil, err
	}

	client = new(UClient)
	client.Iface = iface
	client.RAddr = addr
	client.Conn = conn
	client.NetToIface = make(chan *packet.VLPkt, 1024)
	client.Once = sync.Once{}
	UPool[addr.String()] = client
	return client, nil
}
func (svc *Server) ListenAndServe() error {
	if !config.ValidatePort(svc.Port) {
		return fmt.Errorf("invalidate port %d", svc.Port)
	}

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("0.0.0.0:%d", svc.Port))
	if err != nil {
		return err
	}

	ln, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	if err = svc.SetupLan(); err != nil {
		return err
	}

	for {
		// Max vlpkt len 1502 = 1500(max ethernet pkt) + 2(vlheader)
		var buf [1502]byte
		n, addr, err := ln.ReadFromUDP(buf[:])
		if err != nil {
			return err
		}

		// TODO(shawnlu): Auth client first
		client, err := svc.GetClientForAddr(addr, ln)
		if err != nil {
			return err
		}

		go client.HandleOnce()

		pkt := packet.NewRawPkt(buf[2:n])

		client.NetToIface <- pkt
	}
}

func (svc *Server) Teardown() {
	err := utils.DelLinkByName(svc.Bridge)
	if err != nil {
		fmt.Printf("Delete bridge %s %s\n", svc.Bridge, err.Error())
	}

	os.Exit(0)
}

func (svc *Server) HandleSignal(sigChan chan os.Signal) {
	sig := <-sigChan
	fmt.Printf("Received signal: %v, stop server\n", sig)
	svc.Teardown()
}

func Run(cCtx *cli.Context) error {
	svc := new(Server)

	cfg, err := config.LoadConfigFile(config.GetCfgPath(cCtx.String("config-dir")))
	if err != nil {
		return err
	}

	svc.ServerConfig = cfg

	// Handle signel to delete bridge
	sigChan := make(chan os.Signal, 8)
	signal.Notify(sigChan, unix.SIGTERM, unix.SIGINT)
	go svc.HandleSignal(sigChan)

	return svc.ListenAndServe()
}
