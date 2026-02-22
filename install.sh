#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# Crafter installer
#
# Local usage (from cloned repo):
#   ./install.sh [--global | --local] [--version VERSION]
#
# Remote usage (via curl):
#   curl -fsSL https://raw.githubusercontent.com/richardriman/crafter/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/richardriman/crafter/main/install.sh | bash -s -- --local
#   curl -fsSL https://raw.githubusercontent.com/richardriman/crafter/main/install.sh | bash -s -- --version 0.1.0
# ---------------------------------------------------------------------------

REPO="richardriman/crafter"
VERSION=""         # empty = use main branch
TEMP_DIR=""

# ---------------------------------------------------------------------------
# Detect whether we are running locally (from a cloned repo) or remotely
# (piped through curl/bash).  When piped, BASH_SOURCE[0] is either empty,
# "bash", or does not contain a directory with the expected source files.
# ---------------------------------------------------------------------------
_detect_script_dir() {
  local src="${BASH_SOURCE[0]:-}"
  if [[ -n "$src" && "$src" != "bash" && "$src" != "-" ]]; then
    local candidate
    candidate="$(cd "$(dirname "$src")" 2>/dev/null && pwd)" || true
    if [[ -n "$candidate" && -f "$candidate/VERSION" && -d "$candidate/commands" ]]; then
      echo "$candidate"
      return
    fi
  fi
  echo ""
}

SCRIPT_DIR="$(_detect_script_dir)"

# ---------------------------------------------------------------------------
# Cleanup trap — removes temp dir on exit when in remote mode
# ---------------------------------------------------------------------------
_cleanup() {
  if [[ -n "$TEMP_DIR" && -d "$TEMP_DIR" ]]; then
    rm -rf "$TEMP_DIR"
  fi
}
trap _cleanup EXIT

# ---------------------------------------------------------------------------
# Usage
# ---------------------------------------------------------------------------
usage() {
  cat <<EOF
Usage: ./install.sh [--global | --local] [--version VERSION]

  --global           Install Crafter globally to ~/.claude/ (default)
  --local            Install Crafter locally to .claude/ in the current project
  --version VERSION  Pin a specific release version (e.g. 0.1.0)
                     Defaults to the latest main branch

When run via curl | bash, the installer automatically downloads the required
files from GitHub — no git dependency needed.

After installing, open Claude Code and run /crafter:map-project to set up
project context.
EOF
}

# ---------------------------------------------------------------------------
# Remote mode: download tarball from GitHub and set SCRIPT_DIR
# ---------------------------------------------------------------------------
_download_release() {
  # Require curl
  if ! command -v curl &>/dev/null; then
    echo "Error: curl is required for remote installation but was not found." >&2
    exit 1
  fi
  if ! command -v tar &>/dev/null; then
    echo "Error: tar is required for remote installation but was not found." >&2
    exit 1
  fi

  TEMP_DIR="$(mktemp -d)"

  local tarball_url extract_subdir

  if [[ -z "$VERSION" ]]; then
    # Use main branch
    tarball_url="https://github.com/${REPO}/archive/refs/heads/main.tar.gz"
    extract_subdir="crafter-main"
  else
    # Try tag without 'v' prefix first (e.g. 0.1.0), then with 'v' prefix
    tarball_url="https://github.com/${REPO}/archive/refs/tags/${VERSION}.tar.gz"
    extract_subdir="crafter-${VERSION}"
  fi

  echo "Downloading Crafter from GitHub..."

  local http_code
  http_code="$(curl -fsSL -w "%{http_code}" -o "$TEMP_DIR/crafter.tar.gz" "$tarball_url" 2>/dev/null || true)"

  # If the versionless tag returned 404, try with 'v' prefix
  if [[ -n "$VERSION" && "$http_code" == "404" || -n "$VERSION" && ! -s "$TEMP_DIR/crafter.tar.gz" ]]; then
    local alt_url="https://github.com/${REPO}/archive/refs/tags/v${VERSION}.tar.gz"
    echo "Tag ${VERSION} not found, trying v${VERSION}..."
    http_code="$(curl -fsSL -w "%{http_code}" -o "$TEMP_DIR/crafter.tar.gz" "$alt_url" 2>/dev/null || true)"
    extract_subdir="crafter-v${VERSION}"
  fi

  if [[ ! -s "$TEMP_DIR/crafter.tar.gz" ]]; then
    echo "Error: Failed to download tarball." >&2
    if [[ "$http_code" == "404" ]]; then
      echo "  Version '${VERSION}' was not found on GitHub." >&2
      echo "  Check available releases at: https://github.com/${REPO}/releases" >&2
    else
      echo "  Network error (HTTP ${http_code:-unknown}). Check your internet connection." >&2
    fi
    exit 1
  fi

  if ! tar -tzf "$TEMP_DIR/crafter.tar.gz" &>/dev/null; then
    echo "Error: Downloaded file is not a valid archive." >&2
    exit 1
  fi
  if ! tar -xzf "$TEMP_DIR/crafter.tar.gz" -C "$TEMP_DIR"; then
    echo "Error: Downloaded file is not a valid archive." >&2
    exit 1
  fi

  # Locate extracted directory (handle unexpected subdir names gracefully)
  if [[ -d "$TEMP_DIR/$extract_subdir" ]]; then
    SCRIPT_DIR="$TEMP_DIR/$extract_subdir"
  else
    # Fallback: pick the first extracted directory
    local found
    found="$(find "$TEMP_DIR" -maxdepth 1 -mindepth 1 -type d | head -n 1)"
    if [[ -z "$found" ]]; then
      echo "Error: Could not locate extracted content in downloaded tarball." >&2
      exit 1
    fi
    SCRIPT_DIR="$found"
  fi
}

# ---------------------------------------------------------------------------
# Core install logic (unchanged from original)
# ---------------------------------------------------------------------------
install_to() {
  local base="$1"
  local label="$2"

  local commands_dest="$base/commands/crafter"
  local rules_dest="$base/crafter"
  local templates_dest="$base/crafter/templates"
  local meta_prompts_dest="$base/crafter/meta-prompts"

  echo "Installing Crafter $label..."

  mkdir -p "$commands_dest"
  cp "$SCRIPT_DIR/commands/do.md"          "$commands_dest/do.md"
  cp "$SCRIPT_DIR/commands/debug.md"       "$commands_dest/debug.md"
  cp "$SCRIPT_DIR/commands/status.md"      "$commands_dest/status.md"
  cp "$SCRIPT_DIR/commands/map-project.md" "$commands_dest/map-project.md"

  mkdir -p "$rules_dest"
  cp "$SCRIPT_DIR/VERSION"                 "$rules_dest/VERSION"
  cp "$SCRIPT_DIR/rules/core.md"           "$rules_dest/core.md"
  cp "$SCRIPT_DIR/rules/do-workflow.md"    "$rules_dest/do-workflow.md"
  cp "$SCRIPT_DIR/rules/debug-workflow.md" "$rules_dest/debug-workflow.md"
  cp "$SCRIPT_DIR/rules/delegation.md"     "$rules_dest/delegation.md"
  cp "$SCRIPT_DIR/rules/post-change.md"    "$rules_dest/post-change.md"
  cp "$SCRIPT_DIR/rules/task-lifecycle.md" "$rules_dest/task-lifecycle.md"
  cp "$SCRIPT_DIR/rules/update-check.md"   "$rules_dest/update-check.md"

  mkdir -p "$templates_dest"
  cp "$SCRIPT_DIR/templates/PROJECT.md"          "$templates_dest/PROJECT.md"
  cp "$SCRIPT_DIR/templates/ARCHITECTURE.md"     "$templates_dest/ARCHITECTURE.md"
  cp "$SCRIPT_DIR/templates/STATE.md"            "$templates_dest/STATE.md"
  cp "$SCRIPT_DIR/templates/claude-md.snippet"   "$templates_dest/claude-md.snippet"
  cp "$SCRIPT_DIR/templates/TASK.md"             "$templates_dest/TASK.md"

  mkdir -p "$meta_prompts_dest"
  cp "$SCRIPT_DIR/meta-prompts/planner.md"   "$meta_prompts_dest/planner.md"
  cp "$SCRIPT_DIR/meta-prompts/implement.md" "$meta_prompts_dest/implement.md"
  cp "$SCRIPT_DIR/meta-prompts/verify.md"    "$meta_prompts_dest/verify.md"
  cp "$SCRIPT_DIR/meta-prompts/review.md"    "$meta_prompts_dest/review.md"
  cp "$SCRIPT_DIR/meta-prompts/analyze.md"   "$meta_prompts_dest/analyze.md"
}

install_global() {
  install_to "$HOME/.claude" "globally"
  echo ""
  echo "Crafter installed globally."
  echo ""
  echo "Next steps:"
  echo "  Open Claude Code in your project and run:  /crafter:map-project"
}

install_local() {
  install_to "$(pwd)/.claude" "locally in $(pwd)"
  echo ""
  echo "Crafter installed locally in this project."
  echo ""
  echo "Next steps:"
  echo "  Open Claude Code and run:  /crafter:map-project"
}

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------
MODE="global"   # default

while [[ $# -gt 0 ]]; do
  case "$1" in
    --global)
      MODE="global"
      shift
      ;;
    --local)
      MODE="local"
      shift
      ;;
    --version)
      if [[ -z "${2:-}" ]]; then
        echo "Error: --version requires a value (e.g. --version 0.1.0)" >&2
        exit 1
      fi
      VERSION="$2"
      if [[ ! "$VERSION" =~ ^[a-zA-Z0-9._-]+$ ]]; then
        echo "Error: Invalid version format: '$VERSION'" >&2
        exit 1
      fi
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "Error: Unknown option: $1" >&2
      echo "" >&2
      usage >&2
      exit 1
      ;;
  esac
done

# ---------------------------------------------------------------------------
# Warn if --version is passed when running from a local clone
# ---------------------------------------------------------------------------
if [[ -n "$SCRIPT_DIR" && -n "$VERSION" ]]; then
  echo "Warning: --version is ignored when running from a local clone." >&2
fi

# ---------------------------------------------------------------------------
# Ensure source files are available
# ---------------------------------------------------------------------------
if [[ -z "$SCRIPT_DIR" ]]; then
  # Remote mode: download tarball
  _download_release
fi

# ---------------------------------------------------------------------------
# Run install
# ---------------------------------------------------------------------------
case "$MODE" in
  global)
    install_global
    ;;
  local)
    install_local
    ;;
esac
