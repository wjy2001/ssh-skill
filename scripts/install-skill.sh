#!/usr/bin/env bash
# Minimal installer: only SKILL.md + current-platform binary.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/wjy2001/ssh-skill/master/scripts/install-skill.sh | bash
#   bash install-skill.sh [ref]
set -euo pipefail

REPO_OWNER="${SSH_SKILL_REPO_OWNER:-wjy2001}"
REPO_NAME="${SSH_SKILL_REPO_NAME:-ssh-skill}"
REF="${1:-${SSH_SKILL_REF:-master}}"
RAW_BASE="https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/${REF}/.claude/skills/ssh-skill"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux*|darwin*)
    BIN_NAME="ssh-skill"
    DEST_DIR="${HOME}/.claude/skills/ssh-skill"
    ;;
  msys*|mingw*|cygwin*)
    BIN_NAME="ssh-skill.exe"
    DEST_DIR="${USERPROFILE:-$HOME}/.claude/skills/ssh-skill"
    DEST_DIR="$(cygpath -u "$DEST_DIR" 2>/dev/null || echo "$DEST_DIR")"
    ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

BIN_DIR="${DEST_DIR}/bin"
TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

echo "==> Installing ssh-skill skill (minimal: SKILL.md + ${BIN_NAME})"
echo "    source: ${RAW_BASE}"
echo "    dest:   ${DEST_DIR}"

mkdir -p "$BIN_DIR"

download() {
  local url="$1"
  local out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$out"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$out" "$url"
  else
    echo "Need curl or wget" >&2
    exit 1
  fi
}

download "${RAW_BASE}/SKILL.md" "${TMP_DIR}/SKILL.md"
download "${RAW_BASE}/bin/${BIN_NAME}" "${TMP_DIR}/${BIN_NAME}"

# Overwrite skill files only; never touch ~/.ssh-skill vault.
install -m 0644 "${TMP_DIR}/SKILL.md" "${DEST_DIR}/SKILL.md"
install -m 0755 "${TMP_DIR}/${BIN_NAME}" "${BIN_DIR}/${BIN_NAME}"

# Remove the other platform binary if an old full copy left it behind.
if [[ "$BIN_NAME" == "ssh-skill" ]]; then
  rm -f "${BIN_DIR}/ssh-skill.exe"
else
  rm -f "${BIN_DIR}/ssh-skill"
fi

echo "==> Verifying..."
"${BIN_DIR}/${BIN_NAME}" --version

echo "==> Done."
echo "    skill: ${DEST_DIR}"
echo "    vault is NOT modified (still ~/.ssh-skill if present)"
echo "    next: ask your agent to run vault init / add a server"
