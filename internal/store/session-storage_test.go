package store

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/gotd/td/session"
)

// telethonString builds a Telethon StringSession: version "1" plus base64
// of DC ID, IPv4, port, and a 256-byte auth key.
func telethonString(dc byte) string {
	raw := []byte{dc, 149, 154, 167, 51, 0x01, 0xbb}
	raw = append(raw, bytes.Repeat([]byte{7}, 256)...)
	return "1" + base64.URLEncoding.EncodeToString(raw)
}

func TestLoadSessionNotFound(t *testing.T) {
	st, _ := newTestStore(t)

	_, err := st.SessionStorage().LoadSession(context.Background())
	if !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("LoadSession error = %v, want session.ErrNotFound", err)
	}
}

func TestSessionRoundTrip(t *testing.T) {
	st, _ := newTestStore(t)
	ctx := context.Background()

	if err := st.SessionStorage().StoreSession(ctx, []byte(`{"Version":1}`)); err != nil {
		t.Fatalf("StoreSession: %v", err)
	}
	got, err := st.SessionStorage().LoadSession(ctx)
	if err != nil || string(got) != `{"Version":1}` {
		t.Fatalf("LoadSession = %q, %v", got, err)
	}
}

func TestLoadSessionImportsTelethonSession(t *testing.T) {
	st, mr := newTestStore(t)
	ctx := context.Background()
	mr.Set("telegram-export:session", telethonString(2))

	raw, err := st.SessionStorage().LoadSession(ctx)
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	data, err := (&session.Loader{Storage: st.SessionStorage()}).Load(ctx)
	if err != nil {
		t.Fatalf("Loader.Load: %v", err)
	}
	if data.DC != 2 || data.Addr != "149.154.167.51:443" {
		t.Fatalf("imported session DC=%d Addr=%q", data.DC, data.Addr)
	}
	if stored, _ := mr.Get("telegram-export:session-gotd"); stored != string(raw) {
		t.Fatal("imported session was not saved under the gotd key")
	}
	if old, _ := mr.Get("telegram-export:session"); old != telethonString(2) {
		t.Fatal("Telethon session key was modified")
	}
}

func TestLoadSessionIgnoresBrokenTelethonSession(t *testing.T) {
	st, mr := newTestStore(t)
	mr.Set("telegram-export:session", "not-a-session")

	_, err := st.SessionStorage().LoadSession(context.Background())
	if !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("LoadSession error = %v, want session.ErrNotFound", err)
	}
}
