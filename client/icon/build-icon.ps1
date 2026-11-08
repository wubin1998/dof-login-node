# Regenerates client/icon.ico and client/rsrc.syso from icon.html.
# Requires Microsoft Edge, Node.js and Go (for the rsrc tool).
$ErrorActionPreference = "Stop"

$edge = @(
    "${env:ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe",
    "$env:ProgramFiles\Microsoft\Edge\Application\msedge.exe"
) | Where-Object { Test-Path $_ } | Select-Object -First 1

if (-not $edge) {
    throw "未找到 Microsoft Edge，无法渲染图标。"
}

$url = ([System.Uri](Join-Path $PSScriptRoot "icon.html")).AbsoluteUri + "?v=" + [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$profile = Join-Path $env:TEMP ("dof-icon-" + [guid]::NewGuid().ToString("N"))
$source = Join-Path $PSScriptRoot "icon-source.png"

# Render at 1024x1024 and downscale: Chromium enforces a minimum window width,
# so screenshotting directly at 16/32/48 px would crop the artwork.
& $edge --headless=new --disable-gpu --hide-scrollbars --default-background-color=00000000 `
    --force-device-scale-factor=1 --window-size=1024,1024 `
    --user-data-dir=$profile --no-first-run --screenshot=$source $url | Out-Null

if (-not (Test-Path $source)) {
    throw "图标渲染失败。"
}

Add-Type -AssemblyName System.Drawing
$src = [System.Drawing.Image]::FromFile($source)
try {
    foreach ($size in 16, 24, 32, 48, 64, 128, 256) {
        $bmp = New-Object System.Drawing.Bitmap($size, $size, [System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
        $g = [System.Drawing.Graphics]::FromImage($bmp)
        try {
            $g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
            $g.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
            $g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
            $g.CompositingQuality = [System.Drawing.Drawing2D.CompositingQuality]::HighQuality
            $g.Clear([System.Drawing.Color]::Transparent)
            $g.DrawImage($src, 0, 0, $size, $size)
        } finally {
            $g.Dispose()
        }
        $bmp.Save((Join-Path $PSScriptRoot "icon-$size.png"), [System.Drawing.Imaging.ImageFormat]::Png)
        $bmp.Dispose()
    }
} finally {
    $src.Dispose()
    Remove-Item -Recurse -Force $profile -ErrorAction SilentlyContinue
}

node (Join-Path $PSScriptRoot "make-ico.js")

$goCommand = Get-Command go -ErrorAction SilentlyContinue
$goExe = if ($goCommand) { $goCommand.Source } else { $null }
if (-not $goExe) {
    $candidate = Join-Path $env:ProgramFiles "Go\bin\go.exe"
    if (Test-Path $candidate) { $goExe = $candidate }
}
if (-not $goExe) {
    throw "未找到 Go，无法生成 rsrc.syso。"
}

$rsrc = Join-Path $env:USERPROFILE "go\bin\rsrc.exe"
if (-not (Test-Path $rsrc)) {
    & $goExe install github.com/akavel/rsrc@latest
}

& $rsrc -ico (Join-Path $PSScriptRoot "..\icon.ico") `
    -manifest (Join-Path $PSScriptRoot "..\app.manifest") `
    -o (Join-Path $PSScriptRoot "..\rsrc.syso") `
    -arch amd64

Write-Host "图标已更新：client\icon.ico + client\rsrc.syso" -ForegroundColor Green
