# Task: GH#18 — Reviewer/Verifier/Implementer auto-fix-or-document semantics under --auto

## Metadata
- **Date:** 2026-05-10
- **Work branch:** feat/gh-18-auto-agent-semantics
- **Status:** completed
- **Scope:** Medium

## Request
Update sub-agent prompts (Reviewer, Verifier, Implementer) so that under `--auto` orchestration each agent makes a deterministic decision per finding/drift item: either fix it within scope, or record it in the appropriate buffer (`crafter-buffer uat` / `crafter-buffer gap`) and continue — never block.

**Desired outcome:**

- **Reviewer** classifies every finding into: `auto-fixable` (spawn Implementer fix loop), `manual-only` (→ UAT buffer), or `gap` (→ Gaps buffer).
- **Verifier** under `--auto`: fixable drift → loop with Implementer; manual-check drift → UAT buffer; out-of-scope drift → Gaps buffer.
- **Implementer** under `--auto`: may explicitly call `crafter-buffer uat` for manual-verification steps and `crafter-buffer gap` for out-of-scope constraints.

**Acceptance criteria:**

- [ ] Reviewer prompt under `--auto` produces a deterministic decision per finding (auto-fix / uat / gap)
- [ ] Verifier prompt under `--auto` produces a deterministic decision per drift item
- [ ] Implementer prompt under `--auto` knows when to call `crafter-buffer uat` and `crafter-buffer gap`
- [ ] Auto-fix loop has a documented retry budget; failure to fix within budget → entry in Gaps buffer
- [ ] Ad-hoc clarification escape hatch is documented: criteria, rare use, mechanism
- [ ] `--auto` runs do not block on standard severity findings
- [ ] Without `--auto`, prompts behave identically to today (no regression)

**Non-goals:**

- Changing default (non-`--auto`) behavior of any sub-agent
- New severity taxonomy beyond Critical/Major/Minor/Suggestion
- Cross-agent coordination beyond what already exists

**Dependencies:** #15 (flag — done), #16 (buffer skill — done), pairs with #17 (PR composer — done)

## Plan
**Plan status:** approved

### Complete request

**Goal:** Update the three sub-agent prompt files (`agents/crafter-reviewer.md`, `agents/crafter-verifier.md`, `agents/crafter-implementer.md`) so that under `--auto` orchestration, each agent makes a deterministic decision per finding or drift item — either fix it within scope, or record it in the appropriate buffer (`crafter-buffer uat` / `crafter-buffer gap`) and continue. The orchestrator already has `--auto` semantics for its own workflow (defined in `rules/do-workflow.md` and `skills/crafter-do/SKILL.md`), but the agents themselves have no awareness of `--auto` mode. This task bridges that gap.

**Why this matters:** Without agent-level `--auto` awareness, the orchestrator cannot run unattended end-to-end because agents will produce outputs that assume an interactive human is deciding what to do with findings. The orchestrator needs agents to classify and route findings deterministically so it can execute the removed-gates semantics from `rules/do-workflow.md`.

**Scope:** Medium — cross-cutting prompt changes across 3 agent markdown files plus a documentation update in `rules/do-workflow.md` to remove the "agent-side recognition is NOT defined here" placeholder.

**Acceptance criteria:**
- Reviewer prompt under `--auto` produces a deterministic decision per finding (auto-fix / uat / gap)
- Verifier prompt under `--auto` produces a deterministic decision per drift item (fix / uat / gap)
- Implementer prompt under `--auto` knows when to call `crafter-buffer uat` and `crafter-buffer gap`
- Auto-fix loop has a documented retry budget; failure to fix within budget results in a Gaps buffer entry
- Ad-hoc escape hatch is documented from the agent perspective: criteria, rare use, signal mechanism
- `--auto` runs do not block on standard severity findings
- Without `--auto`, prompts behave identically to today (no regression)

**Constraints:**
- Modify only source files in the repo (`agents/`, `rules/`), never installed copies
- New `--auto` sections must be clearly conditional — existing non-`--auto` behavior must not change
- Agents do not call `crafter-buffer` themselves; they produce structured output that the orchestrator uses to call the buffer CLI. The buffer CLI is a Bash tool, and agents like the Reviewer do not have Write/Edit tools. The Implementer does have Bash, but the pattern should be consistent: agents classify and report, the orchestrator acts on the classification.

**Validation strategy:** Manual review of agent prompts to confirm: (1) `--auto` sections are clearly conditional, (2) existing non-`--auto` output format and behavior are unchanged, (3) every finding type has a deterministic routing rule, (4) retry budget and escape hatch criteria are specified.

### Assumptions / interpretations

1. **Agents classify, orchestrator routes.** The issue says agents "call `crafter-buffer uat ...`" directly. However, examining the agent tool lists: the Reviewer has `Read, Grep, Glob, Bash` but its constraint says "Do not fix anything. Do not modify any file." The Verifier has the same constraint. Only the Implementer has Write/Edit tools and could plausibly call a Bash command to append to buffers. However, the cleaner architecture is for all three agents to produce structured classification output (e.g., annotating each finding with its routing decision), and for the orchestrator to act on that classification by calling `crafter-buffer` itself. This keeps agent prompts focused on analysis/classification rather than side-effecting, and is consistent with the existing separation of concerns (agents report, orchestrator acts). **I will flag this for the user** since the issue explicitly says agents "call" the buffer commands.

2. **Retry budget for auto-fix loop is already defined.** The review fix loop already has a 5-iteration cap (defined in `rules/do-workflow.md` → REVIEW section). Under `--auto`, cap-reached exits with state. This task does not need to define a new budget — it needs to make the agents aware that the orchestrator will use this existing budget, and that budget exhaustion with unresolved findings results in Gaps buffer entries. The Reviewer's `--auto` classification guides which findings enter the fix loop vs. go directly to buffers.

3. **The Reviewer's `--auto` classification is output metadata, not a behavioral change.** Under `--auto`, the Reviewer still reviews the same way and produces the same tables. The addition is a new section in its output format: a routing table that classifies each finding into `auto-fixable`, `manual-only (uat)`, or `gap`. The orchestrator reads this table to decide what to do with each finding.

4. **The Verifier's `--auto` classification maps to its existing recommendation system.** The Verifier already produces a recommended action per drift check (`continue`, `record decision and continue`, `fix current step`, `ask user`, `replan`). Under `--auto`, the "ask user" recommendation needs a sub-classification: is this something that can be downgraded to "record and continue" (UAT or Gaps buffer), or is it genuinely blocking (escape hatch)? The orchestrator already handles this downgrade logic (see `rules/do-workflow.md` → VERIFY → "Under `--auto`" section), but the Verifier needs to provide enough information for the orchestrator to make the right call.

5. **The `--auto` flag is passed to agents as context in the task prompt.** The orchestrator already knows whether `--auto` is active. When spawning agents under `--auto`, it includes this context in the task prompt. The agent prompts define what agents do differently when this context is present.

6. **The ad-hoc escape hatch placeholder in `rules/do-workflow.md`** (the "Scope boundary — agent-side recognition is NOT defined here" paragraph under `#### Ad-hoc escape hatch`) should be updated to reference the agent-side signal mechanism defined by this task.

### Non-goals

- Changing default (non-`--auto`) behavior of any sub-agent
- New severity taxonomy beyond Critical/Major/Minor/Suggestion
- Cross-agent coordination beyond what already exists
- Modifying the orchestrator skill (`skills/crafter-do/SKILL.md`) — the orchestrator already has its `--auto` workflow; this task defines what agents tell it
- Modifying the buffer skill (`skills/crafter-buffer/SKILL.md`) — the buffer CLI interface is stable
- Modifying the delegation rules (`rules/delegation.md`) — agent spawning mechanics are unchanged
- Adding new tools to any agent
- Changing the Implementer's model tier or context budget

### Relevant areas

**Primary files to modify:**
- `agents/crafter-reviewer.md` — add `--auto` classification section to output format
- `agents/crafter-verifier.md` — add `--auto` sub-classification guidance for "ask user" and blocking drift
- `agents/crafter-implementer.md` — add `--auto` awareness for buffer-worthy discoveries
- `rules/do-workflow.md` — update the escape hatch placeholder to reference agent-side signals

**Reference files (read-only, for context):**
- `skills/crafter-buffer/SKILL.md` — buffer CLI interface and entry shapes
- `skills/crafter-do/SKILL.md` — orchestrator workflow, Step 5, 5a, 6, 6b
- `rules/do-workflow.md` — `--auto` section, VERIFY section, REVIEW section
- `.crafter/ARCHITECTURE.md` — project patterns and conventions

### Approach

Add clearly conditional `--auto` sections to each of the three agent prompt files, and update the escape hatch placeholder in `rules/do-workflow.md`. The changes follow a consistent pattern: when the orchestrator passes `--auto` context, the agent adds a structured classification to its existing output format. No existing output format is changed; the `--auto` additions are appended sections. The Implementer gets guidance on when to report buffer-worthy items in its existing deviations/discoveries section.

---

### Phase 1: Agent `--auto` classification semantics

**Outcome:** All three agent prompts contain conditional `--auto` sections that define deterministic classification of findings/drift items, and the escape hatch placeholder in `rules/do-workflow.md` is updated to reference agent-side signals. The full system of `--auto` agent semantics is in place.

**Scope boundary:** Only the four files listed. Only `--auto`-conditional additions. No changes to existing non-`--auto` sections.

**Non-goals:** Modifying orchestrator behavior, buffer CLI, delegation rules, or any other file. Adding new tools to agents. Changing output format for non-`--auto` runs.

**Simplicity constraint:** Each agent gets one new clearly-labeled section. The section is conditional on the orchestrator passing `--auto` context. Existing sections are not reorganized or refactored.

**Drift criteria:** Drift if any existing non-`--auto` output format is altered. Drift if any agent is given new tools. Drift if changes touch files beyond the four listed. Drift if the escape hatch update introduces new orchestrator behavior (it should only document agent-side recognition, not change orchestrator logic).

**Verification evidence:** (1) Each agent file has a clearly labeled `--auto` section. (2) The section is conditional on orchestrator context. (3) Existing output format sections are byte-identical to before the change. (4) The escape hatch placeholder is replaced with concrete agent-side signal criteria. (5) `crafter-buffer` CLI invocation syntax referenced in agent prompts matches `skills/crafter-buffer/SKILL.md`.

**Stop conditions:** Stop if the task requires modifying the orchestrator skill or buffer skill. Stop if a consistent agent-classification-to-orchestrator-action mapping cannot be achieved without modifying orchestrator logic.

#### Steps

- [x] **Step 1.1 — Reviewer `--auto` classification**

  Add a conditional `--auto` section to `agents/crafter-reviewer.md` that defines the three-bucket classification for findings under `--auto` mode.

  **Outcome:** The Reviewer prompt includes an `--auto` section that instructs the agent to append a routing classification to its output when the orchestrator indicates `--auto` mode. Each finding from the Issues table gets classified as `auto-fixable`, `uat`, or `gap`, with clear criteria for each bucket.

  **Scope boundary:** Only `agents/crafter-reviewer.md`. Only add new content — do not modify any existing section.

  **Non-goals:** Changing the Reviewer's non-`--auto` output format. Adding buffer CLI invocation instructions to the Reviewer (the orchestrator calls the buffer CLI based on the Reviewer's classification).

  **Simplicity constraint:** The classification criteria should be a simple decision tree: Critical/Major within scope = auto-fixable; requires manual/external verification = uat; out-of-scope or deferred = gap. Minor/Suggestion findings are always recorded as tech debt by the orchestrator (per `rules/do-workflow.md` → `#### Removed gates`) and do not need classification.

  **Drift criteria:** Drift if any existing section of the Reviewer prompt is modified. Drift if the output format for non-`--auto` runs changes.

  **Verification evidence:** The Reviewer prompt has a new section clearly labeled for `--auto` mode. The section defines criteria for each of the three buckets. Critical/Major findings are routed to auto-fixable unless they meet uat/gap criteria. The section references the existing severity taxonomy without changing it.

  **Stop conditions:** Stop if the three-bucket classification requires a new severity level or changes to the existing severity definitions.

- [x] **Step 1.2 — Verifier `--auto` sub-classification**

  Add a conditional `--auto` section to `agents/crafter-verifier.md` that provides sub-classification guidance for drift items under `--auto` mode.

  **Outcome:** The Verifier prompt includes an `--auto` section that instructs the agent to enrich its existing recommendations with routing metadata when the orchestrator indicates `--auto` mode. Specifically, the "ask user" and "record decision and continue" recommendations need additional context: whether the item is fixable by code change, requires manual verification (uat), is out of scope (gap), or is genuinely blocking (escape hatch signal).

  **Scope boundary:** Only `agents/crafter-verifier.md`. Only add new content — do not modify any existing section.

  **Non-goals:** Changing the Verifier's existing drift classification system or recommendation actions. Adding buffer CLI invocation instructions.

  **Simplicity constraint:** The `--auto` enrichment maps onto the existing five recommendation categories. "continue" and "fix current step" need no enrichment. "record decision and continue" and "ask user" get routing metadata. "replan" becomes an escape-hatch signal.

  **Drift criteria:** Drift if any existing drift classification or recommendation action is modified. Drift if the output format for non-`--auto` runs changes.

  **Verification evidence:** The Verifier prompt has a new section clearly labeled for `--auto` mode. The section maps each existing recommendation to its `--auto` behavior. The escape hatch signal criteria are listed.

  **Stop conditions:** Stop if the sub-classification requires changing the existing five recommendation categories.

- [x] **Step 1.3 — Implementer `--auto` buffer awareness**

  Add a conditional `--auto` section to `agents/crafter-implementer.md` that defines when the Implementer should report buffer-worthy discoveries.

  **Outcome:** The Implementer prompt includes an `--auto` section that instructs the agent to explicitly flag buffer-worthy items in its deviations/discoveries output when the orchestrator indicates `--auto` mode. Two categories: items needing manual verification (uat-worthy) and items that are out of scope (gap-worthy).

  **Scope boundary:** Only `agents/crafter-implementer.md`. Only add new content — do not modify any existing section.

  **Non-goals:** Having the Implementer call `crafter-buffer` directly. Changing the Implementer's existing output format. Adding new output sections for non-`--auto` runs.

  **Simplicity constraint:** The Implementer already reports "deviations/discoveries" in its output. Under `--auto`, it adds a classification tag to each deviation/discovery indicating whether it is uat-worthy, gap-worthy, or neither. The orchestrator reads these tags and calls the buffer CLI.

  **Drift criteria:** Drift if any existing section of the Implementer prompt is modified. Drift if the output format for non-`--auto` runs changes.

  **Verification evidence:** The Implementer prompt has a new section clearly labeled for `--auto` mode. The section defines clear criteria for uat-worthy vs. gap-worthy items. The existing deviations/discoveries output format is preserved for non-`--auto` runs.

  **Stop conditions:** Stop if tagging deviations requires modifying the existing output format section (rather than appending to it).

- [x] **Step 1.4 — Escape hatch agent-side recognition update**

  Update `rules/do-workflow.md` to replace the "agent-side recognition is NOT defined here" placeholder with concrete agent-side signal criteria.

  **Outcome:** The escape hatch section in `rules/do-workflow.md` documents how agents signal genuinely blocking conditions under `--auto` mode, referencing the agent-side mechanisms defined in Steps 1.1–1.3.

  **Scope boundary:** Only `rules/do-workflow.md`, specifically the `#### Ad-hoc escape hatch` subsection's "Scope boundary" paragraph. Only replace the existing placeholder — do not modify other parts of the escape hatch section.

  **Non-goals:** Changing the escape hatch trigger conditions or behavior. Changing orchestrator logic. Adding new retained gates.

  **Simplicity constraint:** The update should be a brief paragraph replacing the existing placeholder, referencing the agent prompts as the source of truth for how each agent signals a blocker.

  **Drift criteria:** Drift if the escape hatch trigger conditions or behavior are changed. Drift if new retained gates are introduced.

  **Verification evidence:** The placeholder paragraph is replaced with a concrete reference to agent-side signals. The replacement does not change the meaning of the escape hatch section — it only fills in the "how agents signal" part that was explicitly deferred to this task.

  **Stop conditions:** Stop if the agent-side signals defined in Steps 1.1–1.3 are insufficient to cover the escape hatch trigger conditions.

- [x] **Phase 1 gate — All agent prompts have `--auto` sections and escape hatch placeholder is resolved**

  **Phase verification criteria:**
  1. `agents/crafter-reviewer.md` contains a conditional `--auto` section with three-bucket classification (auto-fixable, uat, gap)
  2. `agents/crafter-verifier.md` contains a conditional `--auto` section with routing metadata for recommendations
  3. `agents/crafter-implementer.md` contains a conditional `--auto` section with buffer-worthy discovery tagging
  4. `rules/do-workflow.md` escape hatch placeholder is replaced with agent-side signal documentation
  5. No existing non-`--auto` sections are modified in any agent file (byte-identical check via `git diff` on non-`--auto` sections)
  6. All `crafter-buffer` CLI references match the syntax in `skills/crafter-buffer/SKILL.md`
  7. The retry budget references the existing 5-iteration cap from `rules/do-workflow.md` → REVIEW section

### Alternatives considered

1. **Agents call `crafter-buffer` directly via Bash.** This matches the issue's literal wording but breaks the separation of concerns: agents analyze and report, the orchestrator acts. The Reviewer explicitly cannot modify files. The Implementer could call Bash, but having some agents call buffers directly and others not would be inconsistent. Rejected in favor of agents classifying and the orchestrator routing.

2. **Separate `--auto` agent files (e.g., `crafter-reviewer-auto.md`).** This would avoid any risk of regression in non-`--auto` behavior but would duplicate the entire agent prompt. Rejected because the `--auto` additions are small, clearly conditional sections that do not interact with existing logic.

3. **Modify only the orchestrator to interpret existing agent output differently under `--auto`.** The orchestrator already handles some of this (e.g., the Verifier's "ask user" downgrade). However, the agents need to provide enough structured information for the orchestrator to make routing decisions (e.g., why something is uat vs. gap). Without agent-side classification, the orchestrator would have to re-analyze findings, which defeats the purpose of delegation.

### Risks / unknowns / flags

1. **Flag: Agents calling buffers vs. agents classifying.** The issue says agents "call `crafter-buffer uat ...`" but the architecture separates analysis from action. This plan proposes agents classify and the orchestrator calls the buffer CLI. The user should confirm this interpretation before implementation proceeds. If the user prefers agents calling buffers directly, the Implementer's step would need to include direct Bash invocation instructions, and the Reviewer/Verifier prompts would need a different approach (they cannot write files by constraint).

2. **Risk: Output format additions under `--auto` may confuse the orchestrator.** The orchestrator currently parses Reviewer output for severity tables and Verifier output for recommendations. Adding `--auto` classification sections means the orchestrator needs to know to look for them. However, the orchestrator already knows whether `--auto` is active, so it can conditionally parse the extra sections. This is an orchestrator concern that is already handled — the orchestrator's `--auto` logic in `skills/crafter-do/SKILL.md` Step 6 and Step 5 already expects to make routing decisions based on agent output.

3. **Unknown: Exact output format for the `--auto` classification.** The plan defines what information agents should provide but leaves the exact markdown format to the Implementer. The format should be a simple table or list that the orchestrator can reliably parse. This is a local implementation choice within the contract.

## Decisions
- **Decision (Auto-Fixed):** Major — `do-workflow.md` escape hatch summary falsely claimed Reviewer emits blocker signals; fixed to accurately state Reviewer's three-bucket tree is exhaustive with no escape-hatch path.
- **Decision (Tech Debt — auto-recorded):** Minor — Section title naming inconsistency across agents (`--auto classification` vs `--auto sub-classification` vs `--auto mode`).
- **Decision (Tech Debt — auto-recorded):** Minor — Verifier has no `[no-buffer]` equivalent for resolved local drift items, creating asymmetry with Implementer.
- **Decision (Tech Debt — auto-recorded):** Minor — Dense single paragraph in `do-workflow.md` escape hatch summary; consider splitting into bullet list.
- **Decision (Tech Debt — auto-recorded):** Suggestion — Escape hatch criteria duplicated across Verifier, Implementer, and `do-workflow.md`.
- **Decision (Tech Debt — auto-recorded):** Suggestion — Implementer `--auto` section doesn't explicitly cross-reference existing blocker output format.

## Outcome
- **Commit:** 498ecec
- **Summary:** Added conditional `--auto` sections to all three sub-agent prompts (Reviewer, Verifier, Implementer) defining deterministic classification of findings and drift items for unattended orchestration. Reviewer classifies Critical/Major findings into auto-fixable/uat/gap buckets. Verifier enriches recommendations with routing metadata (gap/uat/escape-hatch). Implementer tags deviations with [uat-worthy]/[gap-worthy]/[no-buffer]. Escape hatch placeholder in `rules/do-workflow.md` replaced with concrete agent-side signal documentation. One Major review finding (Reviewer escape-hatch contradiction) was auto-fixed during the review loop. Completes the `--auto` pipeline (GH#15→16→17→18).
