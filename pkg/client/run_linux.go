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

	// Use errChan capture goroutine error
	errChan := make(chan error)

	// Handle signal
	sigChan := make(chan os.Signal, 8)
	signal.Notify(sigChan, unix.SIGTERM, unix.SIGINT)
	go func() {
		if err := c.HandleSignal(sigChan); err != nil {
			errChan <- err
		}
	}()

	// Do auth
	ipChan := make(chan string)
	netToIface := make(chan *packet.VLPkt, 1024)

	// Handle udp packet
	go func() {
		for {
			var buf [65535]byte
			n, _, err := conn.ReadFromUDP(buf[:])

			if err != nil {
				errChan <- fmt.Errorf("read from conn %s", err.Error())
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
					errChan <- errors.New("auth failed")
				case packet.RSP_IP_NOT_MATCH:
					errChan <- errors.New("ip not match")
				case packet.RSP_USER_LOGGED:
					errChan <- errors.New("user already logged by other endpoint")
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
	}()

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
	go func() {
		if err := checkLoginTimeout(authChan); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case ipAddr := <-ipChan:
		// Waiting for dhcp ip
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
		go func() {
			if err := packet.MonitorRouteMulticast(tapIface, strings.Split(c.IPAddr, "/")[0]); err != nil {
				errChan <- err
			}
		}()

		// Send keepalive
		go func() {
			if err := c.DoKeepalive(10); err != nil {
				errChan <- err
			}
		}()

		// Switch io between udp net and tap interface
		go func() {
			if err := c.HandleConn(netToIface); err != nil {
				errChan <- err
			}
		}()

		err = <-errChan
		return err
	}
}
