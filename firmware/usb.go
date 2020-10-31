package main

import (
	"github.com/f-secure-foundry/tamago/soc/imx6/usb"

	"github.com/gsora/fidati"
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

func startUSB() {
	device := &usb.Device{}

	conf := fidati.DefaultConfiguration()

	baseConfiguration(device)

	err := device.AddConfiguration(&conf)
	if err != nil {
		notErr(err)
	}

	err = fidati.ConfigureUSB(&conf, device)
	notErr(err)

	usb.USB1.Init()
	usb.USB1.DeviceMode()
	usb.USB1.Reset()

	// never returns
	usb.USB1.Start(device)
}

// since we're in a critical configuration phase, panic on error.
func notErr(e error) {
	if e != nil {
		panic(e)
	}
}
