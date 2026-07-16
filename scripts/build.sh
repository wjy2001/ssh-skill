#!/usr/bin/env bash
# Build ssh-skill binary into the skill directory
# Usage: ./scripts/build.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GO_DIR="$REPO_ROOT/go"
OUTPUT_DIR="$REPO_ROOT/.claude/skills/ssh-skill/bin"
OUTPUT_BIN="$OUTPUT_DIR/ssh-skill"

echo "==> Building ssh-skill..."
cd "$GO_DIR"
go build -o "$OUTPUT_BIN" ./cmd/ssh-skill/

echo "==> Binary: $OUTPUT_BIN"
echo "==> Done."
