package client

import (
	"bytes"
	"github.com/dedis/prifi/prifi-lib/config"
	"github.com/dedis/prifi/prifi-lib/net"
	"go.dedis.ch/kyber/proof"
	"gopkg.in/dedis/onet.v2/log"
	"time"

	"fmt"
	"go.dedis.ch/kyber"
)

/*
* Received_REL_ALL_DISRUPTION_REVEAL handles REL_ALL_DISRUPTION_REVEAL messages.
* The method calls a function from the DCNet to regenerate the bits from roundID in position BitPos
* The result is sent to the relay.
 */
func (p *PriFiLibClientInstance) Received_REL_ALL_DISRUPTION_REVEAL(msg net.REL_ALL_DISRUPTION_REVEAL) error {
	log.Lvl1("Disruption Phase 1: Received de-anonymization query for round", msg.RoundID, "bit pos", msg.BitPos)

	// TODO: check the proper NIZK
	//CALROS
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

	// as a pseudorandom base point multiplied by our private key.
	suite := config.CryptoSuite
	X := make([]kyber.Point, 1)
	X[0] = p.clientState.PublicKey
	B := suite.Point().Base() //BACK
	// Generate the proof predicate: an OR branch for each public key.
	sec := map[string]kyber.Scalar{"x": p.clientState.privateKey} //BACK
	pub := map[string]kyber.Point{"B": B, "BT": p.clientState.TrusteePublicKey[msg.EntityID], "T": p.clientState.sharedSecrets[msg.EntityID]}
	preds := make([]proof.Predicate, len(X))
	for i := range X {
		name := fmt.Sprintf("X[%d]", i) // "X[0]","X[1]",...
		pub[name] = X[i]                // public point value

		// Predicate indicates knowledge of the private key for X[i]
		// and correspondence of the key with the linkage tag
		preds[i] = proof.And(proof.Rep(name, "x", "B"), proof.Rep("T", "x", "BT"))
	}
	pred := proof.Or(preds...) // make a big Or predicate

	// The prover needs to know which Or branch (mine) is actually true.
	choice := make(map[proof.Predicate]int)
	choice[pred] = 0

	// Generate the signature
	M := "SHAREDKEY"
	prover := pred.Prover(suite, sec, pub, choice)
	NIZK, _ := proof.HashProve(suite, M, prover)

	// Verify the signature
	verifier := pred.Verifier(suite, pub)
	err := proof.HashVerify(suite, M, verifier, NIZK)
	if err != nil {
		log.Lvl1("signature failed to verify: ", err)
	}
	log.Lvl1("Linkable Ring Signature verified.")

	toSend := &net.CLI_REL_SHARED_SECRET{
		ClientID:  p.clientState.ID,
		TrusteeID: msg.EntityID, 
		Secret:    secret,
		NIZK:      make([]byte, 0),
		Pub: 	   pub,
	}

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
	log.Lvl1("POSSIBLE DISRUPTION!", p.clientState.MyLastRound, p.clientState.RoundNo-1)
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
				log.Lvl1("COMPARING", msg.Data, p.clientState.LastMessage)
				for index, b := range msg.Data {
					if b != p.clientState.LastMessage[index] {
						log.Lvl1("MISMATCH", b, p.clientState.LastMessage[index])
						log.Lvl1(index)
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
					log.Error("Disruption protection hash comparison failed.", p.clientState.RoundNo)
					p.clientState.B_echo_last = 1
				} else {
					p.clientState.B_echo_last = 0
				}
			}
		}
	}

	return nil
}
