package multicast

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	MULTICAST_ADDR = "224.0.0.1"
	MULTICAST_PORT = 9999
	MAX_DATA_SIZE  = 8192

	ROUTES_PREFIX = uint16(0x2b00 | 0x01)
)

// TODO: Implement it

func SendToMulticastAddr(data []byte) error {
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

func MulticastRouteStream(stream []byte) error {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, ROUTES_PREFIX)
	if err != nil {
		return err
	}

	rStream := append(buf.Bytes(), stream...)
	if err := SendToMulticastAddr(rStream); err != nil {
		return err
	}

	return nil
}

func MonitorRouteMulticast(iface *net.Interface) error {
	// Listen multicast on tap interface

	// Read data from udp

	// Check source addr same cidr with enpoint ip

	// Strip route prefix

	// Parse route stream

	// Add route, ignore route exist error
	return nil
}
