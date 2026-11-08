package main

import "runtime"

func main() {
	// Win32 window creation, the message loop and all WndProc callbacks must
	// stay on the same OS thread.
	runtime.LockOSThread()
	preloadWin32()

	// Windows 7 needs this to render at the real system DPI.
	_, _, _ = pSetProcessDPIAware.Call()

	app, err := NewApp()
	if err != nil {
		messageBox("DOF 登录器", "初始化失败：\n"+err.Error())
		return
	}
	defer app.Close()
	app.logf("DOF 登录器启动（原生界面）")

	ui := NewUI(app)
	ui.run()
	app.logf("DOF 登录器退出")
}
