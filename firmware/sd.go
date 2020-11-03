package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"

	usbarmory "github.com/f-secure-foundry/tamago/board/f-secure/usbarmory/mark-two"
)

const (
	writeOffset = 1

	blankedTagLBA         = 1
	blankedTagBlockAmount = 1

	storageSizeLBA         = 2
	storageSizeBlockAmount = 1

	storageLBA = 3
)

var noData = errors.New("no data")

// wrongOffset panics if writeOffset is greater than the maximum number of blocks available on the SD.
func wrongOffset(s *sd) {
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

func readStorageData(s *sd) ([]byte, error) {
	wrongOffset(s)
	return readStorageBytes()
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
	ret := make([]byte, numBlocks*usbarmory.SD.Info().BlockSize)

	return ret, usbarmory.SD.ReadBlocks(offset, numBlocks, ret)
}

func readStorageBytes() ([]byte, error) {
	// read the first 8 bytes, excluding the "blank" tag.
	sz, err := sdRead(storageSizeLBA, storageSizeBlockAmount)
	if err != nil {
		return nil, err
	}

	log.Println("storage size bytes read from sd", sz)
	totalLen := int(binary.LittleEndian.Uint64(sz))

	log.Println("total length", totalLen)

	if totalLen == 0 {
		return nil, nil
	}

	blockAmount := closestSectorNumber(totalLen) / 512

	log.Println("block amount", blockAmount)
	ret := make([]byte, closestSectorNumber(totalLen))

	err = usbarmory.SD.ReadBlocks(storageLBA, blockAmount, ret)
	if err != nil {
		return nil, err
	}

	return ret[:totalLen], nil
}

// sd implements a io.Reader for the microSD card slot on the USB Armory Mk.II.
type sd struct{}

// Write implements io.ReadWriter.
func (s *sd) Write(p []byte) (int, error) {
	log.Println("length of stuff to write", len(p))
	psz := make([]byte, 8)
	binary.LittleEndian.PutUint64(psz, uint64(len(p)))

	log.Println("storage size lba", storageSizeLBA)
	err := sdWrite(storageSizeLBA, psz)
	if err != nil {
		return 0, fmt.Errorf("cannot write storage size lba, %w", err)
	}

	err = sdWrite(storageLBA, p)
	if err != nil {
		return 0, fmt.Errorf("cannot write storage, %w", err)
	}

	log.Println("write from sd.Write() finished")

	return len(p), nil
}
