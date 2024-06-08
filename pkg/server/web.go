package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lucheng0127/virtuallan/pkg/users"
)

type webServe struct {
	port int
	svc  *Server
}

type EpEntry struct {
	User  string
	Addr  string
	Iface string
	IP    string
	Login string
}

func listEpEntries(c *gin.Context) {
	// TODO(shawnlu): Add pkt count
	var data []*EpEntry

	for user, addr := range users.UserEPMap {
		c, ok := UPool[addr]
		if !ok {
			continue
		}

		data = append(data, &EpEntry{
			User:  user,
			Addr:  addr,
			Iface: c.Iface.Name(),
			IP:    c.IP.String(),
			Login: c.Login,
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
