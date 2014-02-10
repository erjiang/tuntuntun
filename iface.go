package main

import (
	"errors"
	"fmt"
	"net"
	"socks"
	"strings"
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
	FD           int
	Status       IfaceStatus
	LastRTT      time.Duration
	monitor      chan IfaceStatus
	packets_sent uint64
	packets_recv uint64 // not currently working
	bytes_sent   uint64
	bytes_recv   uint64 // not currently working
	//Conn         *net.UDPConn
}

func (iface *Iface) WriteToUDP(msg []byte, remote_addr *net.UDPAddr) (int, error) {
	return socks.WriteToUDP(iface.FD, msg, remote_addr)
}

func (iface *Iface) ReadFromUDP(buf []byte) (int, *net.UDPAddr, error) {
	return socks.ReadFromUDP(iface.FD, buf)
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

func getIfaceAddr(ifname string) (*net.IP, error) {
	interf, err := net.InterfaceByName(ifname)
	if err != nil {
		return nil, err
	}
	addrs, err := interf.Addrs()
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, errors.New(fmt.Sprintf("Could not get IP for interface %s", ifname))
	}
	debugf(2, "For %s, got addrs %+v", ifname, addrs)

	// strip out the part after the slash in the addr
	// for example: 192.168.0.1/24
	ipstr := addrs[0].String()
	slashindex := strings.Index(ipstr, "/")

	ip := net.ParseIP(ipstr[:slashindex])
	return &ip, nil
}

func upIfaces(ifs []*Iface) int {
	up := 0
	for _, v := range ifs {
		if v.Status == IFACE_STATUS_UP {
			up++
		}
	}
	return up
}
