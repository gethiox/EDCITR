package tasks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/app/journal"
	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/config"
	"github.com/gethiox/EDCITR/internal/inara"
	"github.com/gethiox/EDCITR/internal/parser"
	"github.com/gethiox/EDCITR/internal/parser/message"
	"github.com/gethiox/EDCITR/internal/storage"
	"github.com/gethiox/EDCITR/internal/utils"
	"github.com/gethiox/rotational-velocidensity-buffer"
	"github.com/orcaman/concurrent-map/v2"
	"go.uber.org/zap"
)

const PProfDirectory = "pprof"

type JournalProcessingWorker struct {
	Logger          *zap.Logger
	State           *state.TheHolyGlobalState
	StateUI         *state.UIState
	Config          *config.EDCITRConfig
	ProcessedConfig *config.EDCITRConfig
	CacheStore      *storage.Cache
	InaraClient     *inara.Client
	PFPDownloader   *inara.PFPDownloader

	maxSounds       int
	profileCatchup  bool
	unhandledEvents cmap.ConcurrentMap[string, bool]
}

func NewJournalProcessingJob(
	logger *zap.Logger,
	state *state.TheHolyGlobalState,
	stateUI *state.UIState,
	cfg, processedCfg *config.EDCITRConfig,
	cacheStore *storage.Cache,
	inaraClient *inara.Client,
	pfpDownloader *inara.PFPDownloader,
	maxSounds int,
	profileCatchup bool,
) *JournalProcessingWorker {
	return &JournalProcessingWorker{
		Logger:          logger.Named("Journal"),
		State:           state,
		StateUI:         stateUI,
		Config:          cfg,
		ProcessedConfig: processedCfg,
		CacheStore:      cacheStore,
		InaraClient:     inaraClient,
		PFPDownloader:   pfpDownloader,
		maxSounds:       maxSounds,
		profileCatchup:  profileCatchup,
		unhandledEvents: cmap.New[bool](),
	}
}

func (j *JournalProcessingWorker) Run(ctx context.Context) {
	*j.ProcessedConfig = *j.Config.DeepCopy()
	j.Logger.Info("Starting processing Journal", zap.Int("preload_count", j.ProcessedConfig.Parser.PreloadNLogfiles))

	for _, vp := range j.State.ViewportBuffers {
		vp.ResetBuffer()
		vp.Scroll.ScrollPos = 0
	}
	j.State.SessionCounter.Store(0)
	if j.ProcessedConfig.Stats.AlwaysResetSession {
		j.State.Stats.ResetAll()
	}

	messagesChan := make(chan journal.MessageEvent, 64)
	notifications := make(chan journal.MessageChannel, 16)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(messagesChan)
		j.processEventsAndProduceMessages(ctx, j.ProcessedConfig.Parser.PreloadNLogfiles, j.ProcessedConfig.Parser.ArchiveDirectory, messagesChan)
		j.Logger.Info("processEventsAndProduceMessages finished")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(notifications)
		j.processMessages(messagesChan, notifications)
		j.Logger.Info("processMessages finished")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		j.processNotifications(notifications)
		j.Logger.Info("processNotifications finished")
	}()

	wg.Wait()
	j.Logger.Info("Journal processing shutdown")
}

func handleReceivedText(ev *parser.ReceiveText, currentCMDR string, sessionCounter int, isLiveEvent bool) (journal.MessageEvent, error) {
	channel, ok := journal.StringToMessageChannel[ev.Channel]
	if !ok {
		return nil, fmt.Errorf("unsupported channel '%s'", ev.Channel)
	}

	switch ev.Channel {
	case "npc":
		//if ev.FromLocalised != "" && ev.FromLocalised != "System" { // some NPC, "System" is sent for emote events

		//zaputils.Logger.Named("haha").Info("haha", zap.Any("ev", ev))

		if ev.FromLocalised == "" || ev.FromLocalised == "System" || ev.FromLocalised == "System Colonisation Ship" {

			return &journal.NotificationMessage{
				MessageBase:      journal.NewMessageBase(ev.Timestamp, journal.ChannelNotification, sessionCounter, isLiveEvent),
				From:             utils.UniqueString(ev.From),
				FromLocalised:    utils.UniqueString(ev.FromLocalised),
				Message:          utils.UniqueString(ev.Message),
				MessageLocalised: utils.UniqueString(ev.MessageLocalised),
				DecodedMessage:   message.DecodeMessage(ev.Message),
			}, nil
		} else { // some local notification

			return &journal.ReceivedNPCMessage{
				MessageBase:      journal.NewMessageBase(ev.Timestamp, channel, sessionCounter, isLiveEvent),
				From:             utils.UniqueString(ev.From),
				FromLocalised:    utils.UniqueString(ev.FromLocalised),
				To:               utils.UniqueString(currentCMDR),
				Message:          utils.UniqueString(ev.Message),
				MessageLocalised: utils.UniqueString(ev.MessageLocalised),
				DecodedMessage:   message.DecodeMessage(ev.Message),
			}, nil
		}
	default: // CMDR message
		return &journal.ReceivedCMDRMessage{
			MessageBase: journal.NewMessageBase(ev.Timestamp, channel, sessionCounter, isLiveEvent),
			From:        utils.UniqueString(ev.From),
			To:          utils.UniqueString(currentCMDR),
			Message:     ev.Message,
		}, nil
		//if j.State.Inara.CMDRNamesStreamActive.Load() && j.State.ProcessingLiveEvents.Load() { // TODO: TODO
		//	j.State.Inara.CMDRNamesStream <- ev.From
		//}
	}
}

func (j *JournalProcessingWorker) ProcessEvent(event parser.JournalEvent, events chan<- journal.MessageEvent, sessionCounter int, isLiveEvent bool) {
	switch ev := event.(type) {
	case *parser.Commander:
		j.State.CurrentThings.CMDR.Store(ev.Name)
		j.State.CurrentThings.FID.Store(ev.FID)
		events <- &journal.CommanderEvent{
			Name: ev.Name,
			FID:  ev.FID,
		}

		//if j.State.ProcessingLiveEvents.Load() { // TODO: TODO
		//	if j.State.Inara.CMDRNamesStreamActive.Load() {
		//		j.State.Inara.CMDRNamesStream <- ev.Name
		//	}
		//
		//	if j.Config.InaraSettings.Enabled && !j.InaraClient.Ready() {
		//		err := j.InaraClient.SetHeader(
		//			inara.NewHeader(version.Release, state.APIAppName, j.Config.InaraSettings.Token, ev.Name, ev.FID),
		//		)
		//		if err != nil {
		//			j.Logger.Error("Failed to activate Inara Client", zap.Error(err))
		//		}
		//	}
		//
		//	if j.Config.InaraSettings.Enabled && !j.PFPDownloader.Ready() {
		//		err := j.PFPDownloader.SetHeader(state.APIAppName, version.Release)
		//		if err != nil {
		//			j.Logger.Error("Failed to activate Image Downloader", zap.Error(err))
		//		}
		//	}
		//
		//}
	case *parser.ReceiveText:
		recv, err := handleReceivedText(ev, j.State.CurrentThings.CMDR.Load(), sessionCounter, isLiveEvent)
		if err != nil {
			j.Logger.Error("todo", zap.Error(err), zap.String("channel", ev.Channel))
			return
		}
		events <- recv
	case *parser.SendText:
		toPlayer := ""
		channel, ok := journal.StringToMessageChannel[ev.To]
		if !ok {
			channel = journal.ChannelPlayer
			toPlayer = ev.To
		}

		events <- &journal.SentMessage{
			MessageBase: journal.NewMessageBase(ev.Timestamp, channel, sessionCounter, isLiveEvent),
			From:        utils.UniqueString(j.State.CurrentThings.CMDR.Load()),
			To:          utils.UniqueString(toPlayer),
			Sent:        ev.Sent,
			Message:     ev.Message,
		}
		if j.State.Inara.CMDRNamesStreamActive.Load() && j.State.ProcessingLiveEvents.Load() && toPlayer != "" {
			j.State.Inara.CMDRNamesStream <- toPlayer
		}
	case *parser.Friends:
		switch ev.Status {
		case "Lost":
			j.State.FriendList.Remove(ev.Name)
		case "Offline", "Added":
			j.State.FriendList.Set(ev.Name, false)
		case "Online":
			j.State.FriendList.Set(ev.Name, true)
		case "Requested":
			events <- &journal.FriendInviteMessage{
				MessageBase: journal.NewMessageBase(ev.Timestamp, journal.ChannelNotification, sessionCounter, isLiveEvent),
				From:        utils.UniqueString(ev.Name),
			}
		}
	case *parser.WingInvite:
		events <- &journal.WingInviteMessage{
			MessageBase: journal.NewMessageBase(ev.Timestamp, journal.ChannelNotification, sessionCounter, isLiveEvent),
			From:        utils.UniqueString(ev.Name),
		}
	case *parser.InvitedToSquadron:
		events <- &journal.SquadronInviteMessage{
			MessageBase: journal.NewMessageBase(ev.Timestamp, journal.ChannelNotification, sessionCounter, isLiveEvent),
			Squadron:    utils.UniqueString(ev.SquadronName),
		}
	case *parser.Music:
		j.State.CurrentThings.Music.Store(ev.MusicTrack)
		switch ev.MusicTrack { // TODO: consolidate logic with *Shutdown case
		case "MainMenu":
			startedAt := j.State.CurrentThings.GameLoadedAt.Load()
			switch j.State.CurrentThings.Mode.Load() {
			case "Open":
				j.State.Stats.PlaytimeOpen.Add(ev.Timestamp.Sub(startedAt))
			case "Solo":
				j.State.Stats.PlaytimeSolo.Add(ev.Timestamp.Sub(startedAt))
			case "Group":
				j.State.Stats.PlaytimeGroup.Add(ev.Timestamp.Sub(startedAt))
			}
			j.State.CurrentThings.GameLoadedAt.Store(time.Time{})
			j.State.CurrentThings.Mode.Store("")
			j.resetLocation()
		}
	case *parser.LoadGame:
		j.State.CurrentThings.GameLoadedAt.Store(ev.Timestamp)
		j.State.CurrentThings.Mode.Store(ev.GameMode)
		switch ev.GameMode {
		case "Open":
			j.State.Stats.SessionsOpen.Add(1)
		case "Solo":
			j.State.Stats.SessionsSolo.Add(1)
		case "Group":
			j.State.Stats.SessionsGroup.Add(1)
		}

		body := fmt.Sprintf("Load game: %s", ev.GameMode)
		if ev.Group != "" {
			body += fmt.Sprintf(" (%s)", ev.Group)
		}
		events <- journal.NewGenericMessage(body, sessionCounter, ev.Timestamp, isLiveEvent)
	case *parser.Shutdown:
		if j.State.CurrentThings.Music.Load() != "MainMenu" {
			startedAt := j.State.CurrentThings.GameLoadedAt.Load()
			switch j.State.CurrentThings.Mode.Load() {
			case "Open":
				j.State.Stats.PlaytimeOpen.Add(ev.Timestamp.Sub(startedAt))
			case "Solo":
				j.State.Stats.PlaytimeSolo.Add(ev.Timestamp.Sub(startedAt))
			case "Group":
				j.State.Stats.PlaytimeGroup.Add(ev.Timestamp.Sub(startedAt))
			}
			j.State.CurrentThings.GameLoadedAt.Store(time.Time{})
			j.State.CurrentThings.Mode.Store("")
			j.resetLocation()
		}
		events <- journal.NewGenericMessage("Shutdown", sessionCounter, ev.Timestamp, isLiveEvent)
		j.resetSession()
	case *parser.Fileheader:
		body := fmt.Sprintf("Header: version %s, build %s", ev.Gameversion, ev.Build)
		events <- journal.NewGenericMessage(body, sessionCounter, ev.Timestamp, isLiveEvent)
	case *parser.Died:
		j.State.Stats.Deaths.Add(1)
		var body string
		if strings.HasPrefix(ev.KillerName, "Cmdr ") {
			cmdrName := ev.KillerName[5:]
			j.State.EnemyList.Set(cmdrName, true)
			j.State.Stats.DeathsByPlayer.Add(1)
			body = fmt.Sprintf("Died: CMDR %s", cmdrName)
		} else {
			body = fmt.Sprintf("Died: %s", ev.KillerNameLocalised)
		}
		events <- journal.NewGenericMessage(body, sessionCounter, ev.Timestamp, isLiveEvent)
		j.resetLocation()
	case *parser.FSDJump:
		j.State.Stats.Jumps.Add(1)
		body := fmt.Sprintf("Arrived: %s", ev.StarSystem)
		j.setLocation(&parser.Location{
			StarSystem: ev.StarSystem,
			Population: ev.Population,
			Body:       ev.Body,
			BodyType:   ev.BodyType,
		})
		events <- journal.NewGenericMessage(body, sessionCounter, ev.Timestamp, isLiveEvent)
	case *parser.FSSSignalDiscovered:
		j.State.Stats.SignalsDetected.Add(1)
	case *parser.NpcCrewPaidWage:
		j.State.Stats.CrewWage.Add(ev.Amount)
	case *parser.Docked:
		j.State.Stats.Dockings.Add(1)
	case *parser.MarketBuy:
		j.State.Stats.CreditsNetIncome.Sub(ev.TotalCost)
		j.State.Stats.CreditsExpenses.Sub(ev.TotalCost)
		j.State.Stats.CommodityNet.Sub(ev.TotalCost)
		j.State.Stats.CommodityBuy.Sub(ev.TotalCost)
	case *parser.RefuelAll:
		j.State.Stats.CreditsNetIncome.Sub(ev.Cost)
		j.State.Stats.CreditsExpenses.Sub(ev.Cost)
	case *parser.ModuleSellRemote:
		j.State.Stats.CreditsNetIncome.Add(ev.SellPrice)
		j.State.Stats.CreditsIncome.Add(ev.SellPrice)
	case *parser.MarketSell:
		j.State.Stats.CreditsNetIncome.Add(ev.TotalSale)
		j.State.Stats.CreditsIncome.Add(ev.TotalSale)
		j.State.Stats.CommodityNet.Add(ev.TotalSale)
		j.State.Stats.CommoditySell.Add(ev.TotalSale)
	case *parser.RepairAll:
		j.State.Stats.CreditsNetIncome.Sub(ev.Cost)
		j.State.Stats.CreditsExpenses.Sub(ev.Cost)
	case *parser.ModuleBuy:
		j.State.Stats.CreditsNetIncome.Sub(ev.BuyPrice)
		j.State.Stats.CreditsExpenses.Sub(ev.BuyPrice)
	case *parser.RedeemVoucher:
		j.State.Stats.CreditsNetIncome.Add(ev.Amount)
		j.State.Stats.CreditsIncome.Add(ev.Amount)
	case *parser.BuyDrones:
		j.State.Stats.CreditsNetIncome.Sub(ev.TotalCost)
		j.State.Stats.CreditsExpenses.Sub(ev.TotalCost)
	case *parser.BuyAmmo:
		j.State.Stats.CreditsNetIncome.Sub(ev.Cost)
		j.State.Stats.CreditsExpenses.Sub(ev.Cost)
	case *parser.Repair:
		j.State.Stats.CreditsNetIncome.Sub(ev.Cost)
		j.State.Stats.CreditsExpenses.Sub(ev.Cost)
	case *parser.PayFines:
		j.State.Stats.CreditsNetIncome.Sub(ev.Amount)
		j.State.Stats.CreditsExpenses.Sub(ev.Amount)
	case *parser.ModuleSell:
		j.State.Stats.CreditsNetIncome.Add(ev.SellPrice)
		j.State.Stats.CreditsIncome.Add(ev.SellPrice)
	case *parser.PayBounties:
		j.State.Stats.CreditsNetIncome.Sub(ev.Amount)
		j.State.Stats.CreditsExpenses.Sub(ev.Amount)

	case *parser.SearchAndRescue:
		j.State.Stats.CreditsNetIncome.Add(ev.Reward)
		j.State.Stats.CreditsIncome.Add(ev.Reward)
	case *parser.ShipyardSell:
		j.State.Stats.CreditsNetIncome.Add(ev.ShipPrice)
		j.State.Stats.CreditsIncome.Add(ev.ShipPrice)
	case *parser.SellDrones:
		j.State.Stats.CreditsNetIncome.Add(ev.TotalSale)
		j.State.Stats.CreditsIncome.Add(ev.TotalSale)
	case *parser.CommunityGoalReward:
		j.State.Stats.CreditsNetIncome.Add(ev.Reward)
		j.State.Stats.CreditsIncome.Add(ev.Reward)
	case *parser.SellExplorationData:
		j.State.Stats.CreditsNetIncome.Add(ev.TotalEarnings)
		j.State.Stats.CreditsIncome.Add(ev.TotalEarnings)
	case *parser.BuyWeapon:
		j.State.Stats.CreditsNetIncome.Sub(ev.Price)
		j.State.Stats.CreditsExpenses.Sub(ev.Price)
	case *parser.ModuleBuyAndStore:
		j.State.Stats.CreditsNetIncome.Sub(ev.BuyPrice)
		j.State.Stats.CreditsExpenses.Sub(ev.BuyPrice)
	case *parser.BuyTradeData:
		j.State.Stats.CreditsNetIncome.Sub(ev.Cost)
		j.State.Stats.CreditsExpenses.Sub(ev.Cost)
	case *parser.SellSuit:
		j.State.Stats.CreditsNetIncome.Add(ev.Price)
		j.State.Stats.CreditsIncome.Add(ev.Price)
	case *parser.BuySuit:
		j.State.Stats.CreditsNetIncome.Sub(ev.Price)
		j.State.Stats.CreditsExpenses.Sub(ev.Price)
	case *parser.SellWeapon:
		j.State.Stats.CreditsNetIncome.Add(ev.Price)
		j.State.Stats.CreditsIncome.Add(ev.Price)
	case *parser.ShipyardBuy:
		j.State.Stats.CreditsNetIncome.Sub(ev.ShipPrice)
		j.State.Stats.CreditsExpenses.Sub(ev.ShipPrice)
	case *parser.ShipyardTransfer:
		j.State.Stats.CreditsNetIncome.Sub(ev.TransferPrice)
		j.State.Stats.CreditsExpenses.Sub(ev.TransferPrice)
	case *parser.Resurrect:
		j.State.Stats.CreditsNetIncome.Sub(ev.Cost)
		j.State.Stats.CreditsExpenses.Sub(ev.Cost)
	case *parser.FetchRemoteModule:
		j.State.Stats.CreditsNetIncome.Sub(ev.TransferCost)
		j.State.Stats.CreditsExpenses.Sub(ev.TransferCost)
	case *parser.ColonisationConstructionDepot:
		for _, r := range ev.ResourcesRequired { // TODO: check if legit approach
			j.State.Stats.CreditsNetIncome.Add(r.Payment)
			j.State.Stats.CreditsIncome.Add(r.Payment)
			j.State.Stats.CommodityNet.Add(r.Payment)
			j.State.Stats.CommoditySell.Add(r.Payment)
		}
	case *parser.MultiSellExplorationData:
		j.State.Stats.CreditsNetIncome.Add(ev.TotalEarnings)
		j.State.Stats.CreditsIncome.Add(ev.TotalEarnings)
	case *parser.BuyExplorationData:
		j.State.Stats.CreditsNetIncome.Sub(ev.Cost)
		j.State.Stats.CreditsExpenses.Sub(ev.Cost)
	case *parser.PVPKill:
		j.State.Stats.KillsPlayer.Add(1)
	case *parser.Bounty:
		j.State.Stats.KillsNPC.Add(1)
	case *parser.FactionKillBond:
		j.State.Stats.KillsCombatZone.Add(1)
	case *parser.StartJump:
		if ev.JumpType == "Hyperspace" {
			j.setLocation(&parser.Location{
				StarSystem: "=Hyperspace=",
			})
			txt := fmt.Sprintf("Jumping: %s (class %s)", ev.StarSystem, ev.StarClass)
			events <- journal.NewGenericMessage(txt, sessionCounter, ev.Timestamp, isLiveEvent)
		}
	case *parser.Location:
		j.setLocation(ev)
	case *parser.CargoTransfer:
	case *parser.CarrierBankTransfer:
	case *parser.ColonisationContribution:
	case *parser.FSDTarget:
	case *parser.MissionCompleted:
		if ev.Reward != nil {
			j.State.Stats.CreditsNetIncome.Add(*ev.Reward)
			j.State.Stats.CreditsIncome.Add(*ev.Reward)
		}
	case *parser.MissionFailed:
	case *parser.MissionAccepted:
	case *parser.MissionAbandoned:
	case *parser.Missions:
	case *parser.Resupply:
	case *parser.Screenshot:
	case *parser.SellMicroResources:
	case *parser.ShipLocker:
	case *parser.ShipRedeemed:
	case *parser.ShipyardRedeem:
	case *parser.SquadronStartup:
	case *parser.Statistics:
	case *parser.StoredModules:
	case *parser.TechnologyBroker:
	case *parser.Undocked:
	case *parser.Loadout:
	case *parser.Cargo:
	case *parser.ShipTargeted:
	case *parser.Continued:
		j.Logger.Info("rare continued event")
	default:
		if j.unhandledEvents.SetIfAbsent(event.EventType(), true) {
			j.Logger.Error("unhandled event", zap.Any("type", event.EventType()))
		}
		return
	}
	j.State.Stats.ProcessedEvents.Add(1)
}

func (j *JournalProcessingWorker) setLocation(ev *parser.Location) {
	j.State.CurrentThings.LocationSystem.Store(ev.StarSystem)
	j.State.CurrentThings.LocationBody.Store(ev.Body)
	j.State.CurrentThings.LocationBodyType.Store(ev.BodyType)
	j.State.CurrentThings.LocationPopulation.Store(fmt.Sprintf("%d", ev.Population))
}

func (j *JournalProcessingWorker) resetLocation() {
	j.State.CurrentThings.LocationSystem.Store("")
	j.State.CurrentThings.LocationBody.Store("")
	j.State.CurrentThings.LocationBodyType.Store("")
	j.State.CurrentThings.LocationPopulation.Store("")
}

func (j *JournalProcessingWorker) resetSession() {
	j.State.CurrentThings.CMDR.Store("")
	j.State.CurrentThings.FID.Store("")
	for _, friend := range j.State.FriendList.Keys() {
		j.State.FriendList.Set(friend, false)
	}
	j.State.CurrentThings.Music.Store("")
	j.resetLocation()
}

func (j *JournalProcessingWorker) processEventsAndProduceMessages(ctx context.Context, loadNPrev int, archive string, outcoming chan<- journal.MessageEvent) {
	Logger := j.Logger.Named("MessageProvider")

	j.State.ProcessingLiveEvents.Store(false)
	j.StateUI.PreloadCurrent.Store(0)

	var profileShouldBeClosed bool
	var fdProfile *os.File

	if j.profileCatchup {
		filename := fmt.Sprintf("catchup-%s.pprof", utils.GenerateTimestampForFilename())
		_ = os.MkdirAll(PProfDirectory, 0755)
		fd, err := os.Create(filepath.Join(PProfDirectory, filename))
		if err != nil {
			Logger.Error("Failed to pprof file", zap.Error(err))
		} else {
			fdProfile = fd
			profileShouldBeClosed = true
			_ = pprof.StartCPUProfile(fd)
		}
	}

	started := time.Now()
	prev, current, caughtUp := parser.AttachToLastLog(ctx, loadNPrev, archive)

	var cumSize int64
	var counters []parser.CountResult

	j.StateUI.PreloadMax = len(prev)
	for i, journalEntry := range prev {
		currentN := int64(i + 1)
		j.StateUI.PreloadCurrent.Store(currentN)
		j.StateUI.PreloadMessage.Store(fmt.Sprintf("Currently processing: %s (%3d of %3d)", journalEntry.Entry, currentN, j.StateUI.PreloadMax))
		j.State.CurrentThings.LogFile.Store(journalEntry.Entry.Name)

		var continued bool
		for event := range journalEntry.Events {
			_, ok := event.(*parser.Continued)
			if ok {
				continued = true
			}

			j.ProcessEvent(event, outcoming, j.State.SessionCounter.Load(), false)
		}
		if !continued {
			j.State.SessionCounter.Add(1)
			if j.ProcessedConfig.Stats.AlwaysResetSession {
				j.State.Stats.ResetAll()
			}
		}
		counterResult := <-journalEntry.CountResults
		j.State.Stats.UnprocessedEvents.Add(int(counterResult.Fail))
		counters = append(counters, counterResult)

		j.State.CurrentThings.LogFile.Store("")
		if journalEntry.Entry.Size != -1 {
			cumSize += journalEntry.Entry.Size
		} else {
			Logger.Warn("invalid size", zap.String("entry", journalEntry.Entry.String()))
		}
	}

	duration := time.Since(started)

	processed := j.State.Stats.ProcessedEvents.Load()
	bps := float64(cumSize) / duration.Seconds()
	eps := float64(processed) / duration.Seconds()
	avg := float64(cumSize) / float64(processed)

	Logger.Info("historical journals processed",
		zap.Duration("duration", duration),
		zap.Int("processed_events", processed),
		zap.Float64("processed_MB", float64(cumSize)/1024/1024),
		zap.Float64("MB/s", bps/1024/1024),
	)

	sum := parser.CountResult{}
	for _, counterResult := range counters {
		sum.Success += counterResult.Success
		sum.Fail += counterResult.Fail
		sum.Bytes += counterResult.Bytes
	}

	total := sum.Success + sum.Fail
	bps = float64(sum.Bytes) / duration.Seconds()
	eps = float64(total) / duration.Seconds()
	avg = float64(sum.Bytes) / float64(total)

	Logger.Info("counter report (more truthful)",
		zap.Duration("duration", duration),
		zap.Uint64("processed_events", sum.Success),
		zap.Float64("processed_MB", float64(sum.Bytes)/1024/1024),
		zap.Float64("MB/s", bps/1024/1024),
		zap.Float64("events/s", eps),
		zap.Uint64("failed/skipped", sum.Fail),
		zap.Uint64("total", total),
		zap.Float64("avg_event_size", avg),
	)

	var wg sync.WaitGroup
	// listening for caughtUp only after processing old entries as caughtUp might be emitted instantly
	// when there is no currently active journal that Elite is writing to.
	wg.Add(1)
	go func(caughtUp <-chan struct{}) {
		defer wg.Done()
		defer j.State.ProcessingLiveEventsCond.Broadcast()
		defer j.State.ProcessingLiveEvents.Store(true)

		<-caughtUp
		Logger.Info("caught up with live events", zap.Duration("duration", time.Since(started)), zap.Int("processed_events", j.State.Stats.ProcessedEvents.Load()))
		if j.profileCatchup && profileShouldBeClosed {
			pprof.StopCPUProfile()
			_ = fdProfile.Close()
		}
		time.Sleep(time.Millisecond * 10) // some additional time to process buffers to prevent sound notifications from firing.
	}(caughtUp)

	for journalEntry := range current {
		j.State.CurrentThings.LogFile.Store(journalEntry.Entry.Name)

		var continued bool
		for event := range journalEntry.Events {
			_, ok := event.(*parser.Continued)
			if ok {
				continued = true
			}
			j.ProcessEvent(event, outcoming, j.State.SessionCounter.Load(), true)
		}
		if !continued {
			j.State.SessionCounter.Add(1)
			if j.ProcessedConfig.Stats.AlwaysResetSession {
				j.State.Stats.ResetAll()
			}
		}
		// todo: this is locking
		//counterResult := <-journalEntry.CountResults
		//j.State.Stats.UnprocessedEvents.Add(int(counterResult.Fail))

		j.State.CurrentThings.LogFile.Store("")
	}

	Logger.Info("waiting for waitgroup")
	wg.Wait()
	Logger.Info("waiting finished")
}

type MessagePusher struct {
	channels map[int]chan journal.MessageEvent
	wg       *sync.WaitGroup
}

func NewMessagePusher(chanSize int, viewportChannels map[int]journal.MessageChannel, viewportBuffers map[int]*rvb.RVBuffer[journal.MessageEvent]) *MessagePusher {
	channels := make(map[int]chan journal.MessageEvent)

	wg := sync.WaitGroup{}

	for k, v := range viewportChannels {
		ch := make(chan journal.MessageEvent, chanSize)
		channels[k] = ch
		out := viewportBuffers[k]

		wg.Add(1)
		go func(inChan <-chan journal.MessageEvent, viewport int, channelMask journal.MessageChannel, out *rvb.RVBuffer[journal.MessageEvent]) {
			defer wg.Done()
			for event := range inChan {
				if event.MessageChannel()&channelMask == 0 {
					continue
				}
				out.Push(event)
			}
		}(ch, k, v, out)
	}

	return &MessagePusher{
		channels: channels,
		wg:       &wg,
	}
}

func (m *MessagePusher) Close() {
	for _, ch := range m.channels {
		close(ch)
	}
	m.wg.Wait()
}

func (m *MessagePusher) PushMessage(msg journal.MessageEvent) {
	for _, channel := range m.channels {
		channel <- msg
	}
}

func (j *JournalProcessingWorker) processMessages(messages <-chan journal.MessageEvent, notifications chan<- journal.MessageChannel) {
	Logger := j.Logger.Named("MessageProcessor")

	viewportChannels := j.ProcessedConfig.Viewports.ViewportChannelBitmask

	viewportBuffers := make(map[int]*rvb.RVBuffer[journal.MessageEvent])
	for i, viewport := range j.State.ViewportBuffers {
		viewportBuffers[i] = viewport.Buffer
	}

	pusher := NewMessagePusher(64, viewportChannels, viewportBuffers)
	defer pusher.Close()

	for rawMsg := range messages {
		switch msg := rawMsg.(type) {
		case *journal.SentMessage:
			pusher.PushMessage(msg)
		case *journal.ReceivedCMDRMessage:
			pusher.PushMessage(msg)
		case *journal.ReceivedNPCMessage:
			who := msg.DecodedMessage.What
			subType := msg.DecodedMessage.SubWhat

			blacklistMode := j.ProcessedConfig.Filters.NpcMessagesBlacklistMode
			filter, defined := j.ProcessedConfig.Filters.NpcMessages[who]
			if !defined {
				if blacklistMode {
					pusher.PushMessage(msg)
				}
				continue
			}

			var isChecked bool
			if len(filter.SubTypes) == 0 {
				if subType == "" {
					isChecked = filter.Base.Enabled
				}
			} else {
				subFilter, ok := filter.SubTypes[subType]
				if ok {
					isChecked = subFilter.Enabled
				}
			}

			switch {
			case blacklistMode && isChecked:
				continue
			case !blacklistMode && !isChecked:
				continue
			default:
				pusher.PushMessage(msg)
			}
		case *journal.NotificationMessage:
			who := msg.DecodedMessage.What
			subType := msg.DecodedMessage.SubWhat

			blacklistMode := j.ProcessedConfig.Filters.NotificationsBlacklistMode
			filter, defined := j.ProcessedConfig.Filters.Notifications[who]
			if !defined {
				if blacklistMode {
					pusher.PushMessage(msg)
				}
				continue
			}

			var isChecked bool
			if len(filter.SubTypes) == 0 {
				if subType == "" {
					isChecked = filter.Base.Enabled
				}
			} else {
				subFilter, ok := filter.SubTypes[subType]
				if ok {
					isChecked = subFilter.Enabled
				}
			}

			switch {
			case blacklistMode && isChecked:
				continue
			case !blacklistMode && !isChecked:
				continue
			default:
				pusher.PushMessage(msg)
			}
		case *journal.GenericEvent:
			pusher.PushMessage(msg)
		case *journal.CommanderEvent:
			continue
		case *journal.WingInviteMessage:
			pusher.PushMessage(msg)
		case *journal.FriendInviteMessage:
			pusher.PushMessage(msg)
		//case *journal.SquadronInviteMessage:
		//	pusher.PushMessage(msg)
		default:
			Logger.Error("Unhandled message", zap.String("type", fmt.Sprintf("%T", msg)))
			continue
		}

		j.State.Stats.ProducedMessages.Add(1)

		if j.State.ProcessingLiveEvents.Load() {
			notifications <- rawMsg.MessageChannel()
		}
	}
}

func (j *JournalProcessingWorker) processNotifications(notifications <-chan journal.MessageChannel) {
	j.StateUI.LoadedSoundsCounter = make(map[journal.MessageChannel]int, len(journal.ChannelAllList))
	for ch := range notifications {
		_, ok := j.StateUI.LoadedSoundsSource[ch]
		if !ok {
			continue
		}

		counter := j.StateUI.LoadedSoundsCounter[ch]
		rl.PlaySound(j.StateUI.LoadedSounds[ch][counter])
		j.StateUI.LoadedSoundsCounter[ch] = (j.StateUI.LoadedSoundsCounter[ch] + 1) % j.maxSounds
	}
}
