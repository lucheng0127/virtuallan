package server

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
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

func (svc *Server) ParseRoutesStream(data []byte) map[string]string {
	routes := make(map[string]string)
	reader := bytes.NewReader(data)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		rInfos := strings.Split(scanner.Text(), ">")
		if len(rInfos) == 2 {
			routes[rInfos[0]] = rInfos[1]
		}
	}

	return routes
}
