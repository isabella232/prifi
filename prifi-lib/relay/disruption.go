package relay

import (
	"github.com/dedis/prifi/prifi-lib/config"
	"github.com/dedis/prifi/prifi-lib/net"
	"gopkg.in/dedis/kyber.v2"
	"gopkg.in/dedis/onet.v2/log"
	"strconv"
)

// Received_CLI_REL_BLAME
// CARLOS NEEDS TO IMPLMENT THIS
func (p *PriFiLibRelayInstance) Received_CLI_REL_BLAME(msg net.CLI_REL_DISRUPTION_BLAME) error {

	// TODO: Check NIZK
	p.stateMachine.ChangeState("BLAMING")

	toSend := &net.REL_ALL_DISRUPTION_REVEAL{
		RoundID: msg.RoundID,
		BitPos:  msg.BitPos}

	p.relayState.blamingData[0] = int(msg.RoundID)
	p.relayState.blamingData[1] = msg.BitPos

	// broadcast to all trustees
	for j := 0; j < p.relayState.nTrustees; j++ {
		// send to the j-th trustee
		p.messageSender.SendToTrusteeWithLog(j, toSend, "Reveal message sent to trustee "+strconv.Itoa(j+1))
	}

	// broadcast to all clients
	for i := 0; i < p.relayState.nClients; i++ {
		// send to the i-th client
		p.messageSender.SendToClientWithLog(i, toSend, "Reveal message sent to client "+strconv.Itoa(i+1))
	}

	//Bool var to let the round finish then stop, switch to state blaming, and reveal ?

	return nil
}

/*
* Received_CLI_REL_DISRUPTION_REVEAL handles CLI_REL_DISRUPTION_REVEAL messages
* First, saves the bits reveal by the client.
* It checks that the bits received by the client matches with the ones of the disruptive round.
* For this it XORs the bits revealed together and compares it to the bit in the disruptive position.
* If there is a mismatch, the client is the disruptor.
* Else checks if all the reveals are received to move to next blame phase.
 */
func (p *PriFiLibRelayInstance) Received_CLI_REL_DISRUPTION_REVEAL(msg net.CLI_REL_DISRUPTION_REVEAL) error {

	log.Lvl1("Disruption Phase 1: Received bits from Client", msg.ClientID, "value", msg.Bits)

	p.relayState.clientBitMap[msg.ClientID] = msg.Bits
	result := p.compareBitsClient(msg.ClientID, msg.Bits)
	if result {
		log.Fatal("Disruption Phase 1: Disruptor is Client", msg.ClientID, ".")
	} else if (len(p.relayState.clientBitMap) == p.relayState.nClients) && (len(p.relayState.trusteeBitMap) == p.relayState.nTrustees) {
		log.Lvl1("Disruption Phase 1: Trustee", msg.ClientID, ", is consistent with itself, checking mismatches with all trustees...")
		p.checkMismatchingPairs()
	}

	return nil
}

/*
* Received_TRU_REL_DISRUPTION_REVEAL handles TRU_REL_DISRUPTION_REVEAL messages
* First, saves the bits reveal by the trustee.
* It checks that the bits received by the trustee matches with the ones of the disruptive round.
* For this it XORs the bits revealed together and compares it to the bit in the disruptive position.
* If there is a mismatch, the trustee is the disruptor.
* Else checks if all the reveals are received to move to next blame phase.
 */
func (p *PriFiLibRelayInstance) Received_TRU_REL_DISRUPTION_REVEAL(msg net.TRU_REL_DISRUPTION_REVEAL) error {

	log.Lvl1("Disruption Phase 1: Received bits from Trustee", msg.TrusteeID, "value", msg.Bits)

	p.relayState.trusteeBitMap[msg.TrusteeID] = msg.Bits //LB->CV: Why do you buffer it there, but not when receiving from clients ?
	result := p.compareBitsTrustee(msg.TrusteeID, msg.Bits)
	if result {
		log.Fatal("Disruption Phase 1: Disruptor is Trustee", msg.TrusteeID, ".")
	} else if (len(p.relayState.clientBitMap) == p.relayState.nClients) && (len(p.relayState.trusteeBitMap) == p.relayState.nTrustees) {
		log.Lvl1("Disruption Phase 1: Trustee", msg.TrusteeID, ", is consistent with itself, checking mismatches with all clients...")
		p.checkMismatchingPairs()
	}
	p.relayState.trusteeBitMap[msg.TrusteeID] = msg.Bits
	return nil
}

/*
* Auxiliary function that does the check of the bits revealed with the bit in the disruptive position.
* For clients.
 */
func (p *PriFiLibRelayInstance) compareBitsClient(id int, bits map[int]int) bool {
	round := p.relayState.blamingData[0]
	bitPosition := p.relayState.blamingData[1]
	bytePosition := bitPosition/8 + 9 // LB->CV: why + 9 ? avoid magic numbers :)

	log.Lvl2("Disruption: comparing", bits, "with", p.relayState.CiphertextsHistoryClients[int32(id)][int32(round)])

	byte_toGet := p.relayState.CiphertextsHistoryClients[int32(id)][int32(round)][bytePosition]
	bitInBytePosition := (8-bitPosition%8)%8 - 1
	mask := byte(1 << uint(bitInBytePosition))
	result := 0
	for _, bit := range bits {
		result ^= bit
	}
	bytePreviousResult := mask & byte_toGet
	bitPreviousResult := 0
	if bytePreviousResult != 0 {
		bitPreviousResult = 1
	}

	//LB->CV: You probably want to return "true" if equal or "false" if not
	return !(result == bitPreviousResult)
}

/*
* Auxiliary function that does the check of the bits revealed with the bit in the disruptive position.
* For trustees.
 */
func (p *PriFiLibRelayInstance) compareBitsTrustee(id int, bits map[int]int) bool {

	//LB->CV: This function seems to be identical with the one above. Can you factor out the common part and pass the correct map p.relayState.CiphertextsHistoryTrustees or p.relayState.CiphertextsHistoryClients
	//CV->LB: It will be done in next commit. This commit was just for answering the comments.

	round := p.relayState.blamingData[0]
	bitPosition := p.relayState.blamingData[1]

	bytePosition := int(bitPosition/8) + 9
	byte_toGet := p.relayState.CiphertextsHistoryTrustees[int32(id)][int32(round)][bytePosition]
	bitInBytePosition := (8-bitPosition%8)%8 - 1
	mask := byte(1 << uint(bitInBytePosition))
	result := 0
	for _, bit := range bits {
		result ^= bit
	}
	bytePreviousResult := mask & byte_toGet
	bitPreviousResult := 0
	if bytePreviousResult != 0 {
		bitPreviousResult = 1
	}
	return !(result == bitPreviousResult)
}

/*
* Once the relay have received all the bits revealed and all match the bit in the disruptive position,
* the relay checks the bits between pairsof trustees and clients.
* When a mismatch is found, the Reveal secret message is called to the client and trustee.
 */
func (p *PriFiLibRelayInstance) checkMismatchingPairs() {
	for clientID, clientBits := range p.relayState.clientBitMap {
		for trusteeID, clientBit := range clientBits {
			trusteeBit := p.relayState.trusteeBitMap[trusteeID][clientID]
			if clientBit != trusteeBit {
				log.Error("Disruption Phase 2: mismatch between trustee", trusteeID, "and client", clientID)
				p.relayState.blamingData[2] = clientID
				p.relayState.blamingData[3] = clientBit
				p.relayState.blamingData[4] = trusteeID
				p.relayState.blamingData[5] = trusteeBit

				//LB->CV: "Single Reponsability Principle": each function should do only one thing! This one "checks mismatching pairs", perfect, but so it shouldn't send messages, and this should be done in the parent function
				toClient := &net.REL_ALL_REVEAL_SHARED_SECRETS{
					EntityID: trusteeID,
				}
				toTrustee := &net.REL_ALL_REVEAL_SHARED_SECRETS{
					EntityID: clientID,
				}
				p.messageSender.SendToTrusteeWithLog(clientID, toClient, "")
				p.messageSender.SendToClientWithLog(trusteeID, toTrustee, "")
				return
			}
		}
	}
	log.Fatal("Disruption Phase 2: No mismatching pairs ? this should never occur.")
}

/*
Received_TRU_REL_SHARED_SECRETS handles TRU_REL_SECRET messages
Check the NIZK, if correct regenerate the cipher up to the disrupted round and check if this trustee is the disruptor
*/
func (p *PriFiLibRelayInstance) Received_TRU_REL_SHARED_SECRETS(msg net.TRU_REL_SHARED_SECRET) error {
	log.Lvl1("Disruption Phase 2: Received shared secret from Trustee", msg.TrusteeID, "for client", msg.ClientID, "value", msg.Secret)

	// CARLOS TODO: Check ntzk
	val := p.replayRounds(msg.Secret)
	if val != p.relayState.blamingData[5] {
		log.Fatal("Disruption Phase 2: Disruptor is Trustee", msg.TrusteeID, ".")
	} else {
		log.Lvl1("Disruption Phase 2: Trustee", msg.TrusteeID, "didn't lie, so it should be Client", msg.ClientID, ".")
	}
	return nil
}

/*
Received_CLI_REL_SHARED_SECRET handles CLI_REL_SECRET messages
Check the NIZK, if correct regenerate the cipher up to the disrupted round and check if this client is the disruptor
*/
func (p *PriFiLibRelayInstance) Received_CLI_REL_SHARED_SECRET(msg net.CLI_REL_SHARED_SECRET) error {
	log.Lvl1("Disruption Phase 2: Received shared secret from Client", msg.ClientID, "for Trustee", msg.TrusteeID, "value", msg.Secret)

	// CARLOS TODO: Check NIZK
	val := p.replayRounds(msg.Secret)
	if val != p.relayState.blamingData[3] {
		log.Fatal("Disruption Phase 2: Disruptor is Client", msg.ClientID, ".")
	} else {
		log.Lvl1("Disruption Phase 2: Client", msg.ClientID, "didn't lie, so it should be Client", msg.TrusteeID, ".")
	}
	return nil
}

/*
replayRounds takes the secret revealed by a user and recomputes until the disrupted bit
*/
func (p *PriFiLibRelayInstance) replayRounds(secret kyber.Point) int {
	seed, err := secret.MarshalBinary()
	if err != nil {
		log.Fatal("Could not extract data from shared key", err)
	}
	sharedPRNG := config.CryptoSuite.XOF(seed)

	round := 0
	disruptive_round := p.relayState.blamingData[0]
	for round < disruptive_round {

		dummy := make([]byte, p.relayState.DCNet.DCNetPayloadSize)
		sharedPRNG.XORKeyStream(dummy, dummy)
		round++

	}

	p_ij := make([]byte, p.relayState.DCNet.DCNetPayloadSize)
	sharedPRNG.XORKeyStream(p_ij, p_ij)

	var rtn int

	bitPosition := p.relayState.blamingData[1]
	bytePosition := int(bitPosition/8) + 1
	byte_toGet := p_ij[bytePosition]
	bitInByte := (8-bitPosition%8)%8 - 1
	mask := byte(1 << uint(bitInByte))
	if (byte_toGet & mask) == 0 {
		rtn = 0
	} else {
		rtn = 1
	}

	return rtn
}
