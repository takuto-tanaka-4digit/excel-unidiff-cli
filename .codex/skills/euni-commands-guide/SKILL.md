---
name: euni-commands-guide
description: English runbook for euni command usage in local development and CI, including exit-code and UG-code troubleshooting.
---

# euni Commands (English)

Operational playbook for `excel-unidiff-cli` (`euni`).
Goal: decide quickly which command to run, in what order, and how to interpret results.

## Command Map

- `euni init-policy`
  - Create `.euni.yml` template in `--repo`.
- `euni check`
  - Non-destructive check for policy drift + Unicode findings on tracked files.
- `euni apply`
  - Apply policy drift fixes via `git config --local` only.
  - Use `--dry-run` before writing.
- `euni doctor`
  - Combined diagnosis (`check` + scan-like Unicode analysis).
- `euni scan`
  - Unicode path analysis on tracked + untracked files (no policy input).
- `euni version`
  - Print `euni <version> (<commit>)`.

## Core Rules

- Subcommand required; no standalone `help` command.
- `--recursive` default is `false` (root repo only).
- Write operations are only `apply` and `init-policy`.
- `scan` does not accept `--policy`.
- `version` accepts no options.
- `--format` is `text` or `json` only.

## Recommended Flows

### 1) First-time setup

```bash
euni init-policy --repo .
euni check --repo . --recursive --policy ./.euni.yml
```

### 2) Drift remediation loop

```bash
euni apply --repo . --recursive --policy ./.euni.yml --dry-run
euni apply --repo . --recursive --policy ./.euni.yml
euni check --repo . --recursive --policy ./.euni.yml
```

### 3) Hard-case diagnosis

```bash
euni doctor --repo . --recursive --policy ./.euni.yml
euni scan --repo . --recursive --format json
```

## CI Recipe (Deterministic Gate)

```bash
euni check \
  --repo . \
  --recursive \
  --policy ./.euni.yml \
  --non-interactive \
  --format json > euni-report.json
```

Gate expectations:

- Validate against `schema/euni-report.schema.json`.
- Ensure process exit code matches `report.exitCode`.
- Treat `exit 1` as findings (operational failure), `exit 2` as execution/runtime failure.
- Upload `euni-report.json` as artifact.

JSON contract:

- `stdout`: exactly one JSON object.
- `stderr`: logs/progress/error details.

## Exit Codes

- `0`: no findings, no errors.
- `1`: findings present.
- `2`: execution error.

Precedence: if any execution error exists, exit is `2`.

## UG Code Triage (Quick)

Common findings (`exit 1`):

- `UG004`: config drift detected.
- `UG005`: NFC collision detected.
- `UG011`: combining mark in path.
- `UG012`: non-standard FS entry (symlink/reparse/mount).
- `UG013`: ambiguous policy path key (case-only collision).

Common execution errors (`exit 2`):

- `UG001`: invalid/inaccessible `--repo`.
- `UG002`: git command/runtime failure.
- `UG003`: policy load failure.
- `UG006`: submodule uninitialized/inaccessible.
- `UG007`: invalid policy structure/unsupported keys.
- `UG008`: `.euni.yml` exists without `--force`.
- `UG009`: unsupported command/option usage.
- `UG010`: gitdir/top-level outside `--repo` boundary.

## Local Source Run Note

When running via `go run`, non-zero app exits may appear as process exit `1` (Go tool wrapper behavior).
For strict `0/1/2` contract checks, build and run the binary directly:

```bash
go build -o ./bin/euni ./cmd/euni
./bin/euni check --repo . --policy ./.euni.yml
echo $?
```
