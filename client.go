package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"tun"
)

var packet_seq uint64 = 0

var send_buf []byte = make([]byte, BUF_SIZE)

func client(remote_addr *net.UDPAddr) {
	log.Print("Initializing tun device")
	tundev, err := tun.Tun_alloc(tun.IFF_TUN | tun.IFF_NO_PI)
	if err != nil {
		log.Print("Could not allocate tun device")
		log.Fatal(err)
	}
	log.Print("Opened up tun device " + tundev.Name())

	log.Print("Initializing UDP connection to " + remote_addr.String())

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: TUNTUNTUN_CLIENT_PORT})
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Configuring device with ifconfig")
	err = tun.Ifconfig(tundev.Name(), TTT_CLIENT_IP, TTT_SERVER_IP)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Ready ...")

	tun_read_buf := make([]byte, BUF_SIZE)
	udp_read_buf := make([]byte, BUF_SIZE)

	// set up listening channels for udp and tun
	tunchan := make(chan int)
	udpchan := make(chan UDPRecv)

	go listenTun(tundev, tun_read_buf, tunchan)
	go listenUDP(conn, udp_read_buf, udpchan)

	for {
		select {
		case count, ok := <-tunchan:
			if !ok {
				log.Fatal("Error reading from tun")
			}
			log.Printf("Got a packet of %d bytes for %s", count,
				get_ip_dest(tun_read_buf[:count]))
			log.Printf("Sending to " + remote_addr.String())
			// pass along packet
			forward_packet(conn, remote_addr, tun_read_buf[:count])
		case udpr, ok := <-udpchan:
			if !ok {
				log.Fatal("Error reading from udp")
			}
			count := udpr.Count
			remote_addr := udpr.RemoteAddr
			log.Print("Got packet of len %d from %s", count, remote_addr)
			switch udp_read_buf[0] {
			case TTT_DATA: // packet to be forwarded
				pkt := udp_read_buf[ENVELOPE_LENGTH:count]
				// pass along packet
				tundev.Write(pkt)
			default:
				log.Print("Received packet of type ", udp_read_buf[0])
			}
		}
	}
}

func register(conn *net.UDPConn, remote_addr *net.UDPAddr) error {
    registration := []byte{TTT_REGISTER}
    _, err := conn.WriteToUDP(registration, remote_addr)
    // TODO: wait for registration acknownledgment
    return err
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
