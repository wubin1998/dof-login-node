package main

import (
	"bytes"
	_ "embed"
	"image/png"
	"math"
	"syscall"
	"time"
	"unsafe"
)

//go:embed assets/background.png
var backgroundPNG []byte

const (
	designW = 1040
	designH = 680

	cardX = 598
	cardY = 100
	cardW = 400
	cardH = 480

	contentX = cardX + 28
	contentW = cardW - 56

	// Control IDs
	idTabLogin    = 1001
	idTabRegister = 1002
	idGear        = 1003
	idUser        = 1004
	idPass        = 1005
	idConfirm     = 1006
	idEyePass     = 1007
	idEyeConfirm  = 1008
	idRemember    = 1009
	idForgot      = 1010
	idToLogin     = 1011
	idSubmit      = 1012

	idCfgServer = 1020
	idCfgPort   = 1021
	idCfgPass   = 1022
	idCfgGame   = 1023
	idCfgTest   = 1024
	idCfgCancel = 1025
	idCfgSave   = 1026

	// Timer IDs
	timerSpin    = 1
	timerToast   = 3
	timerSuccess = 4
	timerStatus  = 5

	// Custom window messages
	wmAppStatus   = wmApp + 1
	wmAppLogin    = wmApp + 2
	wmAppRegister = wmApp + 3
	wmAppTest     = wmApp + 4
)

var (
	colCard        = rgb(14, 16, 26)
	colCardBorder  = rgb(42, 45, 60)
	colField       = rgb(20, 23, 36)
	colFieldBorder = rgb(42, 46, 64)
	colTabBg       = rgb(21, 24, 35)
	colTabActive   = rgb(42, 35, 22)
	colGold        = rgb(247, 198, 90)
	colGoldSoft    = rgb(255, 233, 173)
	colEmber       = rgb(255, 122, 60)
	colText        = rgb(242, 239, 233)
	colMuted       = rgb(139, 143, 163)
	colDim         = rgb(99, 103, 121)
	colGreen       = rgb(69, 212, 131)
	colRed         = rgb(255, 95, 116)
	colLink        = rgb(163, 167, 184)
	colBtnText     = rgb(33, 21, 5)
	colChipBg      = rgb(21, 24, 35)
	colChipBorder  = rgb(42, 46, 62)
	colTip         = rgb(102, 107, 125)
)

type fontKey struct {
	size   int32
	weight int32
}

type penKey struct {
	color uint32
	width int32
}

type uiResultLogin struct {
	user      string
	uid       int
	bad       bool
	err       error
	launched  bool
	launchErr error
}

type uiResultRegister struct {
	user    string
	created bool
	err     error
}

type uiResultTest struct {
	latency time.Duration
	err     error
}

type uiResultStatus struct {
	seq     int
	latency time.Duration
	err     error
}

type fieldLayout struct {
	id     uint32
	r      rect
	icon   int // 0 user, 1 lock, 2 shield
	hasEye bool
}

type formLayout struct {
	fields   []fieldLayout
	metaY    int32
	submit   rect
	tip      rect
	remember rect
	forgot   rect
	toLogin  rect
	hint     rect
}

type UI struct {
	app  *App
	hwnd HWND
	inst HINSTANCE

	scale  float64
	client rect

	activeTab    int // 0 login, 1 register
	loading      bool
	success      bool
	modal        bool
	passShown    bool
	confirmShown bool
	remember     bool
	focusID      uint32

	statusKind    int // 0 busy, 1 ok, 2 err
	statusText    string
	statusSeq     int
	statusRunning bool

	toastKind int // 0 info, 1 ok, 2 err
	toastText string

	testKind int // -1 none, 0 busy, 1 ok, 2 err
	testText string
	testBusy bool

	spinPhase int

	form formLayout

	// Controls
	tabLogin, tabRegister, gear        HWND
	userEdit, passEdit, confirmEdit    HWND
	eyePass, eyeConfirm                HWND
	rememberBtn, forgotBtn, toLoginBtn HWND
	submitBtn                          HWND

	cfgServer, cfgPort, cfgPass, cfgGame HWND
	cfgTest, cfgCancel, cfgSave          HWND

	// GDI caches
	fonts   map[fontKey]HFONT
	brushes map[uint32]HBRUSH
	pens    map[penKey]HPEN

	bgDC    HDC
	bgBmp   HBITMAP
	backDC  HDC
	backBmp HBITMAP

	done chan struct{}

	loginCh    chan uiResultLogin
	registerCh chan uiResultRegister
	testCh     chan uiResultTest
	statusCh   chan uiResultStatus
}

func NewUI(app *App) *UI {
	return &UI{
		app:        app,
		fonts:      map[fontKey]HFONT{},
		brushes:    map[uint32]HBRUSH{},
		pens:       map[penKey]HPEN{},
		done:       make(chan struct{}),
		loginCh:    make(chan uiResultLogin, 1),
		registerCh: make(chan uiResultRegister, 1),
		testCh:     make(chan uiResultTest, 1),
		statusCh:   make(chan uiResultStatus, 1),
		statusKind: 0,
		statusText: "正在检测服务器…",
		testKind:   -1,
	}
}

func (u *UI) run() {
	u.createWindow()
	u.msgLoop()
}

func (u *UI) createWindow() {
	initCommonControls()
	u.inst = getModuleHandle()

	className := "DOFNativeLogin"
	title := "DOF 登录器 · 地下城与勇士"

	icon := loadIcon(u.inst, 2, int32(smCXIconValue()), int32(smCYIconValue()))
	iconSm := loadIcon(u.inst, 2, int32(smCXSmIconValue()), int32(smCYSmIconValue()))
	cursor, _, _ := pLoadCursorW.Call(0, idcArrow)

	wc := wndClassExW{
		Size:  uint32(unsafe.Sizeof(wndClassExW{})),
		Style: csHRedraw | csVRedraw,
		WndProc: syscall.NewCallback(func(hwnd, msg, w, l uintptr) uintptr {
			return u.wndProc(HWND(hwnd), uint32(msg), w, l)
		}),
		Instance:  u.inst,
		Icon:      icon,
		Cursor:    cursor,
		ClassName: utf16Ptr(className),
		IconSm:    iconSm,
	}
	registerClassEx(&wc)

	style := uint32(wsOverlapped | wsCaption | wsSysMenu | wsMinimizeBox | wsClipChildren)
	u.hwnd = createWindowEx(
		uint32(wsExControlParent|wsExAppWindow),
		className, title, style,
		0, 0, designW, designH,
		0, 0, u.inst, 0,
	)
	if u.hwnd == 0 {
		messageBox("DOF 登录器", "窗口创建失败。")
		return
	}

	// Work out the effective scale and size the window before showing it.
	dpi := u.getDPI()
	u.scale = float64(dpi) / 96.0
	u.clampScale()
	cw := int32(math.Round(designW * u.scale))
	ch := int32(math.Round(designH * u.scale))

	var wr rect
	wr.Right = cw
	wr.Bottom = ch
	_, _, _ = pAdjustWindowRectEx.Call(
		uintptr(unsafe.Pointer(&wr)), uintptr(style), 0,
		uintptr(uint32(wsExControlParent|wsExAppWindow)),
	)

	work, _ := u.workArea()
	winW := wr.Right - wr.Left
	winH := wr.Bottom - wr.Top
	x := work.Left + (work.Right-work.Left-winW)/2
	y := work.Top + (work.Bottom-work.Top-winH)/2
	setWindowPos(u.hwnd, x, y, winW, winH, swpNoZOrder|swpNoActivate|swpFrameChanged)

	u.client = rect{0, 0, cw, ch}
	u.createBackBuffer()
	u.loadBackground()
	u.createControls()
	u.layoutForm()
	u.applyDarkTitleBar()

	showWindow(u.hwnd, swShow)
	_, _, _ = pUpdateWindow.Call(uintptr(u.hwnd))
	u.app.logf("native window ready: %dx%d scale=%.3f visible=%v", cw, ch, u.scale, isWindowVisible(u.hwnd))

	if cfg := u.app.store.Get(); cfg.LastUser != "" {
		u.remember = true
		setWindowText(u.userEdit, cfg.LastUser)
		invalidate(u.rememberBtn, nil)
		setFocus(u.passEdit)
	} else {
		setFocus(u.userEdit)
	}

	u.startStatusCheck()
	setTimer(u.hwnd, timerStatus, 30000)
}

func (u *UI) clampScale() {
	work, ok := u.workArea()
	if !ok {
		return
	}
	maxW := float64(work.Right-work.Left) * 0.94 / designW
	maxH := float64(work.Bottom-work.Top) * 0.92 / designH
	if maxW < u.scale {
		u.scale = maxW
	}
	if maxH < u.scale {
		u.scale = maxH
	}
	if u.scale < 0.75 {
		u.scale = 0.75
	}
}

func smCXIconValue() int32 {
	r, _, _ := pGetSystemMetrics.Call(smCXIcon)
	return int32(r)
}
func smCYIconValue() int32 {
	r, _, _ := pGetSystemMetrics.Call(smCYIcon)
	return int32(r)
}
func smCXSmIconValue() int32 {
	r, _, _ := pGetSystemMetrics.Call(smCXSmIcon)
	return int32(r)
}
func smCYSmIconValue() int32 {
	r, _, _ := pGetSystemMetrics.Call(smCYSmIcon)
	return int32(r)
}

func (u *UI) getDPI() int {
	if err := pGetDpiForWindow.Find(); err == nil {
		if dpi, _, _ := pGetDpiForWindow.Call(uintptr(u.hwnd)); dpi != 0 {
			return int(dpi)
		}
	}
	if hdc, _, _ := pGetDC.Call(0); hdc != 0 {
		dpi, _, _ := pGetDeviceCaps.Call(hdc, 88) // LOGPIXELSX
		_, _, _ = pReleaseDC.Call(0, hdc)
		if dpi != 0 {
			return int(dpi)
		}
	}
	return 96
}

func (u *UI) workArea() (rect, bool) {
	mon, _, _ := pMonitorFromWindow.Call(uintptr(u.hwnd), monitorDefaultToNearest)
	if mon == 0 {
		return rect{}, false
	}
	info := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
	ok, _, _ := pGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&info)))
	if ok == 0 {
		return rect{}, false
	}
	return info.Work, true
}

func (u *UI) applyDarkTitleBar() {
	dark := int32(1)
	_, _, _ = pDwmSetWindowAttribute.Call(uintptr(u.hwnd), 20, uintptr(unsafe.Pointer(&dark)), unsafe.Sizeof(dark))
	caption := uint32(0x00120B0A)
	text := uint32(0x00E9EFF2)
	_, _, _ = pDwmSetWindowAttribute.Call(uintptr(u.hwnd), 35, uintptr(unsafe.Pointer(&caption)), unsafe.Sizeof(caption))
	_, _, _ = pDwmSetWindowAttribute.Call(uintptr(u.hwnd), 36, uintptr(unsafe.Pointer(&text)), unsafe.Sizeof(text))
}

func (u *UI) msgLoop() {
	var m msg
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) == 0 || int32(r) == -1 {
			break
		}
		if r, _, _ := pIsDialogMessageW.Call(uintptr(u.hwnd), uintptr(unsafe.Pointer(&m))); r != 0 {
			continue
		}
		_, _, _ = pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		_, _, _ = pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func (u *UI) shutdown() {
	select {
	case <-u.done:
	default:
		close(u.done)
	}
	for _, f := range u.fonts {
		deleteObject(HGDIOBJ(f))
	}
	for _, b := range u.brushes {
		deleteObject(HGDIOBJ(b))
	}
	for _, p := range u.pens {
		deleteObject(HGDIOBJ(p))
	}
	if u.backDC != 0 {
		_, _, _ = pDeleteDC.Call(uintptr(u.backDC))
		u.backDC = 0
	}
	if u.backBmp != 0 {
		deleteObject(HGDIOBJ(u.backBmp))
	}
	if u.bgDC != 0 {
		_, _, _ = pDeleteDC.Call(uintptr(u.bgDC))
		u.bgDC = 0
	}
	if u.bgBmp != 0 {
		deleteObject(HGDIOBJ(u.bgBmp))
	}
}

func (u *UI) S(v int32) int32 {
	return int32(math.Round(float64(v) * u.scale))
}

func (u *UI) brush(color uint32) HBRUSH {
	if b, ok := u.brushes[color]; ok {
		return b
	}
	b := createSolidBrush(color)
	u.brushes[color] = b
	return b
}

func (u *UI) pen(color uint32, width int32) HPEN {
	key := penKey{color: color, width: width}
	if p, ok := u.pens[key]; ok {
		return p
	}
	p := createPen(psSolid, width, color)
	u.pens[key] = p
	return p
}

func (u *UI) nullBrush() HBRUSH { return HBRUSH(getStockObject(nullBrush)) }
func (u *UI) nullPen() HPEN     { return HPEN(getStockObject(nullPen)) }

func (u *UI) font(size int32, weight int32) HFONT {
	key := fontKey{size: size, weight: weight}
	if f, ok := u.fonts[key]; ok {
		return f
	}
	f := createFont(size, weight)
	u.fonts[key] = f
	return f
}

func (u *UI) createBackBuffer() {
	dc, _, _ := pGetDC.Call(uintptr(u.hwnd))
	if dc == 0 {
		return
	}
	dcMem, _, _ := pCreateCompatibleDC.Call(dc)
	bmpMem, _, _ := pCreateCompatibleBitmap.Call(dc, uintptr(u.client.Right), uintptr(u.client.Bottom))
	u.backDC = HDC(dcMem)
	u.backBmp = HBITMAP(bmpMem)
	if u.backDC != 0 && u.backBmp != 0 {
		_ = selectObject(u.backDC, HGDIOBJ(u.backBmp))
	}
	_, _, _ = pReleaseDC.Call(uintptr(u.hwnd), dc)
}

func (u *UI) loadBackground() {
	img, err := png.Decode(bytes.NewReader(backgroundPNG))
	if err != nil {
		u.app.logf("decode background: %v", err)
		return
	}
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()

	pixels := make([]byte, sw*sh*4)
	i := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			pixels[i+0] = byte(bl >> 8)
			pixels[i+1] = byte(g >> 8)
			pixels[i+2] = byte(r >> 8)
			pixels[i+3] = byte(a >> 8)
			i += 4
		}
	}

	bmi := bitmapInfo{}
	bmi.Header.Size = uint32(unsafe.Sizeof(bitmapInfoHeader{}))
	bmi.Header.Width = int32(sw)
	bmi.Header.Height = int32(-sh)
	bmi.Header.Planes = 1
	bmi.Header.BitCount = 32
	bmi.Header.Compression = biRGB
	bmi.Header.SizeImage = uint32(len(pixels))

	var bits unsafe.Pointer
	hbm, _, _ := pCreateDIBSection.Call(0, uintptr(unsafe.Pointer(&bmi)), dibRGBColors, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if hbm == 0 || bits == nil {
		u.app.logf("CreateDIBSection failed")
		return
	}
	copy(unsafe.Slice((*byte)(bits), len(pixels)), pixels)

	srcDC, _, _ := pCreateCompatibleDC.Call(0)
	old, _, _ := pSelectObject.Call(srcDC, hbm)

	windowDC, _, _ := pGetDC.Call(uintptr(u.hwnd))
	bgDC, _, _ := pCreateCompatibleDC.Call(windowDC)
	bgBmp, _, _ := pCreateCompatibleBitmap.Call(windowDC, uintptr(u.client.Right), uintptr(u.client.Bottom))
	u.bgDC = HDC(bgDC)
	u.bgBmp = HBITMAP(bgBmp)
	_ = selectObject(u.bgDC, HGDIOBJ(u.bgBmp))

	_, _, _ = pSetStretchBltMode.Call(uintptr(u.bgDC), halfTone)
	_, _, _ = pSetBrushOrgEx.Call(uintptr(u.bgDC), 0, 0, 0)
	_, _, _ = pStretchBlt.Call(
		uintptr(u.bgDC), 0, 0, uintptr(u.client.Right), uintptr(u.client.Bottom),
		srcDC, 0, 0, uintptr(sw), uintptr(sh),
		srcCopy,
	)
	_, _, _ = pReleaseDC.Call(uintptr(u.hwnd), windowDC)

	_, _, _ = pSelectObject.Call(srcDC, old)
	_, _, _ = pDeleteDC.Call(srcDC)
	_, _, _ = pDeleteObject.Call(hbm)
}

// control helpers

func (u *UI) createEdit(id uint32, password bool, maxLen int32, cue string) HWND {
	style := uint32(wsChild | wsTabStop | esLeft | esAutoHScroll | esNoHideSel)
	if password {
		style |= esPassword
	}
	h := createWindowEx(0, "EDIT", "", style, 0, 0, 10, 10, u.hwnd, uintptr(id), u.inst, 0)
	if h == 0 {
		return 0
	}
	sendMessage(h, wmSetFont, uintptr(u.font(u.S(17), 400)), 1)
	sendMessage(h, emSetLimitText, uintptr(maxLen), 0)
	if password {
		sendMessage(h, emSetPasswordChar, 0x25CF, 0)
	}
	setWindowTheme(h, "", "")
	if cue != "" {
		sendMessage(h, emSetCueBanner, 1, uintptr(unsafe.Pointer(utf16Ptr(cue))))
	}
	return h
}

func (u *UI) createButton(id uint32, text string, extraStyle uint32) HWND {
	style := uint32(wsChild|wsTabStop|bsOwnerDraw) | extraStyle
	h := createWindowEx(0, "BUTTON", text, style, 0, 0, 10, 10, u.hwnd, uintptr(id), u.inst, 0)
	if h != 0 {
		sendMessage(h, wmSetFont, uintptr(u.font(u.S(16), 700)), 1)
		setWindowTheme(h, "", "")
	}
	return h
}

func (u *UI) createControls() {
	u.tabLogin = u.createButton(idTabLogin, "登录", 0)
	u.tabRegister = u.createButton(idTabRegister, "注册", 0)
	u.gear = u.createButton(idGear, "", 0)

	u.userEdit = u.createEdit(idUser, false, 32, "请输入账号")
	u.passEdit = u.createEdit(idPass, true, 64, "请输入密码")
	u.confirmEdit = u.createEdit(idConfirm, true, 64, "再次输入密码")

	u.eyePass = u.createButton(idEyePass, "", 0)
	u.eyeConfirm = u.createButton(idEyeConfirm, "", 0)

	u.rememberBtn = u.createButton(idRemember, "记住账号", 0)
	u.forgotBtn = u.createButton(idForgot, "忘记密码？", 0)
	u.toLoginBtn = u.createButton(idToLogin, "已有账号？", 0)

	u.submitBtn = u.createButton(idOK, "进入游戏", bsDefPushButton)

	u.cfgServer = u.createEdit(idCfgServer, false, 128, "192.168.200.131")
	u.cfgPort = u.createEdit(idCfgPort, false, 5, "3306")
	u.cfgPass = u.createEdit(idCfgPass, true, 128, "留空保持不变")
	u.cfgGame = u.createEdit(idCfgGame, false, 512, "留空则使用同目录 DNF.exe")
	u.cfgTest = u.createButton(idCfgTest, "测试连接", 0)
	u.cfgCancel = u.createButton(idCancel, "取消", 0)
	u.cfgSave = u.createButton(idCfgSave, "保存设置", bsDefPushButton)

	// Fixed header positions.
	u.moveControlXYWH(u.tabLogin, 630, 130, 84, 34, true)
	u.moveControlXYWH(u.tabRegister, 714, 130, 84, 34, true)
	u.moveControlXYWH(u.gear, 932, 128, 38, 38, true)

	// Hidden by default.
	for _, h := range []HWND{u.confirmEdit, u.eyeConfirm, u.toLoginBtn,
		u.cfgServer, u.cfgPort, u.cfgPass, u.cfgGame, u.cfgTest, u.cfgCancel, u.cfgSave} {
		if h != 0 {
			showWindow(h, swHide)
		}
	}
}

func (u *UI) moveControlXYWH(h HWND, x, y, w, hgt int32, visible bool) {
	if h == 0 {
		return
	}
	setWindowPos(h, u.S(x), u.S(y), u.S(w), u.S(hgt), swpNoZOrder|swpNoActivate)
	if visible {
		showWindow(h, swShow)
	} else {
		showWindow(h, swHide)
	}
}

func (u *UI) layoutForm() {
	if u.hwnd == 0 {
		return
	}
	reg := u.activeTab == 1
	fields := 2
	if reg {
		fields = 3
	}
	fieldH := int32(56)
	fieldGap := int32(14)
	contentH := int32(fields)*fieldH + int32(fields-1)*fieldGap + 130
	areaTop := int32(190)
	areaBottom := int32(532)
	startY := areaTop + (areaBottom-areaTop-contentH)/2

	u.form.fields = u.form.fields[:0]
	for i := 0; i < fields; i++ {
		icon := 0
		hasEye := false
		id := uint32(idUser)
		if i == 1 {
			icon = 1
			hasEye = true
			id = idPass
		} else if i == 2 {
			icon = 2
			hasEye = true
			id = idConfirm
		}
		r := rect{u.S(contentX), u.S(startY + int32(i)*(fieldH+fieldGap)),
			u.S(contentX + contentW), u.S(startY + int32(i)*(fieldH+fieldGap) + fieldH)}
		u.form.fields = append(u.form.fields, fieldLayout{id: id, r: r, icon: icon, hasEye: hasEye})
	}
	fieldsBottom := startY + int32(fields)*fieldH + int32(fields-1)*fieldGap
	metaY := fieldsBottom + 12
	submitY := metaY + 36
	tipY := submitY + 66

	u.form.metaY = u.S(metaY)
	u.form.submit = rect{u.S(contentX), u.S(submitY), u.S(contentX + contentW), u.S(submitY + 54)}
	u.form.tip = rect{u.S(contentX), u.S(tipY), u.S(contentX + contentW), u.S(tipY + 16)}
	u.form.remember = rect{u.S(contentX + 2), u.S(metaY), u.S(contentX + 2 + 104), u.S(metaY + 20)}
	u.form.forgot = rect{u.S(contentX + contentW - 92), u.S(metaY), u.S(contentX + contentW), u.S(metaY + 20)}
	u.form.toLogin = rect{u.S(contentX + contentW - 92), u.S(metaY), u.S(contentX + contentW), u.S(metaY + 20)}
	u.form.hint = rect{u.S(contentX + 2), u.S(metaY), u.S(contentX + contentW - 100), u.S(metaY + 20)}

	userField := u.form.fields[0]
	passField := u.form.fields[1]
	u.moveEdit(u.userEdit, userField.r, false)
	u.moveEdit(u.passEdit, passField.r, true)
	if reg {
		u.moveEdit(u.confirmEdit, u.form.fields[2].r, true)
		showWindow(u.confirmEdit, swShow)
		showWindow(u.eyeConfirm, swShow)
	} else {
		showWindow(u.confirmEdit, swHide)
		showWindow(u.eyeConfirm, swHide)
	}

	u.moveEye(u.eyePass, passField.r)
	if reg {
		u.moveEye(u.eyeConfirm, u.form.fields[2].r)
	}

	u.moveControlRect(u.rememberBtn, u.form.remember, !reg)
	u.moveControlRect(u.forgotBtn, u.form.forgot, !reg)
	u.moveControlRect(u.toLoginBtn, u.form.toLogin, reg)
	u.moveControlRect(u.submitBtn, u.form.submit, true)

	if u.loading {
		u.setLoading(true)
	}
	u.InvalidateCard()
}

func (u *UI) moveControlRect(h HWND, r rect, visible bool) {
	if h == 0 {
		return
	}
	setWindowPos(h, r.Left, r.Top, r.Right-r.Left, r.Bottom-r.Top, swpNoZOrder|swpNoActivate)
	if visible {
		showWindow(h, swShow)
	} else {
		showWindow(h, swHide)
	}
}

func (u *UI) moveEdit(h HWND, field rect, hasEye bool) {
	if h == 0 {
		return
	}
	left := field.Left + u.S(48)
	right := field.Right - u.S(16)
	if hasEye {
		right = field.Right - u.S(48)
	}
	top := field.Top + u.S(16)
	bottom := field.Bottom - u.S(16)
	setWindowPos(h, left, top, right-left, bottom-top, swpNoZOrder|swpNoActivate)
	showWindow(h, swShow)
}

func (u *UI) moveEye(h HWND, field rect) {
	if h == 0 {
		return
	}
	right := field.Right - u.S(10)
	top := (field.Top+field.Bottom)/2 - u.S(17)
	setWindowPos(h, right-u.S(34), top, u.S(34), u.S(34), swpNoZOrder|swpNoActivate)
	showWindow(h, swShow)
}
