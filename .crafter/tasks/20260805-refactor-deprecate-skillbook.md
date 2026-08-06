# Task: Deprecate skillbook in favor of native Claude Code per-agent memory

## Metadata
- **Date:** 2026-08-05
- **Work branch:** refactor/deprecate-skillbook
- **Status:** completed
- **Scope:** Large

## Request
Deprecate the skillbook mechanism in favor of native Claude Code per-agent memory (`memory: project` frontmatter). Scope items:

1. Add `memory: project` to all six `agents/*.md` frontmatter.
2. Remove the skillbook injection rule from `rules/delegation.md` and the Update Skillbook step from `rules/post-change.md` (and its mention in `rules/do/step-7-9-post-change.md` commit bundling).
3. Add a short memory-discipline instruction to agent prompts replacing the CLI flow (0-3 observations, project-specific only, curation rules equivalent to the former post-change.md guidance).
4. One-time migration documented in release notes: existing `.crafter/skillbook.json` rules → `.claude/agent-memory/<agent>/MEMORY.md` (no automated migrator unless trivial).
5. Remove Go CLI skillbook code (`cli/cmd/skillbook*.go`, `cli/internal/skillbook/`) + tests + doc/spec reference updates (`doc/spec/features/skillbook-learning-system.md`, `.crafter/ARCHITECTURE.md` skillbook section, `install.sh` if it references skillbook, `doc/00-index.md`).

Context: research confirmed skillbook is defective in practice — confidence promotion never fires (Jaccard 0.6 threshold on free-form rules), appliedCount ranking lock-in makes new entries unreachable once an agent has ≥10 entries, no prune/deprecate/edit path exists, and `install.sh` never opts projects in. Native subagent memory (`memory: user|project|local` frontmatter → `.claude/agent-memory/<agent>/`, MEMORY.md auto-loaded, agent writes its own lessons) covers the use case per official Claude Code docs.

Dependency: stacked on unmerged PR #54 (`refactor/opus5-workflow-alignment`), which touches `delegation.md`, `post-change.md`, `agents/*.md`. User approved stacking; PR base = `refactor/opus5-workflow-alignment`, auto-retargets to `main` after #54 merges.

## Plan

**Plan status:** approved

### 1. Complete request

Retire the skillbook mechanism end to end and let native Claude Code per-agent memory carry project-level learning instead.

Today the orchestrator does two things around every task: before each spawn it shells out to `crafter skillbook get` and pastes a "Learned Guidelines" block into the agent prompt (`rules/delegation.md` § Skillbook), and after each task it reflects and calls `crafter skillbook add` (`rules/post-change.md` § Update Skillbook). Research already established the mechanism is defective in practice — Jaccard 0.6 dedup on free-form sentences essentially never merges, so confidence never gets promoted; `appliedCount` ranking locks the top 10 so new entries are unreachable; there is no prune, edit, or deprecate path; and `install.sh` never opts a project in, so almost no installation has a skillbook at all. Native subagent memory (`memory: project` in agent frontmatter → `.claude/agent-memory/<agent>/MEMORY.md`, auto-loaded on spawn, agent curates its own file) covers the same use case with none of the custom machinery.

**Acceptance criteria**

- All six `agents/crafter-*.md` carry `memory: project` frontmatter plus a short memory-discipline instruction (0–3 observations per task, project-specific patterns only, no generic programming advice, curate/replace rather than append forever) — the curation intent that is being deleted from `post-change.md` survives, relocated to the agent that actually does the work.
- No skillbook CLI invocation, section, or instruction remains in `rules/`, `skills/`, or `agents/`.
- `cli/cmd/skillbook*.go` and `cli/internal/skillbook/` (including the three test files) are gone; `mise exec -- go build ./...` and `mise exec -- go test ./...` pass; `crafter --help` no longer lists `skillbook`.
- Docs no longer describe or link a skillbook: `doc/spec/features/skillbook-learning-system.md` removed, `doc/00-index.md` entry removed, `.crafter/ARCHITECTURE.md` skillbook content replaced by a short agent-memory description.
- A manual migration paragraph exists in prose the release can reuse (existing `.crafter/skillbook.json` entries → `.claude/agent-memory/<agent>/MEMORY.md`, then delete the JSON).
- `bash tests/test_install.sh` still passes unchanged.

**Constraints**

- Minimal diffs. Content edits only in `rules/` — no file added, renamed, or deleted under `rules/do/`.
- No VERSION bump (separate release flow).
- One commit per phase, per the standard workflow.
- `.crafter/ARCHITECTURE.md` edits must be delegated to the Implementer (the orchestrator does not edit it directly).
- The branch is `refactor/deprecate-skillbook`, stacked on unmerged PR #54; the PR base is `refactor/opus5-workflow-alignment`.

**Validation strategy** — grep-based absence checks over source trees (excluding untracked build artifacts and historical records), the Go build+test suite, the installer test suite, and one real end-to-end observation that a spawned agent actually gets and writes memory.

**Why it matters** — the skillbook is a maintained-but-broken subsystem: ~8 Go source files, 3 test files, two prompt integrations, one spec doc, and an architecture section, all serving a feature that demonstrably does not learn. Replacing it with a platform feature deletes far more than it adds.

### 2. Assumptions / interpretations

- **Native memory is available and works with restricted `tools:` lists.** Four of the six agents (`analyzer`, `reviewer`, `verifier`, `step-runner`) have no `Write`/`Edit` tool. The plan assumes Claude Code provides the memory capability from the `memory:` frontmatter field independently of the `tools:` allowlist. Phase 1 Step 1 verifies this before anything else is touched; if it is false the phase stops (see Risks).
- **"Never modifies files" contracts need an explicit carve-out.** Analyzer, reviewer, verifier, and step-runner all state they never modify files. Curating MEMORY.md is a file write. The memory-discipline note is interpreted as also being the carve-out — it must say that the agent's own memory file is the single exception. Without this, those agents will correctly refuse to write memory.
- **The discipline note lives in each agent file, not in `rules/core.md`.** The project's usual pattern is one source of truth plus pointers, but agents only *pointer-reference* `rules/core.md` (for Jargon Confinement) and do not reliably load it. Memory curation is per-agent behavior at the moment the agent finishes, so the text belongs in the agent file. Keep the wording identical across the six files so there is one editable shape.
- **The migration paragraph goes in the task file's `## Outcome` section.** Step 9b composes the PR body from `## Plan → Approach` and `## Outcome`, and release notes are drafted from there — so Outcome is the natural host and no new doc is needed.
- **`.crafter/skillbook.json` is tracked and should be deleted in this task.** It is dead data once the read path is gone. Whether its current entries are first migrated into this repo's own `.claude/agent-memory/` is flagged as an open decision, not assumed.
- **`cli/bin/` binaries are untracked local artifacts** — they still contain the string `skillbook` after this change and are rebuilt by the release flow. Absence greps must exclude them.

### 3. Non-goals

- No VERSION bump, no release, no rebuild/commit of `cli/bin/` binaries.
- No automated migrator command, and no deprecation shim that keeps `crafter skillbook` alive with a warning.
- No renames or file additions/deletions under `rules/do/`.
- No edits to `cli/internal/statusline/parse.go` or `statusline_test.go` — their "skillbook update" strings are task-file plan-line fixtures, not skillbook usage.
- No rewriting of historical records: `.crafter/tasks/*` from earlier tasks and existing `.crafter/STATE.md` "Recent Changes" rows keep their skillbook mentions.
- No changes to `install.sh` file lists or `tests/test_install.sh` (neither references skillbook; verify, do not edit).
- No memory for the orchestrator or for skills — memory is a subagent feature and this task only touches the six agents.
- No broadening of agent `tools:` allowlists beyond whatever native memory strictly requires.
- No redesign of what gets remembered beyond the curation rules being relocated.

### 4. Relevant areas

Prompt layer:
- `agents/crafter-planner.md`, `crafter-implementer.md`, `crafter-verifier.md`, `crafter-reviewer.md`, `crafter-analyzer.md`, `crafter-step-runner.md` — frontmatter blocks (lines 1–8ish) and the "Critical Rules" / constraints sections where the never-modify-files statements live.
- `rules/delegation.md` § "Skillbook — Learned Guidelines" (lines 43–56) — sits between the Model Configuration section and § "Skill Directives — Caveman and Ponytail".
- `rules/post-change.md` § "Update Skillbook" (lines 43–70) and § "Consolidated End-of-Task Commit" (lines 72–82, one bullet + one sentence mention skillbook).
- `rules/do/step-7-9-post-change.md` — intro line and checklist item 2.
- `rules/do/step-9b-pr-composition.md` — trigger paragraph lists skillbook among possible follow-up-commit contents.
- `skills/crafter-do/SKILL.md` line ~372, Steps 7–9 checklist item 2 — same consolidated-commit sentence. This is the file the orchestrator actually executes; leaving it stale would keep the behavior live.

Go CLI:
- `cli/cmd/skillbook.go`, `skillbook_add.go`, `skillbook_get.go`, `skillbook_init.go` — the only importers of the package; each self-registers via `init()`, so `cmd/root.go` needs no edit.
- `cli/internal/skillbook/` — `types.go`, `store.go`, `jaccard.go`, `format.go` + `store_test.go`, `jaccard_test.go`, `format_test.go`. Nothing outside `cli/cmd/skillbook*.go` imports this package (verified).
- `cli/internal/buffer/format.go` (~line 23) and `cli/internal/claudesettings/store.go` (~line 108) — comments that reference skillbook as a pattern precedent; they become dangling references.

Docs:
- `doc/spec/features/skillbook-learning-system.md` — the only file in `doc/spec/features/`.
- `doc/00-index.md` line 19 under "## Feature Specifications" — that heading has exactly one entry.
- `doc/documentation-handbook.md` line 42 — uses the spec filename as a naming-convention example.
- `.crafter/ARCHITECTURE.md` — tree entries (lines ~27–30, ~40), CLI subcommand bullets (~130–132), and § "Skillbook — Project-Level Learning" (~152–158). Delegate all edits here.

Data / tests:
- `.crafter/skillbook.json` (tracked).
- `tests/test_install.sh` — enumerates deployed agent/rule filenames (lines ~411–416); confirm no skillbook assertions exist.

### 5. Phases and steps

---

#### Phase 1 — Agents learn via native memory; the skillbook prompt path is gone

**Phase Karpathy Contract**

- **Outcome:** after this phase, a `/crafter-do` run never calls the skillbook CLI, and every spawned agent has its own project-scoped memory file with explicit curation rules. The prompt layer is internally consistent — no rule, skill, or agent file mentions a skillbook.
- **Scope boundary:** `agents/*.md`, `rules/delegation.md`, `rules/post-change.md`, `rules/do/step-7-9-post-change.md`, `rules/do/step-9b-pr-composition.md`, `skills/crafter-do/SKILL.md`. Nothing in `cli/`, `doc/`, or `install.sh`.
- **Non-goals:** the Go binary still builds and still ships a `skillbook` subcommand at the end of this phase — that is expected and removed in Phase 2. No restructuring of the surrounding sections beyond removing skillbook content and closing the resulting seams.
- **Simplicity constraint:** deletion first. The only addition is the frontmatter line plus one short discipline block per agent, worded identically across the six files. No new rules module, no pointer indirection, no configurability.
- **Drift criteria:** any new file; any edit under `rules/do/` beyond removing skillbook mentions; any change to the caveman/ponytail propagation rule in `delegation.md`; the discipline note growing into a general "how to be a good agent" essay; different wording per agent beyond the role name.
- **Verification evidence:** `grep -rn -i skillbook rules/ skills/ agents/` returns nothing; the six agent files show `memory: project` and the discipline block; one real spawn creates `.claude/agent-memory/<agent>/` and the orchestrator's spawn sequence contains no `crafter skillbook` Bash call; `bash tests/test_install.sh` passes.
- **Stop conditions:** native `memory:` frontmatter is not supported by the installed Claude Code, or the memory capability is blocked by the `tools:` allowlist for the four read-only agents and the workaround would require widening those allowlists to `Write`/`Edit`. Either case → stop, report, ask the user before continuing.

**Steps**

- [x] **Step 1.1 — Prove native memory works for a read-only agent, then wire the frontmatter.** Confirm the installed Claude Code honors `memory: project` in subagent frontmatter and that an agent whose `tools:` list has no `Write`/`Edit` can still persist memory. Then add `memory: project` to all six `agents/crafter-*.md` frontmatter blocks.
  - *Outcome:* six agents declared as project-memory agents, on a mechanism confirmed to actually work here.
  - *Scope boundary:* frontmatter blocks only; plus whatever read-only checking (docs, a probe spawn, inspecting `.claude/agent-memory/`) the verification needs.
  - *Non-goals:* no prompt-body edits yet; no `tools:` changes; no memory for skills or the orchestrator.
  - *Simplicity constraint:* one line per file.
  - *Drift criteria:* adding `Write`/`Edit` to an agent that did not have it; adding memory-related fields other than `memory:`; touching agent bodies.
  - *Verification evidence:* the six frontmatter blocks; concrete evidence (a created `.claude/agent-memory/<agent>/` directory or equivalent) that the mechanism fires for at least one no-Write agent.
  - *Stop conditions:* the mechanism does not fire, or firing requires widening a tool allowlist → stop and report; do not proceed to Step 1.2.

- [x] **Step 1.2 — Give each agent its curation rules and the memory-file carve-out.** Add a short, identical memory-discipline block to each of the six agent files: 0–3 observations per task and only when genuinely useful; project-specific patterns only (the good/bad examples currently in `post-change.md` are the reference for the bar); curate the file — replace and prune stale entries rather than appending forever; and, for the agents whose contract says they never modify files, an explicit statement that their own memory file is the sole exception.
  - *Outcome:* the curation guidance deleted from `post-change.md` in Step 1.4 survives at the point of use, and read-only agents are unblocked from writing memory.
  - *Scope boundary:* the six `agents/*.md` bodies; place the block consistently (e.g. near each file's existing constraints/critical-rules area).
  - *Non-goals:* no new `rules/` module; no pointer-only indirection to `core.md`; no per-agent bespoke advice about what to remember.
  - *Simplicity constraint:* a handful of lines, same text everywhere except the role word. If it reads like a policy document, it is too long.
  - *Drift criteria:* per-agent divergence in wording; the block turning into generic prompt-engineering advice; contradicting an agent's existing never-modify-files rule instead of carving out from it.
  - *Verification evidence:* the six blocks side by side are textually consistent; each read-only agent's constraint section no longer conflicts with memory writing.
  - *Stop conditions:* an agent's existing constraints cannot be reconciled without rewriting them substantively → report before rewriting.

- [x] **Step 1.3 — Delete skillbook injection from the delegation rule.** Remove § "Skillbook — Learned Guidelines" from `rules/delegation.md` and close the seam so the Model Configuration section flows into § "Skill Directives — Caveman and Ponytail".
  - *Outcome:* pre-spawn no longer shells out to the CLI; nothing is appended to agent prompts from a skillbook.
  - *Scope boundary:* that one section and its surrounding whitespace/heading order.
  - *Non-goals:* the caveman/ponytail propagation rule is untouched; no renumbering or rewording of the agent-roles or model tables.
  - *Simplicity constraint:* pure deletion; no replacement paragraph explaining that memory exists (agent frontmatter already carries it).
  - *Drift criteria:* any content edit inside § Skill Directives; adding a new "Agent Memory" section to `delegation.md`.
  - *Verification evidence:* `grep -n -i skillbook rules/delegation.md` is empty; the caveman/ponytail section is byte-identical to before.
  - *Stop conditions:* none expected.

- [x] **Step 1.4 — Delete the Update Skillbook step and de-skillbook the consolidated-commit wording.** Remove § "Update Skillbook" from `rules/post-change.md` and drop skillbook from the consolidated end-of-task commit list there; apply the same removal to the matching sentences in `rules/do/step-7-9-post-change.md` (intro + checklist item 2), `rules/do/step-9b-pr-composition.md` (trigger paragraph), and `skills/crafter-do/SKILL.md` (Steps 7–9 checklist item 2).
  - *Outcome:* the post-task flow is docs + STATE.md only; the four places that enumerate what the consolidated commit may contain agree with each other and with `post-change.md`.
  - *Scope boundary:* the named sections in those four files.
  - *Non-goals:* no change to the consolidated-commit mechanism itself, the 5-item checklist structure, its numbering, or the skill-directive markers in `SKILL.md`.
  - *Simplicity constraint:* remove the skillbook item from each list; do not restructure the lists.
  - *Drift criteria:* checklist items renumbered or merged; `SKILL.md` spawn markers altered; a new "update memory" step added to the orchestrator flow (agents curate their own memory — the orchestrator has no role).
  - *Verification evidence:* `grep -rn -i skillbook rules/ skills/ agents/` returns nothing; the four consolidated-commit descriptions still list the same remaining items in the same order.
  - *Stop conditions:* none expected.

- [x] **Phase verification** — `grep -rn -i skillbook rules/ skills/ agents/` empty; six agent files carry `memory: project` + the discipline block; evidence that memory fires for a no-Write agent; `bash tests/test_install.sh` passes; `git diff --stat` shows only the six agent files and the five prompt files.
- [x] **Phase review** — clean after 2 fix-loop iterations; committed as 3ad6a6c

---

#### Phase 2 — The skillbook subsystem no longer exists in code or docs

**Phase Karpathy Contract**

- **Outcome:** the Go binary, the documentation set, the architecture description, and the repo's own data file all reflect a world without a skillbook, and the manual migration path is written down where the release can reuse it.
- **Scope boundary:** `cli/cmd/skillbook*.go`, `cli/internal/skillbook/`, two comment references in `cli/internal/`, `doc/spec/features/`, `doc/00-index.md`, `doc/documentation-handbook.md`, `.crafter/ARCHITECTURE.md`, `.crafter/skillbook.json`, and this task file's `## Outcome`.
- **Non-goals:** no VERSION bump; no `cli/bin/` rebuild or commit; no `install.sh` or `tests/test_install.sh` edits; no statusline fixture edits; no rewriting of historical task files or STATE.md rows.
- **Simplicity constraint:** deletions plus one short architecture paragraph and one migration paragraph. No new doc file, no migrator, no shim.
- **Drift criteria:** `cmd/root.go` edited (the commands self-register, so it should not need to be); any non-skillbook Go file changed beyond the two dangling comments; a new doc created to host the migration text; the ARCHITECTURE.md replacement section growing into a feature spec.
- **Verification evidence:** `mise exec -- go build ./...` and `mise exec -- go test ./...` pass from `cli/`; the built binary's `--help` has no `skillbook`; `grep -rn -i skillbook` over tracked sources (excluding `cli/bin/`, `.claude/crafter/`, `.crafter/tasks/` history, and STATE.md history rows) returns nothing; `bash tests/test_install.sh` passes.
- **Stop conditions:** the Go build or tests fail for a reason unrelated to the deletion; deleting the last feature spec would leave a doc structure the Implementer cannot resolve without inventing a new convention → report and ask.

**Steps**

- [x] **Step 2.1 — Remove the skillbook Go code.** Delete `cli/cmd/skillbook.go`, `skillbook_add.go`, `skillbook_get.go`, `skillbook_init.go` and the whole `cli/internal/skillbook/` package including its three test files.
  - *Outcome:* `crafter` builds and tests green with no `skillbook` subcommand.
  - *Scope boundary:* those files only.
  - *Non-goals:* no other command touched; `cmd/root.go` should not need an edit — if it does, that is a discovery to report.
  - *Simplicity constraint:* deletion only.
  - *Drift criteria:* any surviving import of the package; edits to unrelated Go files; test files preserved "just in case".
  - *Verification evidence:* `mise exec -- go build ./...`, `mise exec -- go test ./...`, and `--help` output without `skillbook`.
  - *Stop conditions:* a non-obvious dependency on the package surfaces → report rather than refactor around it.

- [x] **Step 2.2 — Clear the two dangling code comments.** Update the comments in `cli/internal/buffer/format.go` and `cli/internal/claudesettings/store.go` that cite the skillbook package as a precedent, so they no longer point at deleted code.
  - *Outcome:* no comment references a package that no longer exists.
  - *Scope boundary:* those two comments.
  - *Non-goals:* no behavior change, no refactor of the surrounding functions.
  - *Simplicity constraint:* restate the fact inline or drop the cross-reference — one line each.
  - *Drift criteria:* code changes in either file; comment rewrites beyond removing the stale reference.
  - *Verification evidence:* `mise exec -- go test ./...` still passes; the diff shows comment-only lines.
  - *Stop conditions:* none expected.

- [x] **Step 2.3 — Remove the skillbook from the documentation set.** Delete `doc/spec/features/skillbook-learning-system.md`, remove its entry from `doc/00-index.md`, and stop using its filename as the naming example in `doc/documentation-handbook.md`.
  - *Outcome:* the docs tree has no skillbook spec and no link to one.
  - *Scope boundary:* those three files.
  - *Non-goals:* no new feature spec written to replace it; no reorganization of `doc/`.
  - *Simplicity constraint:* the index's "Feature Specifications" heading has exactly one entry — if removing it empties the section, drop the empty heading too rather than leaving a stub. The handbook example is illustrative and does not need to name a real file.
  - *Drift criteria:* creating a replacement spec; restructuring `doc/00-index.md` beyond the affected lines; changing the handbook's naming rules themselves.
  - *Verification evidence:* `grep -rn -i skillbook doc/` empty; `doc/00-index.md` contains no dead link.
  - *Stop conditions:* none expected.

- [x] **Step 2.4 — Bring ARCHITECTURE.md in line (delegated).** Via the Implementer, remove the four skillbook lines from the source tree diagram, the `internal/skillbook/` line, the three `crafter skillbook …` subcommand bullets, and § "Skillbook — Project-Level Learning"; replace that section with a short description of native per-agent memory (`memory: project` frontmatter → `.claude/agent-memory/<agent>/MEMORY.md`, agent-curated, project-scoped), sized like the neighbouring § "Skill Adaptation — Caveman and Ponytail".
  - *Outcome:* the architecture document describes the mechanism that now exists.
  - *Scope boundary:* `.crafter/ARCHITECTURE.md` only.
  - *Non-goals:* no other architecture sections touched; no restatement of the agent-side curation rules (they live in the agent files).
  - *Simplicity constraint:* one short paragraph, no bullet taxonomy of mechanics.
  - *Drift criteria:* the new section exceeding the length of the section it replaces; documenting behavior not implemented in Phase 1; edits to unrelated sections.
  - *Verification evidence:* `grep -n -i skillbook .crafter/ARCHITECTURE.md` empty; the new section names the frontmatter field and the memory path and nothing else.
  - *Stop conditions:* the ARCHITECTURE.md check surfaces other stale content — report it, do not fix it in this task.

- [x] **Step 2.5 — Retire the data file and write the migration paragraph.** Remove the tracked `.crafter/skillbook.json` and record, in this task file's `## Outcome` section, a short migration paragraph the release notes can reuse: skillbook is removed; anyone with a `.crafter/skillbook.json` (or legacy `.planning/skillbook.json`) should copy any entries still worth keeping into `.claude/agent-memory/<agent>/MEMORY.md` in that project, then delete the JSON; nothing is migrated automatically and nothing breaks if it is ignored.
  - *Outcome:* no dead data in the repo, and the one-time manual step is written down in prose the release flow already consumes.
  - *Scope boundary:* `.crafter/skillbook.json` and the `## Outcome` section of this task file.
  - *Non-goals:* no automated migrator; no new migration guide under `doc/`; no changes to other task files.
  - *Simplicity constraint:* one paragraph. If the entries in this repo's own skillbook are worth keeping, hand-copying them is a judgement call — see the open decision under Risks; do not build tooling for it.
  - *Drift criteria:* writing a migration script; creating a new doc; editing sections of the task file other than `## Outcome`.
  - *Verification evidence:* the file is gone from `git ls-files`; the `## Outcome` paragraph names both the old and new locations and states that migration is manual and optional.
  - *Stop conditions:* the open decision about migrating this repo's own entries has not been answered → apply the default (delete without migrating) and note it, or ask if the user is available.

- [x] **Phase verification** — `mise exec -- go build ./...` and `mise exec -- go test ./...` pass; `--help` has no `skillbook`; `grep -rn -i skillbook` over tracked sources excluding `cli/bin/`, `.claude/crafter/`, `.crafter/tasks/` history and STATE.md history rows returns nothing; `bash tests/test_install.sh` passes; `git status` shows the deletions staged. *(All criteria carry fresh verifier evidence from the two Phase 2 step drift checks, which re-ran every listed command; a separate identical verifier pass was consolidated away — orchestrator decision.)*
- [x] **Phase review** — clean after 1 fix-loop iteration (9 optional findings dispositioned, 0 Critical/Major)

---

### 6. Alternatives considered

- **Automated migrator (`crafter skillbook migrate` or a script).** Rejected. One-time, per-project, a handful of free-form sentences whose value is exactly what a human should judge. Building a migrator means keeping skillbook parsing code alive to delete the skillbook.
- **Keep the CLI, drop only the prompt integration.** Rejected. That leaves ~11 Go files and a spec doc maintained for a command nothing invokes — the opposite of the point.
- **Deprecation shim: `crafter skillbook` prints a warning and exits.** Rejected. Rules and binary ship together in one install, so there is no version-skew window where old prompts meet a new binary. Nobody is opted in by the installer.
- **Put the memory-discipline text in `rules/core.md` and pointer to it from agents.** Rejected for this case. Agents only pointer-reference `core.md` and do not reliably load it; memory curation happens inside the agent at the end of its own run, so the instruction has to be present in the prompt that is actually loaded. The identical-wording constraint keeps the six copies from drifting.
- **Give memory only to the agents that plausibly learn (planner, implementer, reviewer).** Rejected because the request names all six; noted under Risks that step-runner memory is likely low value and can be pruned later if it turns noisy.
- **Write the migration text into a new `doc/guides/` page.** Rejected. A one-time transition note that stops being true after one release does not deserve a permanent doc; the task `## Outcome` already feeds the PR body and release notes.

### 7. Risks / unknowns / flags

1. **Blocking — native memory support is unverified here.** Whether the installed Claude Code honors `memory: project` in subagent frontmatter, and whether the memory capability is available to agents whose `tools:` allowlist excludes `Write`/`Edit`, has not been confirmed in this environment. Phase 1 Step 1 exists to settle this before anything else changes. If it fails, the whole approach is obsolete and the phase stops for a user decision.
2. **Contract conflict with read-only agents.** `crafter-analyzer`, `crafter-reviewer`, `crafter-verifier`, and `crafter-step-runner` all declare that they never modify files. Unless the carve-out in Step 1.2 is explicit, they will follow their contract and never write memory — the change would silently no-op.
3. **Open decision — `.claude/agent-memory/` and version control.** It is not in `.gitignore`, and `.claude/` is essentially untracked in this repo (one skill file aside). Committing agent memory would mirror the old committed skillbook (shared team learning); ignoring it makes memory per-developer. This needs a user decision; it may warrant a `.gitignore` line, which is currently outside the plan's scope.
4. **Open decision — migrate this repo's own skillbook entries?** `.crafter/skillbook.json` holds real entries (including three planner guidelines that were injected into this very planning run). Default assumed in Step 2.5 is delete-without-migrating. Hand-copying the good ones into `.claude/agent-memory/crafter-*/MEMORY.md` would be dogfooding and depends on decision 3.
5. **Behavior change: the orchestrator loses its view of learned guidelines.** Previously the orchestrator ran `skillbook get` and could see what was injected; with native memory the content is loaded inside the agent and is invisible to the orchestrator and to the user unless someone opens the MEMORY.md. Accepted, but worth naming — there is no longer a "briefly tell the user what was learned" moment.
6. **Stacked-branch churn.** PR #54 touches `delegation.md`, `post-change.md`, and `agents/*.md` — the same files as Phase 1. If #54 gets further review changes, this branch will need a rebase and the Phase 1 diff may conflict.
7. **Deleting the last feature spec empties `doc/spec/features/`.** Step 2.3 handles it by also dropping the now-empty index heading, but if the project wants to keep the section for future specs, say so.
8. **Memory quality is unbounded and unreviewed.** Agents will write their own files with no dedup, no cap, and no review gate — the failure mode is the opposite of the skillbook's (noise accumulation rather than lock-in). The 0–3-per-task and curate-don't-append rules in Step 1.2 are the only defense; a periodic manual read of the MEMORY.md files is the fallback.

## Decisions
- **Decision (User Accepted):** `.claude/agent-memory/` is per-developer — add it to `.gitignore` rather than committing memory files. **Reason:** avoids MEMORY.md merge conflicts and PR noise; shared conventions belong in CLAUDE.md/rules, not in agent memory.
- **Decision (User Accepted):** Migrate curated, still-valid project-specific rules from `.crafter/skillbook.json` into the matching agents' `.claude/agent-memory/<agent>/MEMORY.md` (untracked), then delete the JSON. **Reason:** preserves real lessons; discards stale/generic entries.
- **Decision (User Accepted):** Removing the last feature spec also removes the now-empty features section heading from `doc/00-index.md`. **Reason:** empty section is noise; a future spec re-adds it.
- **Decision (User Accepted):** Phase 1 review finding #6 (`.ai/local/bootstrapper-context.yaml` references the spec file deleted in Phase 2) is deferred into Phase 2 Step 2.3 scope. Findings #1-#5 and #7-#12 are fixed in the Phase 1 fix loop. **Reason:** #6 only dangles once Phase 2 deletes the spec; fixing it belongs with that deletion.
- **Decision (User Accepted):** Iteration-1 review Major #1 (no in-environment proof that `memory: project` creates `.claude/agent-memory/<agent>/`) is accepted as a deferred UAT item, not a code fix: live verification requires merging this change and reinstalling crafter, then spawning a no-Write agent (verifier or step-runner) and checking the directory + MEMORY.md load. Official docs (code.claude.com/docs/en/sub-agents.md § Enable persistent memory) confirm the mechanism, including auto-enabled Read/Write/Edit for memory files independent of `tools:`. **Reason:** environment constraint — agent definitions register at session start and the installed copy predates this change; the risk is documented, not removable in-session.
- **Decision (Tech Debt — recorded):** Suggestion — the downstream "add `.claude/agent-memory/` to `.gitignore`" MUST in `rules/delegation.md` has no enforcing actor; consider install-time scaffolding of both `.crafter/run/` and `.claude/agent-memory/` gitignore entries in a future task.
- **Decision (Tech Debt — recorded):** `doc/documentation-handbook.md` lines 20-21 still describe a `doc/spec/` structure (including a never-existing `nonfunctional.md`) that no longer has any backing files after the spec deletion; handbook restructure is out of this task's scope.
- **Decision (User Accepted):** Phase 2 review findings #1, #3, #4, #8 fixed in the fix loop; #7 handled in the Steps 7-9 STATE.md update; #2 covered by the recorded install-scaffolding tech debt; #5/#6 recorded as tech debt above; #9 left as-is (dropped entries recoverable from git history of the deleted JSON). **Reason:** user approved this disposition.
- **Decision (Orchestrator Accepted):** Phase 2 phase-verification gate satisfied by the two step drift checks instead of a fourth verifier spawn — both drift checks re-ran every phase criterion (Go build+test, `--help`, exclusion-grep, `tests/test_install.sh`, staged deletions) with fresh evidence; a separate pass would have executed identical commands. **Reason:** avoids a redundant verification run; all criteria have cited evidence in this task's step records.

## Outcome

Completed in two phases on branch `refactor/deprecate-skillbook` (stacked on PR #54):

- **Phase 1 — `3ad6a6c`:** six agents moved to native per-agent memory (`memory: project` + `## Memory` discipline block with per-agent carve-outs); skillbook injection and Update Skillbook flow removed from `rules/`; `.claude/agent-memory/` gitignored. Review: 2 fix-loop iterations, 17 findings fixed/dispositioned, 0 unresolved.
- **Phase 2 — `3a66b47`:** skillbook Go CLI (11 files), feature spec, index entry, handbook example, bootstrapper reference, and tracked `.crafter/skillbook.json` deleted; ARCHITECTURE.md section replaced; curated 26→21 entry migration into `.claude/agent-memory/crafter-{implementer,reviewer,planner,verifier}/MEMORY.md`. Review: clean after 1 fix-loop iteration.
- Go build/tests and `tests/test_install.sh` (64/64) green throughout; `crafter --help` no longer lists `skillbook`.
- **Deferred UAT:** after merge + reinstall, spawn a no-`Write` agent (verifier or step-runner) in a project and confirm `.claude/agent-memory/<agent>/` is created and MEMORY.md loads.

### Migration note

The skillbook is removed — the `crafter skillbook` subcommands, the prompt injection, and the post-task update step are all gone. Agents now learn through native Claude Code per-agent memory (`memory: project` frontmatter → `.claude/agent-memory/<agent>/MEMORY.md`, agent-curated and project-scoped). If a project has a `.crafter/skillbook.json` (or a legacy `.planning/skillbook.json`), copy any entries still worth keeping into `.claude/agent-memory/<agent>/MEMORY.md` in that project — mapping the entry's `agent` field to the matching agent name, e.g. `implementer` → `crafter-implementer` — then delete the JSON file. Nothing is migrated automatically, and nothing breaks if the file is simply left in place or ignored: it is dead data that is no longer read. Downstream projects should also add `.claude/agent-memory/` to their `.gitignore` — agent memory is per-developer and is not meant to be committed.
