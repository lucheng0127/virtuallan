package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func GetLoginInfo() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Username:")
	user, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Println("Password:")
	bytePasswd, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}

	passwd := string(bytePasswd)

	return strings.TrimSpace(user), strings.TrimSpace(passwd), nil
}

func Run(cCtx *cli.Context) error {
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

	udpAddr, err := net.ResolveUDPAddr("udp4", cCtx.String("target"))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return err
	}

	// Do auth
	ipAddr := cCtx.String("addr")
	netToIface := make(chan *packet.VLPkt, 1024)
	var wg sync.WaitGroup
	wg.Add(3)

	// Handle udp packet
	go func() {
		for {
			var buf [1520]byte
			n, _, err := conn.ReadFromUDP(buf[:])

			if err != nil {
				log.Error("read from conn ", err)
				os.Exit(1)
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
					log.Error("auth failed")
					os.Exit(1)
				case packet.RSP_IP_CONFLICET:
					log.Error("conflicet ip ", ipAddr)
					os.Exit(1)
				case packet.RSP_USER_LOGGED:
					log.Error("user already logged by other endpoint")
					os.Exit(1)
				default:
					continue
				}
			case packet.P_RAW:
				netToIface <- pkt
			default:
				log.Debug("unknow stream, do nothing")
				continue
			}
		}
	}()

	authPkt := packet.NewAuthPkt(user, passwd)
	authStream, err := authPkt.Encode()
	if err != nil {
		log.Error("encode auth packet ", err)
		os.Exit(1)
	}

	_, err = conn.Write(authStream)
	if err != nil {
		log.Error("send auth packet ", err)
		os.Exit(1)
	}

	iface, err := utils.NewTap("")
	if err != nil {
		return err
	}

	if err = utils.AsignAddrToLink(iface.Name(), ipAddr, true); err != nil {
		return err
	}

	// Send keepalive
	go DoKeepalive(conn, strings.Split(ipAddr, "/")[0], 10)

	// Switch io between udp net and tap interface
	go HandleConn(iface, netToIface, conn)

	wg.Wait()
	return nil
}
