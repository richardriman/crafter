# Task: Skills-first runtime portability (VS Code, Copilot CLI, OpenCode)

## Problem

Crafter currently treats `commands/` as the primary source, which makes runtime behavior inconsistent as runtimes evolve toward skill discovery. We need a skills-first model so the same workflow works reliably across VS Code, Copilot CLI, and OpenCode.

## Proposed Approach

Adopt a single canonical skill source and generate runtime-specific artifacts during install/build. Keep compatibility wrappers during migration to avoid breaking existing users.

## Scope

### In scope

- Define canonical Crafter skills format and naming.
- Add runtime adapters for:
  - VS Code / Copilot CLI
  - OpenCode
- Update installer to produce runtime-specific outputs from canonical skills.
- Preserve backward compatibility for existing command entry points during transition.
- Add install and smoke coverage for all supported runtime targets.

### Out of scope

- Redesigning Crafter workflows (`do/debug/map-project/status`) behavior.
- Major agent role changes unrelated to runtime packaging.

## Migration Plan

- [ ] **Step 1: Runtime compatibility contract**
  - Document expected invocation, discovery path, and file layout for each target runtime.
  - Freeze command/skill naming policy (`crafter-*`) and alias policy.

- [ ] **Step 2: Canonical source layout**
  - Introduce canonical skill source directory in repo (skills-first source of truth).
  - Keep current `commands/` files untouched initially.

- [ ] **Step 3: Build/transform pipeline**
  - Implement deterministic transform from canonical skills to each runtime output format.
  - Ensure frontmatter and path references are rewritten per runtime profile.

- [ ] **Step 4: Installer integration**
  - Update installer to install generated runtime artifacts by selected runtime.
  - Add stale-artifact cleanup logic to avoid mixed old/new layouts.

- [ ] **Step 5: Backward compatibility layer**
  - Keep legacy command entry points as wrappers/aliases during migration window.
  - Ensure old invocation forms still route to the same behavior.

- [ ] **Step 6: Test matrix**
  - Extend tests to verify:
    - install output structure per runtime
    - command/skill discoverability
    - at least one end-to-end smoke path per runtime

- [ ] **Step 7: Documentation rollout**
  - Update README and architecture docs to explain skills-first source + runtime adapters.
  - Add migration notes for existing users (global/local installs).

## Acceptance Criteria

- A single canonical skills source produces runtime-specific artifacts for all target runtimes.
- Fresh install works in VS Code, Copilot CLI, and OpenCode without manual edits.
- Existing users keep working invocation patterns during migration.
- Test suite covers packaging + discoverability for all target runtimes.

## Risks and Mitigations

- **Risk:** Runtime-specific discovery rules drift over time.  
  **Mitigation:** Keep adapter logic isolated per runtime and covered by explicit tests.

- **Risk:** Breaking existing users during transition.  
  **Mitigation:** Keep compatibility wrappers and phased deprecation.

- **Risk:** Duplicate sources diverge.  
  **Mitigation:** Enforce one canonical source and generated outputs only.

