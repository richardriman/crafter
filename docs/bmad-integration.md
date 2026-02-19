# BMAD Integration

Crafter is designed to work completely on its own. But for decisions that benefit from multiple perspectives, it pairs well with BMAD party mode.

---

## What Is BMAD Party Mode?

BMAD party mode is a multi-agent discussion simulation. Claude takes on several roles simultaneously — typically Product Manager, Architect, Senior Developer, and QA — and has them debate a topic. Each role brings a different lens:

- The **PM** thinks about user impact and priorities
- The **Architect** thinks about system design and long-term consequences
- The **Developer** thinks about implementation complexity and code quality
- The **QA** thinks about edge cases, failure modes, and testability

The result is a richer discussion than any single perspective would produce.

---

## When to Use It

BMAD party mode is most valuable when:

- You're making an **architectural decision** with significant tradeoffs (e.g. choosing a data model, picking a communication pattern between services)
- You're **choosing between multiple valid approaches** and want to pressure-test each one
- A change has **cross-domain impact** that's hard to reason about from a single perspective
- You're doing a **brainstorm or ideation session** and want to explore the problem space broadly
- You're running a **post-mortem** and want multiple angles on what went wrong

For day-to-day tasks — implementing a feature, fixing a bug, refactoring a module — Crafter handles it directly. BMAD party mode is for the big decisions.

---

## How It Works Alongside Crafter

1. **Identify the decision** — recognize that a task involves a significant tradeoff or architectural choice that would benefit from multiple perspectives
2. **Run a BMAD party mode session** — bring the question to Claude and ask it to run a multi-agent discussion
3. **Capture the conclusions** — note the key decisions and their rationale
4. **Feed into .planning/ files:**
   - Technical decisions → `.planning/ARCHITECTURE.md` (Key Patterns & Decisions section)
   - Strategic decisions → `.planning/PROJECT.md` (Key Decisions table)
5. **Proceed with `/crafter:do`** — implement the agreed-upon approach with full Crafter workflow

---

## Important

**Crafter does not depend on BMAD.** You don't need to install BMAD, configure it, or use it for anything. The integration is purely advisory — when Crafter detects a decision that could benefit from multi-perspective analysis, it may suggest running a BMAD party mode session. You can always skip it.

Think of it as a recommendation from a colleague: "This is a big enough decision that it might be worth talking it through before we start building." You're free to say "no, let's just do it."
