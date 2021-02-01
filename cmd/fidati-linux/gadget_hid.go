package main

// #cgo LDFLAGS: -lusbgx
// #include "gadget-hid.h"
// #include <stdlib.h>
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/gsora/fidati/u2fhid"
)

func configureHidg(configfsPath string) error {
	reportDescC := (*C.char)(unsafe.Pointer(&u2fhid.DefaultReport[0]))

	serial := C.CString("4242424242")
	manufacturer := C.CString("gsora")
	product := C.CString("fidati desktop")
	cfp := C.CString(configfsPath)

	defer func() {
		C.free(unsafe.Pointer(serial))
		C.free(unsafe.Pointer(manufacturer))
		C.free(unsafe.Pointer(product))
		C.free(unsafe.Pointer(cfp))
	}()

	res := C.configure_hidg(
		serial,
		manufacturer,
		product,
		cfp,
		reportDescC,
		C.ulong(len(u2fhid.DefaultReport)),
	)

	if res != C.USBG_SUCCESS {
		rres := C.usbg_error(res)
		errName := C.GoString(C.usbg_error_name(rres))
		stdErr := C.GoString(C.usbg_strerror(rres))

		return fmt.Errorf("libusbgx failure, %s: %s", errName, stdErr)
	}

	return nil
}

func cleanupHidg(configfsPath string) error {
	cfp := C.CString(configfsPath)

	defer func() {
		C.free(unsafe.Pointer(cfp))
	}()

	C.cleanup_usbg(cfp)
	return nil

}
