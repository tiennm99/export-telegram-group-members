package members

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/tiennm99/export-telegram-group-members/internal/store"
)

func ids(list []store.Member) []int64 {
	out := []int64{}
	for _, member := range list {
		out = append(out, member.ID)
	}
	return out
}

func membersWithIDs(values ...int64) []store.Member {
	var list []store.Member
	for _, id := range values {
		list = append(list, store.Member{ID: id})
	}
	return list
}

func TestSplit(t *testing.T) {
	onlyA, onlyB, both := Split(membersWithIDs(3, 1, 2), membersWithIDs(4, 2, 3))

	if !reflect.DeepEqual(ids(onlyA), []int64{1}) {
		t.Errorf("onlyA = %v", ids(onlyA))
	}
	if !reflect.DeepEqual(ids(onlyB), []int64{4}) {
		t.Errorf("onlyB = %v", ids(onlyB))
	}
	if !reflect.DeepEqual(ids(both), []int64{2, 3}) {
		t.Errorf("both = %v", ids(both))
	}
}

func TestDiff(t *testing.T) {
	added, removed := Diff(membersWithIDs(1, 2, 0), membersWithIDs(2, 3))

	if !reflect.DeepEqual(ids(added), []int64{3}) || !reflect.DeepEqual(ids(removed), []int64{1}) {
		t.Fatalf("Diff added=%v removed=%v", ids(added), ids(removed))
	}
}

func TestTableRowAlignsAndTrims(t *testing.T) {
	amy := store.Member{ID: 1, Username: "amy", FirstName: "Amy", LastName: "Lê"}
	bob := store.Member{ID: 22, FirstName: "Bob"}
	table := NewTable([]store.Member{amy, bob})

	if got, want := table.Row("-", amy), "- 1   @amy      Amy Lê"; got != want {
		t.Errorf("Row(amy) = %q, want %q", got, want)
	}
	if got, want := table.Row("+", bob), "+ 22            Bob"; got != want {
		t.Errorf("Row(bob) = %q, want %q", got, want)
	}
}

func TestPrinterColorsOnlyWhenEnabled(t *testing.T) {
	var plain, colored bytes.Buffer
	Printer{Out: &plain}.Line("x", Red)
	Printer{Out: &colored, Color: true}.Line("x", Red)

	if plain.String() != "x\n" {
		t.Errorf("plain = %q", plain.String())
	}
	if colored.String() != Red+"x"+reset+"\n" {
		t.Errorf("colored = %q", colored.String())
	}
}
