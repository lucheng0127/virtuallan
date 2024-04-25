package utils

import (
	"net"
	"strings"
)

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

func ValidatePasswd(passwd string) bool {
	if strings.Contains(passwd, " ") {
		return false
	}

	if len(passwd) < 8 || len(passwd) > 64 {
		return false
	}

	return true
}

func ValidateUsername(name string) bool {
	if strings.Contains(name, " ") {
		return false
	}

	if len(name) == 0 || len(name) > 16 {
		return false
	}

	return true
}
