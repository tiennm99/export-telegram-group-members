package telegram

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/gotd/td/tg"

	"github.com/tiennm99/export-telegram-group-members/internal/store"
)

// participantsPageSize is the largest page channels.getParticipants allows.
const participantsPageSize = 200

// Group is one fetched group with its members sorted by ID.
type Group struct {
	ID      int64
	Title   string
	Members []store.Member
}

// fetchChannelMembers pages through a channel or supergroup's members.
func fetchChannelMembers(ctx context.Context, api *tg.Client, channel *tg.InputChannel) ([]store.Member, error) {
	byID := map[int64]store.Member{}
	for offset := 0; ; {
		resp, err := withFloodWait(ctx, func() (tg.ChannelsChannelParticipantsClass, error) {
			return api.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
				Channel: channel,
				Filter:  &tg.ChannelParticipantsSearch{},
				Offset:  offset,
				Limit:   participantsPageSize,
			})
		})
		if err != nil {
			return nil, fmt.Errorf("get participants: %w", err)
		}
		page, ok := resp.(*tg.ChannelsChannelParticipants)
		if !ok || len(page.Participants) == 0 {
			break
		}
		users := usersByID(page.Users)
		for _, participant := range page.Participants {
			if userID, ok := channelParticipantUserID(participant); ok {
				if user, ok := users[userID]; ok {
					byID[userID] = memberFromUser(user)
				}
			}
		}
		offset += len(page.Participants)
	}
	return sortedMembers(byID), nil
}

// fetchBasicGroupMembers reads a basic group's members from its full info.
func fetchBasicGroupMembers(ctx context.Context, api *tg.Client, chatID int64) ([]store.Member, error) {
	full, err := withFloodWait(ctx, func() (*tg.MessagesChatFull, error) {
		return api.MessagesGetFullChat(ctx, chatID)
	})
	if err != nil {
		return nil, fmt.Errorf("get full chat: %w", err)
	}
	chatFull, ok := full.FullChat.(*tg.ChatFull)
	if !ok {
		return nil, errors.New("unexpected full chat type")
	}
	participants, ok := chatFull.Participants.(*tg.ChatParticipants)
	if !ok {
		return nil, errors.New("member list is hidden")
	}

	users := usersByID(full.Users)
	byID := map[int64]store.Member{}
	for _, participant := range participants.Participants {
		if user, ok := users[participant.GetUserID()]; ok {
			byID[user.ID] = memberFromUser(user)
		}
	}
	return sortedMembers(byID), nil
}

// channelParticipantUserID returns the member's user ID. Admin entries also
// reference inviters and promoters, so users come from participants only.
func channelParticipantUserID(participant tg.ChannelParticipantClass) (int64, bool) {
	switch p := participant.(type) {
	case *tg.ChannelParticipant:
		return p.UserID, true
	case *tg.ChannelParticipantSelf:
		return p.UserID, true
	case *tg.ChannelParticipantAdmin:
		return p.UserID, true
	case *tg.ChannelParticipantCreator:
		return p.UserID, true
	case *tg.ChannelParticipantBanned:
		if peer, ok := p.Peer.(*tg.PeerUser); ok {
			return peer.UserID, true
		}
	case *tg.ChannelParticipantLeft:
		if peer, ok := p.Peer.(*tg.PeerUser); ok {
			return peer.UserID, true
		}
	}
	return 0, false
}

func usersByID(users []tg.UserClass) map[int64]*tg.User {
	byID := make(map[int64]*tg.User, len(users))
	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			byID[user.ID] = user
		}
	}
	return byID
}

func memberFromUser(user *tg.User) store.Member {
	return store.Member{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}

func sortedMembers(byID map[int64]store.Member) []store.Member {
	list := make([]store.Member, 0, len(byID))
	for _, member := range byID {
		list = append(list, member)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list
}
