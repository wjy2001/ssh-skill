# Minimal installer: SKILL.md + platform-agnostic launcher + Windows binary.
# Usage:
#   irm https://raw.githubusercontent.com/wjy2001/ssh-skill/master/scripts/install-skill.ps1 | iex
#   .\install-skill.ps1 [-Ref master]
param(
    [string]$Ref = $(if ($env:SSH_SKILL_REF) { $env:SSH_SKILL_REF } else { "master" }),
    [string]$RepoOwner = $(if ($env:SSH_SKILL_REPO_OWNER) { $env:SSH_SKILL_REPO_OWNER } else { "wjy2001" }),
    [string]$RepoName = $(if ($env:SSH_SKILL_REPO_NAME) { $env:SSH_SKILL_REPO_NAME } else { "ssh-skill" })
)

$ErrorActionPreference = "Stop"

$RawBase = "https://raw.githubusercontent.com/$RepoOwner/$RepoName/$Ref/.claude/skills/ssh-skill"
$DestDir = Join-Path $env:USERPROFILE ".claude\skills\ssh-skill"
$BinDir = Join-Path $DestDir "bin"

# Prebuilt binaries are amd64 only; ARM64 users must build from source.
$IsArm = ($env:PROCESSOR_ARCHITECTURE -match "ARM") -or ($env:PROCESSOR_ARCHITEW6432 -match "ARM")
if ($IsArm) {
    throw "No ARM64 prebuilt binary; run .\scripts\build.ps1 to build ssh-skill-windows-arm64.exe"
}
$BinName = "ssh-skill-windows-amd64.exe"

Write-Host "==> Installing ssh-skill skill (SKILL.md + launcher + $BinName)"
Write-Host "    source: $RawBase"
Write-Host "    dest:   $DestDir"

New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
$Tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("ssh-skill-install-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Force -Path $Tmp | Out-Null

try {
    $skillTmp = Join-Path $Tmp "SKILL.md"
    $launcherTmp = Join-Path $Tmp "ssh-skill"
    $binTmp = Join-Path $Tmp $BinName
    Invoke-WebRequest -Uri "$RawBase/SKILL.md" -OutFile $skillTmp -UseBasicParsing
    Invoke-WebRequest -Uri "$RawBase/bin/ssh-skill" -OutFile $launcherTmp -UseBasicParsing
    Invoke-WebRequest -Uri "$RawBase/bin/$BinName" -OutFile $binTmp -UseBasicParsing

    # Overwrite skill files only; never touch ~/.ssh-skill vault.
    Copy-Item -Force $skillTmp (Join-Path $DestDir "SKILL.md")
    Copy-Item -Force $launcherTmp (Join-Path $BinDir "ssh-skill")
    Copy-Item -Force $binTmp (Join-Path $BinDir $BinName)

    # Remove stale binaries left by older installs / other platforms.
    Remove-Item -Force (Join-Path $BinDir "ssh-skill.exe") -ErrorAction SilentlyContinue
    Get-ChildItem -Path $BinDir -Filter "ssh-skill-linux-*" -ErrorAction SilentlyContinue | Remove-Item -Force
    Get-ChildItem -Path $BinDir -Filter "ssh-skill-darwin-*" -ErrorAction SilentlyContinue | Remove-Item -Force

    Write-Host "==> Verifying..."
    & (Join-Path $BinDir $BinName) --version
    if ($LASTEXITCODE -ne 0) { throw "binary verification failed" }

    Write-Host "==> Done."
    Write-Host "    skill: $DestDir"
    Write-Host "    vault is NOT modified (still ~/.ssh-skill if present)"
    Write-Host "    next: ask your agent to run vault init / add a server"
}
finally {
    Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
}