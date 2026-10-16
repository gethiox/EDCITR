package tasks

import (
	"context"
	"encoding/json/v2"

	"github.com/gethiox/EDCITR/internal/app/journal"
	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/app/tasks/generate"
	"github.com/gethiox/EDCITR/internal/config"
	parser2 "github.com/gethiox/EDCITR/internal/parser"
	"github.com/gethiox/EDCITR/internal/storage"
	"go.uber.org/zap"
)

type JSONPayload struct {
	From             string `json:"From"`
	FromLocalised    string `json:"FromLocalised"`
	Message          string `json:"Message"`
	MessageLocalised string `json:"MessageLocalised"`
}

type GenerateEvents struct {
	JsonStore *storage.JsonStore
	Logger    *zap.Logger
	Config    *config.EDCITRConfig
	State     *state.DataState
}

func (g *GenerateEvents) Run(ctx context.Context) {
	params := generate.GenericGenerator[parser2.UnsupportedEvent]{
		Logger:       g.Logger.Named("UnsupportedEvents"),
		ArchiveDir:   g.Config.Parser.ArchiveDirectory,
		StatePtr:     g.State,
		LogProducer:  parser2.ReadLogsUnsupported,
		FileProducer: g.JsonStore.NewEventFile,
		Evaluator:    g.evaluate,
	}

	params.Run(ctx)
}

func (g *GenerateEvents) evaluate(ev parser2.UnsupportedEvent) (bool, string, []byte) {
	return true, ev.Event, []byte(ev.Raw) // parser.ReadLogsUnsupported always returns desired event
}

type GenerateNPCMessages struct {
	JsonStore *storage.JsonStore
	Logger    *zap.Logger
	Config    *config.EDCITRConfig
	State     *state.DataState
}

func (g *GenerateNPCMessages) Run(ctx context.Context) {
	params := generate.GenericGenerator[parser2.JournalEvent]{
		Logger:       g.Logger.Named("UnsupportedNPCMessages"),
		ArchiveDir:   g.Config.Parser.ArchiveDirectory,
		StatePtr:     g.State,
		LogProducer:  parser2.ReadLogs,
		FileProducer: g.JsonStore.NewNPCMessageFile,
		Evaluator:    g.evaluate,
	}

	params.Run(ctx)
}

func (g *GenerateNPCMessages) evaluate(event parser2.JournalEvent) (bool, string, []byte) {
	msg, ok := event.(*parser2.ReceiveText)
	if !ok {
		return false, "", nil // skip if not ReceiveText
	}

	handledMsg, err := handleReceivedText(msg, "", 0, false)
	if err != nil {
		g.Logger.Error("Failed to handle received text", zap.Error(err))
		return false, "", nil
	}

	npcMessage, ok := handledMsg.(*journal.ReceivedNPCMessage)
	if !ok {
		return false, "", nil
	}

	who := npcMessage.DecodedMessage.What
	subType := npcMessage.DecodedMessage.SubWhat

	if who == "" {
		g.Logger.Error("who is empty", zap.Any("message", npcMessage))
		return false, "", nil
	}

	if g.Config.Filters.NpcMessages.Exists(who, subType) {
		return false, "", nil
	}

	identifier := npcMessage.DecodedMessage.OriginalIdentifier()

	payload := JSONPayload{
		From:             npcMessage.From,
		FromLocalised:    npcMessage.FromLocalised,
		Message:          npcMessage.Message,
		MessageLocalised: npcMessage.MessageLocalised,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		g.Logger.Error("Failed to marshal npc Message", zap.Error(err))
		return false, "", nil
	}

	return true, identifier, payloadBytes
}

type GenerateNotifications struct {
	JsonStore *storage.JsonStore
	Logger    *zap.Logger
	Config    *config.EDCITRConfig
	State     *state.DataState
}

func (g *GenerateNotifications) Run(ctx context.Context) {
	params := generate.GenericGenerator[parser2.JournalEvent]{
		Logger:       g.Logger.Named("UnsupportedNotifications"),
		ArchiveDir:   g.Config.Parser.ArchiveDirectory,
		StatePtr:     g.State,
		LogProducer:  parser2.ReadLogs,
		FileProducer: g.JsonStore.NewNotificationFile,
		Evaluator:    g.evaluate,
	}

	params.Run(ctx)
}

func (g *GenerateNotifications) evaluate(event parser2.JournalEvent) (bool, string, []byte) {
	msg, ok := event.(*parser2.ReceiveText)
	if !ok {
		return false, "", nil // skip if not ReceiveText
	}

	handledMsg, err := handleReceivedText(msg, "", 0, false)
	if err != nil {
		g.Logger.Error("Failed to handle received text", zap.Error(err))
		return false, "", nil
	}

	notification, ok := handledMsg.(*journal.NotificationMessage)
	if !ok {
		return false, "", nil
	}

	who := notification.DecodedMessage.What
	subType := notification.DecodedMessage.SubWhat

	if who == "" {
		g.Logger.Error("who is empty", zap.Any("notification", notification))
		return false, "", nil
	}

	if g.Config.Filters.Notifications.Exists(who, subType) {
		return false, "", nil
	}

	identifier := notification.DecodedMessage.OriginalIdentifier()

	payload := JSONPayload{
		Message:          notification.Message,
		MessageLocalised: notification.MessageLocalised,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		g.Logger.Error("Failed to marshal notification", zap.Error(err))
		return false, "", nil
	}

	return true, identifier, payloadBytes
}
