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
