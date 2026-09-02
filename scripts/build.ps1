# Build ssh-skill Windows binary into the skill directories.
# The skill entrypoint is a platform-agnostic launcher (`bin/ssh-skill`), so the
# real binary carries a platform-suffixed name to avoid clashing with it.
# Usage: .\scripts\build.ps1
$ErrorActionPreference = "Stop"

$RepoRoot = (Resolve-Path "$PSScriptRoot\..").Path
$GoDir = Join-Path $RepoRoot "go"
$OutputDir = Join-Path $RepoRoot ".claude\skills\ssh-skill\bin"
$AgentsDir = Join-Path $RepoRoot ".agents\skills\ssh-skill\bin"
$GoArch = if ($env:PROCESSOR_ARCHITECTURE -match "ARM") { "arm64" } else { "amd64" }
$OutputBin = Join-Path $OutputDir "ssh-skill-windows-$GoArch.exe"

Write-Output "==> Building ssh-skill (windows/$GoArch)..."
Push-Location $GoDir
try {
    $env:GOOS = "windows"
    $env:GOARCH = $GoArch
    $env:CGO_ENABLED = "0"
    go build -o $OutputBin .\cmd\ssh-skill\
} finally {
    Pop-Location
}
Write-Output "==> Binary: $OutputBin"

# Mirror to the Codex distribution copy so both stay in sync.
if (Test-Path $AgentsDir) {
    Copy-Item -Force $OutputBin (Join-Path $AgentsDir "ssh-skill-windows-$GoArch.exe")
    Write-Output "==> Mirrored: $AgentsDir"
}
Write-Output "==> Done."