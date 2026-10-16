package state

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/assets"
	"github.com/gethiox/EDCITR/internal/app/journal"
	"github.com/gethiox/EDCITR/internal/config"
	"github.com/gethiox/EDCITR/internal/elite_status"
	"github.com/gethiox/EDCITR/internal/storage"
	"github.com/gethiox/EDCITR/internal/task"
	"github.com/gethiox/EDCITR/internal/utils"
	"github.com/gethiox/EDCITR/internal/version"
	"github.com/gethiox/EDCITR/internal/zaputils"
	"github.com/gethiox/rotational-velocidensity-buffer"
	"github.com/orcaman/concurrent-map/v2"
	"go.uber.org/zap"
)

var Logger = zaputils.Logger.Named("state")

const (
	StateNone DataState = iota
	StateQueued
	StateDownloading
	StateError
	StateOK
	StateOKButMissing
)

type DataState uint8

type ScrollInfo struct {
	ScrollPos            int
	ScrollBrightenByUser float64
	ScrollCheckpoint     rvb.Checkpoint
}

type GopherMascot struct {
	Gopher      rl.Texture2D
	GopherEye   rl.Texture2D
	GopherHands rl.Texture2D
}

type FSMascot struct {
	FS        rl.Texture2D
	FSEyeIris rl.Texture2D
	FSEyeDot  rl.Texture2D
	FSEyeHl   rl.Texture2D
}

type Fonts struct {
	UI rl.Font

	Main     rl.Font
	MainMono rl.Font
}

type Textures struct {
	Gopher GopherMascot
	FS     FSMascot

	EliteLoader rl.Texture2D
	Terminator  rl.RenderTexture2D
	Cog         rl.RenderTexture2D
	LED         rl.RenderTexture2D
}

type ServerStatus struct {
	Err       error
	Timestamp time.Time
	Status    elite_status.ServerStatus
}

type InaraThings struct {
	CMDRInfoChan              chan string
	CMDRNamesStreamActive     atomic.Bool
	CMDRNamesRefreshAvailable atomic.Bool
	CMDRNamesStream           chan string
	CMDRInfoQueued            atomic.Int64

	CMDRInfo              cmap.ConcurrentMap[string, storage.InfoWithTime]
	CMDRInfoState         cmap.ConcurrentMap[string, DataState]
	CMDRTextures          cmap.ConcurrentMap[string, rl.Texture2D]
	CMDRTexturesState     cmap.ConcurrentMap[string, DataState]
	CMDRTexturesKeepAlive cmap.ConcurrentMap[string, int64] // unixNano
	CMDRInfoKeepAlive     cmap.ConcurrentMap[string, int64] // unixNano

	Powers      map[int]rl.Texture2D         // inapi powerID > texture
	Superpowers map[int]rl.Texture2D         // inapi superpowerID > texture
	Ranks       map[int]map[int]rl.Texture2D // inapi rankID > rankValue > texture
	InaraLogo   rl.Texture2D

	AvatarPlaceholderCMDRQuestion  rl.Texture2D
	AvatarPlaceholderCMDRHead      rl.Texture2D
	AvatarPlaceholderCMDRShoulders rl.Texture2D
}

type Viewport struct {
	BufferSize int
	Buffer     *rvb.RVBuffer[journal.MessageEvent]
	Scroll     *ScrollInfo
}

func (vp *Viewport) ResetBuffer() {
	vp.Buffer = rvb.NewBuffer[journal.MessageEvent](vp.BufferSize)
}

type TheHolyGlobalState struct {
	Dev bool

	SessionCounter AtomicInt

	CurrentThings CurrentThings

	Inara InaraThings

	ServerStatusRVB           *rvb.RVBuffer[ServerStatus]
	ServerStatusTimestampsRVB *rvb.RVBuffer[time.Time]

	FriendList cmap.ConcurrentMap[string, bool]
	EnemyList  cmap.ConcurrentMap[string, bool]

	LogsRVB *rvb.RVBuffer[string]

	ForceShowPing bool

	// these will be handled at runtime
	Tasks                     Jobs
	BubbleMessage             BubbleMessage
	Stats                     Stats
	Fonts                     Fonts
	Textures                  Textures
	ExportStateEvents         DataState
	ExportStateNPCMessages    DataState
	ExportStateNotifications  DataState
	UpdateCheckLast           AtomicValue[time.Time]
	UpdateCheckState          AtomicValue[DataState]
	UpdateCheckVersion        AtomicValue[version.Version]
	UpdateCheckReleaseTime    AtomicValue[time.Time]
	ShowPingHover             bool
	AudioReloadRequired       bool
	LogsVerticalPeek          float32
	ActivelyGrabbingStatusbar bool
	ProcessingLiveEvents      atomic.Bool
	ProcessingLiveEventsCond  sync.Cond

	// this section have to be constructed from already loaded config
	ViewportBuffers          map[int]*Viewport
	WordListEditField        string
	WordList                 map[string]bool
	TimestampFormatEditField string
}

type Jobs struct {
	ServerStatus  *task.TaskManager
	UpdateChecker *task.TaskManager
	Inara         *task.TaskManager
	Wakeup        *task.TaskManager
	Journal       *task.TaskManager

	GenJsons         *task.TaskManager
	GenNPCMessages   *task.TaskManager
	GenNotifications *task.TaskManager
}

type BubbleMessage struct {
	EventTime time.Time
	MousePos  rl.Vector2
	Text      string
}

func (b *BubbleMessage) SetMessage(text string, currentMousePos rl.Vector2) {
	b.EventTime = time.Now()
	b.MousePos = currentMousePos
	b.Text = text
}

type AtomicValue[T any] struct {
	v        atomic.Value
	modified atomic.Int64
}

func (a *AtomicValue[T]) Load() T {
	val := a.v.Load()
	if val == nil {
		var zero T
		return zero
	}
	return val.(T)
}

func (a *AtomicValue[T]) Modified() time.Time {
	return time.UnixMilli(a.modified.Load())
}

func (a *AtomicValue[T]) Store(val T) {
	a.v.Store(val)
	a.modified.Store(time.Now().UnixMilli())
}

func (a *AtomicValue[T]) Swap(new T) (old T) {
	val := a.v.Swap(new)
	a.modified.Store(time.Now().UnixMilli())
	if val == nil {
		var zero T
		return zero
	}
	return val.(T)

}

type AtomicInt struct {
	v        atomic.Int64
	modified atomic.Int64
}

func (s *AtomicInt) Add(value int) {
	s.v.Add(int64(value))
	s.modified.Store(time.Now().UnixMilli())
}

func (s *AtomicInt) Sub(value int) {
	s.v.Add(int64(-value))
	s.modified.Store(time.Now().UnixMilli())
}

func (s *AtomicInt) Reset() {
	s.v.Store(0)
	s.modified.Store(time.Now().UnixMilli())
}

func (s *AtomicInt) Store(value int) {
	s.v.Store(int64(value))
	s.modified.Store(time.Now().UnixMilli())
}

func (s *AtomicInt) Load() int {
	return int(s.v.Load())
}

func (s *AtomicInt) Modified() time.Time {
	return time.UnixMilli(s.modified.Load())
}

type AtomicDuration struct {
	v        atomic.Int64
	modified atomic.Int64
}

func (s *AtomicDuration) Add(value time.Duration) {
	s.v.Add(int64(value))
	s.modified.Store(time.Now().UnixMilli())
}

func (s *AtomicDuration) Sub(value time.Duration) {
	s.v.Add(int64(-value))
	s.modified.Store(time.Now().UnixMilli())
}

func (s *AtomicDuration) Reset() {
	s.v.Store(0)
	s.modified.Store(time.Now().UnixMilli())
}

func (s *AtomicDuration) Store(value time.Duration) {
	s.v.Store(int64(value))
	s.modified.Store(time.Now().UnixMilli())
}

func (s *AtomicDuration) Load() time.Duration {
	return time.Duration(s.v.Load())
}

func (s *AtomicDuration) Modified() time.Time {
	return time.UnixMilli(s.modified.Load())
}

type Resettable interface {
	Reset()
}

type Stats struct {
	CreditsNetIncome AtomicInt
	CreditsExpenses  AtomicInt
	CreditsIncome    AtomicInt
	CommoditySell    AtomicInt
	CommodityBuy     AtomicInt
	CommodityNet     AtomicInt
	Deaths           AtomicInt
	DeathsByPlayer   AtomicInt
	KillsCombatZone  AtomicInt
	KillsNPC         AtomicInt
	KillsPlayer      AtomicInt
	Jumps            AtomicInt
	CrewWage         AtomicInt
	Dockings         AtomicInt
	SignalsDetected  AtomicInt

	ProducedMessages  AtomicInt
	ProcessedEvents   AtomicInt
	UnprocessedEvents AtomicInt

	SessionsSolo  AtomicInt
	SessionsOpen  AtomicInt
	SessionsGroup AtomicInt

	PlaytimeSolo  AtomicDuration
	PlaytimeOpen  AtomicDuration
	PlaytimeGroup AtomicDuration
}

func (s *Stats) ResetAll() {
	val := reflect.ValueOf(s).Elem()
	typ := val.Type()

	var successes int

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		if !field.CanAddr() {
			continue
		}

		ptr := field.Addr().Interface()

		resettable, ok := ptr.(Resettable)
		if !ok {
			continue
		}
		resettable.Reset()
		fieldName := typ.Field(i).Name
		Logger.Debug("stat reset", zap.String("stat", fieldName))
		successes++
	}
	if successes == 0 {
		Logger.Error("no stats were reset", zap.Int("count", successes))
	} else {
		Logger.Debug("stats reset", zap.Int("count", successes))
	}
}

type CurrentThings struct {
	LogFile      AtomicValue[string]
	CMDR         AtomicValue[string]
	FID          AtomicValue[string]
	Music        AtomicValue[string]
	Mode         AtomicValue[string]
	GameLoadedAt AtomicValue[time.Time]

	LocationSystem     AtomicValue[string]
	LocationBody       AtomicValue[string]
	LocationBodyType   AtomicValue[string]
	LocationPopulation AtomicValue[string]
}

func InitializeTheHolyGlobalState(config *config.EDCITRConfig) *TheHolyGlobalState {
	bufferSize := config.Viewports.BufferSize
	if bufferSize <= 0 {
		bufferSize = 1024 * 1024
	}

	state := TheHolyGlobalState{
		Dev: true,

		SessionCounter: AtomicInt{},

		CurrentThings: CurrentThings{
			LogFile:      AtomicValue[string]{},
			CMDR:         AtomicValue[string]{},
			FID:          AtomicValue[string]{},
			Music:        AtomicValue[string]{},
			Mode:         AtomicValue[string]{},
			GameLoadedAt: AtomicValue[time.Time]{},

			LocationSystem:     AtomicValue[string]{},
			LocationBody:       AtomicValue[string]{},
			LocationBodyType:   AtomicValue[string]{},
			LocationPopulation: AtomicValue[string]{},
		},

		Inara: InaraThings{
			CMDRInfoChan:              make(chan string, 16),
			CMDRNamesRefreshAvailable: atomic.Bool{},
			CMDRNamesStreamActive:     atomic.Bool{},
			CMDRNamesStream:           make(chan string, 16),
			CMDRInfoQueued:            atomic.Int64{},

			CMDRInfo:              cmap.New[storage.InfoWithTime](),
			CMDRInfoState:         cmap.New[DataState](),
			CMDRTextures:          cmap.New[rl.Texture2D](),
			CMDRTexturesState:     cmap.New[DataState](),
			CMDRTexturesKeepAlive: cmap.New[int64](),
			CMDRInfoKeepAlive:     cmap.New[int64](),
			//CMDRTexturesLastDrawnFrame: cmap.New[int](),

			// These will be handled at runtime when safe to do so
			Powers:                         nil,
			Superpowers:                    nil,
			Ranks:                          nil,
			InaraLogo:                      rl.Texture2D{},
			AvatarPlaceholderCMDRQuestion:  rl.Texture2D{},
			AvatarPlaceholderCMDRHead:      rl.Texture2D{},
			AvatarPlaceholderCMDRShoulders: rl.Texture2D{},
		},

		FriendList: cmap.New[bool](),
		EnemyList:  cmap.New[bool](),

		LogsRVB: rvb.NewBuffer[string](1000),

		ForceShowPing: false,

		// these will be handled at runtime
		Tasks:                     Jobs{},
		ServerStatusRVB:           nil,
		ServerStatusTimestampsRVB: nil,
		BubbleMessage:             BubbleMessage{},
		Stats:                     Stats{},
		Fonts:                     Fonts{},
		Textures:                  Textures{},
		ProcessingLiveEvents:      atomic.Bool{},
		ExportStateEvents:         0,
		ExportStateNPCMessages:    0,
		ExportStateNotifications:  0,
		UpdateCheckState:          AtomicValue[DataState]{},
		UpdateCheckLast:           AtomicValue[time.Time]{},
		UpdateCheckVersion:        AtomicValue[version.Version]{},
		ShowPingHover:             false,
		AudioReloadRequired:       false,
		LogsVerticalPeek:          0,
		ActivelyGrabbingStatusbar: false,

		// this section have to be constructed from already loaded config
		ViewportBuffers:          nil,
		WordListEditField:        "",
		WordList:                 nil,
		TimestampFormatEditField: "",
	}

	state.ViewportBuffers = make(map[int]*Viewport, len(config.Viewports.ViewportChannelBitmask))
	for k, v := range config.Viewports.ViewportChannelBitmask {
		if v != journal.ChannelUndefined {
			state.ViewportBuffers[k] = &Viewport{
				BufferSize: bufferSize,
				Buffer:     rvb.NewBuffer[journal.MessageEvent](bufferSize),
				Scroll:     &ScrollInfo{},
			}
		} else {
			state.ViewportBuffers[k] = &Viewport{
				BufferSize: 1,
				Buffer:     rvb.NewBuffer[journal.MessageEvent](1),
				Scroll:     &ScrollInfo{},
			}
		}
	}

	state.WordList = make(map[string]bool, len(config.Viewports.Appearance.HighlightWords))
	for _, word := range config.Viewports.Appearance.HighlightWords {
		state.WordList[word] = true
	}
	state.WordListEditField = strings.Join(config.Viewports.Appearance.HighlightWords, ",")
	state.TimestampFormatEditField = config.Viewports.Appearance.TimestampFormat

	return &state
}

// LoadInaraGraphics loads inara-related textures into GPU, populates them into HolyGlobalState.Inara for access
func (s *TheHolyGlobalState) LoadInaraGraphics() {
	s.Inara.Ranks = make(map[int]map[int]rl.Texture2D, 6)
	for rankID := 1; rankID <= 6; rankID++ {
		s.Inara.Ranks[rankID] = make(map[int]rl.Texture2D, 14)
		for rankValue := 0; rankValue <= 13; rankValue++ {
			data := utils.UnsafeLoadData(assets.GraphicsFS, fmt.Sprintf("graphics/inara/ranks/%d/%d.png", rankID, rankValue))
			img := rl.LoadImageFromMemory(".png", data, int32(len(data)))
			s.Inara.Ranks[rankID][rankValue] = rl.LoadTextureFromImage(img)
			rl.SetTextureFilter(s.Inara.Ranks[rankID][rankValue], rl.FilterBilinear)
			rl.UnloadImage(img)
		}
	}
	for powerID := 12; powerID <= 15; powerID++ {
		s.Inara.Ranks[powerID] = make(map[int]rl.Texture2D, 1)
		data := utils.UnsafeLoadData(assets.GraphicsFS, fmt.Sprintf("graphics/inara/ranks/%d/0.png", powerID))
		img := rl.LoadImageFromMemory(".png", data, int32(len(data)))
		s.Inara.Ranks[powerID][0] = rl.LoadTextureFromImage(img)
		rl.SetTextureFilter(s.Inara.Ranks[powerID][0], rl.FilterBilinear)
		rl.UnloadImage(img)
	}

	s.Inara.Superpowers = make(map[int]rl.Texture2D, 4)
	for powerID := 1; powerID <= 4; powerID++ {
		data := utils.UnsafeLoadData(assets.GraphicsFS, fmt.Sprintf("graphics/inara/superpowers/%d.png", powerID))
		img := rl.LoadImageFromMemory(".png", data, int32(len(data)))
		s.Inara.Superpowers[powerID] = rl.LoadTextureFromImage(img)
		rl.SetTextureFilter(s.Inara.Superpowers[powerID], rl.FilterBilinear)
		rl.UnloadImage(img)
	}

	s.Inara.Powers = make(map[int]rl.Texture2D, 13)
	for powerID := 1; powerID <= 13; powerID++ {
		data := utils.UnsafeLoadData(assets.GraphicsFS, fmt.Sprintf("graphics/inara/powers/%d.png", powerID))
		img := rl.LoadImageFromMemory(".png", data, int32(len(data)))
		s.Inara.Powers[powerID] = rl.LoadTextureFromImage(img)
		rl.SetTextureFilter(s.Inara.Powers[powerID], rl.FilterBilinear)
		rl.UnloadImage(img)
	}

	data := utils.UnsafeLoadData(assets.GraphicsFS, "graphics/inara/inara.png")
	img := rl.LoadImageFromMemory(".png", data, int32(len(data)))
	s.Inara.InaraLogo = rl.LoadTextureFromImage(img)
	rl.SetTextureFilter(s.Inara.InaraLogo, rl.FilterBilinear)
	rl.UnloadImage(img)

	data = utils.UnsafeLoadData(assets.GraphicsFS, "graphics/inara/avatars/default-question.png")
	img = rl.LoadImageFromMemory(".png", data, int32(len(data)))
	s.Inara.AvatarPlaceholderCMDRQuestion = rl.LoadTextureFromImage(img)
	rl.SetTextureFilter(s.Inara.AvatarPlaceholderCMDRQuestion, rl.FilterBilinear)
	rl.UnloadImage(img)

	data = utils.UnsafeLoadData(assets.GraphicsFS, "graphics/inara/avatars/default-head.png")
	img = rl.LoadImageFromMemory(".png", data, int32(len(data)))
	s.Inara.AvatarPlaceholderCMDRHead = rl.LoadTextureFromImage(img)
	rl.SetTextureFilter(s.Inara.AvatarPlaceholderCMDRHead, rl.FilterBilinear)
	rl.UnloadImage(img)

	data = utils.UnsafeLoadData(assets.GraphicsFS, "graphics/inara/avatars/default-shoulders.png")
	img = rl.LoadImageFromMemory(".png", data, int32(len(data)))
	s.Inara.AvatarPlaceholderCMDRShoulders = rl.LoadTextureFromImage(img)
	rl.SetTextureFilter(s.Inara.AvatarPlaceholderCMDRShoulders, rl.FilterBilinear)
	rl.UnloadImage(img)
}

// UnloadInaraGraphics unloads inara-related textures from GPU, clearing all references in HolyGlobalState.Inara
func (s *TheHolyGlobalState) UnloadInaraGraphics() {
	for rankID := 1; rankID <= 6; rankID++ {
		for rankValue := 0; rankValue <= 13; rankValue++ {
			rl.UnloadTexture(s.Inara.Ranks[rankID][rankValue])
		}
	}
	s.Inara.Ranks = nil
	for powerID := 12; powerID <= 15; powerID++ {
		rl.UnloadTexture(s.Inara.Ranks[powerID][0])
	}
	s.Inara.Ranks = nil

	for powerID := 1; powerID <= 4; powerID++ {
		rl.UnloadTexture(s.Inara.Superpowers[powerID])
	}
	s.Inara.Superpowers = nil

	for powerID := 1; powerID <= 13; powerID++ {
		rl.UnloadTexture(s.Inara.Powers[powerID])
	}
	s.Inara.Powers = nil
}

// LoadGeneralGraphics loads mascots and loader textures into GPU, populates them into HolyGlobals.Textures for access
// additionally pre-rendering HolyGlobals.Textures.Terminator
func (s *TheHolyGlobalState) LoadGeneralGraphics() {
	type TextureInfo struct {
		path   string
		dest   *rl.Texture2D
		filter rl.TextureFilterMode
	}

	for _, ti := range []TextureInfo{ // TODO: avoid duplication?
		{
			path:   "graphics/mascots/gopher/gopher.png",
			dest:   &s.Textures.Gopher.Gopher,
			filter: rl.FilterPoint,
		},
		{
			path:   "graphics/mascots/gopher/gopher-eye.png",
			dest:   &s.Textures.Gopher.GopherEye,
			filter: rl.FilterPoint,
		},
		{
			path:   "graphics/mascots/gopher/gopher-raylib.png",
			dest:   &s.Textures.Gopher.GopherHands,
			filter: rl.FilterBilinear,
		},
		{
			path:   "graphics/mascots/fs/fs.png",
			dest:   &s.Textures.FS.FS,
			filter: rl.FilterPoint,
		},
		{
			path:   "graphics/mascots/fs/fs-eye2-iris.png",
			dest:   &s.Textures.FS.FSEyeIris,
			filter: rl.FilterBilinear,
		},
		{
			path:   "graphics/mascots/fs/fs-eye3-dot.png",
			dest:   &s.Textures.FS.FSEyeDot,
			filter: rl.FilterBilinear,
		},
		{
			path:   "graphics/mascots/fs/fs-eye4-hl.png",
			dest:   &s.Textures.FS.FSEyeHl,
			filter: rl.FilterBilinear,
		},
		{
			path:   "graphics/loader/EDLoader_1-30.png",
			dest:   &s.Textures.EliteLoader,
			filter: rl.FilterBilinear,
		},
	} {
		data := utils.UnsafeLoadData(assets.GraphicsFS, ti.path)
		img := rl.LoadImageFromMemory(".png", data, int32(len(data)))
		*ti.dest = rl.LoadTextureFromImage(img)
		rl.SetTextureFilter(*ti.dest, ti.filter)
		rl.UnloadImage(img)
	}

	// pre-render HolyGlobals.Textures.Terminator
	s.Textures.Terminator = rl.LoadRenderTexture(60, 200)
	rl.BeginTextureMode(s.Textures.Terminator)
	for i := 0; i < 8; i++ {
		rl.DrawRectanglePro(
			rl.Rectangle{-180 + 60*float32(i), float32(s.Textures.Terminator.Texture.Height) + 20, 20, 400},
			rl.Vector2{0, 0},
			180+45,
			rl.White,
		)
	}
	rl.EndTextureMode()
	rl.SetTextureFilter(s.Textures.Terminator.Texture, rl.FilterBilinear)

	// pre-render settings cog
	// --- Configuration ---
	width := int32(300)
	height := int32(300)
	// Use float32 explicitly for Vector2
	cx := float32(width) / 2
	cy := float32(height) / 2

	// Gear Geometry
	innerRadius := float32(100) // Hub
	outerRadius := float32(140) // Tooth Tip
	holeRadius := float32(50)   // Axle hole

	// Tooth Shape (Trapezoid)
	numTeeth := 6
	baseHalfWidth := float64(20) // Degrees (half of 40)
	tipHalfWidth := float64(10)  // Degrees (half of 20) - smaller = more tapered

	s.Textures.Cog = rl.LoadRenderTexture(width, height)

	rl.BeginTextureMode(s.Textures.Cog)
	rl.ClearBackground(rl.Blank)

	// 1. Draw the Hub
	rl.DrawCircleV(rl.NewVector2(cx, cy), innerRadius, rl.White)

	// 2. Draw the Teeth
	for i := 0; i < numTeeth; i++ {
		// Current tooth center angle
		angleDeg := float64(i) * 360.0 / float64(numTeeth)

		// Calculate radians for the 4 points
		// We convert (Angle +/- Width) to radians
		radBaseLeft := (angleDeg - baseHalfWidth) * (math.Pi / 180.0)
		radTipLeft := (angleDeg - tipHalfWidth) * (math.Pi / 180.0)
		radTipRight := (angleDeg + tipHalfWidth) * (math.Pi / 180.0)
		radBaseRight := (angleDeg + baseHalfWidth) * (math.Pi / 180.0)

		// Calculate Coordinates
		// P1: Base Left (Inner)
		p1 := rl.NewVector2(
			cx+float32(math.Cos(radBaseLeft))*innerRadius,
			cy+float32(math.Sin(radBaseLeft))*innerRadius,
		)
		// P2: Tip Left (Outer)
		p2 := rl.NewVector2(
			cx+float32(math.Cos(radTipLeft))*outerRadius,
			cy+float32(math.Sin(radTipLeft))*outerRadius,
		)
		// P3: Tip Right (Outer)
		p3 := rl.NewVector2(
			cx+float32(math.Cos(radTipRight))*outerRadius,
			cy+float32(math.Sin(radTipRight))*outerRadius,
		)
		// P4: Base Right (Inner)
		p4 := rl.NewVector2(
			cx+float32(math.Cos(radBaseRight))*innerRadius,
			cy+float32(math.Sin(radBaseRight))*innerRadius,
		)

		// Draw the Trapezoid (2 Triangles)
		// We draw them in standard order AND reverse order to prevent invisible back-faces

		// Left side of tooth
		rl.DrawTriangle(p1, p2, p3, rl.White) // Normal
		rl.DrawTriangle(p3, p2, p1, rl.White) // Reverse (Back-up)

		// Right side of tooth
		rl.DrawTriangle(p1, p3, p4, rl.White) // Normal
		rl.DrawTriangle(p4, p3, p1, rl.White) // Reverse (Back-up)
	}

	// 3. Cut the Hole
	rl.SetBlendMode(rl.BlendCustom)
	rl.SetBlendFactors(rl.Zero, rl.Zero, rl.FuncAdd)
	rl.DrawCircleV(rl.NewVector2(cx, cy), holeRadius, rl.White)

	// Cleanup
	rl.SetBlendMode(rl.BlendAlpha)
	rl.EndTextureMode()
	rl.SetTextureFilter(s.Textures.Cog.Texture, rl.FilterBilinear)

	// pre-render LED
	s.Textures.LED = rl.LoadRenderTexture(300, 300)
	rl.BeginTextureMode(s.Textures.LED)
	var color rl.Color
	for i := 0; i < 10; i++ {
		progress := float64(i) / 10
		color = rl.ColorAlpha(rl.Black, float32(progress))
		rl.DrawCircle(
			s.Textures.LED.Texture.Width/2,
			s.Textures.LED.Texture.Height/2,
			float32(s.Textures.LED.Texture.Width)/2-float32(i),
			color,
		)
	}
	color = rl.Black
	for i := 0; i < 10; i++ {
		rl.DrawCircle(
			s.Textures.LED.Texture.Width/2,
			s.Textures.LED.Texture.Height/2,
			float32(s.Textures.LED.Texture.Width)/2-10-float32(i),
			color,
		)
	}
	for i := 0; i < 10; i++ {
		progress := float64(i) / 10
		color = rl.ColorFromHSV(0, 0, float32(utils.EaseOutExpo(progress)/4))
		rl.DrawCircle(
			s.Textures.LED.Texture.Width/2,
			s.Textures.LED.Texture.Height/2,
			float32(s.Textures.LED.Texture.Width)/2-20-float32(i),
			color,
		)
	}
	for i := 0; i < 280; i++ {
		progress := float64(i) / 280
		color = rl.ColorFromHSV(0, 0, float32(utils.EaseOutExpo(progress)))
		rl.DrawCircle(
			s.Textures.LED.Texture.Width/2,
			s.Textures.LED.Texture.Height/2,
			float32(s.Textures.LED.Texture.Width)/2-30-float32(i),
			color,
		)
	}
	rl.EndTextureMode()
	rl.SetTextureFilter(s.Textures.LED.Texture, rl.FilterBilinear)
}

// UnloadGeneralGraphics unloads inara-related textures from GPU, clearing all references in HolyGlobals.Inara
func (s *TheHolyGlobalState) UnloadGeneralGraphics() {
	for _, dest := range []*rl.Texture2D{ // TODO: avoid duplication from LoadGeneralGraphics?
		&s.Textures.Gopher.Gopher,
		&s.Textures.Gopher.GopherEye,
		&s.Textures.Gopher.GopherHands,
		&s.Textures.FS.FS,
		&s.Textures.FS.FSEyeIris,
		&s.Textures.FS.FSEyeDot,
		&s.Textures.FS.FSEyeHl,
		&s.Textures.EliteLoader,
	} {
		rl.UnloadTexture(*dest)
		*dest = rl.Texture2D{}
	}
	rl.UnloadRenderTexture(s.Textures.Terminator)
	s.Textures.Terminator = rl.RenderTexture2D{}
	rl.UnloadRenderTexture(s.Textures.Cog)
	s.Textures.Cog = rl.RenderTexture2D{}
	rl.UnloadRenderTexture(s.Textures.LED)
	s.Textures.LED = rl.RenderTexture2D{}
}
