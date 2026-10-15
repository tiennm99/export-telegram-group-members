// Package configure stores the Telegram crawl config in Redis.
package configure

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	"github.com/tiennm99/export-telegram-group-members/internal/prompt"
	"github.com/tiennm99/export-telegram-group-members/internal/store"
)

// Run asks for any value not given as a flag, then saves the config.
func Run(ctx context.Context, args []string, term cli.IO) error {
	flags := flag.NewFlagSet("configure", flag.ContinueOnError)
	flags.SetOutput(term.Err)
	flags.Usage = func() {
		fmt.Fprintln(term.Err, "usage: go run ./cmd/configure [--api-id ID] [--api-hash HASH] [--phone PHONE] [--groups ID,ID]")
		fmt.Fprintln(term.Err, "\nStore Telegram crawl config in Redis.")
		flags.PrintDefaults()
	}
	apiID := flags.Int("api-id", 0, "Telegram API ID")
	apiHash := flags.String("api-hash", "", "Telegram API hash")
	phone := flags.String("phone", "", "Telegram account phone")
	groups := flags.String("groups", "", "Comma-separated Telegram group IDs to crawl")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	ask := prompt.New(term.In, term.Out)
	cfg, err := collect(ask, *apiID, *apiHash, *phone, *groups)
	if err != nil {
		return err
	}

	st, err := store.OpenFromEnv(ctx)
	if err != nil {
		return err
	}
	defer st.Close()

	if err := st.SaveConfig(ctx, cfg); err != nil {
		return err
	}
	fmt.Fprintf(term.Out, "saved Redis config for %s: %d group(s).\n", cfg.Phone, len(cfg.GroupIDs))
	return nil
}

func collect(ask *prompt.Prompter, apiID int, apiHash, phone, groups string) (store.AppConfig, error) {
	var err error
	if apiID == 0 {
		raw, err := required(ask.Line, "API ID")
		if err != nil {
			return store.AppConfig{}, err
		}
		if apiID, err = strconv.Atoi(raw); err != nil {
			return store.AppConfig{}, errors.New("api id must be an integer")
		}
	}
	if apiHash == "" {
		if apiHash, err = required(ask.Secret, "API hash"); err != nil {
			return store.AppConfig{}, err
		}
	}
	if phone == "" {
		if phone, err = required(ask.Line, "Phone"); err != nil {
			return store.AppConfig{}, err
		}
	}
	if groups == "" {
		if groups, err = required(ask.Line, "Group IDs"); err != nil {
			return store.AppConfig{}, err
		}
	}
	groupIDs, err := ParseGroupList(groups)
	if err != nil {
		return store.AppConfig{}, err
	}
	return store.AppConfig{APIID: apiID, APIHash: apiHash, Phone: phone, GroupIDs: groupIDs}, nil
}

func required(read func(string) (string, error), label string) (string, error) {
	value, err := read(label)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", fmt.Errorf("%s is required", strings.ToLower(label))
	}
	return value, nil
}

// ParseGroupList parses comma-separated group IDs, ignoring empty entries.
func ParseGroupList(value string) ([]int64, error) {
	var ids []int64
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, errors.New("group ids must be comma-separated integers")
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, errors.New("at least one group id is required")
	}
	return ids, nil
}
