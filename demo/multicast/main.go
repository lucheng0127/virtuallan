package main

import (
	"fmt"
	"net"
	"time"
)

const (
	SVC_ADDR      = "224.0.0.1:9000"
	MAX_DATA_SIZE = 8192
)

func ping(addr string) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		panic(err)
	}

	c, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		panic(err)
	}

	idx := 0
	for {
		idx++
		msg := fmt.Sprintf("ping %d", idx)

		_, err := c.Write([]byte(msg))
		if err != nil {
			panic(err)
		}

		time.Sleep(3 * time.Second)
	}
}

func listenMulticast(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	// Listen multicast on all network interface
	ln, err := net.ListenMulticastUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}

	for {
		buf := make([]byte, MAX_DATA_SIZE)

		n, srcAddr, err := ln.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		fmt.Printf("read %s from %s\n", string(buf[:n]), srcAddr.String())
		fmt.Println("")
	}
}

func main() {
	go ping(SVC_ADDR)
	if err := listenMulticast(SVC_ADDR); err != nil {
		panic(err)
	}
}
