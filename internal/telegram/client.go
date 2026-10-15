// Package telegram logs in to Telegram and reads group members.
package telegram

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"github.com/tiennm99/export-telegram-group-members/internal/prompt"
	"github.com/tiennm99/export-telegram-group-members/internal/store"
)

// Run connects with the stored session, signs in only when that session is
// missing or expired, and calls fn while the connection is open.
func Run(
	ctx context.Context,
	cfg store.AppConfig,
	storage session.Storage,
	ask *prompt.Prompter,
	fn func(ctx context.Context, api *tg.Client) error,
) error {
	client := telegram.NewClient(cfg.APIID, cfg.APIHash, telegram.Options{
		SessionStorage: storage,
		NoUpdates:      true,
	})
	return client.Run(ctx, func(ctx context.Context) error {
		flow := auth.NewFlow(terminalAuth{phone: cfg.Phone, ask: ask}, auth.SendCodeOptions{})
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("telegram login: %w", err)
		}
		return fn(ctx, client.API())
	})
}

// terminalAuth answers the login flow from the terminal.
type terminalAuth struct {
	phone string
	ask   *prompt.Prompter
}

func (a terminalAuth) Phone(context.Context) (string, error) { return a.phone, nil }

func (a terminalAuth) Password(context.Context) (string, error) {
	return a.ask.Secret("Password")
}

func (a terminalAuth) Code(context.Context, *tg.AuthSentCode) (string, error) {
	return a.ask.Line("Enter code")
}

func (a terminalAuth) AcceptTermsOfService(context.Context, tg.HelpTermsOfService) error {
	return errors.New("this phone has no Telegram account; sign up in an official app first")
}

func (a terminalAuth) SignUp(context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("this phone has no Telegram account; sign up in an official app first")
}

// maxFloodWait caps how long one call may sleep on FLOOD_WAIT before giving up.
const maxFloodWait = 5 * time.Minute

// withFloodWait retries call after Telegram's requested FLOOD_WAIT delay, so
// paging a large group does not fail on rate limits.
func withFloodWait[T any](ctx context.Context, call func() (T, error)) (T, error) {
	for {
		result, err := call()
		wait, ok := tgerr.AsFloodWait(err)
		if !ok || wait > maxFloodWait {
			return result, err
		}
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(wait + time.Second):
		}
	}
}
