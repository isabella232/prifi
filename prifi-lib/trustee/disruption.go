package trustee

import (
	"fmt"
	"github.com/dedis/prifi/prifi-lib/config"
	"github.com/dedis/prifi/prifi-lib/net"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/proof"
	"gopkg.in/dedis/onet.v2/log"
	"strconv"
)

/*
* Received_REL_ALL_DISRUPTION_REVEAL handles REL_ALL_DISRUPTION_REVEAL messages.
* The method calls a function from the DCNet to regenerate the bits from roundID in position BitPos
* The result is sent to the relay.
 */
func (p *PriFiLibTrusteeInstance) Received_REL_ALL_DISRUPTION_REVEAL(msg net.REL_ALL_DISRUPTION_REVEAL) error {
	log.Lvl1("Disruption Phase 1: Received de-anonymization query for round", msg.RoundID, "bit pos", msg.BitPos)
	bitMap, PRGs := p.trusteeState.DCNet.GetBitsOfRound(int32(msg.RoundID), int32(msg.BitPos))

	var pred_array []proof.Predicate
	sval := make(map[string]kyber.Scalar)
	pval := make(map[string]kyber.Point)
	suite := config.CryptoSuite
	B := suite.Point().Base()
	pval["B"] = B
	for i, prg := range PRGs {
		i_string := strconv.Itoa(i)
		pred_array = append(pred_array, proof.Rep("T"+i_string, "t"+i_string, "B"))
		p_i := suite.Scalar().SetBytes(prg)
		P_i := suite.Point().Mul(p_i, B)
		sval["t"+i_string] = p_i
		pval["T"+i_string] = P_i
	}
	pred := proof.And(pred_array...)

	prover := pred.Prover(suite, sval, pval, nil)
	NIZK, _ := proof.HashProve(suite, "DISRUPTION", prover)

	verifier := pred.Verifier(suite, pval)
	err := proof.HashVerify(suite, "DISRUPTION", verifier, NIZK)
	if err != nil {
		log.Error("EE Proof failed to verify: ")
	}
	log.Lvl1("EE Proof verified.")
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
	// TODO: check that the relay asks for the correct entity, and not a honest entity. There should be a signature check on the TRU_REL_DISRUPTION_REVEAL the relay received (and forwarded to the client)
	secret := p.trusteeState.sharedSecrets[msg.EntityID]
	// as a pseudorandom base point multiplied by our private key.
	suite := config.CryptoSuite
	X := make([]kyber.Point, 1)
	X[0] = p.trusteeState.PublicKey
	B := suite.Point().Base() //BACK
	// Generate the proof predicate: an OR branch for each public key.
	sec := map[string]kyber.Scalar{"x": p.trusteeState.privateKey} //BACK
	pub := map[string]kyber.Point{"B": B, "BT": p.trusteeState.ClientPublicKeys[msg.EntityID], "T": p.trusteeState.sharedSecrets[msg.EntityID]}
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

	toSend := &net.TRU_REL_SHARED_SECRET{
		TrusteeID: p.trusteeState.ID,
		ClientID:  msg.EntityID,
		Secret:    secret,
		NIZK:      NIZK,
		Pub:       pub,
	}
	p.messageSender.SendToRelayWithLog(toSend, "Sent secret to relay")
	log.Lvl1("Reveling secret with client", msg.EntityID)
	return nil
}
