package configure

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/tiennm99/export-telegram-group-members/internal/prompt"
)

func TestParseGroupList(t *testing.T) {
	got, err := ParseGroupList(" -1001, ,-2 ")
	if err != nil || !reflect.DeepEqual(got, []int64{-1001, -2}) {
		t.Fatalf("ParseGroupList = %v, %v", got, err)
	}
	for _, bad := range []string{"", " , ", "abc"} {
		if _, err := ParseGroupList(bad); err == nil {
			t.Errorf("ParseGroupList(%q) accepted bad input", bad)
		}
	}
}

func TestCollectPromptsOnlyForMissingValues(t *testing.T) {
	var out bytes.Buffer
	ask := prompt.New(strings.NewReader("+1555\n-100,-200\n"), &out)

	cfg, err := collect(ask, 123, "hash", "", "")
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if cfg.Phone != "+1555" || !reflect.DeepEqual(cfg.GroupIDs, []int64{-100, -200}) || cfg.APIID != 123 {
		t.Fatalf("collect = %+v", cfg)
	}
	if strings.Contains(out.String(), "API ID") || !strings.Contains(out.String(), "Phone: ") {
		t.Fatalf("unexpected prompts: %q", out.String())
	}
}

func TestCollectRejectsBadAPIID(t *testing.T) {
	ask := prompt.New(strings.NewReader("abc\n"), &bytes.Buffer{})
	if _, err := collect(ask, 0, "", "", ""); err == nil || !strings.Contains(err.Error(), "integer") {
		t.Fatalf("collect error = %v", err)
	}
}
