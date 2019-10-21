# My notes

## Changing the disruption protection code

- The HMAC function is done by the client and send to the realy to check the message has arrived correctly. Now this has changed. Now is the realy who sends a hash to all clients and the one who send it should check it is correct.

Error at the beggining:

    E : (relay.(*PriFiLibRelayInstance).upstreamPhase2b_extractPayload: 385) - Warning: Disruption Protection check failed

Coment the part of the HMAC in raley, new error:

    E : (relay.(*PriFiLibRelayInstance).upstreamPhase2b_extractPayload: 451) - Relay : DecodeCell produced wrong-size payload, 5000!=5000

We decoment the following line of code and it works:

    upstreamPlaintext = upstreamPlaintext[32:]

**Now that we have removed this verification from the relay, we proceed to remove the HMAC generation in client.**

DONE

**Let's now make the relay send and hash in response**

First let's see where it should do it. --> It should be in the downstream message, but:

- *IS THERE A DOWNSTREAM MESSAGE EVERY UPSTREAM MESSAGE*
Yes
- *IS THE HASH SENT JUST BEFORE THE UPSTREAM?*
Yes

**Now let's find the function where we should do this**

    downstreamPhase1_openRoundAndSendData(...)

We need to understand the code in order to see how to add the new feature.

Make sure you are doing everything when the disruption protection is enabled

- // used if we're replaying a pcap. The first message we decode is "time0" *WHAT IS A PCAP?*

Let's try to print the message sent and received

Lets see the types of the code

*Try to do HASHForClients as a chan but did not make it*

**It does not work when de client sends [] to the realy. The realy reads [0 ... 0] (len of 5000), so the hashses do not match**

*Revise everytime they say "if disrruption protection actived"*

### Explanation 1

What the code does in `dcnet.go` is that if the payload is nil (this happend when the payload is an empty slice of bytes), it is transformed into a slice of 0 bytes of the size of the cell. Therefore, for the hash function to be equal in client and relay, we need to has this 0 byte slice and not the empty one.


*THE PROGRAM WORKS FINE FOR A WHILE AND SUDDENLY CRASHES*

*DCNetPayloadSize IS THIS ONLY THE MESSAGE?* *WHAT IF I ADD A FLAG AT THE END, IS IT INCLUDED?*