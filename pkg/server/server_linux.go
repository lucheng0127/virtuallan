package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/erikdubbelboer/gspt"
	"github.com/lucheng0127/virtuallan/pkg/cipher"
	"github.com/lucheng0127/virtuallan/pkg/config"
	"github.com/lucheng0127/virtuallan/pkg/packet"
	"github.com/lucheng0127/virtuallan/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type Server struct {
	*config.ServerConfig
	userDb  string
	UsedIP  []int
	IPStart net.IP
	IPCount int
	MLock   sync.Mutex
	Routes  map[string]string // Nexthop username as key, user ipv4 addr as value
}

func NewServer() *Server {
	svc := new(Server)

	svc.UsedIP = make([]int, 0)
	svc.MLock = sync.Mutex{}
	svc.Routes = make(map[string]string)

	return svc
}

func (svc *Server) SetupLan() error {
	// Create bridge
	la := netlink.NewLinkAttrs()
	la.Name = svc.Bridge
	br := &netlink.Bridge{LinkAttrs: la}

	err := netlink.LinkAdd(br)
	if err != nil {
		return fmt.Errorf("create bridge %s %s", la.Name, err.Error())
	}

	// Add ip and set up
	err = utils.AsignAddrToLink(la.Name, svc.IP, true)
	if err != nil {
		return fmt.Errorf("assign ip to bridge %s", err.Error())
	}

	// Add multicast route 224.0.0.1 dev br
	if err := utils.AddMulticastRouteToIface(fmt.Sprintf("%s/32", packet.MULTICAST_ADDR), br.Index); err != nil {
		return err
	}

	return nil
}

func (svc *Server) Teardown() {
	err := utils.DelLinkByName(svc.Bridge)
	if err != nil {
		log.Errorf("delete bridge %s %s\n", svc.Bridge, err.Error())
	}

	os.Exit(0)
}

func (svc *Server) HandleSignal(sigChan chan os.Signal) {
	sig := <-sigChan
	log.Infof("received signal: %v, stop server\n", sig)
	svc.Teardown()
}

func (svc *Server) InitRoutes() {
	for _, route := range svc.ServerConfig.Routes {
		if route.Nexthop == "SERVER" {
			// Parse route to virtuallan server
			svc.Routes[route.Nexthop] = strings.Split(svc.ServerConfig.IP, "/")[0]
		} else {
			svc.Routes[route.Nexthop] = utils.UNKNOW_IP
		}
	}
}

func (svc *Server) UpdateRoutes(nexthop, ip string) {
	nexthopIP, ok := svc.Routes[nexthop]
	if !ok {
		// Not a nexthop user
		return
	}

	if nexthopIP != ip {
		svc.MLock.Lock()
		svc.Routes[nexthop] = ip
		svc.MLock.Unlock()
	}
}

func Run(cCtx *cli.Context) error {
	// Hide process arguments, it contains too many infos
	gspt.SetProcTitle(os.Args[0] + " server")

	// New server and do cfg parse
	svc := NewServer()

	cfgDir := cCtx.String("config-dir")
	cfg, err := config.LoadConfigFile(config.GetCfgPath(cfgDir))
	if err != nil {
		return err
	}

	svc.ServerConfig = cfg
	svc.userDb = filepath.Join(cfgDir, "users")

	if err := cipher.SetAESKey(cfg.Key); err != nil {
		return err
	}

	switch strings.ToUpper(cfg.LogLevel) {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	log.SetOutput(os.Stdout)

	if !utils.ValidateIPv4WithNetmask(svc.ServerConfig.IP) {
		return fmt.Errorf("invalidate ip %s, <ip>/<netmask len>", svc.ServerConfig.IP)
	}

	if err := svc.ParseDHCPRange(); err != nil {
		return err
	}

	// Init svc.Routes with unknow ip for each route nexthop
	svc.InitRoutes()

	// Run web server
	if svc.ServerConfig.WebConfig.Enable {
		webSvc := &webServe{
			port:  svc.ServerConfig.WebConfig.Port,
			index: svc.ServerConfig.WebConfig.Index,
		}
		go webSvc.Serve()
		log.Info("run web server on port ", webSvc.port)
	}

	// Handle signel to delete bridge
	sigChan := make(chan os.Signal, 8)
	signal.Notify(sigChan, unix.SIGTERM, unix.SIGINT)
	go svc.HandleSignal(sigChan)

	// Setup local bridge
	if err := svc.SetupLan(); err != nil {
		return err
	}

	// Launch udp server
	return svc.ListenAndServe()
}
