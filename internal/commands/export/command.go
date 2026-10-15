// Package export writes one saved group crawl from Redis to a CSV file.
package export

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tiennm99/export-telegram-group-members/internal/commands/cli"
	"github.com/tiennm99/export-telegram-group-members/internal/store"
)

const usage = `usage: go run ./cmd/export [group_id [timecrawl]]

Export a Redis-backed Telegram group crawl to CSV. Without timecrawl
(yyyymmddhhmmss), the latest crawl is used; without a group, the first
configured group.`

// outputDir is where CSV files land; a var so tests can redirect it.
var outputDir = "output"

// Run exports one crawl to output/<group-id>-<yyyymmddhhmmss>.csv.
func Run(ctx context.Context, args []string, term cli.IO) error {
	if cli.WantsHelp(args) {
		fmt.Fprintln(term.Out, usage)
		return nil
	}
	if len(args) > 2 {
		return errors.New(usage)
	}
	var (
		groupID    int64
		hasGroupID bool
		runTime    string
		err        error
	)
	if len(args) >= 1 {
		if groupID, err = cli.ParseGroupID(args[0]); err != nil {
			return fmt.Errorf("%w\n%s", err, usage)
		}
		hasGroupID = true
	}
	if len(args) == 2 {
		runTime = args[1]
		if err := validateRunTime(runTime); err != nil {
			return fmt.Errorf("%w\n%s", err, usage)
		}
	}

	st, err := store.OpenFromEnv(ctx)
	if err != nil {
		return err
	}
	defer st.Close()

	record, groupID, runTime, err := loadRecord(ctx, st, groupID, hasGroupID, runTime)
	if err != nil {
		return err
	}

	rows, err := validateRecord(record, groupID, runTime)
	if err == nil {
		var path string
		if path, err = writeCSV(outputDir, groupID, runTime, rows); err == nil {
			fmt.Fprintf(term.Out, "exported %d members to %s\n", len(rows), path)
			return nil
		}
	}
	return fmt.Errorf("could not export group %d at %s: %w", groupID, runTime, err)
}

// loadRecord resolves the default group and latest time, then reads that
// export. It returns the resolved group ID and time with the record.
func loadRecord(ctx context.Context, st *store.Store, groupID int64, hasGroupID bool, runTime string) (store.GroupExport, int64, string, error) {
	if !hasGroupID {
		cfg, err := st.LoadConfig(ctx)
		if err != nil {
			return store.GroupExport{}, 0, "", fmt.Errorf("could not read export configuration or history: %w", err)
		}
		groupID = cfg.GroupIDs[0]
	}
	if runTime == "" {
		times, err := st.GroupExportTimes(ctx, groupID)
		if err != nil {
			return store.GroupExport{}, 0, "", fmt.Errorf("could not read export configuration or history: %w", err)
		}
		if len(times) == 0 {
			return store.GroupExport{}, 0, "", fmt.Errorf("no exports found for group %d", groupID)
		}
		runTime = times[len(times)-1]
	}

	record, err := st.GetGroupExport(ctx, groupID, runTime)
	if errors.Is(err, store.ErrExportNotFound) {
		return store.GroupExport{}, 0, "", fmt.Errorf("export not found or invalid for group %d at %s", groupID, runTime)
	}
	if err != nil {
		return store.GroupExport{}, 0, "", fmt.Errorf("could not read export configuration or history: %w", err)
	}
	return record, groupID, runTime, nil
}

func validateRunTime(value string) error {
	if !store.IsRunTime(value) {
		return errors.New("timecrawl must use yyyymmddhhmmss format")
	}
	if _, err := time.Parse(store.RunTimeLayout, value); err != nil {
		return errors.New("timecrawl must be a valid date and time")
	}
	return nil
}
