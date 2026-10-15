package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

// AppConfig is the Telegram crawl config stored as one JSON key.
type AppConfig struct {
	APIID    int     `json:"api_id"`
	APIHash  string  `json:"api_hash"`
	Phone    string  `json:"phone"`
	GroupIDs []int64 `json:"group_ids"`
}

func configKey() string { return key("config") }

// LoadConfig reads the Telegram config and rejects missing fields.
func (s *Store) LoadConfig(ctx context.Context) (AppConfig, error) {
	raw, err := s.rdb.Get(ctx, configKey()).Result()
	if errors.Is(err, redis.Nil) {
		return AppConfig{}, errors.New("telegram config not found in Redis. run configure first")
	}
	if err != nil {
		return AppConfig{}, fmt.Errorf("read telegram config: %w", err)
	}

	var cfg AppConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return AppConfig{}, errors.New("telegram config in Redis is not valid JSON")
	}

	var missing []string
	if cfg.APIID == 0 {
		missing = append(missing, "api_id")
	}
	if cfg.APIHash == "" {
		missing = append(missing, "api_hash")
	}
	if cfg.Phone == "" {
		missing = append(missing, "phone")
	}
	if len(cfg.GroupIDs) == 0 {
		missing = append(missing, "group_ids")
	}
	if len(missing) > 0 {
		return AppConfig{}, fmt.Errorf("telegram config missing: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

// SaveConfig replaces the Telegram config.
func (s *Store) SaveConfig(ctx context.Context, cfg AppConfig) error {
	raw, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode telegram config: %w", err)
	}
	if err := s.rdb.Set(ctx, configKey(), raw, 0).Err(); err != nil {
		return fmt.Errorf("save telegram config: %w", err)
	}
	return nil
}
