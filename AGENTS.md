# AGENTS.md

Global repository policy for all coding agents.

## Highest-priority content policy

Do **not** sign outputs with model identity in commits, pull requests, or releases.

Forbidden in commit messages, PR text, and release text:

- any `Co-authored-by` trailer
- any `Signed-off-by` trailer
- any model/vendor attribution ("Claude", "Copilot", "GPT", "Gemini", etc.)
- any "AI-generated" signature line or footer

## Enforcement rules

1. Before finalizing commit text, PR text, or release notes, strip model signatures/attributions.
2. Use concise, project-focused wording only.
3. If an automation path conflicts with this policy, prefer this policy for generated text content.
