package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/songgao/water"
	"github.com/urfave/cli/v2"
	"github.com/vishvananda/netlink"
	"golang.org/x/net/ipv4"
)

const (
	PORT         = 8000
	ADDR         = "10.67.0.254"
	SIP          = "10.69.0.254/24"
	MAX_PKT_SIZE = 4096
)

var EP map[string]net.Conn
var idx = 1

func init() {
	EP = make(map[string]net.Conn)
}

func handleConn(conn net.Conn, iface *water.Interface) {
	EP[fmt.Sprintf("10.69.0.%d", idx)] = conn
	idx += 1

	go func() {
		for {
			buf := make([]byte, MAX_PKT_SIZE)
			n, err := iface.Read(buf)
			if err != nil {
				fmt.Println(err)
				continue
			}

			h, err := ipv4.ParseHeader(buf[:n])
			if err != nil {
				fmt.Println(err)
				continue
			}
			dst := h.Dst.String()
			fmt.Printf("iface dst %s\n", dst)
			p, ok := EP[dst]
			if !ok {
				fmt.Printf("pkt from server dst %s no route\n", dst)
				continue
			}

			_, err = p.Write(buf[:n])
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}()

	for {
		buf := make([]byte, MAX_PKT_SIZE)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}

		h, err := ipv4.ParseHeader(buf[:n])
		if err != nil {
			fmt.Println(err)
			continue
		}
		dst := h.Dst.String()
		fmt.Printf("conn dst %s\n", dst)

		p, ok := EP[dst]
		if !ok {
			fmt.Printf("pkt from client dst %s send to loacal", dst)
			_, err = iface.Write(buf[:n])
			if err != nil {
				fmt.Println(err)
				continue
			}
		} else {
			_, err = p.Write(buf[:n])
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
}

func setupTun(addr string) (*water.Interface, error) {
	tname := "tun0"
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = tname

	iface, err := water.New(config)
	if err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(tname)
	if err != nil {
		return nil, err
	}

	a, err := netlink.ParseAddr(addr)
	if err != nil {
		return nil, err
	}

	err = netlink.AddrAdd(link, a)
	if err != nil {
		return nil, err
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		return nil, err
	}

	return iface, nil
}

func sRun(cCtx *cli.Context) error {
	ln, err := net.Listen("tcp4", fmt.Sprintf(":%d", PORT))
	if err != nil {
		return err
	}
	fmt.Println("server listen on port 8000")

	iface, err := setupTun(SIP)
	if err != nil {
		return err
	}
	fmt.Printf("create tun %s\n", iface.Name())

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go handleConn(conn, iface)
	}
}

func cRun(cCtx *cli.Context) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ADDR, PORT))
	if err != nil {
		return err
	}
	fmt.Println("client dial 8000")

	iface, err := setupTun(cCtx.String("addr"))
	if err != nil {
		return err
	}

	for {
		go io.Copy(iface, conn)
		io.Copy(conn, iface)
	}
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "server",
				Action: sRun,
			},
			{
				Name:   "client",
				Action: cRun,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "addr",
						Aliases:  []string{"a"},
						Required: true,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
