// Package comparegroups fetches two groups live from Telegram and compares
// their members. It saves no crawl to Redis.
package comparegroups

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotd/td/tg"

	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	"github.com/tiennm99/export-telegram-group-members/internal/members"
	"github.com/tiennm99/export-telegram-group-members/internal/prompt"
	"github.com/tiennm99/export-telegram-group-members/internal/store"
	"github.com/tiennm99/export-telegram-group-members/internal/telegram"
)

const usage = `usage: go run ./cmd/compare-groups <group_a> <group_b>

Fetch two Telegram groups live and compare their members. No crawl is saved
to Redis; Redis only supplies the Telegram config and login session.`

// Run fetches both groups and prints who is in both, only A, and only B.
func Run(ctx context.Context, args []string, term cli.IO) error {
	if cli.WantsHelp(args) {
		fmt.Fprintln(term.Out, usage)
		return nil
	}
	if len(args) != 2 {
		return errors.New(usage)
	}
	groupIDs, err := cli.ParseGroupIDs(args)
	if err != nil {
		return fmt.Errorf("%w\n%s", err, usage)
	}

	st, err := store.OpenFromEnv(ctx)
	if err != nil {
		return err
	}
	defer st.Close()

	cfg, err := st.LoadConfig(ctx)
	if err != nil {
		return err
	}

	ask := prompt.New(term.In, term.Out)
	return telegram.Run(ctx, cfg, st.SessionStorage(), ask, func(ctx context.Context, api *tg.Client) error {
		fetcher := telegram.NewFetcher(api)
		var groups [2]telegram.Group
		for i, groupID := range groupIDs {
			group, err := fetcher.FetchGroup(ctx, groupID, func(title string) {
				fmt.Fprintf(term.Out, "fetching %s (ID: %d)\n", title, groupID)
			})
			if errors.Is(err, telegram.ErrNotGroup) {
				return fmt.Errorf("%d is not a group", groupID)
			}
			if err != nil {
				return fmt.Errorf("error accessing group %d: %w", groupID, err)
			}
			groups[i] = group
		}

		fmt.Fprintln(term.Out)
		printReport(members.Printer{Out: term.Out, Color: term.Color}, groups[0], groups[1])
		return nil
	})
}

func printReport(out members.Printer, a, b telegram.Group) {
	onlyA, onlyB, both := members.Split(a.Members, b.Members)

	out.Line(fmt.Sprintf("diff --telegram-groups %q %q", label(a), label(b)), "")
	out.Line("--- "+label(a), members.Red)
	out.Line("+++ "+label(b), members.Green)
	out.Line(fmt.Sprintf("@@ members: %d vs %d (both %d, -%d only first, +%d only second) @@",
		len(a.Members), len(b.Members), len(both), len(onlyA), len(onlyB)), members.Cyan)
	out.Line("", "")

	table := members.NewTable(onlyA, onlyB, both)
	for _, member := range both {
		out.Line(table.Row(" ", member), "")
	}
	for _, member := range onlyA {
		out.Line(table.Row("-", member), members.Red)
	}
	for _, member := range onlyB {
		out.Line(table.Row("+", member), members.Green)
	}
	if len(a.Members) == 0 && len(b.Members) == 0 {
		out.Line(" both groups have no visible members.", "")
	}
}

func label(group telegram.Group) string {
	return fmt.Sprintf("%s (%d)", group.Title, group.ID)
}
