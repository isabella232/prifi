package client

import (
	"github.com/dedis/prifi/prifi-lib/net"
	"gopkg.in/dedis/onet.v2/log"
	"time"
)

/*
* Received_REL_ALL_DISRUPTION_REVEAL handles REL_ALL_DISRUPTION_REVEAL messages.
* The method calls a function from the DCNet to regenerate the bits from roundID in position BitPos
* The result is sent to the relay.
 */
func (p *PriFiLibClientInstance) Received_REL_ALL_DISRUPTION_REVEAL(msg net.REL_ALL_DISRUPTION_REVEAL) error {
	log.Lvl1("Disruption Phase 1: Received de-anonymization query for round", msg.RoundID, "bit pos", msg.BitPos)

	// TODO: check the NIZK

	bitMap := p.clientState.DCNet.GetBitsOfRound(int32(msg.RoundID), int32(msg.BitPos))
	//send the data to the relay
	toSend := &net.CLI_REL_DISRUPTION_REVEAL{
		ClientID: p.clientState.ID,
		Bits:     bitMap,
	}

	//if ForceDisruptionOnRound, continue lying here!
	//LB->CV: have this as a param in prifi.toml
	ForceDisruptionOnRound := 3
	if ForceDisruptionOnRound > -1 && p.clientState.ID == 0 {
		log.Lvl1("Disruption: Malicious client cheating again, old value", bitMap, "(new value right below)")
		trusteeToAccuse := 0
		// pretend the PRG told me to output a 1, and the trustee is lying with its 0
		bitMap[trusteeToAccuse] = 1
	}
	log.Lvl1("Disruption: Sending previous round to relay (Round: ", msg.RoundID, ", bit position:", msg.BitPos, "), value", bitMap)

	p.messageSender.SendToRelayWithLog(toSend, "")
	return nil
}

/*
* Received_REL_ALL_REVEAL_SHARED_SECRETS handles REL_ALL_REVEAL_SHARED_SECRETS messages.
* The method gets the shared secret and sends it to the relay.
 */
func (p *PriFiLibClientInstance) Received_REL_ALL_REVEAL_SHARED_SECRETS(msg net.REL_ALL_REVEAL_SHARED_SECRETS) error {
	log.Lvl1("Disruption Phase 2: Received a reveal secret message for trustee", msg.EntityID)
	// CARLOS TODO: NIZK
	// TODO: check that the relay asks for the correct entity, and not a honest entity. There should be a signature check on the TRU_REL_DISRUPTION_REVEAL the relay received (and forwarded to the client)
	secret := p.clientState.sharedSecrets[msg.EntityID]
	toSend := &net.CLI_REL_SHARED_SECRET{
		ClientID:  p.clientState.ID,
		TrusteeID: msg.EntityID,
		Secret:    secret,
		NIZK:      make([]byte, 0)}

	//LB->CV: have this as a param in prifi.toml
	ForceDisruptionOnRound := 3
	if ForceDisruptionOnRound > -1 && p.clientState.ID == 0 {
		//this client is hesitant to answer as he will get caught
		time.Sleep(1 * time.Second)
		// this is just to let the honest trustee answer and see what happens
	}

	p.messageSender.SendToRelayWithLog(toSend, "Sent secret to relay")
	log.Lvl1("Reveling secret with trustee", msg.EntityID)
	return nil
}
