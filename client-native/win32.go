package main

import (
	"syscall"
	"unsafe"
)

type (
	HWND      uintptr
	HINSTANCE uintptr
	HDC       uintptr
	HGDIOBJ   uintptr
	HBRUSH    uintptr
	HPEN      uintptr
	HFONT     uintptr
	HICON     uintptr
	HBITMAP   uintptr
	HRGN      uintptr
	HMONITOR  uintptr
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	msimg32  = syscall.NewLazyDLL("msimg32.dll")
	dwmapi   = syscall.NewLazyDLL("dwmapi.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")
	uxtheme  = syscall.NewLazyDLL("uxtheme.dll")
)

var (
	pRegisterClassExW   = user32.NewProc("RegisterClassExW")
	pCreateWindowExW    = user32.NewProc("CreateWindowExW")
	pDefWindowProcW     = user32.NewProc("DefWindowProcW")
	pGetMessageW        = user32.NewProc("GetMessageW")
	pTranslateMessage   = user32.NewProc("TranslateMessage")
	pDispatchMessageW   = user32.NewProc("DispatchMessageW")
	pPostQuitMessage    = user32.NewProc("PostQuitMessage")
	pPostMessageW       = user32.NewProc("PostMessageW")
	pSendMessageW       = user32.NewProc("SendMessageW")
	pGetClientRect      = user32.NewProc("GetClientRect")
	pGetWindowRect      = user32.NewProc("GetWindowRect")
	pSetWindowPos       = user32.NewProc("SetWindowPos")
	pAdjustWindowRectEx = user32.NewProc("AdjustWindowRectEx")
	pShowWindow         = user32.NewProc("ShowWindow")
	pIsWindowVisible    = user32.NewProc("IsWindowVisible")
	pUpdateWindow       = user32.NewProc("UpdateWindow")
	pInvalidateRect     = user32.NewProc("InvalidateRect")
	pBeginPaint         = user32.NewProc("BeginPaint")
	pEndPaint           = user32.NewProc("EndPaint")
	pGetDC              = user32.NewProc("GetDC")
	pReleaseDC          = user32.NewProc("ReleaseDC")
	pLoadImageW         = user32.NewProc("LoadImageW")
	pLoadCursorW        = user32.NewProc("LoadCursorW")
	pLoadIconW          = user32.NewProc("LoadIconW")
	pSetWindowTextW     = user32.NewProc("SetWindowTextW")
	pGetWindowTextW     = user32.NewProc("GetWindowTextW")
	pSetFocus           = user32.NewProc("SetFocus")
	pGetFocus           = user32.NewProc("GetFocus")
	pEnableWindow       = user32.NewProc("EnableWindow")
	pDestroyWindow      = user32.NewProc("DestroyWindow")
	pSetTimer           = user32.NewProc("SetTimer")
	pKillTimer          = user32.NewProc("KillTimer")
	pIsDialogMessageW   = user32.NewProc("IsDialogMessageW")
	pGetDlgItem         = user32.NewProc("GetDlgItem")
	pSetWindowLongPtrW  = user32.NewProc("SetWindowLongPtrW")
	pGetSystemMetrics   = user32.NewProc("GetSystemMetrics")
	pMonitorFromWindow  = user32.NewProc("MonitorFromWindow")
	pGetMonitorInfoW    = user32.NewProc("GetMonitorInfoW")
	pSetProcessDPIAware = user32.NewProc("SetProcessDPIAware")
	pMessageBoxW        = user32.NewProc("MessageBoxW")
	pSetCursor          = user32.NewProc("SetCursor")
	pGetCursorPos       = user32.NewProc("GetCursorPos")
	pScreenToClient     = user32.NewProc("ScreenToClient")
	pGetDpiForWindow    = user32.NewProc("GetDpiForWindow")

	pCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pCreateDIBSection       = gdi32.NewProc("CreateDIBSection")
	pSelectObject           = gdi32.NewProc("SelectObject")
	pDeleteObject           = gdi32.NewProc("DeleteObject")
	pDeleteDC               = gdi32.NewProc("DeleteDC")
	pBitBlt                 = gdi32.NewProc("BitBlt")
	pStretchBlt             = gdi32.NewProc("StretchBlt")
	pSetStretchBltMode      = gdi32.NewProc("SetStretchBltMode")
	pSetBrushOrgEx          = gdi32.NewProc("SetBrushOrgEx")
	pCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	pCreatePen              = gdi32.NewProc("CreatePen")
	pCreateFontW            = gdi32.NewProc("CreateFontW")
	pSetBkMode              = gdi32.NewProc("SetBkMode")
	pSetTextColor           = gdi32.NewProc("SetTextColor")
	pSetBkColor             = gdi32.NewProc("SetBkColor")
	pDrawTextW              = user32.NewProc("DrawTextW")
	pGetTextExtentPoint32W  = gdi32.NewProc("GetTextExtentPoint32W")
	pRoundRect              = gdi32.NewProc("RoundRect")
	pRectangle              = gdi32.NewProc("Rectangle")
	pEllipse                = gdi32.NewProc("Ellipse")
	pMoveToEx               = gdi32.NewProc("MoveToEx")
	pLineTo                 = gdi32.NewProc("LineTo")
	pPolyline               = gdi32.NewProc("Polyline")
	pPolygon                = gdi32.NewProc("Polygon")
	pCreateRoundRectRgn     = gdi32.NewProc("CreateRoundRectRgn")
	pSelectClipRgn          = gdi32.NewProc("SelectClipRgn")
	pGetStockObject         = gdi32.NewProc("GetStockObject")
	pGetDeviceCaps          = gdi32.NewProc("GetDeviceCaps")
	pFillRect               = user32.NewProc("FillRect")
	pFrameRect              = user32.NewProc("FrameRect")
	pSetWindowTheme         = uxtheme.NewProc("SetWindowTheme")
	pInitCommonControlsEx   = comctl32.NewProc("InitCommonControlsEx")

	pGradientFill          = msimg32.NewProc("GradientFill")
	pAlphaBlend            = msimg32.NewProc("AlphaBlend")
	pDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")

	pGetModuleHandleW   = kernel32.NewProc("GetModuleHandleW")
	pGetModuleFileNameW = kernel32.NewProc("GetModuleFileNameW")
)

const (
	wsOverlapped   = 0x00000000
	wsCaption      = 0x00C00000
	wsSysMenu      = 0x00080000
	wsMinimizeBox  = 0x00020000
	wsMaximizeBox  = 0x00010000
	wsThickFrame   = 0x00040000
	wsClipChildren = 0x02000000
	wsVisible      = 0x10000000
	wsChild        = 0x40000000
	wsTabStop      = 0x00010000

	wsExControlParent = 0x00010000
	wsExAppWindow     = 0x00040000

	csHRedraw = 0x0002
	csVRedraw = 0x0001

	swHide = 0
	swShow = 5

	swpNoSize       = 0x0001
	swpNoMove       = 0x0002
	swpNoZOrder     = 0x0004
	swpNoActivate   = 0x0010
	swpFrameChanged = 0x0020

	cwUseDefault = 0x80000000

	wmCreate         = 0x0001
	wmDestroy        = 0x0002
	wmSize           = 0x0005
	wmSetFocus       = 0x0007
	wmKillFocus      = 0x0008
	wmPaint          = 0x000F
	wmClose          = 0x0010
	wmQuit           = 0x0012
	wmEraseBkgnd     = 0x0014
	wmShowWindow     = 0x0018
	wmGetMinMaxInfo  = 0x0024
	wmDrawItem       = 0x002B
	wmSetFont        = 0x0030
	wmGetFont        = 0x0031
	wmTimer          = 0x0113
	wmCommand        = 0x0111
	wmCtlColorBtn    = 0x0135
	wmCtlColorEdit   = 0x0133
	wmCtlColorStatic = 0x0138
	wmDpiChanged     = 0x02E0
	wmApp            = 0x8000
	wmUser           = 0x0400

	wmSetText       = 0x000C
	wmGetText       = 0x000D
	wmGetTextLength = 0x000E

	bmSetCheck = 0x00F1
	bmGetCheck = 0x00F0
	bmClick    = 0x00F5

	emSetPasswordChar = 0x00CC
	emSetCueBanner    = 0x1501
	emSetLimitText    = 0x00C5
	emSetSel          = 0x00B1
	emSetMargins      = 0x00D3

	esLeft        = 0x0000
	esCenter      = 0x0001
	esRight       = 0x0002
	esMultiLine   = 0x0004
	esPassword    = 0x0020
	esAutoHScroll = 0x0080
	esNoHideSel   = 0x0100
	esNumber      = 0x2000

	bsPushButton    = 0x00000000
	bsDefPushButton = 0x00000001
	bsCheckBox      = 0x00000002
	bsAutoCheckBox  = 0x00000003
	bsOwnerDraw     = 0x0000000B
	bsLeft          = 0x00000100
	bsCenter        = 0x00000300

	bnClicked   = 0
	enSetFocus  = 0x0100
	enKillFocus = 0x0200
	enChange    = 0x0300

	odsSelected = 0x0001
	odsGrayed   = 0x0002
	odsDisabled = 0x0004
	odsChecked  = 0x0008
	odsFocus    = 0x0010

	dtLeft        = 0x0000
	dtCenter      = 0x0001
	dtRight       = 0x0002
	dtVCenter     = 0x0004
	dtWordBreak   = 0x0010
	dtSingleLine  = 0x0020
	dtNoPrefix    = 0x0800
	dtEndEllipsis = 0x8000

	transparent = 1

	smCXIcon    = 11
	smCYIcon    = 12
	smCXSmIcon  = 49
	smCYSmIcon  = 50
	smCXScreen  = 0
	smCYScreen  = 1
	smCXVScroll = 2
	smCYVScroll = 20

	imageIcon = 1

	lrDefaultColor = 0x00000000
	lrDefaultSize  = 0x00000040
	lrShared       = 0x00008000

	idcArrow       = 32512
	idiApplication = 32512

	idOK     = 1
	idCancel = 2

	iccStandardClasses = 0x00004000

	monitorDefaultToNearest = 2

	psSolid = 0
	psNull  = 5

	nullBrush = 5
	nullPen   = 8

	dibRGBColors = 0
	biRGB        = 0
	srcCopy      = 0x00CC0020
	halfTone     = 4

	gradientFillRectH = 0
	gradientFillRectV = 1

	gwlpID = ^uintptr(11) // -12

	gwlStyle = ^uintptr(15) // -16
)

type (
	point struct{ X, Y int32 }
	rect  struct{ Left, Top, Right, Bottom int32 }
	size  struct{ CX, CY int32 }

	msg struct {
		Hwnd     HWND
		Message  uint32
		_        uint32
		WParam   uintptr
		LParam   uintptr
		Time     uint32
		Pt       point
		LPrivate uint32
	}

	wndClassExW struct {
		Size       uint32
		Style      uint32
		WndProc    uintptr
		ClsExtra   int32
		WndExtra   int32
		Instance   HINSTANCE
		Icon       HICON
		Cursor     uintptr
		Background HBRUSH
		MenuName   *uint16
		ClassName  *uint16
		IconSm     HICON
	}

	paintStruct struct {
		Hdc       HDC
		Erase     int32
		Paint     rect
		Restore   int32
		IncUpdate int32
		Reserved  [32]byte
	}

	drawItemStruct struct {
		CtlType    uint32
		CtlID      uint32
		ItemID     uint32
		ItemAction uint32
		ItemState  uint32
		HwndItem   HWND
		HDC        HDC
		RcItem     rect
		ItemData   uintptr
	}

	monitorInfo struct {
		Size    uint32
		Monitor rect
		Work    rect
		Flags   uint32
	}

	minMaxInfo struct {
		PtReserved     point
		PtMaxSize      point
		PtMaxPosition  point
		PtMinTrackSize point
		PtMaxTrackSize point
	}

	trivertex struct {
		X, Y        int32
		Red, Green  uint16
		Blue, Alpha uint16
	}

	gradientRect struct {
		UpperLeft, LowerRight uint32
	}

	bitmapInfoHeader struct {
		Size          uint32
		Width         int32
		Height        int32
		Planes        uint16
		BitCount      uint16
		Compression   uint32
		SizeImage     uint32
		XPelsPerMeter int32
		YPelsPerMeter int32
		ClrUsed       uint32
		ClrImportant  uint32
	}

	bitmapInfo struct {
		Header bitmapInfoHeader
		Colors [1]uint32
	}

	initCommonControlsEx struct {
		Size uint32
		ICC  uint32
	}
)

func rgb(r, g, b byte) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func loword(v uintptr) uint32 { return uint32(v & 0xffff) }
func hiword(v uintptr) uint32 { return uint32((v >> 16) & 0xffff) }

func makeIntResource(id uintptr) uintptr { return id }

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func getModuleHandle() HINSTANCE {
	h, _, _ := pGetModuleHandleW.Call(0)
	return HINSTANCE(h)
}

func registerClassEx(wc *wndClassExW) uint16 {
	r, _, _ := pRegisterClassExW.Call(uintptr(unsafe.Pointer(wc)))
	return uint16(r)
}

func createWindowEx(exStyle uint32, class, title string, style uint32, x, y, w, h int32, parent HWND, menu uintptr, inst HINSTANCE, param uintptr) HWND {
	r, _, _ := pCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(utf16Ptr(class))),
		uintptr(unsafe.Pointer(utf16Ptr(title))),
		uintptr(style),
		uintptr(uint32(x)), uintptr(uint32(y)),
		uintptr(uint32(w)), uintptr(uint32(h)),
		uintptr(parent), menu, uintptr(inst), param,
	)
	return HWND(r)
}

func defWindowProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := pDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func postMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) {
	_, _, _ = pPostMessageW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
}

func sendMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := pSendMessageW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func setWindowLongPtr(hwnd HWND, index uintptr, value uintptr) {
	_, _, _ = pSetWindowLongPtrW.Call(uintptr(hwnd), index, value)
}

func setWindowText(hwnd HWND, text string) {
	_, _, _ = pSetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func getWindowText(hwnd HWND) string {
	buf := make([]uint16, 512)
	n, _, _ := pGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:n])
}

func setFocus(hwnd HWND) {
	_, _, _ = pSetFocus.Call(uintptr(hwnd))
}

func getFocus() HWND {
	r, _, _ := pGetFocus.Call()
	return HWND(r)
}

func enableWindow(hwnd HWND, enabled bool) {
	v := uintptr(0)
	if enabled {
		v = 1
	}
	_, _, _ = pEnableWindow.Call(uintptr(hwnd), v)
}

func showWindow(hwnd HWND, cmd int32) {
	_, _, _ = pShowWindow.Call(uintptr(hwnd), uintptr(uint32(cmd)))
}

func isWindowVisible(hwnd HWND) bool {
	r, _, _ := pIsWindowVisible.Call(uintptr(hwnd))
	return r != 0
}

func invalidate(hwnd HWND, r *rect) {
	var p uintptr
	if r != nil {
		p = uintptr(unsafe.Pointer(r))
	}
	_, _, _ = pInvalidateRect.Call(uintptr(hwnd), p, 0)
}

func getClientRect(hwnd HWND) rect {
	var r rect
	_, _, _ = pGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	return r
}

func moveWindow(hwnd HWND, x, y, w, h int32) {
	_, _, _ = pSetWindowPos.Call(uintptr(hwnd), 0, uintptr(uint32(x)), uintptr(uint32(y)), uintptr(uint32(w)), uintptr(uint32(h)), swpNoZOrder|swpNoActivate)
}

func setWindowPos(hwnd HWND, x, y, w, h int32, flags uint32) {
	_, _, _ = pSetWindowPos.Call(uintptr(hwnd), 0, uintptr(uint32(x)), uintptr(uint32(y)), uintptr(uint32(w)), uintptr(uint32(h)), uintptr(flags))
}

func setTimer(hwnd HWND, id uintptr, ms uint32) {
	_, _, _ = pSetTimer.Call(uintptr(hwnd), id, uintptr(ms), 0)
}

func killTimer(hwnd HWND, id uintptr) {
	_, _, _ = pKillTimer.Call(uintptr(hwnd), id)
}

func createSolidBrush(color uint32) HBRUSH {
	r, _, _ := pCreateSolidBrush.Call(uintptr(color))
	return HBRUSH(r)
}

func createPen(style uint32, width int32, color uint32) HPEN {
	r, _, _ := pCreatePen.Call(uintptr(style), uintptr(uint32(width)), uintptr(color))
	return HPEN(r)
}

func createFont(px int32, weight int32) HFONT {
	r, _, _ := pCreateFontW.Call(
		uintptr(uint32(-px)), 0, 0, 0,
		uintptr(uint32(weight)), 0, 0, 0,
		1, // DEFAULT_CHARSET
		0, 0,
		5, // CLEARTYPE_QUALITY
		0,
		uintptr(unsafe.Pointer(utf16Ptr("Microsoft YaHei"))),
	)
	return HFONT(r)
}

func selectObject(hdc HDC, obj HGDIOBJ) HGDIOBJ {
	r, _, _ := pSelectObject.Call(uintptr(hdc), uintptr(obj))
	return HGDIOBJ(r)
}

func deleteObject(obj HGDIOBJ) {
	if obj != 0 {
		_, _, _ = pDeleteObject.Call(uintptr(obj))
	}
}

func getStockObject(id int32) HGDIOBJ {
	r, _, _ := pGetStockObject.Call(uintptr(uint32(id)))
	return HGDIOBJ(r)
}

func setBkMode(hdc HDC, mode int32) {
	_, _, _ = pSetBkMode.Call(uintptr(hdc), uintptr(uint32(mode)))
}

func setTextColor(hdc HDC, color uint32) {
	_, _, _ = pSetTextColor.Call(uintptr(hdc), uintptr(color))
}

func setBkColor(hdc HDC, color uint32) {
	_, _, _ = pSetBkColor.Call(uintptr(hdc), uintptr(color))
}

func fillRect(hdc HDC, r rect, brush HBRUSH) {
	_, _, _ = pFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(&r)), uintptr(brush))
}

func drawText(hdc HDC, text string, r *rect, flags uint32) {
	_, _, _ = pDrawTextW.Call(uintptr(hdc), uintptr(unsafe.Pointer(utf16Ptr(text))), ^uintptr(0), uintptr(unsafe.Pointer(r)), uintptr(flags))
}

func measureText(hdc HDC, text string, font HFONT) int32 {
	old := selectObject(hdc, HGDIOBJ(font))
	var s size
	_, _, _ = pGetTextExtentPoint32W.Call(uintptr(hdc), uintptr(unsafe.Pointer(utf16Ptr(text))), uintptr(len([]rune(text))), uintptr(unsafe.Pointer(&s)))
	selectObject(hdc, old)
	return s.CX
}

func roundRect(hdc HDC, r rect, ellipseW, ellipseH int32, brush HBRUSH, pen HPEN) {
	oldB := selectObject(hdc, HGDIOBJ(brush))
	oldP := selectObject(hdc, HGDIOBJ(pen))
	_, _, _ = pRoundRect.Call(uintptr(hdc), uintptr(uint32(r.Left)), uintptr(uint32(r.Top)), uintptr(uint32(r.Right)), uintptr(uint32(r.Bottom)), uintptr(uint32(ellipseW)), uintptr(uint32(ellipseH)))
	selectObject(hdc, oldB)
	selectObject(hdc, oldP)
}

func ellipse(hdc HDC, r rect, brush HBRUSH, pen HPEN) {
	oldB := selectObject(hdc, HGDIOBJ(brush))
	oldP := selectObject(hdc, HGDIOBJ(pen))
	_, _, _ = pEllipse.Call(uintptr(hdc), uintptr(uint32(r.Left)), uintptr(uint32(r.Top)), uintptr(uint32(r.Right)), uintptr(uint32(r.Bottom)))
	selectObject(hdc, oldB)
	selectObject(hdc, oldP)
}

func line(hdc HDC, x1, y1, x2, y2 int32, pen HPEN) {
	oldP := selectObject(hdc, HGDIOBJ(pen))
	_, _, _ = pMoveToEx.Call(uintptr(hdc), uintptr(uint32(x1)), uintptr(uint32(y1)), 0)
	_, _, _ = pLineTo.Call(uintptr(hdc), uintptr(uint32(x2)), uintptr(uint32(y2)))
	selectObject(hdc, oldP)
}

func polyline(hdc HDC, pts []point, pen HPEN) {
	if len(pts) < 2 {
		return
	}
	oldP := selectObject(hdc, HGDIOBJ(pen))
	_, _, _ = pPolyline.Call(uintptr(hdc), uintptr(unsafe.Pointer(&pts[0])), uintptr(len(pts)))
	selectObject(hdc, oldP)
}

func gradientFill(hdc HDC, r rect, top, bottom uint32) {
	var verts [2]trivertex
	verts[0].X, verts[0].Y = r.Left, r.Top
	verts[0].Red = uint16((top & 0xff) * 256)
	verts[0].Green = uint16(((top >> 8) & 0xff) * 256)
	verts[0].Blue = uint16(((top >> 16) & 0xff) * 256)
	verts[0].Alpha = 0
	verts[1].X, verts[1].Y = r.Right, r.Bottom
	verts[1].Red = uint16((bottom & 0xff) * 256)
	verts[1].Green = uint16(((bottom >> 8) & 0xff) * 256)
	verts[1].Blue = uint16(((bottom >> 16) & 0xff) * 256)
	verts[1].Alpha = 0

	grad := gradientRect{0, 1}
	_, _, _ = pGradientFill.Call(
		uintptr(hdc),
		uintptr(unsafe.Pointer(&verts[0])), 2,
		uintptr(unsafe.Pointer(&grad)), 1,
		gradientFillRectV,
	)
}

func roundedGradient(hdc HDC, r rect, radius int32, top, bottom uint32, border uint32) {
	region, _, _ := pCreateRoundRectRgn.Call(
		uintptr(uint32(r.Left)), uintptr(uint32(r.Top)), uintptr(uint32(r.Right+1)), uintptr(uint32(r.Bottom+1)),
		uintptr(uint32(radius*2)), uintptr(uint32(radius*2)),
	)
	oldRgn, _, _ := pSelectClipRgn.Call(uintptr(hdc), region)
	gradientFill(hdc, r, top, bottom)
	_, _, _ = pSelectClipRgn.Call(uintptr(hdc), oldRgn)
	if region != 0 {
		_, _, _ = pDeleteObject.Call(region)
	}
	pen := createPen(psSolid, 1, border)
	roundRect(hdc, r, radius*2, radius*2, HBRUSH(getStockObject(nullBrush)), pen)
	deleteObject(HGDIOBJ(pen))
}

func alphaFill(hdc HDC, r rect, alpha byte) {
	// A 1x1 black source stretched over the target with constant alpha.
	mem, _, _ := pCreateCompatibleDC.Call(uintptr(hdc))
	if mem == 0 {
		return
	}
	bmp, _, _ := pCreateCompatibleBitmap.Call(uintptr(hdc), 1, 1)
	if bmp == 0 {
		_, _, _ = pDeleteDC.Call(mem)
		return
	}
	old, _, _ := pSelectObject.Call(mem, bmp)
	black := createSolidBrush(rgb(0, 0, 0))
	fillRect(HDC(mem), rect{0, 0, 1, 1}, black)
	deleteObject(HGDIOBJ(black))

	blend := uint32(0) | uint32(alpha)<<16 // AC_SRC_OVER, alpha, AlphaFormat 0
	_, _, _ = pAlphaBlend.Call(
		uintptr(hdc),
		uintptr(uint32(r.Left)), uintptr(uint32(r.Top)), uintptr(uint32(r.Right-r.Left)), uintptr(uint32(r.Bottom-r.Top)),
		mem, 0, 0, 1, 1,
		uintptr(blend),
	)
	_, _, _ = pSelectObject.Call(mem, old)
	_, _, _ = pDeleteObject.Call(bmp)
	_, _, _ = pDeleteDC.Call(mem)
}

func loadIcon(inst HINSTANCE, id uintptr, cx, cy int32) HICON {
	r, _, _ := pLoadImageW.Call(uintptr(inst), id, imageIcon, uintptr(uint32(cx)), uintptr(uint32(cy)), lrDefaultColor|lrShared)
	if r == 0 {
		r, _, _ = pLoadIconW.Call(0, idiApplication)
	}
	return HICON(r)
}

func initCommonControls() {
	icc := initCommonControlsEx{Size: uint32(unsafe.Sizeof(initCommonControlsEx{})), ICC: iccStandardClasses}
	_, _, _ = pInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&icc)))
}

func setWindowTheme(hwnd HWND, subApp, subID string) {
	_, _, _ = pSetWindowTheme.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf16Ptr(subApp))), uintptr(unsafe.Pointer(utf16Ptr(subID))))
}

// preloadWin32 loads every DLL and resolves every function pointer before any
// window is created. Resolving a lazy proc from inside a WndProc callback can
// deadlock on the loader lock while USER32 is creating a child control.
func preloadWin32() {
	for _, dll := range []*syscall.LazyDLL{user32, gdi32, kernel32, msimg32, dwmapi, comctl32, uxtheme} {
		_ = dll.Load()
	}
	for _, p := range allProcs {
		_ = p.Find()
	}
	initCommonControls()
}

var allProcs = []*syscall.LazyProc{
	pRegisterClassExW, pCreateWindowExW, pDefWindowProcW, pGetMessageW,
	pTranslateMessage, pDispatchMessageW, pPostQuitMessage, pPostMessageW,
	pSendMessageW, pGetClientRect, pGetWindowRect, pSetWindowPos,
	pAdjustWindowRectEx, pShowWindow, pUpdateWindow, pInvalidateRect,
	pBeginPaint, pEndPaint, pGetDC, pReleaseDC, pLoadImageW, pLoadCursorW,
	pLoadIconW, pSetWindowTextW, pGetWindowTextW, pSetFocus, pGetFocus,
	pEnableWindow, pDestroyWindow, pSetTimer, pKillTimer, pIsDialogMessageW,
	pGetDlgItem, pSetWindowLongPtrW, pGetSystemMetrics, pMonitorFromWindow,
	pGetMonitorInfoW, pSetProcessDPIAware, pMessageBoxW, pSetCursor,
	pGetCursorPos, pScreenToClient, pGetDpiForWindow, pIsWindowVisible,
	pCreateCompatibleDC, pCreateCompatibleBitmap, pCreateDIBSection,
	pSelectObject, pDeleteObject, pDeleteDC, pBitBlt, pStretchBlt,
	pSetStretchBltMode, pSetBrushOrgEx, pCreateSolidBrush, pCreatePen,
	pCreateFontW, pSetBkMode, pSetTextColor, pSetBkColor, pDrawTextW,
	pGetTextExtentPoint32W, pRoundRect, pRectangle, pEllipse, pMoveToEx,
	pLineTo, pPolyline, pPolygon, pCreateRoundRectRgn, pSelectClipRgn,
	pGetStockObject, pGetDeviceCaps, pFillRect, pFrameRect, pSetWindowTheme,
	pInitCommonControlsEx, pGradientFill, pAlphaBlend,
	pDwmSetWindowAttribute, pGetModuleHandleW, pGetModuleFileNameW,
}

func messageBox(title, text string) {
	_, _, _ = pMessageBoxW.Call(0, uintptr(unsafe.Pointer(utf16Ptr(text))), uintptr(unsafe.Pointer(utf16Ptr(title))), 0x10)
}
