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
	p.relayState.clientBitMap[msg.ClientID] = msg.Bits
	result := p.compareBits(msg.ClientID, msg.Bits, p.relayState.CiphertextsHistoryClients)
	if result {
		log.Fatal("DISRUPTOR IS CLIENT", msg.ClientID, ". Detected in first phase of the blame protocol.")
	} else if (len(p.relayState.clientBitMap) == p.relayState.nClients) && (len(p.relayState.trusteeBitMap) == p.relayState.nTrustees) {
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
	p.relayState.trusteeBitMap[msg.TrusteeID] = msg.Bits
	result := p.compareBits(msg.TrusteeID, msg.Bits, p.relayState.CiphertextsHistoryTrustees)
	if result {
		log.Fatal("DISRUPTOR IS TRUSTEE", msg.TrusteeID, ". Detected in first phase of the blame protocol.")
	} else if (len(p.relayState.clientBitMap) == p.relayState.nClients) && (len(p.relayState.trusteeBitMap) == p.relayState.nTrustees) {
		p.checkMismatchingPairs()
	}
	p.relayState.trusteeBitMap[msg.TrusteeID] = msg.Bits
	return nil
}

/*
* Auxiliary function that does the check of the bits revealed with the bit in the disruptive position.
 */
func (p *PriFiLibRelayInstance) compareBits(id int, bits map[int]int, CiphertextsHistory map[int32][][]byte) bool {
	round := p.relayState.blamingData[0]
	bitPosition := p.relayState.blamingData[1]

	bytePosition := int(bitPosition/8) + 9
	byte_toGet := CiphertextsHistory[int32(id)][int32(round)][bytePosition]
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
				log.Error("Disruption: mismatch between trustee", trusteeID, "and client", clientID)
				p.relayState.blamingData[2] = clientID
				p.relayState.blamingData[3] = clientBit
				p.relayState.blamingData[4] = trusteeID
				p.relayState.blamingData[5] = trusteeBit
				toClient := &net.REL_ALL_REVEAL_SHARED_SECRETS{
					UserID: trusteeID,
				}
				toTrustee := &net.REL_ALL_REVEAL_SHARED_SECRETS{
					UserID: clientID,
				}
				p.messageSender.SendToTrusteeWithLog(clientID, toClient, "")
				p.messageSender.SendToClientWithLog(trusteeID, toTrustee, "")
				return
			}
		}
	}
	log.Fatal("NO DISRUPTION ?")
}

/*
Received_TRU_REL_SHARED_SECRETS handles TRU_REL_SECRET messages
Check the NIZK, if correct regenerate the cipher up to the disrupted round and check if this trustee is the disruptor
*/
func (p *PriFiLibRelayInstance) Received_TRU_REL_SHARED_SECRETS(msg net.TRU_REL_SHARED_SECRET) error {
	// CARLOS TODO: Check ntzk
	val := p.replayRounds(msg.Secret)
	if val != p.relayState.blamingData[5] {
		log.Fatal("DISRUPTOR IS TRUSTEE", p.relayState.blamingData[4], ". Detected in second phase of the blame protocol.")
	}
	return nil
}

/*
Received_CLI_REL_SHARED_SECRET handles CLI_REL_SECRET messages
Check the NIZK, if correct regenerate the cipher up to the disrupted round and check if this client is the disruptor
*/
func (p *PriFiLibRelayInstance) Received_CLI_REL_SHARED_SECRET(msg net.CLI_REL_SHARED_SECRET) error {
	// CARLOS TODO: Check ntzk
	val := p.replayRounds(msg.Secret)
	if val != p.relayState.blamingData[3] {
		log.Fatal("DISRUPTOR IS CLIENT", p.relayState.blamingData[2], ". Detected in second phase of the blame protocol.")
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
