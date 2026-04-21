# Task: Skills-first runtime portability (VS Code, Copilot CLI, OpenCode)

## Problem

This task is still open, but its original assumptions are now outdated.

Crafter has already moved to a skills-first baseline. Compatibility command wrappers were removed, so portability work now needs to focus on runtime adapters, discovery contracts, and model-selection behavior across VS Code, Copilot CLI, and OpenCode.

## Proposed Approach

Keep one canonical skill source (`skills/`) and generate/runtime-install target-specific artifacts from it. Do not reintroduce wrappers as source-of-truth or permanent compatibility layer.

Add explicit runtime contracts for:
- skill discovery/invocation naming
- agent model fallback behavior vs orchestrator override behavior
- install output layout per runtime target

## Scope

### In scope

- Define canonical Crafter skills format and naming.
- Add runtime adapters for:
  - VS Code / Copilot CLI
  - OpenCode
- Update installer to produce runtime-specific outputs from canonical skills.
- Add install and smoke coverage for all supported runtime targets.
- Add migration notes for users coming from removed compatibility wrappers.

### Out of scope

- Redesigning Crafter workflows (`do/debug/map-project/status`) behavior.
- Major agent role changes unrelated to runtime packaging.

## Current Baseline (already completed)

- [x] Skills-first canonical source is active in repo (`skills/`).
- [x] Compatibility wrappers were removed from source and installer deployment.
- [x] Installer still cleans stale legacy command paths during upgrades.

## Migration Plan

- [ ] **Step 1: Runtime compatibility contract**
  - Document expected invocation, discovery path, and file layout for each target runtime.
  - Freeze naming policy (`crafter-*`) and define whether aliases are supported per runtime.
  - Document model behavior contract: orchestrator `model` override vs agent fallback model.

- [ ] **Step 2: Runtime adapter profiles**
  - Define deterministic per-runtime adapter config (frontmatter transforms, paths, naming).
  - Keep adapter logic isolated so runtime drift is localized.

- [ ] **Step 3: Build/transform pipeline**
  - Implement deterministic transform from canonical skills to each runtime output format.
  - Ensure frontmatter and path references are rewritten per runtime profile.

- [ ] **Step 4: Installer integration**
  - Update installer to install generated runtime artifacts by selected runtime.
  - Add stale-artifact cleanup logic to avoid mixed old/new layouts.

- [ ] **Step 5: Test matrix**
  - Extend tests to verify:
    - install output structure per runtime
    - command/skill discoverability
    - at least one end-to-end smoke path per runtime
    - model selection behavior for direct agent invocation vs orchestrated invocation

- [ ] **Step 6: Documentation rollout**
  - Update README and architecture docs to explain skills-first source + runtime adapters.
  - Add migration notes for existing users (global/local installs), including removed wrapper forms.

## Acceptance Criteria

- A single canonical skills source produces runtime-specific artifacts for all target runtimes.
- Fresh install works in VS Code, Copilot CLI, and OpenCode without manual edits.
- Runtime model behavior is documented and predictable:
  - orchestrator-passed `model` wins
  - direct invocation uses agent fallback model
- Test suite covers packaging + discoverability for all target runtimes.

## Risks and Mitigations

- **Risk:** Runtime-specific discovery rules drift over time.  
  **Mitigation:** Keep adapter logic isolated per runtime and covered by explicit tests.

- **Risk:** Breaking existing users after wrapper removal.  
  **Mitigation:** Provide explicit migration notes and optional runtime aliases where needed.

- **Risk:** Duplicate sources diverge.  
  **Mitigation:** Enforce one canonical source and generated outputs only.
