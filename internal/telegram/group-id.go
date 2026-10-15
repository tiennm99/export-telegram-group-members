package telegram

import "errors"

// ErrNotGroup means the ID belongs to a user, not a group or channel.
var ErrNotGroup = errors.New("not a group")

// channelIDOffset is how Bot API style "marked" IDs encode a channel or
// supergroup: -100<channel_id>, i.e. -(1e12 + channel_id).
const channelIDOffset = 1_000_000_000_000

type groupKind int

const (
	basicGroup groupKind = iota
	channelGroup
)

// groupRef is a parsed marked group ID.
type groupRef struct {
	kind groupKind
	id   int64
}

// parseGroupID splits a marked ID (as configured and stored) into a basic
// group or a channel/supergroup and its raw Telegram ID.
func parseGroupID(markedID int64) (groupRef, error) {
	switch {
	case markedID <= -channelIDOffset:
		return groupRef{kind: channelGroup, id: -markedID - channelIDOffset}, nil
	case markedID < 0:
		return groupRef{kind: basicGroup, id: -markedID}, nil
	default:
		return groupRef{}, ErrNotGroup
	}
}
