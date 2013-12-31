package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("server or client?")
		return
	}

	if os.Args[1] == "client" {
		if len(os.Args) < 3 {
			fmt.Printf("Remote server addr?")
			return
		}

		remote_addr, err := net.ResolveUDPAddr("udp", os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		client(remote_addr)
	} else {
		server()
	}
}
