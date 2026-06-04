#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# Test suite for install.sh
# Pure Bash, zero external dependencies.
# ---------------------------------------------------------------------------

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_SH="$REPO_DIR/install.sh"

# ---------------------------------------------------------------------------
# Global cleanup trap — removes any temp directories left over on interrupt
# ---------------------------------------------------------------------------
_ALL_TMPDIRS=()
# Cached path to the real (mutation-capable) crafter binary used by G-series
# statusline tests.  Set once by _build_real_crafter_bin; empty until then.
_REAL_CRAFTER_BIN=""
_global_cleanup() {
  local d
  for d in "${_ALL_TMPDIRS[@]+"${_ALL_TMPDIRS[@]}"}"; do
    [[ -d "$d" ]] && rm -rf "$d"
  done
}
trap _global_cleanup EXIT INT TERM

# Wrapper: create a temp dir and register it for cleanup
_make_tmp() {
  local d
  d="$(mktemp -d)"
  _ALL_TMPDIRS+=("$d")
  echo "$d"
}

# ---------------------------------------------------------------------------
# Colour helpers (gracefully degrade when terminal lacks support)
# ---------------------------------------------------------------------------
if [[ -t 1 ]]; then
  _GREEN='\033[0;32m'
  _RED='\033[0;31m'
  _RESET='\033[0m'
else
  _GREEN=''
  _RED=''
  _RESET=''
fi

# ---------------------------------------------------------------------------
# Test runner state
# ---------------------------------------------------------------------------
_PASS=0
_FAIL=0
_CURRENT_TEST=""

# ---------------------------------------------------------------------------
# Assertion helpers
# ---------------------------------------------------------------------------
assert_file_exists() {
  local path="$1"
  if [[ -f "$path" ]]; then
    return 0
  else
    _fail "assert_file_exists: '$path' does not exist or is not a regular file"
    return 1
  fi
}

assert_dir_exists() {
  local path="$1"
  if [[ -d "$path" ]]; then
    return 0
  else
    _fail "assert_dir_exists: '$path' does not exist or is not a directory"
    return 1
  fi
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  if [[ "$haystack" == *"$needle"* ]]; then
    return 0
  else
    _fail "assert_contains: expected to find '${needle}' in output"
    return 1
  fi
}

assert_exit_code() {
  local expected="$1"
  local actual="$2"
  if [[ "$actual" -eq "$expected" ]]; then
    return 0
  else
    _fail "assert_exit_code: expected exit code ${expected}, got ${actual}"
    return 1
  fi
}

assert_file_nonempty() {
  local path="$1"
  if [[ -s "$path" ]]; then
    return 0
  else
    _fail "assert_file_nonempty: '$path' is empty or does not exist"
    return 1
  fi
}

assert_file_not_exists() {
  local path="$1"
  if [[ ! -e "$path" ]]; then
    return 0
  else
    _fail "assert_file_not_exists: '$path' exists but should have been removed"
    return 1
  fi
}

assert_symlink_target() {
  local path="$1"
  local expected="$2"
  if [[ ! -L "$path" ]]; then
    _fail "assert_symlink_target: '$path' is not a symlink"
    return 1
  fi
  local actual
  actual="$(readlink "$path")"
  if [[ "$actual" != "$expected" ]]; then
    _fail "assert_symlink_target: expected '$path' -> '$expected', got '$actual'"
    return 1
  fi
}

# Internal: record a failure and increment counter (does NOT abort the test)
_fail() {
  printf "  ${_RED}FAIL${_RESET} [%s]: %s\n" "$_CURRENT_TEST" "$1" >&2
  _FAIL=$(( _FAIL + 1 ))
}

# ---------------------------------------------------------------------------
# Test runner
# ---------------------------------------------------------------------------
run_test() {
  local func="$1"
  _CURRENT_TEST="$func"
  local before_fail=$_FAIL
  # Disable -e locally so a failed assertion doesn't abort the whole suite
  set +e
  "$func"
  set -e
  if [[ $_FAIL -eq $before_fail ]]; then
    printf "  ${_GREEN}PASS${_RESET} %s\n" "$func"
    _PASS=$(( _PASS + 1 ))
  fi
}

# ---------------------------------------------------------------------------
# Subprocess helpers
# ---------------------------------------------------------------------------

# Run the installer as a subprocess (global mode by default).
# Sets the caller's variables named by $3 (output) and $4 (exit code).
# Internal variable names are prefixed with _ri_ to avoid colliding with
# callers who name their variables "output" / "ec".
#
# Usage: _run_installer <home_dir> <pwd_dir> <output_var> <ec_var> [args...]
#
# <pwd_dir> sets the working directory for the installer subprocess.
# Global install tests pass any temp dir (HOME is what matters).
# Local install tests pass a dedicated sandbox directory so that
# install_local() writes .claude/ there instead of into the real repo.
_run_installer() {
  local _ri_home="$1" _ri_pwd="$2" _ri_out_var="$3" _ri_ec_var="$4"
  shift 4
  local _ri_output="" _ri_ec=0
  # Seed the fake home with ~/.tool-versions if present, so that version
  # manager shims (e.g. asdf) can resolve tools like node in the subprocess.
  if [[ -f "$HOME/.tool-versions" && ! -f "$_ri_home/.tool-versions" ]]; then
    cp "$HOME/.tool-versions" "$_ri_home/.tool-versions"
  fi
  # Use _ri_pwd as working directory when provided; fall back to REPO_DIR.
  local _ri_cwd="${_ri_pwd:-$REPO_DIR}"
  _ri_output="$(cd "$_ri_cwd" && HOME="$_ri_home" bash "$INSTALL_SH" "$@" 2>&1)" \
    && _ri_ec=$? || _ri_ec=$?
  printf -v "$_ri_out_var"  '%s' "$_ri_output"
  printf -v "$_ri_ec_var"   '%d' "$_ri_ec"
}

# Create a fake "go" executable that satisfies:
#   go build -o <dest> .
# by writing a tiny executable to <dest>. Returns directory path via stdout.
_make_fake_go_bin_dir() {
  local shim_dir
  shim_dir="$(_make_tmp)/fake_go_bin"
  mkdir -p "$shim_dir"
  printf '%s\n' \
    '#!/usr/bin/env bash' \
    'set -euo pipefail' \
    'out=""' \
    'while [[ $# -gt 0 ]]; do' \
    '  case "$1" in' \
    '    -o)' \
    '      out="$2"' \
    '      shift 2' \
    '      ;;' \
    '    *)' \
    '      shift' \
    '      ;;' \
    '  esac' \
    'done' \
    'if [[ -z "$out" ]]; then' \
    '  echo "fake go: missing -o output path" >&2' \
    '  exit 1' \
    'fi' \
    'mkdir -p "$(dirname "$out")"' \
    'printf '"'"'#!/usr/bin/env bash\necho fake-built-crafter\n'"'"' > "$out"' \
    'chmod +x "$out"' \
    > "$shim_dir/go"
  chmod +x "$shim_dir/go"
  echo "$shim_dir"
}

# Build a real (mutation-capable) crafter binary from source and cache the path
# in $_REAL_CRAFTER_BIN.  Prints the cached path to stdout.
# Uses `mise exec -- go build` in $REPO_DIR/cli; the build happens at most once
# per test-suite run.  Exits the whole suite with an error message if the build
# cannot succeed (so no test silently passes against a dummy binary).
_build_real_crafter_bin() {
  if [[ -n "$_REAL_CRAFTER_BIN" && -x "$_REAL_CRAFTER_BIN" ]]; then
    echo "$_REAL_CRAFTER_BIN"
    return 0
  fi
  local out
  out="$(_make_tmp)/crafter-real"
  if ! (cd "$REPO_DIR/cli" && mise exec -- go build -o "$out" . 2>/dev/null); then
    echo "ERROR: _build_real_crafter_bin: failed to build crafter from source." >&2
    echo "  Make sure 'mise exec -- go build' works in $REPO_DIR/cli." >&2
    exit 1
  fi
  chmod +x "$out"
  _REAL_CRAFTER_BIN="$out"
  echo "$out"
}

# Create a fake "go" directory whose shim, instead of compiling, copies the
# pre-built real crafter binary to the -o destination.  This makes the
# installer's _download_cli_binary place a mutation-capable binary so that
# install_statusline can actually write to settings.json.
# Returns the shim directory path via stdout.
_make_real_go_bin_dir() {
  local real_bin
  real_bin="$(_build_real_crafter_bin)"
  local shim_dir
  shim_dir="$(_make_tmp)/real_go_bin"
  mkdir -p "$shim_dir"
  # Write the shim with the real binary path interpolated.
  printf '%s\n' \
    '#!/usr/bin/env bash' \
    'set -euo pipefail' \
    'out=""' \
    'while [[ $# -gt 0 ]]; do' \
    '  case "$1" in' \
    '    -o)' \
    '      out="$2"' \
    '      shift 2' \
    '      ;;' \
    '    *)' \
    '      shift' \
    '      ;;' \
    '  esac' \
    'done' \
    'if [[ -z "$out" ]]; then' \
    '  echo "real-go shim: missing -o output path" >&2' \
    '  exit 1' \
    'fi' \
    'mkdir -p "$(dirname "$out")"' \
    "cp $(printf '%q' "$real_bin") \"\$out\"" \
    'chmod +x "$out"' \
    > "$shim_dir/go"
  chmod +x "$shim_dir/go"
  echo "$shim_dir"
}

# Extract the guidance JSON block from installer output.
# The block appears after the "paste this into your settings.json:" marker line.
# Uses awk brace-depth counting so commands containing literal braces are handled.
# Prints the extracted JSON block to stdout; prints nothing if not found.
_extract_guidance_json() {
  local output="$1"
  printf '%s\n' "$output" | awk '
    /paste this into your settings.json:/ { found=1; next }
    found && /^[[:space:]]*$/ && !started { next }
    found && /^\{/ { started=1; depth=0 }
    started {
      print
      n=split($0,c,"")
      for(i=1;i<=n;i++){
        if(c[i]=="{") depth++
        else if(c[i]=="}") { depth--; if(depth==0) exit }
      }
    }
  '
}

# Decode a JSON string value from a line of the form:
#   `    "command": "...",`
# Strips the key prefix and trailing comma/quotes, then decodes \" -> " and \\ -> \.
_decode_json_cmd_line() {
  local line="$1"
  # Strip leading whitespace, key, and opening quote; strip trailing `",` or `"`.
  local encoded
  encoded="$(printf '%s' "$line" | sed 's/^[[:space:]]*"command": "//; s/",[[:space:]]*$//; s/"[[:space:]]*$//')"
  # Decode JSON string escapes relevant to our output: \" -> " and \\ -> \
  printf '%s' "$encoded" | sed 's/\\"/"/g; s/\\\\/\\/g'
}

# Build a symlink-based "safe_bin" directory that mirrors the real PATH but
# omits the specified command name.  Returns the directory path via stdout.
# The caller is responsible for removing the directory when done.
# Iterates over every directory in the current PATH for full portability,
# including /opt/homebrew/bin (Apple Silicon) and /usr/sbin.
_make_safe_bin_without() {
  local missing_cmd="$1"
  local safe_bin
  safe_bin="$(_make_tmp)/safe_bin"
  mkdir -p "$safe_bin"
  local dir f name
  # Use the live PATH plus any extra directories for completeness; dedup via
  # first-found wins logic below.
  local IFS=":"
  for dir in $PATH /usr/local/bin /usr/bin /bin /usr/sbin /opt/homebrew/bin; do
    [[ -d "$dir" ]] || continue
    for f in "$dir"/*; do
      [[ -x "$f" ]] || continue
      name="$(basename "$f")"
      [[ "$name" == "$missing_cmd" ]] && continue   # omit the command under test
      [[ -e "$safe_bin/$name" ]]      && continue   # first-found wins
      ln -sf "$f" "$safe_bin/$name"
    done
  done
  echo "$safe_bin"
}

# Run the installer with SCRIPT_DIR forced to empty (simulating remote/piped
# execution) and the given command removed from PATH.
# Sets caller's variables named by $2 (output) and $3 (exit code).
_run_installer_no_cmd() {
  local _rn_missing="$1" _rn_out_var="$2" _rn_ec_var="$3"
  local _rn_tmp _rn_safe_bin _rn_wrapper _rn_output="" _rn_ec=0

  _rn_tmp="$(_make_tmp)"
  _rn_safe_bin="$(_make_safe_bin_without "$_rn_missing")"

  # Build a patched copy of install.sh that forces SCRIPT_DIR to empty so
  # the remote-download code path (_download_release) is exercised.
  _rn_wrapper="$_rn_tmp/wrapper.sh"
  printf '#!/usr/bin/env bash\nset -euo pipefail\n' > "$_rn_wrapper"
  awk '
    /^SCRIPT_DIR=/ {
      print $0
      print "SCRIPT_DIR=\"\"   # overridden for test"
      next
    }
    { print }
  ' "$INSTALL_SH" >> "$_rn_wrapper"

  _rn_output="$(PATH="$_rn_safe_bin" bash "$_rn_wrapper" 2>&1)" \
    && _rn_ec=$? || _rn_ec=$?

  printf -v "$_rn_out_var" '%s' "$_rn_output"
  printf -v "$_rn_ec_var"  '%d' "$_rn_ec"
}

# ---------------------------------------------------------------------------
# Shared file list used by global-install tests (issue #9: avoid duplication)
# ---------------------------------------------------------------------------
# 24 files installed by install_global / install_to
_EXPECTED_FILES_REL=(
  "skills/crafter-do/SKILL.md"
  "skills/crafter-debug/SKILL.md"
  "skills/crafter-status/SKILL.md"
  "skills/crafter-map-project/SKILL.md"
  "skills/crafter-scaffold-task/SKILL.md"
  "crafter/VERSION"
  "crafter/rules/core.md"
  "crafter/rules/do-workflow.md"
  "crafter/rules/debug-workflow.md"
  "crafter/rules/delegation.md"
  "crafter/rules/post-change.md"
  "crafter/rules/task-lifecycle.md"
  "crafter/rules/do/flag-validation.md"
  "crafter/rules/do/project-resolution.md"
  "crafter/rules/do/extension-skills.md"
  "crafter/rules/do/step-0-resume.md"
  "crafter/rules/do/step-1-scope.md"
  "crafter/rules/do/step-2-discuss.md"
  "crafter/rules/do/step-3-plan.md"
  "crafter/rules/do/step-4-execute.md"
  "crafter/rules/do/step-5-drift.md"
  "crafter/rules/do/step-5a-phase-verification.md"
  "crafter/rules/do/step-6-review.md"
  "crafter/rules/do/step-6b-phase-summary.md"
  "crafter/rules/do/step-6a-session-break.md"
  "crafter/rules/do/step-7-9-post-change.md"
  "crafter/rules/do/step-9b-pr-composition.md"
  "crafter/templates/PROJECT.md"
  "crafter/templates/ARCHITECTURE.md"
  "crafter/templates/STATE.md"
  "crafter/templates/TASK.md"
  "agents/crafter-planner.md"
  "agents/crafter-implementer.md"
  "agents/crafter-verifier.md"
  "agents/crafter-reviewer.md"
  "agents/crafter-analyzer.md"
)

# ---------------------------------------------------------------------------
# A. Argument parsing
# ---------------------------------------------------------------------------

test_default_mode_is_global() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec
  assert_exit_code 0 "$ec"
  assert_contains "$output" "globally"
}

test_local_flag_sets_local_mode() {
  local tmp home_dir proj_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  # Run installer with working directory set to the sandbox project directory
  # so that install_local() creates .claude/ there, not in the real repo.
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"
  assert_contains "$output" "locally"
  # Verify that .claude/ was actually created in the sandbox project directory
  assert_dir_exists "$proj_dir/.claude"
}

test_version_flag_sets_version() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  # --version is silently ignored (with a warning) when running from a local
  # clone, but must NOT cause an error exit.
  _run_installer "$home_dir" "$tmp" output ec --version 0.1.0
  assert_exit_code 0 "$ec"
}

test_version_flag_accepts_v_prefix() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --version v0.1.0
  assert_exit_code 0 "$ec"
}

test_version_without_value_exits_error() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --version
  assert_exit_code 1 "$ec"
  assert_contains "$output" "--version requires a value"
}

test_version_invalid_chars_exits_error() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --version "bad version!"
  assert_exit_code 1 "$ec"
  assert_contains "$output" "Invalid version format"
}

test_help_flag_exits_zero() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --help
  assert_exit_code 0 "$ec"
  assert_contains "$output" "Usage:"
  assert_contains "$output" "'/crafter-map-project' skill"
}

test_help_short_flag_exits_zero() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec -h
  assert_exit_code 0 "$ec"
  assert_contains "$output" "Usage:"
}

test_unknown_flag_exits_error() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --not-a-real-flag
  assert_exit_code 1 "$ec"
  assert_contains "$output" "Unknown option"
}

# ---------------------------------------------------------------------------
# B. Global install
# ---------------------------------------------------------------------------

test_global_creates_expected_directories() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"
  assert_dir_exists "$home_dir/.claude/skills"
  assert_dir_exists "$home_dir/.claude/crafter"
  assert_dir_exists "$home_dir/.claude/crafter/rules"
  assert_dir_exists "$home_dir/.claude/crafter/templates"
  assert_dir_exists "$home_dir/.claude/agents"
}

test_global_creates_bin_directory() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"
  assert_dir_exists "$home_dir/.claude/crafter/bin"
}

test_global_copies_local_cli_binary() {
  # A global install must produce a binary at ~/.claude/crafter/bin/crafter.
  # Post-#37 there is no cli/bin/crafter copy path; the binary is produced by
  # the source-build fallback in _download_cli_binary.  We inject a fast fake
  # `go` shim so no real compilation is needed.
  local tmp home_dir output ec fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"

  fake_go_bin="$(_make_fake_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$tmp" output ec --global
  result_ec=$ec
  PATH="$old_path"

  assert_exit_code 0 "$result_ec"
  assert_file_exists "$home_dir/.claude/crafter/bin/crafter"
}

test_global_builds_cli_from_source_when_local_binary_missing() {
  local tmp home_dir output ec local_bin backup_bin had_local_bin fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"

  local_bin="$REPO_DIR/cli/bin/crafter"
  backup_bin="$tmp/crafter-bin-backup"
  had_local_bin=0
  if [[ -f "$local_bin" ]]; then
    mv "$local_bin" "$backup_bin"
    had_local_bin=1
  fi

  fake_go_bin="$(_make_fake_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$tmp" output ec --global
  result_ec=$ec
  PATH="$old_path"

  if [[ $had_local_bin -eq 1 ]]; then
    mv "$backup_bin" "$local_bin"
  fi

  assert_exit_code 0 "$result_ec"
  assert_file_exists "$home_dir/.claude/crafter/bin/crafter"
  assert_file_exists "$home_dir/.local/bin/crafter"
  assert_contains "$output" "Building CLI binary from source"
}

test_global_links_cli_to_home_local_bin() {
  # A global install must symlink ~/.local/bin/crafter to the installed binary.
  # Post-#37 the binary is produced by the source-build fallback in
  # _download_cli_binary.  We inject a fast fake `go` shim so no real
  # compilation is needed.
  local tmp home_dir output ec fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"

  fake_go_bin="$(_make_fake_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$tmp" output ec --global
  result_ec=$ec
  PATH="$old_path"

  assert_exit_code 0 "$result_ec"
  assert_file_exists "$home_dir/.local/bin/crafter"
  assert_symlink_target "$home_dir/.local/bin/crafter" "$home_dir/.claude/crafter/bin/crafter"
}

test_local_copies_local_cli_binary() {
  # A local install must produce a binary at <proj>/.claude/crafter/bin/crafter.
  # Post-#37 the binary is produced by the source-build fallback in
  # _download_cli_binary.  We inject a fast fake `go` shim so no real
  # compilation is needed.
  local tmp home_dir proj_dir output ec fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"

  fake_go_bin="$(_make_fake_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$proj_dir" output ec --local
  result_ec=$ec
  PATH="$old_path"

  assert_exit_code 0 "$result_ec"
  assert_file_exists "$proj_dir/.claude/crafter/bin/crafter"
}

test_local_builds_cli_from_source_when_local_binary_missing() {
  local tmp home_dir proj_dir output ec local_bin backup_bin had_local_bin fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"

  local_bin="$REPO_DIR/cli/bin/crafter"
  backup_bin="$tmp/crafter-bin-backup"
  had_local_bin=0
  if [[ -f "$local_bin" ]]; then
    mv "$local_bin" "$backup_bin"
    had_local_bin=1
  fi

  fake_go_bin="$(_make_fake_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$proj_dir" output ec --local
  result_ec=$ec
  PATH="$old_path"

  if [[ $had_local_bin -eq 1 ]]; then
    mv "$backup_bin" "$local_bin"
  fi

  assert_exit_code 0 "$result_ec"
  assert_file_exists "$proj_dir/.claude/crafter/bin/crafter"
  assert_contains "$output" "Building CLI binary from source"
}

test_local_does_not_link_cli_to_home_local_bin() {
  local tmp home_dir proj_dir fake_bin output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"

  fake_bin="$REPO_DIR/cli/bin/crafter"
  local fake_bin_created=0
  if [[ ! -f "$fake_bin" ]]; then
    mkdir -p "$(dirname "$fake_bin")"
    printf '#!/usr/bin/env bash\necho fake-crafter\n' > "$fake_bin"
    chmod +x "$fake_bin"
    fake_bin_created=1
  fi

  _run_installer "$home_dir" "$proj_dir" output ec --local
  local result_ec=$ec

  if [[ $fake_bin_created -eq 1 ]]; then
    rm -f "$fake_bin"
  fi

  assert_exit_code 0 "$result_ec"
  assert_file_not_exists "$home_dir/.local/bin/crafter"
}

test_global_copies_all_expected_files() {
  local tmp home_dir output ec base rel
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"

  base="$home_dir/.claude"
  for rel in "${_EXPECTED_FILES_REL[@]}"; do
    assert_file_exists "$base/$rel"
  done
}

test_global_copied_files_are_nonempty() {
  local tmp home_dir output ec base rel
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"

  base="$home_dir/.claude"
  for rel in "${_EXPECTED_FILES_REL[@]}"; do
    assert_file_nonempty "$base/$rel"
  done
}

test_global_output_contains_installed_globally() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"
  assert_contains "$output" "installed globally"
  assert_contains "$output" "run the '/crafter-map-project' skill"
}

# ---------------------------------------------------------------------------
# C. Local install
# ---------------------------------------------------------------------------

test_local_creates_expected_directories() {
  local tmp home_dir proj_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  # Pass proj_dir as the working directory so .claude/ is created there.
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"
  assert_dir_exists "$proj_dir/.claude/skills"
  assert_dir_exists "$proj_dir/.claude/crafter"
  assert_dir_exists "$proj_dir/.claude/crafter/rules"
  assert_dir_exists "$proj_dir/.claude/crafter/templates"
  assert_dir_exists "$proj_dir/.claude/agents"
}

test_local_copies_all_expected_files() {
  local tmp home_dir proj_dir output ec base rel
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"

  base="$proj_dir/.claude"
  for rel in "${_EXPECTED_FILES_REL[@]}"; do
    assert_file_exists "$base/$rel"
  done
}

test_local_output_contains_installed_locally() {
  local tmp home_dir proj_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"
  assert_contains "$output" "installed locally"
  assert_contains "$output" "run the '/crafter-map-project' skill"
}

test_local_creates_bin_directory() {
  local tmp home_dir proj_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"
  assert_dir_exists "$proj_dir/.claude/crafter/bin"
}

# ---------------------------------------------------------------------------
# D. Idempotency
# ---------------------------------------------------------------------------

test_global_idempotency() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"
  # Second run must also succeed
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"
  assert_contains "$output" "installed globally"
}

test_local_idempotency() {
  local tmp home_dir proj_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"
  # Second run
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"
  assert_contains "$output" "installed locally"
}

# ---------------------------------------------------------------------------
# D1. Placeholder resolution
# ---------------------------------------------------------------------------

test_global_resolves_crafter_home_placeholder() {
  local tmp home_dir output ec base leaked_files
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"

  base="$home_dir/.claude"

  # (a) No literal {CRAFTER_HOME} placeholder must remain under rules or skills
  leaked_files="$(grep -rl '{CRAFTER_HOME}' "$base/crafter/rules" "$base/skills" 2>/dev/null || true)"
  if [[ -n "$leaked_files" ]]; then
    _fail "test_global_resolves_crafter_home_placeholder: literal {CRAFTER_HOME} found in deployed tree: $leaked_files"
  fi

  # (b) Representative path: post-change.md must contain the resolved rules path
  local post_change="$base/crafter/rules/post-change.md"
  local expected_path="$base/crafter/rules/task-lifecycle.md"
  if ! grep -qF "$expected_path" "$post_change" 2>/dev/null; then
    _fail "test_global_resolves_crafter_home_placeholder: '$post_change' does not contain resolved path '$expected_path'"
  fi

  # (c) No wrong-base absolute path: every absolute-path reference to
  # task-lifecycle.md in the deployed rules+skills tree must use exactly $base
  # as the prefix — no other absolute .claude/crafter base must appear.
  # Tilde-form (~/.claude/...) documentation references are excluded since they
  # are not resolved absolute paths; only lines lacking ~/ are checked.
  local wrong_base_refs
  wrong_base_refs="$(grep -rh '/crafter/rules/task-lifecycle\.md' "$base/crafter/rules" "$base/skills" 2>/dev/null \
    | grep -v '~/' \
    | grep -vF "$base/crafter/rules/task-lifecycle.md" || true)"
  if [[ -n "$wrong_base_refs" ]]; then
    _fail "test_global_resolves_crafter_home_placeholder: absolute-path reference to task-lifecycle.md with wrong base found: $wrong_base_refs"
  fi
}

test_local_resolves_crafter_home_placeholder() {
  local tmp home_dir proj_dir output ec base leaked_files
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"

  base="$proj_dir/.claude"

  # (a) No literal {CRAFTER_HOME} placeholder must remain under rules or skills
  leaked_files="$(grep -rl '{CRAFTER_HOME}' "$base/crafter/rules" "$base/skills" 2>/dev/null || true)"
  if [[ -n "$leaked_files" ]]; then
    _fail "test_local_resolves_crafter_home_placeholder: literal {CRAFTER_HOME} found in deployed tree: $leaked_files"
  fi

  # (b) Representative path: post-change.md must contain the resolved rules path
  local post_change="$base/crafter/rules/post-change.md"
  local expected_path="$base/crafter/rules/task-lifecycle.md"
  if ! grep -qF "$expected_path" "$post_change" 2>/dev/null; then
    _fail "test_local_resolves_crafter_home_placeholder: '$post_change' does not contain resolved path '$expected_path'"
  fi

  # (c) No wrong-base absolute path: every absolute-path reference to
  # task-lifecycle.md in the deployed rules+skills tree must use exactly $base
  # as the prefix — no global $HOME/.claude base must appear.
  # Tilde-form (~/.claude/...) documentation references are excluded since they
  # are not resolved absolute paths; only lines lacking ~/ are checked.
  local wrong_base_refs
  wrong_base_refs="$(grep -rh '/crafter/rules/task-lifecycle\.md' "$base/crafter/rules" "$base/skills" 2>/dev/null \
    | grep -v '~/' \
    | grep -vF "$base/crafter/rules/task-lifecycle.md" || true)"
  if [[ -n "$wrong_base_refs" ]]; then
    _fail "test_local_resolves_crafter_home_placeholder: absolute-path reference to task-lifecycle.md with wrong base found: $wrong_base_refs"
  fi
}

# ---------------------------------------------------------------------------
# D2. Upgrade cleans stale files
# ---------------------------------------------------------------------------

test_global_upgrade_removes_stale_files() {
  local tmp home_dir output ec base
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  # First install
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"

  base="$home_dir/.claude"
  # Simulate stale files from a previous version
  echo "stale" > "$base/crafter/rules/old-rule.md"
  echo "stale" > "$base/crafter/templates/old-template.md"
  mkdir -p "$base/commands/crafter"
  echo "stale" > "$base/commands/crafter/old-command.md"
  echo "stale" > "$base/agents/crafter-old-agent.md"
  mkdir -p "$base/skills/crafter-old-skill"
  echo "stale" > "$base/skills/crafter-old-skill/SKILL.md"

  # Second install (upgrade)
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"

  # Stale files must be gone
  assert_file_not_exists "$base/crafter/rules/old-rule.md"
  assert_file_not_exists "$base/crafter/templates/old-template.md"
  assert_file_not_exists "$base/commands/crafter/old-command.md"
  assert_file_not_exists "$base/agents/crafter-old-agent.md"
  assert_file_not_exists "$base/skills/crafter-old-skill/SKILL.md"

  # Current files must still be present
  for rel in "${_EXPECTED_FILES_REL[@]}"; do
    assert_file_exists "$base/$rel"
  done
}

test_local_upgrade_removes_stale_files() {
  local tmp home_dir proj_dir output ec base
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  # First install
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"

  base="$proj_dir/.claude"
  # Simulate stale files
  echo "stale" > "$base/crafter/rules/old-rule.md"
  mkdir -p "$base/commands/crafter"
  echo "stale" > "$base/commands/crafter/old-command.md"
  echo "stale" > "$base/agents/crafter-old-agent.md"
  mkdir -p "$base/skills/crafter-old-skill"
  echo "stale" > "$base/skills/crafter-old-skill/SKILL.md"

  # Second install (upgrade)
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"

  # Stale files must be gone
  assert_file_not_exists "$base/crafter/rules/old-rule.md"
  assert_file_not_exists "$base/commands/crafter/old-command.md"
  assert_file_not_exists "$base/agents/crafter-old-agent.md"
  assert_file_not_exists "$base/skills/crafter-old-skill/SKILL.md"

  # Current files must still be present
  for rel in "${_EXPECTED_FILES_REL[@]}"; do
    assert_file_exists "$base/$rel"
  done
}

test_upgrade_preserves_non_crafter_agents_and_skills() {
  local tmp home_dir output ec base
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  # First install
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"

  base="$home_dir/.claude"
  # Place a non-Crafter agent file
  echo "other tool agent" > "$base/agents/other-tool-agent.md"
  mkdir -p "$base/skills/other-tool-skill"
  echo "other tool skill" > "$base/skills/other-tool-skill/SKILL.md"

  # Second install (upgrade)
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"

  # Non-Crafter agent must survive
  assert_file_exists "$base/agents/other-tool-agent.md"
  assert_file_exists "$base/skills/other-tool-skill/SKILL.md"
}

# ---------------------------------------------------------------------------
# E. Hook installation
# ---------------------------------------------------------------------------

test_hook_source_file_exists_in_repo() {
  assert_file_exists "$REPO_DIR/hooks/crafter-check-update.js"
}

test_global_installs_hook_file() {
  local tmp home_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"
  assert_file_exists "$home_dir/.claude/hooks/crafter-check-update.js"
}

test_local_installs_hook_file() {
  local tmp home_dir proj_dir output ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"
  assert_file_exists "$home_dir/.claude/hooks/crafter-check-update.js"
}

test_global_registers_hook_in_settings() {
  local tmp home_dir output ec settings real_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  # Inject a real mutation-capable crafter binary so `crafter install hook`
  # actually writes settings.json (without a real binary the hook registration
  # is silently skipped and settings.json is never created).
  real_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$real_go_bin:$PATH"
  _run_installer "$home_dir" "$tmp" output ec --global
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"
  assert_file_exists "$home_dir/.claude/settings.json"
  settings="$(cat "$home_dir/.claude/settings.json")"
  assert_contains "$settings" "crafter-check-update.js"
  assert_contains "$settings" "SessionStart"
}

test_local_registers_hook_in_settings() {
  local tmp home_dir proj_dir output ec settings real_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  # Inject a real mutation-capable crafter binary so `crafter install hook`
  # actually writes $HOME/.claude/settings.json (both modes target HOME).
  real_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$real_go_bin:$PATH"
  _run_installer "$home_dir" "$proj_dir" output ec --local
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"
  assert_file_exists "$home_dir/.claude/settings.json"
  settings="$(cat "$home_dir/.claude/settings.json")"
  assert_contains "$settings" "crafter-check-update.js"
  assert_contains "$settings" "SessionStart"
}

test_hook_registration_is_idempotent() {
  local tmp home_dir output ec settings count real_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  # Inject a real mutation-capable crafter binary for both runs.
  # install_to deletes the crafter dir on each run, so the binary must be
  # re-placed by the shim each time (same pattern as G4 statusline tests).
  real_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$real_go_bin:$PATH"
  _run_installer "$home_dir" "$tmp" output ec --global
  assert_exit_code 0 "$ec"
  # Second run must not duplicate the hook entry
  _run_installer "$home_dir" "$tmp" output ec --global
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"
  settings="$(cat "$home_dir/.claude/settings.json")"
  count="$(printf '%s\n' "$settings" | grep -o "crafter-check-update.js" | wc -l | tr -d ' ')"
  if [[ "$count" -ne 1 ]]; then
    _fail "hook_registration_is_idempotent: expected 1 occurrence of crafter-check-update.js in settings.json, got $count"
  fi
}

# ---------------------------------------------------------------------------
# F. _detect_script_dir
# ---------------------------------------------------------------------------
# Extract the real _detect_script_dir function from install.sh and wrap it
# so that tests can supply the source path as a parameter instead of relying
# on the read-only BASH_SOURCE[0].
#
# The wrapper _detect_script_dir_with_src() replaces the BASH_SOURCE[0]
# reference with a caller-supplied $1 argument, but otherwise executes the
# identical production logic extracted from install.sh.

# Extract the function body from install.sh using brace-depth counting so
# nested braces are handled correctly, then source it into the current shell.
_source_detect_func() {
  local func_body
  func_body="$(awk '
    /^_detect_script_dir\(\)/ { found=1; depth=0 }
    found {
      print
      # Count opening and closing braces on this line
      n = split($0, chars, "")
      for (i = 1; i <= n; i++) {
        if (chars[i] == "{") depth++
        else if (chars[i] == "}") {
          depth--
          if (depth == 0) { found=0; exit }
        }
      }
    }
  ' "$INSTALL_SH")"
  # Evaluate the extracted function definition so it is available in-scope.
  eval "$func_body"
}

# Build a wrapper that calls the real production logic with a caller-supplied
# source path instead of BASH_SOURCE[0].  This is the only way to test the
# function without forking a subprocess that sets BASH_SOURCE differently.
_define_detect_wrapper() {
  # Extract the raw function text from install.sh.
  local func_body
  func_body="$(awk '
    /^_detect_script_dir\(\)/ { found=1; depth=0 }
    found {
      print
      n = split($0, chars, "")
      for (i = 1; i <= n; i++) {
        if (chars[i] == "{") depth++
        else if (chars[i] == "}") {
          depth--
          if (depth == 0) { found=0; exit }
        }
      }
    }
  ' "$INSTALL_SH")"

  # Mechanically derive the wrapper from the production function body:
  #   1. Rename the function to _detect_script_dir_with_src.
  #   2. Replace the read-only BASH_SOURCE[0] reference with the $1 argument.
  local wrapper_body
  wrapper_body="$(
    printf '%s\n' "$func_body" \
      | sed 's/^_detect_script_dir()/_detect_script_dir_with_src()/' \
      | sed 's/\${BASH_SOURCE\[0\]:-}/$1/g'
  )"
  eval "$wrapper_body"
}

test_detect_script_dir_returns_repo_when_valid() {
  # Verify that a path whose directory contains VERSION + commands/ is
  # recognised as a valid script dir.  We pass INSTALL_SH (which lives in
  # REPO_DIR) so the candidate resolves to REPO_DIR.
  local result
  result="$(
    _define_detect_wrapper
    _detect_script_dir_with_src "$INSTALL_SH"
  )"
  if [[ "$result" == "$REPO_DIR" ]]; then
    return 0
  else
    _fail "Expected '$REPO_DIR', got '$result'"
  fi
}

test_detect_script_dir_empty_when_source_empty() {
  local result
  result="$(
    _define_detect_wrapper
    _detect_script_dir_with_src ""
  )"
  if [[ -z "$result" ]]; then
    return 0
  else
    _fail "Expected empty string, got '$result'"
  fi
}

test_detect_script_dir_empty_when_source_is_bash() {
  local result
  result="$(
    _define_detect_wrapper
    _detect_script_dir_with_src "bash"
  )"
  if [[ -z "$result" ]]; then
    return 0
  else
    _fail "Expected empty string, got '$result'"
  fi
}

test_detect_script_dir_empty_when_source_is_dash() {
  local result
  result="$(
    _define_detect_wrapper
    _detect_script_dir_with_src "-"
  )"
  if [[ -z "$result" ]]; then
    return 0
  else
    _fail "Expected empty string, got '$result'"
  fi
}

test_detect_script_dir_empty_when_no_version_file() {
  local tmp result
  tmp="$(_make_tmp)"
  mkdir -p "$tmp/commands"
  # No VERSION file — place a fake script inside tmp so dirname == tmp
  touch "$tmp/fake_install.sh"

  result="$(
    _define_detect_wrapper
    _detect_script_dir_with_src "$tmp/fake_install.sh"
  )"
  if [[ -z "$result" ]]; then
    return 0
  else
    _fail "Expected empty string, got '$result'"
  fi
}

test_detect_script_dir_empty_when_no_commands_or_skills_dir() {
  local tmp result
  tmp="$(_make_tmp)"
  echo "0.0.0" > "$tmp/VERSION"
  # No commands/ or skills/ directory — place a fake script inside tmp
  touch "$tmp/fake_install.sh"

  result="$(
    _define_detect_wrapper
    _detect_script_dir_with_src "$tmp/fake_install.sh"
  )"
  if [[ -z "$result" ]]; then
    return 0
  else
    _fail "Expected empty string, got '$result'"
  fi
}

test_detect_script_dir_returns_repo_when_skills_dir_exists() {
  local tmp result
  tmp="$(_make_tmp)"
  echo "0.0.0" > "$tmp/VERSION"
  mkdir -p "$tmp/skills"
  touch "$tmp/fake_install.sh"

  result="$(
    _define_detect_wrapper
    _detect_script_dir_with_src "$tmp/fake_install.sh"
  )"
  if [[ "$result" == "$tmp" ]]; then
    return 0
  else
    _fail "Expected '$tmp', got '$result'"
  fi
}

# ---------------------------------------------------------------------------
# F. Error guards (without network)
# ---------------------------------------------------------------------------

test_error_no_curl_exits_with_message() {
  local output ec
  _run_installer_no_cmd "curl" output ec
  assert_exit_code 1 "$ec"
  assert_contains "$output" "curl is required"
}

# NOTE: This test relies on install.sh checking for tar *before* making any
# network calls inside _download_release().  If the order in install.sh ever
# changes so that curl is called before tar is checked, this test would hang
# or fail for a different reason (network request with a missing tar).
test_error_no_tar_exits_with_message() {
  local output ec
  _run_installer_no_cmd "tar" output ec
  assert_exit_code 1 "$ec"
  assert_contains "$output" "tar is required"
}

# ---------------------------------------------------------------------------
# G. Statusline installation (--with-statusline)
# ---------------------------------------------------------------------------

# G1. Default install (no flag) does NOT add statusLine key — global.
test_global_default_install_no_statusline() {
  local tmp home_dir output ec settings real_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  # Inject a real mutation-capable crafter binary so hook registration writes
  # settings.json (without a real binary the hook registration is skipped and
  # settings.json is never created, making the subsequent assertion vacuous).
  real_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$real_go_bin:$PATH"
  _run_installer "$home_dir" "$tmp" output ec --global
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"
  # settings.json is created by the hook registration; statusLine must be absent.
  assert_file_exists "$home_dir/.claude/settings.json"
  settings="$(cat "$home_dir/.claude/settings.json")"
  if [[ "$settings" == *'"statusLine"'* ]]; then
    _fail "test_global_default_install_no_statusline: settings.json contains statusLine but flag was not passed"
  fi
}

# G1. Default install (no flag) does NOT add statusLine key — local.
test_local_default_install_no_statusline() {
  local tmp home_dir proj_dir output ec settings
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  _run_installer "$home_dir" "$proj_dir" output ec --local
  assert_exit_code 0 "$ec"
  # Local statusline writes to the project .claude/settings.json.
  # That file is only created by install_statusline (hook uses HOME's settings.json).
  # When the flag is absent, the project settings.json must not exist OR lack statusLine.
  local proj_settings="$proj_dir/.claude/settings.json"
  if [[ -f "$proj_settings" ]]; then
    settings="$(cat "$proj_settings")"
    if [[ "$settings" == *'"statusLine"'* ]]; then
      _fail "test_local_default_install_no_statusline: project settings.json contains statusLine but flag was not passed"
    fi
  fi
}

# G2. --with-statusline on a clean settings.json SETS the statusLine key — global.
test_global_with_statusline_sets_statusline() {
  local tmp home_dir output ec settings fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  # Inject a real-go shim so _download_cli_binary places a mutation-capable
  # crafter binary; without a real binary install_statusline skips the write.
  fake_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$tmp" output ec --global --with-statusline
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"
  assert_file_exists "$home_dir/.claude/settings.json"
  settings="$(cat "$home_dir/.claude/settings.json")"
  assert_contains "$settings" '"statusLine"'
  assert_contains "$settings" '"type": "command"'
  # Assert the exact command value: the written JSON must contain the quoted binary path
  # followed by " statusline". In raw JSON the inner double quotes are escaped as \",
  # so the raw file contains: crafter/bin/crafter\" statusline
  assert_contains "$settings" 'crafter/bin/crafter\" statusline'
}

# G2. --with-statusline on a clean settings.json SETS the statusLine key — local.
test_local_with_statusline_sets_statusline() {
  local tmp home_dir proj_dir output ec settings fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  # Inject a real-go shim so _download_cli_binary places a mutation-capable
  # crafter binary; without a real binary install_statusline skips the write.
  fake_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$proj_dir" output ec --local --with-statusline
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"
  local proj_settings="$proj_dir/.claude/settings.json"
  assert_file_exists "$proj_settings"
  settings="$(cat "$proj_settings")"
  assert_contains "$settings" '"statusLine"'
  assert_contains "$settings" '"type": "command"'
  # Assert the exact command value: the written JSON must contain the quoted binary path
  # followed by " statusline". In raw JSON the inner double quotes are escaped as \",
  # so the raw file contains: crafter/bin/crafter\" statusline
  assert_contains "$settings" 'crafter/bin/crafter\" statusline'
}

# G3. --with-statusline when statusLine already exists does NOT overwrite — global.
test_global_with_statusline_no_overwrite_on_collision() {
  local tmp home_dir output ec settings sentinel fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir/.claude"
  sentinel="my-existing-statusline-sentinel"
  # Pre-seed settings.json with an existing statusLine value.
  printf '{"statusLine":{"type":"command","command":"%s"}}\n' "$sentinel" \
    > "$home_dir/.claude/settings.json"
  # Use real-go shim so the crafter binary is placed and can report the collision.
  fake_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$tmp" output ec --global --with-statusline
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"
  settings="$(cat "$home_dir/.claude/settings.json")"
  assert_contains "$settings" "$sentinel"
  # Collision guidance should appear in stdout/stderr output.
  assert_contains "$output" "statusLine already set"
  # JSON structural integrity: the statusLine command value must still be the
  # sentinel (foreign-keep path must not rewrite the statusLine key).
  # Accept both compact (`"command":"<sentinel>"`) and pretty (`"command": "<sentinel>"`)
  # formats since node may or may not reformat the file.
  if ! printf '%s\n' "$settings" | grep -qF "\"$sentinel\""; then
    _fail "test_global_with_statusline_no_overwrite_on_collision: sentinel '$sentinel' missing from settings.json"
  fi
  if ! printf '%s\n' "$settings" | grep -qF '"statusLine"'; then
    _fail "test_global_with_statusline_no_overwrite_on_collision: statusLine key missing from settings.json"
  fi
}

# G3. --with-statusline when statusLine already exists does NOT overwrite — local.
test_local_with_statusline_no_overwrite_on_collision() {
  local tmp home_dir proj_dir output ec settings sentinel fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir/.claude"
  sentinel="my-existing-statusline-sentinel"
  # Pre-seed the project settings.json with an existing statusLine value.
  printf '{"statusLine":{"type":"command","command":"%s"}}\n' "$sentinel" \
    > "$proj_dir/.claude/settings.json"
  # Use real-go shim so the crafter binary is placed and can report the collision.
  fake_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$proj_dir" output ec --local --with-statusline
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"
  settings="$(cat "$proj_dir/.claude/settings.json")"
  assert_contains "$settings" "$sentinel"
  # Collision guidance should appear in stdout/stderr output.
  assert_contains "$output" "statusLine already set"
  # JSON structural integrity: the statusLine command value must still be the
  # sentinel (foreign-keep path must not rewrite the statusLine key).
  # Local settings.json is only touched by install_statusline (not by node hook
  # installer which writes to HOME settings); it may be compact or pretty-printed.
  # Assert the sentinel string appears within a "command" value context, accepting
  # both `"command":"<sentinel>"` (compact) and `"command": "<sentinel>"` (pretty).
  if ! printf '%s\n' "$settings" | grep -qF "\"$sentinel\""; then
    _fail "test_local_with_statusline_no_overwrite_on_collision: sentinel '$sentinel' missing from settings.json"
  fi
  if ! printf '%s\n' "$settings" | grep -qF '"statusLine"'; then
    _fail "test_local_with_statusline_no_overwrite_on_collision: statusLine key missing from settings.json"
  fi
}

# G4. --with-statusline is idempotent — second run does not corrupt settings — global.
test_global_with_statusline_is_idempotent() {
  local tmp home_dir output ec settings count fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  # Use real-go shim for both runs: install_to deletes the crafter dir on each
  # run, so the binary must be re-placed by the shim each time.
  fake_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$tmp" output ec --global --with-statusline
  assert_exit_code 0 "$ec"
  # Second run: ours-identical → noop (no overwrite, no duplicate key).
  _run_installer "$home_dir" "$tmp" output ec --global --with-statusline
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"
  settings="$(cat "$home_dir/.claude/settings.json")"
  # Exactly one statusLine key must exist.
  count="$(printf '%s' "$settings" | grep -o '"statusLine"' | wc -l | tr -d ' ')"
  if [[ "$count" -ne 1 ]]; then
    _fail "test_global_with_statusline_is_idempotent: expected 1 statusLine key, got $count"
  fi
  assert_contains "$settings" 'crafter'
  assert_contains "$settings" 'statusline'
}

# G4. --with-statusline is idempotent — second run does not corrupt settings — local.
test_local_with_statusline_is_idempotent() {
  local tmp home_dir proj_dir output ec settings count fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  proj_dir="$tmp/project"
  mkdir -p "$home_dir" "$proj_dir"
  # Use real-go shim for both runs: install_to deletes the crafter dir on each
  # run, so the binary must be re-placed by the shim each time.
  fake_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$proj_dir" output ec --local --with-statusline
  assert_exit_code 0 "$ec"
  # Second run: ours-identical → noop (no overwrite, no duplicate key).
  _run_installer "$home_dir" "$proj_dir" output ec --local --with-statusline
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"
  local proj_settings="$proj_dir/.claude/settings.json"
  settings="$(cat "$proj_settings")"
  # Exactly one statusLine key must exist.
  count="$(printf '%s' "$settings" | grep -o '"statusLine"' | wc -l | tr -d ' ')"
  if [[ "$count" -ne 1 ]]; then
    _fail "test_local_with_statusline_is_idempotent: expected 1 statusLine key, got $count"
  fi
  assert_contains "$settings" 'crafter'
  assert_contains "$settings" 'statusline'
}

# G5. Collision guidance stdout is valid JSON whose .statusLine.command contains
#     both the existing command and the crafter statusline invocation.
test_collision_guidance_is_valid_json() {
  local tmp home_dir output ec sentinel guidance_json cmd_line cmd_value fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir/.claude"
  sentinel="some-existing-statusline-tool"
  # Pre-seed settings.json with an existing statusLine so a collision is triggered.
  printf '{"statusLine":{"type":"command","command":"%s"}}\n' "$sentinel" \
    > "$home_dir/.claude/settings.json"
  # Use real-go shim so the crafter binary is placed and can emit guidance.
  fake_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$tmp" output ec --global --with-statusline
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"

  # Extract the guidance JSON block using pure-bash awk brace-depth counting so
  # that command values containing literal braces do not cause premature exit.
  guidance_json="$(_extract_guidance_json "$output")"

  if [[ -z "$guidance_json" ]]; then
    _fail "test_collision_guidance_is_valid_json: could not extract statusLine JSON block from installer output"
    return
  fi

  # Structural check: the block must start with '{' and end with '}'.
  local trimmed="${guidance_json#"${guidance_json%%[! ]*}"}"   # ltrim
  if [[ "${trimmed:0:1}" != "{" ]]; then
    _fail "test_collision_guidance_is_valid_json: guidance block does not start with '{': $trimmed"
    return
  fi

  # Extract the "command" value from the guidance JSON and decode JSON escapes.
  # The Go encoder produces exactly one "command": "..." line in the MarshalIndent output.
  cmd_line="$(printf '%s\n' "$guidance_json" | grep '"command":')"
  cmd_value="$(_decode_json_cmd_line "$cmd_line")"

  if [[ -z "$cmd_value" ]]; then
    _fail "test_collision_guidance_is_valid_json: could not extract .command from guidance JSON"
    return
  fi

  # The composite command must reference the existing tool.
  if [[ "$cmd_value" != *"$sentinel"* ]]; then
    _fail "test_collision_guidance_is_valid_json: .command does not contain existing command '$sentinel': $cmd_value"
  fi
  # The composite command must reference the crafter statusline invocation.
  if [[ "$cmd_value" != *"statusline"* ]]; then
    _fail "test_collision_guidance_is_valid_json: .command does not contain 'statusline': $cmd_value"
  fi
}

# G6. Collision guidance: existing command containing a single quote produces
#     a syntactically valid composite bash command (regression for #1).
test_collision_guidance_single_quote_in_existing_cmd_is_valid_shell() {
  local tmp home_dir output ec guidance_json cmd_line cmd_value fake_go_bin old_path result_ec
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir/.claude"

  # Seed an existing statusLine.command that contains single quotes — the
  # exact pattern that breaks an unescaped bash -c '...' wrapper.
  # We JSON-encode the value (single quote has no special meaning in JSON).
  printf '{"statusLine":{"type":"command","command":"awk '"'"'{print $1}'"'"'"}}\n' \
    > "$home_dir/.claude/settings.json"

  # Use real-go shim so the crafter binary is placed and can emit guidance.
  fake_go_bin="$(_make_real_go_bin_dir)"
  old_path="$PATH"
  PATH="$fake_go_bin:$PATH"
  _run_installer "$home_dir" "$tmp" output ec --global --with-statusline
  result_ec=$ec
  PATH="$old_path"
  assert_exit_code 0 "$result_ec"

  # Extract the guidance JSON block using pure-bash awk brace-depth counting.
  guidance_json="$(_extract_guidance_json "$output")"

  if [[ -z "$guidance_json" ]]; then
    _fail "test_collision_guidance_single_quote_in_existing_cmd_is_valid_shell: could not extract guidance JSON"
    return
  fi

  # Extract and decode the .command value from the guidance JSON.
  # The Go encoder (json.MarshalIndent) places "command": "..." on one line.
  cmd_line="$(printf '%s\n' "$guidance_json" | grep '"command":')"
  cmd_value="$(_decode_json_cmd_line "$cmd_line")"

  if [[ -z "$cmd_value" ]]; then
    _fail "test_collision_guidance_single_quote_in_existing_cmd_is_valid_shell: could not decode .command from guidance JSON"
    return
  fi

  # The extracted .command must be syntactically valid bash — no unexpected EOF.
  #
  # WHY we EXECUTE rather than use `bash -n -c`:
  # bash -n -c "$cmd_value" only parses the OUTER level of the command — it sees
  # the tokens `bash`, `-c`, and a string argument, and returns 0 without ever
  # parsing the INNER single-quoted pipeline.  A pre-fix unescaped single quote
  # inside that pipeline (e.g. `awk '{print $1}'`) would make bash -n exit 0
  # for both the broken and the correct command, giving false confidence.
  #
  # By actually EXECUTING with `bash -c "$cmd_value"`, bash parses and runs the
  # inner single-quoted pipeline.  A broken inner quote (pre-fix) causes bash to
  # abort with "unexpected EOF while looking for matching" / "syntax error" and
  # exit 2.  A correct (post-fix, '\''…'\''-escaped) command runs normally and
  # exits 0 even if the crafter binary is absent (a missing-binary error is exit
  # 127 caught inside the inner $(...) substitution, NOT a syntax error).
  #
  # Discrimination proof (verified manually):
  #   pre-fix  (awk '{print $1}' embedded raw) → exit 2, stderr contains "syntax error"
  #   post-fix (awk '\''{ print $1 }'\''  escaped) → exit 0, no syntax error in stderr
  local exec_output="" exec_ec=0
  exec_output="$(printf 'hello world\n' | bash -c "$cmd_value" 2>&1)" || exec_ec=$?

  # A shell SYNTAX error from a broken inner quote exits with code 2 and prints
  # "syntax error" / "unexpected" in stderr.  Any other non-zero exit (e.g. 127
  # for a missing crafter binary) is acceptable — it is a runtime error, not a
  # syntax error.
  if [[ "$exec_ec" -eq 2 ]]; then
    _fail "test_collision_guidance_single_quote_in_existing_cmd_is_valid_shell: composite .command has shell syntax error (exit $exec_ec): $cmd_value"
    return
  fi
  if [[ "$exec_output" == *"syntax error"* || "$exec_output" == *"unexpected EOF"* || "$exec_output" == *"unexpected token"* ]]; then
    _fail "test_collision_guidance_single_quote_in_existing_cmd_is_valid_shell: composite .command produced syntax-error output: $exec_output"
  fi
}

# G7. --with-statusline when the crafter binary is absent: warns, exits 0,
#     and leaves settings unchanged (binary-absent posture from Step 2.2).
test_with_statusline_crafter_binary_missing_warns_and_skips() {
  local tmp home_dir output ec settings
  tmp="$(_make_tmp)"
  home_dir="$tmp/home"
  mkdir -p "$home_dir"
  if [[ -f "$HOME/.tool-versions" && ! -f "$home_dir/.tool-versions" ]]; then
    cp "$HOME/.tool-versions" "$home_dir/.tool-versions"
  fi
  # Run with a minimal PATH (no go, mise, or asdf) so the binary cannot be
  # built.  _download_cli_binary fails gracefully and install_statusline sees
  # a missing binary → prints the warning and exits 0.
  local _ec=0
  output="$(cd "$tmp" && HOME="$home_dir" PATH="/usr/local/bin:/usr/bin:/bin" \
    bash "$INSTALL_SH" --global --with-statusline 2>&1)" \
    && _ec=$? || _ec=$?
  assert_exit_code 0 "$_ec"
  # The installer must emit the binary-not-found warning.
  assert_contains "$output" "crafter binary not found"
  # settings.json must NOT gain a statusLine key (binary absent → skip).
  local settings_file="$home_dir/.claude/settings.json"
  if [[ -f "$settings_file" ]]; then
    settings="$(cat "$settings_file")"
    if [[ "$settings" == *'"statusLine"'* ]]; then
      _fail "test_with_statusline_crafter_binary_missing_warns_and_skips: settings.json contains statusLine but crafter binary was absent"
    fi
  fi
}

# ---------------------------------------------------------------------------
# Syntax check
# ---------------------------------------------------------------------------
test_install_sh_passes_bash_syntax_check() {
  local output ec
  output="$(bash -n "$INSTALL_SH" 2>&1)" && ec=$? || ec=$?
  assert_exit_code 0 "$ec"
}

# ---------------------------------------------------------------------------
# Main: discover and run all test_* functions
# ---------------------------------------------------------------------------
main() {
  echo "Running install.sh test suite..."
  echo ""

  local tests t
  tests="$(declare -F | awk '{print $3}' | grep '^test_')"

  while IFS= read -r t; do
    run_test "$t"
  done <<< "$tests"

  echo ""
  echo "---------------------------------------"
  if [[ $_FAIL -eq 0 ]]; then
    printf "${_GREEN}%d passed, %d failed${_RESET}\n" "$_PASS" "$_FAIL"
  else
    printf "${_RED}%d passed, %d failed${_RESET}\n" "$_PASS" "$_FAIL"
  fi
  echo "---------------------------------------"

  [[ $_FAIL -eq 0 ]]
}

main "$@"
