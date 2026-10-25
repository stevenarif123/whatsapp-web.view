# Builds a release for the in-app updater: WhatsApp.exe with the version embedded and WhatsApp.exe.sha256.
#
#   powershell -ExecutionPolicy Bypass -File tools\release.ps1 -Version 1.2.0
#
# Then create a GitHub release with tag v1.2.0 (not a draft or pre-release) and upload both files.
# Running apps offer "Perbarui sekarang" and only install the exe if its SHA-256 matches the .sha256 file.
param(
    [Parameter(Mandatory = $true)][string]$Version
)

if ($Version -notmatch '^\d+\.\d+\.\d+$') { throw "Versi harus berbentuk 1.2.0" }

$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
    go test .
    if ($LASTEXITCODE -ne 0) { throw "test gagal, rilis dibatalkan" }

    go build -ldflags="-H windowsgui -s -w -X main.appVersion=$Version" -o WhatsApp.exe .
    if ($LASTEXITCODE -ne 0) { throw "build gagal" }

    $hash = (Get-FileHash WhatsApp.exe -Algorithm SHA256).Hash.ToLower()
    [System.IO.File]::WriteAllText((Join-Path $root "WhatsApp.exe.sha256"), "$hash  WhatsApp.exe`n")

    Write-Output "WhatsApp.exe        v$Version"
    Write-Output "WhatsApp.exe.sha256 $hash"
    Write-Output ""
    Write-Output "Buat rilis di https://github.com/stevenarif123/whatsapp-web.view/releases/new"
    Write-Output "  tag: v$Version   unggah: WhatsApp.exe dan WhatsApp.exe.sha256"
}
finally {
    Pop-Location
}
