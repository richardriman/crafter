# Task: Clean install directory before copying files

## Metadata
- **Date:** 2026-02-27
- **Branch:** main
- **Status:** active
- **Scope:** Small

## Request
Pri instalaci craftera se nepromaze slozka crafter v cilove lokaci. To zpusobuje, ze pokud jsou nejake soubory z metody v case odstraneny, nebo jsou presunuty/nahrazeny jinymy, v install location zahnivaji. Mimo tam napr. nacitalo meta-promps, ktere jsou davno nahrazene agenty.

## Plan
**Plan status:** approved

The problem: `install_to()` in `install.sh` uses `mkdir -p` + `cp` to copy files into `$base/crafter/`, `$base/commands/crafter/`, and `$base/agents/`. When files are removed or renamed between Crafter versions (e.g., old meta-prompts replaced by agents), the stale files persist in the install location because nothing cleans the directories before copying.

The fix: Add `rm -rf` calls for the three managed directories (`crafter/`, `commands/crafter/`, `agents/`) at the top of `install_to()`, before the `mkdir -p` + `cp` block. These directories are fully managed by Crafter — they contain no user files — so removing them entirely before copying is safe and correct.

**Why clean all three directories (not just `crafter/`):** The `commands/crafter/` and `agents/` directories are also fully managed. If a command or agent file is renamed or removed in a future version, the same stale-file problem would occur there too.

**Why `rm -rf` + fresh `mkdir -p` rather than selective deletion:** It's simpler, more robust, and guarantees a clean state. There's no need to diff old vs. new file lists. The `mkdir -p` calls immediately after recreate the directories.

**First install (no existing dir):** `rm -rf` on a non-existent path is a no-op, so first installs are unaffected.

**Alternatives considered:**
- Selective `find + delete` of files not in a known list — overly complex for no benefit since these dirs are fully managed.
- Only cleaning `crafter/` — would miss stale files in `commands/crafter/` and `agents/`.

### Stage 1 — Implement the fix and add tests

- [ ] **Step 1: Add directory cleanup to `install_to()` in `install.sh`** (file: `/Users/ret/dev/ai/crafter/install.sh`, lines 149-191)

  Insert three `rm -rf` calls at the beginning of `install_to()`, after the variable declarations (line 157) and before the first `mkdir -p` (line 161). Add them in this order:
  ```bash
  # Clean previously installed files to prevent stale leftovers on upgrade
  rm -rf "$commands_dest" "$crafter_dest" "$agents_dest"
  ```
  This goes between line 157 (`local agents_dest=...`) and line 159 (`echo "Installing Crafter $label..."`). The single `rm -rf` line removes all three directories. The subsequent `mkdir -p` calls will recreate them.

  Note: The `agents_dest` variable points to `$base/agents` which is a shared namespace (other tools could put agent files there too). However, looking at the current design, Crafter agent files all have the `crafter-` prefix. To be safe and avoid deleting non-Crafter agents in the `agents/` directory, we should NOT `rm -rf` the entire `$agents_dest` directory. Instead, remove only Crafter's agent files with a glob pattern:
  ```bash
  rm -rf "$commands_dest" "$crafter_dest"
  rm -f "$agents_dest"/crafter-*.md
  ```
  This is safer: `commands/crafter/` and `crafter/` are fully namespaced to Crafter, but `agents/` might contain other agent files if the user has other tools installed.

- [ ] **Step 2: Add `assert_file_not_exists` helper to test file** (file: `/Users/ret/dev/ai/crafter/tests/test_install.sh`, after `assert_file_nonempty` at line 105)

  Add a new assertion helper after `assert_file_nonempty()` (line 98-105):
  ```bash
  assert_file_not_exists() {
    local path="$1"
    if [[ ! -e "$path" ]]; then
      return 0
    else
      _fail "assert_file_not_exists: '$path' exists but should have been removed"
      return 1
    fi
  }
  ```

- [ ] **Step 3: Add test for stale file cleanup** (file: `/Users/ret/dev/ai/crafter/tests/test_install.sh`, in a new section after the Idempotency section, around line 468)

  Add a new test section between the Idempotency section (D) and the Hook installation section (E). This test verifies that files from a previous install that no longer exist in the source are removed on upgrade:
  ```bash
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
    echo "stale" > "$base/commands/crafter/old-command.md"
    echo "stale" > "$base/agents/crafter-old-agent.md"

    # Second install (upgrade)
    _run_installer "$home_dir" "$tmp" output ec --global
    assert_exit_code 0 "$ec"

    # Stale files must be gone
    assert_file_not_exists "$base/crafter/rules/old-rule.md"
    assert_file_not_exists "$base/crafter/templates/old-template.md"
    assert_file_not_exists "$base/commands/crafter/old-command.md"
    assert_file_not_exists "$base/agents/crafter-old-agent.md"

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
    echo "stale" > "$base/commands/crafter/old-command.md"
    echo "stale" > "$base/agents/crafter-old-agent.md"

    # Second install (upgrade)
    _run_installer "$home_dir" "$proj_dir" output ec --local
    assert_exit_code 0 "$ec"

    # Stale files must be gone
    assert_file_not_exists "$base/crafter/rules/old-rule.md"
    assert_file_not_exists "$base/commands/crafter/old-command.md"
    assert_file_not_exists "$base/agents/crafter-old-agent.md"

    # Current files must still be present
    for rel in "${_EXPECTED_FILES_REL[@]}"; do
      assert_file_exists "$base/$rel"
    done
  }

  test_upgrade_preserves_non_crafter_agents() {
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

    # Second install (upgrade)
    _run_installer "$home_dir" "$tmp" output ec --global
    assert_exit_code 0 "$ec"

    # Non-Crafter agent must survive
    assert_file_exists "$base/agents/other-tool-agent.md"
  }
  ```

- [ ] **Step 4: Run the test suite and verify all tests pass**

  Run `bash /Users/ret/dev/ai/crafter/tests/test_install.sh` and confirm:
  - All existing tests still pass (especially idempotency tests)
  - The three new tests (`test_global_upgrade_removes_stale_files`, `test_local_upgrade_removes_stale_files`, `test_upgrade_preserves_non_crafter_agents`) pass
  - `bash -n install.sh` syntax check passes

### Verification criteria

1. Running `bash tests/test_install.sh` produces 0 failures, including the new stale-file tests.
2. `bash -n install.sh` passes (syntax check — already covered by existing test).
3. On a simulated upgrade, files placed in `crafter/`, `commands/crafter/`, and `agents/crafter-*.md` that don't exist in the source are removed.
4. On a simulated upgrade, non-Crafter agent files in `agents/` are preserved.
5. First install (clean directory) works identically to before.

### Unknowns / flags

- The `agents/` directory is a shared namespace (`$base/agents/`). Other Claude Code tools or the user might place agent files there. The plan handles this by only removing `crafter-*.md` files rather than the whole directory. If the naming convention changes in the future, this glob pattern would need updating.

## Decisions

## Outcome
