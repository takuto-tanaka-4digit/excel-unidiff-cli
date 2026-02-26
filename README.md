# excel-unidiff-cli (`euni`)

[English](./README.md) | [日本語](./README.ja.md)

A cross-platform CLI that prevents Unicode normalization phantom diffs in Git by detecting and fixing NFC/NFD-related drift in a deterministic, CI-friendly way.

## Problem

Mixed macOS/Linux/Windows teams often hit Unicode normalization mismatches:

- `git status` stays dirty after discard.
- Filenames look unchanged but keep appearing in diffs.
- CI runners disagree with local environments.

This is common with spreadsheet-heavy repos, multilingual filenames, and macOS-focused workflows.

## Why This Matters in Teams and CI

Without a shared policy, Unicode drift creates hidden operational cost:

- Review noise: PRs include meaningless filename churn.
- CI instability: checks fail on one runner but pass on another.
- Automation waste: bots and scripts keep retrying "dirty" states.
- Slow incident response: root cause is hard to isolate.

`euni` gives one policy, one contract, and repeatable behavior across runners.

## What `euni` Does

- Detects Git config drift related to Unicode-safe behavior.
- Fixes drift via `git config --local` only.
- Scans paths for NFC/NFD risks, combining marks, and collisions.
- Separates findings from execution errors with stable exit semantics.
- Produces machine-readable JSON output for CI gates.

## Commands

- `euni check`
  - Non-destructive validation.
  - Drift + Unicode health checks on tracked files only.
- `euni apply`
  - Drift remediation only (`git config --local`).
  - Supports `--dry-run`.
- `euni doctor`
  - Integrated diagnosis (`check` + `scan`) with cause classification.
- `euni scan`
  - Unicode path analysis for tracked + untracked files.
- `euni init-policy`
  - Creates `.euni.yml` policy template.
- `euni version`
  - Prints version and commit.

## Scope and Impact Matrix

| Command | `--recursive=false` (default) | `--recursive=true` | Writes data |
| --- | --- | --- | --- |
| `check` | root repo only | root + submodules + nested repos | No |
| `apply` | root repo only | root + submodules + nested repos | Yes (`git config --local` only) |
| `doctor` | root repo only | root + submodules + nested repos | No |
| `scan` | root repo only | root + submodules + nested repos | No |
| `init-policy` | root repo only | N/A | Yes (`.euni.yml`) |
| `version` | N/A | N/A | No |

Note: subtree content is always included in root file scanning.

## Safety Guarantees and Non-Goals

`euni` is designed for safe operations by default.

- Safe default scope: `--recursive=false`.
- Explicit write path: only `apply` and `init-policy` write.
- `apply --dry-run` shows planned changes before write.
- No global/system Git config changes.
- No destructive Git commands:
  - no `git reset --hard`
  - no `git clean -fd`
- No automatic filename rename in v1.

## Quick Start

1. Create policy template.

```bash
euni init-policy --repo .
```

2. Check current state.

```bash
euni check --repo . --recursive --policy ./.euni.yml
```

3. Preview changes, then apply.

```bash
euni apply --repo . --recursive --policy ./.euni.yml --dry-run
euni apply --repo . --recursive --policy ./.euni.yml
```

4. Diagnose hard cases.

```bash
euni doctor --repo . --recursive --policy ./.euni.yml
```

## Install with Homebrew (Same Repo Tap)

Use this repository itself as the tap source.

```bash
brew tap --custom-remote takuto-tanaka-4digit/excel-unidiff-cli https://github.com/takuto-tanaka-4digit/excel-unidiff-cli
brew install takuto-tanaka-4digit/excel-unidiff-cli/euni
```

Upgrade later:

```bash
brew update
brew upgrade takuto-tanaka-4digit/excel-unidiff-cli/euni
```

## CI Integration Essentials

Use JSON mode and validate the report contract.

```bash
euni check --repo . --recursive --policy ./.euni.yml --non-interactive --format json > euni-report.json
```

Recommended CI gates:

- Validate `euni-report.json` against `schema/euni-report.schema.json`.
- Ensure process exit code matches report `status/exitCode`.
- Fail on findings (`1`) and execution errors (`2`).
- Always upload `euni-report.json` as an artifact.

JSON contract rule:

- `stdout`: one JSON object only.
- `stderr`: progress, warnings, and error details.

## Exit Code Contract

- `0`: no problems.
- `1`: findings detected (action needed).
- `2`: execution error (repo/policy/git/runtime failure).

If any execution error exists, `2` has precedence.

## Policy File

`euni` uses `.euni.yml` as the source of truth for expected behavior at root, submodule, and nested-repo levels.

## Who Should Use This

- Teams with macOS developers and Linux/Windows CI.
- Repos with multilingual filenames (Japanese, symbols, spaces).
- Maintainers who want deterministic, auditable Unicode hygiene.
