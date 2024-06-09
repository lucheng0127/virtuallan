package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lucheng0127/virtuallan/pkg/config"
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
}

func NewServer() *Server {
	svc := new(Server)

	svc.UsedIP = make([]int, 0)
	svc.MLock = sync.Mutex{}

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

func Run(cCtx *cli.Context) error {
	// New server and do cfg parse
	svc := NewServer()

	cfgDir := cCtx.String("config-dir")
	cfg, err := config.LoadConfigFile(config.GetCfgPath(cfgDir))
	if err != nil {
		return err
	}

	svc.ServerConfig = cfg
	svc.userDb = filepath.Join(cfgDir, "users")

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

	if !utils.ValidateIPv4WithNetmask(svc.ServerConfig.IP) {
		return fmt.Errorf("invalidate ip %s, <ip>/<netmask len>", svc.ServerConfig.IP)
	}

	if err := svc.ParseDHCPRange(); err != nil {
		return err
	}

	// Run web server
	if svc.ServerConfig.WebConfig.Enable {
		webSvc := &webServe{
			port: svc.ServerConfig.WebConfig.Port,
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
