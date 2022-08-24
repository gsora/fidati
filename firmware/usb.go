package main

import (
	"github.com/usbarmory/tamago/soc/nxp/usb"

	"github.com/gsora/fidati"
	"github.com/gsora/fidati/keyring"
	"github.com/gsora/fidati/u2fhid"
	"github.com/gsora/fidati/u2ftoken"
)

func baseConfiguration(device *usb.Device) {
	// Supported Language Code Zero: English
	device.SetLanguageCodes([]uint16{0x0409})

	// device descriptor
	device.Descriptor = &usb.DeviceDescriptor{}
	device.Descriptor.SetDefaults()

	// HID devices sets those in the Interface descriptor.
	device.Descriptor.DeviceClass = 0x0
	device.Descriptor.DeviceSubClass = 0x0
	device.Descriptor.DeviceProtocol = 0x0

	// http://pid.codes/1209/2702/
	// Standard USB Armory {Vendor,Product}ID
	device.Descriptor.VendorId = 0x1209
	device.Descriptor.ProductId = 0x2702

	device.Descriptor.Device = 0x0001

	iManufacturer, err := device.AddString(`gsora`)
	notErr(err)
	device.Descriptor.Manufacturer = iManufacturer

	iProduct, err := device.AddString(`fidati`)
	notErr(err)
	device.Descriptor.Product = iProduct

	iSerial, err := device.AddString(`0.42`)
	notErr(err)
	device.Descriptor.SerialNumber = iSerial
}

func startUSB(keyring *keyring.Keyring) {
	device := &usb.Device{}

	token, err := u2ftoken.New(keyring, attestationCertificate, attestationPrivkey)
	notErr(err)

	hid, err := u2fhid.NewHandler(token)
	notErr(err)

	conf := fidati.DefaultConfiguration()

	baseConfiguration(device)

	err = device.AddConfiguration(&conf)
	notErr(err)

	err = fidati.ConfigureUSB(&conf, device, hid)
	notErr(err)

	usb.USB1.Init()
	usb.USB1.DeviceMode()
	usb.USB1.Reset()

	// never returns
	usb.USB1.Start(device)
}
