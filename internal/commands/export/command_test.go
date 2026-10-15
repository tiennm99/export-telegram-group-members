package export

import (
	"bytes"
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"

	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	"github.com/tiennm99/export-telegram-group-members/internal/store"
)

func setup(t *testing.T) (*miniredis.Miniredis, string) {
	t.Helper()
	mr := miniredis.RunT(t)
	t.Setenv("REDIS_URL", "redis://"+mr.Addr())
	dir := filepath.Join(t.TempDir(), "output")
	old := outputDir
	outputDir = dir
	t.Cleanup(func() { outputDir = old })
	return mr, dir
}

func TestRunExportsLatestCrawl(t *testing.T) {
	mr, dir := setup(t)
	mr.Set("telegram-export:group:-100:20260701000000", `{"group_id":-100,"time":"20260701000000","members":[]}`)
	mr.Set("telegram-export:group:-100:20260702000000",
		`{"group_id":-100,"time":"20260702000000","members":[{"id":1,"username":"amy","first_name":"Amy","last_name":"Lee"},{"id":2,"username":null,"first_name":"=SUM(A1)","last_name":null}]}`)

	var out bytes.Buffer
	if err := Run(context.Background(), []string{"-100"}, cli.IO{Out: &out}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	path := filepath.Join(dir, "-100-20260702000000.csv")
	if !strings.Contains(out.String(), "exported 2 members to "+path) {
		t.Fatalf("output = %q", out.String())
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open CSV: %v", err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}
	want := [][]string{
		{"id", "username", "first_name", "last_name"},
		{"1", "amy", "Amy", "Lee"},
		{"2", "", "'=SUM(A1)", ""},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("rows = %v, want %v", rows, want)
	}

	for p, mode := range map[string]os.FileMode{dir: 0o700, path: 0o600} {
		if info, err := os.Stat(p); err != nil || info.Mode().Perm() != mode {
			t.Errorf("%s mode = %v, %v; want %v", p, info.Mode().Perm(), err, mode)
		}
	}
}

func TestRunRejectsMismatchedRecord(t *testing.T) {
	mr, _ := setup(t)
	mr.Set("telegram-export:group:-100:20260702000000", `{"group_id":-999,"time":"20260702000000","members":[]}`)

	err := Run(context.Background(), []string{"-100", "20260702000000"}, cli.IO{Out: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "group ID does not match its key") {
		t.Fatalf("Run error = %v", err)
	}
}

func TestRunReportsMissingExports(t *testing.T) {
	setup(t)
	err := Run(context.Background(), []string{"-100"}, cli.IO{Out: &bytes.Buffer{}})
	if err == nil || err.Error() != "no exports found for group -100" {
		t.Fatalf("Run error = %v", err)
	}
}

func TestValidateRunTime(t *testing.T) {
	tests := map[string]string{
		"20260724120000": "",
		"2026072412":     "yyyymmddhhmmss format",
		"20260230120000": "valid date and time",
		"20261324000000": "valid date and time",
	}
	for value, want := range tests {
		err := validateRunTime(value)
		if (want == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), want)) {
			t.Errorf("validateRunTime(%q) = %v, want %q", value, err, want)
		}
	}
}

func TestEscapeFormula(t *testing.T) {
	tests := map[string]string{
		"":       "",
		"alice":  "alice",
		"=1+1":   "'=1+1",
		"  @cmd": "'  @cmd",
		"-5":     "'-5",
		"\tx":    "'\tx",
		"\nx":    "'\nx",
	}
	for in, want := range tests {
		if got := escapeFormula(in); got != want {
			t.Errorf("escapeFormula(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidateRecordRejectsMissingMemberID(t *testing.T) {
	record := store.GroupExport{GroupID: -1, Time: "t", Members: []store.Member{{Username: "x"}}}
	if _, err := validateRecord(record, -1, "t"); err == nil || !strings.Contains(err.Error(), "invalid ID") {
		t.Fatalf("validateRecord error = %v", err)
	}
}
