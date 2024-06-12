package main

import (
	"os"

	"github.com/lucheng0127/virtuallan/pkg/app"
	log "github.com/sirupsen/logrus"
)

func main() {
	app := app.GetApp()

	if err := app.Run(os.Args); err != nil {
		log.Panic(err)
		os.Exit(1)
	}
}
