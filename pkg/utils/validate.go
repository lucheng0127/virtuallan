package utils

import (
	"encoding/binary"
	"net"
	"strconv"
	"strings"
)

func ValidateKey(key string) bool {
	keyLen := len([]byte(key))

	switch keyLen {
	default:
		return false
	case 16, 24, 32:
		return true
	}
}

func ValidatePort(port int) bool {
	if port > 1024 && port < 65535 {
		return true
	}

	return false
}

func ValidateIPv4Addr(addr string) bool {
	// Addr format without netmask
	return net.ParseIP(addr).To4() != nil
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

func ValidateIPv4WithNetmask(addr string) bool {
	addrInfos := strings.Split(addr, "/")

	if len(addrInfos) != 2 {
		return false
	}

	if !ValidateIPv4Addr(addrInfos[0]) {
		return false
	}

	masklen, err := strconv.Atoi(addrInfos[1])
	if err != nil {
		return false
	}

	if masklen < 0 || masklen > 32 {
		return false
	}

	return true
}

func ValidateDHCPRange(ipRange string) (net.IP, int) {
	ipRangeInfos := strings.Split(ipRange, "-")

	if len(ipRangeInfos) != 2 {
		return nil, 0
	}

	start := net.ParseIP(ipRangeInfos[0]).To4()
	end := net.ParseIP(ipRangeInfos[1]).To4()

	if start == nil || end == nil {
		return nil, 0
	}

	sInt := binary.BigEndian.Uint32(start)
	eInt := binary.BigEndian.Uint32(end)
	if eInt < sInt {
		return nil, 0
	}

	return start, int(eInt - sInt)
}
