package store

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func newTestStore(t *testing.T) (*Store, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	st, err := Open(context.Background(), "redis://"+mr.Addr())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st, mr
}
