package main

import (
	"log"
	"net"
	"syscall"
	"tun"
)

var read_buf []byte = make([]byte, BUF_SIZE)

func server() {

	ext_ip, err := get_ext_addr()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("External IP is ", ext_ip.String())

	log.Print("Initializing tun device")
	tun, err := tun.Tun_alloc(tun.IFF_TUN | tun.IFF_NO_PI)
	if err != nil {
		log.Print("Could not allocate tun device")
		log.Fatal(err)
	}
	log.Print("Opened up tun device " + tun.Name())

	log.Print("Listening on 0.0.0.0:", TUNTUNTUN_SERVER_PORT)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: TUNTUNTUN_SERVER_PORT})
	if err != nil {
		log.Fatal(err)
	}

	for {
		count, remote_addr, err := conn.ReadFromUDP(read_buf)
		if err != nil {
			log.Fatal(err)
		}

		pkt := read_buf[ENVELOPE_LENGTH:count]

		log.Printf("Got %d bytes from %s addressed to %s",
			count, remote_addr,
			get_ip_dest(pkt))

		log.Printf("Sending through tun device")
		replace_src_addr(pkt, *ext_ip)
		clear_checksum(pkt)
		tun.Write(read_buf[ENVELOPE_LENGTH:count])
	}
}

// Tries to figure out what IP address is the external address
// by looking up Google's DNS service (8.8.8.8) and seeing which
// interface gets used for it
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
