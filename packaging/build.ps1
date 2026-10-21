$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Mingw = Get-ChildItem -Path "$env:LOCALAPPDATA\Microsoft\WinGet\Packages" -Filter "gcc.exe" -Recurse -ErrorAction SilentlyContinue |
  Select-Object -First 1 -ExpandProperty DirectoryName
if ($Mingw) { $env:PATH = "$Mingw;$env:PATH" }
$env:CGO_ENABLED = "1"
$env:CC = "gcc"
Push-Location $Root
go build -ldflags="-H windowsgui" -o todo-go.exe .
Pop-Location
