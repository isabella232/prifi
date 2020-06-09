package relay

import (
	"github.com/dedis/prifi/prifi-lib/config"
	"github.com/dedis/prifi/prifi-lib/net"
	"go.dedis.ch/kyber"
	"go.dedis.ch/onet/log"

	"fmt"
	"go.dedis.ch/kyber/proof"
	"strconv"
)

// Received_CLI_REL_BLAME
func (p *PriFiLibRelayInstance) Received_CLI_REL_DISRUPTION_BLAME(msg net.CLI_REL_DISRUPTION_BLAME) error {
	// TODO: Check NIZK
	pred := proof.Rep("X", "x", "B")
	suite := config.CryptoSuite
	//B := suite.Point().Base()
	/*for _, key := range(p.relayState.EphemeralPublicKeys) {
		pval := map[string]kyber.Point{"B": B, "X": key}
		verifier := pred.Verifier(suite, pval)
		err := proof.HashVerify(suite, "DISRUPTION", verifier, msg.NIZK)
		if err != nil {
			log.Lvl1("Proof failed to verify: ", key)
			continue
		}
		log.Lvl1("Proof verified.", key)
	}*/
	verifier := pred.Verifier(suite, msg.Pval)
	err := proof.HashVerify(suite, "DISRUPTION", verifier, msg.NIZK)
	if err != nil {
		log.Fatal("Proof failed to verify: ")
	}
	log.Lvl1("Proof verified.")

	// TODO: p.stateMachine.ChangeState("BLAMING")

	toSend := &net.REL_ALL_DISRUPTION_REVEAL{
		RoundID: msg.RoundID,
		BitPos:  msg.BitPos,
		Pval:    msg.Pval,
		NIZK:    msg.NIZK,
	}
	p.relayState.blamingData.RoundID = msg.RoundID
	p.relayState.blamingData.BitPos = msg.BitPos

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
	var pred_array []proof.Predicate
	suite := config.CryptoSuite
	for i := 1; i < p.relayState.nTrustees; i++ {
		i_string := strconv.Itoa(i)
		pred_array = append(pred_array, proof.Rep("T"+i_string, "t"+i_string, "B"))
	}
	pred := proof.And(pred_array...)
	verifier := pred.Verifier(suite, msg.Pval)
	err := proof.HashVerify(suite, "DISRUPTION", verifier, msg.NIZK)
	if err != nil {
		log.Fatal("Proof failed to verify: ")
	}
	log.Lvl3("Proof verified.")

	p.relayState.clientBitMap[msg.ClientID] = msg.Bits

	result := p.compareBits(msg.ClientID, msg.Bits, p.relayState.CiphertextsHistoryClients)
	if !result {
		log.Fatal("Disruption Phase 1: Disruptor is Client", msg.ClientID, ".")
	} else if (len(p.relayState.clientBitMap) == p.relayState.nClients) && (len(p.relayState.trusteeBitMap) == p.relayState.nTrustees) {
		log.Lvl1("Disruption Phase 1: Trustee", msg.ClientID, ", is consistent with itself, checking mismatches with all trustees...")
		mismatch := p.checkMismatchingPairs()
		if mismatch {
			toClient := &net.REL_ALL_REVEAL_SHARED_SECRETS{
				EntityID: p.relayState.blamingData.TrusteeID,
			}
			toTrustee := &net.REL_ALL_REVEAL_SHARED_SECRETS{
				EntityID: p.relayState.blamingData.ClientID,
			}
			p.messageSender.SendToTrusteeWithLog(p.relayState.blamingData.ClientID, toTrustee, "")
			p.messageSender.SendToClientWithLog(p.relayState.blamingData.TrusteeID, toClient, "")
		} else {
			log.Fatal("Disruption Phase 2: No mismatching pairs ? this should never occur.")
		}
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

	var pred_array []proof.Predicate
	suite := config.CryptoSuite
	for i := 1; i < p.relayState.nTrustees; i++ {
		i_string := strconv.Itoa(i)
		pred_array = append(pred_array, proof.Rep("T"+i_string, "t"+i_string, "B"))
	}
	pred := proof.And(pred_array...)
	verifier := pred.Verifier(suite, msg.Pval)
	err := proof.HashVerify(suite, "DISRUPTION", verifier, msg.NIZK)
	if err != nil {
		log.Fatal("Proof failed to verify: ")
	}
	log.Lvl3("Proof verified.")

	p.relayState.trusteeBitMap[msg.TrusteeID] = msg.Bits
	result := p.compareBits(msg.TrusteeID, msg.Bits, p.relayState.CiphertextsHistoryTrustees)
	if !result {
		log.Fatal("Disruption Phase 1: Disruptor is Trustee", msg.TrusteeID, ".")
	} else if (len(p.relayState.clientBitMap) == p.relayState.nClients) && (len(p.relayState.trusteeBitMap) == p.relayState.nTrustees) {
		log.Lvl1("Disruption Phase 1: Trustee", msg.TrusteeID, ", is consistent with itself, checking mismatches with all clients...")
		mismatch := p.checkMismatchingPairs()
		if mismatch {
			toClient := &net.REL_ALL_REVEAL_SHARED_SECRETS{
				EntityID: p.relayState.blamingData.TrusteeID,
			}
			toTrustee := &net.REL_ALL_REVEAL_SHARED_SECRETS{
				EntityID: p.relayState.blamingData.ClientID,
			}
			p.messageSender.SendToTrusteeWithLog(p.relayState.blamingData.ClientID, toTrustee, "")
			p.messageSender.SendToClientWithLog(p.relayState.blamingData.TrusteeID, toClient, "")
		} else {
			log.Fatal("Disruption Phase 2: No mismatching pairs ? this should never occur.")
		}
	}

	return nil
}

/*
* Auxiliary function that does the check of the bits revealed with the bit in the disruptive position.
 */
func (p *PriFiLibRelayInstance) compareBits(id int, bits map[int]int, CiphertextsHistory map[int32]map[int32][]byte) bool {
	round := p.relayState.blamingData.RoundID
	bitPosition := p.relayState.blamingData.BitPos
	bytePosition := bitPosition/8 + 9 // LB->CV: why + 9 ? avoid magic numbers :)

	log.Lvl2("Disruption: comparing", bits, "with", CiphertextsHistory[int32(id)][int32(round)])

	byteToGet := CiphertextsHistory[int32(id)][int32(round)][bytePosition]
	bitInBytePosition := (8-bitPosition%8)%8 - 1
	mask := byte(1 << uint(bitInBytePosition))
	result := 0
	for _, bit := range bits {
		result ^= bit
	}
	bytePreviousResult := mask & byteToGet
	bitPreviousResult := 0
	if bytePreviousResult != 0 {
		bitPreviousResult = 1
	}

	return (result == bitPreviousResult)
}

/*
* Once the relay have received all the bits revealed and all match the bit in the disruptive position,
* the relay checks the bits between pairsof trustees and clients.
* When a mismatch is found, the Reveal secret message is called to the client and trustee.
 */
func (p *PriFiLibRelayInstance) checkMismatchingPairs() bool {
	for clientID, clientBits := range p.relayState.clientBitMap {
		for trusteeID, clientBit := range clientBits {
			trusteeBit := p.relayState.trusteeBitMap[trusteeID][clientID]
			if clientBit != trusteeBit {
				log.Error("Disruption Phase 2: mismatch between trustee", trusteeID, "and client", clientID)
				p.relayState.blamingData.ClientID = clientID
				p.relayState.blamingData.ClientBitRevealed = clientBit
				p.relayState.blamingData.TrusteeID = trusteeID
				p.relayState.blamingData.TrusteeBitRevealed = trusteeBit
				return true
			}
		}
	}
	return false
}

/*
Received_TRU_REL_SHARED_SECRETS handles TRU_REL_SECRET messages
Check the NIZK, if correct regenerate the cipher up to the disrupted round and check if this trustee is the disruptor
*/
func (p *PriFiLibRelayInstance) Received_TRU_REL_SHARED_SECRETS(msg net.TRU_REL_SHARED_SECRET) error {
	log.Lvl1("Disruption Phase 2: Received shared secret from Trustee", msg.TrusteeID, "for client", msg.ClientID, "value", msg.Secret)

	M := "SHAREDKEY"
	X := make([]kyber.Point, 1)
	X[0] = p.relayState.trustees[msg.TrusteeID].PublicKey
	preds := make([]proof.Predicate, len(X))
	for i := range X {
		name := fmt.Sprintf("X[%d]", i) // "X[0]","X[1]",...
		msg.Pub[name] = X[i]            // public point value

		// Predicate indicates knowledge of the private key for X[i]
		// and correspondence of the key with the linkage tag
		preds[i] = proof.And(proof.Rep(name, "x", "B"), proof.Rep("T", "x", "BT"))
	}
	pred := proof.Or(preds...) // make a big Or predicate
	suite := config.CryptoSuite
	// Verify the signature
	verifier := pred.Verifier(suite, msg.Pub)
	err := proof.HashVerify(suite, M, verifier, msg.NIZK)
	if err != nil {
		log.Error("signature failed to verify: ", err)
	}
	log.Lvl3("Linkable Ring Signature verified.")

	val := p.replayRounds(msg.Secret)
	if val != p.relayState.blamingData.TrusteeBitRevealed {
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

	M := "SHAREDKEY"
	X := make([]kyber.Point, 1)
	X[0] = p.relayState.trustees[msg.TrusteeID].PublicKey
	preds := make([]proof.Predicate, len(X))
	for i := range X {
		name := fmt.Sprintf("X[%d]", i) // "X[0]","X[1]",...
		msg.Pub[name] = X[i]            // public point value

		// Predicate indicates knowledge of the private key for X[i]
		// and correspondence of the key with the linkage tag
		preds[i] = proof.And(proof.Rep(name, "x", "B"), proof.Rep("T", "x", "BT"))
	}
	pred := proof.Or(preds...) // make a big Or predicate
	suite := config.CryptoSuite
	// Verify the signature
	verifier := pred.Verifier(suite, msg.Pub)
	err := proof.HashVerify(suite, M, verifier, msg.NIZK)
	if err != nil {
		log.Error("signature failed to verify: ", err)
	}
	log.Lvl3("Linkable Ring Signature verified.")

	val := p.replayRounds(msg.Secret)
	if val != p.relayState.blamingData.ClientBitRevealed {
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

	round := int32(0)
	disruptive_round := p.relayState.blamingData.RoundID
	for round < disruptive_round {

		dummy := make([]byte, p.relayState.DCNet.DCNetPayloadSize)
		sharedPRNG.XORKeyStream(dummy, dummy)
		round++

	}

	p_ij := make([]byte, p.relayState.DCNet.DCNetPayloadSize)
	sharedPRNG.XORKeyStream(p_ij, p_ij)

	var rtn int

	bitPosition := p.relayState.blamingData.BitPos
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
