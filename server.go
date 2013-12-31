package main

import (
	"log"
)

var read_buf []byte = make([]byte, BUF_SIZE)

func server() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: TUNTUNTUN_SERVER_PORT})
	if err != nil {
		log.Fatal(err)
	}
}
