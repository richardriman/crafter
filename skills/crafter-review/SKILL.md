---
name: "crafter-review"
description: "Context-free adversarial review of a PR, a base..head range, or the working tree"
---

Usage: `/crafter-review [<PR>|<base..head>]` — a PR URL or number, a `base..head` range, or nothing for the working tree (uncommitted changes against `HEAD`, including untracked files). Optional paths after the range restrict the review to them; paths alone (no range) mean the working tree restricted to those paths.

The raw input is: $ARGUMENTS

Spawn the **`crafter-reviewer`** agent with `model: opus`. The prompt carries only:

- the range exactly as given (or `working tree` when no range is given — including when only paths are), plus any paths as `restricted to <paths>`;
- one line naming the user's conversation language for the report prose;
- this line: "No greeting, preamble, closing summary, or hedging — only the sections your output format mandates, in order."

Pass nothing else — no description of the change, no conversation context, no project notes. The reviewer is meant to see the change cold.

Reproduce the reviewer's report unchanged — the summary line and the finding blocks as-is, never reformatted into tables, prose, or bullet lists. Do not add your own findings, re-rank, or fix anything.
