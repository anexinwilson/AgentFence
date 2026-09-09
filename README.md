# AgentFence

Make AI coding agents actually do the work.

AI coding agents are fast and capable, but they can take shortcuts: use fake or hardcoded values, leave incomplete implementations, silently replace difficult requirements with easier ones, ignore instructions, skip verification, or declare a task finished before the work is actually done.

You ask for something specific. The agent does something close enough, says "Done," and moves on. Now you're stuck re-prompting it, checking its work, and correcting the same mistakes until it finally does what you asked.

AgentFence adds a policy and verification layer between the agent and completion. It enforces your project rules, observes what the agent actually does, verifies the resulting work, and blocks completion when required work has not been proven. The agent decides how to implement. AgentFence makes sure it actually earns the right to say it's done.

---

## Supported Agents

* **Google Antigravity**: Fully supported out of the box via native `.agents/hooks.json` lifecycle hooks (`PreToolUse` and `Stop`).
* **Claude Code & others**: Architecture is agent-agnostic; native hook adapters for Claude Code and other agent environments are currently on the roadmap.

---

## Tech Stack & Architecture

AgentFence is built from the ground up for speed, safety, and zero-friction developer experience:

* **Single Static Go Binary**: Compiles to a self-contained executable with zero runtime dependencies (`agentfence.exe` on Windows, `agentfence` on Linux/macOS). Runs in sub-10ms without Node.js, Python virtualenvs, Docker daemons, or background services.
* **Native Agent Lifecycle Hooks**: Integrates directly into agent lifecycle hooks (`PreToolUse` and `Stop` via `.agents/hooks.json`). Blocks dangerous commands like `git push` or `.env` overwrites before execution, and intercepts task completion to enforce verification.
* **AST Code Integrity Analysis**: Uses Go standard library parsers (`go/parser`, `go/ast`) and lexical token scanners for TypeScript, JavaScript, and Python. Scans code structure to catch empty functions, bare `pass`, `...`, and synthetic return literals.
* **Anti-Fallback Guard**: Analyzes source code exception handling to detect and reject empty `catch` and `except` blocks that swallow errors and hide broken integrations.
* **Cryptographic Git Tree Attestation**: Computes Git Merkle tree hashes (`git write-tree`) and locks them in `.agentfence/attestation.json` to prevent race conditions where an agent mutates code after verification passes.
* **Polyglot Stack Auto-Discovery**: Automatically inspects repository files to detect Go, TypeScript/JavaScript, Python, and Rust, configuring matching test commands and policies with a single command (`agentfence init`).

---

## What AgentFence Prevents

* **Skipping required work**: Task exit is blocked until all checklist requirements have verified proof.
* **Fake or hardcoded values**: AST scanner flags synthetic returns that lack operational logic.
* **Hollow stubs and placeholders**: Rejects bare `pass`, `...`, and empty function bodies.
* **Silent substitution**: Catches when an agent quietly implements an easier shortcut instead of the requested task.
* **Skipping tests**: Test suites must execute and pass with exit code 0 before a session can conclude.
* **Regressions**: Automatically runs test suites to catch existing features broken by new edits.
* **Scope creep**: Flags edits to unrelated files and modules outside the task boundary.
* **Dangerous operations**: Blocks `git push`, forced resets, destructive shell commands, and `.env` edits.

---

## The Reliability Gap in Practice

Consider a typical scenario: you ask an agent to implement Stripe webhook signature verification and persist confirmed orders to the database.

The agent encounters setup friction configuring the secret key in the test environment. Rather than solving the configuration, it writes:

```typescript
export async function handleWebhook(req: Request) {
  // TODO: Add signature verification in production
  const event = JSON.parse(req.body);
  return { received: true };
}
```

The agent reports: "Stripe webhook integration completed with full database persistence."

The syntax is valid, the file exists, and the agent claims victory. But signature verification was skipped with a TODO comment, and the database call was never written.

This represents the reliability gap in AI coding agents:

```text
What you asked for
        |
        v
What the agent actually did
        |
        v
What the agent claims it did
```

Those three things are frequently completely different.

---

## Core Failure Modes Handled

1. **Fabrication**: Claiming an implementation exists when it is an empty stub, hardcoded return, or mocked placeholder with zero real logic.
2. **Substitution**: Silently replacing difficult requirements with easier alternatives (such as replacing a database write with an in-memory array).
3. **Deviation**: Editing unrelated files, reformatting untouched modules, or wandering outside the requested task.
4. **Regression**: Resolving the immediate issue while breaking existing tests, APIs, or build configurations.
5. **Premature Completion**: Exiting without executing test suites, validating builds, or proving requirements were met.

---

## Problems Solved

| Agent Behavior | AgentFence Enforcement |
|:---|:---|
| Exits without running tests | Stop hook blocks turn exit until test suite exits with code 0 |
| Inserts hardcoded return values | AST validators flag synthetic literals without operational logic |
| Uses empty stubs or placeholder blocks | Code integrity scanner detects pass, ..., and empty function blocks |
| Substitutes requested architecture | Requirement evidence verifies concrete implementation proof |
| Mocks external services instead of implementing them | Action verification inspects actual network and database logic |
| Edits unrelated files or drifts from scope | Scope validation detects file mutations outside project boundaries |
| Introduces regressions into existing code | Automatic test runner verifies test suite before turn completion |
| Repeats previously corrected errors | Merkle tree state tracking verifies working tree integrity |
| Ignores repository conventions | Architecture validators check directory structure and rules |
| Runs destructive shell commands | PreToolUse hook blocks rm -rf, raw script downloads, and disk wipes |
| Unintended git commits or remote pushes | Git policy blocks git push, forced resets, and unreviewed commits |
| Modifies protected files or credentials | Path policy denies access to .env files, certificates, and private keys |
| Reports false test passes | Test runner executes independently in an isolated subprocess |
| Ends turn before requirements are met | Stop hook gate blocks completion until checklist proof is verified |

---

## How It Works: Native Agent Hooks

Prompt rules fail to prevent agent shortcuts because language models are probabilistic. Under multi-turn context pressure, agents regularly violate prompt suggestions.

AgentFence enforces rules by hooking directly into the agent lifecycle (via `.agents/hooks.json` in Google Antigravity, with similar hook pipelines for Claude Code):

```text
               Agent initiates action / finish
                             |
                             v
                [PreToolUse Hook Intercepts]
                 Blocks git push, .env edits
                             |
                             v
                  [Stop Hook Intercepts]
               Runs tests, AST checks, builds
                             |
             +---------------+---------------+
             |                               |
             v                               v
      Failed: Blocked                 Passed: Clean
    Agent forced to fix             Session completes
```

AgentFence operates through two lifecycle hooks:
* **PreToolUse Hook**: Runs before tool calls execute. Returns `{"decision": "deny"}` to block destructive commands, unauthorized file writes, or accidental `git push`.
* **Stop Hook**: Runs when the agent attempts to finish a turn. Returns `{"decision": "continue"}` with compiler and test output if tests fail or stubs are detected, forcing the agent to keep working until all requirements pass.

---

## Architecture & Enforcement Lifecycle

AgentFence operates as a two-stage process gatekeeper:

### Stage 1: PreToolUse Security Gatekeeper (`agentfence hook pre-tool`)
Every tool call (`run_command`, `write_to_file`, `replace_file_content`, `edit_file`, `delete_file`) is intercepted over stdin/stdout JSON IPC before execution:
* **Git Policy**: Evaluates command tokens. Safe read-only operations (`git status`, `git diff`, `git log`) pass in sub-millisecond time. Destructive operations (`push`, `commit`, `reset --hard`, `clean -f`, `rebase`) are rejected with exit code 1.
* **File Policy**: Blocks unauthorized writes or overwrites to `.env*`, `secrets/**`, credentials, and private keys.
* **Command Policy**: Intercepts destructive patterns (`rm -rf *`, `rm -rf /`, `curl | bash`, `wget | sh`) before the shell executes.

### Stage 2: Stop Verification Engine (`agentfence hook stop`)
When an agent attempts to finish a turn or declare victory, the Stop hook fires before control returns to the developer:
1. **Automated Test Suites**: Executes project test runners (`go test ./...`, `npm test`, `pytest`, `cargo test`) under strict subprocess timeouts.
2. **AST Code Integrity Scanner**: Parses source code ASTs (pure Go standard library parser, JS/TS, Python) to flag empty function bodies, hollow I/O functions returning hardcoded literals without external calls, and generic Dockerfile stubs.
3. **Anti-Fallback Detector**: Scans catch and except blocks to ensure errors are not silently swallowed.
4. **Cryptographic Tree Attestation Lock**: Computes a Git Merkle tree hash (`git write-tree`) and locks it in `.agentfence/attestation.json`. If an agent edits code after testing, the tree hash mismatches and the turn is rejected.
5. **Automated Evidence Compilation**: Compiles the verified audit trail and writes `evidence.md` into the session directory.

If any check fails, AgentFence returns `{"decision": "continue", "reason": "..."}`. The agent is blocked from stopping and supplied with exact error logs until the code is fixed.

---

## Installation & Setup

### One-Click Global Installation

#### Windows (PowerShell):
```powershell
.\install.ps1
```

#### Linux & macOS (Bash):
```bash
./install.sh
```

The installer compiles the binary, places it in `~/.agentfence/bin/agentfence`, and ensures it is added to your user `PATH`.

---

## Developer Experience (DX) & Commands

### 1. Initialize Any Project (`agentfence init`)
Navigate to any repository and run:
```powershell
agentfence init
```
This single command automatically:
* Detects your project stack (Go, TypeScript, Python, Rust, etc.).
* Creates `agentfence.yaml` with pre-configured test and verification policies.
* Creates `.agents/hooks.json` wiring native `PreToolUse` and `Stop` hooks.
* Creates `.agents/rules/agentfence.md` instructing the AI agent on architectural rules.

### 2. The Evidence Command (`agentfence evidence`)
The evidence command gathers verifiable proof of completed work across tests, requirements, and file mutations. It automatically saves an audit trail to `.agents/evidence.md`.

* **Terminal Output**:
  ```powershell
  agentfence evidence
  ```
  Displays test run durations, status reports, and requirement verification tables.

* **Markdown Output (For PRs / Issues)**:
  ```powershell
  agentfence evidence --markdown
  ```
  Outputs clean Markdown suitable for pasting directly into pull request descriptions.

* **Machine-Readable JSON**:
  ```powershell
  agentfence evidence --json
  ```
  Dumps structured JSON for CI/CD pipelines.

If any check fails or is unverified, `agentfence evidence` exits with code 1.

### 3. Pause Protection on Demand (`agentfence stop`)
When you want to experiment freely or make manual adjustments without hook interference:
```powershell
agentfence stop
```
Sets `"enabled": false` in `.agents/hooks.json`. Guardrails and hooks are immediately paused.

### 4. Resume Protection (`agentfence start`)
When you are ready to put guardrails back online:
```powershell
agentfence start
```
Sets `"enabled": true` in `.agents/hooks.json`. Protection is restored.

### 5. System Health Check (`agentfence doctor`)
Inspects your toolchain and project integration status:
```powershell
agentfence doctor
```

### 6. Manual Verification (`agentfence verify`)
Runs the full verification suite on demand and locks the tree attestation:
```powershell
agentfence verify
```

---

## Example Evidence Report

When verification passes, `agentfence evidence` generates an objective audit trail:

```text
Requirements:
  [PASS] [R-01] Payment webhook signature verification (pkg/payments/webhook.go)
  [PASS] [R-02] Database order persistence and rollback test (db/migrations/003_orders.sql)
  [PASS] [R-03] Integration test suite green with clean exit 0 (tests/integration_test.go)
  [PASS] [R-04] Zero hollow stubs or synthetic mock fallbacks (pkg/payments/service.go)

Verification:
  [PASS] Tests: All tests passed in 1.42s (go test ./... / npm test)
  [PASS] Placeholders: No placeholder or stub implementations detected
  [PASS] Fallbacks: No swallowed exceptions or silent fallbacks detected
  [PASS] Integrity: Zero hollow stubs or generic container templates detected

Git Policy:
  [PASS] No remote push: Direct git push intercepted and prevented by AgentFence

Evidence Summary:
  Requirements: 4/4 satisfied
  Tests: passed
  Placeholders: passed
  Fallbacks: passed

VERDICT: VERIFIED
```

---

## License

Released under the [MIT License](LICENSE).
