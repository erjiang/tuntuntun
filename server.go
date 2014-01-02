package main

import (
	"log"
	"net"
	"syscall"
	"tun"
)

func server() {
	tun_ip := &net.IP{192, 168, 7, 2}
	log.Print("External IP is ", tun_ip.String())

	log.Print("Initializing tun device")
	tundev, err := tun.Tun_alloc(tun.IFF_TUN | tun.IFF_NO_PI)
	if err != nil {
		log.Print("Could not allocate tun device")
		log.Fatal(err)
	}
	log.Print("Opened up tun device " + tundev.Name())

	log.Print("Configuring device with ifconfig")
	err = tun.Ifconfig(tundev.Name(), TTT_SERVER_IP, TTT_CLIENT_IP)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Listening on 0.0.0.0:", TUNTUNTUN_SERVER_PORT)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: TUNTUNTUN_SERVER_PORT})
	if err != nil {
		log.Fatal(err)
	}

	udp_read_buf := make([]byte, BUF_SIZE)
	tun_read_buf := make([]byte, BUF_SIZE)

	// set up listening channels for udp and tun
	tunchan := make(chan int)
	udpchan := make(chan UDPRecv)

	go listenTun(tundev, tun_read_buf, tunchan)
	go listenUDP(conn, udp_read_buf, udpchan)

	for {
		select {
		// listenTun sends the count of bytes
		case tlen, ok := <-tunchan:
			if !ok {
				log.Fatal("Error reading from tun")
			}
			log.Printf("Got %d bytes from tundev", tlen)
			log.Print("[FIXME] send data to client")
		// listenUDP sends a struct with byte count and remote_addr
		case udpr, ok := <-udpchan:
			if !ok {
				log.Fatal("Error reading from udp")
			}
			count := udpr.Count
			remote_addr := udpr.RemoteAddr
			switch udp_read_buf[0] {
			case 1: // packet to be forwarded
				pkt := udp_read_buf[ENVELOPE_LENGTH:count]

				log.Printf("Got %d bytes from %s addressed to %s",
					count, remote_addr,
					get_ip_dest(pkt))

				log.Printf("Sending through tun device")
				/* this breaks routing
				replace_src_addr(pkt, *tun_ip)
				ReplaceIPHeaderChecksum(pkt)
				*/
				tundev.Write(udp_read_buf[ENVELOPE_LENGTH:count])
			case 2: // registration
				log.Print("Received registration from ", remote_addr)
			default:
				log.Print("Received packet of type ", udp_read_buf[0])
			}
		}
	}

}

// Tries to figure out what IP address is the external address
// by looking up Google's DNS service (8.8.8.8) and seeing which
// interface gets used for it
// DEPRECATED: linux will handle routing anyways
func get_ext_addr() (*net.IP, error) {
	googDns := &syscall.SockaddrInet4{
		Port: 53, // doesn't really matter
		Addr: [4]byte{8, 8, 8, 8},
	}
	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return nil, err
	}

	err = syscall.Connect(sock, googDns)
	if err != nil {
		return nil, err
	}

	ourname, err := syscall.Getsockname(sock)
	if err != nil {
		return nil, err
	}

	// get addr in [4]byte form
	ipb := (ourname.(*syscall.SockaddrInet4)).Addr
	//ip := make(net.IP, 4)
	ip := net.IPv4(ipb[0], ipb[1], ipb[2], ipb[3])
	return &ip, nil
}
