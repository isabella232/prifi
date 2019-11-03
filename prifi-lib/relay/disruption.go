package relay

import (
	"github.com/dedis/prifi/prifi-lib/net"
	"github.com/dedis/prifi/prifi-lib/config"
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
	log.Error("REVEAL", msg.ClientID, msg.Bits)
	p.relayState.clientBitMap[msg.ClientID] = msg.Bits
	result := p.compareBitsClient(msg.ClientID, msg.Bits)
	if result {
		log.Fatal("DISRUPTOR IS CLIENT", msg.ClientID)
	}else if (len(p.relayState.clientBitMap) == p.relayState.nClients) && (len(p.relayState.trusteeBitMap) == p.relayState.nTrustees) {
		p.checkMismachesPairs()
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
	log.Error("REVEAL", msg.TrusteeID, msg.Bits)
	p.relayState.trusteeBitMap[msg.TrusteeID] = msg.Bits
	result := p.compareBitsTrustee(msg.TrusteeID, msg.Bits)
	if result {
		log.Fatal("DISRUPTOR IS TRUSTEE", msg.TrusteeID)
	}else if (len(p.relayState.clientBitMap) == p.relayState.nClients) && (len(p.relayState.trusteeBitMap) == p.relayState.nTrustees) {
		p.checkMismachesPairs()
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

	bytePosition := int(bitPosition/8) + 9
	byte_toGet := p.relayState.ChiperHistoryClient[int32(id)][int32(round)][bytePosition]
	bitInBytePosition := (8 - bitPosition%8)%8 - 1
	mask := byte(1 << uint(bitInBytePosition))
	log.Lvl1(bitPosition, bytePosition, bitInBytePosition, byte_toGet, mask, mask & byte_toGet)
	result := 0
	for _, bit := range(bits){
		result ^= bit
	}
	bytePreviousResult := mask & byte_toGet
	bitPreviousResult := 0
	if bytePreviousResult != 0{
		bitPreviousResult = 1
	}
	return  !(result == bitPreviousResult)
}

/*
* Auxiliary function that does the check of the bits revealed with the bit in the disruptive position. 
* For trustees.
*/
func (p *PriFiLibRelayInstance) compareBitsTrustee(id int, bits map[int]int) bool {
	round := p.relayState.blamingData[0]
	bitPosition := p.relayState.blamingData[1]

	bytePosition := int(bitPosition/8) + 9
	byte_toGet := p.relayState.ChiperHistoryTrustee[int32(id)][int32(round)][bytePosition]
	bitInBytePosition := (8 - bitPosition%8)%8 - 1
	mask := byte(1 << uint(bitInBytePosition))
	log.Lvl1(bitPosition, bytePosition, bitInBytePosition, byte_toGet, mask, mask & byte_toGet)
	result := 0
	for _, bit := range(bits){
		result ^= bit
	}
	bytePreviousResult := mask & byte_toGet
	bitPreviousResult := 0
	if bytePreviousResult != 0{
		bitPreviousResult = 1
	}
	return  !(result == bitPreviousResult)
}

/*
* Once the relay have received all the bits revealed and all match the bit in the disruptive position,
* the relay checks the bits between pairsof trustees and clients.
* When a mismatch is found, the Reveal secret message is called to the client and trustee.
*/
func (p *PriFiLibRelayInstance) checkMismachesPairs() {
	for clientID, clientBits := range p.relayState.clientBitMap {
		for trusteeID, clientBit := range clientBits {
			trusteeBit := p.relayState.trusteeBitMap[trusteeID][clientID];
			if clientBit != trusteeBit {
				log.Error("MISMATCH BETWEEN TRUSTEE", trusteeID, "AND CLIENT", clientID)
				p.relayState.blamingData[2] = clientID
				p.relayState.blamingData[3] = clientBit
				p.relayState.blamingData[4] = trusteeID
				p.relayState.blamingData[5] = trusteeBit
				toClient := &net.REL_ALL_DISRUPTION_SECRET{
					UserID: trusteeID,
				}
				toTrustee := &net.REL_ALL_DISRUPTION_SECRET{
					UserID: clientID,
				}
				p.messageSender.SendToTrusteeWithLog(clientID, toClient, "")
				p.messageSender.SendToClientWithLog(trusteeID, toTrustee, "")
				return
			}
		}
	}
	log.Fatal("NO DISRRUPTION ?")

}

/*
Received_TRU_REL_SECRET handles TRU_REL_SECRET messages
Check the NIZK, if correct regenerate the cipher up to the disrupted round and check if this trustee is the disruptor
*/
func (p *PriFiLibRelayInstance) Received_TRU_REL_SECRET(msg net.TRU_REL_DISRUPTION_SECRET) error {
	// CARLOS TODO: Check ntzk
	val := p.replayRounds(msg.Secret)
	log.Error("RECEIVED TRUSTEE", msg.Secret, val, p.relayState.blamingData)
	if val != p.relayState.blamingData[5] {
		log.Fatal("Trustee ", p.relayState.blamingData[4], " lied and is considered a disruptor")
	}
	return nil
}

/*
Received_CLI_REL_SECRET handles CLI_REL_SECRET messages
Check the NIZK, if correct regenerate the cipher up to the disrupted round and check if this client is the disruptor
*/
func (p *PriFiLibRelayInstance) Received_CLI_REL_SECRET(msg net.CLI_REL_DISRUPTION_SECRET) error {
	// CARLOS TODO: Check ntzk
	log.Error("RECEIVED CLIENT", msg.Secret)
	val := p.replayRounds(msg.Secret)
	log.Error("RECEIVED CLIENT", msg.Secret, val, p.relayState.blamingData)
	if val != p.relayState.blamingData[3] {
		log.Fatal("Client ", p.relayState.blamingData[2], " lied and is considered a disruptor")
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
	byte_2 := p_ij[bytePosition+1]
	bitInByte := (8 - bitPosition%8)%8 - 1
	mask := byte(1 << uint(bitInByte))
	log.Lvl1(bitPosition, bytePosition, bitInByte,byte_2, byte_toGet, mask, mask & byte_toGet)
	if (byte_toGet & mask) == 0{
		rtn = 0
	}else{
		rtn = 1
	}

	return rtn
	/*bytes, err := secret.MarshalBinary()
	if err != nil {
		log.Fatal("Could not marshal point !")
	}
	roundID := p.relayState.blamingData[0]
	sharedPRNG := config.CryptoSuite.XOF(bytes)
	key := make([]byte, config.CryptoSuite.XOF(nil).KeySize())
	sharedPRNG.Partial(key, key, nil)
	dcCipher := config.CryptoSuite.XOF(key)

	for i := 0; i < roundID; i++ {
		//discard crypto material
		dst := make([]byte, p.relayState.PayloadSize)
		dcCipher.Read(dst)
	}

	dst := make([]byte, p.relayState.PayloadSize)
	dcCipher.Read(dst)
	bitPos := p.relayState.blamingData[0]
	m := float64(bitPos) / float64(8)
	m = math.Floor(m)
	m2 := int(m)
	n := bitPos % 8
	mask := byte(1 << uint8(n))
	if (dst[m2] & mask) == 0 {
		return 0
	}
	
	log.Fatal("not implemented")
	return 1*/
}
