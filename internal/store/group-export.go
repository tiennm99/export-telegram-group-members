package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RunTimeLayout is the yyyymmddhhmmss crawl timestamp shared by every group
// in one run. Fixed width, so timestamps sort chronologically as strings.
const RunTimeLayout = "20060102150405"

// Member is one group member as stored in an export.
type Member struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// GroupExport is one group's members for one run. Each export is a single
// self-contained key, so deleting any key never corrupts another.
type GroupExport struct {
	GroupID int64    `json:"group_id"`
	Title   string   `json:"title"`
	Time    string   `json:"time"`
	Members []Member `json:"members"`
}

// NewRunTime returns the current time as a run timestamp.
func NewRunTime() string {
	return time.Now().Format(RunTimeLayout)
}

// IsRunTime reports whether value has the yyyymmddhhmmss shape.
func IsRunTime(value string) bool {
	if len(value) != len(RunTimeLayout) {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func groupExportKey(groupID int64, runTime string) string {
	return key("group", strconv.FormatInt(groupID, 10), runTime)
}

// SaveGroupExport writes one group's export with a single atomic SET.
func (s *Store) SaveGroupExport(ctx context.Context, record GroupExport) error {
	raw, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode export: %w", err)
	}
	if err := s.rdb.Set(ctx, groupExportKey(record.GroupID, record.Time), raw, 0).Err(); err != nil {
		return fmt.Errorf("save export for group %d: %w", record.GroupID, err)
	}
	return nil
}

// GroupExportTimes returns one group's run timestamps, oldest first.
func (s *Store) GroupExportTimes(ctx context.Context, groupID int64) ([]string, error) {
	pattern := groupExportKey(groupID, "*")
	prefix := strings.TrimSuffix(pattern, "*")

	var times []string
	iter := s.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		runTime := strings.TrimPrefix(iter.Val(), prefix)
		if IsRunTime(runTime) {
			times = append(times, runTime)
		}
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("list exports for group %d: %w", groupID, err)
	}
	sort.Strings(times)
	return times, nil
}

// ErrExportNotFound means no readable export exists at that group and time.
var ErrExportNotFound = errors.New("export not found")

// GetGroupExport returns one export, or ErrExportNotFound when it is missing
// or not valid JSON.
func (s *Store) GetGroupExport(ctx context.Context, groupID int64, runTime string) (GroupExport, error) {
	raw, err := s.rdb.Get(ctx, groupExportKey(groupID, runTime)).Bytes()
	if errors.Is(err, redis.Nil) {
		return GroupExport{}, ErrExportNotFound
	}
	if err != nil {
		return GroupExport{}, fmt.Errorf("read export for group %d at %s: %w", groupID, runTime, err)
	}

	var record GroupExport
	if err := json.Unmarshal(raw, &record); err != nil {
		return GroupExport{}, ErrExportNotFound
	}
	return record, nil
}

// LatestGroupExports returns up to n readable exports for a group, oldest
// first. Corrupt values and keys deleted mid-scan are skipped.
func (s *Store) LatestGroupExports(ctx context.Context, groupID int64, n int) ([]GroupExport, error) {
	times, err := s.GroupExportTimes(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var records []GroupExport
	for i := len(times) - 1; i >= 0 && len(records) < n; i-- {
		record, err := s.GetGroupExport(ctx, groupID, times[i])
		if errors.Is(err, ErrExportNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		records = append([]GroupExport{record}, records...)
	}
	return records, nil
}
