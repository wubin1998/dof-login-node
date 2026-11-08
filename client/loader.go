package main

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// WebView2 SDK 1.0.1518.46 is the last SDK release that supports Windows 7
// and Windows 8/8.1. Its loader is embedded so the single-file exe also works
// on those systems, where the Evergreen runtime is pinned to version 109.
//
//go:embed third_party/WebView2Loader-x64.dll
var webView2Loader109 []byte

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procSetDllDirectoryW = kernel32.NewProc("SetDllDirectoryW")
	procLoadLibraryW     = kernel32.NewProc("LoadLibraryW")
)

// prepareWebView2Loader drops the Windows 7-compatible loader into the local
// cache and loads it before go-webview2 starts. go-webview2's native loader
// lookup (LoadLibrary("WebView2Loader")) then reuses the already loaded DLL.
func (a *App) prepareWebView2Loader() {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}

	dir := filepath.Join(base, "DOFLogin", "WebView2Loader109-v1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		a.logf("create WebView2Loader dir: %v", err)
		return
	}

	dllPath := filepath.Join(dir, "WebView2Loader.dll")
	cached, err := os.ReadFile(dllPath)
	if err != nil || !bytes.Equal(cached, webView2Loader109) {
		if err := os.WriteFile(dllPath, webView2Loader109, 0o644); err != nil {
			a.logf("write WebView2Loader: %v", err)
			return
		}
	}

	dirPtr, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		a.logf("encode loader dir: %v", err)
		return
	}
	dllPtr, err := syscall.UTF16PtrFromString(dllPath)
	if err != nil {
		a.logf("encode loader path: %v", err)
		return
	}

	_, _, _ = procSetDllDirectoryW.Call(uintptr(unsafe.Pointer(dirPtr)))
	if h, _, callErr := procLoadLibraryW.Call(uintptr(unsafe.Pointer(dllPtr))); h == 0 {
		a.logf("load WebView2Loader 109 failed: %v", callErr)
		return
	}
	a.logf("loaded WebView2Loader 109 from %s", dllPath)
}
