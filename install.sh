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
REMOTE_MODE=0      # set to 1 when running from downloaded source (curl|bash path)

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
    if [[ -n "$candidate" && -f "$candidate/VERSION" && ( -d "$candidate/skills" || -d "$candidate/commands" ) ]]; then
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

After installing, open Claude Code and run the '/crafter-map-project' skill to
set up project context.
EOF
}

# ---------------------------------------------------------------------------
# Remote mode: download tarball from GitHub and set SCRIPT_DIR
# ---------------------------------------------------------------------------
_download_release() {
  REMOTE_MODE=1

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
    # Prefer v-prefixed tag (common in this repo), then fall back to raw version.
    tarball_url="https://github.com/${REPO}/archive/refs/tags/v${VERSION}.tar.gz"
    extract_subdir="crafter-v${VERSION}"
  fi

  echo "Downloading Crafter from GitHub..."

  local http_code
  http_code="$(curl -fsSL -w "%{http_code}" -o "$TEMP_DIR/crafter.tar.gz" "$tarball_url" 2>/dev/null || true)"

  # If the v-prefixed tag failed, try the raw version tag.
  if [[ -n "$VERSION" && ( "$http_code" == "404" || ! -s "$TEMP_DIR/crafter.tar.gz" ) ]]; then
    local alt_url="https://github.com/${REPO}/archive/refs/tags/${VERSION}.tar.gz"
    echo "Tag v${VERSION} not found, trying ${VERSION}..."
    http_code="$(curl -fsSL -w "%{http_code}" -o "$TEMP_DIR/crafter.tar.gz" "$alt_url" 2>/dev/null || true)"
    extract_subdir="crafter-${VERSION}"
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
# Download CLI binary for the current platform/arch (release mode only)
# ---------------------------------------------------------------------------
_download_cli_binary() {
  local dest_dir="$1"
  local dest_bin="$dest_dir/crafter/bin/crafter"

  # install_to may have already copied a local pre-built binary from
  # SCRIPT_DIR/cli/bin/crafter. If so, avoid network/build work.
  if [[ -x "$dest_bin" ]]; then
    return 0
  fi

  # os/arch are lowercased to match GitHub release asset naming convention
  # (e.g. crafter-darwin-arm64, crafter-linux-amd64).
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  case "$arch" in
    x86_64)       arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *)
      echo "Warning: Unsupported architecture '$arch'; skipping CLI binary download." >&2
      return 0
      ;;
  esac

  local binary_name="crafter-${os}-${arch}"

  # Ensure the bin/ directory exists regardless of whether install_to has
  # already created it (makes this function self-contained).
  mkdir -p "$(dirname "$dest_bin")"

  # In remote mode, VERSION may be empty when installing from main. In that
  # case, try the downloaded source VERSION for release-asset lookup.
  local release_version="$VERSION"
  if [[ -z "$release_version" && "$REMOTE_MODE" -eq 1 && -f "$SCRIPT_DIR/VERSION" ]]; then
    release_version="$(tr -d '[:space:]' < "$SCRIPT_DIR/VERSION")"
    release_version="${release_version#v}"
  fi

  if [[ -n "$release_version" && "$REMOTE_MODE" -eq 1 ]]; then
    if ! command -v curl &>/dev/null; then
      echo "Warning: curl not found; skipping CLI binary download." >&2
    else
      local release_tag http_code binary_url
      for release_tag in "v${release_version}" "${release_version}"; do
        binary_url="https://github.com/${REPO}/releases/download/${release_tag}/${binary_name}"
        echo "Downloading CLI binary ${binary_name} (${release_tag})..."
        http_code="$(curl -fsSL -w "%{http_code}" -o "$dest_bin" "$binary_url" 2>/dev/null || true)"
        if [[ -s "$dest_bin" && "$http_code" != "404" ]]; then
          chmod +x "$dest_bin"
          echo "CLI binary installed to $dest_bin"
          return 0
        fi
        rm -f "$dest_bin"
      done
      echo "Warning: CLI binary not available for this platform/version (${binary_name}); trying source build." >&2
    fi
  fi

  if [[ "$REMOTE_MODE" -eq 1 && -f "$SCRIPT_DIR/cli/go.mod" ]]; then
    if ! command -v go &>/dev/null; then
      echo "Warning: go not found; skipping CLI source build fallback." >&2
      return 0
    fi
    local temp_tool_versions_created=0
    if [[ -f "$HOME/.tool-versions" && ! -f "$SCRIPT_DIR/cli/.tool-versions" ]]; then
      cp "$HOME/.tool-versions" "$SCRIPT_DIR/cli/.tool-versions"
      temp_tool_versions_created=1
    fi
    local go_version=""
    go_version="$(awk '/^go[[:space:]]+/ { print $2; exit }' "$SCRIPT_DIR/cli/go.mod" 2>/dev/null || true)"
    if [[ -n "$go_version" ]] && command -v asdf &>/dev/null; then
      local asdf_go_version=""
      asdf_go_version="$(
        asdf list golang 2>/dev/null \
          | tr -d ' *' \
          | awk -v pfx="$go_version" '$0 ~ "^" pfx "(\\.|$)" { print; exit }'
      )"
      if [[ -n "$asdf_go_version" ]]; then
        go_version="$asdf_go_version"
      fi
    fi
    echo "Building CLI binary from source..."
    if (cd "$SCRIPT_DIR/cli" && ASDF_GOLANG_VERSION="$go_version" go build -o "$dest_bin" .); then
      chmod +x "$dest_bin"
      echo "CLI binary installed to $dest_bin"
    else
      rm -f "$dest_bin"
      echo "Warning: failed to build CLI binary from source; skipping." >&2
    fi
    if [[ "$temp_tool_versions_created" -eq 1 ]]; then
      rm -f "$SCRIPT_DIR/cli/.tool-versions"
    fi
  fi
}

# ---------------------------------------------------------------------------
# Make global CLI available in PATH via ~/.local/bin/crafter symlink
# ---------------------------------------------------------------------------
_link_cli_into_path() {
  local base_dir="$1"
  local installed_bin="$base_dir/crafter/bin/crafter"
  local link_dir="$HOME/.local/bin"
  local link_path="$link_dir/crafter"

  if [[ ! -x "$installed_bin" ]]; then
    return 0
  fi

  mkdir -p "$link_dir"
  ln -sf "$installed_bin" "$link_path"
  echo "CLI command linked at $link_path"

  case ":$PATH:" in
    *":$link_dir:"*) ;;
    *)
      echo "Warning: $link_dir is not in PATH." >&2
      echo "  Add this to your shell profile: export PATH=\"$link_dir:\$PATH\"" >&2
      ;;
  esac
}

# ---------------------------------------------------------------------------
# Core install logic
# ---------------------------------------------------------------------------
install_to() {
  local base="$1"
  local label="$2"

  local commands_dest="$base/commands/crafter"
  local skills_dest="$base/skills"
  local crafter_dest="$base/crafter"
  local rules_dest="$base/crafter/rules"
  local templates_dest="$base/crafter/templates"
  local agents_dest="$base/agents"

  # Clean previously installed files to prevent stale leftovers on upgrade
  rm -rf "$commands_dest" "$crafter_dest"
  rm -f "$agents_dest"/crafter-*.md
  if [[ -d "$skills_dest" ]]; then
    local stale_skill
    for stale_skill in "$skills_dest"/crafter-*; do
      [[ -d "$stale_skill" ]] || continue
      rm -rf "$stale_skill"
    done
  fi

  echo "Installing Crafter $label..."

  mkdir -p "$commands_dest"
  cp "$SCRIPT_DIR/commands/do.md"          "$commands_dest/do.md"
  cp "$SCRIPT_DIR/commands/debug.md"       "$commands_dest/debug.md"
  cp "$SCRIPT_DIR/commands/status.md"      "$commands_dest/status.md"
  cp "$SCRIPT_DIR/commands/map-project.md" "$commands_dest/map-project.md"

  mkdir -p "$skills_dest/crafter-do" "$skills_dest/crafter-debug" "$skills_dest/crafter-status" "$skills_dest/crafter-map-project"
  cp "$SCRIPT_DIR/skills/crafter-do/SKILL.md"          "$skills_dest/crafter-do/SKILL.md"
  cp "$SCRIPT_DIR/skills/crafter-debug/SKILL.md"       "$skills_dest/crafter-debug/SKILL.md"
  cp "$SCRIPT_DIR/skills/crafter-status/SKILL.md"      "$skills_dest/crafter-status/SKILL.md"
  cp "$SCRIPT_DIR/skills/crafter-map-project/SKILL.md" "$skills_dest/crafter-map-project/SKILL.md"

  mkdir -p "$crafter_dest"
  cp "$SCRIPT_DIR/VERSION"                 "$crafter_dest/VERSION"

  mkdir -p "$rules_dest"
  cp "$SCRIPT_DIR/rules/core.md"           "$rules_dest/core.md"
  cp "$SCRIPT_DIR/rules/do-workflow.md"    "$rules_dest/do-workflow.md"
  cp "$SCRIPT_DIR/rules/debug-workflow.md" "$rules_dest/debug-workflow.md"
  cp "$SCRIPT_DIR/rules/delegation.md"     "$rules_dest/delegation.md"
  cp "$SCRIPT_DIR/rules/post-change.md"    "$rules_dest/post-change.md"
  cp "$SCRIPT_DIR/rules/task-lifecycle.md" "$rules_dest/task-lifecycle.md"

  mkdir -p "$templates_dest"
  cp "$SCRIPT_DIR/templates/PROJECT.md"          "$templates_dest/PROJECT.md"
  cp "$SCRIPT_DIR/templates/ARCHITECTURE.md"     "$templates_dest/ARCHITECTURE.md"
  cp "$SCRIPT_DIR/templates/STATE.md"            "$templates_dest/STATE.md"
  cp "$SCRIPT_DIR/templates/TASK.md"             "$templates_dest/TASK.md"

  mkdir -p "$agents_dest"
  cp "$SCRIPT_DIR/agents/crafter-planner.md"     "$agents_dest/crafter-planner.md"
  cp "$SCRIPT_DIR/agents/crafter-implementer.md" "$agents_dest/crafter-implementer.md"
  cp "$SCRIPT_DIR/agents/crafter-verifier.md"    "$agents_dest/crafter-verifier.md"
  cp "$SCRIPT_DIR/agents/crafter-reviewer.md"    "$agents_dest/crafter-reviewer.md"
  cp "$SCRIPT_DIR/agents/crafter-analyzer.md"    "$agents_dest/crafter-analyzer.md"

  mkdir -p "$crafter_dest/bin"

  # Local clone install: copy pre-built binary if it exists
  if [[ -f "$SCRIPT_DIR/cli/bin/crafter" ]]; then
    cp "$SCRIPT_DIR/cli/bin/crafter" "$crafter_dest/bin/crafter"
    chmod +x "$crafter_dest/bin/crafter"
  fi
}

install_hook() {
  local hooks_dir="$HOME/.claude/hooks"
  local settings_file="$HOME/.claude/settings.json"
  local hook_dest="$hooks_dir/crafter-check-update.js"

  mkdir -p "$hooks_dir"
  cp "$SCRIPT_DIR/hooks/crafter-check-update.js" "$hook_dest"

  if ! command -v node &>/dev/null; then
    echo "Warning: node not found, skipping hook registration"
    return 0
  fi

  SETTINGS_FILE="$settings_file" HOOK_CMD="node \"$hook_dest\"" node -e '
    const fs = require("fs");
    const settingsFile = process.env.SETTINGS_FILE;
    const hookCommand = process.env.HOOK_CMD;

    let settings = {};
    try {
      settings = JSON.parse(fs.readFileSync(settingsFile, "utf8"));
    } catch (e) {}

    if (!settings.hooks) settings.hooks = {};
    if (!Array.isArray(settings.hooks.SessionStart)) settings.hooks.SessionStart = [];

    // Check if already registered
    const alreadyRegistered = settings.hooks.SessionStart.some(function(entry) {
      return entry.hooks && entry.hooks.some(function(h) {
        return h.command === hookCommand;
      });
    });

    if (!alreadyRegistered) {
      settings.hooks.SessionStart.push({
        hooks: [{ type: "command", command: hookCommand }]
      });
    }

    fs.writeFileSync(settingsFile, JSON.stringify(settings, null, 2) + "\n");
  '
}

install_global() {
  install_to "$HOME/.claude" "globally"
  _download_cli_binary "$HOME/.claude"
  _link_cli_into_path "$HOME/.claude"
  install_hook
  echo ""
  echo "Crafter installed globally."
  echo ""
  echo "Next steps:"
  echo "  Open Claude Code in your project and run the '/crafter-map-project' skill."
}

install_local() {
  install_to "$(pwd)/.claude" "locally in $(pwd)"
  _download_cli_binary "$(pwd)/.claude"
  install_hook
  echo ""
  echo "Crafter installed locally in this project."
  echo ""
  echo "Next steps:"
  echo "  Open Claude Code and run the '/crafter-map-project' skill."
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
      VERSION="${VERSION#v}"
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
