package fidati

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"

	"github.com/f-secure-foundry/tamago/soc/imx6/usb"
	"github.com/gsora/fidati/u2fhid"
)

const (
	hidRequestSetIdle           = 10
	hidRequestTypeGetDescriptor = 0x21
	descriptorTypeGetReport     = 0x22
)

// hidDescriptor represents a HID standard descriptor.
// Device Class Definition for Human Interface Devices (HID) Version 1.11, pg 22.
type hidDescriptor struct {
	Length               uint8
	Type                 uint8
	bcdHID               uint16
	CountryCode          uint8
	NumDescriptors       uint8
	ReportDescriptorType uint8
	DescriptorLength     uint16
}

// setDefaults sets some standard properties for hidDescriptor.
func (d *hidDescriptor) setDefaults() {
	d.Length = 0x09
	d.Type = 0x21
	d.bcdHID = 0x101
}

// bytes converts the descriptor structure to byte array format.
func (d *hidDescriptor) bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

// configureDevice configures device to use hidSetup Setup function, and adds an HID InterfaceDescriptor to conf
// along with the needed Endpoints.
func configureDevice(device *usb.Device, conf *usb.ConfigurationDescriptor, u2fHandler *u2fhid.Handler) error {
	device.Setup = hidSetup(device)

	id, err := addInterface(device, conf)
	if err != nil {
		return fmt.Errorf("cannot add U2F USB Interface, %w", err)
	}

	endpoints := addEndpoints(id)
	endpoints.in.Function = u2fHandler.Tx
	endpoints.out.Function = u2fHandler.Rx

	addHIDClassDescriptor(id)

	// device qualifier
	device.Qualifier = &usb.DeviceQualifierDescriptor{}
	device.Qualifier.SetDefaults()
	device.Qualifier.NumConfigurations = uint8(len(device.Configurations))

	return nil
}

// addInterface adds a Interface Descriptor with 2 endpoints, with HID interface class.
func addInterface(device *usb.Device, conf *usb.ConfigurationDescriptor) (*usb.InterfaceDescriptor, error) {
	id := &usb.InterfaceDescriptor{}
	id.SetDefaults()

	id.NumEndpoints = 2
	id.InterfaceClass = 0x03
	id.InterfaceSubClass = 0x0
	id.InterfaceProtocol = 0x0

	var err error
	id.Interface, err = device.AddString("fidati interface descriptor")
	if err != nil {
		return nil, err
	}

	conf.AddInterface(id)
	return id, nil
}

// endpoints is a convenience struct, holds input and output endpoints.
type endpoints struct {
	in  *usb.EndpointDescriptor
	out *usb.EndpointDescriptor
}

// addEndpoints adds an input and output endpoint to conf, returns a endpoints instance to let
// the caller determine their behavior.
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

// addHIDClassDescriptor adds a HID class descriptor to conf.
// The report descriptor length is len(u2fhid.DefaultReport).
func addHIDClassDescriptor(conf *usb.InterfaceDescriptor) {
	hid := hidDescriptor{}
	hid.setDefaults()
	hid.CountryCode = 0x0
	hid.NumDescriptors = 0x01
	hid.ReportDescriptorType = 0x22

	hid.DescriptorLength = uint16(len(u2fhid.DefaultReport))

	conf.ClassDescriptors = append(conf.ClassDescriptors, hid.bytes())
}

// hidSetup returns a custom setup function for device.
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

		if int(setup.RequestType) & ^0x80 == hidRequestTypeGetDescriptor {
			if setup.Request == hidRequestSetIdle {
				ack = true
				done = true
				return
			}
		}

		if setup.Request == usb.GET_DESCRIPTOR {
			if bDescriptorType == descriptorTypeGetReport {
				in = u2fhid.DefaultReport.Bytes()
				done = true
				return
			}
		}

		return
	}
}

// DefaultConfiguration returns a usb.ConfigurationDescriptor ready to be used for ConfigureUSB.
func DefaultConfiguration() usb.ConfigurationDescriptor {
	cd := usb.ConfigurationDescriptor{}
	cd.SetDefaults()
	cd.Attributes = 160

	return cd
}

// ConfigureUSB configures device and config to be used as a FIDO2 U2F token.
func ConfigureUSB(config *usb.ConfigurationDescriptor, device *usb.Device, u2fHandler *u2fhid.Handler) error {
	return configureDevice(device, config, u2fHandler)
}
