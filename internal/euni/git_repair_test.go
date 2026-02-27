package euni

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestRepairDeletedTrackedPathsRestoresStagedDeletion(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	mustGit(t, repo, "init")
	mustGit(t, repo, "config", "user.name", "euni-test")
	mustGit(t, repo, "config", "user.email", "euni-test@example.com")

	filePath := filepath.Join(repo, "sample.txt")
	if err := os.WriteFile(filePath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write sample file: %v", err)
	}
	mustGit(t, repo, "add", "sample.txt")
	mustGit(t, repo, "commit", "-m", "init")

	if err := os.Remove(filePath); err != nil {
		t.Fatalf("remove sample file: %v", err)
	}
	mustGit(t, repo, "add", "-u")

	plan, err := gitListDeletedTrackedPathsForRepair(repo)
	if err != nil {
		t.Fatalf("gitListDeletedTrackedPathsForRepair: %v", err)
	}
	if !slices.Equal(plan.StagedPaths, []string{"sample.txt"}) {
		t.Fatalf("staged paths = %v, want [sample.txt]", plan.StagedPaths)
	}
	if len(plan.WorktreePaths) != 0 {
		t.Fatalf("worktree paths = %v, want []", plan.WorktreePaths)
	}

	outcome, err := gitRestoreDeletedTrackedPathsForRepair(repo, plan)
	if err != nil {
		t.Fatalf("gitRestoreDeletedTrackedPathsForRepair: %v", err)
	}
	if outcome.RestoredCount() != 1 {
		t.Fatalf("restored count = %d, want 1", outcome.RestoredCount())
	}
	if len(outcome.SkippedStagedPaths) != 0 {
		t.Fatalf("skipped staged paths = %v, want []", outcome.SkippedStagedPaths)
	}

	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("restored file does not exist: %v", err)
	}

	res, err := runGit(repo, "status", "--porcelain")
	if err != nil {
		t.Fatalf("git status --porcelain failed: %v (%s)", err, strings.TrimSpace(res.stderr))
	}
	if strings.TrimSpace(res.stdout) != "" {
		t.Fatalf("repo not clean after repair: %q", res.stdout)
	}
}

func TestListDeletedTrackedPathsForRepairIncludesUnstagedAndStaged(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	mustGit(t, repo, "init")
	mustGit(t, repo, "config", "user.name", "euni-test")
	mustGit(t, repo, "config", "user.email", "euni-test@example.com")

	if err := os.WriteFile(filepath.Join(repo, "staged.txt"), []byte("staged\n"), 0o644); err != nil {
		t.Fatalf("write staged file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "unstaged.txt"), []byte("unstaged\n"), 0o644); err != nil {
		t.Fatalf("write unstaged file: %v", err)
	}
	mustGit(t, repo, "add", "staged.txt", "unstaged.txt")
	mustGit(t, repo, "commit", "-m", "init")

	if err := os.Remove(filepath.Join(repo, "staged.txt")); err != nil {
		t.Fatalf("remove staged file: %v", err)
	}
	if err := os.Remove(filepath.Join(repo, "unstaged.txt")); err != nil {
		t.Fatalf("remove unstaged file: %v", err)
	}
	mustGit(t, repo, "add", "staged.txt")

	plan, err := gitListDeletedTrackedPathsForRepair(repo)
	if err != nil {
		t.Fatalf("gitListDeletedTrackedPathsForRepair: %v", err)
	}
	if !slices.Equal(plan.StagedPaths, []string{"staged.txt"}) {
		t.Fatalf("staged paths = %v, want [staged.txt]", plan.StagedPaths)
	}
	if !slices.Equal(plan.WorktreePaths, []string{"unstaged.txt"}) {
		t.Fatalf("worktree paths = %v, want [unstaged.txt]", plan.WorktreePaths)
	}
	if plan.TotalCount() != 2 {
		t.Fatalf("total count = %d, want 2", plan.TotalCount())
	}
}

func TestRestoreDeletedTrackedPathsForRepairKeepsStagedModifyOnWorktreeDelete(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	mustGit(t, repo, "init")
	mustGit(t, repo, "config", "user.name", "euni-test")
	mustGit(t, repo, "config", "user.email", "euni-test@example.com")

	filePath := filepath.Join(repo, "sample.txt")
	if err := os.WriteFile(filePath, []byte("base\n"), 0o644); err != nil {
		t.Fatalf("write sample file: %v", err)
	}
	mustGit(t, repo, "add", "sample.txt")
	mustGit(t, repo, "commit", "-m", "init")

	if err := os.WriteFile(filePath, []byte("staged-change\n"), 0o644); err != nil {
		t.Fatalf("update sample file: %v", err)
	}
	mustGit(t, repo, "add", "sample.txt")
	if err := os.Remove(filePath); err != nil {
		t.Fatalf("remove sample file: %v", err)
	}

	plan, err := gitListDeletedTrackedPathsForRepair(repo)
	if err != nil {
		t.Fatalf("gitListDeletedTrackedPathsForRepair: %v", err)
	}
	if len(plan.StagedPaths) != 0 {
		t.Fatalf("staged paths = %v, want []", plan.StagedPaths)
	}
	if !slices.Equal(plan.WorktreePaths, []string{"sample.txt"}) {
		t.Fatalf("worktree paths = %v, want [sample.txt]", plan.WorktreePaths)
	}

	outcome, err := gitRestoreDeletedTrackedPathsForRepair(repo, plan)
	if err != nil {
		t.Fatalf("gitRestoreDeletedTrackedPathsForRepair: %v", err)
	}
	if outcome.RestoredCount() != 1 {
		t.Fatalf("restored count = %d, want 1", outcome.RestoredCount())
	}
	if len(outcome.SkippedStagedPaths) != 0 {
		t.Fatalf("skipped staged paths = %v, want []", outcome.SkippedStagedPaths)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if string(content) != "staged-change\n" {
		t.Fatalf("restored content = %q, want staged content", string(content))
	}

	res, err := runGit(repo, "status", "--porcelain")
	if err != nil {
		t.Fatalf("git status --porcelain failed: %v (%s)", err, strings.TrimSpace(res.stderr))
	}
	if strings.TrimSpace(res.stdout) != "M  sample.txt" {
		t.Fatalf("status after repair = %q, want %q", strings.TrimSpace(res.stdout), "M  sample.txt")
	}
}

func TestRestoreDeletedTrackedPathsForRepairChunking(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	mustGit(t, repo, "init")
	mustGit(t, repo, "config", "user.name", "euni-test")
	mustGit(t, repo, "config", "user.email", "euni-test@example.com")

	paths := make([]string, 0, 205)
	for i := 0; i < 205; i++ {
		name := fmt.Sprintf("f-%03d.txt", i)
		paths = append(paths, name)
		if err := os.WriteFile(filepath.Join(repo, name), []byte("x\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	mustGit(t, repo, "add", ".")
	mustGit(t, repo, "commit", "-m", "init")

	for _, p := range paths {
		if err := os.Remove(filepath.Join(repo, p)); err != nil {
			t.Fatalf("remove %s: %v", p, err)
		}
	}
	mustGit(t, repo, "add", "-u")

	plan, err := gitListDeletedTrackedPathsForRepair(repo)
	if err != nil {
		t.Fatalf("gitListDeletedTrackedPathsForRepair: %v", err)
	}
	if len(plan.StagedPaths) != 205 {
		t.Fatalf("staged deleted count = %d, want 205", len(plan.StagedPaths))
	}
	if len(plan.WorktreePaths) != 0 {
		t.Fatalf("worktree deleted count = %d, want 0", len(plan.WorktreePaths))
	}

	outcome, err := gitRestoreDeletedTrackedPathsForRepair(repo, plan)
	if err != nil {
		t.Fatalf("gitRestoreDeletedTrackedPathsForRepair: %v", err)
	}
	if outcome.RestoredCount() != 205 {
		t.Fatalf("restored count = %d, want 205", outcome.RestoredCount())
	}
	if len(outcome.SkippedStagedPaths) != 0 {
		t.Fatalf("skipped staged paths = %v, want []", outcome.SkippedStagedPaths)
	}

	res, err := runGit(repo, "status", "--porcelain")
	if err != nil {
		t.Fatalf("git status --porcelain failed: %v (%s)", err, strings.TrimSpace(res.stderr))
	}
	if strings.TrimSpace(res.stdout) != "" {
		t.Fatalf("repo not clean after chunked repair: %q", res.stdout)
	}
}

func TestRestoreDeletedTrackedPathsForRepairSkipsStagedDeleteWhenPathAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	mustGit(t, repo, "init")
	mustGit(t, repo, "config", "user.name", "euni-test")
	mustGit(t, repo, "config", "user.email", "euni-test@example.com")

	filePath := filepath.Join(repo, "sample.txt")
	if err := os.WriteFile(filePath, []byte("base\n"), 0o644); err != nil {
		t.Fatalf("write sample file: %v", err)
	}
	mustGit(t, repo, "add", "sample.txt")
	mustGit(t, repo, "commit", "-m", "init")

	if err := os.Remove(filePath); err != nil {
		t.Fatalf("remove sample file: %v", err)
	}
	mustGit(t, repo, "add", "-u")
	if err := os.WriteFile(filePath, []byte("untracked-recreated\n"), 0o644); err != nil {
		t.Fatalf("recreate sample file: %v", err)
	}

	plan, err := gitListDeletedTrackedPathsForRepair(repo)
	if err != nil {
		t.Fatalf("gitListDeletedTrackedPathsForRepair: %v", err)
	}
	if !slices.Equal(plan.StagedPaths, []string{"sample.txt"}) {
		t.Fatalf("staged paths = %v, want [sample.txt]", plan.StagedPaths)
	}

	outcome, err := gitRestoreDeletedTrackedPathsForRepair(repo, plan)
	if err != nil {
		t.Fatalf("gitRestoreDeletedTrackedPathsForRepair: %v", err)
	}
	if outcome.RestoredCount() != 0 {
		t.Fatalf("restored count = %d, want 0", outcome.RestoredCount())
	}
	if !slices.Equal(outcome.SkippedStagedPaths, []string{"sample.txt"}) {
		t.Fatalf("skipped staged paths = %v, want [sample.txt]", outcome.SkippedStagedPaths)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read sample file: %v", err)
	}
	if string(content) != "untracked-recreated\n" {
		t.Fatalf("content = %q, want untracked content", string(content))
	}
}

func TestRestoreDeletedTrackedPathsForRepairSkipsStagedDeleteWhenBrokenSymlinkExists(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not reliable in default Windows CI setup")
	}

	repo := t.TempDir()
	mustGit(t, repo, "init")
	mustGit(t, repo, "config", "user.name", "euni-test")
	mustGit(t, repo, "config", "user.email", "euni-test@example.com")

	filePath := filepath.Join(repo, "sample.txt")
	if err := os.WriteFile(filePath, []byte("base\n"), 0o644); err != nil {
		t.Fatalf("write sample file: %v", err)
	}
	mustGit(t, repo, "add", "sample.txt")
	mustGit(t, repo, "commit", "-m", "init")

	if err := os.Remove(filePath); err != nil {
		t.Fatalf("remove sample file: %v", err)
	}
	mustGit(t, repo, "add", "-u")

	if err := os.Symlink("missing-target.txt", filePath); err != nil {
		if errors.Is(err, fs.ErrPermission) {
			t.Skipf("symlink permission denied: %v", err)
		}
		t.Fatalf("create broken symlink: %v", err)
	}

	plan, err := gitListDeletedTrackedPathsForRepair(repo)
	if err != nil {
		t.Fatalf("gitListDeletedTrackedPathsForRepair: %v", err)
	}
	if !slices.Equal(plan.StagedPaths, []string{"sample.txt"}) {
		t.Fatalf("staged paths = %v, want [sample.txt]", plan.StagedPaths)
	}

	outcome, err := gitRestoreDeletedTrackedPathsForRepair(repo, plan)
	if err != nil {
		t.Fatalf("gitRestoreDeletedTrackedPathsForRepair: %v", err)
	}
	if outcome.RestoredCount() != 0 {
		t.Fatalf("restored count = %d, want 0", outcome.RestoredCount())
	}
	if !slices.Equal(outcome.SkippedStagedPaths, []string{"sample.txt"}) {
		t.Fatalf("skipped staged paths = %v, want [sample.txt]", outcome.SkippedStagedPaths)
	}

	target, err := os.Readlink(filePath)
	if err != nil {
		t.Fatalf("broken symlink must remain: %v", err)
	}
	if target != "missing-target.txt" {
		t.Fatalf("symlink target = %q, want %q", target, "missing-target.txt")
	}
}

func mustGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	res, err := runGit(repo, args...)
	if err != nil {
		t.Fatalf("git %v failed: %v stderr=%q stdout=%q", args, err, strings.TrimSpace(res.stderr), strings.TrimSpace(res.stdout))
	}
}
