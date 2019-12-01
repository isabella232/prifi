package trustee

import (
	"github.com/dedis/prifi/prifi-lib/net"
	"gopkg.in/dedis/onet.v2/log"
)

/*
* Received_REL_ALL_DISRUPTION_REVEAL handles REL_ALL_DISRUPTION_REVEAL messages.
* The method calls a function from the DCNet to regenerate the bits from roundID in position BitPos
* The result is sent to the relay.
 */
func (p *PriFiLibTrusteeInstance) Received_REL_ALL_DISRUPTION_REVEAL(msg net.REL_ALL_DISRUPTION_REVEAL) error {
	log.Lvl1("Disruption Phase 1: Received de-anonymization query for round", msg.RoundID, "bit pos", msg.BitPos)
	bitMap := p.trusteeState.DCNet.GetBitsOfRound(int32(msg.RoundID), int32(msg.BitPos))
	toSend := &net.TRU_REL_DISRUPTION_REVEAL{
		TrusteeID: p.trusteeState.ID,
		Bits:      bitMap,
	}
	p.messageSender.SendToRelayWithLog(toSend, "")
	log.Lvl1("Disruption: Sending previous round to relay (Round: ", msg.RoundID, ", bit position:", msg.BitPos, "), value", bitMap)
	return nil
}

/*
* Received_REL_ALL_REVEAL_SHARED_SECRETS handles REL_ALL_REVEAL_SHARED_SECRETS messages.
* The method gets the shared secret and sends it to the relay.
 */
func (p *PriFiLibTrusteeInstance) Received_REL_ALL_REVEAL_SHARED_SECRETS(msg net.REL_ALL_REVEAL_SHARED_SECRETS) error {
	log.Lvl1("Disruption Phase 2: Received a reveal secret message for client", msg.EntityID)
	// CARLOS TODO: NIZK
	// TODO: check that the relay asks for the correct entity, and not a honest entity. There should be a signature check on the TRU_REL_DISRUPTION_REVEAL the relay received (and forwarded to the client)
	secret := p.trusteeState.sharedSecrets[msg.EntityID]
	toSend := &net.TRU_REL_SHARED_SECRET{
		TrusteeID: p.trusteeState.ID,
		ClientID:  msg.EntityID,
		Secret:    secret,
		NIZK:      make([]byte, 0)}
	p.messageSender.SendToRelayWithLog(toSend, "Sent secret to relay")
	log.Lvl1("Reveling secret with client", msg.EntityID)
	return nil
}
