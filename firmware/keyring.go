package main

import "github.com/gsora/fidati/keyring"

func genKeyring(secret []byte, counter keyring.Counter) *keyring.Keyring {
	return keyring.New(secret, counter)
}
