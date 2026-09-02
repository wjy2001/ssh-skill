#!/usr/bin/env bash
# Build ssh-skill binaries into the skill directories with platform-suffixed names.
# The skill entrypoint is a platform-agnostic launcher (`bin/ssh-skill`), so the
# real binaries must carry a <os>-<arch> suffix to avoid clashing with it.
# Usage:
#   ./scripts/build.sh                   # build for the current platform
#   ./scripts/build.sh linux amd64       # explicit cross-compile
#   ./scripts/build.sh darwin arm64
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GO_DIR="$REPO_ROOT/go"
OUTPUT_DIR="$REPO_ROOT/.claude/skills/ssh-skill/bin"

# Resolve current platform so a bare invocation builds for the running OS.
CURRENT_OS="$(uname -s)"
case "$CURRENT_OS" in
  MINGW*|MSYS*|CYGWIN*|Windows_NT) CURRENT_OS="windows" ;;
  Darwin*) CURRENT_OS="darwin" ;;
  Linux*)  CURRENT_OS="linux" ;;
  *) echo "[build] Unsupported OS: $CURRENT_OS" >&2; exit 1 ;;
esac

GOOS="${1:-$CURRENT_OS}"
GOARCH="${2:-$(uname -m)}"
case "$GOARCH" in
  x86_64|amd64)     GOARCH="amd64" ;;
  aarch64|arm64)    GOARCH="arm64" ;;
  *) echo "[build] Unsupported arch: $GOARCH" >&2; exit 1 ;;
esac

BASE="ssh-skill-${GOOS}-${GOARCH}"
OUTPUT_BIN="$OUTPUT_DIR/$BASE"
if [ "$GOOS" = "windows" ]; then OUTPUT_BIN="$OUTPUT_BIN.exe"; fi

echo "==> Building ssh-skill ($GOOS/$GOARCH)..."
(
  cd "$GO_DIR"
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$OUTPUT_BIN" ./cmd/ssh-skill/
)
echo "==> Binary: $OUTPUT_BIN"

# Mirror to the Codex distribution copy so both stay in sync.
AGENTS_DIR="$REPO_ROOT/.agents/skills/ssh-skill/bin"
if [ -d "$AGENTS_DIR" ] && [ "$OUTPUT_DIR" != "$AGENTS_DIR" ]; then
  cp -f "$OUTPUT_BIN" "$AGENTS_DIR/$(basename "$OUTPUT_BIN")"
  echo "==> Mirrored: $AGENTS_DIR/$(basename "$OUTPUT_BIN")"
fi

echo "==> Done."