package storage

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
)

// Storage is the global KeyStorage instance.
var Storage *KeyStorage

// Device represents a generic interface with a storage component,
// used to persist Storage.
var Device io.ReadWriter = NilDevice{}

// LoadStorage loads a serialized KeyStorage instance from b, and assigns it to
// Storage.
func LoadStorage(b []byte) error {
	rb := bytes.NewReader(b)
	decoder := gob.NewDecoder(rb)
	if decoder == nil {
		return errors.New("gob decoder is nil")
	}

	err := decoder.Decode(Storage)
	if err != nil {
		return fmt.Errorf("error while decoding storage bytes, %w", err)
	}

	return nil
}

// KeyStorage stores registered relying party keys.
type KeyStorage struct {
	M map[[32]byte]KeyItem
}

// New initializes a new keyStorage instance.
func New() *KeyStorage {
	return &KeyStorage{
		M: map[[32]byte]KeyItem{},
	}
}

// Bytes returns the byte slice representation of ks.
func (ks *KeyStorage) Bytes() []byte {
	ret := new(bytes.Buffer)
	enc := gob.NewEncoder(ret)

	gob.Register(elliptic.P256())
	if err := enc.Encode(ks); err != nil {
		panic(fmt.Sprintf("cannot encode KeyStorage: %s", err.Error()))
	}

	return ret.Bytes()
}

// KeyItem represents a registered relying party key pair.
type KeyItem struct {
	ID         [32]byte
	PrivateKey *ecdsa.PrivateKey
	Counter    uint64
}

// NewKeyItem initializes a new keyItem for a given appID, and stores it into ks.
func (ks *KeyStorage) NewKeyItem(appID []byte) (ki *KeyItem, err error) {
	if appID == nil || len(appID) == 0 {
		return nil, errors.New("appID can't be empty")
	}
	ki = &KeyItem{}

	id := []byte{}
	copy(id, appID)

	rd := make([]byte, 32)
	rand.Read(rd)

	id = append(id, rd...)

	rhash := sha256.Sum256(id)
	ki.ID = rhash

	ki.PrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err == nil {
		ks.M[ki.ID] = *ki
	}

	_, err = Device.Write(ks.Bytes())
	if err != nil {
		ki = nil
		err = fmt.Errorf("cannot store new KeyItem on Device, %w", err)
		return
	}

	return
}

// Item returns the keyItem associated with key.
func (ks *KeyStorage) Item(key [32]byte) (KeyItem, error) {
	i, ok := ks.M[key]
	if !ok {
		return KeyItem{}, fmt.Errorf("item not found")
	}

	return i, nil
}

// IncrementKeyItem increments Counter of the keyItem associated with key.
func (ks *KeyStorage) IncrementKeyItem(key [32]byte) (uint32, error) {
	i, ok := ks.M[key]
	if !ok {
		return 0, fmt.Errorf("item not found")
	}

	i.Counter++

	ks.M[key] = i

	_, err := Device.Write(ks.Bytes())
	if err != nil {
		return 0, fmt.Errorf("cannot store new KeyItem on Device, %w", err)
	}

	return uint32(i.Counter), nil
}
