package main

import (
	"log"
	"os/exec"
	"tun"
)

const MY_IP string = "192.168.7.1"
const IFCONFIG string = "ifconfig tun0 192.168.7.1 pointopoint 192.168.7.2 up"
const REMOTE_IP string = "192.168.7.2"

const BUF_SIZE uint = 2048

func client() {
	log.Print("Initializing tun device")
	tun, err := tun.Tun_alloc(tun.IFF_TUN | tun.IFF_NO_PI)
	if err != nil {
		log.Print("Could not allocate tun device")
		log.Fatal(err)
	}
	log.Print("Opened up tun device " + tun.Name())

	log.Print("Configuring device with ifconfig")
	err = ifconfig_tun(tun.Name(), MY_IP, REMOTE_IP)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Ready ...")

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

func ifconfig_tun(name, my_ip, remote_ip string) error {
	cmd := exec.Command("ifconfig", name, my_ip, "pointopoint", remote_ip, "up")

	return cmd.Run()
}
