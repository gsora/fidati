// +build !debug

package main

import (
	"io/ioutil"
	"log"
)

func enableLogs() {
	log.SetOutput(ioutil.Discard)
}
