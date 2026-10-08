# Builds a dev copy of the app and starts it with local DevTools ports (chosen automatically per window),
# so tools/smoke.js can drive it.
# The app must not already be running (it shares the same WebView2 profile). Quit it from the tray first.
#
#   powershell -ExecutionPolicy Bypass -File tools\dev.ps1            # profile 1 (your real session)
#   powershell -ExecutionPolicy Bypass -File tools\dev.ps1 -Profile 3 # throwaway profile, shows the QR screen
param(
    [string]$Profile = "1",
    [int]$WaitSeconds = 18
)

$root = Split-Path -Parent $PSScriptRoot
$exe = Join-Path $env:TEMP "WhatsApp-dev.exe"

Get-Process WhatsApp-dev -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 1

Push-Location $root
go build -ldflags="-H windowsgui -s -w" -o $exe .
$ok = ($LASTEXITCODE -eq 0)
Pop-Location
if (-not $ok) { Write-Output "BUILD FAILED"; exit 1 }

$env:WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS = "--remote-debugging-port=0 --remote-allow-origins=*"
if ($Profile -eq "1") { Start-Process -FilePath $exe -WorkingDirectory $env:TEMP }
else { Start-Process -FilePath $exe -ArgumentList "--profile", $Profile -WorkingDirectory $env:TEMP }

Start-Sleep -Seconds $WaitSeconds
Write-Output "started: $exe"
foreach ($dir in "UserData", "UserData_Telegram") {
    $f = Join-Path $env:APPDATA "WhatsAppDesktopLight\$dir\EBWebView\DevToolsActivePort"
    if (Test-Path $f) { Write-Output ("{0}: DevTools port {1}" -f $dir, (Get-Content $f -TotalCount 1)) }
}
