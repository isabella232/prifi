/*
 * The package name must NOT contain underscore.
 */
package prifiMobile

import (
	prifi_service "github.com/dedis/prifi/sda/services"
	"gopkg.in/dedis/onet.v2"
	"gopkg.in/dedis/onet.v2/log"
	"time"
	"net"
	"github.com/pkg/errors"
	"strconv"
)

var stopChan chan bool
var errorChan chan error
var stopPingChan chan bool

var globalHost *onet.Server
var globalService *prifi_service.ServiceState

var currentRelayHost string
var currentRelayPort string

// The "main" function that is called by Mobile OS in order to launch a client server
func StartClient() error {
	stopChan = make(chan bool, 1)
	errorChan = make(chan error, 1)
	stopPingChan = make(chan bool, 1)

	host, err := GetRelayAddress()
	if err != nil {
		return err
	}

	port, err := GetRelayPort()
	if err != nil {
		return err
	}

	currentRelayHost = host
	currentRelayPort = strconv.Itoa(port)

	if !isRelayReachable() {
		return errors.New("Relay is not reachable at launch.")
	}

	go checkIfRelayIsStillReachable(stopPingChan)
	go func() {
		errorChan <- run()
	}()

	select {
	case err := <-errorChan:
		log.Error("Error occurs", err)
		return err
	case <-stopChan:

		// Stop goroutines
		globalService.ShutdownConnexionToRelay()
		globalService.ShutdownSocks()
		stopPing()

		// Change the protocol state to SHUTDOWN
		globalService.StopPriFiCommunicateProtocol()

		// Clean network-related resources
		globalHost.Close()

		log.Info("PriFi Session Ended")
		return nil
	}
}

func StopClient() {
	stopChan <- true
}

func stopPing() {
	stopPingChan <- true
}

func run() error {
	host, group, service, err := startCothorityNode()
	globalHost = host
	globalService = service

	if err != nil {
		log.Error("Could not start the cothority node:", err)
		return err
	}

	if err := service.StartClient(group, time.Duration(0)); err != nil {
		log.Error("Could not start the PriFi service:", err)
		return err
	}

	//host.Router.AddErrorHandler(networkErrorHappenedForMobile)
	host.Start()

	// Never return
	return nil
}

func timeout() {
	select {
	case <-time.After(10 * time.Second):
		if !globalService.IsPriFiProtocolRunning() {
			log.Lvl2("Timeout triggered")
			StopClient()
		} else {
			log.Lvl2("Timeout not triggered")
			go checkIfRelayIsStillReachable(stopPingChan)
		}
	}
}

func checkIfRelayIsStillReachable(stopChan chan bool) {
	for {
		select {
		case _ = <-stopChan: // Check if we need to stop
			log.Lvl2("Ping Relay Goroutine Stopped")
			return
		case <-time.After(3 * time.Second):
		}

		if !isRelayReachable() {
			log.Lvl2("The relay is not reachable.")
			globalService.StopPriFiCommunicateProtocol()

			doWeDisconnectWhenNetworkError, _ := GetMobileDisconnectWhenNetworkError()

			if doWeDisconnectWhenNetworkError {
				StopClient()
			} else {
				stopPing()
				go timeout()
			}
		}
	}

}

func isRelayReachable() bool {
	relayAddress := currentRelayHost + ":" + currentRelayPort
	conn, err := net.DialTimeout("tcp", relayAddress, 2 * time.Second)
	if err != nil {
		return false
	} else {
		conn.Close()
		return true
	}
}