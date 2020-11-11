package keyring

import "crypto/ecdsa"

var RetrievePrivatekey = retrievePrivkey

type NonceFuncType = func() ([]byte, error)
type KeygenFuncType = func(b []byte) (*ecdsa.PrivateKey, error)

func SetNonceFunc(f NonceFuncType) NonceFuncType {
	nonceFunc = f
	return nonce
}

func SetKeygenFunc(f KeygenFuncType) KeygenFuncType {
	keygenFunc = f
	return generateECKey
}
