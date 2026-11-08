# dof-login-node（旧版）

> 这是最初的 Node.js 命令行版本，仅保留作参考。
> 新的桌面客户端：
> - `client-native/`：原生 Win32 版，Windows 7 零依赖（推荐）
> - `client/`：WebView2 网页版（Win10/11 更精致的界面，Win7 需要 WebView2 109）

```bash
npm install
npm run build
```

旧版通过 `pkg` 打包 Node 18，生成的文件通常在 45 MB 以上。
