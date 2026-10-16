package app

import (
	"math"
	"time"

	"github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/app/journal"
	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/noire"
)

type KeyAction struct {
	key    int32
	action func()
}

type ASDF struct {
	text    string
	color   my_noire.Color
	font    *rl.Font
	spacing float32
}

type Renderer struct {
	app *App
	s   state.UIState

	colorSettings   []*state.ColorSetting
	asdf            []ASDF
	settingsGeneral []Widget
	settingsNetwork []Widget
	settingsParser  []Widget
	settingsStats   []Widget
	exportSettings  []Widget
	ChannelToColor  map[journal.MessageChannel]*my_noire.Color
	keyActions      []KeyAction
}

func NewRenderer(app *App) *Renderer {
	r := &Renderer{
		app: app,
		s: state.UIState{
			FilterExpanded: make(map[string]bool),
		},
	}

	r.settingsGeneral = []Widget{
		Slider(&app.Config.Viewports.Count, "Display Segments ", 1, 8),
		SliderBar(&app.Config.Fonts.Size, "Font Size ", 16, 128, OnChange(func() { r.s.FontReloadRequired = true })), // 14 is minimum to cover rendered logs
		SliderBar(&app.Config.Fonts.Spacing, "Font Spacing ", -8, 16),
		SliderBar(&app.Config.Fonts.SpacingMono, "Mono Font Spacing ", -8, 16),
		SliderBar(&app.Config.Viewports.Appearance.MessageIndentPad, "Message Body LeftPad ", 0, 255),
		SliderBar(&app.Config.Viewports.Appearance.MessageSpacing, "Message Spacing ", 0, 64),
		SliderBar(&app.Config.FPS, "FPS ", 10, 240, OnChange(func() { r.s.FPSReloadRequired = true })),
		SliderBar(&app.Config.Viewports.Appearance.MarkerWidth, "Marker Width ", 0, 64),
		SliderBar(&app.Config.Viewports.Appearance.MarkerHoldTime, "Marker Hold Duration ", 0, 10),
		SliderBar(&app.Config.Viewports.Appearance.MarkerFadeoutTime, "Marker Fadeout Duration ", 0, 10),
		SliderBar(&app.Config.Viewports.BorderSize, "Viewport Border Width ", 0, 64),
		SliderBar(&app.Config.WheelSpeed, "WheelSpeed ", 0.1, 8.0),
		Checkbox(&app.Config.Stats.Enabled, "Enable Statistics (WIP)"),
		Checkbox(&app.Config.Stats.AlwaysResetSession, "Always reset stats on new session"),
		Checkbox(&app.Config.Viewports.Appearance.Toggles.Timestamps, "Timestamps"),
		Checkbox(&app.Config.Viewports.Appearance.Toggles.PrefixCMDR, "CMDR Prefix"),
		Checkbox(&app.Config.Viewports.Appearance.Toggles.PrefixNPC, "NPC Prefix"),
		Checkbox(&app.Config.Viewports.Appearance.Toggles.AlwaysShowTimestamps, "Always Show Timestamps"),
		Checkbox(&app.Config.Viewports.Appearance.Toggles.TimestampEliteFormat, "Elite Timestamp Format"),
	}

	r.settingsNetwork = []Widget{
		Checkbox(&app.Config.ServerStatus.Enabled, "Enable Server Status", OnChange(r.togglePing)),
		Checkbox(&app.Config.GenericImages.Enabled, "Generic Images"),
		SliderBar(&app.Config.GenericImages.BorderWidth, "Generic Image Border Width ", 0, 10),
		SliderBar(&app.Config.GenericImages.ImageSize, "Generic Image Size ", 1, 4),
		SliderBar(&app.Config.GenericImages.HoverSize, "Generic Image Size on Hover", 50, 300),
		Checkbox(&app.Config.InaraSettings.Enabled, "Enable Inara Integration", OnChange(r.toggleInara)),
		Checkbox(&app.Config.InaraSettings.CMDRImages, "Enable Inara Profile Pictures"),
		Checkbox(&app.Config.InaraSettings.ProfileIndicator, "Enable Inara Profile Indicator"),
		DurationSlider(&app.Config.InaraSettings.UnloadPeriod, "UnloadPeriod", time.Second, 0.0, 60.0*5, OnChange(func() {
			r.s.UnloadPeriod.Store(app.Config.InaraSettings.UnloadPeriod.Nanoseconds())
		})),
		Button("Clear Cache", r.clearCache),
		Button("Delete Config", r.deleteConfig),
	}

	r.exportSettings = []Widget{
		Checkbox(&app.Config.StatusBar.Enabled, "Enable Status Bar"),
		Checkbox(&app.Config.StatusBar.ServerStatusText, "Server Status Text"),
		Checkbox(&app.Config.StatusBar.NextMaintenance, "Next Maintenance"),
		Checkbox(&app.Config.StatusBar.InaraStats, "Inara Stas"),
		Checkbox(&app.Config.StatusBar.CurrentJournal, "Current Journal"),
		Checkbox(&app.Config.StatusBar.CheckForUpdates, "check for updates", OnChange(r.toggleUpdateChecker)),
		Button("crash for no reason", func() { panic("crashed for no reason") }),
		//Checkbox(&app.Config.StatusBar.MaintenanceTime, "check for updates", OnChange(r.toggleUpdateChecker)),

		Button("export events", r.startGenerateJsons),
		Button("export NPC messages", r.startGenerateNPCMessages),
		Button("export notifications", r.startGenerateNotifications),
	}

	r.settingsParser = []Widget{
		TextBox(&app.Config.Parser.ArchiveDirectory, "Archive directory", &r.s.EditArchiveDirectory, OnEdit(r.disableGlobalHotkeys), AcceptClipboard(), AcceptDragAndDrop()),
		Spinner(&app.Config.Parser.PreloadNLogfiles, "PreloadNLogfiles ", 0, math.MaxUint16, &r.s.EditPreloadNLogfiles, OnEdit(r.disableGlobalHotkeys)),
		SliderBar(&app.Config.Parser.MaxConcurrentFiles, "MaxConcurrentFiles ", 1, 16, OnChange(r.updateParserSettings)),
		SliderBar(&app.Config.Parser.ProcessorWorkers, "ProcessorWorkers ", 1, 32, OnChange(r.updateParserSettings)),
		SliderBar(&app.Config.Parser.ProcessorWorkerChanSize, "ProcessorWorkerChanSize ", 0, 128, OnChange(r.updateParserSettings)),
		SliderBar(&app.Config.Parser.ProcessorResultChanSize, "ProcessorResultChanSize ", 0, 512, OnChange(r.updateParserSettings)),
	}

	r.settingsStats = []Widget{
		Checkbox(&app.Config.Stats.JumpsAndDockings, "JumpsAndDockings"),
		Checkbox(&app.Config.Stats.Signals, "Signals"),
		Checkbox(&app.Config.Stats.Credits, "Credits"),
		Checkbox(&app.Config.Stats.Commodities, "Commodities"),
		Checkbox(&app.Config.Stats.Playtime, "Playtime"),
		Checkbox(&app.Config.Stats.Kills, "Kills"),
		Checkbox(&app.Config.Stats.Journal, "Journal"),
		Checkbox(&app.Config.Stats.Sessions, "Sessions"),
		Checkbox(&app.Config.Stats.CurrentVariables, "CurrentVariables"),
		Checkbox(&app.Config.Stats.CurrentLocation, "CurrentLocation"),
	}

	r.colorSettings = []*state.ColorSetting{
		state.NewColorSetting("Header Default", &app.Config.Colors.HeaderDefault, app.DefaultConfigReadOnly.Colors.HeaderDefault),
		state.NewColorSetting("Header Local Notification", &app.Config.Colors.HeaderNotification, app.DefaultConfigReadOnly.Colors.HeaderNotification),
		state.NewColorSetting("Header CMDR", &app.Config.Colors.HeaderCMDR, app.DefaultConfigReadOnly.Colors.HeaderCMDR),
		state.NewColorSetting("Header NPC", &app.Config.Colors.HeaderNPC, app.DefaultConfigReadOnly.Colors.HeaderNPC),
		state.NewColorSetting("Header Friend", &app.Config.Colors.HeaderFriend, app.DefaultConfigReadOnly.Colors.HeaderFriend),
		state.NewColorSetting("Header Enemy", &app.Config.Colors.HeaderEnemy, app.DefaultConfigReadOnly.Colors.HeaderEnemy),
		state.NewColorSetting("Header (You)", &app.Config.Colors.HeaderSent, app.DefaultConfigReadOnly.Colors.HeaderSent),
		state.NewColorSetting("Warning Following Disabled", &app.Config.Colors.FollowingDisabled, app.DefaultConfigReadOnly.Colors.FollowingDisabled),

		state.NewColorSetting("Channel Status", &app.Config.Colors.ChannelNotification, app.DefaultConfigReadOnly.Colors.ChannelNotification),
		state.NewColorSetting("Channel Local", &app.Config.Colors.ChannelLocal, app.DefaultConfigReadOnly.Colors.ChannelLocal),
		state.NewColorSetting("Channel Player", &app.Config.Colors.ChannelPlayer, app.DefaultConfigReadOnly.Colors.ChannelPlayer),
		state.NewColorSetting("Channel Wing", &app.Config.Colors.ChannelWing, app.DefaultConfigReadOnly.Colors.ChannelWing),
		state.NewColorSetting("Channel Voice", &app.Config.Colors.ChannelVoice, app.DefaultConfigReadOnly.Colors.ChannelVoice),
		state.NewColorSetting("Channel Squadron", &app.Config.Colors.ChannelSquadron, app.DefaultConfigReadOnly.Colors.ChannelSquadron),
		state.NewColorSetting("Channel StarSystem", &app.Config.Colors.ChannelStarSystem, app.DefaultConfigReadOnly.Colors.ChannelStarSystem),
		state.NewColorSetting("Channel Journal", &app.Config.Colors.ChannelJournal, app.DefaultConfigReadOnly.Colors.ChannelJournal),

		state.NewColorSetting("NPC Normal", &app.Config.Colors.NPCNormal, app.DefaultConfigReadOnly.Colors.NPCNormal),
		state.NewColorSetting("NPC Warning", &app.Config.Colors.NPCWarning, app.DefaultConfigReadOnly.Colors.NPCWarning),
		state.NewColorSetting("NPC Critical", &app.Config.Colors.NPCCritical, app.DefaultConfigReadOnly.Colors.NPCCritical),
		state.NewColorSetting("NPC Positive", &app.Config.Colors.NPCPositive, app.DefaultConfigReadOnly.Colors.NPCPositive),
		state.NewColorSetting("Notification Normal", &app.Config.Colors.NotificationNormal, app.DefaultConfigReadOnly.Colors.NotificationNormal),
		state.NewColorSetting("Notification Warning", &app.Config.Colors.NotificationWarning, app.DefaultConfigReadOnly.Colors.NotificationWarning),
		state.NewColorSetting("Notification Interesting", &app.Config.Colors.NotificationInteresting, app.DefaultConfigReadOnly.Colors.NotificationInteresting),
		state.NewColorSetting("Notification Critical", &app.Config.Colors.NotificationCritical, app.DefaultConfigReadOnly.Colors.NotificationCritical),
		state.NewColorSetting("Notification Positive", &app.Config.Colors.NotificationPositive, app.DefaultConfigReadOnly.Colors.NotificationPositive),

		state.NewColorSetting("General Normal", &app.Config.Colors.GeneralNormal, app.DefaultConfigReadOnly.Colors.GeneralNormal),
		state.NewColorSetting("General Gray", &app.Config.Colors.GeneralGray, app.DefaultConfigReadOnly.Colors.GeneralGray),
		state.NewColorSetting("General Dark Gray", &app.Config.Colors.GeneralDarkGray, app.DefaultConfigReadOnly.Colors.GeneralDarkGray),
		state.NewColorSetting("General Warning", &app.Config.Colors.GeneralWarning, app.DefaultConfigReadOnly.Colors.GeneralWarning),
		state.NewColorSetting("General Critical", &app.Config.Colors.GeneralCritical, app.DefaultConfigReadOnly.Colors.GeneralCritical),
		state.NewColorSetting("General Positive", &app.Config.Colors.GeneralOK, app.DefaultConfigReadOnly.Colors.GeneralOK),
		state.NewColorSetting("Local Time", &app.Config.Colors.TimeFormatLocal, app.DefaultConfigReadOnly.Colors.TimeFormatLocal),
		state.NewColorSetting("Elite Time", &app.Config.Colors.TimeFormatElite, app.DefaultConfigReadOnly.Colors.TimeFormatElite),

		state.NewColorSetting("Highlight Word", &app.Config.Colors.HighlightWord, app.DefaultConfigReadOnly.Colors.HighlightWord),
		state.NewColorSetting("Highlight URL", &app.Config.Colors.HighlightUrl, app.DefaultConfigReadOnly.Colors.HighlightUrl),
		state.NewColorSetting("Terminator (No more messages)", &app.Config.Colors.NoMoreMessages, app.DefaultConfigReadOnly.Colors.NoMoreMessages),
		state.NewColorSetting("Terminator Text (No more messages)", &app.Config.Colors.NoMoreMessagesText, app.DefaultConfigReadOnly.Colors.NoMoreMessagesText),
		state.NewColorSetting("Status Bar 1", &app.Config.Colors.StatusBar1, app.DefaultConfigReadOnly.Colors.StatusBar1),
		state.NewColorSetting("Status Bar 2", &app.Config.Colors.StatusBar2, app.DefaultConfigReadOnly.Colors.StatusBar2),
		state.NewColorSetting("Header 1", &app.Config.Colors.Header1, app.DefaultConfigReadOnly.Colors.Header1),
		state.NewColorSetting("Header 2", &app.Config.Colors.Header2, app.DefaultConfigReadOnly.Colors.Header2),

		state.NewColorSetting("New Message Marker", &app.Config.Colors.MewMessageMarker, app.DefaultConfigReadOnly.Colors.MewMessageMarker),
		state.NewColorSetting("Inara Logo", &app.Config.Colors.InaraLogo, app.DefaultConfigReadOnly.Colors.InaraLogo),
		state.NewColorSetting("Viewport Border", &app.Config.Colors.ViewportBorder, app.DefaultConfigReadOnly.Colors.ViewportBorder),
		state.NewColorSetting("Background", &app.Config.Colors.Background, app.DefaultConfigReadOnly.Colors.Background),
		{}, // padding
		{}, // padding
		{}, // padding
		{}, // padding

		{}, // padding
		{}, // padding
		{}, // padding
		{}, // padding
		{}, // padding
		{}, // padding
		{}, // padding
		{}, // padding

		state.NewColorSetting("Palette 1 (Unused)", &app.Config.Colors.Palette1, app.DefaultConfigReadOnly.Colors.Palette1),
		state.NewColorSetting("Palette 2 (Unused)", &app.Config.Colors.Palette2, app.DefaultConfigReadOnly.Colors.Palette2),
		state.NewColorSetting("Palette 3 (Unused)", &app.Config.Colors.Palette3, app.DefaultConfigReadOnly.Colors.Palette3),
		state.NewColorSetting("Palette 4 (Unused)", &app.Config.Colors.Palette4, app.DefaultConfigReadOnly.Colors.Palette4),
		state.NewColorSetting("Palette 5 (Unused)", &app.Config.Colors.Palette5, app.DefaultConfigReadOnly.Colors.Palette5),
		state.NewColorSetting("Palette 6 (Unused)", &app.Config.Colors.Palette6, app.DefaultConfigReadOnly.Colors.Palette6),
		state.NewColorSetting("Palette 7 (Unused)", &app.Config.Colors.Palette7, app.DefaultConfigReadOnly.Colors.Palette7),
		state.NewColorSetting("Palette 8 (Unused)", &app.Config.Colors.Palette8, app.DefaultConfigReadOnly.Colors.Palette8),
	}

	r.keyActions = []KeyAction{
		{rl.KeyF, rl.ToggleBorderlessWindowed},
		{rl.KeyS, r.toggleMaximize},
		{rl.KeyTab, r.toggleSettingsView},
		{rl.KeyL, func() { r.s.ShowLogs = !r.s.ShowLogs }},
	}

	r.ChannelToColor = map[journal.MessageChannel]*my_noire.Color{
		journal.ChannelLocal:        &app.Config.Colors.ChannelLocal,
		journal.ChannelNPC:          &app.Config.Colors.ChannelLocal,
		journal.ChannelPlayer:       &app.Config.Colors.ChannelPlayer,
		journal.ChannelWing:         &app.Config.Colors.ChannelWing,
		journal.ChannelVoice:        &app.Config.Colors.ChannelVoice,
		journal.ChannelSquadron:     &app.Config.Colors.ChannelSquadron,
		journal.ChannelStarSystem:   &app.Config.Colors.ChannelStarSystem,
		journal.ChannelNotification: &app.Config.Colors.ChannelNotification,
		journal.ChannelJournal:      &app.Config.Colors.ChannelJournal,
	}

	return r
}
