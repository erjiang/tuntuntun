package socks

// #include "socks.h"
// #include <stdlib.h>
// #include <arpa/inet.h>
import "C"

import (
	"errors"
	"fmt"
	"net"
	"unsafe"
)

func CreateDeviceBoundUDPSocket(addr *net.IP, port uint16, device string) (fd int, err error) {
	devcstr := C.CString(device)
    //fmt.Printf("creating device for %s", device)
	defer C.free(unsafe.Pointer(devcstr))
	fd = int(C.createDeviceBoundUDPSocket(C.uint32_t(ipToInt(addr)), C.uint16_t(port), devcstr))
	if fd < 0 {
		return fd, errors.New(fmt.Sprintf("Got return code %d when creating bound socket.", fd))
	}

	return
}

// not thread safe
func WriteToUDP(fd int, buf []byte, dest *net.UDPAddr) (int, error) {
    //fmt.Printf("WriteToUDP (go): fd=%d, buf len=%d, dest=%s\n", fd, len(buf), dest);
	res, errno := C.writeToUDP(C.int(fd), unsafe.Pointer(&buf[0]),
		C.size_t(len(buf)), C.uint32_t(ipToInt(&dest.IP)),
		C.uint16_t(dest.Port))
	if res < 0 {
		return int(res), errors.New(fmt.Sprintf("Got errno %d from sendto", errno))
	}
	return int(res), nil
}

func ReadFromUDP(fd int, buf []byte) (int, *net.UDPAddr, error) {
    var port_buf C.uint16_t
    var ip_buf C.uint32_t
    /*
    fmt.Printf("Calling recvFromUDP(%d, %p, %d, %p, %p)\n", C.int(fd), unsafe.Pointer(&buf[0]),
        C.size_t(len(buf)), &ip_buf, &port_buf)
    */
    res, errno := C.recvFromUDP(C.int(fd), unsafe.Pointer(&buf[0]),
        C.size_t(len(buf)), &ip_buf, &port_buf)

    if res < 0 {
        return int(res), nil, errors.New(fmt.Sprintf("Got errno %d from recvfrom", errno))
    }

    addr := &net.UDPAddr{
        IP: []byte{
            byte(ip_buf >> 24),
            byte(ip_buf >> 16),
            byte(ip_buf >> 8),
            byte(ip_buf >> 0),
        },
        Port: int(port_buf),
    }
    return int(res), addr, nil
}

// helper function for converting ipv4 addr to uint32_t
// doesn't work for ipv6
func ipToInt(ip *net.IP) C.uint32_t {
	addr4 := ip.To4()
	var intaddr uint32
	for i := 0; i < 4; i++ {
		intaddr = intaddr << 8
		intaddr = intaddr + uint32(addr4[i])
	}
    //fmt.Printf("Conv %s->%x\n", ip, intaddr)
	return C.uint32_t(intaddr)
}
