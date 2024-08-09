package server

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/users"
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
	IP         net.IP
	Login      string
	Beat       chan string
	CloseChan  chan string
	Svc        *Server
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

	client.Svc.ReleaseIP(client.IP)
	delete(users.UserEPMap, client.User)
	delete(UPool, client.RAddr.String())

	// Multicast route when client close
	client.Svc.UpdateRoutes(client.User, utils.UNKNOW_IP)
	if err := client.Svc.MulticastRoutes(); err != nil {
		log.Errorf("mulitcast route %s", err.Error())
	}

	// Sync routes
	if err := utils.SyncRoutesForIface(client.Svc.Bridge, strings.Split(client.Svc.IP, "/")[0], client.Svc.GetRouteEntries()); err != nil {
		log.Errorf("sync route for dev %s %s", client.Svc.Bridge, err.Error())
	}

	client.CloseChan <- "FIN"
}

func (client *UClient) Countdown() {
	for {
		select {
		case <-client.CloseChan:
			log.Info("stop heartbeat monitor ", client.RAddr.String())
			return
		case <-client.Beat:
			continue
		case <-time.After(50 * time.Second):
			log.Infof("endpoint %s with raddr %s don't get keepalive pkt for long time, close it\n", client.IP, client.RAddr.String())
			client.Close()
			return
		}
	}
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
			if utils.IsTapNotExist(err) {
				log.Warnf("%s has been deleted", client.Iface.Name())
			} else {
				log.Errorf("read from tap %s %s\n", client.Iface.Name(), err.Error())
			}

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

func (svc *Server) OfferIPToClient(conn *net.UDPConn, ip string, raddr *net.UDPAddr) error {
	masklen := strings.Split(svc.IP, "/")[1]
	ipAddr := fmt.Sprintf("%s/%s", ip, masklen)

	pkt, err := packet.NewDhcpPkt(ipAddr)
	if err != nil {
		return err
	}

	stream, err := pkt.Encode()
	if err != nil {
		return err
	}

	_, err = conn.WriteToUDP(stream, raddr)
	if err != nil {
		log.Errorf("send udp stream to %s %s\n", raddr.String(), err.Error())
		os.Exit(1)
	}

	return nil
}

func (svc *Server) CreateClientForAddr(addr *net.UDPAddr, conn *net.UDPConn, username string) (*UClient, error) {
	iface, err := utils.NewTap(svc.Bridge)
	if err != nil {
		return nil, err
	}

	// Pop a ip for client
	ip, err := svc.IPForUser(username)
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
	client.Beat = make(chan string)
	client.CloseChan = make(chan string)
	client.Svc = svc
	client.IP = ip

	log.Infof("new client remote addr %s ip %s login at %s\n", client.RAddr.String(), client.IP.String(), client.Login)

	UPool[addr.String()] = client

	// Monitor client heartbeat
	go client.Countdown()

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

	log.Infof("run virtuallan server on udp port %d\n", svc.Port)
	ln, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}
	defer ln.Close()

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
			client, err := svc.CreateClientForAddr(addr, ln, u)
			if err != nil {
				log.Errorf("create authed client %s\n", err.Error())
				svc.SendResponse(ln, packet.RSP_INTERNAL_ERR, addr)
				continue
			}
			client.User = u

			// Offer IP to client
			err = svc.OfferIPToClient(ln, client.IP.String(), addr)
			if err != nil {
				log.Errorf("send dhcp for client %s with ip %s failed: %s", addr.String(), client.IP.String(), err.Error())
				svc.SendResponse(ln, packet.RSP_INTERNAL_ERR, addr)
				continue
			}

			log.Infof("client %s auth succeed", addr.String())

			// Parse nexthop user ip
			svc.UpdateRoutes(client.User, client.IP.String())
			if err := svc.MulticastRoutes(); err != nil {
				log.Errorf("mulitcast route %s", err.Error())
			}

			// Sync routes
			if err := utils.SyncRoutesForIface(svc.Bridge, strings.Split(svc.IP, "/")[0], svc.GetRouteEntries()); err != nil {
				log.Errorf("sync route for dev %s %s", svc.Bridge, err.Error())
			}
		case packet.P_KEEPALIVE:
			// Handle keepalive
			err = HandleKeepalive(pkt.VLBody.(*packet.KeepaliveBody).Parse(), addr.String(), svc)

			if err != nil {
				if utils.IsUnauthedErr(err) {
					continue
				}

				svc.SendResponse(ln, packet.RSP_IP_NOT_MATCH, addr)
				log.Warnf("heartbeat from %s %s, send ip not match response", addr.String(), err.Error())
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
		case packet.P_FIN:
			client, ok := UPool[addr.String()]
			if !ok {
				continue
			}

			log.Info("client FIN packet received ", addr.String())
			client.Close()
		default:
			log.Debug("unknow stream, do nothing")
			continue
		}
	}
}
