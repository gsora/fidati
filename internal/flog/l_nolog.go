// +build !fidati_logs

package flog

import (
	"io/ioutil"
	"log"
)

var Logger = log.New(ioutil.Discard, "fidati :: ", log.Lshortfile)
