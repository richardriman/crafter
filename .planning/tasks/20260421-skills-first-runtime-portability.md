# Task: Skills-first runtime portability (VS Code, Copilot CLI, OpenCode)

## Metadata
- **Date:** 2026-04-21
- **Branch:** main
- **Status:** active
- **Scope:** Large

## Request
Make Crafter runtime-portable across VS Code, Copilot CLI, and OpenCode with a skills-first source model and predictable install/discovery behavior.

## Plan
- [ ] **Step 1: Runtime compatibility contract** — document invocation, discovery paths, and file layout for each runtime; freeze naming policy (`crafter-*`) and alias policy; define model contract (orchestrator override vs agent fallback).
- [ ] **Step 2: Runtime adapter profiles** — define deterministic per-runtime adapter config (frontmatter transforms, paths, naming) with isolated runtime-specific logic.
- [ ] **Step 3: Build/transform pipeline** — implement deterministic transform from canonical `skills/` source to runtime-specific outputs.
- [ ] **Step 4: Installer integration** — install runtime-specific generated artifacts and keep stale-artifact cleanup to prevent mixed layouts.
- [ ] **Step 5: Test matrix** — add install/discoverability/smoke tests per runtime, including model-selection behavior for direct agent invocation vs orchestrated invocation.
- [ ] **Step 6: Documentation rollout** — update README/ARCHITECTURE docs for runtime adapters and add migration notes for users coming from removed wrapper commands.

## Decisions
- **Decision:** Keep `skills/` as the only canonical workflow source. **Reason:** avoids divergence between duplicated command/skill sources.
- **Decision:** Do not reintroduce compatibility wrappers as permanent layer. **Reason:** wrappers caused duplication and inconsistent runtime behavior.
- **Decision:** Define explicit model-behavior contract per runtime. **Reason:** direct agent invocation and orchestrated Task invocation may resolve models differently.

## Outcome
Task remains active.

Current baseline already in place:
- Skills-first canonical source is active in repo.
- Compatibility command wrappers were removed from source and installer deployment.
- Installer still cleans stale legacy `commands/crafter` artifacts during upgrades.
