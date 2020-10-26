// +build usbarmory

package main

import (
	"log"
	"runtime"
	"time"

	usbarmory "github.com/f-secure-foundry/tamago/board/f-secure/usbarmory/mark-two"
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

func init() {
	if imx6.Native && (imx6.Family == imx6.IMX6UL || imx6.Family == imx6.IMX6ULL) {
		usbarmory.EnableDebugAccessory()
		log.Println("enabling debug accessory (if needed)")
		blinkLeds()
	}
}

func blinkLeds() {
	go func() {
		go func() {
			lastVal := true
			for {
				runtime.Gosched()
				usbarmory.LED("white", lastVal)
				time.Sleep(1 * time.Second)
				lastVal = !lastVal
			}
		}()

		time.Sleep(1 * time.Second)
		runtime.Gosched()

		go func() {
			lastVal := true
			for {
				runtime.Gosched()
				usbarmory.LED("blue", lastVal)
				time.Sleep(1 * time.Second)
				lastVal = !lastVal
			}
		}()
	}()
}
