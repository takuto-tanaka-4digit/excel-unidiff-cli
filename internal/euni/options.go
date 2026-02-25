package euni

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/takuto-tanaka-4digit/excel-unidiff-cli/internal/policy"
)

type Options struct {
	Command        string
	Repo           string
	Recursive      bool
	PolicyPath     string
	Format         string
	Quiet          bool
	NonInteractive bool
	LogFile        string
	DryRun         bool
	Force          bool
}

func ParseOptions(argv []string) (Options, error) {
	if len(argv) == 0 {
		return Options{}, NewUGError("UG009", "missing command", hintFor("UG009"))
	}

	cmd := argv[0]
	switch cmd {
	case "check", "apply", "doctor", "scan":
		return parseOperationalOptions(cmd, argv[1:])
	case "init-policy":
		return parseInitPolicyOptions(argv[1:])
	case "version":
		if len(argv[1:]) > 0 {
			return Options{}, NewUGError("UG009", "version does not accept options", hintFor("UG009"))
		}
		return Options{Command: "version"}, nil
	default:
		return Options{}, NewUGError("UG009", fmt.Sprintf("unsupported command: %s", cmd), hintFor("UG009"))
	}
}

func parseOperationalOptions(cmd string, args []string) (Options, error) {
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	repo := fs.String("repo", ".", "")
	recursive := fs.Bool("recursive", false, "")
	format := fs.String("format", "text", "")
	quiet := fs.Bool("quiet", false, "")
	nonInteractive := fs.Bool("non-interactive", false, "")
	logFile := fs.String("log-file", "", "")

	var policyRef *string
	var dryRunRef *bool
	if cmd != "scan" {
		policyRef = fs.String("policy", "", "")
	}
	if cmd == "apply" {
		dryRunRef = fs.Bool("dry-run", false, "")
	}

	if err := fs.Parse(args); err != nil {
		return Options{}, NewUGError("UG009", err.Error(), hintFor("UG009"))
	}
	if fs.NArg() > 0 {
		return Options{}, NewUGError("UG009", fmt.Sprintf("unexpected positional args: %v", fs.Args()), hintFor("UG009"))
	}

	if *format != "text" && *format != "json" {
		return Options{}, NewUGError("UG009", fmt.Sprintf("unsupported --format: %s", *format), hintFor("UG009"))
	}

	policyPath := ""
	if cmd != "scan" {
		if policyRef != nil {
			policyPath = *policyRef
		}
		if policyPath == "" {
			policyPath = filepath.Join(*repo, policy.DefaultPolicyFile)
		}
	}

	dryRun := false
	if dryRunRef != nil {
		dryRun = *dryRunRef
	}

	return Options{
		Command:        cmd,
		Repo:           *repo,
		Recursive:      *recursive,
		PolicyPath:     policyPath,
		Format:         *format,
		Quiet:          *quiet,
		NonInteractive: *nonInteractive,
		LogFile:        *logFile,
		DryRun:         dryRun,
	}, nil
}

func parseInitPolicyOptions(args []string) (Options, error) {
	fs := flag.NewFlagSet("init-policy", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	repo := fs.String("repo", ".", "")
	nonInteractive := fs.Bool("non-interactive", false, "")
	force := fs.Bool("force", false, "")

	if err := fs.Parse(args); err != nil {
		return Options{}, NewUGError("UG009", err.Error(), hintFor("UG009"))
	}
	if fs.NArg() > 0 {
		return Options{}, NewUGError("UG009", fmt.Sprintf("unexpected positional args: %v", fs.Args()), hintFor("UG009"))
	}

	return Options{
		Command:        "init-policy",
		Repo:           *repo,
		NonInteractive: *nonInteractive,
		Force:          *force,
	}, nil
}
