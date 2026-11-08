package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// App owns the settings, the log file and the game process.
type App struct {
	store   *ConfigStore
	dataDir string
	exeDir  string

	logger  *log.Logger
	logFile *os.File
}

func NewApp() (*App, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dataDir := filepath.Join(base, "DOFLogin")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}

	store, err := newConfigStore(dataDir)
	if err != nil {
		return nil, err
	}

	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	app := &App{
		store:   store,
		dataDir: dataDir,
		exeDir:  filepath.Dir(exe),
	}
	app.openLog()
	return app, nil
}

func (a *App) openLog() {
	f, err := os.OpenFile(filepath.Join(a.dataDir, "launcher.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		a.logger = log.New(io.Discard, "", 0)
		return
	}
	a.logFile = f
	a.logger = log.New(f, "", log.LstdFlags)
}

func (a *App) logf(format string, args ...any) {
	if a.logger != nil {
		a.logger.Printf(format, args...)
	}
}

func (a *App) Close() {
	if a.logFile != nil {
		_ = a.logFile.Close()
	}
}

func (a *App) gameExecutable(cfg Config) string {
	if cfg.GamePath != "" {
		return cfg.GamePath
	}
	return filepath.Join(a.exeDir, "DNF.exe")
}

func (a *App) LaunchGame(token string) error {
	cfg := a.store.Get()
	path := a.gameExecutable(cfg)

	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return errors.New("未找到 DNF.exe，请把登录器放到游戏目录，或在设置中指定路径")
	}

	cmd := exec.Command(path, token)
	cmd.Dir = filepath.Dir(path)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Start(); err != nil {
		a.logf("launch %s failed: %v", path, err)
		return fmt.Errorf("启动 DNF.exe 失败：%v", err)
	}
	a.logf("launched %s (pid %d)", path, cmd.Process.Pid)

	go func() {
		if err := cmd.Wait(); err != nil {
			a.logf("DNF.exe exited: %v", err)
		}
	}()
	return nil
}
