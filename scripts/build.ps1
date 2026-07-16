# Build ssh-skill binary into the skill directory
# Usage: .\scripts\build.ps1
$ErrorActionPreference = "Stop"

$RepoRoot = (Resolve-Path "$PSScriptRoot\..").Path
$GoDir = Join-Path $RepoRoot "go"
$OutputDir = Join-Path $RepoRoot ".claude\skills\ssh-skill\bin"
$OutputBin = Join-Path $OutputDir "ssh-skill.exe"

Write-Output "==> Building ssh-skill..."
Push-Location $GoDir
try {
    go build -o $OutputBin .\cmd\ssh-skill\
    Write-Output "==> Binary: $OutputBin"
    Write-Output "==> Done."
} finally {
    Pop-Location
}
