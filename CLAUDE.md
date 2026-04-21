# CLAUDE.md

Repository policy for Claude-based agents.

## Primary rule: no model signatures

Never add any AI/model attribution or signature to:

1. commit messages
2. pull request titles or bodies
3. release titles or release notes

This includes (non-exhaustive):

- `Co-authored-by: ...`
- `Signed-off-by: ...`
- `Generated-by: ...`
- `🤖 Generated with ...`
- any mention like "created by Claude/Copilot/GPT/Gemini"

## Required behavior

1. Keep commits, PRs, and releases in a neutral human project voice.
2. If a template or tool auto-inserts an attribution/signature, remove it before final output.
3. If unsure whether text is a signature, treat it as a signature and remove it.
