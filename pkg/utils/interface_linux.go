package utils

import (
	"fmt"
	"math/rand"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

func RandStr(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func NewTap(br string) (*water.Interface, error) {
	config := new(water.Config)
	config.DeviceType = water.TAP
	config.Name = fmt.Sprintf("tap-%s", RandStr(4))

	iface, err := water.New(*config)
	if err != nil {
		return nil, err
	}

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
