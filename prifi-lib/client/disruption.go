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
	if p.clientState.Cheater {
		// TEST DISRUPTION
		//upstreamCell[0] ^= 1
	}

	//send the data to the relay
	toSend := &net.CLI_REL_DISRUPTION_REVEAL{
		ClientID: p.clientState.ID,
		Bits:     upstreamCell,
	}
	p.messageSender.SendToRelayWithLog(toSend, "")
	log.Lvl1("Disruption: Sending previous round to realy (Round: ", msg.RoundID, ", bit position:", msg.BitPos, ")")
	return nil
}

/*
* Received_REL_ALL_DISRUPTION_SECRET handles REL_ALL_DISRUPTION_SECRET messages.
* The method gets the shared secret and sends it to the relay.
 */
func (p *PriFiLibClientInstance) Received_REL_ALL_DISRUPTION_SECRET(msg net.REL_ALL_DISRUPTION_SECRET) error {
	// CARLOS TODO: NIZK
	secret := p.clientState.sharedSecrets[msg.UserID]
	toSend := &net.CLI_REL_DISRUPTION_SECRET{
		Secret: secret,
		NIZK:   make([]byte, 0)}
	p.messageSender.SendToRelayWithLog(toSend, "Sent secret to relay")
	log.Lvl1("Reveling secret with trustee", msg.UserID)
	return nil
}
