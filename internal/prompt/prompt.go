// Package prompt reads answers from the terminal.
package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// Prompter asks questions on out and reads answers from in.
type Prompter struct {
	in  *bufio.Reader
	fd  int
	tty bool
	out io.Writer
}

// New returns a Prompter. Secrets are read without echo when in is a
// terminal.
func New(in io.Reader, out io.Writer) *Prompter {
	p := &Prompter{in: bufio.NewReader(in), out: out}
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		p.fd, p.tty = int(f.Fd()), true
	}
	return p
}

// Line asks for one line of text and trims it.
func (p *Prompter) Line(label string) (string, error) {
	fmt.Fprintf(p.out, "%s: ", label)
	line, err := p.in.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && line != "") {
		return "", fmt.Errorf("read %s: %w", strings.ToLower(label), err)
	}
	return strings.TrimSpace(line), nil
}

// Secret asks for a value without echoing it on a terminal.
func (p *Prompter) Secret(label string) (string, error) {
	if !p.tty {
		return p.Line(label)
	}
	fmt.Fprintf(p.out, "%s: ", label)
	raw, err := term.ReadPassword(p.fd)
	fmt.Fprintln(p.out)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", strings.ToLower(label), err)
	}
	return strings.TrimSpace(string(raw)), nil
}
