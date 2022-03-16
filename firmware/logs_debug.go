// +build debug

package main

import (
	"log"
	"os"

	usbarmory "github.com/usbarmory/tamago/board/f-secure/usbarmory/mark-two"
)

func enableLogs() {
	usbarmory.EnableDebugAccessory()
	log.SetOutput(os.Stdout)
	log.Println("enabled debugging logs")
}
