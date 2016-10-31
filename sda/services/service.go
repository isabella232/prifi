package prifi

/*
* This is the internal part of the API. As probably the prifi-service will
* not have an external API, this will not have any API-functions.
 */

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dedis/cothority/app/lib/config"
	"github.com/dedis/cothority/log"
	"github.com/dedis/cothority/network"
	"github.com/dedis/cothority/sda"
	"github.com/lbarman/prifi_dev/sda/protocols"
)

// ServiceName is the name to refer to the Template service from another
// package.
const ServiceName = "ExampleService"

var serviceID sda.ServiceID

// Register Service with SDA
func init() {
	sda.RegisterNewService(ServiceName, newService)
	serviceID = sda.ServiceFactory.ServiceID(ServiceName)
}

// This struct contains the state of the service
type Service struct {
	// We need to embed the ServiceProcessor, so that incoming messages
	// are correctly handled.
	*sda.ServiceProcessor
	Storage *Storage
	path    string
}

// This structure will be saved, on the contrary of the 'Service'-structure
// which has per-service information stored
type Storage struct {
	TrusteeID string
}

// StartTrustee has to take a configuration and start the necessary
// protocols to enable the trustee-mode.
func (s *Service) StartTrustee() error {
	log.Info("Service", s, "running in trustee mode")
	// Set up the configuration
	return nil
}

// StartRelay has to take a configuration and start the necessary
// protocols to enable the relay-mode.
// In this example it simply starts the demo protocol
func (s *Service) StartRelay(group *config.Group) error {
	log.Info("Service", s, "running in relay mode")

	// Obtain the relay's ServerIdentity from the group file
	// (it is the server with "relay" as a description)
	var relayIdentity *network.ServerIdentity = nil
	nodeList := group.Roster.List
	for i := 0; i < len(nodeList); i++ {
		if group.GetDescription(nodeList[i]) == "relay" {
			relayIdentity = nodeList[i]
			break
		}
	}

	if (relayIdentity == nil) {
		log.Fatal("Group file does not contian a relay")
	}

	// Start the PriFi protocol on a flat tree with the relay as root
	tree := group.Roster.GenerateNaryTreeWithRoot(100, relayIdentity)
	// Create the protocol with SDA: the newProtocolInstance function
	// defined in protocol.go will be called on each node.
	pi, err := s.CreateProtocolSDA(prifi.ProtocolName, tree)
	// Another way to create the protocol is to use CreateProtocolService:
	// in this case the NewProtocol function defined in this file
	// will be called on each node instead of newProtocolInstance.
	//pi, err := s.CreateProtocolService(prifi.ProtocolName, tree)

	if err != nil {
		log.Fatal("Unable to start Prifi protocol:", err)
	}

	pi.Start()

	return nil
}

// StartClient has to take a configuration and start the necessary
// protocols to enable the client-mode.
func (s *Service) StartClient() error {
	log.Info("Service", s, "running in client mode")
	// Set up the configuration
	return nil
}

// NewProtocol is called on all nodes of a Tree (except the root, since it is
// the one starting the protocol) so it's the Service that will be called to
// generate the PI on all others node.
// If you use CreateProtocolSDA, this will not be called, as the SDA will
// instantiate the protocol on its own. If you need more control at the
// instantiation of the protocol, use CreateProtocolService, and you can
// give some extra-configuration to your protocol in here.
func (s *Service) NewProtocol(tn *sda.TreeNodeInstance, conf *sda.GenericConfig) (sda.ProtocolInstance, error) {
	log.Info("Calling service.NewProtocol")

	pi := &prifi.MyProtocol{
		TreeNodeInstance: tn,
	}

	// Register message handler(s)
	if err:= pi.RegisterHandler(pi.HandleMsg); err != nil {
		log.Fatal("Could not register handler:", err)
	}

	return pi, nil
}

// saves the actual identity
func (s *Service) save() {
	log.Lvl3("Saving service")
	b, err := network.MarshalRegisteredType(s.Storage)
	if err != nil {
		log.Error("Couldn't marshal service:", err)
	} else {
		err = ioutil.WriteFile(s.path+"/prifi.bin", b, 0660)
		if err != nil {
			log.Error("Couldn't save file:", err)
		}
	}
}

// Tries to load the configuration and updates if a configuration
// is found, else it returns an error.
func (s *Service) tryLoad() error {
	configFile := s.path + "/identity.bin"
	b, err := ioutil.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Error while reading %s: %s", configFile, err)
	}
	if len(b) > 0 {
		_, msg, err := network.UnmarshalRegistered(b)
		if err != nil {
			return fmt.Errorf("Couldn't unmarshal: %s", err)
		}
		log.Lvl3("Successfully loaded")
		s.Storage = msg.(*Storage)
	}
	return nil
}

// newTemplate receives the context and a path where it can write its
// configuration, if desired. As we don't know when the service will exit,
// we need to save the configuration on our own from time to time.
func newService(c *sda.Context, path string) sda.Service {
	log.Info("Calling newService")
	s := &Service{
		ServiceProcessor: sda.NewServiceProcessor(c),
		path:             path,
	}
	if err := s.tryLoad(); err != nil {
		log.Error(err)
	}
	return s
}
