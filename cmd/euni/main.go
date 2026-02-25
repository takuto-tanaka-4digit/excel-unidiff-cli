package main

import (
	"fmt"
	"os"

	"github.com/takuto-tanaka-4digit/excel-unidiff-cli/internal/euni"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	exitCode := run(os.Args[1:])
	os.Exit(exitCode)
}

func run(args []string) int {
	opts, err := euni.ParseOptions(args)
	if err != nil {
		if ug, ok := err.(euni.UGError); ok {
			fmt.Fprintf(os.Stderr, "[%s] %s (hint: %s)\n", ug.Code, ug.Message, ug.Hint)
			return 2
		}
		fmt.Fprintf(os.Stderr, "[UG002] %s\n", err.Error())
		return 2
	}

	logger, cleanup, logErr := euni.NewLogger(os.Stderr, opts.LogFile, opts.Quiet)
	if logErr != nil {
		fmt.Fprintf(os.Stderr, "[UG002] failed to open log file: %v (hint: %s)\n", logErr, "Retry with a writable --log-file path.")
		return 2
	}
	defer cleanup()

	service := euni.NewService(os.Stdout, logger, version, commit)
	return service.Run(opts)
}
