# Nightshift

> The factory floor where no lights are needed. Lights out, you sleep, code ships.

**Repo description:** Autonomous AI development pipeline

## Concept

Nightshift is an autonomous AI software development pipeline inspired by the "Dark Factory" concept — lights-out manufacturing applied to software. You define intent and acceptance criteria, the pipeline delivers tested code.

## Origin

Evolved from experience building Crafter (supervised AI dev workflow). Key insight: human interventions in the development loop converge to repetitive, automatable patterns ("fix all major, all minor, maybe this suggestion"). Instead of a human routing every step, the pipeline runs autonomously and escalates only when it hits a problem it can't resolve.

## Core Philosophy

- **Supervised (Crafter):** "I want control over every step"
- **Autonomous (Nightshift):** "I want the result"

## Key Components (inherited from Crafter)

- **Planner** — creates implementation plan from intent
- **Implementer** — writes code according to plan
- **Reviewer** — judges code quality, bugs, security, plan deviations
- **Verifier** — runs tests, checks verification criteria

## New / Changed Components

### Autonomous Orchestration
No human gates between steps. Pipeline runs end-to-end:
```
Intent + acceptance criteria --> Plan --> Implement --> Review --> Auto-fix --> Verify --> Done (or escalate)
```

### Test Judge (new agent)
Validates that generated tests are not bullshit:
- Do tests cover all acceptance criteria defined by the human?
- Do they test actual behavior or just implementation details?
- Are there negative tests (what must NOT happen)?
- Could a trivial/nonsensical implementation pass these tests? (mutation testing thinking)

### Human-Defined Acceptance Criteria
The human defines WHAT must be verified, not HOW:
- "User without permissions must not see other users' data"
- "When API fails, the order must not be lost"
- "Empty input must not crash the application"

This is the contract the pipeline is held to. AI generates the tests, Test Judge validates they actually enforce the contract.

## Automation Levels

1. **Semi-dark** — pipeline runs autonomously, stops at Critical findings, escalates to human
2. **Full-dark** — pipeline runs fully including post-review fixes, human only approves final result
3. **Ultra-dark** — even merge is automatic, human defines intent and monitors outcomes

## Circular Validation Problem (why Test Judge exists)

Without Test Judge, the pipeline risks circular validation:
- AI writes code
- AI writes tests that verify what the code DOES (not what it SHOULD do)
- Tests pass --> everyone happy --> but nothing was actually verified

Bullshit tests manifest as:
1. **Tautological** — test repeats the implementation ("function returns X" --> `expect(fn()).toBe(X)`)
2. **Weak** — cover happy path only, skip edge cases and error states

Test Judge breaks this cycle by independently evaluating test quality against human-defined acceptance criteria.

## Development Methodology

**GSD (Get Shit Done)** workflow for building Nightshift.

### Phase 1: Spike-driven discovery

Before building the product, validate core assumptions through time-boxed experiments (spikes). A spike is NOT a prototype — its output is knowledge, not code. Spike code is intentionally throwaway.

**How a spike works:**
1. **Question** — what exactly do you want to find out
2. **Timebox** — how much time you give it (typically hours, max a day). When time's up, stop and evaluate
3. **Success criterion** — how you'll know you have an answer

**Process:**
1. Identify uncertainty that can't be resolved by thinking alone
2. Build minimal code/prompt/setup to answer the question
3. Evaluate: did the spike answer the question? Record the conclusion. Did it open new questions? Create new spikes.
4. Throw away the spike code. The deliverable is a written finding, not code.

**Candidate spikes for Nightshift:**

| Spike | Question | Timebox | Method |
|---|---|---|---|
| S1 | Can LLM reliably detect tautological tests? | 2h | Give it 10 tests (5 good, 5 tautological), measure accuracy |
| S2 | Where exactly did I intervene in the last 10 Crafter tasks? | 1h | Review history, categorize interventions |
| S3 | Can pipeline auto-fix a Major review finding without human? | 3h | Take a real task, run Crafter, but let rules decide instead of human |
| S4 | What format of acceptance criteria works best? | 2h | Try 3 formats on the same task, compare output quality |

**After 5-10 spikes:** extract patterns, define architecture, then move to Phase 2.

### Phase 2: Systematic GSD milestone/phase development

Once spike findings are documented, switch to structured GSD workflow with milestones and phases for building the actual product.

**Crafter stays separate** — a finished product for supervised workflows. Nightshift inherits agent definitions as building blocks but has its own orchestration and identity.
