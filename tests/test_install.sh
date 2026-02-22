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
  # Use _ri_pwd as working directory when provided; fall back to REPO_DIR.
  local _ri_cwd="${_ri_pwd:-$REPO_DIR}"
  _ri_output="$(cd "$_ri_cwd" && HOME="$_ri_home" bash "$INSTALL_SH" "$@" 2>&1)" \
    && _ri_ec=$? || _ri_ec=$?
  printf -v "$_ri_out_var"  '%s' "$_ri_output"
  printf -v "$_ri_ec_var"   '%d' "$_ri_ec"
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
# 22 files installed by install_global / install_to
_EXPECTED_FILES_REL=(
  "commands/crafter/do.md"
  "commands/crafter/debug.md"
  "commands/crafter/status.md"
  "commands/crafter/map-project.md"
  "crafter/VERSION"
  "crafter/rules/core.md"
  "crafter/rules/do-workflow.md"
  "crafter/rules/debug-workflow.md"
  "crafter/rules/delegation.md"
  "crafter/rules/post-change.md"
  "crafter/rules/task-lifecycle.md"
  "crafter/rules/update-check.md"
  "crafter/templates/PROJECT.md"
  "crafter/templates/ARCHITECTURE.md"
  "crafter/templates/STATE.md"
  "crafter/templates/claude-md.snippet"
  "crafter/templates/TASK.md"
  "crafter/meta-prompts/planner.md"
  "crafter/meta-prompts/implement.md"
  "crafter/meta-prompts/verify.md"
  "crafter/meta-prompts/review.md"
  "crafter/meta-prompts/analyze.md"
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
  assert_dir_exists "$home_dir/.claude/commands/crafter"
  assert_dir_exists "$home_dir/.claude/crafter"
  assert_dir_exists "$home_dir/.claude/crafter/rules"
  assert_dir_exists "$home_dir/.claude/crafter/templates"
  assert_dir_exists "$home_dir/.claude/crafter/meta-prompts"
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
  assert_dir_exists "$proj_dir/.claude/commands/crafter"
  assert_dir_exists "$proj_dir/.claude/crafter"
  assert_dir_exists "$proj_dir/.claude/crafter/rules"
  assert_dir_exists "$proj_dir/.claude/crafter/templates"
  assert_dir_exists "$proj_dir/.claude/crafter/meta-prompts"
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
# E. _detect_script_dir
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

test_detect_script_dir_empty_when_no_commands_dir() {
  local tmp result
  tmp="$(_make_tmp)"
  echo "0.0.0" > "$tmp/VERSION"
  # No commands/ directory — place a fake script inside tmp
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
