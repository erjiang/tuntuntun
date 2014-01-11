package main

import (
	"errors"
	"fmt"
	"github.com/mgutz/ansi"
	"log"
	"net"
	"tun"
)

var packet_seq uint64 = 0

var send_buf []byte = make([]byte, BUF_SIZE)

var ifs []*Iface

func client(remote_addr *net.UDPAddr, local_ifs []*net.UDPAddr) {
	log.Print("Initializing tun device")
	tundev, err := tun.Tun_alloc(tun.IFF_TUN | tun.IFF_NO_PI)
	if err != nil {
		log.Print("Could not allocate tun device")
		log.Fatal(err)
	}
	log.Print("Opened up tun device " + tundev.Name())

	log.Print("Initializing UDP connection to " + remote_addr.String())

	// TODO: listen on each iface separately to track where packets come from
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: TUNTUNTUN_CLIENT_PORT})
	if err != nil {
		log.Fatal(err)
	}

	// create list of local Ifs and store in global
	ifs = setupIfs(local_ifs)

	log.Print("Registering with server...")
	for _, iface := range ifs {
		registerClient(iface.Conn, remote_addr)
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
	fwdchan := make(chan []byte)

	go listenTun(tundev, tun_read_buf, tunchan)
	go listenUDP(conn, udp_read_buf, udpchan)
	// put packet forwarding in a separate goroutine to be able to do
	// round-robin load-balancing and more
	go forwardPacketHandler(remote_addr, fwdchan)

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
			fwdchan <- tun_read_buf[:count]
		case udpr, ok := <-udpchan:
			if !ok {
				log.Fatal("Error reading from udp")
			}
			count := udpr.Count
			remote_addr := udpr.RemoteAddr
			log.Printf("Got packet of len %d from %s", count, remote_addr)
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

func setupIfs(addrs []*net.UDPAddr) []*Iface {
	var iflist = make([]*Iface, 0)
	for i, v := range addrs {
		// try listening on if
		conn, err := net.ListenUDP("udp", v)
		if err != nil {
			log.Print("Could not listen to ", v)
			log.Fatal(err)
		}
		// create if struct and add to list
		iflist = append(iflist, &Iface{
			Name:   fmt.Sprintf("Link %d", i),
			IP:     v,
			Conn:   conn,
			Status: IFACE_STATUS_UP,
		})
		debug(1, fmt.Sprintf("Created link %s", v))
	}
	return iflist
}

func registerClient(conn *net.UDPConn, remote_addr *net.UDPAddr) error {
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

func forwardPacketHandler(remote_addr *net.UDPAddr, fwdchan chan []byte) {

	colors := []string{"magenta", "yellow", "cyan", "white:blue", "black:white"}
	ansi_colors := make([]string, len(colors))
	for i, color := range colors {
		ansi_colors[i] = ansi.ColorCode(color)
	}
	ansi_reset := ansi.ColorCode("reset")

	for {

		// round-robin by using packet sequence
		iface := ifs[packet_seq%uint64(len(ifs))]

		pkt := <-fwdchan

		err := forward_packet(iface.Conn, remote_addr, pkt)
		if err != nil {
			log.Print(err)
		}

		// log statistics in if
		iface.packets_sent++
		// does not count envelope in bytes sent
		iface.bytes_sent = iface.bytes_sent + uint64(len(pkt))

		// make colorful display of packets
		// each iface gets a different color, and we print 'S' for sent and 'R' for received
		if DEBUG_LEVEL >= 2 {
			fmt.Print(ansi_colors[int(packet_seq%uint64(len(ifs)))%len(colors)], 'S', ansi_reset)
		}

		if err != nil {
			log.Print(err)
		}
	}
}
