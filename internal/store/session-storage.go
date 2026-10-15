package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotd/td/session"
	"github.com/redis/go-redis/v9"
)

// The login session grants full account access; it lives only in Redis
// (behind TLS/auth), never on local disk. Deleting it forces a clean re-login.
func sessionKey() string { return key("session-gotd") }

// SessionStorage implements gotd's session.Storage on Redis.
type SessionStorage struct {
	store *Store
}

// SessionStorage returns the Redis-backed Telegram session storage.
func (s *Store) SessionStorage() *SessionStorage {
	return &SessionStorage{store: s}
}

// LoadSession returns the stored session, or session.ErrNotFound before the
// first login.
func (ss *SessionStorage) LoadSession(ctx context.Context) ([]byte, error) {
	raw, err := ss.store.rdb.Get(ctx, sessionKey()).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, session.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read session: %w", err)
	}
	return raw, nil
}

// StoreSession persists the session for reuse on any device.
func (ss *SessionStorage) StoreSession(ctx context.Context, data []byte) error {
	if err := ss.store.rdb.Set(ctx, sessionKey(), data, 0).Err(); err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}
