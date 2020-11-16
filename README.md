# `fidati`: DIY FIDO2 U2F token

`fidati` is a FIDO2 U2F token implementation for the F-Secure USB Armory Mk.II, written in Go by leveraging the [Tamago](https://github.com/f-secure-foundry/tamago) compiler.

This repository holds a developer-friendly Tamago firmware, for a more user-friendly one check out [`GoKey`](https://github.com/f-secure-foundry/gokey).

## Project status: **PoC**

This project is still very much a Proof-of-Concept and should be handled as such: **there are exactly zero guarantees about the safety/security of fidati**.

Code is still work-in-progress, expect bugs/bad practices and so on.

What works:
 - HID interface
 - key generation
 - site registration
 - site authentication

`fidati` uses the microSD card as its support for persistency. 

Currently no filesystem is supported, so `fidati` will use up the entire microSD space if needed.

This means that `fidati` can only be ran from the Armory eMMC - a future revision will fix this.

To prepare a microSD for `fidati`, zero out the first 512 bytes:

```bash
dd if=/dev/zero of=/dev/mmcblk0 bs=512 count=1
```

No relying party private key is stored, the microSD is only used to store a monotonic counter.

For more details about how `fidati` deterministic key derivation works, see [here](https://www.yubico.com/blog/yubicos-u2f-key-wrapping/).

## Building and running

You can run `fidati` with or without a bootloader.

By default the project `Makefile` produces a binary with logging disabled.

To enable logging append `TARGET="'usbarmory debug fidati_logs'"` to the `make` parameters.

`fidati` as a library disables logging by default.

To enable it, build your program with the `fidati_logs` build tag.

### Booting via U-Boot

```
$ make
```

This command will produce a self-standing ELF executable, `fidati`, which can be booted via U-Boot in the usual way:

```
ext4load mmc 0:1 0x80800000 /fidati
bootelf 0x80800000
```

### Booting without a bootloader

```
$ make imx
```

This command will produce a i.MX native image, `fidati.imx`, which can be flashed to either the internal Armory eMMC or a microSD.

Refer to [these instructions](https://github.com/f-secure-foundry/usbarmory/wiki/Boot-Modes-(Mk-II)#flashing-imx-native-images) for further instructions.

## Usage as a library

`fidati` can be used as a library, by importing the `github.com/gsora/fidati` package and invoking the `ConfigureUSB()` function.

See `firmware/main.go` and `firmware/usb.go` for an example.

## Technical details

`fidati` implements the bare minimum functionality to act as a FIDO2 U2F token, as detailed by the [FIDO Alliance](https://fidoalliance.org/specifications/download/).

A default attestation certificate and private key are contained in this repository, in the `/certs` directory.

A CLI tool &ndash; `gen-cert` &ndash; is available for those who want to generate their own certificate and private key.

For each relying party, given their `appID` and a device-specific master key `fidati` derives in a deterministic fashion an ECDSA private key, which will then be used in the registration and authentication phase.

The derivation algorithm is defined as follows:

```
nonce := (32 secure random bytes)
relyingPartyPrivateKey := HMAC-SHA256(MasterKey, appID, nonce)
keyHandle := HMAC-SHA256(MasterKey, appID, relyingPartyPrivateKey) + nonce
```

To derive the private key back given a `keyHandle` and `appID`, one must extract the `nonce` by reading the last 32 bytes of `keyHandle` and then execute the algorithm again.

## Debugging

To test U2F token registration and login, the following tools can be used:
 - https://mdp.github.io/u2fdemo/
 - https://demo.yubico.com/webauthn-technical/registration
 - https://github.com/Yubico/java-webauthn-server/
