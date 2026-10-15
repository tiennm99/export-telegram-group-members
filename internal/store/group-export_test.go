package store

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func saveExport(t *testing.T, st *Store, groupID int64, runTime string) {
	t.Helper()
	err := st.SaveGroupExport(context.Background(), GroupExport{
		GroupID: groupID,
		Title:   "Group",
		Time:    runTime,
		Members: []Member{{ID: 1, Username: "amy"}},
	})
	if err != nil {
		t.Fatalf("SaveGroupExport: %v", err)
	}
}

func TestGroupExportRoundTrip(t *testing.T) {
	st, _ := newTestStore(t)
	saveExport(t, st, -100, "20260724120000")

	got, err := st.GetGroupExport(context.Background(), -100, "20260724120000")
	if err != nil {
		t.Fatalf("GetGroupExport: %v", err)
	}
	if got.Title != "Group" || len(got.Members) != 1 || got.Members[0].Username != "amy" {
		t.Fatalf("GetGroupExport = %+v", got)
	}
}

func TestGetGroupExportMissingOrCorrupt(t *testing.T) {
	st, mr := newTestStore(t)
	mr.Set("telegram-export:group:-100:20260724120000", "not json")

	for _, runTime := range []string{"20260724120000", "20990101000000"} {
		_, err := st.GetGroupExport(context.Background(), -100, runTime)
		if !errors.Is(err, ErrExportNotFound) {
			t.Fatalf("GetGroupExport(%s) error = %v, want ErrExportNotFound", runTime, err)
		}
	}
}

func TestGroupExportTimesScopesToGroupAndSorts(t *testing.T) {
	st, mr := newTestStore(t)
	saveExport(t, st, -100, "20260724120000")
	saveExport(t, st, -100, "20260701000000")
	saveExport(t, st, -1000, "20260801000000") // prefix of another group's ID
	mr.Set("telegram-export:group:-100:latest", "{}")

	got, err := st.GroupExportTimes(context.Background(), -100)
	if err != nil {
		t.Fatalf("GroupExportTimes: %v", err)
	}
	want := []string{"20260701000000", "20260724120000"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GroupExportTimes = %v, want %v", got, want)
	}
}

func TestLatestGroupExportsSkipsCorrupt(t *testing.T) {
	st, mr := newTestStore(t)
	saveExport(t, st, -100, "20260701000000")
	saveExport(t, st, -100, "20260702000000")
	mr.Set("telegram-export:group:-100:20260703000000", "{broken")

	got, err := st.LatestGroupExports(context.Background(), -100, 2)
	if err != nil {
		t.Fatalf("LatestGroupExports: %v", err)
	}
	if len(got) != 2 || got[0].Time != "20260701000000" || got[1].Time != "20260702000000" {
		t.Fatalf("LatestGroupExports = %+v", got)
	}
}

func TestIsRunTime(t *testing.T) {
	for value, want := range map[string]bool{
		"20260724120000":  true,
		"2026072412000":   false,
		"2026072412000a":  false,
		"202607241200000": false,
	} {
		if got := IsRunTime(value); got != want {
			t.Errorf("IsRunTime(%q) = %v, want %v", value, got, want)
		}
	}
}
