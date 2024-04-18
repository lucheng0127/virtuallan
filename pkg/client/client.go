package client

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func getLoginInfo() (string, string, error) {
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
		u, p, err := getLoginInfo()
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

	// TODO(shawnlu): Do auth
	fmt.Printf("Auth with user: %s passwd: %s ip: %s\n", user, passwd, cCtx.String("addr"))

	iface, err := utils.NewTap("")
	if err != nil {
		return err
	}

	if err = utils.AsignAddrToLink(iface.Name(), cCtx.String("addr"), true); err != nil {
		return err
	}

	netToIface := make(chan *packet.VLPkt, 1024)

	// Send keepalive
	go DoKeepalive(conn, strings.Split(cCtx.String("addr"), "/")[0], 10)

	// Switch io between udp net and tap interface
	go HandleConn(iface, netToIface, conn)

	// Handle udp packet
	for {
		var buf [1502]byte
		n, _, err := conn.ReadFromUDP(buf[:])

		if err != nil {
			return err
		}

		if n < 2 {
			continue
		}

		headerType := binary.BigEndian.Uint16(buf[:2])

		switch headerType {
		case packet.P_RAW:
			pkt := packet.NewRawPkt(buf[2:n])
			netToIface <- pkt
		default:
			log.Debug("unknow stream, do nothing")
			continue
		}
	}
}
