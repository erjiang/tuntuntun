package main

import (
	"errors"
	"fmt"
	"github.com/mgutz/ansi"
	"log"
	"net"
	"os"
	"socks"
	"time"
	"tun"
)

var packet_seq uint64 = 0

var send_buf []byte = make([]byte, BUF_SIZE)

var ifs []*Iface

func client(remote_addr *net.UDPAddr, local_ifs []string) {
	debug(2, "Initializing tun device")
	tundev, err := tun.Tun_alloc(tun.IFF_TUN | tun.IFF_NO_PI)
	if err != nil {
		log.Print("Could not allocate tun device")
		log.Fatal(err)
	}
	err = tun.SetMTU(tundev.Name(), 1500-TUNNEL_OVERHEAD)
	if err != nil {
		log.Print("Could not set MTU: ", err)
	}
	debug(0, "Opened up tun device "+tundev.Name())

	debug(0, "Initializing interfaces...")
	// create list of local Ifs and store in global
	ifs = setupIfs(local_ifs)

	debug(0, "Sending handshake to server...")
	err = registerBegin(ifs[0], remote_addr)
	if err != nil {
		log.Fatal(err)
	}
	for _, iface := range ifs {
		debugf(1, "Registering %s with server...", iface.IP.IP)
		err = registerClient(iface, remote_addr)
		if err != nil {
			log.Fatal(err)
		}
	}

	debug(1, "Configuring device with ifconfig")
	err = tun.Ifconfig(tundev.Name(), TTT_CLIENT_IP, TTT_SERVER_IP)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Ready ...")

	// set up listening channels for udp and tun
	tunchan := make(chan []byte)
	udpchan := make(chan UDPRecv)
	fwdchan := make(chan []byte)

	go listenTun(tundev, tunchan)
	for _, iface := range ifs {
		go listenUDP(iface, udpchan)
	}
	// put packet forwarding in a separate goroutine to be able to do
	// round-robin load-balancing and more
	go forwardPacketHandler(remote_addr, fwdchan)

	for {
		select {
		case tun_pkt, ok := <-tunchan:
			if !ok {
				log.Fatal("Error reading from tun")
			}
			debugf(3, "Got a packet of %d bytes for %s", len(tun_pkt),
				get_ip_dest(tun_pkt))
			debugf(3, "Sending to "+remote_addr.String())
			// pass along packet
			// TODO: use another buffer to prevent clobbering things
			fwdchan <- tun_pkt
		case udpr, ok := <-udpchan:
			if !ok {
				log.Fatal("Error reading from udp")
			}
			envelope := udpr.Data
			remote_addr := udpr.RemoteAddr
			debugf(3, "Got packet of len %d from %s", len(envelope), remote_addr)
			switch envelope[0] {
			case TTT_DATA: // packet to be forwarded
				pkt := envelope[ENVELOPE_LENGTH:]
				// pass along packet
				tundev.Write(pkt)
			default:
				debug(1, "Received unhandled packet of type ", envelope[0])
			}
		}
	}
}

// gets a list of network interfaces (eth0, eth1, wlan0, ...)
// and does the following:
// 1. gets the (first) IP associated with that interface
// 2. creates a unix socket bound to that interface
// 3. creates a golang udp listener to listen on that address
// 4. creates a tuntuntun Iface struct for that interface
func setupIfs(ifs []string) []*Iface {
	var iflist = make([]*Iface, 0)
	for i, v := range ifs {
		// Figure out interface's IP
		log.Printf("Getting ip of %s", v)
		ip, err := getIfaceAddr(v)
		if err != nil {
			log.Fatal(err)
		}
		debugf(1, "IP of %s is %s", v, ip)

		debug(1, "Disabling reverse path filtering on ", v)
		disableRPFilter(v)

		// create socket bound to this if
		// put every socket on a different port to avoid conflicts
		fd, err := socks.CreateDeviceBoundUDPSocket(ip, uint16(TUNTUNTUN_CLIENT_PORT+i), v)
		if err != nil {
			log.Fatal(err)
		}

		// try listening on this IP
		/* deprecated, listening on unix sockets now
		udpaddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, TUNTUNTUN_CLIENT_PORT))
		if err != nil {
			log.Fatal(err)
		}
		debug(1, "Listening on ", udpaddr)
		conn, err := net.ListenUDP("udp", udpaddr)
		if err != nil {
			log.Print("Could not listen to ", udpaddr)
			log.Fatal(err)
		}
		*/

		// create if struct and add to list
		iflist = append(iflist, &Iface{
			Name: v,
			FD:   fd,
			IP: &net.UDPAddr{
				IP:   *ip,
				Port: TUNTUNTUN_CLIENT_PORT,
			},
			Status: IFACE_STATUS_UP,
			//Conn:   conn,
		})
		debug(1, fmt.Sprintf("Created link %s", v))
	}
	return iflist
}

func registerBegin(iface *Iface, remote_addr *net.UDPAddr) error {
	registration := []byte{TTT_REGISTER_BEGIN}
	debug(1, "Beginning registration...")
	_, err := iface.WriteToUDP(registration, remote_addr)
	if err != nil {
		log.Fatal("Could not send registration: ", err)
	}

	ack_chan := make(chan bool) // alert us if we get an ack from the server

	// keep reading until we get a REGISTER_BEGIN ack from server
	go func() {
		read_buf := make([]byte, 100) // buffer to receive ack
		for {
			count, _, err := iface.ReadFromUDP(read_buf)
			if err != nil {
				log.Fatal(err)
			}
			if count > 0 && read_buf[0] == TTT_REGISTER_BEGIN {
				ack_chan <- true
				break
			}
		}
	}()

	// retry registration every 2 seconds
WAIT:
	for {
		select {
		case did_ack := <-ack_chan:
			if did_ack {
				break WAIT
			}
		case <-time.After(2 * time.Second):
			debug(0, "Resending registration...")
			_, err = iface.WriteToUDP(registration, remote_addr)
			if err != nil {
				log.Fatal("Could not send registration: ", err)
			}
		}
	}
	// TODO: wait for registration acknownledgment
	return err
}

func registerClient(iface *Iface, remote_addr *net.UDPAddr) error {
	registration := []byte{TTT_REGISTER}
	log.Printf("Registering via fd %d", iface.FD)
	b, err := iface.WriteToUDP(registration, remote_addr)
	log.Printf("Sent out %d bytes for registration", b)
	time.Sleep(400 * time.Millisecond)
	// TODO: wait for registration acknownledgment
	return err
}

func forward_packet(writer UDPReadWrite, remote_addr *net.UDPAddr, pkt []byte) error {

	total_len := len(pkt) + ENVELOPE_LENGTH

	if uint(total_len) > BUF_SIZE {
		return errors.New(fmt.Sprintf("%d packet too long for %d", len(pkt), BUF_SIZE))
	}

	send_buf[0] = ENV_DATA
	send_buf[1] = byte(packet_seq >> 24)
	send_buf[2] = byte(packet_seq >> 16)
	send_buf[3] = byte(packet_seq >> 8)
	send_buf[4] = byte(packet_seq >> 0)
	packet_seq++

	copy(send_buf[5:], pkt)

	_, err := writer.WriteToUDP(send_buf[:total_len], remote_addr)

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

		//fmt.Printf("sending out conn %p", iface.IP)
		err := forward_packet(iface, remote_addr, pkt)
		if err != nil {
			log.Print(err)
		}

		// log statistics in if
		iface.packets_sent++
		// does not count envelope in bytes sent
		iface.bytes_sent = iface.bytes_sent + uint64(len(pkt))

		// make colorful display of packets
		// each iface gets a different color, and we print 'S' for sent and 'R' for received
		if DEBUG_LEVEL >= 1 {
			fmt.Print(ansi_colors[int(packet_seq%uint64(len(ifs)))%len(colors)], "S", ansi_reset)
		}

		if err != nil {
			log.Print(err)
		}
	}
}

// disables reverse-path filtering for the given device (or "all")
func disableRPFilter(device string) error {
	f, err := os.Open("/proc/sys/net/ipv4/conf/" + device + "/rp_filter")
	if err != nil {
		return err
	}

	_, err = f.Write([]byte{'0'})
	if err != nil {
		return err
	}

	return f.Close()
}
