#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
  cat <<EOF
Usage: ./install.sh [--global | --local]

  --global   Install Crafter globally to ~/.claude/ (available in all projects)
  --local    Install Crafter locally to .claude/ in the current project (committable, team-shareable)

Choose one. Then open Claude Code and run /crafter:map-project to set up project context.
EOF
}

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
  cp "$SCRIPT_DIR/rules/core.md"           "$rules_dest/core.md"
  cp "$SCRIPT_DIR/rules/do-workflow.md"    "$rules_dest/do-workflow.md"
  cp "$SCRIPT_DIR/rules/debug-workflow.md" "$rules_dest/debug-workflow.md"
  cp "$SCRIPT_DIR/rules/delegation.md"     "$rules_dest/delegation.md"
  cp "$SCRIPT_DIR/rules/post-change.md"    "$rules_dest/post-change.md"
  cp "$SCRIPT_DIR/rules/task-lifecycle.md"    "$rules_dest/task-lifecycle.md"

  mkdir -p "$templates_dest"
  cp "$SCRIPT_DIR/templates/PROJECT.md"          "$templates_dest/PROJECT.md"
  cp "$SCRIPT_DIR/templates/ARCHITECTURE.md"     "$templates_dest/ARCHITECTURE.md"
  cp "$SCRIPT_DIR/templates/STATE.md"            "$templates_dest/STATE.md"
  cp "$SCRIPT_DIR/templates/claude-md.snippet"   "$templates_dest/claude-md.snippet"
  cp "$SCRIPT_DIR/templates/TASK.md"            "$templates_dest/TASK.md"

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
  echo "✓ Crafter installed globally."
  echo ""
  echo "Next steps:"
  echo "  Open Claude Code in your project and run:  /crafter:map-project"
}

install_local() {
  install_to "$(pwd)/.claude" "locally in $(pwd)"
  echo ""
  echo "✓ Crafter installed locally in this project."
  echo ""
  echo "Next steps:"
  echo "  Open Claude Code and run:  /crafter:map-project"
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
