// +build usbarmory

package leds

import (
	"context"
	"runtime"
	"time"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
)

// cancelFuncs holds references to the functions needed to stop
// the LED blink.
var cancelFuncs []context.CancelFunc

// StartBlink starts the white and blue LED blinking.
func StartBlink() {
	whiteCtx, whiteCancel := context.WithCancel(context.Background())
	blueCtx, blueCancel := context.WithCancel(context.Background())

	go func() { // start led blinking in a goroutine, don't block the main thread
		go blink(whiteCtx, "white")

		time.Sleep(1 * time.Second)
		runtime.Gosched()

		go blink(blueCtx, "blue")
	}()

	cancelFuncs = []context.CancelFunc{
		whiteCancel,
		blueCancel,
	}
}

// StopBlink stops the blinking, and turns both LEDs off.
func StopBlink() {
	for _, cf := range cancelFuncs {
		cf()
	}

	cancelFuncs = nil
}

// Panic stops blinking, turns both LEDs on.
func Panic() {
	StopBlink()
	usbarmory.LED("blue", true)
	usbarmory.LED("white", true)
}

func blink(ctx context.Context, led string) {
	lastVal := true
	for {
		select {
		case <-ctx.Done():
			usbarmory.LED(led, false)
			return
		default:
			runtime.Gosched()
			usbarmory.LED(led, lastVal)
			time.Sleep(1 * time.Second)
			lastVal = !lastVal
		}
	}
}
