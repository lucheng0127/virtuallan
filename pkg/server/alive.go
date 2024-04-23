package server

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Endpoints struct {
	IP   string
	Addr string
	Beat chan string
}

var EPMap map[string]*Endpoints

func init() {
	EPMap = make(map[string]*Endpoints, 1024)
}

func (ep *Endpoints) Close() {
	client, ok := UPool[ep.Addr]
	if ok {
		client.Close()
	}

	delete(EPMap, ep.IP)
}

func (ep *Endpoints) Countdown() {
	for {
		select {
		case <-ep.Beat:
			continue
		case <-time.After(50 * time.Second):
			log.Infof("endpoint %s with raddr %s don't get keepalive pkt for long time, close it\n", ep.IP, ep.Addr)
			ep.Close()
			return
		}
	}
}

func GetOrCreateEp(ip, raddr string) (*Endpoints, error) {
	ep, ok := EPMap[ip]
	if ok {
		if ep.Addr != raddr {
			return nil, fmt.Errorf("ip %s used by other endpoint", ip)
		}

		return ep, nil
	}

	ep = new(Endpoints)
	ep.IP = ip
	ep.Addr = raddr
	ep.Beat = make(chan string)
	EPMap[ip] = ep

	go ep.Countdown()
	return ep, nil
}

func HandleKeepalive(ip, raddr string) error {
	_, ok := UPool[raddr]
	if !ok {
		return fmt.Errorf("unauthed client")
	}

	log.Debugf("handle keepalive pkt for %s with ip %s", raddr, ip)

	ep, err := GetOrCreateEp(ip, raddr)
	if err != nil {
		return err
	}

	ep.Beat <- "ok"
	return nil
}
