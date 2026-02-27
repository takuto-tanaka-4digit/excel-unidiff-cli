package euni

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type gitExecResult struct {
	stdout   string
	stderr   string
	exitCode int
}

func runGit(repo string, args ...string) (gitExecResult, error) {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err == nil {
		return gitExecResult{stdout: out.String(), stderr: errBuf.String(), exitCode: 0}, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return gitExecResult{stdout: out.String(), stderr: errBuf.String(), exitCode: exitErr.ExitCode()}, err
	}
	return gitExecResult{stdout: out.String(), stderr: errBuf.String(), exitCode: -1}, err
}

func gitShowTopLevel(repo string) (string, error) {
	res, err := runGit(repo, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("show-toplevel failed: %w (%s)", err, strings.TrimSpace(res.stderr))
	}
	return strings.TrimSpace(res.stdout), nil
}

func gitAbsoluteGitDir(repo string) (string, error) {
	res, err := runGit(repo, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return "", fmt.Errorf("absolute-git-dir failed: %w (%s)", err, strings.TrimSpace(res.stderr))
	}
	return strings.TrimSpace(res.stdout), nil
}

func gitListTracked(repo string) ([]string, error) {
	res, err := runGit(repo, "ls-files", "-z")
	if err != nil {
		return nil, fmt.Errorf("ls-files tracked failed: %w (%s)", err, strings.TrimSpace(res.stderr))
	}
	return splitNullSeparated(res.stdout), nil
}

func gitListUntracked(repo string) ([]string, error) {
	res, err := runGit(repo, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, fmt.Errorf("ls-files untracked failed: %w (%s)", err, strings.TrimSpace(res.stderr))
	}
	return splitNullSeparated(res.stdout), nil
}

func gitConfigGetBool(repo, key string) (bool, error) {
	res, err := runGit(repo, "config", "--local", "--type=bool", "--get", key)
	if err != nil {
		if res.exitCode == 1 {
			return false, nil
		}
		return false, fmt.Errorf("git config get failed for %s: %w (%s)", key, err, strings.TrimSpace(res.stderr))
	}

	v := strings.TrimSpace(strings.ToLower(res.stdout))
	switch v {
	case "true", "yes", "on", "1":
		return true, nil
	case "false", "no", "off", "0", "":
		return false, nil
	default:
		parsed, parseErr := strconv.ParseBool(v)
		if parseErr != nil {
			return false, fmt.Errorf("invalid bool for %s: %s", key, v)
		}
		return parsed, nil
	}
}

func gitConfigSetBool(repo, key string, value bool) error {
	res, err := runGit(repo, "config", "--local", key, strconv.FormatBool(value))
	if err != nil {
		return fmt.Errorf("git config set failed for %s=%t: %w (%s)", key, value, err, strings.TrimSpace(res.stderr))
	}
	return nil
}

type submoduleEntry struct {
	Path          string
	Uninitialized bool
}

func gitSubmoduleStatus(repo string) ([]submoduleEntry, error) {
	res, err := runGit(repo, "submodule", "status", "--recursive")
	if err != nil {
		// Repo without submodules may still return 0; non-zero is treated as execution failure.
		return nil, fmt.Errorf("submodule status failed: %w (%s)", err, strings.TrimSpace(res.stderr))
	}
	lines := strings.Split(strings.ReplaceAll(res.stdout, "\r\n", "\n"), "\n")
	entries := make([]submoduleEntry, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		trimmed := strings.TrimRight(line, "\r")
		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}
		prefix := byte(' ')
		if len(trimmed) > 0 {
			prefix = trimmed[0]
		}
		entries = append(entries, submoduleEntry{
			Path:          fields[1],
			Uninitialized: prefix == '-',
		})
	}
	return entries, nil
}

func splitNullSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "\x00")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

type deletedTrackedPathsForRepair struct {
	StagedPaths   []string
	WorktreePaths []string
}

type deletedTrackedPathsRepairOutcome struct {
	RestoredStagedPaths   []string
	RestoredWorktreePaths []string
	SkippedStagedPaths    []string
}

func (d deletedTrackedPathsForRepair) TotalCount() int {
	seen := make(map[string]struct{}, len(d.StagedPaths)+len(d.WorktreePaths))
	for _, p := range d.StagedPaths {
		seen[p] = struct{}{}
	}
	for _, p := range d.WorktreePaths {
		seen[p] = struct{}{}
	}
	return len(seen)
}

func (d deletedTrackedPathsRepairOutcome) RestoredCount() int {
	return len(d.RestoredStagedPaths) + len(d.RestoredWorktreePaths)
}

func gitListDeletedTrackedPathsForRepair(repo string) (deletedTrackedPathsForRepair, error) {
	collect := func(staged bool) ([]string, error) {
		args := []string{"-c", "core.precomposeunicode=false", "diff"}
		if staged {
			args = append(args, "--cached")
		}
		args = append(args, "--name-only", "--diff-filter=D", "-z")

		res, err := runGit(repo, args...)
		if err != nil {
			return nil, fmt.Errorf("git diff deleted paths failed: %w (%s)", err, strings.TrimSpace(res.stderr))
		}
		return splitNullSeparated(res.stdout), nil
	}

	stagedPaths, err := collect(true)
	if err != nil {
		return deletedTrackedPathsForRepair{}, err
	}
	worktreePaths, err := collect(false)
	if err != nil {
		return deletedTrackedPathsForRepair{}, err
	}

	stagedSet := make(map[string]struct{}, len(stagedPaths))
	for _, p := range stagedPaths {
		stagedSet[p] = struct{}{}
	}

	worktreeSet := make(map[string]struct{}, len(worktreePaths))
	for _, p := range worktreePaths {
		if _, ok := stagedSet[p]; ok {
			continue
		}
		worktreeSet[p] = struct{}{}
	}

	stagedOut := make([]string, 0, len(stagedSet))
	for p := range stagedSet {
		stagedOut = append(stagedOut, p)
	}
	worktreeOut := make([]string, 0, len(worktreeSet))
	for p := range worktreeSet {
		worktreeOut = append(worktreeOut, p)
	}
	sort.Strings(stagedOut)
	sort.Strings(worktreeOut)

	return deletedTrackedPathsForRepair{
		StagedPaths:   stagedOut,
		WorktreePaths: worktreeOut,
	}, nil
}

func splitRestorableAndConflictingStagedDeletedPaths(repo string, stagedPaths []string) ([]string, []string, error) {
	restorable := make([]string, 0, len(stagedPaths))
	conflicting := make([]string, 0)
	for _, p := range stagedPaths {
		abs := filepath.Join(repo, filepath.FromSlash(p))
		_, err := os.Lstat(abs)
		if err == nil {
			conflicting = append(conflicting, p)
			continue
		}
		if errors.Is(err, os.ErrNotExist) {
			restorable = append(restorable, p)
			continue
		}
		return nil, nil, fmt.Errorf("stat staged deleted path failed for %s: %w", p, err)
	}
	return restorable, conflicting, nil
}

func gitRestoreDeletedTrackedPathsForRepair(repo string, plan deletedTrackedPathsForRepair) (deletedTrackedPathsRepairOutcome, error) {
	if len(plan.StagedPaths) == 0 && len(plan.WorktreePaths) == 0 {
		return deletedTrackedPathsRepairOutcome{}, nil
	}

	runChunkedRestore := func(paths []string, argsPrefix []string, errMsg string) error {
		if len(paths) == 0 {
			return nil
		}
		const chunkSize = 200
		for i := 0; i < len(paths); i += chunkSize {
			end := i + chunkSize
			if end > len(paths) {
				end = len(paths)
			}
			args := append([]string{}, argsPrefix...)
			args = append(args, paths[i:end]...)
			res, err := runGit(repo, args...)
			if err != nil {
				return fmt.Errorf("%s: %w (%s)", errMsg, err, strings.TrimSpace(res.stderr))
			}
		}
		return nil
	}

	restorableStagedPaths, skippedStagedPaths, err := splitRestorableAndConflictingStagedDeletedPaths(repo, plan.StagedPaths)
	if err != nil {
		return deletedTrackedPathsRepairOutcome{}, err
	}
	if err := runChunkedRestore(
		restorableStagedPaths,
		[]string{"-c", "core.precomposeunicode=false", "restore", "--staged", "--worktree", "--"},
		"git restore staged deleted paths failed",
	); err != nil {
		return deletedTrackedPathsRepairOutcome{}, err
	}
	if err := runChunkedRestore(
		plan.WorktreePaths,
		[]string{"-c", "core.precomposeunicode=false", "restore", "--worktree", "--"},
		"git restore worktree deleted paths failed",
	); err != nil {
		return deletedTrackedPathsRepairOutcome{}, err
	}
	return deletedTrackedPathsRepairOutcome{
		RestoredStagedPaths:   restorableStagedPaths,
		RestoredWorktreePaths: plan.WorktreePaths,
		SkippedStagedPaths:    skippedStagedPaths,
	}, nil
}
