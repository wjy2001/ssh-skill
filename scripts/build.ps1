# Build ssh-mcp binary into the skill directory
# Usage: .\scripts\build.ps1
$ErrorActionPreference = "Stop"

$RepoRoot = (Resolve-Path "$PSScriptRoot\..").Path
$GoDir = Join-Path $RepoRoot "go"
$OutputDir = Join-Path $RepoRoot ".claude\skills\ssh-ops\bin"
$OutputBin = Join-Path $OutputDir "ssh-mcp.exe"

Write-Output "==> Building ssh-mcp..."
Push-Location $GoDir
try {
    go build -o $OutputBin .\cmd\ssh-mcp\
    Write-Output "==> Binary: $OutputBin"
    Write-Output "==> Done."
} finally {
    Pop-Location
}
