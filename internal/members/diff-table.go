package members

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/tiennm99/export-telegram-group-members/internal/store"
)

// ANSI colors for git-diff-style output.
const (
	Red   = "\033[31m"
	Green = "\033[32m"
	Cyan  = "\033[36m"
	reset = "\033[0m"
)

// Printer writes diff lines, coloring them only when Color is set (stdout is
// a terminal), so piped output stays plain.
type Printer struct {
	Out   io.Writer
	Color bool
}

// Line prints text in the given color, or plain when color is empty.
func (p Printer) Line(text, color string) {
	if p.Color && color != "" {
		text = color + text + reset
	}
	fmt.Fprintln(p.Out, text)
}

// Table aligns member rows: ID, @username, and full name.
type Table struct {
	widths [3]int
}

// NewTable sizes the columns to fit every given member.
func NewTable(lists ...[]store.Member) Table {
	t := Table{widths: [3]int{2, 8, 4}}
	for _, list := range lists {
		for _, member := range list {
			for i, value := range columns(member) {
				t.widths[i] = max(t.widths[i], utf8.RuneCountInString(value))
			}
		}
	}
	return t
}

// Row formats one member behind a one-character prefix.
func (t Table) Row(prefix string, member store.Member) string {
	cols := columns(member)
	row := prefix + " " +
		pad(cols[0], t.widths[0]) + "  " +
		pad(cols[1], t.widths[1]) + "  " +
		pad(cols[2], t.widths[2])
	return strings.TrimRight(row, " ")
}

func columns(member store.Member) [3]string {
	username := ""
	if member.Username != "" {
		username = "@" + member.Username
	}
	var names []string
	for _, part := range []string{member.FirstName, member.LastName} {
		if part != "" {
			names = append(names, part)
		}
	}
	return [3]string{
		strconv.FormatInt(member.ID, 10),
		username,
		strings.Join(names, " "),
	}
}

func pad(value string, width int) string {
	return value + strings.Repeat(" ", max(0, width-utf8.RuneCountInString(value)))
}
