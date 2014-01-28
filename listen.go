package main

import (
	"fmt"
	"github.com/mgutz/ansi"
	"log"
	"net"
	"os"
)

type UDPRecv struct {
	Data       []byte // UDP payload
	RemoteAddr *net.UDPAddr
}

var listenerID int = 0

// listenUDP goroutine has its own internal read buf that it gives a slice into
// when it receives a packet. It passes the slice and the source address into
// the channel.
// TODO: make this take an Iface so it can log statistics
func listenUDP(conn UDPReadWrite, c chan UDPRecv) error {
	myID := listenerID // for debugging output
	listenerID++

	colors := []string{"magenta", "yellow", "cyan", "white:blue", "black:white"}
	mycolor := ansi.ColorCode(colors[myID%len(colors)])
	ansi_reset := ansi.ColorCode("reset")

	debugf(1, "%sListening to %p...%s", mycolor, conn, ansi_reset)

	// switch between two buffers so we don't clobber something that the end
	// end of the pipe is busy processing
	read_buf := make([]byte, BUF_SIZE)
	read_buf2 := make([]byte, BUF_SIZE)
	curr_buf := &read_buf
	other_buf := &read_buf2
	for {
		count, remote_addr, err := conn.ReadFromUDP(*curr_buf)
		if err != nil {
			log.Print(err)
			return err
		}
		debugf(3, "%sGot %d bytes from %s%s", mycolor, count, remote_addr, ansi_reset)
		// some debugging by printing out packet numbers
		if count > 5 {
			e := parse_envelope(read_buf)
			debugf(4, "rfUDP: %d", e.sequence)
		}
		c <- UDPRecv{
			Data:       (*curr_buf)[:count],
			RemoteAddr: remote_addr,
		}
		if DEBUG_LEVEL >= 1 {
			fmt.Print(mycolor, "R", ansi_reset)
		}
		// swap the buffers
		temp_swp := curr_buf
		curr_buf = other_buf
		other_buf = temp_swp
	}
}

func listenTun(tundev *os.File, c chan []byte) error {
	read_buf := make([]byte, BUF_SIZE)
	read_buf2 := make([]byte, BUF_SIZE)
	curr_buf := &read_buf
	other_buf := &read_buf2
	for {
		count, err := tundev.Read(*curr_buf)
		if err != nil {
			log.Print(err)
			return err
		}
		c <- (*curr_buf)[:count]
		temp_swp := curr_buf
		curr_buf = other_buf
		other_buf = temp_swp
	}
}
