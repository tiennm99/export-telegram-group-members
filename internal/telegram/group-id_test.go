package telegram

import (
	"errors"
	"testing"

	"github.com/gotd/td/tg"

	"github.com/tiennm99/export-telegram-group-members/internal/store"
)

func TestParseGroupID(t *testing.T) {
	tests := []struct {
		marked int64
		want   groupRef
	}{
		{-1001234567890, groupRef{kind: channelGroup, id: 1234567890}},
		{-1000000000001, groupRef{kind: channelGroup, id: 1}},
		{-123456, groupRef{kind: basicGroup, id: 123456}},
		{-999999999999, groupRef{kind: basicGroup, id: 999999999999}},
	}
	for _, tt := range tests {
		got, err := parseGroupID(tt.marked)
		if err != nil || got != tt.want {
			t.Errorf("parseGroupID(%d) = %+v, %v; want %+v", tt.marked, got, err, tt.want)
		}
	}

	for _, userID := range []int64{0, 777} {
		if _, err := parseGroupID(userID); !errors.Is(err, ErrNotGroup) {
			t.Errorf("parseGroupID(%d) error = %v, want ErrNotGroup", userID, err)
		}
	}
}

func TestChannelParticipantUserID(t *testing.T) {
	tests := []struct {
		name        string
		participant tg.ChannelParticipantClass
		want        int64
		ok          bool
	}{
		{"member", &tg.ChannelParticipant{UserID: 1}, 1, true},
		{"self", &tg.ChannelParticipantSelf{UserID: 2}, 2, true},
		{"admin", &tg.ChannelParticipantAdmin{UserID: 3, InviterID: 99}, 3, true},
		{"creator", &tg.ChannelParticipantCreator{UserID: 4}, 4, true},
		{"banned user", &tg.ChannelParticipantBanned{Peer: &tg.PeerUser{UserID: 5}}, 5, true},
		{"banned channel", &tg.ChannelParticipantBanned{Peer: &tg.PeerChannel{ChannelID: 6}}, 0, false},
	}
	for _, tt := range tests {
		got, ok := channelParticipantUserID(tt.participant)
		if got != tt.want || ok != tt.ok {
			t.Errorf("%s: got %d, %v; want %d, %v", tt.name, got, ok, tt.want, tt.ok)
		}
	}
}

func TestMemberFromUserAndSorting(t *testing.T) {
	users := usersByID([]tg.UserClass{
		&tg.User{ID: 9, Username: "zed", FirstName: "Z"},
		&tg.User{ID: 3, FirstName: "C", LastName: "D"},
		&tg.UserEmpty{ID: 5},
	})
	if len(users) != 2 {
		t.Fatalf("usersByID kept %d users, want 2", len(users))
	}

	got := sortedMembers(map[int64]store.Member{
		9: memberFromUser(users[9]),
		3: memberFromUser(users[3]),
	})
	if got[0].ID != 3 || got[0].LastName != "D" || got[1].Username != "zed" {
		t.Fatalf("sortedMembers = %+v", got)
	}
}
