package utils

import "net"

func ValidatePort(port int) bool {
	if port > 1024 && port < 65535 {
		return true
	}

	return false
}

func ValidateIPv4Addr(addr string) bool {
	// Addr format without netmask
	ip, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		return false
	}

	v4ip := ip.IP.To4()
	return v4ip != nil
}
