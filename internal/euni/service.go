package euni

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/takuto-tanaka-4digit/excel-unidiff-cli/internal/policy"
	"github.com/takuto-tanaka-4digit/excel-unidiff-cli/internal/report"
	"github.com/takuto-tanaka-4digit/excel-unidiff-cli/internal/scan"
)

type Service struct {
	stdout  io.Writer
	logger  *Logger
	version string
	commit  string
}

func NewService(stdout io.Writer, logger *Logger, version, commit string) *Service {
	return &Service{stdout: stdout, logger: logger, version: version, commit: commit}
}

func (s *Service) Run(opts Options) int {
	switch opts.Command {
	case "version":
		return s.runVersion()
	case "init-policy":
		return s.runInitPolicy(opts)
	case "check", "apply", "doctor", "scan":
		return s.runOperational(opts)
	default:
		err := NewUGError("UG009", "unsupported command", hintFor("UG009"))
		s.logger.UGLine(err.Code, err.Message, err.Hint)
		return 2
	}
}

func (s *Service) runVersion() int {
	fmt.Fprintf(s.stdout, "euni %s (%s)\n", s.version, s.commit)
	return 0
}

func (s *Service) runInitPolicy(opts Options) int {
	rootPath, _, err := resolveRepoPaths(opts.Repo)
	if err != nil {
		s.logger.UGLine("UG001", "--repo is not a valid Git repository", hintFor("UG001"))
		return 2
	}
	target := filepath.Join(filepath.FromSlash(rootPath), policy.DefaultPolicyFile)

	info, statErr := os.Lstat(target)
	if statErr == nil {
		if !opts.Force {
			s.logger.UGLine("UG008", ".euni.yml already exists", hintFor("UG008"))
			return 2
		}
		if info.Mode()&os.ModeSymlink != 0 {
			s.logger.UGLine("UG010", ".euni.yml is a symlink/reparse point", hintFor("UG010"))
			return 2
		}
		resolved, err := canonicalPath(target)
		if err != nil || !isWithinBoundary(rootPath, resolved) {
			s.logger.UGLine("UG010", ".euni.yml resolves outside --repo boundary", hintFor("UG010"))
			return 2
		}
	} else if !os.IsNotExist(statErr) {
		s.logger.UGLine("UG002", "failed to inspect .euni.yml", hintFor("UG002"))
		return 2
	}

	if err := os.WriteFile(target, []byte(policy.TemplatePolicyYAML), 0o644); err != nil {
		s.logger.UGLine("UG002", "failed to write .euni.yml", hintFor("UG002"))
		return 2
	}
	fmt.Fprintf(s.stdout, "%s\n", filepath.ToSlash(target))
	return 0
}

func (s *Service) runOperational(opts Options) int {
	start := time.Now().UTC()
	rpt := report.Report{
		SchemaVersion: report.SchemaVersion,
		Command:       opts.Command,
		Recursive:     opts.Recursive,
		Results:       []report.Result{},
		Errors:        []report.ErrorItem{},
		StartedAt:     start.Format(time.RFC3339),
	}

	discovery := discoverRepositories(opts.Repo, opts.Recursive)
	rpt.Repo = discovery.RootCanonical
	if rpt.Repo == "" {
		rpt.Repo = opts.Repo
	}
	rpt.Summary.TargetRepos = len(discovery.Repos)

	for _, f := range discovery.Findings {
		if f.RepoPath == "" {
			f.RepoPath = rpt.Repo
		}
		rpt.Results = append(rpt.Results, f)
	}
	rpt.Errors = append(rpt.Errors, discovery.Errors...)

	if len(rpt.Errors) == 0 {
		var p *policy.Policy
		if opts.Command == "check" || opts.Command == "apply" || opts.Command == "doctor" {
			loaded, err := policy.Load(opts.PolicyPath)
			if err != nil {
				code := "UG003"
				message := "failed to load policy file"
				if errors.Is(err, policy.ErrInvalidPolicy) {
					code = "UG007"
					message = "policy contains unsupported keys or invalid structure"
				}
				rpt.Errors = append(rpt.Errors, errorItem(code, message, strPtr(opts.PolicyPath), nil))
			} else {
				p = &loaded
			}
		}

		if len(rpt.Errors) == 0 {
			results, errors, summary := s.evaluate(discovery.Repos, opts, p)
			rpt.Results = append(rpt.Results, results...)
			rpt.Errors = append(rpt.Errors, errors...)
			rpt.Summary.NFCOnly += summary.NFCOnly
			rpt.Summary.NFDOnly += summary.NFDOnly
			rpt.Summary.NFCCollisions += summary.NFCCollisions
			rpt.Summary.CombiningMarkPaths += summary.CombiningMarkPaths
		}
	}

	rpt.Summary.Findings = len(rpt.Results)
	rpt.Summary.Errors = len(rpt.Errors)
	if unique := uniqueRepoPathsInReport(rpt); unique > rpt.Summary.TargetRepos {
		rpt.Summary.TargetRepos = unique
	}
	rpt.Summary.DurationMs = int(time.Since(start).Milliseconds())
	rpt.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	rpt.Status, rpt.ExitCode = report.DeriveStatus(rpt.Summary.Findings, rpt.Summary.Errors)

	report.SortResultsAndErrors(&rpt)
	if opts.Format == "json" {
		encoder := json.NewEncoder(s.stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(rpt); err != nil {
			s.logger.UGLine("UG002", "failed to serialize JSON report", hintFor("UG002"))
			return 2
		}
		return rpt.ExitCode
	}

	s.emitTextReport(rpt)
	return rpt.ExitCode
}

func (s *Service) evaluate(repos []repoTarget, opts Options, p *policy.Policy) ([]report.Result, []report.ErrorItem, report.Summary) {
	results := make([]report.Result, 0)
	errs := make([]report.ErrorItem, 0)
	summary := report.Summary{}

	includeUntracked := opts.Command == "scan" || opts.Command == "doctor"
	applyMode := opts.Command == "apply"

	for _, repoTarget := range repos {
		repoPath := repoTarget.AbsPath

		if p != nil {
			expected, unresolved := p.ExpectedFor(repoTarget.RelPath, repoTarget.Kind)
			if unresolved != nil {
				details := map[string]any{
					"normalizedPath": unresolved.Normalized,
					"candidates":     unresolved.Candidates,
				}
				path := repoTarget.RelPath
				entry := resultItem(
					"UG013",
					"policy",
					"policy path is ambiguous by case-only collision",
					&path,
					"path",
					nil,
					nil,
					strPtr("resolve case-only path collision in .euni.yml"),
					details,
				)
				entry.RepoPath = repoPath
				results = append(results, entry)
			}

			keys := mapKeys(expected)
			sort.Strings(keys)
			for _, key := range keys {
				expectedValue := expected[key]
				actualValue, err := gitConfigGetBool(repoTarget.AbsPath, key)
				if err != nil {
					errs = append(errs, errorItem("UG002", "failed to read git config", strPtr(repoPath), strPtr(key)))
					continue
				}
				if actualValue == expectedValue {
					continue
				}

				expectedText := strconv.FormatBool(expectedValue)
				actualText := strconv.FormatBool(actualValue)
				if applyMode && !opts.DryRun {
					if err := gitConfigSetBool(repoTarget.AbsPath, key, expectedValue); err != nil {
						errs = append(errs, errorItem("UG002", "failed to set git config", strPtr(repoPath), strPtr(key)))
						continue
					}
					continue
				}

				action := "run euni apply"
				if applyMode && opts.DryRun {
					action = fmt.Sprintf("would set %s=%s", key, expectedText)
				}
				entry := resultItem(
					"UG004",
					"drift",
					fmt.Sprintf("config drift detected for %s", key),
					nil,
					"configKey",
					&expectedText,
					&actualText,
					&action,
					nil,
				)
				entry.RepoPath = repoPath
				results = append(results, entry)
			}
		}

		if applyMode && opts.RepairUnicodeDeletes {
			repairPlan, err := gitListDeletedTrackedPathsForRepair(repoTarget.AbsPath)
			if err != nil {
				errs = append(errs, errorItem("UG002", "failed to list deleted tracked paths for unicode repair", strPtr(repoPath), nil))
			} else if repairPlan.TotalCount() > 0 {
				restorableStagedPaths, conflictingStagedPaths, splitErr := splitRestorableAndConflictingStagedDeletedPaths(repoTarget.AbsPath, repairPlan.StagedPaths)
				if splitErr != nil {
					errs = append(errs, errorItem("UG002", "failed to inspect staged deleted paths for unicode repair", strPtr(repoPath), nil))
					continue
				}
				plannedRestoreCount := len(restorableStagedPaths) + len(repairPlan.WorktreePaths)

				if opts.DryRun {
					count := strconv.Itoa(plannedRestoreCount)
					action := "would restore staged deletions with --staged --worktree and worktree deletions with --worktree (precomposeunicode=false)"
					message := fmt.Sprintf("unicode delete repair would restore %s path(s)", count)
					if len(conflictingStagedPaths) > 0 {
						message = fmt.Sprintf(
							"unicode delete repair would restore %s path(s); %d staged path(s) would be skipped to avoid overwriting existing files",
							count,
							len(conflictingStagedPaths),
						)
					}
					entry := resultItem(
						"UG014",
						"environment",
						message,
						nil,
						"path",
						nil,
						nil,
						&action,
						map[string]any{
							"detectedCount":         repairPlan.TotalCount(),
							"restoreCount":          plannedRestoreCount,
							"stagedDeleteCount":     len(repairPlan.StagedPaths),
							"worktreeDeleteCount":   len(repairPlan.WorktreePaths),
							"restorableStagedPaths": restorableStagedPaths,
							"worktreePaths":         repairPlan.WorktreePaths,
							"skipCount":             len(conflictingStagedPaths),
							"skipPaths":             conflictingStagedPaths,
						},
					)
					entry.RepoPath = repoPath
					results = append(results, entry)
				} else {
					outcome, err := gitRestoreDeletedTrackedPathsForRepair(repoTarget.AbsPath, repairPlan)
					if err != nil {
						errs = append(errs, errorItem("UG002", "failed to restore deleted tracked paths for unicode repair", strPtr(repoPath), nil))
					} else {
						if outcome.RestoredCount() > 0 {
							s.logger.Progressf("repaired unicode delete drift in %s: restored %d path(s)", repoPath, outcome.RestoredCount())
						}
						if len(outcome.SkippedStagedPaths) > 0 {
							action := "review skipped paths and resolve manually to avoid overwriting existing files"
							entry := resultItem(
								"UG014",
								"environment",
								fmt.Sprintf(
									"unicode delete repair skipped %d staged path(s) to avoid overwriting existing files",
									len(outcome.SkippedStagedPaths),
								),
								nil,
								"path",
								nil,
								nil,
								&action,
								map[string]any{
									"detectedCount":         repairPlan.TotalCount(),
									"restoreCount":          outcome.RestoredCount(),
									"skipCount":             len(outcome.SkippedStagedPaths),
									"skipPaths":             outcome.SkippedStagedPaths,
									"restoredStagedPaths":   outcome.RestoredStagedPaths,
									"restoredWorktreePaths": outcome.RestoredWorktreePaths,
								},
							)
							entry.RepoPath = repoPath
							results = append(results, entry)
						}
					}
				}
			}
		}

		tracked, err := gitListTracked(repoTarget.AbsPath)
		if err != nil {
			errs = append(errs, errorItem("UG002", "failed to list tracked files", strPtr(repoPath), nil))
			continue
		}
		paths := append([]string{}, tracked...)
		if includeUntracked {
			untracked, untrackedErr := gitListUntracked(repoTarget.AbsPath)
			if untrackedErr != nil {
				errs = append(errs, errorItem("UG002", "failed to list untracked files", strPtr(repoPath), nil))
				continue
			}
			paths = append(paths, untracked...)
		}

		analysis := scan.AnalyzePaths(paths)
		summary.NFCOnly += analysis.Metrics.NFCOnly
		summary.NFDOnly += analysis.Metrics.NFDOnly
		summary.NFCCollisions += analysis.Metrics.NFCCollisions
		summary.CombiningMarkPaths += analysis.Metrics.CombiningMarkPaths

		for _, collision := range analysis.Collisions {
			normalizedPath := collision.NormalizedPath
			entry := resultItem(
				"UG005",
				"unicode",
				"NFC path collision detected",
				&normalizedPath,
				"path",
				nil,
				nil,
				strPtr("rename colliding paths to unique NFC names"),
				map[string]any{
					"normalizedPath": collision.NormalizedPath,
					"collidingPaths": collision.CollidingPaths,
				},
			)
			entry.RepoPath = repoPath
			results = append(results, entry)
		}
		for _, pth := range analysis.CombiningMarkPaths {
			pathCopy := pth
			entry := resultItem(
				"UG011",
				"unicode",
				"combining mark detected in path",
				&pathCopy,
				"path",
				nil,
				nil,
				strPtr("rename path without combining marks"),
				nil,
			)
			entry.RepoPath = repoPath
			results = append(results, entry)
		}
	}

	return results, errs, summary
}

func (s *Service) emitTextReport(rpt report.Report) {
	fmt.Fprintf(s.stdout, "command=%s status=%s exit=%d\n", rpt.Command, rpt.Status, rpt.ExitCode)
	fmt.Fprintf(s.stdout, "repo=%s recursive=%t targets=%d findings=%d errors=%d durationMs=%d\n", rpt.Repo, rpt.Recursive, rpt.Summary.TargetRepos, rpt.Summary.Findings, rpt.Summary.Errors, rpt.Summary.DurationMs)
	fmt.Fprintf(s.stdout, "unicode: nfcOnly=%d nfdOnly=%d nfcCollisions=%d combiningMarkPaths=%d\n", rpt.Summary.NFCOnly, rpt.Summary.NFDOnly, rpt.Summary.NFCCollisions, rpt.Summary.CombiningMarkPaths)

	for _, item := range rpt.Results {
		hint := hintFor(item.Code)
		fmt.Fprintf(s.stdout, "[%s] %s (hint: %s)\n", item.Code, item.Message, hint)
	}
	for _, item := range rpt.Errors {
		hint := ""
		if item.Hint != nil {
			hint = *item.Hint
		}
		fmt.Fprintf(s.stdout, "[%s] %s (hint: %s)\n", item.Code, item.Message, hint)
	}
}

func mapKeys[K comparable, V any](in map[K]V) []K {
	out := make([]K, 0, len(in))
	for key := range in {
		out = append(out, key)
	}
	return out
}

func uniqueRepoPathsInReport(r report.Report) int {
	seen := make(map[string]struct{})
	for _, item := range r.Results {
		if item.RepoPath != "" {
			seen[item.RepoPath] = struct{}{}
		}
	}
	for _, item := range r.Errors {
		if item.RepoPath != nil && *item.RepoPath != "" {
			seen[*item.RepoPath] = struct{}{}
		}
	}
	return len(seen)
}
