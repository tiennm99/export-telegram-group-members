package compare

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"

	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
)

func TestParseArgs(t *testing.T) {
	for _, args := range [][]string{{}, {"-100"}, {"-100", "t1", "t2"}} {
		if _, err := parseArgs(args); err != nil {
			t.Errorf("parseArgs(%v): %v", args, err)
		}
	}
	for _, args := range [][]string{{"-100", "t1"}, {"a", "b", "c", "d"}, {"x"}} {
		if _, err := parseArgs(args); err == nil {
			t.Errorf("parseArgs(%v) accepted bad input", args)
		}
	}
}

func TestRunComparesLatestTwoCrawls(t *testing.T) {
	mr := miniredis.RunT(t)
	t.Setenv("REDIS_URL", "redis://"+mr.Addr())
	mr.Set("telegram-export:group:-100:20260701000000",
		`{"group_id":-100,"title":"Team","time":"20260701000000","members":[{"id":1,"username":"amy"},{"id":2,"username":null}]}`)
	mr.Set("telegram-export:group:-100:20260702000000",
		`{"group_id":-100,"title":"Team","time":"20260702000000","members":[{"id":2},{"id":3,"username":"cat","first_name":"Cat"}]}`)

	var out bytes.Buffer
	if err := Run(context.Background(), []string{"-100"}, cli.IO{Out: &out}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := strings.Join([]string{
		`diff --telegram-group "Team (-100)"`,
		`--- crawl/20260701000000`,
		`+++ crawl/20260702000000`,
		`@@ members: 2 -> 2 (+1 -1) @@`,
		``,
		`- 1   @amy`,
		`+ 3   @cat      Cat`,
		``,
	}, "\n")
	if out.String() != want {
		t.Fatalf("output:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestRunNeedsTwoCrawls(t *testing.T) {
	mr := miniredis.RunT(t)
	t.Setenv("REDIS_URL", "redis://"+mr.Addr())

	err := Run(context.Background(), []string{"-100"}, cli.IO{Out: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "need at least 2 exports for group -100") {
		t.Fatalf("Run error = %v", err)
	}
}
