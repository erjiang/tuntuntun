package main

import (
	"log"
	"net"
)

var read_buf []byte = make([]byte, BUF_SIZE)

func server() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: TUNTUNTUN_SERVER_PORT})
	if err != nil {
		log.Fatal(err)
	}

	for {
		count, remote_addr, err := conn.ReadFromUDP(read_buf)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Got %d bytes from %s addressed to %s",
			count, remote_addr,
			get_ip_dest(read_buf[ENVELOPE_LENGTH:]))
	}
}
