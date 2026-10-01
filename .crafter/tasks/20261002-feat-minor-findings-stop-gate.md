# Task: Restore a Minor-findings stop gate in crafter-do

## Metadata
- **Date:** 2026-10-02
- **Work branch:** feat/minor-findings-stop-gate
- **Status:** completed
- **Scope:** Medium

## Request
Restore a Minor-findings stop gate in the crafter-do flow (default mode). Current behavior since PR #58: after the Checker pass, Minor and Suggestion findings are auto-recorded as `Decision (Tech Debt — auto-recorded)` and the commit proceeds; the flow STOPs only for Critical/Major. New behavior:

1. Critical/Major — unchanged STOP + fix loop.
2. Minor — the flow STOPs and asks the user to choose: fix all, pick numbers to fix, or defer all to tech debt; chosen fixes go through the same fix loop (delta pass) as Major; deferred ones are recorded as tech-debt Decisions with a non-auto wording (user-deferred).
3. Suggestion — unchanged: auto-recorded as tech debt, no stop.
4. Under `--auto` — unchanged: Minor and Suggestion auto-recorded as tech debt, no stop.
5. `--fast` stays removed; update its removal error message in `skills/crafter-do/SKILL.md` and `rules/do/flag-validation.md`, which currently claims "Minor findings now auto-proceed".

Rationale: Opus 5.5 code review catches more real bugs with fewer false alarms and the Checker has a reachability gate, so Minor findings are likely real and cheap to fix while context is fresh; Suggestions are low-confidence by definition. Keep the manual-verification exception untouched. Update docs that describe the commit model (README.md, docs/philosophy.md, .crafter/ARCHITECTURE.md, .crafter/STATE.md) to match. Prompt/markdown-only change; work on a new branch off main, open a PR at the end, do not merge.

## Plan
**Plan status:** approved

### Complete request
Re-introduce a user stop for Minor findings in the default (non-`--auto`) `crafter-do` flow. After any Checker report (full or delta pass) that contains Minor findings, the orchestrator relays the report verbatim, then STOPs and asks: **fix all / pick numbers / defer all**. Chosen Minors enter the existing fix loop (delta passes, same 5-iteration cap) alongside any Critical/Major; deferred Minors are recorded as `Decision (Tech Debt — user-deferred): Minor — <description>`. Suggestions stay auto-recorded (`Decision (Tech Debt — auto-recorded)`), no stop. `--auto` behavior unchanged (Minor + Suggestion auto-recorded). `--fast` stays removed; its error text loses the "Minor findings now auto-proceed" claim. Manual-verification exception untouched. Docs describing the commit model updated to match. Why: Opus 5.5 review plus the reachability gate makes Minors likely real and cheap to fix while context is fresh; Suggestions are low-confidence by definition.

Acceptance: no source file under `skills/`, `rules/`, `agents/`, `README.md`, `docs/philosophy.md`, `.crafter/ARCHITECTURE.md` still says Minor findings auto-proceed / never gate in the default flow; `--auto` text still says Minor auto-records; SKILL.md residue and `rules/do/*` modules agree.

### Assumptions / interpretations
- New label is `Decision (Tech Debt — user-deferred)`. Verified: the CLI (`cli/internal/prbody`, statusline) never matches the `Tech Debt — auto-recorded` string — it copies the Decisions section wholesale — so no Go change is needed.
- The Minor gate applies to every pass report, including delta passes that surface new Minors (consistent with "Minor stops"). Reachability send-back stays Critical/Major only.
- Minors the user chose to fix count as open for the Check-gate tick and the transition to Step 6b, and keep the fix loop going until resolved; at the cap they fall under the existing (a)/(b)/(c) choice. Minor findings still never block the per-step batch tick.
- If the same report has Critical/Major findings too, one stop covers both: Critical/Major are mandatory, the Minor choice is asked in the same prompt.
- Step 6b auto-commit condition now accepts remaining findings that are Suggestions (auto-recorded) or Minors (user-deferred). Steps 7–9 deferred-findings list includes both labels.

### Non-goals
- No change to `--auto`, the Checker's `--auto` classification table, severity definitions, reachability gate, fix-loop cap, or the manual-verification exception.
- No CLI/Go or test changes (none assert these strings). No new rule module or mechanism — reuse the fix loop.
- Historical `.crafter/STATE.md` rows and the Ideas entry about `--fast` stay as written.

### Relevant areas
- `skills/crafter-do/SKILL.md` — Flag Validation error, Master Plan table (Step 5, Step 6b rows), Flag branching, Step 5 residue (b), (c), (d), (f) entry/close conditions, (g), (h.1), Step 6b `--auto` branch (keep) and path (1), Steps 7–9 item 5.
- `rules/do/flag-validation.md`, `rules/do/step-5-check.md` (Acting on the report 1–9, Fix loop entry/close), `rules/do/step-7-9-post-change.md` item 5.
- `rules/do-workflow.md` — CHECK bullets (Minor/Suggestion, STOP gate, tick, "user may ask for a deferred Minor"), Fix loop entry condition, `--auto` Removed gates bullet ("default behavior in all runs" is no longer true for Minor).
- `rules/post-change.md` — COMMIT precondition 1.
- `agents/crafter-checker.md` — Recommendations block ("Record and continue" lumps Minor with Suggestion); the `--auto` note that Minor/Suggestion are excluded from the classification table stays.
- `README.md` (lines ~143, ~147), `docs/philosophy.md` (~19), `.crafter/ARCHITECTURE.md` (~90), `.crafter/STATE.md` (Recent Changes row; Current Focus may name v0.16.0).

### Steps
- [x] Rewrite the Minor rule in the rule modules: `rules/do-workflow.md`, `rules/do/step-5-check.md`, `rules/post-change.md`, `rules/do/step-7-9-post-change.md`, and the `--fast` error in `rules/do/flag-validation.md` — Minor stop with the three options, chosen Minors into the fix loop, deferred → `user-deferred` Decision, Suggestion auto-recorded, `--auto` unchanged.
- [x] Mirror the same behavior in `skills/crafter-do/SKILL.md` (Flag Validation error, Master Plan rows, Flag branching, Step 5 residue items, Step 6b path (1), Steps 7–9 item 5) so the residue letters and rule sub-steps say the same thing.
- [x] Update `agents/crafter-checker.md` Recommendations so Minor findings are listed in their own bucket ("needs a user decision: fix or defer") separate from record-and-continue Suggestions/beneficial drift.
- [x] Update commit-model prose in `README.md`, `docs/philosophy.md`, `docs/core-capabilities.md` (stale `--fast` silence-as-approval row), `.crafter/ARCHITECTURE.md`; add a dated Recent Changes row to `.crafter/STATE.md`.
- [x] Run a repo-wide search for leftover default-flow claims ("auto-proceed", "Minor … do not gate", "Minor/Suggestion → auto-record") outside `.crafter/tasks/` and historical STATE rows; run `tests/test_install.sh` to confirm nothing broke.
- [x] Check

### Contract
- **Outcome:** default flow stops on Minor with fix-all / pick / defer-all; Suggestion and `--auto` behavior unchanged in meaning.
- **Scope boundary:** markdown in `skills/crafter-do/`, `rules/`, `agents/crafter-checker.md`, `README.md`, `docs/philosophy.md`, `docs/core-capabilities.md`, `.crafter/ARCHITECTURE.md`, `.crafter/STATE.md`. Never `~/.claude/crafter/`.
- **Non-goals:** as listed above.
- **Seams:** see below.
- **Verification evidence:** grep output showing no stale auto-proceed-for-Minor claims in the default flow; side-by-side confirmation that each SKILL.md Step 5 residue item matches its `step-5-check.md` sub-step; `--auto` passages still auto-record Minor; `tests/test_install.sh` passes.
- **Stop conditions:** a needed edit would change `--auto`, severity definitions, the cap, or CLI parsing; or a SKILL.md/rule pair cannot be made consistent without restructuring — stop and report.

### Seams
- The `--fast` removal error string (identical in `skills/crafter-do/SKILL.md` and `rules/do/flag-validation.md`).
- Decision labels: existing `Decision (Tech Debt — auto-recorded)` (Suggestion; Minor+Suggestion under `--auto`) and new `Decision (Tech Debt — user-deferred)` (Minor, default flow).
- Step 5 residue (b)–(h) in SKILL.md ↔ `rules/do/step-5-check.md` sub-steps 1–9 and Fix loop 1–6.
- Step 6b path (1) conditions ↔ `rules/post-change.md` COMMIT precondition 1.
- Checker report Recommendations block (consumed by the orchestrator's relay).

### Alternatives considered
- Separate Minor-only fix mechanism: rejected — the existing fix loop already does delta passes and caps.
- Keep the single `auto-recorded` label for user-deferred Minors: rejected — the request asks for non-auto wording, and the CLI does not parse the label so a second label is free.
- Stop on Minors only on the first full pass: rejected as inconsistent; a delta-pass Minor is as real as a first-pass one.

### Risks / unknowns / flags
- `docs/core-capabilities.md` line ~46 still describes the removed `--fast` silence-as-approval path — pre-existing staleness, out of scope unless the user wants it folded in.
- Extra stops lengthen interactive runs when the Checker reports many Minors; mitigated by "defer all" as a one-word answer.
- Edge: if the delta pass keeps surfacing new Minors, each pass asks again; the cap bounds it.


## Decisions
- **Decision (User Accepted):** Added `docs/core-capabilities.md` to scope to fix its stale `--fast` silence-as-approval description. **Reason:** Same topic, one-line fix; flagged by the Planner.
- **Decision (Orchestrator Accepted):** `docs/core-capabilities.md` is fixed via a clause in its "Superseded snapshot" banner instead of rewriting historical table row 46. **Reason:** The document is a historical record; the banner already notes the `--fast` removal; same outcome, lower risk.

- **Decision (Auto-Fixed):** Minor — four consistency gaps found by the Checker, fixed in fix-loop iteration 1 on the user's "fix all" choice: cap bullet in `rules/do-workflow.md`, fix-loop entry in SKILL.md High-risk routing chains, Step 6b entry condition in SKILL.md, and a path for findings the reachability send-back lowers to Minor (`rules/do/step-5-check.md`, SKILL.md). Delta pass: 4/4 resolved, no new findings.

## Outcome
Default crafter-do flow now stops on Minor findings with fix all / pick numbers / defer all; chosen Minors go through the existing fix loop, deferred ones are recorded as `Decision (Tech Debt — user-deferred)`. Suggestions stay auto-recorded and `--auto` is unchanged. `--fast` error text, Checker Recommendations, README, philosophy, core-capabilities banner, ARCHITECTURE and STATE updated. Checker full pass: 0 Critical/Major, 4 Minor (all fixed in iteration 1), 1 beneficial local drift accepted. `tests/test_install.sh` 64/0. Commit: see PR on branch `feat/minor-findings-stop-gate`.
