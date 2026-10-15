// Package members compares member lists and formats them as diff rows. It
// stays pure: no Redis and no network.
package members

import (
	"sort"

	"github.com/tiennm99/export-telegram-group-members/internal/store"
)

// Diff returns members added in after and removed from before, matched by
// Telegram member ID and sorted by ID.
func Diff(before, after []store.Member) (added, removed []store.Member) {
	onlyBefore, onlyAfter, _ := Split(before, after)
	return onlyAfter, onlyBefore
}

// Split returns members only in a, only in b, and in both, matched by
// Telegram member ID and sorted by ID. Members in both come from b.
func Split(a, b []store.Member) (onlyA, onlyB, both []store.Member) {
	aByID := index(a)
	bByID := index(b)

	for id, member := range aByID {
		if _, ok := bByID[id]; !ok {
			onlyA = append(onlyA, member)
		}
	}
	for id, member := range bByID {
		if _, ok := aByID[id]; ok {
			both = append(both, member)
		} else {
			onlyB = append(onlyB, member)
		}
	}
	sortByID(onlyA)
	sortByID(onlyB)
	sortByID(both)
	return onlyA, onlyB, both
}

// index keys members by ID; ID 0 is never a real Telegram user.
func index(list []store.Member) map[int64]store.Member {
	byID := make(map[int64]store.Member, len(list))
	for _, member := range list {
		if member.ID != 0 {
			byID[member.ID] = member
		}
	}
	return byID
}

func sortByID(list []store.Member) {
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
}
