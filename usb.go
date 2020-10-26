package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"

	"github.com/f-secure-foundry/tamago/soc/imx6/usb"

	"github.com/f-secure-foundry/tamago-example/u2fhid"
)

const (
	HID_SET_IDLE       = 10
	GET_HID_DESCRIPTOR = 0x21
	GET_REPORT         = 0x22
)

type hidDescriptor struct {
	Length               uint8
	Type                 uint8
	bcdHID               uint16
	CountryCode          uint8
	NumDescriptors       uint8
	ReportDescriptorType uint8
	DescriptorLength     uint16
}

func (d *hidDescriptor) SetDefaults() {
	d.Length = 0x09
	d.Type = 0x21
	d.bcdHID = 0x101
}

// Bytes converts the descriptor structure to byte array format.
func (d *hidDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

func configureDevice(device *usb.Device, u2fHandler *u2fhid.Handler) {
	// Supported Language Code Zero: English
	device.SetLanguageCodes([]uint16{0x0409})

	// device descriptor
	device.Descriptor = &usb.DeviceDescriptor{}
	device.Descriptor.SetDefaults()

	// p5, Table 1-1. Device Descriptor Using Class Codes for IAD,
	// USB Interface Association Descriptor Device Class Code and Use Model.
	device.Descriptor.DeviceClass = 0x0
	device.Descriptor.DeviceSubClass = 0x0
	device.Descriptor.DeviceProtocol = 0x0

	// http://pid.codes/1209/2702/
	device.Descriptor.VendorId = 0x1209
	device.Descriptor.ProductId = 0x2702

	device.Descriptor.Device = 0x0001

	iManufacturer, _ := device.AddString(`gsora`)
	device.Descriptor.Manufacturer = iManufacturer

	iProduct, _ := device.AddString(`fidati`)
	device.Descriptor.Product = iProduct

	iSerial, _ := device.AddString(`0.42`)
	device.Descriptor.SerialNumber = iSerial

	device.Setup = hidSetup(device)

	conf := addConfiguration(device)
	id := addInterface(device, conf)
	endpoints := addEndpoints(id)
	endpoints.in.Function = u2fHandler.Tx
	endpoints.out.Function = u2fHandler.Rx

	addClassDescriptors(id)

	// device qualifier
	device.Qualifier = &usb.DeviceQualifierDescriptor{}
	device.Qualifier.SetDefaults()
	device.Qualifier.NumConfigurations = uint8(len(device.Configurations))
}

func addConfiguration(device *usb.Device) *usb.ConfigurationDescriptor {
	cd := &usb.ConfigurationDescriptor{}
	cd.SetDefaults()
	cd.Attributes = 160
	device.AddConfiguration(cd)

	return cd
}

func addInterface(device *usb.Device, conf *usb.ConfigurationDescriptor) *usb.InterfaceDescriptor {

	id := &usb.InterfaceDescriptor{}
	id.SetDefaults()

	id.NumEndpoints = 2
	id.InterfaceClass = 0x03
	id.InterfaceSubClass = 0x0
	id.InterfaceProtocol = 0x0
	id.Interface, _ = device.AddString("fidati interface descriptor")

	conf.AddInterface(id)
	return id
}

type endpoints struct {
	in  *usb.EndpointDescriptor
	out *usb.EndpointDescriptor
}

func addEndpoints(conf *usb.InterfaceDescriptor) endpoints {
	var e endpoints

	e.in = &usb.EndpointDescriptor{}
	e.in.SetDefaults()
	e.in.Attributes = 0x03
	e.in.EndpointAddress = 0x81
	e.in.MaxPacketSize = 63
	e.in.Interval = 5

	e.out = &usb.EndpointDescriptor{}
	e.out.SetDefaults()
	e.out.Attributes = 0x03
	e.out.EndpointAddress = 0x01
	e.out.MaxPacketSize = 63
	e.out.Interval = 5

	conf.Endpoints = append(conf.Endpoints, e.out, e.in)

	return e
}

func addClassDescriptors(conf *usb.InterfaceDescriptor) {
	hid := hidDescriptor{}
	hid.SetDefaults()
	hid.CountryCode = 0x0
	hid.NumDescriptors = 0x01
	hid.ReportDescriptorType = 0x22

	hid.DescriptorLength = uint16(len(u2fhid.DefaultReport))

	conf.ClassDescriptors = append(conf.ClassDescriptors, hid.Bytes())
}

func hidSetup(device *usb.Device) usb.SetupFunction {
	return func(setup *usb.SetupData) (in []byte, done, ack bool, err error) {
		bDescriptorType := setup.Value & 0xff

		log.Println("descriptor type:", bDescriptorType, setup)

		if setup.Request == usb.SET_FEATURE {
			// stall here
			err = errors.New("should stall")
			done = true
			return
		}

		if int(setup.RequestType) & ^0x80 == GET_HID_DESCRIPTOR {
			if setup.Request == HID_SET_IDLE {
				ack = true
				done = true
				return
			}
		}

		if setup.Request == usb.GET_DESCRIPTOR {
			if bDescriptorType == GET_REPORT {
				log.Println("handling hid get_report")
				in = u2fhid.DefaultReport.Bytes()
				done = true
				return
			}
		}

		return
	}
}

func StartUSB() {
	device := &usb.Device{}
	u2f := &u2fhid.Handler{}

	configureDevice(device, u2f)

	usb.USB1.Init()
	usb.USB1.DeviceMode()
	usb.USB1.Reset()

	// never returns
	usb.USB1.Start(device)
}
