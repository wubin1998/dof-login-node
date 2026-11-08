package main

import (
	"syscall"
	"unsafe"
)

var (
	dwmapi                    = syscall.NewLazyDLL("dwmapi.dll")
	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
)

// setDarkTitleBar makes the native caption blend with the dark UI.
// The caption/text colors only work on Windows 11; older systems simply
// keep the dark-mode caption.
func setDarkTitleBar(hwnd uintptr) {
	if hwnd == 0 {
		return
	}

	dark := int32(1)
	r, _, _ := procDwmSetWindowAttribute.Call(
		hwnd, 20, uintptr(unsafe.Pointer(&dark)), unsafe.Sizeof(dark),
	)
	if r != 0 {
		_, _, _ = procDwmSetWindowAttribute.Call(
			hwnd, 19, uintptr(unsafe.Pointer(&dark)), unsafe.Sizeof(dark),
		)
	}

	// COLORREF is 0x00BBGGRR.
	caption := uint32(0x00120B0A) // #0a0b12
	text := uint32(0x00E9EFF2)    // #f2efe9
	_, _, _ = procDwmSetWindowAttribute.Call(
		hwnd, 35, uintptr(unsafe.Pointer(&caption)), unsafe.Sizeof(caption),
	)
	_, _, _ = procDwmSetWindowAttribute.Call(
		hwnd, 36, uintptr(unsafe.Pointer(&text)), unsafe.Sizeof(text),
	)
}
