# `fidati`: DIY FIDO2 U2F token

`fidati` is a FIDO2 U2F token implementation for the F-Secure USB Armory Mk.II, written in Go by leveraging the [Tamago](https://github.com/f-secure-foundry/tamago) compiler.

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

## Building and running

You can run `fidati` with or without a bootloader.

By default the project `Makefile` produces a binary with logging disabled.

To enable logging append `TARGET="'usbarmory debug'"` to the `make` parameters.

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

For each relying party `fidati` creates a new ECDSA keypair.

The key handle is defined as follows:

```
keyHandle := applicationID + attestationPrivateKey
``` 

## Debugging

To test U2F token registration and login, the following tools can be used:
 - https://mdp.github.io/u2fdemo/
 - https://demo.yubico.com/webauthn-technical/registration
 - https://github.com/Yubico/java-webauthn-server/
