package server

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

func HandleKeepalive(ipAddr, raddr string, svc *Server) error {
	c, ok := UPool[raddr]
	if !ok {
		return errors.New("unauthed client")
	}

	log.Debugf("handle keepalive pkt for %s with ip %s", raddr, ipAddr)

	if c.IP.String() != strings.Split(ipAddr, "/")[0] {
		return fmt.Errorf("client %s ip should be %s, wrong ip %s in keepalive pkt", c.RAddr.String(), c.IP.String(), ipAddr)
	}

	// Heartbeat
	c.Beat <- "ok"

	return nil
}
