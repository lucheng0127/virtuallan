package server

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

func HandleKeepalive(ipAddr, raddr string, svc *Server) error {
	c, ok := UPool[raddr]
	if !ok {
		return fmt.Errorf("unauthed client")
	}

	log.Debugf("handle keepalive pkt for %s with ip %s", raddr, ipAddr)

	// TODO: IP maybe conflict, use dhcp
	ip := net.ParseIP(ipAddr).To4()

	if !svc.IPInPool(ip) {
		c.IP = ip
		ipIdx := svc.IdxFromIP(ip)
		svc.MLock.Lock()
		svc.UsedIP = append(svc.UsedIP, ipIdx)
		svc.MLock.Unlock()

		go c.Countdown()
	}

	c.Beat <- "ok"

	return nil
}
