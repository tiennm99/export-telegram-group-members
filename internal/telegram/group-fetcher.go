package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
)

// dialogsPageSize is how many dialogs each lookup request fetches.
const dialogsPageSize = 100

// Fetcher resolves group IDs and fetches their members.
//
// A marked group ID alone cannot address a supergroup: Telegram also needs
// its access hash. The account's dialog list carries both, and the account
// must already be in the group to see its members, so the list is read once
// and reused for every group in a run.
type Fetcher struct {
	api    *tg.Client
	groups map[groupRef]resolvedGroup
}

type resolvedGroup struct {
	title   string
	channel *tg.InputChannel
}

// NewFetcher returns a Fetcher on an open connection.
func NewFetcher(api *tg.Client) *Fetcher {
	return &Fetcher{api: api}
}

// FetchGroup returns the group's title and members. It returns ErrNotGroup
// for a user ID.
func (f *Fetcher) FetchGroup(ctx context.Context, markedID int64, onStart func(title string)) (Group, error) {
	ref, err := parseGroupID(markedID)
	if err != nil {
		return Group{}, err
	}
	if err := f.loadDialogs(ctx); err != nil {
		return Group{}, err
	}
	resolved, ok := f.groups[ref]
	if !ok {
		return Group{}, fmt.Errorf("group %d not found in this account's chats", markedID)
	}

	onStart(resolved.title)
	group := Group{ID: markedID, Title: resolved.title}
	if ref.kind == channelGroup {
		group.Members, err = fetchChannelMembers(ctx, f.api, resolved.channel)
	} else {
		group.Members, err = fetchBasicGroupMembers(ctx, f.api, ref.id)
	}
	if err != nil {
		return Group{}, err
	}
	return group, nil
}

func (f *Fetcher) loadDialogs(ctx context.Context) error {
	if f.groups != nil {
		return nil
	}
	groups := map[groupRef]resolvedGroup{}
	iter := dialogs.NewQueryBuilder(f.api).GetDialogs().BatchSize(dialogsPageSize).Iter()
	for iter.Next(ctx) {
		elem := iter.Value()
		switch peer := elem.Peer.(type) {
		case *tg.InputPeerChat:
			if chat, ok := elem.Entities.Chat(peer.ChatID); ok {
				groups[groupRef{kind: basicGroup, id: peer.ChatID}] = resolvedGroup{title: chat.Title}
			}
		case *tg.InputPeerChannel:
			if channel, ok := elem.Entities.Channel(peer.ChannelID); ok {
				groups[groupRef{kind: channelGroup, id: peer.ChannelID}] = resolvedGroup{
					title:   channel.Title,
					channel: &tg.InputChannel{ChannelID: peer.ChannelID, AccessHash: peer.AccessHash},
				}
			}
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("list chats: %w", err)
	}
	f.groups = groups
	return nil
}
