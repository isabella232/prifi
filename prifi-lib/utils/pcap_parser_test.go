package utils

import (
	"log"
	"testing"
)

func TestPCAPReader(t *testing.T) {

	t.Skip()
	packets, error := ParsePCAP("../../pcap/test.pcap", 100, 0)

	if error != nil {
		log.Fatal(error)
	}

	if len(packets) != 689 {
		log.Fatal("Expected 689 packets")
	}
}

func TestPKTSReader(t *testing.T) {

	packets, error := ParsePKTS("../../pcap/test.pkts", 100, 0)

	if error != nil {
		log.Fatal(error)
	}

	if len(packets) != 12 {
		log.Fatal("Expected 12 packets")
	}
}
