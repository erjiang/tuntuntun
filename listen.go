package main

import (
	"log"
	"net"
	"os"
)

type UDPRecv struct {
	Data       []byte // UDP payload
	RemoteAddr *net.UDPAddr
}

// listenUDP goroutine has its own internal read buf that it gives a slice into
// when it receives a packet. It passes the slice and the source address into
// the channel.
// TODO: make this take an Iface so it can log statistics
func listenUDP(conn *net.UDPConn, c chan UDPRecv) error {
	read_buf := make([]byte, BUF_SIZE)
	for {
		count, remote_addr, err := conn.ReadFromUDP(read_buf)
		if err != nil {
			log.Print(err)
			return err
		}
		c <- UDPRecv{
			Data:       read_buf[:count],
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
