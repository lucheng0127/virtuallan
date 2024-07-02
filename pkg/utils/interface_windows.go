package utils

import (
	"fmt"
	"math/rand"
	"net"
	"os/exec"
	"strings"

	"github.com/songgao/water"
)

// TODO: Implement it in windows

type LinkMessages struct {
	InterfaceName string
	RX_SIZE       string
	TX_SIZE       string
	RX_PKT        uint64
	TX_PKT        uint64
}

func RandStr(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetMacFromIP(ip net.IP) net.HardwareAddr {
	ip = ip.To4()
	return net.HardwareAddr{0x60, 0xe2, ip[0], ip[1], ip[2], ip[3]}
}

func GetIfaceByName(name string) (*net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if iface.Name == name {
			return &iface, nil
		}
	}

	return nil, fmt.Errorf("interface %s not exist", name)
}

// when a endpoint close then reconnect to it, if there is another
// endpoint try to access this endpoint, the ip neigh entry is
// still the old, it must waiting for the ip neigh entry staled
// maybe we can generate mac address according the ip address
func SetMacToTap(name, ip string) error {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return fmt.Errorf("not validate ipv4 address %s", ip)
	}

	_, err := GetIfaceByName(name)
	if err != nil {
		return err
	}

	mac := GetMacFromIP(ipv4)

	fmt.Println(mac.String())
	//if err := setWindowsMacAddress(link, mac); err != nil {
	//	return err
	//}

	return nil
}

func NewTap(ip string) (*water.Interface, error) {
	config := new(water.Config)
	config.DeviceType = water.TAP
	config.PlatformSpecificParams = water.PlatformSpecificParams{
		//InterfaceName: "TAP-Windows Adapter V9",
		//InterfaceName: fmt.Sprintf("tap-%s", RandStr(4)),
		Network:     ip,
		ComponentID: "tap0901",
		//ComponentID: "root\\tap0901",
	}

	iface, err := water.New(*config)
	if err != nil {
		return nil, err
	}

	return iface, nil
}

func AsignAddrToLink(name, addr string) error {
	cmd := exec.Command("netsh", "interface", "ip", "set", "address", fmt.Sprintf("name=%s", name), "static", strings.Split(addr, "/")[0], "255.255.255.0")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("assign ip %s to tap %s %s", addr, name, err.Error())
	}
	return nil
}
