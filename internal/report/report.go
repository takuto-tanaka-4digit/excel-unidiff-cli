package report

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/unicode/norm"
)

const SchemaVersion = "1.0.0"

type Status string

const (
	StatusOK       Status = "ok"
	StatusFindings Status = "findings"
	StatusError    Status = "error"
)

type Summary struct {
	TargetRepos        int `json:"targetRepos"`
	Findings           int `json:"findings"`
	Errors             int `json:"errors"`
	DurationMs         int `json:"durationMs"`
	NFCOnly            int `json:"nfcOnly"`
	NFDOnly            int `json:"nfdOnly"`
	NFCCollisions      int `json:"nfcCollisions"`
	CombiningMarkPaths int `json:"combiningMarkPaths"`
}

type Result struct {
	RepoPath   string         `json:"repoPath"`
	Kind       string         `json:"kind"`
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Path       *string        `json:"path"`
	TargetType string         `json:"targetType"`
	Expected   *string        `json:"expected"`
	Actual     *string        `json:"actual"`
	Action     *string        `json:"action"`
	Details    map[string]any `json:"details"`
}

type ErrorItem struct {
	Code     string  `json:"code"`
	Message  string  `json:"message"`
	RepoPath *string `json:"repoPath"`
	Path     *string `json:"path"`
	Hint     *string `json:"hint"`
}

type Report struct {
	SchemaVersion string      `json:"schemaVersion"`
	Command       string      `json:"command"`
	Repo          string      `json:"repo"`
	Recursive     bool        `json:"recursive"`
	Status        Status      `json:"status"`
	ExitCode      int         `json:"exitCode"`
	Summary       Summary     `json:"summary"`
	Results       []Result    `json:"results"`
	Errors        []ErrorItem `json:"errors"`
	StartedAt     string      `json:"startedAt"`
	FinishedAt    string      `json:"finishedAt"`
}

func DeriveStatus(findings, errors int) (Status, int) {
	if errors > 0 {
		return StatusError, 2
	}
	if findings > 0 {
		return StatusFindings, 1
	}
	return StatusOK, 0
}

func ValidateInvariants(r Report, processExitCode int) []string {
	var errs []string

	if r.SchemaVersion != SchemaVersion {
		errs = append(errs, fmt.Sprintf("schemaVersion mismatch: want %s, got %s", SchemaVersion, r.SchemaVersion))
	}

	expectedByStatus, ok := mapStatusToExitCode(r.Status)
	if !ok {
		errs = append(errs, fmt.Sprintf("status-exit mismatch: unknown status %q", r.Status))
	} else if r.ExitCode != expectedByStatus {
		errs = append(errs, fmt.Sprintf("status-exit mismatch: status %s expects %d, got %d", r.Status, expectedByStatus, r.ExitCode))
	}

	if processExitCode >= 0 && r.ExitCode != processExitCode {
		errs = append(errs, fmt.Sprintf("process exit code mismatch: process=%d report=%d", processExitCode, r.ExitCode))
	}

	if !isStatusInvariantValid(r) {
		errs = append(errs, "status invariant violation")
	}

	if r.Summary.Findings != len(r.Results) {
		errs = append(errs, fmt.Sprintf("summary.findings mismatch: want %d, got %d", len(r.Results), r.Summary.Findings))
	}
	if r.Summary.Errors != len(r.Errors) {
		errs = append(errs, fmt.Sprintf("summary.errors mismatch: want %d, got %d", len(r.Errors), r.Summary.Errors))
	}

	if uniqueRepoPathCount(r.Results, r.Errors) > r.Summary.TargetRepos {
		errs = append(errs, "summary.targetRepos too small")
	}

	return errs
}

func SortResultsAndErrors(r *Report) {
	sort.SliceStable(r.Results, func(i, j int) bool {
		left := r.Results[i]
		right := r.Results[j]

		if c := compareCanonical(left.RepoPath, right.RepoPath); c != 0 {
			return c < 0
		}
		if c := strings.Compare(left.Kind, right.Kind); c != 0 {
			return c < 0
		}
		if c := compareOptionalString(left.Path, right.Path, true); c != 0 {
			return c < 0
		}
		if c := strings.Compare(left.Code, right.Code); c != 0 {
			return c < 0
		}
		if c := compareOptionalString(left.Expected, right.Expected, false); c != 0 {
			return c < 0
		}
		return compareOptionalString(left.Actual, right.Actual, false) < 0
	})

	sort.SliceStable(r.Errors, func(i, j int) bool {
		left := r.Errors[i]
		right := r.Errors[j]

		if c := strings.Compare(left.Code, right.Code); c != 0 {
			return c < 0
		}
		if c := compareOptionalString(left.RepoPath, right.RepoPath, true); c != 0 {
			return c < 0
		}
		return compareOptionalString(left.Path, right.Path, true) < 0
	})
}

func mapStatusToExitCode(s Status) (int, bool) {
	switch s {
	case StatusOK:
		return 0, true
	case StatusFindings:
		return 1, true
	case StatusError:
		return 2, true
	default:
		return 0, false
	}
}

func isStatusInvariantValid(r Report) bool {
	switch r.Status {
	case StatusOK:
		return r.Summary.Findings == 0 && r.Summary.Errors == 0
	case StatusFindings:
		return r.Summary.Findings > 0 && r.Summary.Errors == 0
	case StatusError:
		return r.Summary.Errors > 0 && r.ExitCode == 2
	default:
		return false
	}
}

func uniqueRepoPathCount(results []Result, errors []ErrorItem) int {
	seen := make(map[string]struct{})
	for _, item := range results {
		if item.RepoPath != "" {
			seen[item.RepoPath] = struct{}{}
		}
	}
	for _, item := range errors {
		if item.RepoPath != nil && *item.RepoPath != "" {
			seen[*item.RepoPath] = struct{}{}
		}
	}
	return len(seen)
}

func compareOptionalString(left, right *string, canonicalPath bool) int {
	if left == nil && right == nil {
		return 0
	}
	if left == nil {
		return 1
	}
	if right == nil {
		return -1
	}
	if canonicalPath {
		return compareCanonical(*left, *right)
	}
	return strings.Compare(norm.NFC.String(*left), norm.NFC.String(*right))
}

func compareCanonical(left, right string) int {
	leftN := canonicalForSort(left)
	rightN := canonicalForSort(right)
	return strings.Compare(leftN, rightN)
}

func canonicalForSort(s string) string {
	v := strings.ReplaceAll(s, "\\", "/")
	if len(v) >= 2 && v[1] == ':' {
		v = strings.ToUpper(v[:1]) + v[1:]
	}
	return norm.NFC.String(v)
}
