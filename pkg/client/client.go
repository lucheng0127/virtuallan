package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/lucheng0127/virtuallan/pkg/packet"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"golang.org/x/term"
)

type Client struct {
	target   string
	user     string
	password string
	key      string
	logLevel string
	Conn     *net.UDPConn
	IPAddr   string
	Iface    *water.Interface
}

type ClientOpt func(*Client)

func ClientSetTarget(t string) ClientOpt {
	return func(c *Client) {
		c.target = t
	}
}

func ClientSetUser(u string) ClientOpt {
	return func(c *Client) {
		c.user = u
	}
}

func ClientSetPasswd(p string) ClientOpt {
	return func(c *Client) {
		c.password = p
	}
}

func ClientSetKey(k string) ClientOpt {
	return func(c *Client) {
		c.key = k
	}
}

func ClientSetLogLevel(l string) ClientOpt {
	return func(c *Client) {
		c.logLevel = l
	}
}

func NewClient(opts ...ClientOpt) *Client {
	client := new(Client)

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func GetLoginInfo() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Username:")
	user, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Println("Password:")
	bytePasswd, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}

	passwd := string(bytePasswd)

	return strings.TrimSpace(user), strings.TrimSpace(passwd), nil
}

func checkLoginTimeout(c chan string) error {
	select {
	case <-c:
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("login timeout")
	}
}
func (c *Client) Close() error {
	finPkt := packet.NewFinPkt()

	stream, err := finPkt.Encode()
	if err != nil {
		return err
	}

	_, err = c.Conn.Write(stream)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) HandleSignal(sigChan chan os.Signal) error {
	sig := <-sigChan
	log.Infof("received signal: %v, send fin pkt to close conn\n", sig)

	if err := c.Close(); err != nil {
		return fmt.Errorf("send fin pkt %s", err.Error())
	}

	os.Exit(0)
	return nil
}

func (c *Client) SetLogLevel() {
	switch strings.ToUpper(c.logLevel) {
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
}

func (c *Client) CheckAuthed() bool {
	return c.IPAddr != ""
}
