package main

import (
	"fmt"
	"math"
	"unsafe"
)

func (u *UI) wndProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmEraseBkgnd:
		return 1
	case wmPaint:
		u.onPaint()
		return 0
	case wmCommand:
		u.onCommand(loword(wParam), hiword(wParam), HWND(lParam))
		return 0
	case wmDrawItem:
		u.onDrawItem(lParam)
		return 0
	case wmTimer:
		u.onTimer(wParam)
		return 0
	case wmCtlColorEdit:
		hdc := HDC(wParam)
		setTextColor(hdc, colText)
		setBkColor(hdc, colField)
		return uintptr(u.brush(colField))
	case wmCtlColorBtn:
		setBkColor(HDC(wParam), colCard)
		return uintptr(u.brush(colCard))
	case wmClose:
		_, _, _ = pDestroyWindow.Call(uintptr(hwnd))
		return 0
	case wmDestroy:
		u.shutdown()
		_, _, _ = pPostQuitMessage.Call(0)
		return 0
	case wmAppStatus:
		u.onStatusResult()
		return 0
	case wmAppLogin:
		u.onLoginResult()
		return 0
	case wmAppRegister:
		u.onRegisterResult()
		return 0
	case wmAppTest:
		u.onTestResult()
		return 0
	}
	return defWindowProc(hwnd, msg, wParam, lParam)
}

func (u *UI) onPaint() {
	var ps paintStruct
	hdc, _, _ := pBeginPaint.Call(uintptr(u.hwnd), uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	if u.backDC != 0 {
		u.drawUI(u.backDC)
		_, _, _ = pBitBlt.Call(
			uintptr(hdc), 0, 0, uintptr(u.client.Right), uintptr(u.client.Bottom),
			uintptr(u.backDC), 0, 0, srcCopy,
		)
	}
	_, _, _ = pEndPaint.Call(uintptr(u.hwnd), uintptr(unsafe.Pointer(&ps)))
}

func (u *UI) drawUI(dc HDC) {
	if u.bgDC != 0 {
		_, _, _ = pBitBlt.Call(uintptr(dc), 0, 0, uintptr(u.client.Right), uintptr(u.client.Bottom), uintptr(u.bgDC), 0, 0, srcCopy)
	} else {
		fillRect(dc, u.client, u.brush(rgb(6, 7, 12)))
	}

	u.drawCard(dc)
	if u.success {
		u.drawSuccess(dc)
	} else if u.modal {
		u.drawModal(dc)
	}
	u.drawToast(dc)
}

func (u *UI) rrect(dc HDC, r rect, radius int32, fill, border uint32) {
	roundRect(dc, r, radius*2, radius*2, u.brush(fill), u.pen(border, 1))
}

func (u *UI) drawCard(dc HDC) {
	card := rect{u.S(cardX), u.S(cardY), u.S(cardX + cardW), u.S(cardY + cardH)}
	u.rrect(dc, card, u.S(26), colCard, colCardBorder)

	// headers/footer separators and subtle gold corner accents
	line(dc, card.Left+u.S(44), card.Top, card.Right-u.S(44), card.Top, u.pen(rgb(74, 66, 48), 1))
	line(dc, card.Left+u.S(12), card.Top+u.S(12), card.Left+u.S(28), card.Top+u.S(12), u.pen(rgb(150, 120, 62), 1))
	line(dc, card.Left+u.S(12), card.Top+u.S(12), card.Left+u.S(12), card.Top+u.S(28), u.pen(rgb(150, 120, 62), 1))
	line(dc, card.Right-u.S(12), card.Bottom-u.S(12), card.Right-u.S(28), card.Bottom-u.S(12), u.pen(rgb(150, 120, 62), 1))
	line(dc, card.Right-u.S(12), card.Bottom-u.S(12), card.Right-u.S(12), card.Bottom-u.S(28), u.pen(rgb(150, 120, 62), 1))

	// tab track behind the owner-drawn tab buttons
	track := rect{u.S(contentX), u.S(126), u.S(contentX + 176), u.S(168)}
	u.rrect(dc, track, u.S(13), colTabBg, colCardBorder)

	for _, f := range u.form.fields {
		focused := u.focusID == f.id && !u.modal
		border := colFieldBorder
		iconColor := colMuted
		if focused {
			border = colGold
			iconColor = colGold
		}
		u.rrect(dc, f.r, u.S(14), colField, border)
		cx := f.r.Left + u.S(27)
		cy := (f.r.Top + f.r.Bottom) / 2
		switch f.icon {
		case 0:
			u.drawUserIcon(dc, cx, cy, iconColor)
		case 1:
			u.drawLockIcon(dc, cx, cy, iconColor)
		default:
			u.drawShieldIcon(dc, cx, cy, iconColor)
		}
	}

	if u.activeTab == 0 {
		// The remembered/forgot buttons are native owner-drawn controls.
	} else {
		u.textRect(dc, "注册后自动获得 100000 点券", 12, 400, colDim, u.form.hint, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
	}

	u.textRect(dc, u.tipText(), 11, 400, colTip, u.form.tip, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)
	u.drawCardFooter(dc)
	u.drawBrandFooter(dc)
	u.drawChips(dc)
}

func (u *UI) tipText() string {
	if u.activeTab == 1 {
		return "请牢记账号密码，注册成功后即可登录"
	}
	return "登录成功后将自动启动 DNF.exe"
}

func (u *UI) drawCardFooter(dc HDC) {
	y := u.S(542)
	line(dc, u.S(contentX), y, u.S(contentX+contentW), y, u.pen(colFieldBorder, 1))
	textY := u.S(548)
	u.statusDot(dc, u.S(contentX+3), textY+u.S(7))
	rr := rect{u.S(contentX + 14), textY - u.S(2), u.S(contentX + contentW - 52), textY + u.S(18)}
	u.textRect(dc, u.statusText, 11, 400, u.statusColor(), rr, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
	ver := rect{u.S(contentX + contentW - 48), textY - u.S(2), u.S(contentX + contentW), textY + u.S(18)}
	u.textRect(dc, "v1.0.0", 10, 400, rgb(79, 83, 100), ver, dtRight|dtVCenter|dtSingleLine|dtNoPrefix)
}

func (u *UI) drawBrandFooter(dc HDC) {
	y := u.S(634)
	line(dc, u.S(42), y, u.S(556), y, u.pen(colFieldBorder, 1))
	textY := u.S(640)
	u.statusDot(dc, u.S(45), textY+u.S(10))
	rr := rect{u.S(56), textY, u.S(440), textY + u.S(20)}
	u.textRect(dc, u.statusText, 11, 400, u.statusColor(), rr, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
	sig := rect{u.S(360), textY, u.S(556), textY + u.S(20)}
	u.textRect(dc, "DOF LOGIN CLIENT", 10, 400, rgb(79, 83, 100), sig, dtRight|dtVCenter|dtSingleLine|dtNoPrefix)
}

func (u *UI) drawChips(dc HDC) {
	cfg := u.app.store.Get()
	text := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
	font := u.font(u.S(12), 400)
	tw := measureText(dc, text, font)
	x := u.S(42)
	y := u.S(515)
	h := u.S(30)
	w := u.S(26) + u.S(5) + u.S(8) + tw
	r := rect{x, y, x + w, y + h}
	u.rrect(dc, r, h/2, colChipBg, colChipBorder)
	ellipse(dc, rect{r.Left + u.S(13), y + h/2 - u.S(2), r.Left + u.S(17), y + h/2 + u.S(2)}, u.brush(colGold), u.nullPen())
	u.textRect(dc, text, 12, 400, rgb(168, 172, 189),
		rect{r.Left + u.S(26), y, r.Right - u.S(6), y + h}, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)

	x2 := r.Right + u.S(10)
	ver := "v1.0.0"
	vw := measureText(dc, ver, font) + u.S(26)
	r2 := rect{x2, y, x2 + vw, y + h}
	u.rrect(dc, r2, h/2, colChipBg, colChipBorder)
	u.textRect(dc, ver, 12, 400, colDim, r2, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)
}

func (u *UI) statusColor() uint32 {
	switch u.statusKind {
	case 1:
		return rgb(154, 199, 173)
	case 2:
		return rgb(201, 142, 150)
	}
	return colGold
}

func (u *UI) statusDotColor() uint32 {
	switch u.statusKind {
	case 1:
		return colGreen
	case 2:
		return colRed
	}
	return colGold
}

func (u *UI) statusDot(dc HDC, cx, cy int32) {
	r := rect{cx - u.S(3), cy - u.S(3), cx + u.S(3), cy + u.S(3)}
	ellipse(dc, r, u.brush(u.statusDotColor()), u.nullPen())
}

func (u *UI) textRect(dc HDC, text string, size, weight int32, color uint32, r rect, flags uint32) {
	f := u.font(u.S(size), weight)
	old := selectObject(dc, HGDIOBJ(f))
	setBkMode(dc, transparent)
	setTextColor(dc, color)
	rr := r
	drawText(dc, text, &rr, flags)
	selectObject(dc, old)
}

// ------------------------------- modal ------------------------------------

func (u *UI) drawModal(dc HDC) {
	alphaFill(dc, u.client, 150)

	card := rect{u.S(290), u.S(100), u.S(750), u.S(580)}
	u.rrect(dc, card, u.S(22), colCard, colCardBorder)
	u.textRect(dc, "服务器设置", 20, 600, colText,
		rect{u.S(314), u.S(120), u.S(726), u.S(148)}, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
	u.textRect(dc, "修改后立即生效，并保存在本机", 12, 400, colDim,
		rect{u.S(314), u.S(150), u.S(726), u.S(170)}, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)

	u.settingsLabel(dc, "服务器地址", 314, 178)
	u.settingsField(dc, rect{u.S(314), u.S(196), u.S(726), u.S(240)}, idCfgServer)

	u.settingsLabel(dc, "数据库端口", 314, 254)
	u.settingsField(dc, rect{u.S(314), u.S(272), u.S(446), u.S(316)}, idCfgPort)
	u.settingsLabel(dc, "数据库密码", 458, 254)
	u.settingsField(dc, rect{u.S(458), u.S(272), u.S(726), u.S(316)}, idCfgPass)

	u.settingsLabel(dc, "游戏路径  (可选)", 314, 330)
	u.settingsField(dc, rect{u.S(314), u.S(348), u.S(726), u.S(392)}, idCfgGame)

	if u.testText != "" {
		color := colDim
		if u.testKind == 1 {
			color = colGreen
		} else if u.testKind == 2 {
			color = colRed
		}
		u.textRect(dc, u.testText, 12, 400, color,
			rect{u.S(430), u.S(414), u.S(726), u.S(446)}, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
	}
}

func (u *UI) settingsLabel(dc HDC, text string, x, y int32) {
	u.textRect(dc, text, 12, 400, colMuted,
		rect{u.S(x), u.S(y), u.S(x + 200), u.S(y + 18)}, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
}

func (u *UI) settingsField(dc HDC, r rect, id uint32) {
	border := colFieldBorder
	if u.focusID == id {
		border = colGold
	}
	u.rrect(dc, r, u.S(11), colField, border)
}

// ------------------------------- success ----------------------------------

func (u *UI) drawSuccess(dc HDC) {
	alphaFill(dc, u.client, 225)
	cx := u.client.Right / 2
	cy := u.client.Bottom/2 - u.S(30)

	r := u.S(39)
	ring := rect{cx - r, cy - r, cx + r, cy + r}
	ellipse(dc, ring, u.brush(rgb(18, 20, 31)), u.pen(colGold, u.S(2)))
	u.drawCheckIcon(dc, cx, cy+u.S(2), colGold, u.S(18))

	u.textRect(dc, "登录成功", 22, 600, colText,
		rect{cx - u.S(160), cy + r + u.S(18), cx + u.S(160), cy + r + u.S(48)}, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)
	u.textRect(dc, "正在启动 DNF.exe，请稍候…", 12, 400, colMuted,
		rect{cx - u.S(200), cy + r + u.S(50), cx + u.S(200), cy + r + u.S(72)}, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)

	bar := rect{cx - u.S(88), cy + r + u.S(88), cx + u.S(88), cy + r + u.S(91)}
	u.rrect(dc, bar, u.S(2), colFieldBorder, colFieldBorder)
	gold := rect{bar.Left, bar.Top, bar.Left + u.S(74), bar.Bottom}
	u.rrect(dc, gold, u.S(2), colGold, colGold)
}

// -------------------------------- toast -----------------------------------

func (u *UI) toastRect() rect {
	return rect{u.S(330), u.S(596), u.S(710), u.S(644)}
}

func (u *UI) drawToast(dc HDC) {
	if u.toastText == "" {
		return
	}
	r := u.toastRect()
	u.rrect(dc, r, u.S(13), rgb(18, 20, 31), colCardBorder)

	accent := colGold
	if u.toastKind == 1 {
		accent = colGreen
	} else if u.toastKind == 2 {
		accent = colRed
	}
	ax := r.Left + u.S(1)
	ay := r.Top + u.S(10)
	ah := r.Bottom - r.Top - u.S(20)
	fillRect(dc, rect{ax, ay, ax + u.S(3), ay + ah}, u.brush(accent))

	cy := (r.Top + r.Bottom) / 2
	switch u.toastKind {
	case 1:
		u.drawCheckIcon(dc, r.Left+u.S(24), cy, accent, u.S(7))
	case 2:
		u.drawErrorIcon(dc, r.Left+u.S(24), cy, accent, u.S(8))
	default:
		u.drawInfoIcon(dc, r.Left+u.S(24), cy, accent, u.S(8))
	}
	u.textRect(dc, u.toastText, 12, 400, colText,
		rect{r.Left + u.S(42), r.Top + u.S(4), r.Right - u.S(12), r.Bottom - u.S(4)},
		dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
}

func (u *UI) setStatus(kind int, text string) {
	u.statusKind = kind
	u.statusText = text
	invalidate(u.hwnd, nil)
}

func (u *UI) showToast(kind int, text string) {
	u.toastKind = kind
	u.toastText = text
	setTimer(u.hwnd, timerToast, 3800)
	r := u.toastRect()
	invalidate(u.hwnd, &r)
}

// -------------------------------- icons -----------------------------------

func (u *UI) drawUserIcon(dc HDC, cx, cy int32, color uint32) {
	pen := u.pen(color, 1)
	ellipse(dc, rect{cx - u.S(4), cy - u.S(8), cx + u.S(4), cy}, u.nullBrush(), pen)
	pts := []point{
		{cx - u.S(8), cy + u.S(8)},
		{cx - u.S(7), cy + u.S(4)},
		{cx - u.S(4), cy + u.S(2)},
		{cx, cy + u.S(1)},
		{cx + u.S(4), cy + u.S(2)},
		{cx + u.S(7), cy + u.S(4)},
		{cx + u.S(8), cy + u.S(8)},
	}
	polyline(dc, pts, pen)
}

func (u *UI) drawLockIcon(dc HDC, cx, cy int32, color uint32) {
	pen := u.pen(color, 1)
	body := rect{cx - u.S(7), cy - u.S(1), cx + u.S(7), cy + u.S(9)}
	roundRect(dc, body, u.S(3), u.S(3), u.nullBrush(), pen)
	pts := []point{
		{cx - u.S(4), cy - u.S(1)},
		{cx - u.S(4), cy - u.S(6)},
		{cx - u.S(2), cy - u.S(9)},
		{cx + u.S(2), cy - u.S(9)},
		{cx + u.S(4), cy - u.S(6)},
		{cx + u.S(4), cy - u.S(1)},
	}
	polyline(dc, pts, pen)
	ellipse(dc, rect{cx - u.S(1), cy + u.S(3), cx + u.S(1), cy + u.S(5)}, u.brush(color), u.nullPen())
}

func (u *UI) drawShieldIcon(dc HDC, cx, cy int32, color uint32) {
	pen := u.pen(color, 1)
	pts := []point{
		{cx, cy - u.S(9)},
		{cx + u.S(7), cy - u.S(5)},
		{cx + u.S(6), cy + u.S(3)},
		{cx, cy + u.S(9)},
		{cx - u.S(6), cy + u.S(3)},
		{cx - u.S(7), cy - u.S(5)},
		{cx, cy - u.S(9)},
	}
	polyline(dc, pts, pen)
	check := []point{{cx - u.S(3), cy}, {cx - u.S(1), cy + u.S(2)}, {cx + u.S(4), cy - u.S(3)}}
	polyline(dc, check, pen)
}

func (u *UI) drawEyeIcon(dc HDC, cx, cy int32, color uint32, shown bool) {
	pen := u.pen(color, 1)
	ellipse(dc, rect{cx - u.S(8), cy - u.S(4), cx + u.S(8), cy + u.S(4)}, u.nullBrush(), pen)
	ellipse(dc, rect{cx - u.S(2), cy - u.S(2), cx + u.S(2), cy + u.S(2)}, u.brush(color), u.nullPen())
	if shown {
		line(dc, cx-u.S(9), cy+u.S(6), cx+u.S(9), cy-u.S(6), pen)
	}
}

func (u *UI) drawGearIcon(dc HDC, cx, cy int32, color uint32) {
	pen := u.pen(color, 1)
	ellipse(dc, rect{cx - u.S(5), cy - u.S(5), cx + u.S(5), cy + u.S(5)}, u.nullBrush(), pen)
	ellipse(dc, rect{cx - u.S(2), cy - u.S(2), cx + u.S(2), cy + u.S(2)}, u.nullBrush(), pen)
	for i := 0; i < 8; i++ {
		a := float64(i) * math.Pi / 4
		x1 := cx + int32(math.Cos(a)*float64(u.S(7)))
		y1 := cy + int32(math.Sin(a)*float64(u.S(7)))
		x2 := cx + int32(math.Cos(a)*float64(u.S(9)))
		y2 := cy + int32(math.Sin(a)*float64(u.S(9)))
		line(dc, x1, y1, x2, y2, pen)
	}
}

func (u *UI) drawCheckIcon(dc HDC, cx, cy int32, color uint32, size int32) {
	pen := u.pen(color, maxInt32(u.S(2), 2))
	pts := []point{
		{cx - size, cy},
		{cx - size/3, cy + size/2},
		{cx + size, cy - size},
	}
	polyline(dc, pts, pen)
}

func (u *UI) drawErrorIcon(dc HDC, cx, cy int32, color uint32, size int32) {
	pen := u.pen(color, maxInt32(u.S(1), 1))
	ellipse(dc, rect{cx - size, cy - size, cx + size, cy + size}, u.nullBrush(), pen)
	line(dc, cx-u.S(4), cy-u.S(4), cx+u.S(4), cy+u.S(4), pen)
	line(dc, cx+u.S(4), cy-u.S(4), cx-u.S(4), cy+u.S(4), pen)
}

func (u *UI) drawInfoIcon(dc HDC, cx, cy int32, color uint32, size int32) {
	pen := u.pen(color, maxInt32(u.S(1), 1))
	ellipse(dc, rect{cx - size, cy - size, cx + size, cy + size}, u.nullBrush(), pen)
	line(dc, cx, cy-u.S(4), cx, cy+u.S(4), pen)
	line(dc, cx, cy-u.S(6), cx, cy-u.S(5), pen)
}

func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func (u *UI) InvalidateCard() {
	invalidate(u.hwnd, nil)
}
