package server

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/lucheng0127/virtuallan/pkg/packet"
)

func (svc *Server) AvaliabelRoutes() string {
	rawData := ""
	for _, route := range svc.ServerConfig.Routes {
		nexthopIP := svc.Routes[route.Nexthop]
		if nexthopIP == "" || nexthopIP == UNKNOW_IP {
			continue
		}

		rawData += fmt.Sprintf("%s>%s\n", route.CIDR, nexthopIP)
	}

	return rawData
}

func (svc *Server) AvaliabelRouteStreams() []byte {
	return []byte(svc.AvaliabelRoutes())
}

func (svc *Server) MulticastRoutes() error {
	// Add routes prefix in multicast data
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, packet.ROUTES_PREFIX); err != nil {
		return err
	}

	stream := append(buf.Bytes(), svc.AvaliabelRouteStreams()...)
	if err := packet.MulticastStream(stream); err != nil {
		return err
	}

	return nil
}
