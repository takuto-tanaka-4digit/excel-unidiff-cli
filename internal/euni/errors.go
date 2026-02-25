package euni

import (
	"fmt"

	"github.com/takuto-tanaka-4digit/excel-unidiff-cli/internal/report"
)

type UGError struct {
	Code    string
	Message string
	Hint    string
}

func (e UGError) Error() string {
	if e.Hint == "" {
		return fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s (hint: %s)", e.Code, e.Message, e.Hint)
}

func NewUGError(code, message, hint string) UGError {
	return UGError{Code: code, Message: message, Hint: hint}
}

func hintFor(code string) string {
	switch code {
	case "UG001":
		return "Verify --repo points to a readable Git repository."
	case "UG002":
		return "Retry and inspect stderr from git command execution."
	case "UG003":
		return "Check policy file path, YAML syntax, and required keys."
	case "UG004":
		return "Run euni apply to align local config with policy."
	case "UG005":
		return "Rename colliding paths to unique NFC-normalized names."
	case "UG006":
		return "Initialize submodules and ensure access permissions."
	case "UG007":
		return "Fix unsupported keys in .euni.yml (strict mode)."
	case "UG008":
		return "Use --force to overwrite existing .euni.yml explicitly."
	case "UG009":
		return "Remove unsupported options for this command."
	case "UG010":
		return "Use a repository whose gitdir/top-level stay inside --repo boundary."
	case "UG011":
		return "Avoid combining marks in filenames when possible."
	case "UG012":
		return "Inspect non-standard filesystem entries (symlink/reparse/mount)."
	case "UG013":
		return "Resolve ambiguous policy path keys and keep exact-case unique paths."
	default:
		return "See runbook for this code."
	}
}

func resultItem(code, kind, message string, path *string, targetType string, expected, actual, action *string, details map[string]any) report.Result {
	return report.Result{
		RepoPath:   "",
		Kind:       kind,
		Code:       code,
		Message:    message,
		Path:       path,
		TargetType: targetType,
		Expected:   expected,
		Actual:     actual,
		Action:     action,
		Details:    details,
	}
}

func errorItem(code, message string, repoPath, path *string) report.ErrorItem {
	return report.ErrorItem{
		Code:     code,
		Message:  message,
		RepoPath: repoPath,
		Path:     path,
		Hint:     strPtr(hintFor(code)),
	}
}

func strPtr(s string) *string {
	return &s
}
