package packet

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/lucheng0127/virtuallan/pkg/utils"
	log "github.com/sirupsen/logrus"
)

const (
	MULTICAST_ADDR = "224.0.0.1"
	MULTICAST_PORT = 9999
	MAX_DATA_SIZE  = 8192

	ROUTES_PREFIX = uint16(0x2b00 | 0x01)
)

func MulticastStream(data []byte) error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", MULTICAST_ADDR, MULTICAST_PORT))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}

	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}

func MonitorRouteMulticast(iface *net.Interface, tapIP string) error {
	// Monitor route multicast will run as a goruntine in client so log error but don't exit
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", MULTICAST_ADDR, MULTICAST_PORT))
	if err != nil {
		return fmt.Errorf("parse multicast addr %s", err.Error())
	}

	// Listen multicast on tap interface
	ln, err := net.ListenMulticastUDP("udp", iface, udpAddr)
	if err != nil {
		return fmt.Errorf("listen multicast address %s", err.Error())
	}

	// Read data from udp
	for {
		buf := make([]byte, MAX_DATA_SIZE)

		n, srcAddr, err := ln.ReadFromUDP(buf)
		if err != nil {
			log.Errorf("read from multicast %s", err.Error())
			continue
		}

		log.Debugf("read multicast from %s", srcAddr.String())
		// XXX: Check source addr same cidr with enpoint ip maybe

		// Strip route prefix
		if n < 2 {
			// Not validate stream
			continue
		}

		streamPrefix := binary.BigEndian.Uint16(buf[:2])
		if streamPrefix != ROUTES_PREFIX {
			// Not route stream
			continue
		}

		// Parse route stream
		routes := utils.ParseRoutesStream(buf[2:])
		log.Debugf("receive route multicast:\n%+v", routes)

		// TODO: Implement in windows
		// Sync routes, use flag replace, for unknow ip need delete
		if err := utils.SyncRoutesForIface(iface.Name, tapIP, routes); err != nil {
			return fmt.Errorf("sync route for %s %s", iface.Name, err.Error())
		}
	}
}
