package main

import (
	"fmt"
	"log"
	"runtime"
	"runtime/debug"

	"github.com/gsora/fidati/firmware/leds"

	usbarmory "github.com/usbarmory/tamago/board/f-secure/usbarmory/mark-two"
	"github.com/usbarmory/tamago/soc/imx6"

	_ "github.com/gsora/fidati/firmware/certs"
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

	err := usbarmory.SD.Detect()
	if err != nil {
		panic(err)
	}

	readCertPrivkey()

	leds.StartBlink()
}

func main() {
	defer catchPanic()

	log.Println(banner)

	go rebootWatcher()

	//s := &sd{}

	/*if err := blank(); err != nil {
		panic(err)
	}

	var store *storage.Storage
	data, err := readStorageData(s)
	if err != nil && err != noData {
		panic(err)
	} else if err == noData {
		st, err := storage.New(s, nil)
		if err != nil {
			panic(err)
		}

		store = st
	} else {
		st, err := storage.New(s, data)
		if err != nil {
			panic(err)
		}

		store = st
	}*/

	counter, err := readSdCounter()
	if err != nil {
		panic(err)
	}

	k := genKeyring(attestationPrivkey, counter)
	startUSB(k)
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
			imx6.Reset()
		}

		buf[0] = 0
	}
}

// catchPanic catches every panic(), sets the LEDs into error mode and prints the stacktrace.
func catchPanic() {
	if r := recover(); r != nil {
		leds.Panic()
		fmt.Printf("panic: %v\n\n", r)
		fmt.Println(string(debug.Stack()))

		for {
		} // stuck here forever!
	}
}

// since we're in a critical configuration phase, panic on error.
func notErr(e error) {
	if e != nil {
		panic(e)
	}
}
