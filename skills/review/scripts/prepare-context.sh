#!/usr/bin/env bash
# prepare-context.sh — Collects review context for the code-review skill.
#
# Behavior depends on where the skill lives:
#
# A) Skill is in a subproject (no nested git repos):
#    ./prepare-context.sh                     # review changes in this project
#    ./prepare-context.sh file1 file2         # review specific files
#
# B) Skill is in a workspace with subprojects (dirs with .git):
#    ./prepare-context.sh --list-projects     # list available subprojects
#    ./prepare-context.sh --project=ruby-prj  # review changes in subproject
#    ./prepare-context.sh --project=.         # review workspace repo itself
#
# Output for --list-projects:
#   == PROJECTS ==
#   .  (workspace-name)
#   elixir-prj
#   frontend-v2
#   ruby-prj
#
# Output for review (stdout):
#   == PROJECT ==
#   frontend-v2
#   == LANGUAGES ==
#   elm, javascript
#   == RULES ==
#   --- elm-complexity.md ---
#   <contents>
#   == FILES ==
#   src/Page/Home.elm
#   src/utils.js

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ASSETS_DIR="$SKILL_DIR/assets"
# Skill is always at <project>/.claude/skills/<skill-name>/
PROJECT_ROOT="$(cd "$SKILL_DIR/../../.." && pwd)"

# --- Parse arguments ---------------------------------------------------------
PROJECT=""
LIST_PROJECTS=false
POSITIONAL=()
INSTALL_AGENT=false

for arg in "$@"; do
  case "$arg" in
    --list-projects)  LIST_PROJECTS=true ;;
    --install-agent)  INSTALL_AGENT=true ;;
    --project=*)      PROJECT="${arg#--project=}" ;;
    *)                POSITIONAL+=("$arg") ;;
  esac
done

# --- Agent installation -------------------------------------------------------
# The agent definition lives in assets/ (distribution source) and gets copied
# to .claude/agents/ on first use. Do not edit the copy in agents/ directly —
# it will be overwritten on next install.
AGENT_FILE="$PROJECT_ROOT/.claude/agents/code-reviewer.md"
AGENT_ASSET="$ASSETS_DIR/code-review-agent.md"

if [[ "$INSTALL_AGENT" == true ]]; then
  mkdir -p "$(dirname "$AGENT_FILE")"
  cp "$AGENT_ASSET" "$AGENT_FILE"
  echo "== AGENT_INSTALLED =="
  echo "The code-reviewer agent has been installed."
  echo "Tell the user to restart Claude Code (/exit and relaunch) and run /review again."
  exit 0
fi

if [[ ! -f "$AGENT_FILE" ]]; then
  echo "== ASK_USER =="
  cat <<'ASKEOF'
{
  "questions": [{
    "question": "The code-reviewer agent is not installed yet. Install it now?",
    "header": "Install",
    "multiSelect": false,
    "options": [
      { "label": "Yes", "description": "Install the agent (requires restart to use /review)" },
      { "label": "No", "description": "Skip installation" }
    ]
  }]
}
ASKEOF
  echo "== ON_YES =="
  echo ".claude/skills/review/scripts/prepare-context.sh --install-agent"
  exit 0
fi

# --- Detect if workspace has subprojects -------------------------------------
has_subprojects=false
for dir in "$PROJECT_ROOT"/*/; do
  if [[ -d "$dir/.git" ]]; then
    has_subprojects=true
    break
  fi
done

# --- List projects mode (workspace only) -------------------------------------
if [[ "$LIST_PROJECTS" == true ]]; then
  if [[ "$has_subprojects" == false ]]; then
    echo "Single project — no subprojects to list." >&2
    exit 0
  fi

  # Collect all projects
  all_projects=()
  if [[ -d "$PROJECT_ROOT/.git" ]]; then
    all_projects+=(".")
  fi
  for dir in "$PROJECT_ROOT"/*/; do
    [[ -d "$dir/.git" ]] || continue
    all_projects+=("$(basename "$dir")")
  done

  # Build AskUserQuestion JSON with up to 4 options
  echo "== ASK_USER =="
  max=4
  options=""
  sep=""
  for i in "${!all_projects[@]}"; do
    (( i >= max )) && break
    p="${all_projects[$i]}"
    if [[ "$p" == "." ]]; then
      label="."; desc="workspace root ($(basename "$PROJECT_ROOT"))"
    else
      label="$p"; desc="$p project"
    fi
    options+="${sep}{\"label\":\"$label\",\"description\":\"$desc\"}"
    sep=","
  done

  cat <<ASKEOF
{"questions":[{"question":"Which project do you want to review?","header":"Project","multiSelect":false,"options":[$options]}]}
ASKEOF

  # If more than 4, list the rest for "Other" fallback
  if (( ${#all_projects[@]} > max )); then
    echo "== OTHER_PROJECTS =="
    for i in "${!all_projects[@]}"; do
      (( i < max )) && continue
      echo "${all_projects[$i]}"
    done
  fi

  exit 0
fi

# --- Determine working directory ---------------------------------------------
if [[ "$has_subprojects" == true ]]; then
  # Workspace mode — require --project
  if [[ -z "$PROJECT" ]]; then
    echo "Error: this is a workspace with subprojects. Use --project=<name> or --list-projects." >&2
    exit 1
  fi
  if [[ "$PROJECT" == "." ]]; then
    WORK_DIR="$PROJECT_ROOT"
  else
    WORK_DIR="$PROJECT_ROOT/$PROJECT"
  fi
  if [[ ! -d "$WORK_DIR/.git" ]]; then
    echo "Error: '$PROJECT' is not a git repository (no .git in $WORK_DIR)" >&2
    exit 1
  fi
else
  # Single project mode — work directly here
  WORK_DIR="$PROJECT_ROOT"
  PROJECT="$(basename "$PROJECT_ROOT")"
fi

echo "== PROJECT =="
echo "$PROJECT"

# --- Collect changed files ---------------------------------------------------
cd "$WORK_DIR"

if [[ ${#POSITIONAL[@]} -gt 0 ]]; then
  files=("${POSITIONAL[@]}")
else
  # Auto-detect: staged → unstaged → untracked → diff against default branch
  mapfile -t staged < <(git diff --name-only --cached 2>/dev/null)
  mapfile -t unstaged < <(git diff --name-only 2>/dev/null)
  mapfile -t untracked < <(git ls-files --others --exclude-standard 2>/dev/null)

  files=("${staged[@]}" "${unstaged[@]}" "${untracked[@]}")

  if [[ ${#files[@]} -gt 0 ]]; then
    mapfile -t files < <(printf '%s\n' "${files[@]}" | sort -u)
  fi

  # Fallback: compare against default branch
  if [[ ${#files[@]} -eq 0 ]]; then
    for candidate in main master; do
      if git rev-parse --verify "$candidate" &>/dev/null; then
        mapfile -t files < <(git diff --name-only "$candidate"...HEAD 2>/dev/null)
        break
      fi
    done
  fi
fi

if [[ ${#files[@]} -eq 0 ]]; then
  echo "No files to review (no staged, unstaged, untracked, or branch changes detected)." >&2
  exit 0
fi

# --- Detect languages from extensions ----------------------------------------
declare -A lang_map=(
  [elm]=elm
  [rb]=ruby
  [rake]=ruby
  [ex]=elixir
  [exs]=elixir
  [js]=javascript
  [ts]=typescript
  [tsx]=typescript
  [jsx]=javascript
  [rs]=rust
  [py]=python
  [sh]=bash
  [bash]=bash
  [css]=css
  [scss]=css
  [html]=html
  [eex]=elixir
  [heex]=elixir
  [leex]=elixir
)

declare -A detected_langs=()

for f in "${files[@]}"; do
  ext="${f##*.}"
  if [[ -n "${lang_map[$ext]+_}" ]]; then
    detected_langs["${lang_map[$ext]}"]=1
  fi
done

if [[ ${#detected_langs[@]} -eq 0 ]]; then
  echo "No recognized languages in changed files." >&2
fi

# --- Find and load rule files ------------------------------------------------
echo "== LANGUAGES =="
printf '%s\n' "${!detected_langs[@]}" | sort | paste -sd', '

echo "== RULES =="
rules_found=0
for lang in "${!detected_langs[@]}"; do
  for rule_file in "$ASSETS_DIR"/${lang}-*.md; do
    [[ -f "$rule_file" ]] || continue
    rules_found=1
    echo "--- $(basename "$rule_file") ---"
    cat "$rule_file"
    echo ""
  done
done

if [[ $rules_found -eq 0 ]]; then
  echo "(no language-specific rules found)"
fi

echo "== FILES =="
printf '%s\n' "${files[@]}"

# --- Run language-specific tools ---------------------------------------------
if [[ -n "${detected_langs[elm]+_}" ]]; then
  echo ""
  echo "== ELM-REVIEW =="
  if command -v elm-review &>/dev/null; then
    cd "$WORK_DIR"
    elm-review "${POSITIONAL[@]}" 2>&1 || true
  else
    echo "(elm-review not available)"
  fi
fi

if [[ -n "${detected_langs[rust]+_}" ]]; then
  echo ""
  echo "== RUST-CLIPPY =="
  if command -v cargo &>/dev/null && cargo clippy --version &>/dev/null; then
    cd "$WORK_DIR"
    cargo clippy --workspace --message-format=short 2>&1 || true
  else
    echo "(cargo clippy not available)"
  fi
fi
