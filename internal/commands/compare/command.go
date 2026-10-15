// Package compare shows member changes for one group between two saved
// crawls.
package compare

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	"github.com/tiennm99/export-telegram-group-members/internal/members"
	"github.com/tiennm99/export-telegram-group-members/internal/store"
)

const usage = `usage: go run ./cmd/compare [group_id [time1 time2]]

Compare member changes between two crawls for one group. Without times, the
latest two crawls are used; without a group, the first configured group.`

type request struct {
	groupID    int64
	hasGroupID bool
	times      []string
}

// Run compares two crawls of one group.
func Run(ctx context.Context, args []string, term cli.IO) error {
	if cli.WantsHelp(args) {
		fmt.Fprintln(term.Out, usage)
		return nil
	}
	req, err := parseArgs(args)
	if err != nil {
		return fmt.Errorf("%w\n%s", err, usage)
	}

	st, err := store.OpenFromEnv(ctx)
	if err != nil {
		return err
	}
	defer st.Close()

	if !req.hasGroupID {
		cfg, err := st.LoadConfig(ctx)
		if err != nil {
			return err
		}
		req.groupID = cfg.GroupIDs[0]
	}

	before, after, err := loadPair(ctx, st, req)
	if err != nil {
		return err
	}
	added, removed := members.Diff(before.Members, after.Members)
	printReport(members.Printer{Out: term.Out, Color: term.Color}, req.groupID, before, after, added, removed)
	return nil
}

func parseArgs(args []string) (request, error) {
	switch len(args) {
	case 0:
		return request{}, nil
	case 1, 3:
		groupID, err := cli.ParseGroupID(args[0])
		if err != nil {
			return request{}, err
		}
		return request{groupID: groupID, hasGroupID: true, times: args[1:]}, nil
	default:
		return request{}, errors.New("provide both time1 and time2, or omit both to compare latest two crawls")
	}
}

func loadPair(ctx context.Context, st *store.Store, req request) (store.GroupExport, store.GroupExport, error) {
	if len(req.times) == 2 {
		var pair [2]store.GroupExport
		for i, runTime := range req.times {
			record, err := st.GetGroupExport(ctx, req.groupID, runTime)
			if errors.Is(err, store.ErrExportNotFound) {
				return store.GroupExport{}, store.GroupExport{}, fmt.Errorf("export not found for group %d at %s", req.groupID, runTime)
			}
			if err != nil {
				return store.GroupExport{}, store.GroupExport{}, err
			}
			pair[i] = record
		}
		return pair[0], pair[1], nil
	}

	records, err := st.LatestGroupExports(ctx, req.groupID, 2)
	if err != nil {
		return store.GroupExport{}, store.GroupExport{}, err
	}
	if len(records) < 2 {
		return store.GroupExport{}, store.GroupExport{}, fmt.Errorf("need at least 2 exports for group %d", req.groupID)
	}
	return records[0], records[1], nil
}

func printReport(out members.Printer, groupID int64, before, after store.GroupExport, added, removed []store.Member) {
	title := after.Title
	if title == "" {
		title = before.Title
	}
	label := strconv.FormatInt(groupID, 10)
	if title != "" {
		label = fmt.Sprintf("%s (%d)", title, groupID)
	}

	out.Line(fmt.Sprintf("diff --telegram-group %q", label), "")
	out.Line("--- crawl/"+before.Time, members.Red)
	out.Line("+++ crawl/"+after.Time, members.Green)
	out.Line(fmt.Sprintf("@@ members: %d -> %d (+%d -%d) @@",
		len(before.Members), len(after.Members), len(added), len(removed)), members.Cyan)
	out.Line("", "")

	table := members.NewTable(removed, added)
	for _, member := range removed {
		out.Line(table.Row("-", member), members.Red)
	}
	for _, member := range added {
		out.Line(table.Row("+", member), members.Green)
	}
	if len(added) == 0 && len(removed) == 0 {
		out.Line(" no membership changes.", "")
	}
}
