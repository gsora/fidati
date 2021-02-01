# fidati-linux

This directory holds `fidati-linux`, a Go program which leverages Linux kernel to run a `fidati` U2F token in userspace.

This is a **development tool**, since it has no security guarantees and doesn't store the usage counter in a persistent way.

## Dependencies

`fidati-linux` requires the following components to run:

 - a Linux kernel configured with the `libcomposite`, `dummy_hcd`, `configfs` modules
 - root privileges
 - [`libusbgx`](https://github.com/libusbgx/libusbgx)

`fidati-linux` simulates a full-blown USB HID device by leveraging the `dummy_hcd` kernel module.

The `libusbgx` dependency is needed to properly configure and tear down the virtual USB device.

## Building and usage

To build `fidati-linux`:

```bash
make fidati-linux
```

Run `./fidati-linux -h` to see every configuration parameter.
