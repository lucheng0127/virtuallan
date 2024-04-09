package server

import (
	"fmt"
	"net"

	"github.com/lucheng0127/virtuallan/pkg/config"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	"github.com/urfave/cli/v2"
	"github.com/vishvananda/netlink"
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
		var buf [1500]byte
		n, addr, err := ln.ReadFromUDP(buf[:])
		if err != nil {
			return err
		}

		// TODO(shawnlu): Auth client first
		client, err := svc.GetClientForAddr(addr, ln)
		if err != nil {
			return err
		}

		client.Iface.Write(buf[:n])

		go func() {
			for {
				var buf [1500]byte
				n, err := client.Iface.Read(buf[:])
				if err != nil {
					fmt.Println(err)
					continue
				}

				ln.WriteToUDP(buf[:n], client.RAddr)
			}
		}()

	}
}

func Run(cCtx *cli.Context) error {
	svc := new(Server)

	cfg, err := config.LoadConfigFile(cCtx.String("config-file"))
	if err != nil {
		return err
	}

	svc.ServerConfig = cfg

	return svc.ListenAndServe()
}
