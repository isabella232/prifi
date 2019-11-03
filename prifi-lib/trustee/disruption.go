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
	upstreamCell := p.trusteeState.DCNet.GetBitsOfRound(int32(msg.RoundID), int32(msg.BitPos))
	toSend := &net.TRU_REL_DISRUPTION_REVEAL{
		TrusteeID: p.trusteeState.ID,
		Bits: upstreamCell,
	}
	p.messageSender.SendToRelayWithLog(toSend, "")
	log.Error("REVEAL", msg.RoundID, msg.BitPos, upstreamCell)
	return nil
}

/*
* Received_REL_ALL_DISRUPTION_SECRET handles REL_ALL_DISRUPTION_SECRET messages.
* The method gets the shared secret and sends it to the relay.
*/
func (p *PriFiLibTrusteeInstance) Received_REL_ALL_DISRUPTION_SECRET(msg net.REL_ALL_DISRUPTION_SECRET) error {
	// CARLOS TODO: NIZK
	secret := p.trusteeState.sharedSecrets[msg.UserID]
	toSend := &net.TRU_REL_DISRUPTION_SECRET{
		Secret: secret,
		NIZK:   make([]byte, 0)}
	p.messageSender.SendToRelayWithLog(toSend, "Sent secret to relay")
	log.Lvl1("REVELING SECRET WITH CLIENT", msg.UserID)
	return nil
}
