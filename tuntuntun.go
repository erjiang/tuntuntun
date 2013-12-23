package main

import (
	"log"
	"tun"
)

const BUF_SIZE uint = 2048

func main() {
	tun, err := tun.Tun_alloc(tun.IFF_TUN | tun.IFF_NO_PI)
	if err != nil {
		log.Print("Could not allocate tun device")
		log.Print(err)
	}
	data := make([]byte, BUF_SIZE)
	for {
		count, err := tun.Read(data)
		if err != nil {
			log.Print(err)
		} else {
			log.Printf("Got a packet of %d bytes", count)
		}
	}
}
