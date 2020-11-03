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

// Storage holds a KeyStorge instance and a low-level device io.ReadWriter.
type Storage struct {
	ks     KeyStorage
	device io.ReadWriter
}

// New returns a new Storage instance with the given device.
// If device is nil, the returned Storage instance have it set to a NilDevice instance.
// If data is nil, the returned Storage instance have it set to an empty KeyStorage.
func New(device io.ReadWriter, data []byte) (*Storage, error) {
	if device == nil {
		device = NilDevice{}
	}

	var ks *KeyStorage
	if data == nil {
		ks = NewKeyStorage()
	} else {
		k, err := LoadStorage(data)
		if err != nil {
			return nil, err
		}

		ks = &k
	}

	return &Storage{
		*ks,
		device,
	}, nil
}

// LoadStorage loads a serialized KeyStorage instance from b.
func LoadStorage(b []byte) (KeyStorage, error) {
	rb := bytes.NewReader(b)
	decoder := gob.NewDecoder(rb)
	if decoder == nil {
		return KeyStorage{}, errors.New("gob decoder is nil")
	}

	var ks KeyStorage
	err := decoder.Decode(&ks)
	if err != nil {
		return KeyStorage{}, fmt.Errorf("error while decoding storage bytes, %w", err)
	}

	return ks, nil
}

// KeyStorage stores registered relying party keys.
type KeyStorage struct {
	M map[[32]byte]KeyItem
}

// NewKeyStorage initializes a new keyStorage instance.
func NewKeyStorage() *KeyStorage {
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
func (s *Storage) NewKeyItem(appID []byte) (ki *KeyItem, err error) {
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
		s.ks.M[ki.ID] = *ki
	}

	_, err = s.device.Write(s.ks.Bytes())
	if err != nil {
		ki = nil
		err = fmt.Errorf("cannot store new KeyItem on Device, %w", err)
		return
	}

	return
}

// Item returns the keyItem associated with key.
func (s *Storage) Item(key [32]byte) (KeyItem, error) {
	i, ok := s.ks.M[key]
	if !ok {
		return KeyItem{}, fmt.Errorf("item not found")
	}

	return i, nil
}

// IncrementKeyItem increments Counter of the keyItem associated with key.
func (s *Storage) IncrementKeyItem(key [32]byte) (uint32, error) {
	i, ok := s.ks.M[key]
	if !ok {
		return 0, fmt.Errorf("item not found")
	}

	i.Counter++

	s.ks.M[key] = i

	_, err := s.device.Write(s.ks.Bytes())
	if err != nil {
		return 0, fmt.Errorf("cannot store new KeyItem on Device, %w", err)
	}

	return uint32(i.Counter), nil
}
