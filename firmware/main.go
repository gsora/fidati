package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/gsora/fidati/leds"

	"github.com/f-secure-foundry/tamago/soc/imx6"
)

var (
	// Build is a string which contains build user, host and date.
	Build string

	// Revision contains the git revision (last hash and/or tag).
	Revision string

	banner string
)

func init() {
	banner = fmt.Sprintf("%s/%s (%s) • %s %s",
		runtime.GOOS, runtime.GOARCH, runtime.Version(),
		Revision, Build)

	log.SetFlags(log.Lshortfile)
	enableLogs()

	model := imx6.Model()
	_, family, revMajor, revMinor := imx6.SiliconVersion()

	if !imx6.Native {
		log.Fatal("running fidati on emulated hardware is not supported")
	}

	if err := imx6.SetARMFreq(900); err != nil {
		log.Printf("WARNING: error setting ARM frequency: %v", err)
	}

	banner += fmt.Sprintf(" • %s %d MHz", model, imx6.ARMFreq()/1000000)

	log.Printf("imx6_soc: %s (%#x, %d.%d) @ %d MHz - native:%v",
		model, family, revMajor, revMinor, imx6.ARMFreq()/1000000, imx6.Native)

	leds.StartBlink()
}

func main() {
	log.Println(banner)

	go rebootWatcher()

	startUSB()
}

func rebootWatcher() {
	buf := make([]byte, 1)

	for {
		runtime.Gosched()
		imx6.UART2.Read(buf)
		if buf[0] == 0 {
			continue
		}

		if buf[0] == 'r' {
			log.Println("rebooting...")
			imx6.Reboot()
		}

		buf[0] = 0
	}
}
