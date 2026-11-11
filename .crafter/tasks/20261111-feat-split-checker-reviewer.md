# Task: Split crafter-checker into a contract checker and a generic reviewer

## Metadata
- **Date:** 2026-11-11
- **Work branch:** feat/split-checker-reviewer
- **Status:** active
- **Scope:** Medium

## Request
Split `crafter-checker` into two agents:

1. **`crafter-checker`** — contract drift only (current Part A): drift classification, verification evidence, contract deviations, Replan, delta pass against the contract. Likely `effort: medium`.
2. **A new generic reviewer** — adversarial, with no task context. It receives only the change range (a PR link, `base..head`, or the working tree); no contract, no implementer report, no review checklist, no ARCHITECTURE.md (it may discover repo conventions itself). It keeps: a one-sentence role, the Critical/Major/Minor/Suggestion severity scale, the reachability gate for Critical/Major, the completeness and empty-table rules, the output format (summary line, findings table, max 2 sentences + `file:line`), read-only behavior. It drops: the Recommendations buckets, the `--auto` classification table (moves to the orchestrator as a severity→action mapping), delta pass (the orchestrator instead runs a review of the fix diff), and project memory. The reviewer must also be usable standalone outside Crafter on any PR.

Update the orchestrator and rules (mainly `skills/crafter-do/SKILL.md`, `rules/do-workflow.md`, `rules/delegation.md`) so that after implementation both agents run (ideally in parallel) and their outputs merge into the existing decisions: fix loop, Minor stop gate, `--auto` routing.

## Plan
**Plan status:** approved

### Approach
Split today's single `crafter-checker` along the context line. The Checker keeps everything that needs the task context (contract drift, verification evidence, contract deviations, Replan, delta pass). A new context-free `crafter-reviewer` judges the change the way an outsider would: it gets only a change range. The orchestrator runs both in parallel after every implementation unit and every fix. It merges their reports into the existing decisions: fix loop, Minor stop gate, Suggestion auto-record, `--auto` routing. A thin `crafter-review` skill makes the reviewer usable on any PR outside the crafter-do flow. Why: a reviewer that reads the author's contract and report is anchored to the author's view, and today every review is anchored.

### Request (restated)
- **Goal:** two agents with separate inputs. The Checker gets the full task context. The Reviewer gets the change range only. Both run after implementation, their outputs feed the existing gates, and the Reviewer also works standalone on any PR.
- **Acceptance:** (1) `agents/crafter-reviewer.md` exists with the kept/dropped elements from `## Request`. (2) `agents/crafter-checker.md` has no code-review role and runs at `effort: medium`. (3) `/crafter-do` Step 5, the fix loop, Steps 7–9 deferred-fix and `/crafter-debug` verification spawn both agents. Every existing decision rule (cap 5, iteration counting, Minor stop, Suggestion auto-record, reachability send-back, widening, verbatim relay, `--auto`) has a defined owner and input. (4) The installer ships the agent and the skill, and `tests/test_install.sh` passes. (5) No live doc still describes the Checker as a code reviewer.
- **Constraints:** source files only, never `~/.claude/`. Persistent files stay in English. The Reviewer's spawn prompt carries no task context.

### Assumptions / interpretations (decisions proposed — confirm at approval)
1. **Agent name `crafter-reviewer` (reused).** The installer's `crafter-*.md` cleanup glob handles it. The old name survives only in historical records: closed task files, `.crafter/STATE.md` history rows, the "Superseded snapshot" `docs/core-capabilities.md`, and arbitrary `CreatedBy` fixture strings in `cli/internal/buffer/format_test.go`. All of those stay untouched.
2. **Ship a standalone skill `crafter-review`.** It is not called `review` because that collides with Claude Code's built-in `/review` and with the repo's non-installed `skills/review/`. It spawns the reviewer with only the range (a PR link or number, `base..head`, or nothing for the working tree) plus the conversation-language line, and relays the report unchanged.
3. **Reviewer range forms:** a PR, a `base..head` range, or the working tree (uncommitted changes vs HEAD, including untracked files). Each form can be restricted to a path list. In crafter-do the range is always the working tree, because the unit is uncommitted until Step 6b.
4. **Fix-loop re-check:** both agents run in parallel.
   - The **Checker delta pass** gets every prior open finding, drift and code alike. It reports `resolved` / `still present` / `partially resolved` for each, with evidence, plus any new drift in the delta and whether widening is needed. It is the only agent with the context to decide whether a fix did what was asked, so it owns resolved status for code findings too.
   - The **Reviewer** runs on the working tree restricted to the files the fix changed and reports new code findings only.
   - Open set after a pass = prior findings that are not resolved, plus new findings from both reports. The orchestrator does not deduplicate by judgment, so an overlapping Reviewer row stays as its own finding.
   - **Widening** (reported by the Checker) makes the next pass a full pass for both agents. Iteration count and cap are unaffected.
5. **Reachability send-back targets the Reviewer.** It runs as a narrow variant: the same range plus the one finding row, still with no task context. It does not count as an iteration, and each finding gets at most one send-back.
6. **`--auto` routing:**
   - Code findings use an orchestrator-side mapping: Critical/Major go to the fix loop, Minor and Suggestion are auto-recorded.
   - Drift findings keep the Checker's own `--auto` bucket table (`gap` / `uat` / `auto-fixable` / `escape-hatch`), narrowed to drift. The Checker is the only agent with the contract context to judge them. Plan-obsoleting drift still routes to `escape-hatch`.
   - A fix-loop code finding the Implementer leaves unfixed and tags `[uat-worthy]` / `[gap-worthy]` (its existing mechanism) becomes a buffer entry and counts as handled under the buffered-finding carve-out.
7. **Step ticks:** attribution comes from the Checker's drift findings. Reviewer findings carry no step attribution, so they gate only the `Check` line. Resume then re-enters at the pending gate.
8. **The Diff summary stays with the Checker.** The Reviewer output is only a summary line and a findings table.
9. **`--ext`:** matching extension-skill context goes to the Checker only (advisory), never to the Reviewer.
10. **Effort and memory.** Checker: `opus`/`medium`. Reviewer: `opus`/`high`, with no `memory: project` and no Jargon Confinement pointer (no Crafter vocabulary ever reaches it). Both get caveman-lite (both are relayed verbatim). Neither gets ponytail.
11. **Debug flow:** `/crafter-debug` Step 6 spawns both agents in parallel. The Checker verifies the symptom is fixed without a contract, and the Reviewer reviews the working tree.

### Non-goals
- No reviewer delta mode, no review checklist, and no task context of any kind in the Reviewer.
- No Go CLI change. The `Check` gate line, `crafter pr-body` and the buffers stay as they are.
- No edits to historical records: task files, `.crafter/STATE.md` (Steps 7–9 owns it), `docs/core-capabilities.md`, the Go test fixtures.
- No change to the non-installed `skills/review/`.
- No change to fix-loop semantics beyond the new re-check inputs: the cap, the user choices at the cap, and the green-commit invariant stay.

### Relevant areas
`agents/crafter-checker.md`, new `agents/crafter-reviewer.md`, new `skills/crafter-review/SKILL.md`, `skills/crafter-do/SKILL.md` (Master Plan, scope spawn counts, high-risk routing, Step 5, Steps 7–9 item 5, closing Reminder), `rules/do/step-5-check.md`, `rules/do-workflow.md`, `rules/do/step-7-9-post-change.md`, `rules/post-change.md`, `rules/delegation.md`, `rules/core.md`, `rules/do/extension-skills.md`, `skills/crafter-debug/SKILL.md`, `install.sh`, `tests/test_install.sh`, `README.md`, `docs/philosophy.md`, `docs/skill-contract.md`, `docs/plugin-system.md`, `.crafter/ARCHITECTURE.md`, `doc/overview/architecture-overview.md`.

### Steps
- [x] Step 1: **Reviewer agent exists.** `agents/crafter-reviewer.md` covers:
  - frontmatter: `opus`/`high`, tools Bash/Read/Grep/Glob, no memory
  - a one-sentence adversarial role and the range forms from Assumption 3
  - the four-level severity scale, the reachability gate for Critical/Major, and the completeness and empty-table rules
  - the output format: summary line, findings table, at most 2 sentences plus `file:line` per row
  - the reachability-check variant, the shared `## Ending your turn` section, and strict read-only behavior: no edits, commits, pushes, checkout or branch changes, and no GitHub writes or comments
  - tests only when the range's head is the current checkout
- [x] Step 2: **Checker is contract-only.** In `agents/crafter-checker.md`:
  - drop Part B, the severities, the Issues table, the reachability gate and its variant
  - add explicit coverage of verification evidence and of overengineering or drive-by edits as scope drift
  - make the delta pass take every prior finding (drift and code) and report the status of each
  - narrow Recommendations and the `--auto` table to drift
  - set `effort: medium` and update the description
- [x] Step 3: **Standalone skill ships.** Add `skills/crafter-review/SKILL.md` per Assumption 2. Add the agent and the skill to `install.sh` and to the expected-files list in `tests/test_install.sh`.
- [x] Step 4: **crafter-do runs both agents.**
  - `skills/crafter-do/SKILL.md`, `rules/do/step-5-check.md`, `rules/do-workflow.md`, `rules/do/step-7-9-post-change.md` and `rules/post-change.md` encode Assumptions 4–9: parallel spawn, verbatim relay of both reports, reachability send-back to the Reviewer, the `--auto` mapping, the step-tick rule, widening and the deferred-fix re-check.
  - Spawn counts, routing chains and the Master Plan row are updated.
  - Each `rules/do/step-*.md` change is mirrored in its SKILL.md residue item.
- [x] Step 5: **Shared rules agree.**
  - `rules/delegation.md`: roles table, model/effort table and rationale, caveman levels, ponytail exclusions, a Report Language note that the language line is not task context.
  - `rules/core.md`: caveman audience line, carve-out (a) covering both reports, ponytail scope.
  - `rules/do/extension-skills.md`: extension context goes to the Checker.
- [x] Step 6: **Debug flow updated.** `skills/crafter-debug/SKILL.md` Step 6 per Assumption 11. Remove its "Part A" wording if the Checker no longer uses that label.
- [x] Step 7: **Docs sweep.** Update the agent and role descriptions in `README.md` (agent table, command table with `/crafter-review`, check paragraphs), `docs/philosophy.md`, `docs/skill-contract.md` (core-agent list, safety envelope item 5), `docs/plugin-system.md`, `.crafter/ARCHITECTURE.md` (structure, agent count, Model Selection, HITL gates, Check Pass, Skill Adaptation, Agent Memory exception) and `doc/overview/architecture-overview.md`.
- [x] Check

### Contract
- **Outcome:** the Checker judges drift and evidence against the contract. The Reviewer judges the code from the change range alone. Every existing decision rule has a single defined owner and input, and the Reviewer works standalone.
- **Scope boundary:** the files in Relevant areas. The Implementer agent file may get at most a one-line clarification for Assumption 6, and only if the spawn prompt cannot carry it.
- **Non-goals:** as listed above.
- **Seams:**
  - The Reviewer's spawn input: range (+ optional paths), language line, no-filler rule, skill directives — nothing else.
  - The Reviewer's output format and its reachability-variant shape.
  - The Checker's full-pass and delta-pass output sections, including prior-finding status for code findings.
  - The `/crafter-review [<PR>|<base..head>]` invocation.
  - The installer's file list.
- **Verification evidence:**
  - `bash tests/test_install.sh` passes.
  - A grep sweep over live files (excluding historical records) finds no Checker code-review role ("Part B", "drift + code review", "QA + code reviewer", "Checker's classification table" for code findings).
  - Every decision named in Acceptance (3) is traceable to one rule each in SKILL.md and `rules/do/step-5-check.md`, and the two agree.
  - The Reviewer file contains no contract, report or ARCHITECTURE input.
- **Stop conditions:**
  - An existing decision rule cannot be given an owner without task context reaching the Reviewer.
  - The fix-loop re-check needs commits or checkout changes to isolate the fix diff.
  - Any change to the Go CLI or the statusline turns out to be required.

### Alternatives considered
- **Reviewer delta mode for resolved status.** Rejected: the request drops it, and resolution is a context question the Checker already answers.
- **Checker classifies the Reviewer's findings under `--auto`.** Rejected: the Checker would have to run after the Reviewer, which kills parallelism, and it would be judging another agent's code findings.
- **Snapshot the working tree (`git stash create`) to isolate the fix diff.** Rejected: it misses untracked files and adds machinery. A path restriction on the working tree is enough.
- **A new agent name (e.g. `crafter-code-reviewer`).** Viable, but longer, and it avoids nothing that matters, because old references are history only.

### Risks / unknowns / flags
- Confirm Assumptions 1, 2, 6, 7 and 9 at approval. They change user-visible behavior:
  - 6: code findings can no longer be routed straight to `gap`/`uat` by the agent that found them.
  - 7: steps can be ticked while a code finding is still open.
  - 9: `--ext` review plugins now advise the drift checker, not the reviewer.
- Since the Reviewer is context-free, it may re-flag a prior or deferred issue, so a deferred Minor can come back as a new Minor stop. This is accepted noise, not a bug.
- `.crafter/STATE.md` Known Issues mentions the old reachability-variant/`--auto` ambiguity. Steps 7–9 should drop it once this lands.

## Decisions
- **Decision (Tech Debt — auto-recorded):** Suggestion — `/crafter-review` called with only paths (e.g. `src/foo`) is ambiguous: the skill passes it as a range, not as a path-restricted working tree (`skills/crafter-review/SKILL.md:6`).
- **Decision (Tech Debt — auto-recorded):** Suggestion — stale comment "24 files installed" at `tests/test_install.sh:378` (list now has 34 entries).
- **Decision (Tech Debt — auto-recorded):** Suggestion — the fix-loop Implementer spawn field list "(severity or classification, file, line, description)" omits the finding number that the new `--auto` instruction asks the Implementer to return (`rules/do/step-5-check.md:51`, `skills/crafter-do/SKILL.md:246`).

## Outcome
