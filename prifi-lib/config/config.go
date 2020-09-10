/*
Package config contains the cryptographic primitives that are used by the PriFi library.
*/
package config

import (
	"go.dedis.ch/kyber/v3/suites"
)

// the suite used in the prifi-lib
var CryptoSuite = suites.MustFind("Ed25519")
