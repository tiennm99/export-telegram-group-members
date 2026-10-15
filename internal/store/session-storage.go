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
//
// The Python version stored a Telethon StringSession under "session". This
// version keeps its own format under a separate key and only reads the old
// one once, so an existing login carries over without a new code.
func sessionKey() string         { return key("session-gotd") }
func telethonSessionKey() string { return key("session") }

// SessionStorage implements gotd's session.Storage on Redis.
type SessionStorage struct {
	store *Store
}

// SessionStorage returns the Redis-backed Telegram session storage.
func (s *Store) SessionStorage() *SessionStorage {
	return &SessionStorage{store: s}
}

// LoadSession returns the stored session, importing a Telethon session when
// no gotd session exists yet.
func (ss *SessionStorage) LoadSession(ctx context.Context) ([]byte, error) {
	raw, err := ss.store.rdb.Get(ctx, sessionKey()).Bytes()
	if err == nil {
		return raw, nil
	}
	if !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("read session: %w", err)
	}

	telethon, err := ss.store.rdb.Get(ctx, telethonSessionKey()).Result()
	if errors.Is(err, redis.Nil) || telethon == "" {
		return nil, session.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read Telethon session: %w", err)
	}

	raw, err = convertTelethonSession(ctx, telethon)
	if err != nil {
		// An unreadable old session should mean a fresh login, not a dead end.
		return nil, session.ErrNotFound
	}
	if err := ss.StoreSession(ctx, raw); err != nil {
		return nil, err
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

func convertTelethonSession(ctx context.Context, telethon string) ([]byte, error) {
	data, err := session.TelethonSession(telethon)
	if err != nil {
		return nil, err
	}
	var mem session.StorageMemory
	if err := (&session.Loader{Storage: &mem}).Save(ctx, data); err != nil {
		return nil, err
	}
	return mem.LoadSession(ctx)
}
