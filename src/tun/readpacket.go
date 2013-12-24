package tun

// #include "readpacket.h"
// #include <stdlib.h>
import "C"

import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

const IFF_TUN int = 0x0001
const IFF_TAP int = 0x0002
const IFF_NO_PI int = 0x1000

func Tun_alloc(flags int) (*os.File, error) {
	devname := C.alloc_tun_name()
	fd := C.tun_alloc(devname, C.int(flags))
	if fd < 0 {
		return nil, errors.New(fmt.Sprintf(
			"Error trying to create tun device. Got error code %d. Do you have permission to create a tun device?", int(fd)))
	}
	godevname := C.GoString(devname)
	C.free(unsafe.Pointer(devname))
	f := os.NewFile(uintptr(fd), godevname)
	return f, nil
}
