package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lucheng0127/virtuallan/pkg/config"
	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/users"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type Server struct {
	*config.ServerConfig
	userDb string
}

func (svc *Server) SetupLan() error {
	// Create bridge
	la := netlink.NewLinkAttrs()
	la.Name = svc.Bridge
	br := &netlink.Bridge{LinkAttrs: la}

	err := netlink.LinkAdd(br)
	if err != nil {
		return fmt.Errorf("create bridge %s %s", la.Name, err.Error())
	}

	// Add ip and set up
	err = utils.AsignAddrToLink(la.Name, svc.IP, true)
	if err != nil {
		return err
	}

	return nil
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
	client.Once = sync.Once{}
	UPool[addr.String()] = client
	return client, nil
}
func (svc *Server) ListenAndServe() error {
	if !utils.ValidatePort(svc.Port) {
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
		// for encrypted data len should be n*16(aes block size) + 16(key len)
		// so buf len should be 94 * 16 + 16 = 1520
		var buf [65535]byte
		n, addr, err := ln.ReadFromUDP(buf[:])
		if err != nil {
			return err
		}

		if n < 2 {
			continue
		}

		// For wrong AES key, will return pkt to nill or unsupported pkt error, just skip
		pkt, err := packet.Decode(buf[:n])
		if pkt == nil {
			continue
		}

		if err != nil {
			if utils.IsUnsupportedPkt(err) {
				log.Warn(err)
				continue
			}

			log.Error("parse packet ", err)
		}

		// TODO(shawnlu): Add close conn
		switch pkt.Type {
		case packet.P_AUTH:
			u, p := pkt.VLBody.(*packet.AuthBody).Parse()

			// Check user logged
			if _, ok := users.UserEPMap[u]; ok {
				svc.SendResponse(ln, packet.RSP_USER_LOGGED, addr)
				continue
			}

			// Auth user
			err = users.ValidateUser(svc.userDb, u, p)

			if err != nil {
				log.Warn(err)
				svc.SendResponse(ln, packet.RSP_AUTH_REQUIRED, addr)
				continue
			}

			users.UserEPMap[u] = addr.String()
			log.Infof("client %s login to %s succeed\n", addr.String(), u)

			// Create client for authed addr
			client, err := svc.CreateClientForAddr(addr, ln)
			if err != nil {
				log.Errorf("create authed client %s\n", err.Error())
			}
			client.User = u

			log.Infof("client %s auth succeed", addr.String())
		case packet.P_KEEPALIVE:
			// Handle keepalive
			err = HandleKeepalive(pkt.VLBody.(*packet.KeepaliveBody).Parse(), addr.String())

			if err != nil {
				if utils.IsUnauthedErr(err) {
					continue
				}

				svc.SendResponse(ln, packet.RSP_IP_CONFLICET, addr)
				log.Warnf("heartbeat from %s %s, send ip conflicet response", addr.String(), err.Error())
			}
		case packet.P_RAW:
			// Get authed client from UPool
			client, ok := UPool[addr.String()]
			if !ok {
				svc.SendResponse(ln, packet.RSP_AUTH_REQUIRED, addr)
				continue
			}

			go client.HandleOnce()

			client.NetToIface <- pkt
		default:
			log.Debug("unknow stream, do nothing")
			continue
		}
	}
}

func (svc *Server) Teardown() {
	err := utils.DelLinkByName(svc.Bridge)
	if err != nil {
		log.Errorf("delete bridge %s %s\n", svc.Bridge, err.Error())
	}

	os.Exit(0)
}

func (svc *Server) HandleSignal(sigChan chan os.Signal) {
	sig := <-sigChan
	log.Infof("received signal: %v, stop server\n", sig)
	svc.Teardown()
}

// TODO(shawnlu): Add dhcp
func Run(cCtx *cli.Context) error {
	svc := new(Server)

	cfgDir := cCtx.String("config-dir")
	cfg, err := config.LoadConfigFile(config.GetCfgPath(cfgDir))
	if err != nil {
		return err
	}

	svc.ServerConfig = cfg
	svc.userDb = filepath.Join(cfgDir, "users")

	switch strings.ToUpper(cfg.LogLevel) {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	// Run web server
	if svc.ServerConfig.WebConfig.Enable {
		webSvc := &webServe{port: svc.ServerConfig.WebConfig.Port}
		go webSvc.Serve()
		log.Info("run web server on port ", webSvc.port)
	}

	// Handle signel to delete bridge
	sigChan := make(chan os.Signal, 8)
	signal.Notify(sigChan, unix.SIGTERM, unix.SIGINT)
	go svc.HandleSignal(sigChan)

	return svc.ListenAndServe()
}
