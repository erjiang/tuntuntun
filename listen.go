package main

import (
	"log"
	"net"
	"os"
)

type UDPRecv struct {
	Count int
	RemoteAddr *net.UDPAddr
}

func listenUDP(conn *net.UDPConn, read_buf []byte, c chan UDPRecv) error {
	for {
		count, remote_addr, err := conn.ReadFromUDP(read_buf)
		if err != nil {
			log.Print(err)
			return err
		}
		c <- UDPRecv{
			Count: count,
			RemoteAddr: remote_addr,
		}
	}
}

func listenTun(tundev *os.File, read_buf []byte, c chan int) error {
	for {
		count, err := tundev.Read(read_buf)
		if err != nil {
			log.Print(err)
			return err
		}
		c <- count
	}
}
