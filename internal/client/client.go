package client

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/T-Systems-MMS/oc-daemon/pkg/client"
	log "github.com/sirupsen/logrus"
)

const (
	// maxReconnectTries is the maximum amount or reconnect retries
	maxReconnectTries = 5
)

// readXMLProfile reads the contents of the XML profile
func readXMLProfile() []byte {
	b, err := os.ReadFile(xmlProfile)
	if err != nil {
		return nil
	}
	return b
}

// connectVPN connects to the VPN if necessary
func connectVPN() {
	// create client
	c := client.NewClient()

	// try to read current xml profile
	pre := readXMLProfile()

	// authenticate
	c.ClientCertificate = config.ClientCertificate
	c.ClientKey = config.ClientKey
	c.CACertificate = config.CACertificate
	c.XMLProfile = xmlProfile
	c.VPNServer = config.VPNServer
	c.User = config.User
	c.Password = config.Password

	if err := c.Authenticate(); err != nil {
		log.WithError(err).Fatal("error authenticating user for VPN")
	}

	// warn user if profile changed
	post := readXMLProfile()
	if !bytes.Equal(pre, post) {
		time.Sleep(2 * time.Second)
		log.Warnln("XML Profile was updated. Connection attempt " +
			"might fail. Please, check status and reconnect " +
			"if necessary.")
		time.Sleep(2 * time.Second)
	}

	// connect
	if err := c.Connect(); err != nil {
		log.WithError(err).Fatal("error connecting to VPN")
	}
}

// disconnectVPN disconnects the VPN
func disconnectVPN() {
	// create client
	c := client.NewClient()

	// disconnect
	err := c.Disconnect()
	if err != nil {
		log.WithError(err).Fatal("error disconnecting from VPN")
	}
}

// reconnectVPN reconnects to the VPN
func reconnectVPN() {
	// create client
	client := client.NewClient()

	// check status
	status, err := client.Query()
	if err != nil {
		log.WithError(err).Fatal("error reconnecting to VPN")
	}

	// disconnect if needed
	if status.Running {
		// send disconnect request
		disconnectVPN()
	}

	// wait for status to switch to untrusted network and not running
	try := 0
	for {
		status, err := client.Query()
		if err != nil {
			log.WithError(err).Fatal("error reconnecting to VPN")
		}

		if !status.TrustedNetwork.Trusted() &&
			!status.ConnectionState.Connected() &&
			!status.Running {
			// authenticate and connect
			connectVPN()
			return
		}

		try++
		if try >= maxReconnectTries {
			// too many tries, abort
			log.Fatal("error reconnecting to VPN")
		}

		// sleep a second before retry
		time.Sleep(time.Second)
	}

}

// getStatus gets the VPN status from the daemon
func getStatus() {
	c := client.NewClient()
	status, err := c.Query()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Trusted Network:  %s\n", status.TrustedNetwork)
	fmt.Printf("Connection State: %s\n", status.ConnectionState)
	fmt.Printf("IP:               %s\n", status.IP)
	log.Printf("Running: %t", status.Running)
	log.Printf("Config: %+v", status.Config)
}
