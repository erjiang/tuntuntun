package main

import (
	"log"
	"net"
)

type Envelope struct {
	env_type byte
	sequence uint64 // big endian?
	packet   []byte
}

const ENV_DATA byte = 1
const ENV_PING byte = 2
const ENV_RECV byte = 3

const IP4_ICMP byte = 1
const IP4_UDP byte = 17
const IP4_TCP byte = 6

const ENVELOPE_LENGTH int = 5 // 5 byte envelope

func replace_sender_ip(pkt []byte, new_ip net.IP) []byte {
	pkt[0] = new_ip[0]
	pkt[1] = new_ip[1]
	pkt[2] = new_ip[2]
	pkt[3] = new_ip[3]
	return pkt
}

func parse_envelope(raw []byte) Envelope {
	return Envelope{
		env_type: raw[0],
		sequence: uint64((raw[0])<<3) + uint64((raw[1])<<2) + uint64((raw[2])<<1) + uint64(raw[3]),
		packet:   raw[ENVELOPE_LENGTH:],
	}
}

// returns 4 for IPv4, 6 for IPv6
func get_ip_version(pkt []byte) byte {
	// get first 4 bits
	return (pkt[0] & 0xF0) / 0x10
}

func get_ip_src(pkt []byte) net.IP {
	switch get_ip_version(pkt) {
	case 4:
		return net.IP(pkt[12:16])
	default:
		log.Printf("IPv%d packets not supported", get_ip_version(pkt))
	}
	return net.IP{0, 0, 0, 0}
}

func get_ip_dest(pkt []byte) net.IP {
	switch get_ip_version(pkt) {
	case 4:
		return net.IP(pkt[16:20])
	default:
		log.Printf("IPv%d packets not supported", get_ip_version(pkt))
	}
	return net.IP{0, 0, 0, 0}
}

/*
func replace_sender_port(pkt []byte) {

}
*/

func get_ip_proto(pkt []byte) byte {
	return pkt[9]
}
