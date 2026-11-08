package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"
)

func (a *App) StartServer(web fs.FS) (string, error) {
	index, err := fs.ReadFile(web, "index.html")
	if err != nil {
		return "", err
	}
	if !strings.Contains(string(index), `"__DOF_TOKEN__"`) {
		return "", errors.New("index.html 缺少令牌占位符")
	}
	indexHTML := strings.Replace(string(index), `"__DOF_TOKEN__"`, `"`+a.token+`"`, 1)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = io.WriteString(w, indexHTML)
	})

	static := http.FileServer(http.FS(web))
	cache := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			next.ServeHTTP(w, r)
		})
	}
	mux.Handle("/style.css", cache(static))
	mux.Handle("/app.js", cache(static))

	mux.HandleFunc("/api/status", a.api(a.handleStatus))
	mux.HandleFunc("/api/config", a.api(a.handleConfig))
	mux.HandleFunc("/api/test", a.api(a.handleTest))
	mux.HandleFunc("/api/login", a.api(a.handleLogin))
	mux.HandleFunc("/api/register", a.api(a.handleRegister))
	mux.HandleFunc("/api/exit", a.api(a.handleExit))

	// A stable port lets WebView2 keep localStorage (the remembered account)
	// between launches. Fall back to a random port if it is already taken.
	ln, err := net.Listen("tcp", "127.0.0.1:27183")
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return "", err
		}
	}

	srv := &http.Server{
		Handler:           localOnly(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		_ = srv.Serve(ln)
	}()

	a.logf("local ui listening on %s", ln.Addr())
	return "http://" + ln.Addr().String() + "/", nil
}

// localOnly rejects requests whose Host is not the loopback address. This
// blocks DNS-rebinding pages from reading the token out of the UI document.
func localOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Host, "127.0.0.1:") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// api rejects anything that did not come from our own page (random token),
// which keeps other local processes and websites out of the DB endpoints.
func (a *App) api(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		got := r.Header.Get("X-DOF-Token")
		if subtle.ConstantTimeCompare([]byte(got), []byte(a.token)) != 1 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	_ = json.NewEncoder(w).Encode(v)
}

func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeJSON(w, map[string]any{"ok": false, "code": "bad-request", "message": "请求格式错误"})
		return false
	}
	return true
}

func fail(w http.ResponseWriter, code, message string) {
	writeJSON(w, map[string]any{"ok": false, "code": code, "message": message})
}

func (a *App) handleStatus(w http.ResponseWriter, r *http.Request) {
	cfg := a.store.Get()
	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	db, err := a.openDB(ctx, cfg)
	if err != nil {
		writeJSON(w, map[string]any{"ok": true, "online": false, "message": err.Error()})
		return
	}
	_ = db.Close()
	writeJSON(w, map[string]any{
		"ok":        true,
		"online":    true,
		"latencyMs": time.Since(start).Milliseconds(),
	})
}

type configRequest struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	GamePath string `json:"gamePath"`
}

func (a *App) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		cfg := a.store.Get()
		writeJSON(w, map[string]any{
			"ok":        true,
			"server":    cfg.Server,
			"port":      cfg.Port,
			"gamePath":  cfg.GamePath,
			"hasCustom": cfg.DBPassword != defaultConfig().DBPassword,
		})
		return
	}

	var req configRequest
	if !decodeBody(w, r, &req) {
		return
	}
	err := a.store.Update(Config{
		Server:     req.Server,
		Port:       req.Port,
		DBPassword: req.Password,
		GamePath:   req.GamePath,
	})
	if err != nil {
		fail(w, "bad-config", err.Error())
		return
	}
	a.logf("config updated: %s:%d game=%q", req.Server, req.Port, req.GamePath)
	cfg := a.store.Get()
	writeJSON(w, map[string]any{
		"ok":       true,
		"server":   cfg.Server,
		"port":     cfg.Port,
		"gamePath": cfg.GamePath,
		"message":  "设置已保存",
	})
}

type testRequest struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Password string `json:"password"`
}

func (a *App) handleTest(w http.ResponseWriter, r *http.Request) {
	var req testRequest
	if !decodeBody(w, r, &req) {
		return
	}
	cfg := a.store.Get()
	if strings.TrimSpace(req.Server) != "" {
		cfg.Server = strings.TrimSpace(req.Server)
	}
	if req.Port > 0 && req.Port <= 65535 {
		cfg.Port = req.Port
	}
	if req.Password != "" {
		cfg.DBPassword = req.Password
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	db, err := a.openDB(ctx, cfg)
	if err != nil {
		writeJSON(w, map[string]any{"ok": true, "online": false, "message": err.Error()})
		return
	}
	_ = db.Close()
	writeJSON(w, map[string]any{
		"ok":        true,
		"online":    true,
		"latencyMs": time.Since(start).Milliseconds(),
		"message":   "连接成功",
	})
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeBody(w, r, &req) {
		return
	}
	user := strings.TrimSpace(req.Username)
	if user == "" {
		fail(w, "empty-user", "请输入账号")
		return
	}
	if req.Password == "" {
		fail(w, "empty-password", "请输入密码")
		return
	}
	if len(user) > 32 || len(req.Password) > 64 {
		fail(w, "too-long", "账号或密码过长")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	uid, err := a.Login(ctx, user, req.Password)
	if err != nil {
		a.logf("login %q failed: %v", user, err)
		fail(w, "db", err.Error())
		return
	}
	if uid < 0 {
		writeJSON(w, map[string]any{"ok": false, "code": "credentials", "message": "账号或密码错误"})
		return
	}

	token, err := generateToken(uid)
	if err != nil {
		a.logf("token generation failed: %v", err)
		fail(w, "token", "登录凭证生成失败，请重启登录器")
		return
	}

	if err := a.LaunchGame(token); err != nil {
		a.logf("login %q uid=%d launch failed: %v", user, uid, err)
		writeJSON(w, map[string]any{
			"ok":       true,
			"uid":      uid,
			"launched": false,
			"message":  "登录成功，但启动游戏失败：" + err.Error(),
		})
		return
	}

	a.logf("login %q uid=%d ok", user, uid)
	writeJSON(w, map[string]any{
		"ok":       true,
		"uid":      uid,
		"launched": true,
		"message":  "登录成功，正在启动游戏",
	})
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !decodeBody(w, r, &req) {
		return
	}
	user := strings.TrimSpace(req.Username)
	if user == "" {
		fail(w, "empty-user", "请输入账号")
		return
	}
	if len(user) < 2 || len(user) > 20 {
		fail(w, "bad-user", "账号长度需要 2-20 个字符")
		return
	}
	if strings.ContainsAny(user, " \t\r\n") {
		fail(w, "bad-user", "账号不能包含空格")
		return
	}
	if len(req.Password) < 1 || len(req.Password) > 64 {
		fail(w, "bad-password", "密码长度需要 1-64 个字符")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	created, err := a.Register(ctx, user, req.Password)
	if err != nil {
		a.logf("register %q failed: %v", user, err)
		fail(w, "db", err.Error())
		return
	}
	if !created {
		writeJSON(w, map[string]any{"ok": false, "code": "exists", "message": "该账号已存在"})
		return
	}

	a.logf("register %q ok", user)
	writeJSON(w, map[string]any{"ok": true, "message": "注册成功，请登录"})
}

func (a *App) handleExit(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"ok": true})
	go func() {
		time.Sleep(150 * time.Millisecond)
		a.Quit()
	}()
}
