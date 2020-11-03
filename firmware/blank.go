package main

import (
	"fmt"
	"log"
)

// blank zeroes the first 8 bytes after writeOffset.
// Only the first blank() call will actually blank out the SD, subsequent (and future) calls
// will return nil.
func blank() error {
	blanked, err := wasBlanked()
	if err != nil {
		return fmt.Errorf("cannot read blank flag")
	}

	log.Println("was blanked before:", blanked)

	if blanked {
		return nil
	}

	log.Println("writing blanked tag")
	err = sdWrite(blankedTagLBA, []byte{1})
	if err != nil {
		return fmt.Errorf("cannot write blank tag, %w", err)
	}

	log.Println("writing empty data len")
	return sdWrite(storageSizeLBA, make([]byte, 8))
}

func wasBlanked() (bool, error) {
	tag, err := sdRead(blankedTagLBA, blankedTagBlockAmount)
	if err != nil {
		return false, err
	}

	if tag[0] == 1 {
		return true, nil
	}

	return false, nil
}
