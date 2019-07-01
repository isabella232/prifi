package prifimobile

import (
	"errors"
	"gopkg.in/dedis/kyber.v2/suites"
	"gopkg.in/dedis/kyber.v2/util/encoding"
	"gopkg.in/dedis/kyber.v2/util/key"
	"gopkg.in/dedis/onet.v2/network"
	"strconv"
)

const relayIndex = 0
const separatorHostPort = ":"

// PriFi Port
func GetPrifiPort() (int, error) {
	c, err := getPrifiConfig()
	if err != nil {
		return 0, err
	}

	return c.SocksServerPort, nil
}

// Relay Address
func GetRelayAddress() (string, error) {
	c, err := getGroupConfig()
	relayAddress := c.Roster.Get(relayIndex).Address.Host()
	return relayAddress, err
}

// Sets RelayAddress
func SetRelayAddress(host string) error {
	c, err := getGroupConfig()
	if err != nil {
		return err
	}

	port := c.Roster.Get(relayIndex).Address.Port()
	fullAddress := network.NewAddress(network.PlainTCP, host+separatorHostPort+port)
	if fullAddress.Valid() {
		c.Roster.Get(relayIndex).Address = fullAddress
		return nil
	}
	return errors.New("not a host:port address")
}

// Relay Port
func GetRelayPort() (int, error) {
	c, err := getGroupConfig()
	portString := c.Roster.Get(relayIndex).Address.Port()
	port, _ := strconv.Atoi(portString)
	return port, err
}

// Sets Relay Port
func SetRelayPort(port int) error {
	c, err := getGroupConfig()
	if err != nil {
		return err
	}

	relayAddress := c.Roster.Get(relayIndex).Address.Host()
	newPort := strconv.Itoa(port)
	fullAddress := network.NewAddress(network.PlainTCP, relayAddress+separatorHostPort+newPort)
	if fullAddress.Valid() {
		c.Roster.Get(relayIndex).Address = fullAddress
		return nil
	}
	return errors.New("not a host:port address")
}

// Relay Socks Port
func GetRelaySocksPort() (int, error) {
	c, err := getPrifiConfig()
	return c.SocksClientPort, err
}

// Sets relay socks port
func SetRelaySocksPort(port int) error {
	c, err := getPrifiConfig()
	if err != nil {
		return err
	}
	c.SocksClientPort = port
	return nil
}

// Keys
func GenerateNewKeyPairAndAssign() error {
	// Generate new raw key pair
	suite := suites.MustFind("Ed25519") // May crash
	kp := key.NewKeyPair(suite)

	// Parse private key
	priStr, err := encoding.ScalarToStringHex(suite, kp.Private)
	if err != nil {
		return err
	}

	// Parse public key
	pubStr, err := encoding.PointToStringHex(suite, kp.Public)
	if err != nil {
		return err
	}

	err = SetPublicKey(pubStr)
	if err != nil {
		return err
	}

	err = SetPrivateKey(priStr)
	if err != nil {
		return err
	}

	return nil
}

// Gets Public Key
func GetPublicKey() (string, error) {
	c, err := getCothorityConfig()
	return c.Public, err
}

// Sets Public Key
func SetPublicKey(pubKey string) error {
	c, err := getCothorityConfig()
	if err != nil {
		return err
	}

	c.Public = pubKey
	return nil
}

// Get Private Key
func GetPrivateKey() (string, error) {
	c, err := getCothorityConfig()
	return c.Private, err
}

//Sets Private Key
func SetPrivateKey(priKey string) error {
	c, err := getCothorityConfig()
	if err != nil {
		return err
	}

	c.Private = priKey
	return nil
}

// Support functions
func getFullAddress() string {
	c, _ := getGroupConfig()
	return c.Roster.Get(relayIndex).Address.String()
}
