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

const TTT_DATA byte = 1
const TTT_RESEND_REQ byte = 2
const TTT_ACK byte = 3

const TTT_REGISTER byte = 4
const TTT_REGISTER_REQ byte = 5
const TTT_REGISTER_ACK byte = 6

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
		// TODO: replace with encoding/binary
		sequence: uint64((raw[0])<<24) + uint64((raw[1])<<16) + uint64((raw[2])<<8) + uint64(raw[3]),
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

func replace_src_addr(pkt []byte, new_ip net.IP) {
	switch get_ip_version(pkt) {
	case 4:
		copy(pkt[12:], new_ip.To4()[0:4])
	default:
		log.Printf("IPv%d packets not supported", get_ip_version(pkt))
	}
}

// zero out the IP header checksum
func clear_checksum(pkt []byte) {
	switch get_ip_version(pkt) {
	case 4:
		pkt[10] = 0
		pkt[11] = 0
	default:
		log.Printf("IPv%d packets not supported", get_ip_version(pkt))
	}
}

// replace a packet's checksum with a newly calculated sum
func ReplaceIPHeaderChecksum(pkt []byte) {
	checksum := calculateIPHeaderChecksum(pkt)
	pkt[10] = byte(checksum >> 8)
	pkt[11] = byte(checksum)
}

func calculateIPHeaderChecksum(pkt []byte) uint16 {
	// loop through the 10 16-bit values
	var sum uint32 = 0
	for i := 0; i < 10; i++ {
		if i == 5 {
			continue
		}
		// sum up each 16-bit value
		sum = sum + (uint32(pkt[i*2]) << 8) + uint32(pkt[i*2+1])
	}

	// carry over first bits beyond lower 16
	var checksum uint16 = uint16(sum) + uint16(sum>>16)
	checksum = ^checksum
	return checksum
}

func get_ip_proto(pkt []byte) byte {
	return pkt[9]
}
