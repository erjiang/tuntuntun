package readpacket

// #include <readpacket.h>
// #include <stdlib.h>
import "C"

import (
	"errors"
	"fmt"
    "os"
)

func tun_alloc(flags int32) (os.File, error) {
	devname := C.alloc_tun_name()
	fd := C.tun_alloc(devname, flags)
	if fd < 0 {
		return 0, "", errors.new(fmt.sprintf(
			"Error trying to create tun device. Got error code %ld. Do you have permission to create a tun device?", fd))
	}
	godevname := C.GoString(devname)
	C.free(devname)
    f := os.NewFile(fd, godevname)
	return f, nil
}
