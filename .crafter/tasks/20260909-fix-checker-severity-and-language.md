# Task: Calibrate checker severity and propagate the user's language to subagents

## Metadata
- **Date:** 2026-09-09
- **Work branch:** fix/checker-severity-and-language
- **Status:** completed
- **Scope:** Medium

## Request

Calibrate `crafter-checker` severity and propagate the user's conversation language to subagents. Source files in the repo (`agents/`, `rules/`), never `~/.claude/crafter/`. Four changes:

1. `agents/crafter-checker.md` severity list (lines 57–60): add a gate — Critical and Major require a reachable trigger. Before assigning either, the checker must name the concrete input, caller, or sequence that produces the failure and how reachability was established (a caller read, a test run, a probe executed). "A caller could send X" is not reachability; "the only caller is `lib/tasks/is.rake`, line N, sends X" is. If no current caller produces the input and none is planned in the task's contract, the finding is at most Minor and must say which callers were checked. Exception: genuinely open trust boundaries (unauthenticated / third-party traffic) are reachable by definition — there assume hostile input exists. Do not suppress security findings on public surfaces.

2. `agents/crafter-checker.md` after the "Report completely" rule (line 62): add that an empty findings table is a valid, expected outcome (especially on a delta pass over a small or deletion-only change). Report it plainly and stop. Do not pad with observations that describe correct behavior, restate the diff, or comment on wording in another agent's prose report. If the only thing a row can say is that something *could* be different, it belongs in Suggestion or nowhere. Keep the existing "Report completely" rule intact.

3. `rules/do/step-5-check.md` sub-step 3 (Critical/Major → STOP): before entering the fix loop (interactive and `--auto`), the orchestrator verifies each Critical/Major finding names a reachable trigger per change 1. A finding without one is downgraded to Minor (recorded as tech debt) or sent back to the checker for a reachability check — orchestrator picks; state the rule explicitly. Do NOT weaken the "no proceed anyway" rule, do NOT touch the verbatim-relay rule (line 14), do NOT reduce checker passes or scope.

4. Language propagation. `rules/core.md` Language Rules cover only the orchestrator; subagent reports come back in English even when the user writes Czech, and are relayed verbatim. Fix: in `rules/delegation.md` (alongside the caveman/ponytail directives) the orchestrator passes the user's conversation language into every spawned agent's prompt; agents write free-text prose of their returned reports (finding descriptions, recommendations, summaries) in that language. Unchanged and always English: code, identifiers, file paths, required headings/table columns/status line formats, and all persistent files (`.crafter/*`, task files, plans — `task-lifecycle.md` rule stays). Mirror a one-line note in `rules/core.md` Language Rules so the two files agree. Crafter must not target Czech specifically — it respects whatever language the user uses.

Not requested: ponytail for the checker (deliberately skipped). Context brief: `crafter-checker-severity-brief.md` in repo root (untracked; do not commit it).

## Plan

**Plan status:** approved

### Goal and why it matters

Four targeted prose edits across five source files. Two calibrate `crafter-checker` severity (a three-state reachability gate on Critical/Major, and permission to report an empty findings table), one makes the orchestrator send stateless Critical/Major findings back to the Checker instead of downgrading them itself, and one propagates the user's conversation language into every spawned agent so returned prose comes back in the user's language instead of always English. Motivation is in `crafter-checker-severity-brief.md` (untracked, repo root): a run burned two fix-loop iterations on a Major finding no caller could produce, and a third pass produced seven non-defects because a findings table was expected.

Acceptance: the four changes are present in the repo source files, the three protected rules are byte-identical afterwards, and `rules/do/step-5-check.md` and `skills/crafter-do/SKILL.md` state the same rule rather than disagreeing.

### Assumptions / interpretations

- The wording quoted in `## Request` and in the brief is a **specification of intent**, not text to paste verbatim. The Implementer may tighten the prose to match each file's register, but every named condition must survive: how reachability was established, "a caller could send X" is not reachability, the at-most-Minor fallback gated on positive evidence, "must say which callers were checked", and — overriding the brief's narrower wording — the **three-state** model below. "Cannot establish reachability" is not "unreachable"; unknown reachability defaults to reachable and keeps the severity. The brief's open-trust-boundary exception is absorbed as one case of state 3.
- The three states the Checker must choose between for every Critical/Major finding:
  1. **Verified reachable** — the caller / input / sequence is found and named → severity stands.
  2. **Verified unreachable** — callers were checked and are named (files/lines), none produces the input, none is planned in the contract, and the surface is not an open trust boundary → at most Minor, and the finding states which callers were checked. Downgrade only on this positive evidence.
  3. **Cannot determine** — external or unknown consumers, public API, library code, third-party or unauthenticated traffic, or the check simply did not settle it → severity **kept** as Critical/Major.
- The language directive is **unconditional** — emitted on every spawn for all four agents, independent of the caveman/ponytail markers. It therefore cannot live inside `delegation.md` § "Skill Directives — Caveman and Ponytail" item list, whose §3 says "when both markers are absent, append nothing". It goes alongside that section as its own always-on rule.
- "The user's conversation language" is whatever the orchestrator already detects per `rules/core.md` Language Rules. No detection mechanism, marker file, or config is introduced — the orchestrator just names the language in the spawn prompt.
- Because `delegation.md` is loaded by `crafter-debug` and `crafter-map-project` too, putting the rule there gives those orchestrators parity for free. Only `crafter-do`'s `## Pre-Spawn Gate` needs a pointer, because it is the file the orchestrator actually reads at spawn time.
- No agent file needs editing for change 4. There are no `## Behavior under caveman` sections in `agents/` today (ARCHITECTURE.md line 140 is stale on this), and `agents/crafter-planner.md:44` is already scoped to the plan written into the task file, not to the returned summary. Spawn-time directive suffices — the smaller change.
- **Only the Checker assigns or lowers severity.** The orchestrator is a dispatcher with no code context; it never downgrades. Its single action on a Critical/Major finding that names neither a trigger nor a reachability state is to **send it back to the Checker** for a targeted reachability check. Behavior is identical interactive and under `--auto` — no mode-dependent default.
- The send-back is a **cheap Checker-only pass**, not a fix-loop iteration: no Implementer spawn, no code change, so it does **not** increment the iteration count and does not touch the 5-iteration cap. It sits before fix-loop entry.
- The send-back is bounded: at most one per finding. Whatever the Checker returns the second time is final — in practice state 3 (severity kept), which is the safe default.
- The relay is untouched: the send-back happens *after* the verbatim relay of the report that triggered it, and the report it returns is itself relayed verbatim.

### Non-goals

- Ponytail for the checker (brief's change 3) — deliberately skipped by the user, do not add it.
- Reducing checker passes, narrowing its review scope, or weakening "no proceed anyway" for Critical/Major.
- Touching the verbatim-relay rule (`rules/do/step-5-check.md` line 14 and its SKILL.md mirror `(a)`).
- Touching `rules/task-lifecycle.md` line 11 (task files always English), the Checker's Output-format tables, its `## Behavior under --auto` classification table, the 5-iteration cap, or the `gap`/`uat`/`auto-fixable`/`escape-hatch` buckets.
- Any change under `~/.claude/crafter/` or `~/.claude/agents/`, the Go CLI, `install.sh`, or the templates.
- Committing `crafter-checker-severity-brief.md`; it stays untracked.
- Translating persistent artifacts — task files, plans, `.crafter/*` stay English.
- Refreshing the stale ARCHITECTURE.md line 140 claim; leave it to the Steps 7–9 docs check.

### Relevant areas

- `agents/crafter-checker.md` — severity list (lines 57–60) and the "Report completely" paragraph (line 62). Both changes land in § Task, above § Delta pass.
- `rules/do/step-5-check.md` — § "Acting on the report" sub-step 3 (line 22); line 14 relay and the fix-loop section are context, not edit targets.
- `skills/crafter-do/SKILL.md` — Step 5 orchestrator residue `(d)` (line 223) and `## Pre-Spawn Gate` (lines 88–90).
- `rules/delegation.md` — § "Skill Directives — Caveman and Ponytail" (lines 41–69).
- `rules/core.md` — § Language Rules (lines 3–8).
- Read-only context: `crafter-checker-severity-brief.md`, `.crafter/ARCHITECTURE.md`.

### Steps

- [x] Step 1: Add the reachability gate to the Checker's severity definitions in `agents/crafter-checker.md`. The Checker is the one who establishes the trigger — it has Read/Grep/Glob/Bash and must actually look (grep the callers, read them, run a probe) before assigning Critical or Major, and must record which of the three states applies: **verified reachable** (names the input/caller/sequence and how it was established → severity stands), **verified unreachable** (names the callers checked with files/lines, none produces the input, none planned in the contract, surface is not an open trust boundary → at most Minor, listing those callers), or **cannot determine** (external/unknown consumers, public API, library code, third-party or unauthenticated traffic, or the check did not settle → severity **kept**). State explicitly that "a caller could send X" is not reachability, that a downgrade requires the positive evidence of state 2, and that uncertainty keeps the severity — unknown reachability defaults to reachable. Do not suppress security findings on public surfaces. The four existing severity bullets keep their current meaning.
- [x] Step 2: Add the empty-findings-table legitimacy rule to `agents/crafter-checker.md` after the "Report completely" paragraph, covering: an empty table is a valid and expected outcome (notably on a delta pass over a small or deletion-only change), report it plainly and stop, no padding with correct-behavior observations, diff restatements, or comments on another agent's prose report; "could be different" belongs in Suggestion or nowhere. "Report completely" itself stays intact.
- [x] Step 3: State the orchestrator's send-back rule in `rules/do/step-5-check.md` sub-step 3, and mirror it in `skills/crafter-do/SKILL.md` Step 5 residue `(d)` so the runtime file and the rule file agree. Content: the orchestrator **never downgrades a severity** — it has no code context. Before entering the fix loop, a Critical/Major finding that names neither a trigger nor one of the three reachability states is **sent back to the Checker** for a targeted reachability check; findings that already name a trigger, or already say "cannot determine — severity kept", go straight to the fix loop. The send-back is a Checker-only pass (no Implementer spawn), so it does **not** increment the fix-loop iteration count and leaves the 5-iteration cap untouched; at most one send-back per finding. Only the Checker's state 2 lowers a severity. Behavior is identical interactive and under `--auto`. The send-back happens after the verbatim relay, and the report it returns is relayed verbatim too; the "no proceed anyway" sentence and the relay rule stay byte-identical in both files.
- [x] Step 4: Make the orchestrator pass the user's conversation language into every spawn — an always-on rule in `rules/delegation.md` next to the marker-gated skill directives, naming what the agents write in that language (finding descriptions, recommendations, summaries — free-text prose of the returned report) and what stays English regardless (code, identifiers, file paths, required headings, table columns, status-line formats, and all persistent files). Extend `skills/crafter-do/SKILL.md` § Pre-Spawn Gate so the orchestrator applies it at spawn time, and mirror one line into `rules/core.md` § Language Rules so the two files agree. Crafter must not name Czech or any specific language.
- [x] Check

### Contract

- **Outcome:** the four prose changes exist in `agents/crafter-checker.md`, `rules/do/step-5-check.md`, `skills/crafter-do/SKILL.md`, `rules/delegation.md`, and `rules/core.md`; the three-state reachability gate, the empty-table rule, the orchestrator send-back rule, and the unconditional language directive are each stated once and consistently across the rule file and the SKILL.md mirror. Severity is assigned and lowered by the Checker only; unknown reachability keeps Critical/Major.
- **Scope boundary:** those five files only. Markdown prose only — no Go, no shell, no frontmatter changes, no new files.
- **Non-goals:** as listed above; in particular no checker ponytail, no weakening of "no proceed anyway", no edit to the verbatim relay or to `rules/task-lifecycle.md`.
- **Seams:** listed in the next section.
- **Verification evidence:** for each of the five files, a `grep -n` hit showing the new text; a `grep -n` hit showing all three reachability states present in `agents/crafter-checker.md` and the "orchestrator never downgrades" wording present in both `rules/do/step-5-check.md` and `skills/crafter-do/SKILL.md`; `git diff` showing no hunk on `rules/do/step-5-check.md` line 14, on either "no proceed anyway" sentence, on the 5-iteration-cap text, or on `rules/task-lifecycle.md`; `bash tests/test_install.sh` green; `git status` showing `crafter-checker-severity-brief.md` still untracked and not staged.
- **Stop conditions:** a change appears to require editing a sixth file, an agent file, or anything under `~/.claude/`; the language directive cannot be made unconditional without restructuring `delegation.md` § Skill Directives; the send-back cannot be phrased without softening "no proceed anyway" or without touching the iteration count / 5-cap; the send-back appears to require a new Checker mode beyond `full pass` / `delta pass` semantics; `tests/test_install.sh` fails.

### Seams

- `agents/crafter-checker.md` § Task — the severity list and the completeness paragraph are the Checker's prompt contract. The severity vocabulary (`Critical` / `Major` / `Minor` / `Suggestion`) and every Output-format table must stay exactly as-is.
- `rules/do/step-5-check.md` § "Acting on the report" sub-step 3 — the send-back rule's canonical home.
- `skills/crafter-do/SKILL.md` Step 5 residue `(d)` — the runtime mirror; must state the same rule as sub-step 3.
- The fix-loop iteration count and its 5-cap (`step-5-check.md` § Fix loop sub-step 1, SKILL.md `(f.1)`) — must not move; the send-back sits outside them.
- `skills/crafter-do/SKILL.md` § Pre-Spawn Gate — the single enforcement point for everything appended to a spawn prompt.
- `rules/delegation.md` § "Skill Directives — Caveman and Ponytail" and its neighbourhood — the propagation rule's home; the existing `## Active skill directives` block format is not to be repurposed for the language rule.
- `rules/core.md` § Language Rules — the four bullets are the canonical statement the mirror must not contradict.
- Must not move: `rules/do/step-5-check.md` line 14; the "There is no 'proceed anyway'" sentence in both files; `rules/task-lifecycle.md` line 11.

### Alternatives considered

- **Language rule inside the caveman/ponytail directive block.** Rejected: that block is emitted only when a marker exists, so the language would silently drop for users running without either skill.
- **A `## Report language` section in each of the four agent files.** Rejected: four edits instead of one, and the spawn-time directive already reaches every agent. Prefer the smaller diff.
- **Orchestrator downgrades an unproven Critical/Major to Minor.** Rejected by the user: the orchestrator has no code context, so a downgrade there is a guess, and "cannot establish reachability" would silently become "unreachable". Severity moves only on the Checker's positive evidence.
- **Two-state model (reachable / not reachable).** Rejected: it forces the unknown case into "not reachable" and loses weight on exactly the findings that matter — public APIs, library code, unauthenticated traffic. The third state makes unknown default to reachable.
- **Send-back counted as a fix-loop iteration.** Rejected: it spawns no Implementer and changes no code, so charging it against the 5-cap would spend the budget on bookkeeping.
- **Reachability check only in `rules/do/step-5-check.md`.** Rejected: SKILL.md Step 5 `(d)` is what the orchestrator reads at runtime; leaving it silent would let the two disagree, the exact failure the 2026-07-03 fix addressed.
- **Making the Checker refuse to emit a report without reachability.** Rejected: outside the request, and it would collide with "report completely".

### Risks / unknowns / flags

- **Send-back costs a round trip.** The gate trades one cheap Checker spawn for avoiding a wrong downgrade. Bounded at one send-back per finding; if the Checker still cannot determine reachability, the severity stands and the fix loop runs.
- **Checker self-consistency.** After Step 1 the Checker should rarely emit a stateless Critical/Major, so the send-back is a backstop for older/partial reports, not the normal path. If the Implementer finds the two rules reading as duplicated enforcement, keep both — the Checker states, the orchestrator only asks.
- **Send-back mode vocabulary.** The Checker has two modes (`full pass`, `delta pass`). The send-back is a targeted question, not a third review mode; if it cannot be phrased without inventing a mode, stop and ask.
- **Prose-only change, no test coverage.** Verification is grep and `git diff` only; a regression here surfaces at runtime, not in CI.
- **Stale doc noted, not fixed.** `.crafter/ARCHITECTURE.md` line 140 says each agent file carries a `## Behavior under caveman` section; none do. Out of scope here, worth a separate cleanup.

## Decisions

- **Decision (Orchestrator Accepted):** Fix-loop iteration 1 edited `agents/crafter-checker.md` line 28 (mode sentence) to admit the named variant `delta pass` with `reachability check on finding #N`. **Reason:** Smallest change that lets the Checker accept a send-back spawn; stays inside the five-file boundary and does not add a third top-level mode (contract stop condition 4 not triggered).
- **Decision (Tech Debt — auto-recorded):** Minor — `rules/do/step-5-check.md:24` cross-reference "sub-step 5 below" is ambiguous (Acting-on-the-report sub-step 5 vs Fix loop sub-step 5); SKILL.md mirror uses the unambiguous "(f.5)".
- **Decision (Tech Debt — auto-recorded):** Minor — `agents/crafter-checker.md:83` reachability-check variant has no defined output shape; the delta-pass summary line / Prior-findings-status table cannot express the three reachability states.
- **Decision (Tech Debt — auto-recorded):** Minor — `rules/delegation.md:75` rationale calls `.crafter/run/` buffers "persistent files" while `rules/do-workflow.md:116-129` calls the directory scratch, gitignored, and deleted on cleanup; the rule itself stands on `rules/core.md:7`.
- **Decision (Tech Debt — auto-recorded):** Minor — `rules/do-workflow.md:36` `#### Check pass` summary bullet does not mention the reachability send-back (third mirror, outside the five-file scope).
- **Decision (Tech Debt — auto-recorded):** Suggestion — `skills/crafter-do/SKILL.md:88` heading "Pre-Spawn Gate — Skill Directives" no longer covers the unconditional Report Language pointer it now contains.
- **Decision (Tech Debt — auto-recorded):** Suggestion — `rules/core.md:9` Language Rules bullet is ~3× longer than its siblings.
- **Decision (Tech Debt — auto-recorded):** Suggestion — `.crafter/ARCHITECTURE.md` § Skill Adaptation (~line 140) describes only marker-gated caveman/ponytail propagation and claims agent files carry `## Behavior under caveman` sections (they do not); the always-on language rule and the send-back are absent. Deferred to the Steps 7–9 docs check.

## Outcome

Committed as `f9367c4` on `fix/checker-severity-and-language`. All four changes landed in the five contract files (+23/−2 after the fix loop), protected lines byte-identical, `tests/test_install.sh` 64/0. Checker full pass raised 1 harmful drift (send-back had no Checker mode) and 2 Major (probe wording vs Bash-only rule; missing buffer-bound-text carve-out in the language rule) — all fixed in fix-loop iteration 1 together with 2 Minor; delta pass closed clean with 3 new Minor recorded as tech debt. The brief `crafter-checker-severity-brief.md` stays untracked. Deviation from the request as written: the orchestrator never downgrades (the request offered downgrade-or-send-back); severity is assigned and lowered by the Checker only, per user decision during plan revision 1.
