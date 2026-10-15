package store

import (
	"context"
	"errors"
	"testing"

	"github.com/gotd/td/session"
)

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
