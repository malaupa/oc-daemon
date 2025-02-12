package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/telekom-mms/oc-daemon/internal/devmon"
)

func main() {
	log.SetLevel(log.DebugLevel)
	d := devmon.NewDevMon()
	go d.Start()
	for u := range d.Updates() {
		log.Println(u)
	}
}
