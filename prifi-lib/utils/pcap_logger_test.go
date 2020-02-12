package utils

import (
	"testing"
)

func TestPCAPLogger(t *testing.T) {

	l := NewPCAPLog()

	l.ReceivedPcap(0,0,true, 0,0,100)

	// should call print on its own
}
