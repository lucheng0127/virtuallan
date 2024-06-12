package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/lucheng0127/virtuallan/pkg/utils"
)

func (svc *Server) ParseDHCPRange() error {
	ipStart, ipCount := utils.ValidateDHCPRange(svc.ServerConfig.DHCPRange)
	if ipCount == 0 {
		return fmt.Errorf("invalidate dhcp range %s, <ip start>-<ip end>", svc.ServerConfig.DHCPRange)
	}

	svc.IPStart = ipStart
	svc.IPCount = ipCount
	return nil
}

func (svc *Server) IPIdxInPool(idx int) bool {
	for _, i := range svc.UsedIP {
		if idx == i {
			return true
		}
	}

	return false
}

func (svc *Server) IPInPool(ip net.IP) bool {
	ipInt := binary.BigEndian.Uint32(ip)
	ipStartInt := binary.BigEndian.Uint32(svc.IPStart)

	idx := ipInt - ipStartInt

	return svc.IPIdxInPool(int(idx))
}

func (svc *Server) IPFromIdx(idx int) net.IP {
	ipStartInt := binary.BigEndian.Uint32(svc.IPStart)
	ipInt := ipStartInt + uint32(idx)

	ipBytes := make([]byte, 4)

	binary.BigEndian.PutUint32(ipBytes, ipInt)

	return net.IP(ipBytes)
}

func (svc *Server) IdxFromIP(ip net.IP) int {
	ipStartInt := binary.BigEndian.Uint32(svc.IPStart)
	ipInt := binary.BigEndian.Uint32(ip)

	return int(ipInt - ipStartInt)
}

func (svc *Server) PopIPFromPool() (net.IP, error) {
	for idx := 0; idx < svc.IPCount; idx++ {
		if svc.IPIdxInPool(idx) {
			continue
		}

		svc.MLock.Lock()
		svc.UsedIP = append(svc.UsedIP, idx)
		svc.MLock.Unlock()

		return svc.IPFromIdx(idx), nil
	}

	return nil, errors.New("run out of ip")
}

func (svc *Server) ReleaseIP(ip net.IP) {
	idx := svc.IdxFromIP(ip)

	targetPoolIdx := -1

	for i, ipIdx := range svc.UsedIP {
		if ipIdx == idx {
			targetPoolIdx = i
			break
		}
	}

	if targetPoolIdx == -1 {
		return
	}

	svc.MLock.Lock()
	svc.UsedIP = append(svc.UsedIP[:targetPoolIdx], svc.UsedIP[targetPoolIdx+1:]...)
	svc.MLock.Unlock()
}
