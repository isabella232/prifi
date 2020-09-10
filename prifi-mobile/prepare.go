package prifimobile

// Functions that are needed to initialize a server are all here

import (
	prifi_protocol "github.com/dedis/prifi/sda/protocols"
	prifi_service "github.com/dedis/prifi/sda/services"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/encoding"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/app"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/onet/v3/network"
)

func startCothorityNode() (*onet.Server, *app.Group, *prifi_service.ServiceState, error) {
	prifiConfig, err := parsePrifi()
	if err != nil {
		log.Error("Could not read prifi config")
		return nil, nil, nil, err
	}

	host, err := parseCothority()
	if err != nil {
		log.Error("Could not start cothority")
		return nil, nil, nil, err
	}

	group, err := parseGroup()
	if err != nil {
		log.Error("Could not read the group description:", err)
		return nil, nil, nil, err
	}

	service := host.Service(prifi_service.ServiceName).(*prifi_service.ServiceState)
	service.SetConfigFromToml(prifiConfig)

	// TODO Replace getCommitID
	prifiConfig.ProtocolVersion = "v1" // standard string for all nodes

	return host, group, service, nil
}

func parsePrifi() (*prifi_protocol.PrifiTomlConfig, error) {
	c, err := getPrifiConfig()
	if err != nil {
		return nil, err
	}

	if c.OverrideLogLevel > 0 {
		log.Lvl3("Log level set to", c.OverrideLogLevel)
		log.SetDebugVisible(c.OverrideLogLevel)
	}

	return c, nil
}

func parseCothority() (*onet.Server, error) {
	c, err := getCothorityConfig()
	if err != nil {
		return nil, err
	}

	if c.Suite == "" {
		c.Suite = "Ed25519"
	}
	suite, err := suites.Find(c.Suite)
	if err != nil {
		return nil, err
	}

	secret, err := encoding.StringHexToScalar(suite, c.Private)
	if err != nil {
		return nil, err
	}

	point, err := encoding.StringHexToPoint(suite, c.Public)
	if err != nil {
		return nil, err
	}

	serverIdentity := network.NewServerIdentity(point, c.Address)
	serverIdentity.SetPrivate(secret)
	serverIdentity.Description = c.Description
	server := onet.NewServerTCPWithListenAddr(serverIdentity, suite, c.ListenAddress)

	// Don't handle Websocket TLC

	return server, nil
}

func parseGroup() (*app.Group, error) {
	c, err := getGroupConfig()
	if err != nil {
		return nil, err
	}

	return c, nil
}
