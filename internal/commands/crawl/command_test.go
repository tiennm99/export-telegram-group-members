package crawl

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
)

func TestRunRejectsBadGroupIDBeforeRedis(t *testing.T) {
	t.Setenv("REDIS_URL", "")
	var out bytes.Buffer

	err := Run(context.Background(), []string{"not-a-group"}, cli.IO{Out: &out, Err: &out})
	if err == nil || !strings.Contains(err.Error(), `invalid group id "not-a-group"`) {
		t.Fatalf("Run error = %v", err)
	}
}

func TestRunHelp(t *testing.T) {
	var out bytes.Buffer
	if err := Run(context.Background(), []string{"-h"}, cli.IO{Out: &out}); err != nil {
		t.Fatalf("Run -h: %v", err)
	}
	if !strings.Contains(out.String(), "go run ./cmd/crawl") {
		t.Fatalf("help = %q", out.String())
	}
}
