package state

import (
	"image"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/app/journal"
	"github.com/gethiox/EDCITR/internal/noire"
	"github.com/gethiox/rotational-velocidensity-buffer"
)

type Tab int

const (
	TabGeneral Tab = iota
	TabFilters
	TabColors
	TabAudio
	TabNetwork
	TabExport
	TabParser
	TabStats

	numTabs
)

var tabNames = [numTabs]string{
	TabGeneral: "General",
	TabFilters: "Filters",
	TabColors:  "Colors",
	TabAudio:   "Audio",
	TabNetwork: "Network",
	TabExport:  "Export",
	TabParser:  "Parser",
	TabStats:   "Stats",
}

func (t Tab) String() string {
	if t < 0 || t >= numTabs {
		return "Unknown"
	}
	return tabNames[t]
}

var TabListAll = func() []Tab {
	tabs := make([]Tab, numTabs)
	for i := Tab(0); i < numTabs; i++ {
		tabs[i] = i
	}
	return tabs
}()

type TexLoadInfo struct {
	CmdrName string
	Img      image.Image
}

type UIState struct { // TODO: todo
	CanvasSizeX               int
	CanvasSizeY               int
	CanvasMiddleX             int
	CanvasMiddleY             int
	MousePosition             rl.Vector2
	MouseFrameDelta           rl.Vector2
	MouseOverWindow           bool
	MouseOverWindowEngaged    time.Time
	MouseOverWindowDisengaged time.Time
	FontReloadRequired        bool
	FPSReloadRequired         bool
	VolumeReloadRequired      bool
	SoundListReloadRequired   bool
	IsMaximized               bool

	OptionsActiveTab          Tab
	OptionViewActive          bool
	CurrentlyModifiedColor    *ColorSetting
	CurrentlyPressedColor     *ColorSetting
	CurrentlyReleasedColor    *ColorSetting
	HighlightWordsEditing     bool
	HighlightWordsEditingWarn bool
	DisableShortcuts          bool

	BuiltInSoundList   [][2]string // name, path
	ShowLogs           bool
	ShowLogsPos        int
	ShowLogsCheckpoint rvb.Checkpoint

	PreloadCurrent atomic.Int64
	PreloadMax     int
	PreloadMessage AtomicValue[string]

	ProfilePicturesToLoad    []TexLoadInfo
	ProfilePicturesToLoadMtx sync.Mutex

	GopherSound rl.Sound
	FSSound     rl.Sound

	CurrentTime atomic.Int64 // unix nano

	LoadedSoundsSource  map[journal.MessageChannel]rl.Sound
	LoadedSounds        map[journal.MessageChannel][]rl.Sound
	LoadedSoundsCounter map[journal.MessageChannel]int

	NotificationsScrollY float32
	FiltersScrollY       float32

	StatePeek            bool
	StateHoverPeek       bool
	StatePeekAt          time.Time
	StatePeekReleasedAt  time.Time
	StatePeekChangedAt   float32
	StatePeekCurrent     float32 // 0-1
	GuiCurrentAlpha      float32
	GuiCurrentAlphaImage float32
	GuiCurrentBgAlpha    float32

	HoverPos                               rl.Vector2
	AlternativeHoverPos                    rl.Vector2
	LastCmdrNameHover                      string
	IsCurrentlyHoveringOverCmdrName        bool
	LastExtraParametersHover               map[string]string
	IsCurrentlyHoveringOverExtraParameters bool
	OverrideAlpha                          int

	FilterScrollY         float32
	FilterExpanded        map[string]bool
	LastTotalHeight       float32
	CurrentlyEditingSound journal.MessageChannel
	SoundListCurrentPos   int

	IsEditing            bool
	EditArchiveDirectory bool
	EditPreloadNLogfiles bool
	UnloadPeriod         atomic.Int64
}

type ColorSetting struct {
	Name         string
	Color        *my_noire.Color
	Hsv          rl.Vector3
	DefaultColor my_noire.Color
}

func (s *ColorSetting) UpdateHSV() {
	hh, ss, vv := s.Color.HSV()
	s.Hsv = rl.Vector3{
		float32(hh),
		float32(ss / 100),
		float32(vv / 100),
	}
}

func NewColorSetting(name string, color *my_noire.Color, defaultColor my_noire.Color) *ColorSetting {
	s := ColorSetting{
		Name:         name,
		Color:        color,
		DefaultColor: defaultColor,
	}
	s.UpdateHSV()
	return &s
}
