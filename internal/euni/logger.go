package euni

import (
	"fmt"
	"io"
	"os"
)

type Logger struct {
	w     io.Writer
	quiet bool
}

func NewLogger(stderr io.Writer, logFile string, quiet bool) (*Logger, func(), error) {
	writer := stderr
	cleanup := func() {}

	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return nil, cleanup, err
		}
		writer = io.MultiWriter(stderr, f)
		cleanup = func() { _ = f.Close() }
	}

	return &Logger{w: writer, quiet: quiet}, cleanup, nil
}

func (l *Logger) Progressf(format string, args ...any) {
	if l.quiet {
		return
	}
	fmt.Fprintf(l.w, format+"\n", args...)
}

func (l *Logger) Linef(format string, args ...any) {
	fmt.Fprintf(l.w, format+"\n", args...)
}

func (l *Logger) UGLine(code, message, hint string) {
	fmt.Fprintf(l.w, "[%s] %s (hint: %s)\n", code, message, hint)
}
