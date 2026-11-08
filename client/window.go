package main

import (
	"math"
	"syscall"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
)

// The UI is designed for a 1040x680 CSS-pixel viewport. On high-DPI screens
// the WebView2 CSS viewport is smaller than the physical window, so scale the
// native window up to keep the design at its intended size.

var (
	procGetDpiForWindow   = user32.NewProc("GetDpiForWindow")
	procMonitorFromWindow = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfoW   = user32.NewProc("GetMonitorInfoW")
	procGetWindowRect     = user32.NewProc("GetWindowRect")
	procSetWindowPos      = user32.NewProc("SetWindowPos")
	procGetDC             = user32.NewProc("GetDC")
	procReleaseDC         = user32.NewProc("ReleaseDC")

	gdi32             = syscall.NewLazyDLL("gdi32.dll")
	procGetDeviceCaps = gdi32.NewProc("GetDeviceCaps")
)

type winRect struct {
	Left, Top, Right, Bottom int32
}

type monitorInfo struct {
	Size    uint32
	Monitor winRect
	Work    winRect
	Flags   uint32
}

func windowDPI(hwnd uintptr) float64 {
	// GetDpiForWindow exists on Windows 10 1607+. On Windows 7/8 the
	// LazyProc lookup fails, so fall back to the screen DC.
	if err := procGetDpiForWindow.Find(); err == nil {
		if dpi, _, _ := procGetDpiForWindow.Call(hwnd); dpi != 0 {
			return float64(dpi) / 96.0
		}
	}

	if hdc, _, _ := procGetDC.Call(0); hdc != 0 {
		const logPixelsX = 88
		dpi, _, _ := procGetDeviceCaps.Call(hdc, logPixelsX)
		_, _, _ = procReleaseDC.Call(0, hdc)
		if dpi != 0 {
			return float64(dpi) / 96.0
		}
	}
	return 1
}

func monitorWorkArea(hwnd uintptr) (winRect, bool) {
	const monitorDefaultToNearest = 2

	monitor, _, _ := procMonitorFromWindow.Call(hwnd, monitorDefaultToNearest)
	if monitor == 0 {
		return winRect{}, false
	}

	info := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
	if ok, _, _ := procGetMonitorInfoW.Call(monitor, uintptr(unsafe.Pointer(&info))); ok == 0 {
		return winRect{}, false
	}
	return info.Work, true
}

func centerOnMonitor(hwnd uintptr) {
	work, ok := monitorWorkArea(hwnd)
	if !ok {
		return
	}

	var wr winRect
	if ok, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&wr))); ok == 0 {
		return
	}

	width := wr.Right - wr.Left
	height := wr.Bottom - wr.Top
	x := work.Left + (work.Right-work.Left-width)/2
	y := work.Top + (work.Bottom-work.Top-height)/2

	const (
		swpNoSize     = 0x0001
		swpNoZOrder   = 0x0004
		swpNoActivate = 0x0010
	)
	_, _, _ = procSetWindowPos.Call(
		hwnd, 0,
		uintptr(x), uintptr(y), 0, 0,
		swpNoSize|swpNoZOrder|swpNoActivate,
	)
}

func fitWindow(w webview2.WebView) {
	hwnd := uintptr(w.Window())
	if hwnd == 0 {
		return
	}

	scale := windowDPI(hwnd)
	width := int(math.Round(1040 * scale))
	height := int(math.Round(680 * scale))
	minWidth := int(math.Round(920 * scale))
	minHeight := int(math.Round(620 * scale))

	// Keep the window inside the usable screen area on small/high-DPI
	// displays. The page itself is responsive down to 920x620 CSS pixels.
	if work, ok := monitorWorkArea(hwnd); ok {
		maxWidth := int(math.Round(float64(work.Right-work.Left) * 0.94))
		maxHeight := int(math.Round(float64(work.Bottom-work.Top) * 0.94))
		if maxWidth > 0 && maxHeight > 0 && (width > maxWidth || height > maxHeight) {
			ratio := math.Min(float64(maxWidth)/float64(width), float64(maxHeight)/float64(height))
			width = int(math.Round(float64(width) * ratio))
			height = int(math.Round(float64(height) * ratio))
			minWidth = int(math.Min(float64(minWidth), float64(width)))
			minHeight = int(math.Min(float64(minHeight), float64(height)))
		}
	}

	w.SetSize(width, height, webview2.HintNone)
	w.SetSize(minWidth, minHeight, webview2.HintMin)
	centerOnMonitor(hwnd)
}
