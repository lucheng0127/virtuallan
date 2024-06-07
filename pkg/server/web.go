package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lucheng0127/virtuallan/pkg/users"
	"github.com/lucheng0127/virtuallan/pkg/utils"
)

type webServe struct {
	port int
}

type EpEntry struct {
	User   string
	Addr   string
	Iface  string
	IP     string
	TX_PKT uint64
	RX_PKT uint64
	TX     string
	RX     string
	Login  string
}

func listEpEntries(c *gin.Context) {
	// Add pkt count
	var data []*EpEntry

	// Get all link stats
	linkStats := utils.GetLinkStats()

	// Format endpoint stats data
	for user, addr := range users.UserEPMap {
		c, ok := UPool[addr]
		if !ok {
			continue
		}

		rxPkt, txPkt, rx, tx := utils.GetLinkStatsByName(c.Iface.Name(), linkStats)

		data = append(data, &EpEntry{
			User:   user,
			Addr:   addr,
			Iface:  c.Iface.Name(),
			IP:     c.IP.String(),
			Login:  c.Login,
			RX_PKT: rxPkt,
			RX:     rx,
			TX_PKT: txPkt,
			TX:     tx,
		})
	}

	c.JSON(http.StatusOK, data)
}

func (svc *webServe) Serve() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	router := gin.Default()
	router.LoadHTMLFiles("./static/index.html")

	router.GET("/endpoints", listEpEntries)
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.Run(fmt.Sprintf(":%d", svc.port))
}
