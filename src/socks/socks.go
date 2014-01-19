package socks

// #include "socks.h"
// #include <stdlib.h>
import "C"

import (
    "errors"
    "fmt"
    "net"
    "unsafe"
)

func CreateDeviceBoundUDPSocket(addr net.IP, port uint16, device string) (fd int, err error) {
    addr4 := addr.To4();
    var intaddr uint32
    // convert []byte to uint32
    for i := 0; i < 4; i++ {
        intaddr = intaddr << 4;
        intaddr = intaddr + uint32(addr4[i])
    }
    devcstr := C.CString(device)
    defer C.free(unsafe.Pointer(devcstr))
    fd = int(C.createDeviceBoundUDPSocket(C.uint32_t(intaddr), C.uint16_t(port), devcstr))
    if fd < 0 {
        return fd, errors.New(fmt.Sprintf("Got return code %d when creating bound socket.", fd))
    }

    return
}
