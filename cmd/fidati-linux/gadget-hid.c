/*
 * Originally taken from libusbgx examples, modified for fidati-linux 
 * needs.
 *
 * Copyright (C) 2014 Samsung Electronics
 *
 * Krzysztof Opasiak <k.opasiak@samsung.com>
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 */

#include "gadget-hid.h"

int configure_hidg(
		const char* serial, 
		const char* manufacturer, 
		const char* product,
		const char* configfs_path,
		const char* report_descriptor, 
		size_t report_descriptor_len
		) {
	usbg_state *s;
	usbg_gadget *g;
	usbg_config *c;
	usbg_function *f_hid;
	int ret = -EINVAL;
	int usbg_ret;

	struct usbg_gadget_attrs g_attrs = {
		.bcdUSB = 0x0200,
		.bDeviceClass =	USB_CLASS_PER_INTERFACE,
		.bDeviceSubClass = 0x00,
		.bDeviceProtocol = 0x00,
		.bMaxPacketSize0 = 64, /* Max allowed ep0 packet size */
		.idVendor = VENDOR,
		.idProduct = PRODUCT,
		.bcdDevice = 0x0001, /* Verson of device */
	};

	struct usbg_gadget_strs g_strs = {
		.serial = (char *)serial, /* Serial number */
		.manufacturer = (char *)manufacturer, /* Manufacturer */
		.product = (char *)product /* Product string */
	};

	struct usbg_config_strs c_strs = {
		.configuration = "1xHID"
	};

	struct usbg_f_hid_attrs f_attrs = {
		.protocol = 0x21,
		.report_desc = {
			.desc = (char *)report_descriptor,
			.len = report_descriptor_len,
		},
		.report_length = 64,
		.subclass = 0,
	};

	usbg_ret = usbg_init(configfs_path, &s);
	if (usbg_ret != USBG_SUCCESS) {
		goto out1;
	}

	usbg_ret = usbg_create_gadget(s, "g1", &g_attrs, &g_strs, &g);
	if (usbg_ret != USBG_SUCCESS) {
		goto out2;
	}

	usbg_ret = usbg_create_function(g, USBG_F_HID, "usb0", &f_attrs, &f_hid);
	if (usbg_ret != USBG_SUCCESS) {
		goto out2;
	}

	usbg_ret = usbg_create_config(g, 1, "fidati-linux", NULL, &c_strs, &c);
	if (usbg_ret != USBG_SUCCESS) {
		goto out2;
	}

	usbg_ret = usbg_add_config_function(c, "u2fhid", f_hid);
	if (usbg_ret != USBG_SUCCESS) {
		goto out2;
	}

	usbg_ret = usbg_enable_gadget(g, DEFAULT_UDC);
	if (usbg_ret != USBG_SUCCESS) {
		goto out2;
	}

	ret = 0;
out2:
	usbg_cleanup(s);

out1:
	return ret;
}

int remove_gadget(usbg_gadget *g)
{
	int usbg_ret;
	usbg_udc *u;

	/* Check if gadget is enabled */
	u = usbg_get_gadget_udc(g);

	/* If gadget is enable we have to disable it first */
	if (u) {
		usbg_ret = usbg_disable_gadget(g);
		if (usbg_ret != USBG_SUCCESS) {
			goto out;
		}
	}

	/* Remove gadget with USBG_RM_RECURSE flag to remove
	 * also its configurations, functions and strings */
	usbg_ret = usbg_rm_gadget(g, USBG_RM_RECURSE);

out:
	return usbg_ret;
}

int cleanup_usbg(const char* configfs_path)
{
	int usbg_ret;
	int ret = -EINVAL;
	usbg_state *s;
	usbg_gadget *g;
	struct usbg_gadget_attrs g_attrs;

	usbg_ret = usbg_init(configfs_path, &s);
	if (usbg_ret != USBG_SUCCESS) {
		goto out1;
	}

	g = usbg_get_first_gadget(s);
	while (g != NULL) {
		/* Get current gadget attrs to be compared */
		usbg_ret = usbg_get_gadget_attrs(g, &g_attrs);
		if (usbg_ret != USBG_SUCCESS) {
			goto out2;
		}

		/* Compare attrs with given values and remove if suitable */
		if (g_attrs.idVendor == VENDOR && g_attrs.idProduct == PRODUCT) {
			usbg_gadget *g_next = usbg_get_next_gadget(g);

			usbg_ret = remove_gadget(g);
			if (usbg_ret != USBG_SUCCESS)
				goto out2;

			g = g_next;
		} else {
			g = usbg_get_next_gadget(g);
		}
	}

out2:
	usbg_cleanup(s);
out1:
	return ret;
}
