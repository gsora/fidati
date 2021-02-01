//go:generate go run github.com/rakyll/statik -src=../certs -p=certs

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gsora/fidati/keyring"
	"github.com/gsora/fidati/u2fhid"
	"github.com/gsora/fidati/u2ftoken"
	"github.com/rakyll/statik/fs"

	_ "github.com/gsora/fidati/cmd/fidati-linux/certs"
)

var (
	// X.509 attestation certificate, sent along in registration requests
	attestationCertificate []byte

	// ECDSA private key, used to sign registration requests
	attestationPrivkey []byte
)

func cliArgs() (hidg, configfsPath string, mustClean bool) {
	flag.StringVar(&hidg, "hidg", "/dev/hidg0", "/dev/hidgX file descriptor path")
	flag.StringVar(&configfsPath, "configfs-path", "/sys/kernel/config", "configfs path")
	flag.BoolVar(&mustClean, "clean", false, "clean existing hidg descriptors and exit")
	flag.Parse()

	return
}

func main() {
	hidg, configfsPath, mustClean := cliArgs()

	if mustClean {
		if err := cleanupHidg(configfsPath); err != nil {
			panic(err)
		}

		return
	}

	if err := configureHidg(configfsPath); err != nil {
		panic(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	readCertPrivkey()

	hidRx, err := os.OpenFile(hidg, os.O_RDWR, 0666)
	notErr(err)

	log.Println("done, polling...")
	d := &dumbCounter{}
	k := genKeyring(attestationPrivkey, d)

	token, err := u2ftoken.New(k, attestationCertificate, attestationPrivkey)
	notErr(err)

	hid, err := u2fhid.NewHandler(token)
	notErr(err)

	// add 50ms delay in both rx and tx
	// we don't wanna burn laptop cpus :^)

	// rx
	go func() {
		for {
			time.Sleep(50 * time.Millisecond)
			buf := make([]byte, 64)
			_, err := hidRx.Read(buf)
			notErr(err)

			_, err = hid.Rx(buf, nil)
			if err != nil {
				log.Println("rx error:", err)
				continue
			}
		}
	}()

	go func() {
		for {
			time.Sleep(50 * time.Millisecond)
			data, err := hid.Tx(nil, nil)
			if err != nil {
				log.Println("tx error:", err)
				continue
			}

			if data == nil {
				continue
			}

			_, err = hidRx.Write(data)
			notErr(err)
		}
	}()

	log.Println("running...")

	<-sigs

	fmt.Println()
	log.Println("cleaning...")

	if err := cleanupHidg(configfsPath); err != nil {
		panic(err)
	}
}

func genKeyring(secret []byte, counter keyring.Counter) *keyring.Keyring {
	return keyring.New(secret, counter)
}

func readCertPrivkey() {
	statikFS, err := fs.New()
	notErr(err)

	aCert, err := statikFS.Open("/attestation_certificate.pem")
	notErr(err)

	aPk, err := statikFS.Open("/ecdsa_privkey.pem")
	notErr(err)

	aCertBytes, err := ioutil.ReadAll(aCert)
	notErr(err)

	aPkBytes, err := ioutil.ReadAll(aPk)
	notErr(err)

	attestationCertificate = aCertBytes
	attestationPrivkey = aPkBytes
}

// since we're in a critical configuration phase, panic on error.
func notErr(e error) {
	if e != nil {
		panic(e)
	}
}

type dumbCounter struct {
	i uint32
}

func (d *dumbCounter) Increment(appID []byte, challenge []byte, keyHandle []byte) (uint32, error) {
	d.i++
	return d.i, nil
}

func (d *dumbCounter) UserPresence() bool {
	return true
}
