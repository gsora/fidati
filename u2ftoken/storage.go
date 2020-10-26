package u2ftoken

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

var ks *keyStorage

type keyStorage struct {
	M map[[32]byte]keyItem
}

func newKeyStorage() *keyStorage {
	return &keyStorage{
		M: map[[32]byte]keyItem{},
	}
}

func (ks *keyStorage) Bytes() []byte {
	ret := new(bytes.Buffer)
	enc := gob.NewEncoder(ret)

	if err := enc.Encode(ks); err != nil {
		panic(fmt.Sprintf("cannot encode keyStorage: %s", err.Error()))
	}

	return ret.Bytes()
}

type keyItem struct {
	ID         [32]byte
	PrivateKey *ecdsa.PrivateKey
	Counter    uint64
}

func (ks *keyStorage) newKeyItem(appID []byte) (ki *keyItem, err error) {
	ki = &keyItem{}

	id := []byte{}
	copy(id, appID)

	serPk := elliptic.Marshal(elliptic.P256(), attestationPrivkey.X, attestationPrivkey.X)
	id = append(id, serPk...)

	rhash := sha256.Sum256(id)
	ki.ID = rhash

	ki.PrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err == nil {
		ks.M[ki.ID] = *ki
	}

	return
}

func (ks *keyStorage) item(key [32]byte) (keyItem, error) {
	i, ok := ks.M[key]
	if !ok {
		return keyItem{}, fmt.Errorf("item not found")
	}

	return i, nil
}

func (ks *keyStorage) incrementKeyItem(key [32]byte) (uint32, error) {
	i, ok := ks.M[key]
	if !ok {
		return 0, fmt.Errorf("item not found")
	}

	i.Counter++

	ks.M[key] = i
	return uint32(i.Counter), nil
}
