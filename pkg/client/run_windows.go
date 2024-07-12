package client

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/lucheng0127/virtuallan/pkg/cipher"
	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

func Run(cCtx *cli.Context) error {
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
	signal.Notify(sigChan, os.Interrupt)
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
				return fmt.Errorf("read from conn %s", err)
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

	iface, err := utils.NewTap(ipAddr)
	if err != nil {
		return err
	}
	c.Iface = iface

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
