package cli

import (
	"reflect"
	"testing"
)

func TestParseGroupIDsAcceptsNegativeIDs(t *testing.T) {
	got, err := ParseGroupIDs([]string{"-1001234567890", "-42"})
	if err != nil {
		t.Fatalf("ParseGroupIDs: %v", err)
	}
	if !reflect.DeepEqual(got, []int64{-1001234567890, -42}) {
		t.Fatalf("ParseGroupIDs = %v", got)
	}
}

func TestParseGroupIDsRejectsNonInteger(t *testing.T) {
	if _, err := ParseGroupIDs([]string{"not-a-group"}); err == nil {
		t.Fatal("ParseGroupIDs accepted a non-integer")
	}
}

func TestWantsHelp(t *testing.T) {
	if !WantsHelp([]string{"-100", "--help"}) || WantsHelp([]string{"-100"}) {
		t.Fatal("WantsHelp misread arguments")
	}
}
