package comparegroups

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	"github.com/tiennm99/export-telegram-group-members/internal/members"
	"github.com/tiennm99/export-telegram-group-members/internal/store"
	"github.com/tiennm99/export-telegram-group-members/internal/telegram"
)

func TestPrintReport(t *testing.T) {
	a := telegram.Group{ID: -101, Title: "A", Members: []store.Member{
		{ID: 1, Username: "amy"}, {ID: 2, Username: "bob"},
	}}
	b := telegram.Group{ID: -202, Title: "B", Members: []store.Member{
		{ID: 2, Username: "bob"}, {ID: 3, Username: "cat"},
	}}

	var out bytes.Buffer
	printReport(members.Printer{Out: &out}, a, b)

	want := strings.Join([]string{
		`diff --telegram-groups "A (-101)" "B (-202)"`,
		`--- A (-101)`,
		`+++ B (-202)`,
		`@@ members: 2 vs 2 (both 1, -1 only first, +1 only second) @@`,
		``,
		`  2   @bob`,
		`- 1   @amy`,
		`+ 3   @cat`,
		``,
	}, "\n")
	if out.String() != want {
		t.Fatalf("output:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestRunRequiresTwoGroupIDsBeforeRedis(t *testing.T) {
	t.Setenv("REDIS_URL", "")
	for _, args := range [][]string{{"-100"}, {"-100", "x"}} {
		err := Run(context.Background(), args, cli.IO{Out: &bytes.Buffer{}})
		if err == nil || strings.Contains(err.Error(), "REDIS_URL") {
			t.Errorf("Run(%v) error = %v, want a usage error", args, err)
		}
	}
}
