#requires -Version 5.1
[CmdletBinding()]
param(
    # Compress the finished exe with UPX (must be installed).
    [switch]$Upx,

    # Use a local Go 1.20.x instead of the cached/downloaded toolchain.
    [string]$GoExe = ""
)

$ErrorActionPreference = "Stop"
try { [Console]::OutputEncoding = [System.Text.Encoding]::UTF8 } catch { }

$clientDir = $PSScriptRoot
$rootDir = Split-Path -Parent $clientDir
$distDir = Join-Path $rootDir "dist"
$outputExe = Join-Path $distDir "DOF登录器.exe"

# Go 1.20 is the last Go release that produces binaries for Windows 7 / 8.
# The resulting exe runs on Windows 7 through Windows 11.
$goVersion = "go1.20.14"
$goSHA256 = "0e0d0190406ead891d94ecf00f961bb5cfa15ddd47499d2649f12eee80aee110"
$toolchainDir = Join-Path $rootDir ".toolchain"
$toolchainRoot = Join-Path $toolchainDir $goVersion
$toolchainGo = Join-Path $toolchainRoot "go\bin\go.exe"

function Test-Go120([string]$path) {
    if (-not (Test-Path $path)) { return $false }
    $version = & $path version 2>$null
    return ($version -match "^go version go1\.20\.")
}

if ($GoExe) {
    if (-not (Test-Path $GoExe)) {
        throw "找不到指定的 Go：$GoExe"
    }
    if (-not (Test-Go120 $GoExe)) {
        throw "需要 Go 1.20.x（最后一个支持 Windows 7 的 Go 版本）。"
    }
    $goCommand = (Resolve-Path $GoExe).Path
} elseif (Test-Go120 $toolchainGo) {
    $goCommand = $toolchainGo
} else {
    New-Item -ItemType Directory -Force -Path $toolchainDir | Out-Null
    $zipPath = Join-Path $toolchainDir "$goVersion.windows-amd64.zip"

    if (-not (Test-Path $zipPath)) {
        Write-Host "==> 首次构建：下载 Go $goVersion（约 110 MB，仅需一次）..." -ForegroundColor Cyan
        Invoke-WebRequest -UseBasicParsing -Uri "https://go.dev/dl/$goVersion.windows-amd64.zip" -OutFile $zipPath
    }

    $hash = (Get-FileHash $zipPath -Algorithm SHA256).Hash.ToLower()
    if ($hash -ne $goSHA256) {
        Remove-Item $zipPath -Force -ErrorAction SilentlyContinue
        throw "Go 工具链校验失败，请删除 .toolchain 后重试。"
    }

    Write-Host "==> 解压 Go $goVersion..." -ForegroundColor Cyan
    if (Test-Path $toolchainRoot) {
        Remove-Item -Recurse -Force $toolchainRoot
    }
    New-Item -ItemType Directory -Force -Path $toolchainRoot | Out-Null
    Expand-Archive -Path $zipPath -DestinationPath $toolchainRoot -Force

    if (-not (Test-Go120 $toolchainGo)) {
        throw "Go 工具链解压失败，请删除 .toolchain 后重试。"
    }
    $goCommand = $toolchainGo
}

$running = Get-Process -Name "DOF登录器" -ErrorAction SilentlyContinue
if ($running) {
    throw "请先关闭正在运行的 DOF登录器.exe，再重新构建。"
}

New-Item -ItemType Directory -Force -Path $distDir | Out-Null
$env:GOTOOLCHAIN = "local"

Push-Location $clientDir
try {
    Write-Host "==> go test ./..." -ForegroundColor Cyan
    & $goCommand test ./...
    if ($LASTEXITCODE -ne 0) {
        throw "测试失败，已停止构建。"
    }

    Write-Host "==> go build（Windows 7 兼容，Go $goVersion）" -ForegroundColor Cyan
    & $goCommand build -trimpath -ldflags="-s -w -H=windowsgui" -o $outputExe .
    if ($LASTEXITCODE -ne 0) {
        throw "编译失败。"
    }

    $size = (Get-Item $outputExe).Length
    Write-Host ("完成：{0}  ({1:N2} MB)" -f $outputExe, ($size / 1MB)) -ForegroundColor Green

    if ($Upx) {
        $upxCommand = Get-Command upx -ErrorAction SilentlyContinue
        if (-not $upxCommand) {
            throw "未找到 UPX。可先安装 UPX，或去掉 -Upx 参数。"
        }
        Write-Host "==> upx --best --lzma" -ForegroundColor Cyan
        $upxOutput = Join-Path $distDir "DOF登录器-UPX.exe"
        Copy-Item $outputExe $upxOutput -Force
        & $upxCommand.Source --best --lzma $upxOutput
        if ($LASTEXITCODE -ne 0) {
            throw "UPX 压缩失败。"
        }
        $size = (Get-Item $upxOutput).Length
        Write-Host ("UPX 完成：{0}  ({1:N2} MB)" -f $upxOutput, ($size / 1MB)) -ForegroundColor Green
    }
} finally {
    Pop-Location
}
