package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

const TUNTUNTUN_CLIENT_PORT int = 70
const TTT_CLIENT_IP string = "192.168.7.1"
const TTT_SERVER_IP string = "192.168.7.2"
const TUNTUNTUN_SERVER_PORT int = 71

const BUF_SIZE uint = 2048

var DEBUG_LEVEL int = 0

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("server or client?")
		return
	}

	// check if DEBUG is a valid number, and if so
	// put it in the global DEBUG_LEVEL var
	env_debug, err := strconv.ParseInt(os.Getenv("DEBUG_LEVEL"), 10, 32)
	if err == nil && env_debug > 0 {
		DEBUG_LEVEL = int(env_debug)
	}

	if os.Args[1] == "client" {
		if len(os.Args) < 3 {
			fmt.Printf("Remote server addr?")
			return
		}

		var local_ifs = make([]*net.UDPAddr, len(os.Args[3:]))
		for i, addr := range os.Args[3:] {
			local_ifs[i], err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", addr, TUNTUNTUN_CLIENT_PORT))
			if err != nil {
				log.Fatal(err)
			}
		}
		remote_addr, err := net.ResolveUDPAddr("udp", os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		client(remote_addr, local_ifs)
	} else {
		server()
	}
}

func debug(lvl int, msg ...interface{}) {
	if DEBUG_LEVEL >= lvl {
		log.Print(msg...)
	}
}

func debugf(lvl int, format string, stuff ...interface{}) {
	if DEBUG_LEVEL >= lvl {
		log.Printf(format, stuff...)
	}
}
