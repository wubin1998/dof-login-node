package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func (u *UI) onCommand(id, code uint32, ctrl HWND) {
	if code == enSetFocus || code == enKillFocus {
		if code == enSetFocus {
			u.focusID = id
		} else if u.focusID == id {
			u.focusID = 0
		}
		invalidate(u.hwnd, nil)
		return
	}

	switch id {
	case idOK:
		if u.modal {
			u.saveSettings()
		} else {
			u.onSubmit()
		}
		return
	case idCancel:
		if u.modal {
			u.closeSettings()
		}
		return
	case idSubmit:
		u.onSubmit()
		return
	case idCfgSave:
		if u.modal {
			u.saveSettings()
		}
		return
	}

	if code != bnClicked {
		return
	}

	switch id {
	case idTabLogin:
		u.setTab(false)
	case idTabRegister:
		u.setTab(true)
	case idGear:
		u.openSettings()
	case idEyePass:
		u.togglePassword(u.passEdit, u.eyePass, &u.passShown)
	case idEyeConfirm:
		u.togglePassword(u.confirmEdit, u.eyeConfirm, &u.confirmShown)
	case idRemember:
		u.remember = !u.remember
		if u.rememberBtn != 0 {
			invalidate(u.rememberBtn, nil)
		}
	case idForgot:
		u.showToast(0, "请联系服务器管理员协助重置密码")
	case idToLogin:
		u.setTab(false)
	case idCfgTest:
		u.startTest()
	}
}

func (u *UI) onTimer(id uintptr) {
	switch id {
	case timerSpin:
		u.spinPhase = (u.spinPhase + 1) % 8
		if u.submitBtn != 0 {
			invalidate(u.submitBtn, nil)
		}
		if !u.loading {
			killTimer(u.hwnd, timerSpin)
		}
	case timerToast:
		u.toastText = ""
		killTimer(u.hwnd, timerToast)
		r := u.toastRect()
		invalidate(u.hwnd, &r)
	case timerSuccess:
		postMessage(u.hwnd, wmClose, 0, 0)
	case timerStatus:
		u.startStatusCheck()
	}
}

func (u *UI) setLoading(loading bool) {
	u.loading = loading
	enableWindow(u.tabLogin, !loading)
	enableWindow(u.tabRegister, !loading)
	enableWindow(u.gear, !loading)
	enableWindow(u.submitBtn, !loading)
	if loading {
		setTimer(u.hwnd, timerSpin, 60)
	} else {
		killTimer(u.hwnd, timerSpin)
		u.spinPhase = 0
	}
	if u.submitBtn != 0 {
		invalidate(u.submitBtn, nil)
	}
	invalidate(u.hwnd, nil)
}

func (u *UI) onSubmit() {
	if u.loading || u.modal {
		return
	}
	user := strings.TrimSpace(getWindowText(u.userEdit))
	pass := getWindowText(u.passEdit)
	if user == "" {
		u.focusField(idUser, u.userEdit)
		u.showToast(2, "请输入账号")
		return
	}
	if pass == "" {
		u.focusField(idPass, u.passEdit)
		u.showToast(2, "请输入密码")
		return
	}

	if u.activeTab == 1 {
		if len(user) < 2 || len(user) > 20 {
			u.focusField(idUser, u.userEdit)
			u.showToast(2, "账号长度需要 2-20 个字符")
			return
		}
		if strings.ContainsAny(user, " \t\r\n") {
			u.focusField(idUser, u.userEdit)
			u.showToast(2, "账号不能包含空格")
			return
		}
		confirm := getWindowText(u.confirmEdit)
		if confirm == "" {
			u.focusField(idConfirm, u.confirmEdit)
			u.showToast(2, "请再次输入密码")
			return
		}
		if pass != confirm {
			u.focusField(idConfirm, u.confirmEdit)
			u.showToast(2, "两次输入的密码不一致")
			return
		}
		u.startRegister(user, pass)
		return
	}

	u.startLogin(user, pass)
}

func (u *UI) focusField(id uint32, h HWND) {
	u.focusID = id
	if h != 0 {
		setFocus(h)
		sendMessage(h, emSetSel, ^uintptr(0), ^uintptr(0))
	}
	invalidate(u.hwnd, nil)
}

func (u *UI) startLogin(user, pass string) {
	u.setLoading(true)
	u.setStatus(0, "正在验证账号…")
	u.focusID = 0

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		res := uiResultLogin{user: user}
		uid, err := u.app.Login(ctx, user, pass)
		res.uid = uid
		res.err = err
		res.bad = err == nil && uid < 0
		if err == nil && uid >= 0 {
			token, terr := generateToken(uid)
			if terr != nil {
				res.err = terr
			} else if lerr := u.app.LaunchGame(token); lerr != nil {
				res.launchErr = lerr
			} else {
				res.launched = true
			}
		}
		select {
		case u.loginCh <- res:
		default:
		}
		postMessage(u.hwnd, wmAppLogin, 0, 0)
	}()
}

func (u *UI) onLoginResult() {
	var res uiResultLogin
	select {
	case res = <-u.loginCh:
	default:
		return
	}
	u.setLoading(false)

	if res.err != nil {
		u.setStatus(2, "登录失败")
		u.showToast(2, res.err.Error())
		return
	}
	if res.bad {
		u.setStatus(2, "账号或密码错误")
		u.showToast(2, "账号或密码错误")
		u.focusField(idPass, u.passEdit)
		return
	}

	if u.remember {
		if err := u.app.store.SetLastUser(res.user); err != nil {
			u.app.logf("save last user: %v", err)
		}
	} else {
		_ = u.app.store.SetLastUser("")
	}

	if res.launchErr != nil {
		u.setStatus(2, "启动游戏失败")
		u.showToast(2, "登录成功，但启动游戏失败："+res.launchErr.Error())
		return
	}

	u.toastText = ""
	u.setStatus(1, "已启动 DNF.exe")
	u.success = true
	u.hideMainControls()
	u.hideSettingsControls()
	setTimer(u.hwnd, timerSuccess, 2600)
	invalidate(u.hwnd, nil)
}

func (u *UI) startRegister(user, pass string) {
	u.setLoading(true)
	u.setStatus(0, "正在创建账号…")
	u.focusID = 0

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		created, err := u.app.Register(ctx, user, pass)
		res := uiResultRegister{user: user, created: created, err: err}
		select {
		case u.registerCh <- res:
		default:
		}
		postMessage(u.hwnd, wmAppRegister, 0, 0)
	}()
}

func (u *UI) onRegisterResult() {
	var res uiResultRegister
	select {
	case res = <-u.registerCh:
	default:
		return
	}
	u.setLoading(false)

	if res.err != nil {
		u.setStatus(2, "注册失败")
		u.showToast(2, res.err.Error())
		return
	}
	if !res.created {
		u.setStatus(2, "该账号已存在")
		u.showToast(2, "该账号已存在")
		u.focusField(idUser, u.userEdit)
		return
	}

	u.setStatus(1, "注册成功，请登录")
	u.showToast(1, "账号 "+res.user+" 已创建，请登录")
	u.setTab(false)
	setWindowText(u.userEdit, res.user)
	setWindowText(u.passEdit, "")
	setWindowText(u.confirmEdit, "")
	u.focusField(idPass, u.passEdit)
}

func (u *UI) startStatusCheck() {
	u.statusSeq++
	seq := u.statusSeq
	u.statusRunning = true
	u.setStatus(0, "正在检测服务器…")

	go func() {
		cfg := u.app.store.Get()
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		latency, err := u.app.TestConnection(ctx, cfg)
		res := uiResultStatus{seq: seq, latency: latency, err: err}
		select {
		case u.statusCh <- res:
		default:
		}
		postMessage(u.hwnd, wmAppStatus, 0, 0)
	}()
}

func (u *UI) onStatusResult() {
	var res uiResultStatus
	select {
	case res = <-u.statusCh:
	default:
		return
	}
	if res.seq != u.statusSeq {
		return
	}
	u.statusRunning = false
	if u.loading {
		return
	}
	if res.err != nil {
		u.setStatus(2, res.err.Error())
		return
	}
	u.setStatus(1, fmt.Sprintf("服务器在线 · %dms", res.latency.Milliseconds()))
}

func (u *UI) startTest() {
	if u.testBusy || !u.modal {
		return
	}
	host := strings.TrimSpace(getWindowText(u.cfgServer))
	portText := strings.TrimSpace(getWindowText(u.cfgPort))
	pass := getWindowText(u.cfgPass)

	if host == "" {
		u.focusField(idCfgServer, u.cfgServer)
		u.showToast(2, "请输入服务器地址")
		return
	}
	port := 3306
	if portText != "" {
		p, err := strconv.Atoi(portText)
		if err != nil || p < 1 || p > 65535 {
			u.focusField(idCfgPort, u.cfgPort)
			u.showToast(2, "端口必须是 1-65535 之间的数字")
			return
		}
		port = p
	}

	cfg := u.app.store.Get()
	cfg.Server = host
	cfg.Port = port
	if pass != "" {
		cfg.DBPassword = pass
	}

	u.testBusy = true
	u.testKind = 0
	u.testText = "正在连接…"
	enableWindow(u.cfgTest, false)
	invalidate(u.hwnd, nil)
	u.app.logf("test connection to %s:%d", cfg.Server, cfg.Port)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		latency, err := u.app.TestConnection(ctx, cfg)
		res := uiResultTest{latency: latency, err: err}
		select {
		case u.testCh <- res:
		default:
		}
		postMessage(u.hwnd, wmAppTest, 0, 0)
	}()
}

func (u *UI) onTestResult() {
	var res uiResultTest
	select {
	case res = <-u.testCh:
	default:
		return
	}
	u.testBusy = false
	enableWindow(u.cfgTest, true)
	if !u.modal {
		return
	}
	if res.err != nil {
		u.testKind = 2
		u.testText = res.err.Error()
	} else {
		u.testKind = 1
		u.testText = fmt.Sprintf("连接成功 · %dms", res.latency.Milliseconds())
	}
	invalidate(u.hwnd, nil)
}

func (u *UI) setTab(register bool) {
	if u.loading || u.modal {
		return
	}
	if register {
		u.activeTab = 1
	} else {
		u.activeTab = 0
	}
	u.focusID = 0
	u.layoutForm()
	u.focusActive()
	if u.submitBtn != 0 {
		invalidate(u.submitBtn, nil)
	}
	if u.tabLogin != 0 {
		invalidate(u.tabLogin, nil)
	}
	if u.tabRegister != 0 {
		invalidate(u.tabRegister, nil)
	}
	invalidate(u.hwnd, nil)
}

func (u *UI) focusActive() {
	if u.activeTab == 0 {
		if strings.TrimSpace(getWindowText(u.userEdit)) == "" {
			u.focusField(idUser, u.userEdit)
		} else {
			u.focusField(idPass, u.passEdit)
		}
		return
	}
	if strings.TrimSpace(getWindowText(u.userEdit)) == "" {
		u.focusField(idUser, u.userEdit)
	} else if getWindowText(u.passEdit) == "" {
		u.focusField(idPass, u.passEdit)
	} else {
		u.focusField(idConfirm, u.confirmEdit)
	}
}

func (u *UI) togglePassword(edit, eye HWND, shown *bool) {
	*shown = !*shown
	if *shown {
		sendMessage(edit, emSetPasswordChar, 0, 0)
	} else {
		sendMessage(edit, emSetPasswordChar, 0x25CF, 0)
	}
	if edit != 0 {
		invalidate(edit, nil)
		setFocus(edit)
	}
	if eye != 0 {
		invalidate(eye, nil)
	}
}

func (u *UI) hideMainControls() {
	for _, h := range []HWND{
		u.tabLogin, u.tabRegister, u.gear, u.userEdit, u.passEdit, u.confirmEdit,
		u.eyePass, u.eyeConfirm, u.rememberBtn, u.forgotBtn, u.toLoginBtn, u.submitBtn,
	} {
		if h != 0 {
			showWindow(h, swHide)
		}
	}
}

func (u *UI) showMainControls() {
	for _, h := range []HWND{
		u.tabLogin, u.tabRegister, u.gear, u.userEdit, u.passEdit,
		u.rememberBtn, u.forgotBtn, u.submitBtn,
	} {
		if h != 0 {
			showWindow(h, swShow)
		}
	}
	u.layoutForm()
}

func (u *UI) hideSettingsControls() {
	for _, h := range []HWND{
		u.cfgServer, u.cfgPort, u.cfgPass, u.cfgGame, u.cfgTest, u.cfgCancel, u.cfgSave,
	} {
		if h != 0 {
			showWindow(h, swHide)
		}
	}
}

func (u *UI) showSettingsControls() {
	for _, h := range []HWND{
		u.cfgServer, u.cfgPort, u.cfgPass, u.cfgGame, u.cfgTest, u.cfgCancel, u.cfgSave,
	} {
		if h != 0 {
			showWindow(h, swShow)
		}
	}
}

func (u *UI) openSettings() {
	if u.loading || u.modal {
		return
	}
	u.modal = true
	u.focusID = 0
	u.testBusy = false
	u.testKind = -1
	u.testText = "尚未测试"

	u.hideMainControls()
	u.showSettingsControls()
	u.layoutSettings()

	cfg := u.app.store.Get()
	setWindowText(u.cfgServer, cfg.Server)
	setWindowText(u.cfgPort, strconv.Itoa(cfg.Port))
	setWindowText(u.cfgPass, "")
	setWindowText(u.cfgGame, cfg.GamePath)

	setWindowLongPtr(u.submitBtn, gwlpID, idSubmit)
	setWindowLongPtr(u.cfgSave, gwlpID, idOK)

	invalidate(u.hwnd, nil)
	u.focusField(idCfgServer, u.cfgServer)
}

func (u *UI) layoutSettings() {
	u.setCtrl(u.cfgServer, 328, 206, 384, 24)
	u.setCtrl(u.cfgPort, 328, 282, 104, 24)
	u.setCtrl(u.cfgPass, 472, 282, 240, 24)
	u.setCtrl(u.cfgGame, 328, 358, 384, 24)
	u.setCtrl(u.cfgTest, 314, 404, 104, 38)
	u.setCtrl(u.cfgCancel, 584, 516, 64, 38)
	u.setCtrl(u.cfgSave, 658, 516, 82, 38)
}

func (u *UI) setCtrl(h HWND, x, y, w, hgt int32) {
	if h == 0 {
		return
	}
	setWindowPos(h, u.S(x), u.S(y), u.S(w), u.S(hgt), swpNoZOrder|swpNoActivate)
	showWindow(h, swShow)
}

func (u *UI) closeSettings() {
	if !u.modal {
		return
	}
	u.modal = false
	u.testBusy = false
	u.testKind = -1
	u.testText = ""
	enableWindow(u.cfgTest, true)

	setWindowLongPtr(u.cfgSave, gwlpID, idCfgSave)
	setWindowLongPtr(u.submitBtn, gwlpID, idOK)

	u.hideSettingsControls()
	u.showMainControls()
	invalidate(u.hwnd, nil)
}

func (u *UI) saveSettings() {
	if !u.modal {
		return
	}
	host := strings.TrimSpace(getWindowText(u.cfgServer))
	portText := strings.TrimSpace(getWindowText(u.cfgPort))
	pass := getWindowText(u.cfgPass)
	game := strings.TrimSpace(getWindowText(u.cfgGame))

	if host == "" {
		u.focusField(idCfgServer, u.cfgServer)
		u.showToast(2, "请输入服务器地址")
		return
	}
	port := 3306
	if portText != "" {
		p, err := strconv.Atoi(portText)
		if err != nil || p < 1 || p > 65535 {
			u.focusField(idCfgPort, u.cfgPort)
			u.showToast(2, "端口必须是 1-65535 之间的数字")
			return
		}
		port = p
	}

	err := u.app.store.Update(Config{
		Server:     host,
		Port:       port,
		DBPassword: pass,
		GamePath:   game,
	})
	if err != nil {
		u.showToast(2, err.Error())
		return
	}

	u.app.logf("settings saved: %s:%d game=%q", host, port, game)
	u.closeSettings()
	u.setStatus(1, "设置已保存")
	u.showToast(1, "服务器设置已更新")
	invalidate(u.hwnd, nil)
	u.startStatusCheck()
}

// ------------------------------ owner draw --------------------------------

func (u *UI) onDrawItem(l uintptr) {
	dis := (*drawItemStruct)(unsafe.Pointer(l))
	switch dis.CtlID {
	case idOK:
		if u.modal {
			u.drawSaveButton(dis)
		} else {
			u.drawSubmitButton(dis)
		}
	case idSubmit:
		u.drawSubmitButton(dis)
	case idCancel:
		u.drawGhostButton(dis, "取消")
	case idCfgTest:
		u.drawGhostButton(dis, "测试连接")
	case idTabLogin:
		u.drawTabButton(dis, "登录", u.activeTab == 0)
	case idTabRegister:
		u.drawTabButton(dis, "注册", u.activeTab == 1)
	case idGear:
		u.drawGearButton(dis)
	case idEyePass:
		u.drawEyeButton(dis, u.passShown)
	case idEyeConfirm:
		u.drawEyeButton(dis, u.confirmShown)
	case idRemember:
		u.drawRememberButton(dis)
	case idForgot:
		u.drawLinkButton(dis, "忘记密码？")
	case idToLogin:
		u.drawLinkButton(dis, "已有账号？")
	default:
		fillRect(dis.HDC, dis.RcItem, u.brush(colCard))
	}
}

func (u *UI) drawSubmitButton(dis *drawItemStruct) {
	r := dis.RcItem
	fillRect(dis.HDC, r, u.brush(colCard))

	if u.loading {
		u.roundedGradient(dis.HDC, r, u.S(14), rgb(206, 170, 92), rgb(150, 95, 44), rgb(110, 82, 40))
		u.drawSpinner(dis.HDC, r)
		return
	}

	top, bottom, border := colGoldSoft, colEmber, rgb(255, 210, 140)
	if dis.ItemState&odsSelected != 0 {
		top, bottom = colGold, rgb(226, 110, 44)
	}
	if dis.ItemState&odsDisabled != 0 {
		top, bottom, border = rgb(160, 140, 90), rgb(120, 84, 44), rgb(100, 74, 36)
	}
	u.roundedGradient(dis.HDC, r, u.S(14), top, bottom, border)
	if dis.ItemState&odsFocus != 0 {
		inner := rect{r.Left + 2, r.Top + 2, r.Right - 2, r.Bottom - 2}
		roundRect(dis.HDC, inner, u.S(12), u.S(12), u.nullBrush(), u.pen(rgb(255, 245, 210), 1))
	}
	text := "进入游戏"
	if u.activeTab == 1 {
		text = "创建账号"
	}
	u.textRect(dis.HDC, text, 16, 700, colBtnText, r, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)
}

func (u *UI) drawSaveButton(dis *drawItemStruct) {
	r := dis.RcItem
	fillRect(dis.HDC, r, u.brush(colCard))
	u.roundedGradient(dis.HDC, r, u.S(10), colGoldSoft, colEmber, rgb(255, 210, 140))
	if dis.ItemState&odsFocus != 0 {
		inner := rect{r.Left + 2, r.Top + 2, r.Right - 2, r.Bottom - 2}
		roundRect(dis.HDC, inner, u.S(8), u.S(8), u.nullBrush(), u.pen(rgb(255, 245, 210), 1))
	}
	u.textRect(dis.HDC, "保存设置", 13, 700, colBtnText, r, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)
}

func (u *UI) drawGhostButton(dis *drawItemStruct, text string) {
	r := dis.RcItem
	fillRect(dis.HDC, r, u.brush(colCard))
	fill := colTabBg
	if dis.ItemState&odsSelected != 0 {
		fill = rgb(28, 32, 46)
	}
	u.rrect(dis.HDC, r, u.S(10), fill, colFieldBorder)
	color := colLink
	if dis.ItemState&odsDisabled != 0 || u.testBusy {
		color = colDim
	}
	u.textRect(dis.HDC, text, 13, 600, color, r, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)
}

func (u *UI) drawTabButton(dis *drawItemStruct, text string, active bool) {
	r := dis.RcItem
	fillRect(dis.HDC, r, u.brush(colTabBg))
	if active {
		u.rrect(dis.HDC, r, u.S(10), colTabActive, rgb(120, 100, 60))
		u.textRect(dis.HDC, text, 14, 600, colGoldSoft, r, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)
		return
	}
	color := colMuted
	if dis.ItemState&odsSelected != 0 {
		color = colText
	}
	u.textRect(dis.HDC, text, 14, 600, color, r, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)
}

func (u *UI) drawGearButton(dis *drawItemStruct) {
	r := dis.RcItem
	fillRect(dis.HDC, r, u.brush(colCard))
	inner := rect{r.Left + 1, r.Top + 1, r.Right - 1, r.Bottom - 1}
	u.rrect(dis.HDC, inner, u.S(11), colTabBg, colFieldBorder)
	u.drawGearIcon(dis.HDC, (r.Left+r.Right)/2, (r.Top+r.Bottom)/2, colMuted)
}

func (u *UI) drawEyeButton(dis *drawItemStruct, shown bool) {
	r := dis.RcItem
	fillRect(dis.HDC, r, u.brush(colField))
	color := colMuted
	if shown {
		color = colGold
	}
	u.drawEyeIcon(dis.HDC, (r.Left+r.Right)/2, (r.Top+r.Bottom)/2, color, shown)
}

func (u *UI) drawRememberButton(dis *drawItemStruct) {
	r := dis.RcItem
	fillRect(dis.HDC, r, u.brush(colCard))
	box := rect{r.Left + u.S(2), (r.Top+r.Bottom)/2 - u.S(8), r.Left + u.S(2) + u.S(16), (r.Top+r.Bottom)/2 + u.S(8)}
	checked := u.remember
	if checked {
		u.rrect(dis.HDC, box, u.S(4), colGold, colGold)
		u.drawCheckIcon(dis.HDC, (box.Left+box.Right)/2, (box.Top+box.Bottom)/2, colBtnText, u.S(4))
	} else {
		u.rrect(dis.HDC, box, u.S(4), rgb(25, 28, 42), colFieldBorder)
	}
	u.textRect(dis.HDC, "记住账号", 12, 400, colMuted,
		rect{r.Left + u.S(25), r.Top, r.Right, r.Bottom}, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
}

func (u *UI) drawLinkButton(dis *drawItemStruct, text string) {
	r := dis.RcItem
	fillRect(dis.HDC, r, u.brush(colCard))
	color := colLink
	if dis.ItemState&odsSelected != 0 {
		color = colGold
	}
	u.textRect(dis.HDC, text, 12, 400, color, r, dtRight|dtVCenter|dtSingleLine|dtNoPrefix)
}

func (u *UI) roundedGradient(dc HDC, r rect, radius int32, top, bottom, border uint32) {
	roundedGradient(dc, r, radius, top, bottom, border)
}

func (u *UI) drawSpinner(dc HDC, r rect) {
	cx := (r.Left + r.Right) / 2
	cy := (r.Top + r.Bottom) / 2
	radius := u.S(9)
	for i := 0; i < 8; i++ {
		a := float64((i+u.spinPhase)%8) * math.Pi / 4
		x := cx + int32(math.Cos(a)*float64(radius))
		y := cy + int32(math.Sin(a)*float64(radius))
		color := rgb(60, 40, 12)
		if i == 0 {
			color = rgb(40, 26, 6)
		} else if i < 3 {
			color = rgb(80, 55, 18)
		}
		ellipse(dc, rect{x - u.S(2), y - u.S(2), x + u.S(2), y + u.S(2)}, u.brush(color), u.nullPen())
	}
}
