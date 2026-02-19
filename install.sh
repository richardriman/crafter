#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
  cat <<EOF
Usage: ./install.sh [--global | --local]

  --global   Install Crafter commands and rules globally to ~/.claude/
  --local    Set up .planning/ context files and CLAUDE.md in the current project
EOF
}

install_global() {
  local commands_dest="$HOME/.claude/commands/crafter"
  local rules_dest="$HOME/.claude/crafter"
  local templates_dest="$HOME/.claude/crafter/templates"

  echo "Installing Crafter globally..."

  mkdir -p "$commands_dest"
  cp "$SCRIPT_DIR/commands/do.md"          "$commands_dest/do.md"
  cp "$SCRIPT_DIR/commands/debug.md"       "$commands_dest/debug.md"
  cp "$SCRIPT_DIR/commands/status.md"      "$commands_dest/status.md"
  cp "$SCRIPT_DIR/commands/map-project.md" "$commands_dest/map-project.md"

  mkdir -p "$rules_dest"
  cp "$SCRIPT_DIR/rules/rules.md" "$rules_dest/rules.md"

  mkdir -p "$templates_dest"
  cp "$SCRIPT_DIR/templates/PROJECT.md"          "$templates_dest/PROJECT.md"
  cp "$SCRIPT_DIR/templates/ARCHITECTURE.md"     "$templates_dest/ARCHITECTURE.md"
  cp "$SCRIPT_DIR/templates/STATE.md"            "$templates_dest/STATE.md"
  cp "$SCRIPT_DIR/templates/claude-md.snippet"   "$templates_dest/claude-md.snippet"

  echo ""
  echo "✓ Crafter installed globally."
  echo ""
  echo "Next steps:"
  echo "  1. In your project directory, run:  $SCRIPT_DIR/install.sh --local"
  echo "  2. Open Claude Code and run:        /crafter:map-project"
}

install_local() {
  local planning_dir=".planning"
  local snippet="$SCRIPT_DIR/templates/claude-md.snippet"

  echo "Setting up Crafter in $(pwd)..."

  # --- .planning/ context files ---
  mkdir -p "$planning_dir"

  for tpl in PROJECT.md ARCHITECTURE.md STATE.md; do
    if [ -f "$planning_dir/$tpl" ]; then
      echo "  ✓ $planning_dir/$tpl already exists — skipping"
    else
      cp "$SCRIPT_DIR/templates/$tpl" "$planning_dir/$tpl"
      echo "  ✓ Created $planning_dir/$tpl"
    fi
  done

  # --- CLAUDE.md ---
  local snippet_content
  snippet_content="$(cat "$snippet")"

  if [ ! -f "CLAUDE.md" ]; then
    # No CLAUDE.md — create it with the snippet
    cp "$snippet" "CLAUDE.md"
    echo "  ✓ Created CLAUDE.md with Crafter snippet"
  elif grep -q "<!-- crafter:start -->" "CLAUDE.md"; then
    # Marker exists — replace only the crafter section
    local tmpfile
    tmpfile="$(mktemp)"
    awk -v snippet="$snippet_content" '
      /<!-- crafter:start -->/ { print snippet; skip=1; next }
      /<!-- crafter:end -->/   { skip=0; next }
      !skip                    { print }
    ' "CLAUDE.md" > "$tmpfile"

    mv "$tmpfile" "CLAUDE.md"
    echo "  ✓ Updated Crafter section in existing CLAUDE.md"
  else
    # CLAUDE.md exists but no marker — append the snippet
    echo "" >> "CLAUDE.md"
    cat "$snippet" >> "CLAUDE.md"
    echo "  ✓ Appended Crafter snippet to existing CLAUDE.md"
    echo ""
    echo "  Added:"
    cat "$snippet"
  fi

  # Recommend map-project for large CLAUDE.md files
  if [ -f "CLAUDE.md" ]; then
    local line_count
    line_count="$(wc -l < "CLAUDE.md")"
    if [ "$line_count" -gt 50 ]; then
      echo ""
      echo "  ⚠  Your CLAUDE.md has $line_count lines."
      echo "     Consider running /crafter:map-project to migrate content to .planning/ files."
    fi
  fi

  echo ""
  echo "✓ Crafter set up in this project."
  echo ""
  echo "Next steps:"
  echo "  1. Open Claude Code in this directory"
  echo "  2. Run: /crafter:map-project   (to analyze and populate .planning/ files)"
  echo "  3. Run: /crafter:do <your task>"
}

if [ $# -eq 0 ]; then
  usage
  exit 0
fi

case "${1:-}" in
  --global)
    install_global
    ;;
  --local)
    install_local
    ;;
  *)
    echo "Unknown option: $1"
    echo ""
    usage
    exit 1
    ;;
esac
