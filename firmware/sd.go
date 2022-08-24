package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
)

const (
	writeOffset = 1

	counterLBA         = 1
	counterBlockAmount = 1
)

var noData = errors.New("no data")

// wrongOffset panics if writeOffset is greater than the maximum number of blocks available on the SD.
func wrongOffset() {
	info := usbarmory.SD.Info()
	if info.Blocks < writeOffset {
		panic(fmt.Sprintf("specified base write offset %d bigger than SD total block number %d", writeOffset, info.Blocks))
	}
}

// closestSectorNumber returns the closest number of sectors divisible by the microSD block size,
// by rounding up.
func closestSectorNumber(n int) int {
	m := float64(usbarmory.SD.Info().BlockSize)
	return int(math.Ceil(float64(n)/m) * m)
}

func sdWrite(offset int, data []byte) error {
	var nd []byte

	if len(data) < usbarmory.SD.Info().BlockSize {
		nd = make([]byte, usbarmory.SD.Info().BlockSize)
		copy(nd, data)
	} else if len(data) == usbarmory.SD.Info().BlockSize {
		nd = data
	} else {
		nd = make([]byte, closestSectorNumber(len(data)))
		copy(nd, data)
	}

	log.Printf("writing %v at lba %v", nd, offset)
	err := usbarmory.SD.WriteBlocks(offset, nd)
	log.Println("finished writing")
	return err
}

func sdRead(offset, numBlocks int) ([]byte, error) {
	ret := make([]byte, numBlocks)

	return ret, usbarmory.SD.ReadBlocks(offset, ret)
}

type sdCounter struct {
	counter uint32
}

func (s *sdCounter) UserPresence() bool {
	// we always say yes :)]
	return true
}

func readSdCounter() (*sdCounter, error) {
	cbytes, err := sdRead(counterLBA, counterBlockAmount)
	if err != nil {
		return nil, err
	}

	return &sdCounter{
		counter: binary.LittleEndian.Uint32(cbytes),
	}, nil
}

func (s *sdCounter) Increment(_ []byte, _ []byte, _ []byte) (uint32, error) {
	s.counter++
	cbytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(cbytes, s.counter)
	err := sdWrite(counterLBA, cbytes)
	if err != nil {
		return 0, err
	}

	return s.counter, nil
}
