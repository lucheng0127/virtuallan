package utils

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
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

// When create tap interface, the mac address will change
// when a endpoint close then reconnect to it, if there is another
// endpoint try to access this endpoint, the ip neigh entry is
// still the old, it must waiting for the ip neigh entry staled
// maybe we can generate mac address according the ip address
func SetMacToTap(name, ip string) error {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return fmt.Errorf("not validate ipv4 address %s", ip)
	}

	link, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	mac := GetMacFromIP(ipv4)

	if err := netlink.LinkSetHardwareAddr(link, mac); err != nil {
		return err
	}

	return nil
}

func NewTap(br string) (*water.Interface, error) {
	config := new(water.Config)
	config.DeviceType = water.TAP
	config.Name = fmt.Sprintf("tap-%s", RandStr(4))

	iface, err := water.New(*config)
	if err != nil {
		return nil, err
	}

	// Add into bridge if needed
	if br == "" {
		return iface, nil
	}

	err = SetLinkMaster(config.Name, br)
	if err != nil {
		return nil, err
	}

	return iface, nil
}

func AsignAddrToLink(name, addr string, up bool) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	ipAddr, err := netlink.ParseAddr(addr)
	if err != nil {
		return err
	}

	if err := netlink.AddrAdd(link, ipAddr); err != nil {
		return err
	}

	if up {
		if err = netlink.LinkSetUp(link); err != nil {
			return err
		}
	}

	return nil
}

func SetLinkMaster(iface, br string) error {
	ln, err := netlink.LinkByName(iface)
	if err != nil {
		return err
	}

	brLn, err := netlink.LinkByName(br)
	if err != nil {
		return err
	}

	if err = netlink.LinkSetMaster(ln, brLn); err != nil {
		return err
	}

	return netlink.LinkSetUp(ln)
}

func DelLinkByName(name string) error {
	ln, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	return netlink.LinkDel(ln)
}

func GetLinkStats() []LinkMessages {
	// Get all link stats
	var linkMesages []LinkMessages

	links, err := netlink.LinkList()
	if err != nil {
		return linkMesages
	}

	for _, link := range links {
		stats := link.Attrs().Statistics

		LinkMessage := LinkMessages{
			InterfaceName: link.Attrs().Name,
			RX_PKT:        stats.RxPackets,
			TX_PKT:        stats.TxPackets,
			RX_SIZE:       ConvertBytes(stats.RxBytes),
			TX_SIZE:       ConvertBytes(stats.TxBytes),
		}

		linkMesages = append(linkMesages, LinkMessage)
	}

	return linkMesages
}

func GetLinkStatsByName(name string, linkMsg []LinkMessages) (uint64, uint64, string, string) {
	for _, msg := range linkMsg {
		if msg.InterfaceName == name {
			return msg.RX_PKT, msg.TX_PKT, msg.RX_SIZE, msg.TX_SIZE
		}
	}

	return 0, 0, "", ""
}
