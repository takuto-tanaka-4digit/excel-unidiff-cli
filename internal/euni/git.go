package euni

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
