package euni

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/takuto-tanaka-4digit/excel-unidiff-cli/internal/policy"
	"github.com/takuto-tanaka-4digit/excel-unidiff-cli/internal/report"
)

func TestApplyDryRunRepairUnicodeDeletesUsesSchemaCompatibleKind(t *testing.T) {
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

	policyPath := filepath.Join(repo, policy.DefaultPolicyFile)
	if err := os.WriteFile(policyPath, []byte(policy.TemplatePolicyYAML), 0o644); err != nil {
		t.Fatalf("write policy file: %v", err)
	}

	if err := os.Remove(filePath); err != nil {
		t.Fatalf("remove sample file: %v", err)
	}
	mustGit(t, repo, "add", "-u")

	logger, cleanup, err := NewLogger(io.Discard, "", true)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer cleanup()

	var out bytes.Buffer
	svc := NewService(&out, logger, "test", "commit")
	exitCode := svc.Run(Options{
		Command:              "apply",
		Repo:                 repo,
		PolicyPath:           policyPath,
		Format:               "json",
		DryRun:               true,
		RepairUnicodeDeletes: true,
	})
	if exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", exitCode)
	}

	var rpt report.Report
	if err := json.Unmarshal(out.Bytes(), &rpt); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}

	found := false
	for _, r := range rpt.Results {
		if r.Code != "UG014" {
			continue
		}
		found = true
		if r.Kind != "environment" {
			t.Fatalf("UG014 kind = %q, want environment", r.Kind)
		}
		if r.TargetType != "path" {
			t.Fatalf("UG014 targetType = %q, want path", r.TargetType)
		}
		if r.Details == nil {
			t.Fatalf("UG014 details is nil")
		}
		if v, ok := r.Details["restoreCount"].(float64); !ok || int(v) != 1 {
			t.Fatalf("UG014 details.restoreCount = %#v, want 1", r.Details["restoreCount"])
		}
		stagedRaw, ok := r.Details["restorableStagedPaths"].([]any)
		if !ok {
			t.Fatalf("UG014 details.restorableStagedPaths type = %T, want []any", r.Details["restorableStagedPaths"])
		}
		staged := toStringSlice(t, stagedRaw, "UG014 details.restorableStagedPaths")
		if !slices.Equal(staged, []string{"sample.txt"}) {
			t.Fatalf("UG014 details.restorableStagedPaths = %v, want [sample.txt]", staged)
		}
		worktreeRaw, ok := r.Details["worktreePaths"].([]any)
		if !ok {
			t.Fatalf("UG014 details.worktreePaths type = %T, want []any", r.Details["worktreePaths"])
		}
		if len(worktreeRaw) != 0 {
			t.Fatalf("UG014 details.worktreePaths = %v, want []", worktreeRaw)
		}
		if v, ok := r.Details["skipCount"].(float64); !ok || int(v) != 0 {
			t.Fatalf("UG014 details.skipCount = %#v, want 0", r.Details["skipCount"])
		}
	}
	if !found {
		t.Fatalf("UG014 result not found in report: %+v", rpt.Results)
	}
}

func TestApplyRepairUnicodeDeletesSkipsConflictingStagedDelete(t *testing.T) {
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

	policyPath := filepath.Join(repo, policy.DefaultPolicyFile)
	if err := os.WriteFile(policyPath, []byte(policy.TemplatePolicyYAML), 0o644); err != nil {
		t.Fatalf("write policy file: %v", err)
	}

	if err := os.Remove(filePath); err != nil {
		t.Fatalf("remove sample file: %v", err)
	}
	mustGit(t, repo, "add", "-u")
	if err := os.WriteFile(filePath, []byte("untracked-recreated\n"), 0o644); err != nil {
		t.Fatalf("recreate sample file: %v", err)
	}

	logger, cleanup, err := NewLogger(io.Discard, "", true)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer cleanup()

	var out bytes.Buffer
	svc := NewService(&out, logger, "test", "commit")
	exitCode := svc.Run(Options{
		Command:              "apply",
		Repo:                 repo,
		PolicyPath:           policyPath,
		Format:               "json",
		RepairUnicodeDeletes: true,
	})
	if exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", exitCode)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read file after apply: %v", err)
	}
	if string(content) != "untracked-recreated\n" {
		t.Fatalf("content after apply = %q, want untracked content", string(content))
	}

	var rpt report.Report
	if err := json.Unmarshal(out.Bytes(), &rpt); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}

	found := false
	for _, r := range rpt.Results {
		if r.Code != "UG014" {
			continue
		}
		found = true
		if v, ok := r.Details["skipCount"].(float64); !ok || int(v) != 1 {
			t.Fatalf("UG014 details.skipCount = %#v, want 1", r.Details["skipCount"])
		}
		skipRaw, ok := r.Details["skipPaths"].([]any)
		if !ok {
			t.Fatalf("UG014 details.skipPaths type = %T, want []any", r.Details["skipPaths"])
		}
		if !slices.Equal(toStringSlice(t, skipRaw, "UG014 details.skipPaths"), []string{"sample.txt"}) {
			t.Fatalf("UG014 details.skipPaths = %v, want [sample.txt]", skipRaw)
		}
	}
	if !found {
		t.Fatalf("UG014 result not found in report: %+v", rpt.Results)
	}
}

func toStringSlice(t *testing.T, in []any, fieldName string) []string {
	t.Helper()
	out := make([]string, 0, len(in))
	for _, item := range in {
		s, ok := item.(string)
		if !ok {
			t.Fatalf("%s item type = %T, want string", fieldName, item)
		}
		out = append(out, s)
	}
	return out
}
