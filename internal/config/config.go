package config

import (
	"time"

	"github.com/gethiox/EDCITR/internal/app/journal"
	"github.com/gethiox/EDCITR/internal/noire"
	"github.com/gethiox/EDCITR/internal/version"
)

type Level uint

const (
	LevelUndefined Level = iota

	LevelUseless
	LevelInteresting
	LevelGood
	LevelWarning
	LevelCritical
)

type MsgType struct {
	ValueForUser       string
	DescriptionForUser string

	Value string
	Level Level
}

type Stats struct {
	Enabled            bool
	AlwaysResetSession bool

	JumpsAndDockings bool
	Signals          bool
	Credits          bool
	Commodities      bool
	Playtime         bool
	Kills            bool
	Journal          bool
	Sessions         bool
	CurrentVariables bool
	CurrentLocation  bool
}

type StatusBar struct {
	Enabled bool

	ServerStatusText bool
	CheckForUpdates  bool
	NextMaintenance  bool
	InaraStats       bool
	CurrentJournal   bool

	MaintenanceTime time.Time //utc
}

type GenericImages struct {
	Enabled     bool
	ImageSize   float32
	HoverSize   float32
	BorderWidth int
}

type InaraSettings struct {
	Enabled bool

	UnloadPeriod      time.Duration `json:",format:sec"`
	WaitBeforeRequest time.Duration `json:",format:sec"`
	CMDRImages        bool
	ProfileIndicator  bool
	Token             string
}

type ServerStatusSettings struct {
	Enabled     bool
	Timeout     time.Duration `json:",format:sec"`
	Period      time.Duration `json:",format:sec"`
	HistorySize int
}

type Toggles struct {
	PrefixCMDR           bool
	PrefixNPC            bool
	TimestampEliteFormat bool
	Timestamps           bool
	AlwaysShowTimestamps bool
}

type Parser struct {
	PreloadNLogfiles int
	ArchiveDirectory string

	MaxConcurrentFiles      int
	ProcessorWorkers        int
	ProcessorWorkerChanSize int
	ProcessorResultChanSize int
}

type Colors struct {
	HeaderCMDR         my_noire.Color
	HeaderNPC          my_noire.Color
	HeaderNotification my_noire.Color
	HeaderFriend       my_noire.Color
	HeaderEnemy        my_noire.Color
	HeaderSent         my_noire.Color
	HeaderDefault      my_noire.Color

	ChannelLocal        my_noire.Color
	ChannelPlayer       my_noire.Color
	ChannelStarSystem   my_noire.Color
	ChannelSquadron     my_noire.Color
	ChannelWing         my_noire.Color
	ChannelVoice        my_noire.Color
	ChannelJournal      my_noire.Color
	ChannelNotification my_noire.Color

	NPCPositive             my_noire.Color
	NPCWarning              my_noire.Color
	NPCNormal               my_noire.Color
	NPCCritical             my_noire.Color
	NotificationWarning     my_noire.Color
	NotificationInteresting my_noire.Color
	NotificationNormal      my_noire.Color
	NotificationCritical    my_noire.Color
	NotificationPositive    my_noire.Color

	GeneralWarning  my_noire.Color
	GeneralGray     my_noire.Color
	GeneralDarkGray my_noire.Color
	GeneralNormal   my_noire.Color
	GeneralCritical my_noire.Color
	GeneralOK       my_noire.Color

	HighlightWord      my_noire.Color
	HighlightUrl       my_noire.Color
	NoMoreMessages     my_noire.Color
	NoMoreMessagesText my_noire.Color
	FollowingDisabled  my_noire.Color
	StatusBar1         my_noire.Color
	StatusBar2         my_noire.Color
	Header1            my_noire.Color
	Header2            my_noire.Color
	MewMessageMarker   my_noire.Color
	InaraLogo          my_noire.Color
	TimeFormatLocal    my_noire.Color
	TimeFormatElite    my_noire.Color
	ViewportBorder     my_noire.Color
	Background         my_noire.Color

	Palette1 my_noire.Color
	Palette2 my_noire.Color
	Palette3 my_noire.Color
	Palette4 my_noire.Color
	Palette5 my_noire.Color
	Palette6 my_noire.Color
	Palette7 my_noire.Color
	Palette8 my_noire.Color
}

type ViewportAppearance struct {
	LeftPad           int
	RightPad          int
	MessageSpacing    int      // vertical spacing between separate messages
	MessageIndentPad  int      // horizontal message body indentation size, relative to "header" (CMDR name) position
	MarkerWidth       int      // width of *new message* marker box
	MarkerHoldTime    float32  // seconds
	MarkerFadeoutTime float32  // seconds
	HighlightWords    []string // comma-separated list of words to highligh in messages
	TimestampFormat   string
	Toggles           Toggles
}

type Viewports struct {
	BufferSize             int
	Count                  int
	BorderSize             int // width of column borders
	Appearance             ViewportAppearance
	ViewportChannelBitmask map[int]journal.MessageChannel // channel configuration for given column
}

type Audio struct {
	MasterVolume      float32
	ChannelSoundPaths map[journal.MessageChannel]string
	ChannelVolume     map[journal.MessageChannel]float32
}

type FilterInfo struct {
	Enabled bool
	Level   Level
}

type Filter struct {
	Base     *FilterInfo
	SubTypes map[string]*FilterInfo
}

type FilterMap map[string]Filter

func (f *FilterMap) Exists(base, subType string) bool {
	if subType == "" {
		_, ok := (*f)[base]
		return ok
	}

	return f.IsSubtype(base, subType)
}

func (f *FilterMap) IsSubtype(base, subType string) bool {
	filter, ok := (*f)[base]
	if !ok {
		return false
	}

	_, ok = filter.SubTypes[subType]
	return ok
}

type Filters struct {
	Notifications              FilterMap
	NpcMessages                FilterMap
	NotificationsBlacklistMode bool
	NpcMessagesBlacklistMode   bool
}

type EDCITRConfig struct {
	Version    version.Version
	FPS        int
	WheelSpeed float32

	Audio         Audio
	Colors        Colors
	Filters       Filters
	Fonts         Fonts
	GenericImages GenericImages
	InaraSettings InaraSettings
	Parser        Parser
	ServerStatus  ServerStatusSettings
	Stats         Stats
	StatusBar     StatusBar
	Viewports     Viewports

	IsPonyEnjoyer bool
}

type Fonts struct {
	MainPath string
	MonoPath string

	Size        int
	Spacing     int
	SpacingMono int
}

func GetDefaultConfig() *EDCITRConfig {
	config := EDCITRConfig{
		Version:    version.Release,
		FPS:        60,
		WheelSpeed: 1.0,

		Audio: Audio{
			MasterVolume: 0.5,
			ChannelSoundPaths: map[journal.MessageChannel]string{
				journal.ChannelLocal:        "builtin:sounds/games/Oddworld SligHi.ogg",
				journal.ChannelNPC:          "builtin:sounds/games/Doom ImpIdle.ogg",
				journal.ChannelPlayer:       "builtin:sounds/random/MineTurtleHello.ogg",
				journal.ChannelStarSystem:   "builtin:sounds/coms/Twitter.ogg",
				journal.ChannelSquadron:     "builtin:sounds/coms/Skype Message.ogg",
				journal.ChannelWing:         "builtin:sounds/systems/Oneplus Meet.ogg",
				journal.ChannelVoice:        "builtin:sounds/systems/Oneplus Meet.ogg",
				journal.ChannelJournal:      "builtin:sounds/games/Doom Pickup.ogg",
				journal.ChannelNotification: "builtin:sounds/games/BeyondAllReason beep4.ogg",
			},
			ChannelVolume: map[journal.MessageChannel]float32{
				journal.ChannelLocal:        0.7,
				journal.ChannelNPC:          0.5,
				journal.ChannelPlayer:       0.7,
				journal.ChannelStarSystem:   0.7,
				journal.ChannelSquadron:     0.7,
				journal.ChannelWing:         0.7,
				journal.ChannelVoice:        0.7,
				journal.ChannelJournal:      0.4,
				journal.ChannelNotification: 0.5,
			},
		},

		Colors: Colors{ // TODO: shitty noire doesn't keep alpha values on Hex encoding/decoding
			HeaderDefault:      my_noire.NewHex("#888888ff"),
			HeaderNotification: my_noire.NewHex("#bdbdbdff"),
			HeaderCMDR:         my_noire.NewHex("#ffd619ff"),
			HeaderNPC:          my_noire.NewHex("#bdbdbdff"),
			HeaderFriend:       my_noire.NewHex("#20ff08ff"),
			HeaderEnemy:        my_noire.NewHex("#ff4619ff"),
			HeaderSent:         my_noire.NewHex("#00f4ffff"),
			FollowingDisabled:  my_noire.NewHex("#aa2200ff"),

			ChannelLocal:        my_noire.NewHex("#d1d1d1ff"),
			ChannelPlayer:       my_noire.NewHex("#5ccde4ff"),
			ChannelStarSystem:   my_noire.NewHex("#f8c589ff"),
			ChannelSquadron:     my_noire.NewHex("#70c665ff"),
			ChannelWing:         my_noire.NewHex("#7290ffff"),
			ChannelVoice:        my_noire.NewHex("#7290ffff"),
			ChannelJournal:      my_noire.NewHex("#848484ff"),
			ChannelNotification: my_noire.NewHex("#848484ff"),

			NPCNormal:               my_noire.NewHex("#bdbdbdff"),
			NPCWarning:              my_noire.NewHex("#ffd619ff"),
			NPCCritical:             my_noire.NewHex("#ff4619ff"),
			NPCPositive:             my_noire.NewHex("#20ff08ff"),
			NotificationNormal:      my_noire.NewHex("#bdbdbdff"),
			NotificationWarning:     my_noire.NewHex("#ffd619ff"),
			NotificationInteresting: my_noire.NewHex("#0a8bd6ff"),
			NotificationCritical:    my_noire.NewHex("#ff4619ff"),
			NotificationPositive:    my_noire.NewHex("#20ff08ff"),

			GeneralNormal:   my_noire.NewHex("#d1d1d1ff"),
			GeneralGray:     my_noire.NewHex("#bdbdbdff"),
			GeneralDarkGray: my_noire.NewHex("#6d6d6dff"),
			GeneralWarning:  my_noire.NewHex("#ff9f00ff"),
			GeneralCritical: my_noire.NewHex("#ff0000ff"),
			GeneralOK:       my_noire.NewHex("#00ff00ff"),
			TimeFormatLocal: my_noire.NewHex("#888888ff"),
			TimeFormatElite: my_noire.NewHex("#00f4ffff"),

			HighlightWord:      my_noire.NewHex("#ff4619ff"),
			HighlightUrl:       my_noire.NewHex("#99c3ffff"),
			NoMoreMessages:     my_noire.NewHex("#aa0000ff"),
			NoMoreMessagesText: my_noire.NewHex("#000000ff"),
			StatusBar1:         my_noire.NewHex("#000000ff"),
			StatusBar2:         my_noire.NewHex("#222222ff"),
			Header1:            my_noire.NewHex("#000000ff"),
			Header2:            my_noire.NewHex("#222222ff"),

			MewMessageMarker: my_noire.NewHex("#20ff08ff"),
			InaraLogo:        my_noire.NewHex("#d1823fff"),
			ViewportBorder:   my_noire.NewHex("#00740bff"),
			Background:       my_noire.NewHex("#000000ff"),

			Palette1: my_noire.NewHex("#00b3f7"),
			Palette2: my_noire.NewHex("#0a8bd6"),
			Palette3: my_noire.NewHex("#029e4c"),
			Palette4: my_noire.NewHex("#ff0000"),
			Palette5: my_noire.NewHex("#ff7100"),
			Palette6: my_noire.NewHex("#ffb000"),
			Palette7: my_noire.NewHex("#f07b05"),
			Palette8: my_noire.NewHex("#000000"),
		},

		Filters: Filters{
			Notifications:              nil, // will be constructed dynamically
			NpcMessages:                nil, // will be constructed dynamically
			NotificationsBlacklistMode: true,
			NpcMessagesBlacklistMode:   true,
		},

		Fonts: Fonts{
			MainPath:    "builtin:fonts/Dosis-Bold.ttf",
			MonoPath:    "builtin:fonts/Dosis-SpaceMono-Bold.ttf",
			Size:        30,
			Spacing:     0,
			SpacingMono: -1,
		},

		GenericImages: GenericImages{
			Enabled:     true,
			ImageSize:   1.0,
			HoverSize:   300,
			BorderWidth: 2,
		},

		InaraSettings: InaraSettings{
			Enabled:           true,
			UnloadPeriod:      time.Second * 60,
			WaitBeforeRequest: time.Second * 5,
			CMDRImages:        true,
			ProfileIndicator:  true,
			Token:             "<Inara API token>",
		},

		Parser: Parser{
			PreloadNLogfiles: 2560,
			ArchiveDirectory: ``,

			MaxConcurrentFiles:      8,
			ProcessorWorkers:        16,
			ProcessorWorkerChanSize: 64,
			ProcessorResultChanSize: 128,
		},

		ServerStatus: ServerStatusSettings{
			Enabled:     true,
			Timeout:     time.Second * 60,
			Period:      time.Second * 20,
			HistorySize: 60,
		},

		Stats: Stats{
			Enabled:            false,
			AlwaysResetSession: false,

			JumpsAndDockings: true,
			Signals:          true,
			Credits:          true,
			Commodities:      false,
			Playtime:         false,
			Kills:            true,
			Journal:          false,
			Sessions:         false,
			CurrentVariables: true,
			CurrentLocation:  true,
		},

		StatusBar: StatusBar{
			Enabled:          true,
			ServerStatusText: true,
			CheckForUpdates:  false,
			NextMaintenance:  true,
			InaraStats:       true,
			CurrentJournal:   true,
		},

		Viewports: Viewports{
			BufferSize: 100_000,
			Count:      4,
			BorderSize: 6,
			Appearance: ViewportAppearance{
				LeftPad:           0,
				RightPad:          0,
				MessageSpacing:    8,
				MessageIndentPad:  24,
				MarkerWidth:       20,
				MarkerHoldTime:    1.0,
				MarkerFadeoutTime: 4.0,
				HighlightWords:    []string{"FDev", "StellarForge", "Cobra", "o7"},
				TimestampFormat:   "02.01.2006 15:04",
				Toggles: Toggles{
					PrefixCMDR:           false,
					PrefixNPC:            true,
					TimestampEliteFormat: true,
					Timestamps:           true,
					AlwaysShowTimestamps: false,
				},
			},
			ViewportChannelBitmask: map[int]journal.MessageChannel{
				0: journal.ChannelNotification,
				1: journal.ChannelLocal,
				2: journal.ChannelNPC,
				3: journal.ChannelPlayer,
				4: journal.ChannelWing | journal.ChannelVoice,
				5: journal.ChannelSquadron,
				6: journal.ChannelStarSystem,
				7: journal.ChannelJournal,
			},
		},

		IsPonyEnjoyer: false,
	}

	// building notification filter list
	config.Filters.Notifications = make(FilterMap, len(LocalNotificationsAvailable))
	for _, s := range LocalNotificationsAvailable {
		config.Filters.Notifications[s.Who] = Filter{
			Base: &FilterInfo{
				Enabled: true,
				Level:   s.BaseLevel,
			},
			SubTypes: make(map[string]*FilterInfo, len(s.MsgTypes)),
		}
		if len(s.MsgTypes) == 0 {
			continue
		}

		for _, t := range s.MsgTypes {
			config.Filters.Notifications[s.Who].SubTypes[t.Value] = &FilterInfo{
				Enabled: true,
				Level:   t.Level,
			}
		}
	}
	//config.Filters.Notifications["STATION"].SubTypes["docking_denied"].Enabled = true
	//config.Filters.Notifications["DockingFailed"].SubTypes["Offences"].Enabled = true

	// building npc messages filter list
	config.Filters.NpcMessages = make(FilterMap, len(NPCMessagesAvailable))
	for _, s := range NPCMessagesAvailable {
		config.Filters.NpcMessages[s.Who] = Filter{
			Base: &FilterInfo{
				Enabled: true,
				Level:   s.BaseLevel,
			},
			SubTypes: make(map[string]*FilterInfo, len(s.MsgTypes)),
		}
		if len(s.MsgTypes) == 0 {
			continue
		}

		for _, t := range s.MsgTypes {
			config.Filters.NpcMessages[s.Who].SubTypes[t.Value] = &FilterInfo{
				Enabled: true,
				Level:   t.Level,
			}
		}
	}
	//config.Filters.NpcMessages["Mission"].SubTypes["Collect_MessengerChat"].Enabled = true             // TODO: testing, remove
	//config.Filters.NpcMessages["Wreckage"].SubTypes["SearchandRescue_AccidentalDamage"].Enabled = true // TODO: testing, remove

	return &config
}
