package utils

import (
	"fmt"
	"math/rand"
	"net"
	"os/exec"

	"github.com/songgao/water"
)

const (
	UNKNOW_IP = "UNKNOW_IP"
)

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

//func SetMACAddress(interfaceName, macAddress string) error {
//	// Disable the network interface
//	disableCmd := exec.Command("netsh", "interface", "set", "interface", interfaceName, "admin=disable")
//	if err := disableCmd.Run(); err != nil {
//		return fmt.Errorf("failed to disable interface: %v", err)
//	}
//
//	// Set the new MAC address
//	setMacCmd := exec.Command("netsh", "interface", "set", "interface", interfaceName, "newmac", macAddress)
//	if err := setMacCmd.Run(); err != nil {
//		return fmt.Errorf("failed to set MAC address: %v", err)
//	}
//
//	// Enable the network interface
//	enableCmd := exec.Command("netsh", "interface", "set", "interface", interfaceName, "admin=enable")
//	if err := enableCmd.Run(); err != nil {
//		return fmt.Errorf("failed to enable interface: %v", err)
//	}
//
//	return nil
//}

func NewTap(ipAddr string) (*water.Interface, error) {
	// Create tap
	config := new(water.Config)
	config.DeviceType = water.TAP
	config.PlatformSpecificParams = water.PlatformSpecificParams{
		ComponentID: "tap0901",
	}

	iface, err := water.New(*config)
	if err != nil {
		return nil, fmt.Errorf("create tap %s", err.Error())
	}

	ip, ipNet, err := net.ParseCIDR(ipAddr)
	if err != nil {
		return nil, fmt.Errorf("parse cidr %s %s", ipAddr, err.Error())
	}

	// Add ip to tap
	netMask := fmt.Sprintf("%d.%d.%d.%d", ipNet.Mask[0], ipNet.Mask[1], ipNet.Mask[2], ipNet.Mask[3])
	cmd := exec.Command("netsh", "interface", "ip", "set", "address", fmt.Sprintf("name=%s", iface.Name()), "static", ip.String(), netMask)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("assign ip %s to %s %s", ipAddr, iface.Name(), err.Error())
	}

	// Set tap mac
	//mac := GetMacFromIP(ip)
	//if err := SetMACAddress(iface.Name(), mac.String()); err != nil {
	//	return nil, err
	//}
	return iface, nil
}
