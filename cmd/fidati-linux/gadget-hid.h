#include <errno.h>
#include <stdio.h>
#include <linux/usb/ch9.h>
#include <usbg/usbg.h>
#include <usbg/function/hid.h>

#define VENDOR          0x1d6b
#define PRODUCT         0x0142

int configure_hidg(
		const char* serial, 
		const char* manufacturer, 
		const char* product,
		const char* configfs_path,
		const char* report_descriptor,
		size_t report_descriptor_len
		);

int cleanup_usbg(const char* configfs_path);


