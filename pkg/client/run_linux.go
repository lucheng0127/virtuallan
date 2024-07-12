package client

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/erikdubbelboer/gspt"
	"github.com/lucheng0127/virtuallan/pkg/cipher"
	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
)

func Run(cCtx *cli.Context) error {
	// Hide process arguments, it contains too many infos
	gspt.SetProcTitle(os.Args[0] + " client")

	// Parse args
	logLevel := cCtx.String("log-level")
	key := cCtx.String("key")
	target := cCtx.String("target")

	var user, passwd string

	if cCtx.String("passwd") == "" || cCtx.String("user") == "" {
		u, p, err := GetLoginInfo()
		if err != nil {
			return err
		}

		user = u
		passwd = p
	} else {
		user = cCtx.String("user")
		passwd = cCtx.String("passwd")
	}

	client := NewClient(
		ClientSetKey(key),
		ClientSetTarget(target),
		ClientSetLogLevel(logLevel),
		ClientSetUser(user),
		ClientSetPasswd(passwd),
	)

	return client.Launch()
}

func (c *Client) Launch() error {
	// Set AES key
	if err := cipher.SetAESKey(c.key); err != nil {
		return err
	}

	// Set log
	c.SetLogLevel()

	// Connect to server
	udpAddr, err := net.ResolveUDPAddr("udp4", c.target)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return err
	}

	c.Conn = conn

	// Use errgroup check goruntine error
	g := new(errgroup.Group)

	// Handle signal
	sigChan := make(chan os.Signal, 8)
	signal.Notify(sigChan, unix.SIGTERM, unix.SIGINT)
	g.Go(func() error {
		return c.HandleSignal(sigChan)
	})

	// Do auth
	ipChan := make(chan string)
	netToIface := make(chan *packet.VLPkt, 1024)

	// Handle udp packet
	g.Go(func() error {
		for {
			var buf [65535]byte
			n, _, err := conn.ReadFromUDP(buf[:])

			if err != nil {
				return fmt.Errorf("read from conn %s", err.Error())
			}

			if n < 2 {
				continue
			}

			pkt, err := packet.Decode(buf[:n])
			if err != nil {
				log.Error("parse packet ", err)
				continue
			}

			switch pkt.Type {
			case packet.P_RESPONSE:
				switch pkt.VLBody.(*packet.RspBody).Code {
				case packet.RSP_AUTH_REQUIRED:
					return errors.New("auth failed")
				case packet.RSP_IP_NOT_MATCH:
					return errors.New("ip not match")
				case packet.RSP_USER_LOGGED:
					return errors.New("user already logged by other endpoint")
				default:
					continue
				}
			case packet.P_RAW:
				netToIface <- pkt
			case packet.P_DHCP:
				ipAddr := pkt.VLBody.(*packet.KeepaliveBody).Parse()
				ipChan <- ipAddr
			default:
				log.Debug("unknow stream, do nothing")
				continue
			}
		}
	})

	// Auth
	authPkt := packet.NewAuthPkt(c.user, c.password)
	authStream, err := authPkt.Encode()
	if err != nil {
		return fmt.Errorf("encode auth packet %s", err.Error())
	}

	_, err = conn.Write(authStream)
	if err != nil {
		return fmt.Errorf("send auth packet %s", err.Error())
	}

	authChan := make(chan string, 1)
	g.Go(func() error {
		return checkLoginTimeout(authChan)
	})

	// Waiting for dhcp ip
	ipAddr := <-ipChan
	authChan <- "ok"
	log.Infof("auth with %s succeed, endpoint ip %s\n", c.user, ipAddr)
	c.IPAddr = ipAddr

	iface, err := utils.NewTap("")
	if err != nil {
		return err
	}
	c.Iface = iface

	// Set tap mac address according to ipv4 address,
	// it will make sure each ip with a fixed mac address,
	// so the arp entry will always be correct even when
	// tap interface has been recreate
	if err := utils.SetMacToTap(c.Iface.Name(), strings.Split(c.IPAddr, "/")[0]); err != nil {
		return err
	}

	if err := utils.AsignAddrToLink(c.Iface.Name(), c.IPAddr, true); err != nil {
		return err
	}

	// Add multicast route 224.0.0.1 dev tap
	tapIface, err := net.InterfaceByName(c.Iface.Name())
	if err != nil {
		return fmt.Errorf("get tap interface %s", err.Error())
	}

	if err := utils.AddMulticastRouteToIface(fmt.Sprintf("%s/32", packet.MULTICAST_ADDR), tapIface.Index); err != nil {
		return err
	}

	// XXX: Sometime when client restart too fast will not reveice the first multicast pkt
	// Monitor multicast for route bordcast
	g.Go(func() error {
		return packet.MonitorRouteMulticast(tapIface, strings.Split(c.IPAddr, "/")[0])
	})

	// Send keepalive
	go c.DoKeepalive(10)

	// Switch io between udp net and tap interface
	g.Go(func() error {
		return c.HandleConn(netToIface)
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
