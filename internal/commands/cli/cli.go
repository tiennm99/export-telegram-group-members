// Package cli holds what every command shares: process wiring and
// positional-argument parsing.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/mattn/go-isatty"
)

// IO is a command's terminal. Color is set when Out is a terminal, so piped
// output stays free of ANSI codes.
type IO struct {
	In    io.Reader
	Out   io.Writer
	Err   io.Writer
	Color bool
}

// RunFunc is a command entry point. It returns an error to exit with status 1.
type RunFunc func(ctx context.Context, args []string, term IO) error

// Main runs a command against the process terminal and exits on error.
func Main(run RunFunc) {
	term := IO{
		In:    os.Stdin,
		Out:   os.Stdout,
		Err:   os.Stderr,
		Color: isatty.IsTerminal(os.Stdout.Fd()),
	}
	if err := run(context.Background(), os.Args[1:], term); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// WantsHelp reports whether args ask for usage. Commands parse positionals
// by hand because the flag package would read a negative group ID such as
// -1001234567890 as an unknown flag.
func WantsHelp(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

// ParseGroupID parses one group ID argument.
func ParseGroupID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid group id %q: must be an integer", value)
	}
	return id, nil
}

// ParseGroupIDs parses every argument as a group ID.
func ParseGroupIDs(args []string) ([]int64, error) {
	ids := make([]int64, 0, len(args))
	for _, arg := range args {
		id, err := ParseGroupID(arg)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
