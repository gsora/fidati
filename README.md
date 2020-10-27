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

What doesn't work:
- persistence (rebooting the Armory makes everything go away)

Since `fidati` currently doesn't store the keys it generates, it's practically useless.

A persistence strategy will be implemented sometime soon.

## Building and running

You can run `fidati` with or without a bootloader.

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
