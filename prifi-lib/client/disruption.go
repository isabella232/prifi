package client

import (
	"github.com/dedis/prifi/prifi-lib/net"
	"gopkg.in/dedis/onet.v2/log"
)

/*
* Received_REL_ALL_DISRUPTION_REVEAL handles REL_ALL_DISRUPTION_REVEAL messages.
* The method calls a function from the DCNet to regenerate the bits from roundID in position BitPos
* The result is sent to the relay.
 */
func (p *PriFiLibClientInstance) Received_REL_ALL_DISRUPTION_REVEAL(msg net.REL_ALL_DISRUPTION_REVEAL) error {
	upstreamCell := p.clientState.DCNet.GetBitsOfRound(int32(msg.RoundID), int32(msg.BitPos))
	//send the data to the relay
	toSend := &net.CLI_REL_DISRUPTION_REVEAL{
		ClientID: p.clientState.ID,
		Bits:     upstreamCell,
	}

	//LB->CV: if ForceDisruptionOnRound, continue lying here!
	if true {
		upstreamCell[15] = 1
	}

	p.messageSender.SendToRelayWithLog(toSend, "")
	log.Lvl1("Disruption: Sending previous round to relay (Round: ", msg.RoundID, ", bit position:", msg.BitPos, ")")
	return nil
}

/*
* Received_REL_ALL_REVEAL_SHARED_SECRETS handles REL_ALL_REVEAL_SHARED_SECRETS messages.
* The method gets the shared secret and sends it to the relay.
 */
func (p *PriFiLibClientInstance) Received_REL_ALL_REVEAL_SHARED_SECRETS(msg net.REL_ALL_REVEAL_SHARED_SECRETS) error {
	// CARLOS TODO: NIZK
	// TODO: check that the relay asks for the correct entity, and not a honest entity. There should be a signature check on the TRU_REL_DISRUPTION_REVEAL the relay received (and forwarded to the client)
	secret := p.clientState.sharedSecrets[msg.UserID]
	toSend := &net.CLI_REL_SHARED_SECRET{
		Secret: secret,
		NIZK:   make([]byte, 0)}
	p.messageSender.SendToRelayWithLog(toSend, "Sent secret to relay")
	log.Lvl1("Reveling secret with trustee", msg.UserID)
	return nil
}
