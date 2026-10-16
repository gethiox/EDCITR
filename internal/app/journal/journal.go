package journal

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gethiox/EDCITR/internal/parser/message"
)

type MessageChannel uint32

func (c *MessageChannel) UnmarshalJSON(bytes []byte) error {
	s := string(bytes)[1 : 1+32] // truncate quotes ("), see MarshalJSON
	v, err := strconv.ParseInt(s, 2, 64)
	if err != nil {
		return err
	}
	if v < 0 || v > math.MaxUint32 {
		return errors.New("message channel out of uint8 range")
	}

	*c = MessageChannel(v)
	return nil
}

func (c *MessageChannel) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%032b"`, *c)), nil
}

const ChannelUndefined MessageChannel = 0

const (
	ChannelLocal MessageChannel = 1 << iota
	ChannelNPC
	ChannelPlayer
	ChannelWing
	ChannelVoice
	ChannelSquadron
	ChannelStarSystem
	ChannelNotification
	ChannelJournal
)

var ChannelAllList = []MessageChannel{
	ChannelLocal,
	ChannelNPC,
	ChannelPlayer,
	ChannelWing,
	ChannelVoice,
	ChannelSquadron,
	ChannelStarSystem,
	ChannelNotification,
	ChannelJournal,
}

var MessageChannelToString = map[MessageChannel]string{
	ChannelLocal:        "local",
	ChannelNPC:          "npc",
	ChannelPlayer:       "player",
	ChannelWing:         "wing",
	ChannelVoice:        "voicechat",
	ChannelSquadron:     "squadron",
	ChannelStarSystem:   "starsystem",
	ChannelNotification: "notification",
	ChannelJournal:      "=Journal=",
}

func (c *MessageChannel) ExtractList() []MessageChannel {
	var out = make([]MessageChannel, 0, len(ChannelAllList))

	for _, ch := range ChannelAllList {
		if *c&ch != 0 {
			out = append(out, ch)
		}
	}

	return out
}

var StringToMessageChannel = map[string]MessageChannel{
	"local":        ChannelLocal,
	"npc":          ChannelNPC,
	"player":       ChannelPlayer,
	"wing":         ChannelWing,
	"voicechat":    ChannelVoice,
	"squadron":     ChannelSquadron,
	"starsystem":   ChannelStarSystem,
	"notification": ChannelNotification,
}

type MessageBase struct {
	SessionCounter int
	IsLiveEvent    bool
	GeneratedAt    time.Time

	Timestamp time.Time
	Channel   MessageChannel
}

func NewMessageBase(timestamp time.Time, channel MessageChannel, session int, isLive bool) MessageBase {
	return MessageBase{
		SessionCounter: session,
		IsLiveEvent:    isLive,
		GeneratedAt:    time.Now(),
		Timestamp:      timestamp,
		Channel:        channel,
	}
}

func (m *MessageBase) IsMessageEvent() bool {
	return true
}

func (m *MessageBase) MessageChannel() MessageChannel {
	return m.Channel
}

func (m *MessageBase) IsLive() bool {
	return m.IsLiveEvent
}

func (m *MessageBase) Session() int {
	return m.SessionCounter
}

func (m *MessageBase) GeneratedTime() time.Time {
	return m.GeneratedAt
}

func (m *MessageBase) TimestampTime() time.Time {
	return m.Timestamp
}

type MessageEvent interface {
	IsMessageEvent() bool
	MessageChannel() MessageChannel
	IsLive() bool
	Session() int
	GeneratedTime() time.Time
	TimestampTime() time.Time
}

type NotificationMessage struct {
	MessageBase

	From             string
	FromLocalised    string
	Message          string
	MessageLocalised string
	DecodedMessage   message.DecodedMessage
}

type WingInviteMessage struct {
	MessageBase

	From string
}

type FriendInviteMessage struct {
	MessageBase

	From string
}

type SquadronInviteMessage struct {
	MessageBase

	Squadron string
}

type ReceivedNPCMessage struct {
	MessageBase

	From             string
	FromLocalised    string
	To               string
	Message          string
	MessageLocalised string
	DecodedMessage   message.DecodedMessage
}

type ReceivedCMDRMessage struct {
	MessageBase

	From    string
	To      string
	Message string
}

type SentMessage struct {
	MessageBase

	From    string
	To      string
	Message string

	Sent bool
}

type GenericEvent struct {
	MessageBase

	MessageBody string
}

type CommanderEvent struct {
	MessageBase

	Name string
	FID  string
}

func NewGenericMessage(body string, session int, timestamp time.Time, isLiveEvent bool) *GenericEvent {
	return &GenericEvent{
		MessageBase: MessageBase{
			SessionCounter: session,
			IsLiveEvent:    isLiveEvent,
			GeneratedAt:    time.Now(),
			Timestamp:      timestamp,
			Channel:        ChannelJournal,
		},
		MessageBody: body,
	}
}
