package keyring

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
)

var nonceFunc = nonce
var keygenFunc = generateECKey

// Counter is some sort of interface to a counter (like, a monotonic counter) and to a
// user presence confirmation device.
type Counter interface {
	Increment(appID, challenge, keyHandle []byte) (uint32, error)
	UserPresence() bool
}

// Keyring represents a mechanism to derive deterministic relying party authentication private keys
// given a master key.
// A Keyring needs a Counter to be able to pass along the counter value recommended by the FIDO U2F standard.
// Keyring implements the key wrapping method described by Yubico: https://www.yubico.com/blog/yubicos-u2f-key-wrapping/.
type Keyring struct {
	Counter   Counter
	MasterKey []byte
}

func (k *Keyring) validate() error {
	switch {
	case k.MasterKey == nil:
		return errors.New("master key is nil")
	case k.Counter == nil:
		return errors.New("counter is nil")
	default:
		return nil
	}
}

// New returns a Keyring pointer given a master key and a Counter.
func New(mk []byte, counter Counter) *Keyring {
	return &Keyring{
		MasterKey: mk,
		Counter:   counter,
	}
}

// NonceFromKeyHandle returns the nonce from a given keyhandle.
// Assumes SHA-256 as hashing function.
func (k *Keyring) NonceFromKeyHandle(kh []byte) []byte {
	if len(kh) < sha256.Size {
		return nil
	}

	return kh[sha256.Size:]
}

// Register deterministically derives an ECDSA public key given an application ID.
// It also returns a key handle (also deterministic) and an error.
// If nonce is not nil, it will be used for the derivation process.
func (k *Keyring) Register(appID []byte, nonce []byte) (*ecdsa.PublicKey, []byte, error) {
	if err := k.validate(); err != nil {
		return nil, nil, err
	}

	if appID == nil {
		return nil, nil, errors.New("appID is nil")
	}

	if nonce == nil {
		var err error
		nonce, err = nonceFunc()
		if err != nil {
			return nil, nil, err
		}
	}

	mac := hmac.New(sha256.New, k.MasterKey)
	_, err := mac.Write(appID)
	if err != nil {
		return nil, nil, err
	}

	_, err = mac.Write(nonce)
	if err != nil {
		return nil, nil, err
	}

	rpPrivKey := mac.Sum(nil)

	mac.Reset()

	_, err = mac.Write(appID)
	if err != nil {
		return nil, nil, err
	}

	_, err = mac.Write(rpPrivKey)
	if err != nil {
		return nil, nil, err
	}

	keyHandle := append(mac.Sum(nil), nonce...)

	ecPrivKey, err := keygenFunc(rpPrivKey)
	if err != nil {
		return nil, nil, err
	}

	return &ecPrivKey.PublicKey, keyHandle, nil
}

// retrievePrivkey returns the private key associated to a given application ID and key handle.
func retrievePrivkey(appID, keyHandle, masterKey []byte) (*ecdsa.PrivateKey, error) {
	if len(keyHandle) < 32 {
		return nil, errors.New("key handle is shorter than 32 bytes")
	}

	nonce := keyHandle[32:]

	mac := hmac.New(sha256.New, masterKey)
	_, err := mac.Write(appID)
	if err != nil {
		return nil, err
	}

	_, err = mac.Write(nonce)
	if err != nil {
		return nil, err
	}

	rpPrivKey := mac.Sum(nil)

	ecPrivKey, err := keygenFunc(rpPrivKey)
	if err != nil {
		return nil, err
	}

	return ecPrivKey, nil
}

// Authenticate returns a valid FIDO2 U2F authentication signature for the given application ID,
// authentication challenge, key handle and a byte indicating whether user presence was confirmed or not.
// It also returns the updated count to be used in the authentication message, and an error.
func (k *Keyring) Authenticate(appID, challenge, keyHandle []byte, userPresence bool) ([]byte, uint32, error) {
	if err := k.validate(); err != nil {
		return nil, 0, err
	}

	switch {
	case appID == nil:
		return nil, 0, fmt.Errorf("appID is nil")
	case challenge == nil:
		return nil, 0, fmt.Errorf("challenge is nil")
	case keyHandle == nil:
		return nil, 0, fmt.Errorf("keyHandle is nil")
	}

	privKey, err := retrievePrivkey(appID, keyHandle, k.MasterKey)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot derive private key from appID and keyHandle, %w", err)
	}

	userPresenceByte := byte(0)
	if userPresence {
		userPresenceByte = 1
	}

	count, err := k.Counter.Increment(appID, challenge, keyHandle)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot increment counter, %w", err)
	}

	sp := signaturePayload(
		appID,
		count,
		challenge,
		userPresenceByte,
	)

	sph := sha256.Sum256(sp)
	spHash := sph[:]

	sign, err := ecdsa.SignASN1(rand.Reader, privKey, spHash)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot execute authentication signature, %w", err)
	}

	return sign, count, nil
}

// signaturePayload returns the byte slice to be signed to validate an authentication request.
func signaturePayload(appParam []byte, counter uint32, challengeParam []byte, userPresenceByte byte) []byte {
	ret := new(bytes.Buffer)

	ret.Write(appParam)
	ret.WriteByte(userPresenceByte)
	counterBytes := [4]byte{}

	binary.BigEndian.PutUint32(counterBytes[:], counter)

	ret.Write(counterBytes[:])
	ret.Write(challengeParam)

	return ret.Bytes()
}

// nonce returns a byte slice with 32 bytes of randomness inside.
func nonce() ([]byte, error) {
	n := make([]byte, 32)
	_, err := rand.Read(n)
	return n, err
}

// generateECKey generates a ECDSA private key given b bytes.
// This function is deterministic, and will return always the same *ecdsa.PrivateKey given
// the same b bytes.
func generateECKey(b []byte) (*ecdsa.PrivateKey, error) {
	// code adapted from https://golang.org/src/crypto/ecdsa/ecdsa.go?#L133
	// we had to hand-roll this because ecdsa.GenerateKey() expects 40 bytes of random data
	// instead of just 32.
	// We're basically using b as our privkey bytes representation.
	c := elliptic.P256()
	params := c.Params()
	var one = new(big.Int).SetInt64(1)
	k := new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(params.N, one)
	k.Mod(k, n)
	k.Add(k, one)

	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = c
	priv.D = k
	priv.PublicKey.X, priv.PublicKey.Y = c.ScalarBaseMult(k.Bytes())
	return priv, nil
}
