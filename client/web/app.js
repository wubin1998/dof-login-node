(function () {
  "use strict";

  const TOKEN = window.__DOF_TOKEN__ || "";

  const $ = (sel) => document.querySelector(sel);

  const el = {
    tabs: $("#tabs"),
    loginForm: $("#loginForm"),
    registerForm: $("#registerForm"),
    loginUser: $("#loginUser"),
    loginPass: $("#loginPass"),
    remember: $("#remember"),
    loginBtn: $("#loginBtn"),
    registerBtn: $("#registerBtn"),
    regUser: $("#regUser"),
    regPass: $("#regPass"),
    regPass2: $("#regPass2"),
    card: $("#card"),
    successVeil: $("#successVeil"),
    toasts: $("#toasts"),
    serverChip: $("#serverChip"),
    brandStatus: $("#brandStatus"),
    cardStatus: $("#cardStatus"),
    cfgMask: $("#cfgMask"),
    cfgServer: $("#cfgServer"),
    cfgPort: $("#cfgPort"),
    cfgPass: $("#cfgPass"),
    cfgGame: $("#cfgGame"),
    testResult: $("#testResult"),
  };

  const state = {
    config: { server: "", port: 3306, gamePath: "", hasCustom: false },
    loginBusy: false,
    registerBusy: false,
    cfgBusy: false,
    finished: false,
  };

  /* ---------------------------------- api --------------------------------- */

  async function api(path, body, method) {
    const opts = {
      method: method || (body === undefined ? "GET" : "POST"),
      headers: { "X-DOF-Token": TOKEN },
      cache: "no-store",
    };
    if (opts.method === "POST") {
      opts.headers["Content-Type"] = "application/json";
      opts.body = JSON.stringify(body || {});
    }
    const res = await fetch(path, opts);
    if (!res.ok) {
      const msg = res.status === 403 ? "本地安全校验失败，请重启登录器" : "请求失败 (" + res.status + ")";
      throw new Error(msg);
    }
    return res.json();
  }

  /* -------------------------------- feedback ------------------------------ */

  const ICONS = {
    success:
      '<svg viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.6"/><path d="m8.2 12.4 2.6 2.6 5.2-5.4" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/></svg>',
    error:
      '<svg viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.6"/><path d="m9 9 6 6M15 9l-6 6" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/></svg>',
    info:
      '<svg viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.6"/><path d="M12 11v5.2M12 7.8v.2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>',
  };

  function toast(message, type, title) {
    const kind = type === "success" || type === "error" ? type : "info";
    while (el.toasts.children.length >= 3) {
      el.toasts.removeChild(el.toasts.firstElementChild);
    }

    const node = document.createElement("div");
    node.className = "toast " + kind;

    const icon = document.createElement("span");
    icon.className = "t-icon";
    icon.innerHTML = ICONS[kind];

    const body = document.createElement("div");
    body.className = "t-body";
    if (title) {
      const head = document.createElement("div");
      head.className = "t-title";
      head.textContent = title;
      body.appendChild(head);
    }
    const text = document.createElement("div");
    text.textContent = message;
    body.appendChild(text);

    node.appendChild(icon);
    node.appendChild(body);
    el.toasts.appendChild(node);

    const close = () => {
      if (node.classList.contains("out")) return;
      node.classList.add("out");
      setTimeout(() => node.remove(), 280);
    };
    const timer = setTimeout(close, 3800);
    node.addEventListener("click", () => {
      clearTimeout(timer);
      close();
    });
  }

  function setNet(node, kind, text) {
    if (!node) return;
    node.classList.remove("busy", "ok", "err");
    if (kind) node.classList.add(kind);
    const span = node.querySelector("span");
    if (span) {
      span.textContent = text;
      span.title = text;
    }
  }

  function setStatus(kind, text) {
    setNet(el.cardStatus, kind, text);
    setNet(el.brandStatus, kind, text);
  }

  function setLoading(button, loading) {
    button.classList.toggle("loading", loading);
    button.disabled = loading;
  }

  function markInvalid(field) {
    if (!field) return;
    const box = field.closest(".field");
    if (!box) return;
    box.classList.remove("invalid");
    // restart the shake animation
    void box.offsetWidth;
    box.classList.add("invalid");
    const input = box.querySelector("input");
    if (input) input.focus();
  }

  function shakeCard() {
    if (!el.card.animate) return;
    el.card.animate(
      [
        { transform: "translateX(0)" },
        { transform: "translateX(-6px)" },
        { transform: "translateX(6px)" },
        { transform: "translateX(-4px)" },
        { transform: "translateX(4px)" },
        { transform: "translateX(0)" },
      ],
      { duration: 380, easing: "ease-in-out" }
    );
  }

  /* --------------------------------- tabs --------------------------------- */

  function setTab(name) {
    el.tabs.dataset.active = name;
    el.tabs.querySelectorAll(".tab").forEach((tab) => {
      tab.classList.toggle("active", tab.dataset.tab === name);
    });
    el.loginForm.classList.toggle("active", name === "login");
    el.registerForm.classList.toggle("active", name === "register");
    setTimeout(() => {
      if (name === "login") el.loginUser.focus();
      else el.regUser.focus();
    }, 140);
  }

  el.tabs.addEventListener("click", (event) => {
    const tab = event.target.closest(".tab");
    if (tab) setTab(tab.dataset.tab);
  });

  $("#toLogin").addEventListener("click", () => setTab("login"));

  /* --------------------------------- eyes --------------------------------- */

  document.querySelectorAll("[data-eye]").forEach((button) => {
    button.addEventListener("click", () => {
      const input = button.parentElement.querySelector("input");
      const show = input.type === "password";
      input.type = show ? "text" : "password";
      button.classList.toggle("show", show);
      button.setAttribute("aria-label", show ? "隐藏密码" : "显示密码");
      input.focus();
    });
  });

  document.querySelectorAll(".field input").forEach((input) => {
    input.addEventListener("input", () => {
      const box = input.closest(".field");
      if (box) box.classList.remove("invalid");
    });
  });

  el.loginUser.addEventListener("keydown", (event) => {
    if (event.key === "Enter" && !el.loginPass.value) {
      event.preventDefault();
      el.loginPass.focus();
    }
  });

  /* --------------------------------- login -------------------------------- */

  function rememberUser(user) {
    try {
      if (el.remember.checked) localStorage.setItem("dof.user", user);
      else localStorage.removeItem("dof.user");
    } catch (_) {
      /* storage can be disabled; ignore */
    }
  }

  el.loginForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (state.loginBusy || state.finished) return;

    const user = el.loginUser.value.trim();
    const pass = el.loginPass.value;

    if (!user) {
      markInvalid(el.loginUser);
      toast("请输入账号", "error");
      return;
    }
    if (!pass) {
      markInvalid(el.loginPass);
      toast("请输入密码", "error");
      return;
    }

    state.loginBusy = true;
    setLoading(el.loginBtn, true);
    setStatus("busy", "正在验证账号…");

    try {
      const res = await api("/api/login", { username: user, password: pass });

      if (!res.ok) {
        setStatus("err", res.message || "登录失败");
        toast(res.message || "登录失败", "error");
        if (res.code === "credentials") {
          markInvalid(el.loginPass);
          shakeCard();
        }
        setLoading(el.loginBtn, false);
        state.loginBusy = false;
        return;
      }

      rememberUser(user);

      if (!res.launched) {
        setStatus("err", "登录成功，但启动游戏失败");
        toast(res.message || "启动游戏失败", "error", "已登录");
        setLoading(el.loginBtn, false);
        state.loginBusy = false;
        return;
      }

      state.finished = true;
      setStatus("ok", "已启动 DNF.exe");
      el.successVeil.hidden = false;

      // The launcher has done its job. Closing it also avoids stray processes.
      setTimeout(() => {
        api("/api/exit", {}).catch(() => {});
      }, 2400);
    } catch (err) {
      setStatus("err", err.message || "登录失败");
      toast(err.message || "登录失败，请重试", "error");
      setLoading(el.loginBtn, false);
      state.loginBusy = false;
    }
  });

  /* -------------------------------- register ------------------------------ */

  el.registerForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (state.registerBusy) return;

    const user = el.regUser.value.trim();
    const pass = el.regPass.value;
    const pass2 = el.regPass2.value;

    if (!user) {
      markInvalid(el.regUser);
      toast("请输入账号", "error");
      return;
    }
    if (user.length < 2 || user.length > 20) {
      markInvalid(el.regUser);
      toast("账号长度需要 2-20 个字符", "error");
      return;
    }
    if (/\s/.test(user)) {
      markInvalid(el.regUser);
      toast("账号不能包含空格", "error");
      return;
    }
    if (!pass) {
      markInvalid(el.regPass);
      toast("请设置密码", "error");
      return;
    }
    if (pass !== pass2) {
      markInvalid(el.regPass2);
      toast("两次输入的密码不一致", "error");
      return;
    }

    state.registerBusy = true;
    setLoading(el.registerBtn, true);
    setStatus("busy", "正在创建账号…");

    try {
      const res = await api("/api/register", { username: user, password: pass });
      if (!res.ok) {
        setStatus("err", res.message || "注册失败");
        toast(res.message || "注册失败", "error");
        if (res.code === "exists") markInvalid(el.regUser);
        setLoading(el.registerBtn, false);
        state.registerBusy = false;
        return;
      }

      toast("账号 " + user + " 已创建，请登录", "success", "注册成功");
      setStatus("ok", "注册成功，请登录");
      el.regUser.value = "";
      el.regPass.value = "";
      el.regPass2.value = "";
      el.loginUser.value = user;
      el.loginPass.value = "";
      el.remember.checked = true;
      setTab("login");
      setTimeout(() => el.loginPass.focus(), 200);
    } catch (err) {
      setStatus("err", err.message || "注册失败");
      toast(err.message || "注册失败，请重试", "error");
    } finally {
      setLoading(el.registerBtn, false);
      state.registerBusy = false;
    }
  });

  /* --------------------------------- status ------------------------------- */

  let statusTimer = 0;

  async function refreshStatus() {
    setStatus("busy", "正在检测服务器…");
    try {
      const res = await api("/api/status", {});
      if (res.online) {
        setStatus("ok", "服务器在线" + (typeof res.latencyMs === "number" ? " · " + res.latencyMs + "ms" : ""));
      } else {
        setStatus("err", res.message || "服务器离线");
      }
    } catch (_) {
      setStatus("err", "无法检测服务器状态");
    }
    clearTimeout(statusTimer);
    statusTimer = setTimeout(refreshStatus, 30000);
  }

  /* -------------------------------- settings ------------------------------ */

  function fillConfigForm() {
    el.cfgServer.value = state.config.server || "";
    el.cfgPort.value = state.config.port || 3306;
    el.cfgPass.value = "";
    el.cfgPass.placeholder = state.config.hasCustom ? "已设置自定义密码（留空不变）" : "默认密码（留空不变）";
    el.cfgGame.value = state.config.gamePath || "";
    el.testResult.textContent = "尚未测试";
    el.testResult.className = "test-result";
  }

  async function loadConfig() {
    const res = await api("/api/config");
    if (!res.ok) throw new Error(res.message || "读取设置失败");
    state.config = res;
    el.serverChip.textContent = res.server + ":" + res.port;
    fillConfigForm();
  }

  function openSettings() {
    fillConfigForm();
    el.cfgMask.hidden = false;
    setTimeout(() => el.cfgServer.focus(), 120);
  }

  function closeSettings() {
    el.cfgMask.hidden = true;
  }

  $("#settingsBtn").addEventListener("click", openSettings);
  $("#cfgClose").addEventListener("click", closeSettings);
  $("#cfgCancel").addEventListener("click", closeSettings);

  el.cfgMask.addEventListener("click", (event) => {
    if (event.target === el.cfgMask) closeSettings();
  });

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && !el.cfgMask.hidden) closeSettings();
  });

  $("#testBtn").addEventListener("click", async () => {
    if (state.cfgBusy) return;
    state.cfgBusy = true;
    const button = $("#testBtn");
    button.disabled = true;
    el.testResult.className = "test-result";
    el.testResult.textContent = "正在连接…";

    try {
      const res = await api("/api/test", {
        server: el.cfgServer.value.trim(),
        port: parseInt(el.cfgPort.value, 10) || 0,
        password: el.cfgPass.value,
      });
      if (res.online) {
        el.testResult.className = "test-result ok";
        el.testResult.textContent = "连接成功" + (typeof res.latencyMs === "number" ? " · " + res.latencyMs + "ms" : "");
      } else {
        el.testResult.className = "test-result err";
        el.testResult.textContent = res.message || "连接失败";
      }
    } catch (err) {
      el.testResult.className = "test-result err";
      el.testResult.textContent = err.message || "连接失败";
    } finally {
      button.disabled = false;
      state.cfgBusy = false;
    }
  });

  $("#cfgSave").addEventListener("click", async () => {
    if (state.cfgBusy) return;

    const server = el.cfgServer.value.trim();
    const port = parseInt(el.cfgPort.value, 10);
    if (!server) {
      el.cfgServer.focus();
      toast("请输入服务器地址", "error");
      return;
    }
    if (!port || port < 1 || port > 65535) {
      el.cfgPort.focus();
      toast("端口必须是 1-65535 之间的数字", "error");
      return;
    }

    state.cfgBusy = true;
    const button = $("#cfgSave");
    button.disabled = true;

    try {
      const res = await api("/api/config", {
        server: server,
        port: port,
        password: el.cfgPass.value,
        gamePath: el.cfgGame.value.trim(),
      });
      if (!res.ok) {
        toast(res.message || "保存失败", "error");
        return;
      }
      state.config = {
        server: res.server,
        port: res.port,
        gamePath: res.gamePath,
        hasCustom: state.config.hasCustom || !!el.cfgPass.value,
      };
      el.serverChip.textContent = res.server + ":" + res.port;
      closeSettings();
      toast("服务器设置已更新", "success");
      refreshStatus();
    } catch (err) {
      toast(err.message || "保存失败", "error");
    } finally {
      button.disabled = false;
      state.cfgBusy = false;
    }
  });

  /* -------------------------------- password ------------------------------ */

  $("#forgotBtn").addEventListener("click", () => {
    toast("请联系服务器管理员协助重置密码", "info", "忘记密码");
  });

  /* -------------------------------- startup ------------------------------- */

  try {
    const saved = localStorage.getItem("dof.user");
    if (saved) {
      el.loginUser.value = saved;
      el.remember.checked = true;
    }
  } catch (_) {
    /* ignore */
  }

  loadConfig()
    .catch((err) => toast(err.message || "读取设置失败", "error"))
    .finally(() => {
      refreshStatus();
      setTimeout(() => {
        if (el.loginPass.value === "" && el.loginUser.value) el.loginPass.focus();
        else if (!el.loginUser.value) el.loginUser.focus();
      }, 300);
    });
})();
