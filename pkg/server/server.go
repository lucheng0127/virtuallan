package server

import (
	"fmt"
	"net"

	"github.com/urfave/cli/v2"
)

func Run(cCtx *cli.Context) error {
	addr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:6123")
	if err != nil {
		return err
	}

	ln, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		var buf [1500]byte
		n, addr, err := ln.ReadFromUDP(buf[:])
		if err != nil {
			return err
		}

		client, err := GetClientForAddr(addr, ln)
		if err != nil {
			return err
		}

		client.Iface.Write(buf[:n])

		go func() {
			for {
				var buf [1500]byte
				n, err := client.Iface.Read(buf[:])
				if err != nil {
					fmt.Println(err)
					continue
				}

				ln.WriteToUDP(buf[:n], client.RAddr)
			}
		}()
	}
}
