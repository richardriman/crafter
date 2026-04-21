---
name: "crafter-status"
description: "Show current project state from .crafter/STATE.md (with .planning fallback)"
---

Resolve context directory:
- Prefer `.crafter/STATE.md`.
- If only legacy `.planning/STATE.md` exists, use it as fallback and proactively suggest migration via `git mv .planning .crafter`.

Read the resolved `STATE.md` file and display its full contents.

Then provide a brief, conversational summary covering:

- What is currently being worked on
- What was recently completed
- What is coming up next
- Any known issues or blockers worth flagging

Keep the summary concise — a few sentences is enough. The full STATE.md content already provides the detail.
