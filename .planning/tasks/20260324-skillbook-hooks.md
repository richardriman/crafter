# Task: Skillbook — Self-Learning Agents via Go CLI

**Date:** 2026-03-24
**Status:** active
**Scope:** Large

## Request

Implement a skillbook system for Crafter that allows agents to learn from experience across sessions. After each task completion, observations (patterns, mistakes, successes) are extracted and stored in a project-level `skillbook.json`. Before agent invocation, relevant skills are injected into the agent's prompt.

Architecture decision: build a `crafter` Go CLI binary with `skillbook` subcommands. The binary handles all deterministic work (JSON CRUD, Jaccard dedup, confidence promotion, atomic writes) that LLMs do poorly. Orchestration stays in markdown prompts — the orchestrator calls the CLI via Bash and uses its output. Inspired by Nightshift's `skillbook.ts` + `reflector.ts`.

Key constraints:
- Crafter orchestration remains prompt-only — the CLI is a utility tool, not orchestration
- Skillbook is project-level (`.planning/skillbook.json`), not global
- Go CLI is a single static binary with zero runtime dependencies
- Cross-compiled for darwin-arm64, darwin-amd64, linux-amd64, linux-arm64
- Distributed via GitHub releases, downloaded by `install.sh`
- No Node.js dependency (Claude Code installer no longer requires npm)

## Plan

**Plan status:** approved

### What and Why

Crafter agents start each task with a blank slate — they have no memory of what worked or failed in previous tasks on the same project. The skillbook system adds project-level learning: observations from completed tasks are stored in `.planning/skillbook.json` and relevant skills are injected into agent prompts before they run.

The architecture went through several iterations (hooks, prompt-only, Node.js script) before settling on a **Go CLI binary**. The reasoning:

- **Hooks (rejected):** PreToolUse hooks cannot inject into spawned agent contexts; PostToolUse parsing of free-form LLM output is fragile.
- **Prompt-only (rejected):** Token-hungry (LLM reads full JSON, computes dedup, writes back). LLM-based Jaccard is non-deterministic.
- **Node.js utility (rejected):** Claude Code's recommended install path is `curl | bash`, not npm. Cannot assume Node.js is available.
- **Go CLI (selected):** Single static binary, zero runtime dependencies, deterministic, cross-compiled for all platforms.

The Go binary is called `crafter` and has `skillbook` as its first subcommand, but is designed for future extensibility (model profiles, state management, etc.).

**How it works:**
1. **Injection:** Before spawning an agent, the orchestrator runs `crafter skillbook get --agent <name> --file {PROJECT_PATH}/.planning/skillbook.json` via Bash. The CLI outputs a small markdown block (~100-200 tokens) that the orchestrator appends to the agent's task prompt. Zero LLM token cost for JSON parsing, filtering, sorting, or formatting.
2. **Capture:** After task completion, the orchestrator reflects and formulates observations, then runs `crafter skillbook add --agent <name> --rule "..." --rationale "..." --task "<taskfile>" --file {PROJECT_PATH}/.planning/skillbook.json` for each observation. The CLI handles Jaccard dedup, confidence promotion, and atomic JSON write. The orchestrator never reads or writes skillbook.json directly.
3. **Init:** `crafter skillbook init --file {PROJECT_PATH}/.planning/skillbook.json` creates an empty skillbook.

### Alternatives Considered

Covered in the "What and Why" section above — hooks, prompt-only, and Node.js approaches were all explored and rejected for the reasons stated.

### Files Affected

**New files:**
- `cli/` — entire new Go module directory:
  - `cli/go.mod`, `cli/go.sum` — Go module definition
  - `cli/main.go` — entry point, root cobra command
  - `cli/cmd/root.go` — root command setup
  - `cli/cmd/skillbook.go` — `skillbook` parent command
  - `cli/cmd/skillbook_get.go` — `skillbook get` subcommand
  - `cli/cmd/skillbook_add.go` — `skillbook add` subcommand
  - `cli/cmd/skillbook_init.go` — `skillbook init` subcommand
  - `cli/internal/skillbook/types.go` — JSON schema types
  - `cli/internal/skillbook/store.go` — read/write/atomic file operations
  - `cli/internal/skillbook/jaccard.go` — Jaccard similarity + tokenization
  - `cli/internal/skillbook/format.go` — markdown formatting for `get` output
  - `cli/internal/skillbook/jaccard_test.go` — tests for Jaccard similarity
  - `cli/internal/skillbook/store_test.go` — tests for store operations (add with dedup, confidence promotion)
  - `cli/internal/skillbook/format_test.go` — tests for markdown formatting
  - `cli/Makefile` — cross-compilation targets

**Modified files:**
- `rules/delegation.md` — add skill injection instructions using `crafter skillbook get`
- `rules/post-change.md` — add observation extraction using `crafter skillbook add`
- `commands/do.md` — update mandatory checklist to include skillbook step
- `install.sh` — add binary download (platform detection, `~/.claude/crafter/bin/crafter`)
- `.planning/ARCHITECTURE.md` — document the CLI and skillbook system
- `.planning/PROJECT.md` — add Go to the stack
- `tests/test_install.sh` — add test for bin directory creation

### Stages

---

### Stage 1 — Go CLI (core logic)

Build the `crafter` Go CLI binary with `skillbook` subcommands. This stage produces a working, tested binary that can be built locally with `go build`.

- [x] **Step 1: Initialize Go module and set up project structure.**

  Create the `cli/` directory at the repo root (`/Users/ret/dev/ai/crafter/cli/`).

  **`cli/go.mod`:** Initialize with `module github.com/richardriman/crafter/cli`, Go 1.22+ (for consistency with modern toolchains). Add `github.com/spf13/cobra` as the only external dependency.

  **`cli/main.go`:** Minimal entry point that calls `cmd.Execute()`.

  **`cli/cmd/root.go`:** Root cobra command `crafter` with a short description ("Crafter CLI — utilities for AI development workflow"). No persistent flags on root yet — the `--file` flag is on the `skillbook` subcommand.

  **`cli/cmd/skillbook.go`:** Parent `skillbook` command that groups subcommands. Add a persistent `--file` flag (required, string, path to skillbook.json) on this command so all subcommands inherit it.

  After this step, `cd cli && go build -o crafter .` should produce a binary that shows help text.

- [x] **Step 2: Implement JSON schema types and file I/O (store layer).**

  **`cli/internal/skillbook/types.go`:** Define the skillbook JSON schema as Go structs:

  ```go
  type Skillbook struct {
      Version   int     `json:"version"`
      UpdatedAt string  `json:"updatedAt"`
      Disabled  bool    `json:"disabled,omitempty"`
      Skills    []Skill `json:"skills"`
  }

  type Skill struct {
      ID           string `json:"id"`
      Agent        string `json:"agent"`
      Rule         string `json:"rule"`
      Rationale    string `json:"rationale"`
      SourceTask   string `json:"sourceTask"`
      Confidence   string `json:"confidence"`   // "low", "medium", "high"
      AddedAt      string `json:"addedAt"`
      AppliedCount int    `json:"appliedCount"`
      Deprecated   bool   `json:"deprecated"`
  }
  ```

  Key differences from Nightshift: `sourceRunId` becomes `sourceTask` (Crafter uses task files, not run IDs). No `deprecatedReason` field for simplicity.

  **`cli/internal/skillbook/store.go`:** File operations:
  - `Load(path string) (*Skillbook, error)` — read and unmarshal JSON. Return empty skillbook if file does not exist.
  - `Save(path string, sb *Skillbook) error` — atomic write: marshal JSON with indent, write to `path + ".tmp"`, `os.Rename` to `path`. Update `updatedAt` timestamp before writing.
  - `NewSkillbook() *Skillbook` — create empty skillbook with version 1.

- [x] **Step 3: Implement Jaccard similarity and the `skillbook add` subcommand.**

  **`cli/internal/skillbook/jaccard.go`:**
  - `Tokenize(text string) map[string]struct{}` — lowercase, strip non-alphanumeric (keep hyphens), split on whitespace, discard tokens with length <= 2. Return as a set.
  - `JaccardSimilarity(a, b map[string]struct{}) float64` — compute |intersection| / |union|. Return 0.0 if both sets are empty.
  - `FindDuplicate(skills []Skill, agent, rule string, threshold float64) (int, bool)` — find the first non-deprecated skill for the given agent with Jaccard similarity > threshold. Return index and found flag.

  **`cli/cmd/skillbook_add.go`:** The `skillbook add` subcommand with flags:
  - `--agent` (required, string) — agent name (e.g., "implementer")
  - `--rule` (required, string) — the learned guideline text
  - `--rationale` (required, string) — why this was observed
  - `--task` (required, string) — source task filename
  - `--file` (inherited from parent)

  Logic:
  1. Load skillbook from `--file` (create if not exists).
  2. If `disabled` is true, print message to stderr and exit 0.
  3. Call `FindDuplicate` with threshold 0.6.
  4. If duplicate found: append rationale (separated by "; "), promote confidence (low->medium, medium->high, high stays), print "Merged with existing skill: <id>" to stdout.
  5. If no duplicate: create new skill with UUID, confidence "low", appliedCount 0, print "Added new skill: <id>" to stdout.
  6. Atomic save.

  For UUID generation: use `crypto/rand` to generate UUID v4 — no external dependency needed.

- [x] **Step 4: Implement `skillbook get` and `skillbook init` subcommands.**

  **`cli/internal/skillbook/format.go`:**
  - `FormatMarkdown(skills []Skill, agent string, limit int) string` — filter by agent (non-deprecated), sort by confidence descending (high=3, medium=2, low=1) then appliedCount descending, take top `limit` (default 10). Format as:
    ```
    ## Learned Guidelines (from project skillbook)

    - **IMPORTANT:** <rule text>
    - **Guideline:** <rule text>
    - **Consider:** <rule text>
    ```
    Return empty string if no skills match (so the orchestrator knows to skip injection).
  - `IncrementApplied(sb *Skillbook, agent string, limit int) []string` — find the same skills that `FormatMarkdown` would select, increment their `appliedCount`, return their IDs. This is called after `get` to update counts.

  **`cli/cmd/skillbook_get.go`:** The `skillbook get` subcommand with flags:
  - `--agent` (required, string)
  - `--limit` (optional, int, default 10)
  - `--file` (inherited)

  Logic:
  1. Load skillbook from `--file`. If file does not exist, exit 0 silently (no output).
  2. If `disabled` is true, exit 0 silently.
  3. Call `FormatMarkdown`. Print result to stdout.
  4. Call `IncrementApplied`. Atomic save to update the appliedCount values.
  5. If no skills match, print nothing and exit 0.

  **`cli/cmd/skillbook_init.go`:** The `skillbook init` subcommand:
  - `--file` (inherited)

  Logic:
  1. If file already exists, print "Skillbook already exists at <path>" to stderr and exit 0.
  2. Create parent directories if needed (`os.MkdirAll`).
  3. Create new empty skillbook via `NewSkillbook()`.
  4. Save to `--file`.
  5. Print "Created empty skillbook at <path>" to stdout.

- [x] **Step 5: Write tests for the Go code.**

  **`cli/internal/skillbook/jaccard_test.go`:** Test cases:
  - Empty strings return similarity 0.0
  - Identical strings return similarity 1.0
  - Known example: "always use descriptive variable names" vs "use descriptive naming for variables" should be above 0.6
  - Short words (<=2 chars) are stripped: "if a or b" produces empty token set
  - `FindDuplicate` returns correct index when above threshold, returns false when below

  **`cli/internal/skillbook/store_test.go`:** Test cases:
  - `Load` of non-existent file returns empty skillbook (not error)
  - `Save` + `Load` round-trip preserves all fields
  - Atomic write: if the directory exists, `.tmp` file is cleaned up
  - `NewSkillbook` returns version 1 with empty skills slice

  **`cli/internal/skillbook/format_test.go`:** Test cases:
  - No matching skills returns empty string
  - Skills are sorted by confidence then appliedCount
  - Only top N skills are returned (test with limit=2 when 5 exist)
  - Deprecated skills are excluded
  - Confidence prefixes: high -> IMPORTANT, medium -> Guideline, low -> Consider
  - `IncrementApplied` increments only selected skills

  Run with: `cd cli && go test ./...`

---

### Stage 2 — Build and Distribution

Set up cross-compilation and update the install script to download and place the binary. This stage makes the binary available to end users.

- [x] **Step 6: Create Makefile for cross-compilation.**

  **`cli/Makefile`:**
  - `VERSION` variable read from `../VERSION`
  - `build` target: `go build -o bin/crafter .` (local platform)
  - `release` target: cross-compile for 4 platform/arch combinations using `GOOS` and `GOARCH`:
    - `darwin/arm64` -> `bin/crafter-darwin-arm64`
    - `darwin/amd64` -> `bin/crafter-darwin-amd64`
    - `linux/amd64` -> `bin/crafter-linux-amd64`
    - `linux/arm64` -> `bin/crafter-linux-arm64`
  - Use `CGO_ENABLED=0` for fully static binaries
  - Use `-ldflags="-s -w"` to strip debug info and reduce binary size
  - `test` target: `go test ./...`
  - `clean` target: `rm -rf bin/`

  Binaries are attached to GitHub releases manually for now (via the existing `/crafter:release` command or `gh release`). CI automation is a future enhancement.

- [x] **Step 7: Update `install.sh` to download and install the CLI binary.**

  At `/Users/ret/dev/ai/crafter/install.sh`:

  Add a new function `_download_cli_binary()` (around line 144, after `_download_release`). This function:
  1. Detects platform: `uname -s | tr '[:upper:]' '[:lower:]'` -> `darwin` or `linux`
  2. Detects architecture: `uname -m` -> map `x86_64` to `amd64`, `arm64`/`aarch64` to `arm64`
  3. Constructs the binary name: `crafter-${os}-${arch}`
  4. Downloads from GitHub releases: `https://github.com/${REPO}/releases/download/${VERSION}/crafter-${os}-${arch}` (only when VERSION is set; skip for main branch installs since binaries are only on release tags)
  5. Places binary at `$1/crafter/bin/crafter` (where `$1` is the base dir, e.g., `~/.claude` or `.claude`)
  6. Makes it executable: `chmod +x`
  7. If download fails (e.g., no binary for this platform/version, or installing from main branch), print a warning but do NOT fail the install — the CLI is optional, and the core Crafter functionality works without it.

  Update `install_to()` function (line 149) to:
  - Create `$crafter_dest/bin/` directory: add `mkdir -p "$crafter_dest/bin"` after line 171
  - For local installs from a clone: check if `$SCRIPT_DIR/cli/bin/crafter` exists (pre-built local binary) and copy it to `$crafter_dest/bin/crafter`.

  Update `tests/test_install.sh`:
  - Add a new test `test_global_creates_bin_directory` that verifies `$base/crafter/bin/` directory is created after install.
  - The `_EXPECTED_FILES_REL` array (line 233) does NOT need the binary added since binary download only happens in remote mode with a version tag, and local clone may not have a pre-built binary.

---

### Stage 3 — Prompt Integration

Wire the CLI into Crafter's orchestrator prompts. This stage makes the skillbook system functional end-to-end.

- [x] **Step 8: Update `rules/delegation.md` — add skill injection instructions.**

  At `/Users/ret/dev/ai/crafter/rules/delegation.md`, after line 35 ("Always include the `model` parameter in every Task tool invocation..."), add a new section:

  ```markdown
  ## Skillbook — Learned Guidelines

  Before spawning any agent via the Task tool, check if the `crafter` CLI binary is available at `~/.claude/crafter/bin/crafter` (or `.claude/crafter/bin/crafter` for local installs). If available:

  1. Run via Bash: `~/.claude/crafter/bin/crafter skillbook get --agent <agent-short-name> --file {PROJECT_PATH}/.planning/skillbook.json`
  2. Agent name mapping: strip the `crafter-` prefix (e.g., `crafter-implementer` -> `implementer`, `crafter-planner` -> `planner`).
  3. If the command produces output (non-empty stdout), append it verbatim to the agent's task prompt. The output is already formatted as a "Learned Guidelines" markdown section.
  4. If the command produces no output, the agent has no learned guidelines — proceed normally without mentioning it.
  5. If the command fails (non-zero exit), log a warning but proceed with agent spawning — skillbook is optional.

  If the CLI binary does not exist, skip skillbook injection silently.
  ```

- [x] **Step 9: Update `rules/post-change.md` — add observation capture.**

  At `/Users/ret/dev/ai/crafter/rules/post-change.md`, after the "Complete Task File" section (line 33, after "complete it per `~/.claude/crafter/rules/task-lifecycle.md`.") and before "## Session Wrap-Up" (line 35), add a new section:

  ```markdown
  ## Update Skillbook

  After completing the task file, reflect on the task and extract observations for the project's skillbook. Only do this if the `crafter` CLI binary is available at `~/.claude/crafter/bin/crafter`.

  1. Review what happened during the task: Did the implementer struggle with something project-specific? Did the reviewer flag a recurring pattern? Did the planner miss something about the project structure?
  2. Formulate 0-3 observations (only if genuinely useful — do not force observations for trivial tasks). Each observation needs:
     - **agent**: which agent this applies to (implementer, reviewer, planner, verifier, analyzer)
     - **rule**: the learned guideline, written as an instruction
     - **rationale**: what happened that led to this observation
  3. For each observation, run via Bash:
     ```
     ~/.claude/crafter/bin/crafter skillbook add \
       --agent "<agent>" \
       --rule "<rule text>" \
       --rationale "<rationale text>" \
       --task "<task-filename>" \
       --file {PROJECT_PATH}/.planning/skillbook.json
     ```
  4. The CLI handles deduplication and confidence promotion automatically. If a similar skill already exists, it will be merged and promoted.
  5. Briefly tell the user what was learned (e.g., "Added 2 observations to the project skillbook: ...").
  6. If the CLI binary is not available or the command fails, skip silently — skillbook is optional.

  Focus on project-specific patterns, not general programming knowledge:
  - Good: "This project uses X pattern for Y", "Tests need Z setup", "The review found A was a recurring issue"
  - Bad: "Always use descriptive variable names" (too generic), "Fixed a typo" (not a pattern)
  ```

- [x] **Step 10: Update `commands/do.md` — update mandatory checklist.**

  At `/Users/ret/dev/ai/crafter/commands/do.md`:

  Lines 174-182 — update the mandatory checklist to add a skillbook step between items 4 and 5 (renumbering accordingly):

  Current (lines 179-180):
  ```
  4. **Complete the task file** — set Status to `completed`, fill in the `## Outcome` section, check off remaining plan steps. The task file is in `{PROJECT_PATH}/.planning/tasks/`.
  5. **Suggest session wrap-up** — ...
  ```

  New:
  ```
  4. **Complete the task file** — set Status to `completed`, fill in the `## Outcome` section, check off remaining plan steps. The task file is in `{PROJECT_PATH}/.planning/tasks/`.
  5. **Update skillbook** — extract observations from the completed task per the "Update Skillbook" section in `~/.claude/crafter/rules/post-change.md`. Only for non-trivial tasks with genuine learnings. Skip if the `crafter` CLI is not available.
  6. **Suggest session wrap-up** — ...
  ```

  Update the count reference on line 182: "Do not end the conversation until all **6** items above are addressed."

  Note: do NOT add a `skillbook.md` to the rule loading list — there is no `rules/skillbook.md` file in this architecture. The rules are embedded in `delegation.md` and `post-change.md`, and the logic lives in the Go binary.

- [x] **Step 11: Update documentation files.**

  **`.planning/ARCHITECTURE.md`** (`/Users/ret/dev/ai/crafter/.planning/ARCHITECTURE.md`):

  Add `cli/` directory to the structure tree (after line 16, after the hooks section):
  ```
  ├── cli/                        # Go CLI binary source (crafter utility tool)
  │   ├── main.go                 # Entry point
  │   ├── cmd/                    # Cobra command definitions
  │   ├── internal/skillbook/     # Skillbook logic (types, store, jaccard, format)
  │   ├── Makefile                # Cross-compilation targets
  │   ├── go.mod                  # Go module definition
  │   └── go.sum                  # Dependency checksums
  ```

  After the "Dual Installation Model" section (line 80), add two new Key Patterns subsections:

  ```markdown
  ### Crafter CLI — Utility Binary

  A Go CLI binary (`crafter`) provides deterministic utilities that LLMs handle poorly. The binary is a utility tool, NOT orchestration — orchestration stays in markdown prompts. The CLI is invoked via Bash by the orchestrator.

  Current subcommands:
  - `crafter skillbook get` — read skillbook, filter/sort skills, format as markdown, increment appliedCount
  - `crafter skillbook add` — add observation with Jaccard dedup and confidence promotion
  - `crafter skillbook init` — create empty skillbook

  Distribution: cross-compiled for darwin-arm64, darwin-amd64, linux-amd64, linux-arm64. Binaries attached to GitHub releases. `install.sh` downloads the correct binary to `~/.claude/crafter/bin/crafter`.

  ### Skillbook — Project-Level Learning

  The skillbook system lets agents learn from experience across sessions. After each task, the orchestrator reflects on what happened and captures observations via `crafter skillbook add`. Before spawning an agent, the orchestrator calls `crafter skillbook get` and appends the output to the agent's task prompt.

  Key mechanics: Jaccard keyword-overlap deduplication (threshold 0.6), three confidence tiers (low/medium/high) with promotion on repeated observations, top-10 skill selection sorted by confidence then usage count, atomic file writes.

  The skillbook file (`.planning/skillbook.json`) is project-level — agents learn project-specific patterns, not general knowledge.
  ```

  **`.planning/PROJECT.md`** (`/Users/ret/dev/ai/crafter/.planning/PROJECT.md`):

  Update the Stack section (lines 7-10) to add Go:
  ```markdown
  ## Stack

  - **Language:** Markdown (all commands, rules, agents, templates, documentation)
  - **Language:** Go (CLI utility binary — `cli/` directory)
  - **Scripting:** Bash (install.sh only)
  - **Runtime platform:** Claude Code CLI (custom slash commands)
  - **Framework:** cobra (Go CLI command framework)
  ```

  Add to Key Decisions table (line 22):
  ```markdown
  | 2026-03-24 | Go CLI binary for deterministic utilities | LLMs do JSON CRUD, Jaccard similarity, and atomic writes poorly — a static binary with zero runtime deps handles these reliably |
  ```

### Verification Criteria

1. `cd /Users/ret/dev/ai/crafter/cli && go build -o crafter .` succeeds and produces a binary
2. `cd /Users/ret/dev/ai/crafter/cli && go test ./...` passes all tests
3. `./crafter skillbook init --file /tmp/test-skillbook.json` creates a valid JSON file
4. `./crafter skillbook add --agent implementer --rule "Always run tests" --rationale "Tests were missing" --task "test-task.md" --file /tmp/test-skillbook.json` adds a skill
5. Running the same `add` again with similar rule text triggers dedup (merge + confidence promotion)
6. `./crafter skillbook get --agent implementer --file /tmp/test-skillbook.json` outputs a markdown "Learned Guidelines" section
7. `./crafter skillbook get --agent planner --file /tmp/test-skillbook.json` outputs nothing (no skills for that agent)
8. `cd /Users/ret/dev/ai/crafter/cli && make release` produces 4 binaries in `bin/`
9. `rules/delegation.md` contains "Skillbook" section with `crafter skillbook get` instructions
10. `rules/post-change.md` contains "Update Skillbook" section with `crafter skillbook add` instructions
11. `commands/do.md` mandatory checklist has 6 items including "Update skillbook"
12. `install.sh` creates `crafter/bin/` directory and attempts binary download in remote mode
13. `.planning/ARCHITECTURE.md` documents both the CLI and skillbook system
14. `.planning/PROJECT.md` lists Go in the stack
15. `bash tests/test_install.sh` still passes (existing tests not broken)

### Unknowns / Flags

1. **Go version requirement:** The plan specifies Go 1.22+ in `go.mod`. This assumes the developer building the binary has Go installed. End users do NOT need Go — they get a pre-compiled binary. The CI/release workflow for building binaries is manual for now; a GitHub Actions workflow could be added later.

2. **Binary size:** A cobra-based Go binary with the skillbook logic will be roughly 5-8 MB (stripped). This is small enough to distribute alongside tarballs in GitHub releases without concern.

3. **Binary path convention:** The plan places the binary at `~/.claude/crafter/bin/crafter`. This means the orchestrator needs to know this path. It is hardcoded in `delegation.md` and `post-change.md`. If the user uses `--local` install, the path is `.claude/crafter/bin/crafter`. The orchestrator should try both paths. This is documented in the delegation.md instructions.

4. **Graceful degradation:** The entire skillbook system is optional. If the binary is not installed (e.g., old Crafter version, main-branch install without a build step), all skillbook-related prompt instructions say "skip silently". Core Crafter functionality is unaffected.

5. **`install.sh` binary download for main-branch installs:** When users install from main branch (no `--version` flag), there is no GitHub release to download a binary from. The binary will only be available for tagged releases. Main-branch users who want the CLI must build it locally with `cd cli && go build`. The install script handles this gracefully by not failing if the binary download is skipped.

## Decisions

_(none yet)_

## Outcome

_(pending)_
