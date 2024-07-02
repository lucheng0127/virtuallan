package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/erikdubbelboer/gspt"
	"github.com/lucheng0127/virtuallan/pkg/cipher"
	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/sys/unix"
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

func checkLoginTimeout(c chan string) {
	select {
	case <-c:
		return
	case <-time.After(10 * time.Second):
		log.Error("login timeout")
		os.Exit(1)
	}
}

func handleSignal(conn *net.UDPConn, sigChan chan os.Signal) {
	sig := <-sigChan
	log.Infof("received signal: %v, send fin pkt to close conn\n", sig)
	finPkt := packet.NewFinPkt()

	stream, err := finPkt.Encode()
	if err != nil {
		log.Error(err)
	}

	_, err = conn.Write(stream)
	if err != nil {
		log.Error(err)
	}

	os.Exit(0)
}

func Run(cCtx *cli.Context) error {
	// Hide process arguments, it contains too many infos
	gspt.SetProcTitle(os.Args[0] + " client")

	logLevel := cCtx.String("log-level")

	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	var user, passwd string

	if err := cipher.SetAESKey(cCtx.String("key")); err != nil {
		return err
	}

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

	// Handle signal
	sigChan := make(chan os.Signal, 8)
	signal.Notify(sigChan, unix.SIGTERM, unix.SIGINT)
	go handleSignal(conn, sigChan)

	// Do auth
	ipChan := make(chan string)
	netToIface := make(chan *packet.VLPkt, 1024)
	var wg sync.WaitGroup
	wg.Add(3)

	// Handle udp packet
	go func() {
		for {
			var buf [65535]byte
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
				case packet.RSP_IP_NOT_MATCH:
					log.Error("ip not match")
					os.Exit(1)
				case packet.RSP_USER_LOGGED:
					log.Error("user already logged by other endpoint")
					os.Exit(1)
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

	authChan := make(chan string, 1)
	go checkLoginTimeout(authChan)

	// Waiting for dhcp ip
	ipAddr := <-ipChan
	authChan <- "ok"
	log.Infof("auth with %s succeed, endpoint ip %s\n", user, ipAddr)

	iface, err := utils.NewTap("")
	if err != nil {
		return err
	}

	// Set tap mac address according to ipv4 address,
	// it will make sure each ip with a fixed mac address,
	// so the arp entry will always be correct even when
	// tap interface has been recreate
	if err := utils.SetMacToTap(iface.Name(), strings.Split(ipAddr, "/")[0]); err != nil {
		return err
	}

	if err := utils.AsignAddrToLink(iface.Name(), ipAddr, true); err != nil {
		return err
	}

	// Add multicast route 224.0.0.1 dev tap
	tapIface, err := net.InterfaceByName(iface.Name())
	if err != nil {
		return fmt.Errorf("get tap interface %s", err.Error())
	}

	if err := utils.AddMulticastRouteToIface(fmt.Sprintf("%s/32", packet.MULTICAST_ADDR), tapIface.Index); err != nil {
		return err
	}

	// XXX: Sometime when client restart too fast will not reveice the first multicast pkt
	// Monitor multicast for route bordcast
	go packet.MonitorRouteMulticast(tapIface, strings.Split(ipAddr, "/")[0])

	// Send keepalive
	go DoKeepalive(conn, ipAddr, 10)

	// Switch io between udp net and tap interface
	go HandleConn(iface, netToIface, conn)

	wg.Wait()
	return nil
}
