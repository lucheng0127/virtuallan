package server

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Endpoints struct {
	IP      string
	Addr    string
	Expired *time.Timer
	Once    sync.Once
}

var EPMap map[string]*Endpoints

func init() {
	EPMap = make(map[string]*Endpoints, 1024)
}

func (ep *Endpoints) UpdateExpired(n int) {
	ep.Expired = time.NewTimer(time.Second * time.Duration(n))
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
		<-ep.Expired.C
		log.Infof("endpoint %s with raddr %s don't get keepalive pkt for long time, close it\n", ep.IP, ep.Addr)
		ep.Close()
		break
	}
}

func (ep *Endpoints) CountdownOnce() {
	ep.Once.Do(ep.Countdown)
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
	ep.Once = sync.Once{}
	return ep, nil
}

func HandleKeepalive(ip, raddr string) error {
	log.Debugf("handle keepalive pkt for %s with ip %s", raddr, ip)

	ep, err := GetOrCreateEp(ip, raddr)
	if err != nil {
		return err
	}

	ep.UpdateExpired(50)
	go ep.CountdownOnce()
	return nil
}
