// +build fidati_logs

package flog

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "fidati :: ", log.Lshortfile)
