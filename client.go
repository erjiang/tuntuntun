package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"tun"
)

var packet_seq uint64 = 0

const MY_IP string = "192.168.7.1"
const IFCONFIG string = "ifconfig tun0 192.168.7.1 pointopoint 192.168.7.2 up"
const REMOTE_IP string = "192.168.7.2"

const TUNTUNTUN_CLIENT_PORT int = 70

const BUF_SIZE uint = 2048

var send_buf []byte = make([]byte, BUF_SIZE)

func client(remote_addr *net.UDPAddr) {
	log.Print("Initializing tun device")
	tun, err := tun.Tun_alloc(tun.IFF_TUN | tun.IFF_NO_PI)
	if err != nil {
		log.Print("Could not allocate tun device")
		log.Fatal(err)
	}
	log.Print("Opened up tun device " + tun.Name())

	log.Print("Initializing UDP connection to " + remote_addr.String())

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: TUNTUNTUN_CLIENT_PORT})
	if err != nil {
		log.Fatal(err)
	}

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
			log.Printf("Got a packet of %d bytes for %s", count,
				get_ip_dest(data[:count]))
			log.Printf("Sending to " + remote_addr.String())
			forward_packet(conn, remote_addr, data[:count])
		}
	}
}

func forward_packet(conn *net.UDPConn, remote_addr *net.UDPAddr, pkt []byte) error {

	total_len := len(pkt) + ENVELOPE_LENGTH

	if uint(total_len) > BUF_SIZE {
		return errors.New(fmt.Sprintf("%d packet too long for %d", len(pkt), BUF_SIZE))
	}

	send_buf[0] = ENV_DATA
	send_buf[1] = byte(packet_seq >> 3)
	send_buf[2] = byte(packet_seq >> 2)
	send_buf[3] = byte(packet_seq >> 1)
	send_buf[4] = byte(packet_seq >> 0)
	packet_seq++

	copy(send_buf[5:], pkt)

	_, err := conn.WriteToUDP(send_buf[:total_len], remote_addr)

	if err != nil {
		return err
	}

	return nil
}

func ifconfig_tun(name, my_ip, remote_ip string) error {
	cmd := exec.Command("ifconfig", name, my_ip, "pointopoint", remote_ip, "up")

	return cmd.Run()
}
