package server

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/lucheng0127/virtuallan/pkg/packet"
)

func (svc *Server) GetRoutes() string {
	rawData := ""
	for _, route := range svc.ServerConfig.Routes {
		nexthopIP := svc.Routes[route.Nexthop]
		if nexthopIP == "" {
			continue
		}

		rawData += fmt.Sprintf("%s>%s\n", route.CIDR, nexthopIP)
	}

	return rawData
}

func (svc *Server) GetRouteStreams() []byte {
	return []byte(svc.GetRoutes())
}

func (svc *Server) MulticastRoutes() error {
	if !svc.RouteChange {
		// Route not change
		return nil
	}

	// Update svc.RouteChange flag to false, when svc.Routes change, it will set to true
	svc.MLock.Lock()
	svc.RouteChange = false
	svc.MLock.Unlock()

	// Add routes prefix in multicast data
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, packet.ROUTES_PREFIX); err != nil {
		return err
	}

	stream := append(buf.Bytes(), svc.GetRouteStreams()...)
	if err := packet.MulticastStream(stream); err != nil {
		return err
	}

	return nil
}
