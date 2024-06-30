package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/utils"
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

func (svc *Server) GetRouteEntries() map[string]string {
	return utils.ParseRoutesStream(svc.GetRouteStreams())
}

func (svc *Server) MulticastRoutes() error {
	// Add routes prefix in multicast data
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, packet.ROUTES_PREFIX); err != nil {
		return err
	}

	stream := append(buf.Bytes(), svc.GetRouteStreams()...)

	// XXX: Make 50ms delay to send multicast routes, to prevent endpoint start too fast don't receive route multicast
	time.Sleep(50 * time.Microsecond)
	if err := packet.MulticastStream(stream); err != nil {
		return err
	}

	return nil
}
