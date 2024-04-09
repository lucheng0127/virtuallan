package utils

import "github.com/vishvananda/netlink"

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
