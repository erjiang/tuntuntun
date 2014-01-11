package main

import (
	"net"
	"time"
)

type IfaceStatus int

// when monitoring a server, ping no faster than once per PING_INTERVAL
const PING_INTERVAL time.Duration = 1 * time.Second

// how to wait before giving up on a ping
const PING_TIMEOUT time.Duration = 5 * time.Second

// ping (RTT) threshold beyond which an iface is considered
// congested or down
const IFACE_RTT_CONGESTED time.Duration = 1 * time.Second
const IFACE_RTT_DOWN time.Duration = 2 * time.Second

const (
	IFACE_STATUS_DOWN      = iota
	IFACE_STATUS_CONGESTED = iota
	IFACE_STATUS_UP        = iota
)

type Iface struct {
	Name         string
	IP           *net.UDPAddr
	Conn         *net.UDPConn
	Status       IfaceStatus
	LastRTT      time.Duration
	monitor      chan IfaceStatus
	packets_sent uint64
	packets_recv uint64 // not currently working
	bytes_sent   uint64
	bytes_recv   uint64 // not currently working
}

// function to repeatedly ping a host to monitor its response time
// requires that acks get forwarded from the connection listener to the monitor
func monitorIface(conn *net.UDPConn, remote_addr *net.UDPAddr, acks chan uint32, times chan time.Duration) {
	var ping_serial uint32 = 0
	var last_ping_time time.Time
	var last_rtt time.Duration
	for {
		// TODO: check for error from writetoudp
		conn.WriteToUDP(pingPacket(ping_serial), remote_addr)
		last_ping_time = time.Now()
		select {
		// TODO: check if channel was closed
		case ack_serial, _ := <-acks:
			if ack_serial == ping_serial {
				last_rtt = (time.Now().Sub(last_ping_time))
				times <- last_rtt
			}
		case <-time.After(PING_TIMEOUT):
			// TODO: think of a better way to signal timeout than -1 sec
			times <- (-1 * time.Second)
		}
		// if the ping came back faster than PING_INTERVAL
		// then wait until it's time to send the next ping
		if PING_INTERVAL-last_rtt > 0 {
			<-time.After(PING_INTERVAL - last_rtt)
		}

		ping_serial++
	}
}

func pingPacket(serial uint32) []byte {
	return []byte{TTT_PING_REQ}
}
