package main

import (
	"flag"
	"github.com/armon/go-socks5"
	"go.dedis.ch/onet/log"
	"strconv"
)

const defaultBugLevel = 1
const defaultPort = 8090

var onetDebugLevels = []int{1, 2, 3, 4, 5}

// Launch a SOCKS5 server that listens to PriFi traffic and forwards all connections
func main() {

	// Command-line flags
	var debugFlag = flag.Int("debug", defaultBugLevel, "debug-level")
	var portFlag = flag.Int("port", defaultPort, "port")
	flag.Parse()
	log.SetDebugVisible(*debugFlag)

	// Check if the port is valid
	if *portFlag <= 1024 {
		log.Lvl1("Port number below 1024. Without super-admin privileges, this server will crash.")
	}

	if *portFlag > 65535 {
		log.Fatal("Port number above 65535. Exiting...")
	}

	// Construct the correct server address (for example :8090)
	port := ":" + strconv.Itoa(*portFlag)

	log.Lvl2("Starting a SOCKS5 server...")

	// Create a SOCKS5 server
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", port); err != nil {
		log.Fatal("Could not listen on port", port, "error is", err)
	}
}

func contains(intSlice []int, searchInt int) bool {
	for _, value := range intSlice {
		if value == searchInt {
			return true
		}
	}
	return false
}
