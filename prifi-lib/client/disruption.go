package client

import (
	"bytes"
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

	if p.clientState.ForceDisruptionSinceRound3 && p.clientState.ID == 0 {
		log.Lvl1("Disruption: Malicious client cheating again, old value", bitMap, "(new value right below)")
		trusteeToAccuse := 0
		if bitMap[trusteeToAccuse] == 0 {
			bitMap[trusteeToAccuse] = 1
		} else {
			bitMap[trusteeToAccuse] = 0
		}
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

	if p.clientState.ForceDisruptionSinceRound3 && p.clientState.ID == 0 {
		//this client is hesitant to answer as he will get caught
		//CV->LB: How do we handle this in the relay?
		time.Sleep(1 * time.Second)
		// this is just to let the honest trustee answer and see what happens
	}

	p.messageSender.SendToRelayWithLog(toSend, "Sent secret to relay")
	log.Lvl1("Reveling secret with trustee", msg.EntityID)
	return nil
}

func (p *PriFiLibClientInstance) handlePossibleDisruption(msg net.REL_CLI_DOWNSTREAM_DATA) error {
	if p.clientState.RoundNo-1 == p.clientState.MyLastRound {

		if p.clientState.B_echo_last == 1 {
			log.Lvl1("We previously set b_echo_last=1, relay retransmitted message", msg.Data)
			// We are in the disruption protection blame protocol
			if bytes.Equal(msg.Data, p.clientState.LastMessage) {
				p.clientState.B_echo_last = 0
				p.clientState.DisruptionWrongBitPosition = -1
				log.Error("There was no disruption; the relay is lying about disruption (outside of threat model).")
			} else {
				log.Lvl1("We previously set b_echo_last=1, comparing messages: (only on log level 3)")
				log.Lvl3(msg.Data)
				log.Lvl3(p.clientState.LastMessage)

				// Get the l-th bit
				found := false
				for index, b := range msg.Data {
					if b != p.clientState.LastMessage[index] {
						//Get the bit
						for j := 0; j < 8; j++ {
							mask := byte(1 << uint(j))
							if (b & mask) != (p.clientState.LastMessage[index] & mask) {
								bitPos := index*8 + (7 - j)

								if (p.clientState.LastMessage[index] & mask) == 1 {
									log.Lvl1("Bit at position", bitPos, "was a 1 toggled to 0, ignoring...")
								} else {
									// Found bit
									p.clientState.DisruptionWrongBitPosition = bitPos
									log.Lvl1("Disruptive bit position:", p.clientState.DisruptionWrongBitPosition)
									found = true
									break
								}
							}
						}
						if found {
							break
						}
					}
				}
			}
		} else {
			p.clientState.B_echo_last = 0
			p.clientState.DisruptionWrongBitPosition = -1
			if len(msg.HashOfPreviousUpstreamData) != 32 {
				log.Error("The relay did not send the hash back. This should never happen.")
				p.clientState.B_echo_last = 1
			} else {
				// Getting hash sent by relay
				hash := msg.HashOfPreviousUpstreamData

				// Getting previously calculated hash
				previousHash := p.clientState.HashFromPreviousMessage[:]

				// Comparing both hashes
				if !bytes.Equal(hash, previousHash) {
					log.Error("Disruption protection hash comparison failed.")
					p.clientState.B_echo_last = 1
				} else {
					p.clientState.B_echo_last = 0
				}
			}
		}
	}

	return nil
}
