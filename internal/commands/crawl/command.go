// Package crawl fetches group members from Telegram and saves each group as
// one export in Redis.
package crawl

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotd/td/tg"

	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	"github.com/tiennm99/export-telegram-group-members/internal/prompt"
	"github.com/tiennm99/export-telegram-group-members/internal/store"
	"github.com/tiennm99/export-telegram-group-members/internal/telegram"
)

const usage = `usage: go run ./cmd/crawl [group_id ...]

Export members from Telegram groups. Group IDs given here replace the
configured list for this run only.`

// Run crawls the given groups, or the configured ones when none are given.
func Run(ctx context.Context, args []string, term cli.IO) error {
	if cli.WantsHelp(args) {
		fmt.Fprintln(term.Out, usage)
		return nil
	}
	// Parse before touching Redis or Telegram, so bad input fails offline.
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
	if len(groupIDs) == 0 {
		groupIDs = cfg.GroupIDs
	}

	ask := prompt.New(term.In, term.Out)
	return telegram.Run(ctx, cfg, st.SessionStorage(), ask, func(ctx context.Context, api *tg.Client) error {
		runTime := store.NewRunTime()
		saved := crawlGroups(ctx, term, telegram.NewFetcher(api), st, groupIDs, runTime)
		fmt.Fprintf(term.Out, "saved run %s: %d group(s) to Redis.\n", runTime, saved)
		return nil
	})
}

// crawlGroups saves every group it can read. One failing group is reported
// and skipped so the rest of the run still lands.
func crawlGroups(
	ctx context.Context,
	term cli.IO,
	fetcher *telegram.Fetcher,
	st *store.Store,
	groupIDs []int64,
	runTime string,
) int {
	saved := 0
	for _, groupID := range groupIDs {
		group, err := fetcher.FetchGroup(ctx, groupID, func(title string) {
			fmt.Fprintf(term.Out, "exporting %s (ID: %d)\n", title, groupID)
		})
		if errors.Is(err, telegram.ErrNotGroup) {
			fmt.Fprintf(term.Out, "skip %d because it is not a group.\n", groupID)
			continue
		}
		if err == nil {
			err = st.SaveGroupExport(ctx, store.GroupExport{
				GroupID: groupID,
				Title:   group.Title,
				Time:    runTime,
				Members: group.Members,
			})
		}
		if err != nil {
			fmt.Fprintf(term.Out, "error accessing group %d: %v\n", groupID, err)
			continue
		}
		saved++
		fmt.Fprintf(term.Out, "export %s done.\n", group.Title)
	}
	return saved
}
