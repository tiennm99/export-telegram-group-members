package store

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestConfigRoundTrip(t *testing.T) {
	st, _ := newTestStore(t)
	ctx := context.Background()
	want := AppConfig{APIID: 123, APIHash: "hash", Phone: "+15551234567", GroupIDs: []int64{-1001, -2}}

	if err := st.SaveConfig(ctx, want); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	got, err := st.LoadConfig(ctx)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadConfig = %+v, want %+v", got, want)
	}
}

func TestLoadConfigReadsPythonWrittenJSON(t *testing.T) {
	st, mr := newTestStore(t)
	mr.Set("telegram-export:config",
		`{"api_id": 123, "api_hash": "hash", "phone": "+1555", "group_ids": [-1001234567890]}`)

	got, err := st.LoadConfig(context.Background())
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got.APIID != 123 || got.GroupIDs[0] != -1001234567890 {
		t.Fatalf("LoadConfig = %+v", got)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"missing key", "", "run configure first"},
		{"corrupt json", "{", "not valid JSON"},
		{"missing fields", `{"api_id": 1}`, "missing: api_hash, phone, group_ids"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, mr := newTestStore(t)
			if tt.value != "" {
				mr.Set("telegram-export:config", tt.value)
			}
			_, err := st.LoadConfig(context.Background())
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("LoadConfig error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}
