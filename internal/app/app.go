package app

import (
	"bytes"
	"errors"
	"image/png"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gen2brain/raylib-go/raygui"
	"github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/assets"
	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/app/tasks"
	"github.com/gethiox/EDCITR/internal/ascii"
	"github.com/gethiox/EDCITR/internal/config"
	"github.com/gethiox/EDCITR/internal/inara"
	"github.com/gethiox/EDCITR/internal/noire"
	"github.com/gethiox/EDCITR/internal/parser"
	"github.com/gethiox/EDCITR/internal/storage"
	"github.com/gethiox/EDCITR/internal/task"
	"github.com/gethiox/EDCITR/internal/utils"
	"github.com/gethiox/EDCITR/internal/version"
	"github.com/gethiox/EDCITR/internal/zaputils"
	"github.com/spf13/pflag"
	"github.com/teacat/noire"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	zaputils.SetLevel(zapcore.InfoLevel)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.SetTraceLogCallback(TraceLogCallbackFun)

	codepoints = make([]rune, 0)
	for _, cpRange := range codepointRanges {
		for cp := cpRange[0]; cp <= cpRange[1]; cp++ {
			codepoints = append(codepoints, cp)
		}
	}
}

const (
	UIFontSize    = 24
	UIFontSpacing = 1
	MaxSounds     = 8
	RenderPad     = 128 // additional space for RenderTexture (avoiding shader blur edge artifacts)

	DefaultScreenWidth  = 1280
	DefaultScreenHeight = 720
	MinimumScreenWidth  = 1280
	MinimumScreenHeight = 720

	asciiArtWidth  = 88
	asciiArtHeight = 32
)

var (
	profileCatchup = pflag.Bool("profile-catchup", false, "generate cpu profile file for catchup process")
)

var (
	codepointRanges = [][2]rune{ // from, to, including both values
		{0x20, 0x7e}, // ASCII: ' '-`~`

		// partial Latin-1 Supplement: ' '-'¿' (NO-BREAK SPACE, INVERTED QUESTION MARK)
		// excluding SOFT HYPHEN (0xad)
		{0xa0, 0xac},
		{0xae, 0xbf},

		//{0x2580, 0x259F}, // Block Elements
	}

	codepoints []rune

	Logger   = zaputils.Logger.Named(state.APIAppName)
	RlLogger = zaputils.LoggerWithoutCaller.Named("raylib")

	ColorUI       = my_noire.Color{noire.NewHex("#d4d4d4ff")}
	ColorUIHLFdev = my_noire.Color{noire.NewHex("#EB891Dff")}
	GuiNormalBg   = rl.Color{0x00, 0x00, 0x00, 0xcc}
)

func initColorSettingsHSV(colorSettings []*state.ColorSetting) {
	for _, cs := range colorSettings {
		if cs.Color == nil {
			continue
		}
		cs.UpdateHSV()
	}
}

type App struct {
	Config config.EDCITRConfig
	State  state.TheHolyGlobalState

	Render *Renderer

	DefaultConfigReadOnly config.EDCITRConfig
	ProcessedConfig       *config.EDCITRConfig
	SavedConfig           *config.EDCITRConfig

	ConfigStore *storage.Config
	CacheStore  *storage.Cache
	JsonStore   *storage.JsonStore

	PFPDownloader  *inara.PFPDownloader
	PFPResultsChan <-chan inara.DownloadImageResult
	InaraClient    *inara.Client

	startedAt time.Time
}

func NewApp() *App {
	a := &App{
		DefaultConfigReadOnly: *config.GetDefaultConfig(),
	}
	a.Render = NewRenderer(a)
	return a
}

func (a *App) Run() {
	var err error

	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	a.startedAt = time.Now()
	Logger.Info("Starting application", zap.String("runtime", runtime.Version()), zap.String("version", version.Release.FormatFull()), zap.Strings("args", args))
	Logger.Info("Codepoints", zap.Int("count", len(codepoints)))

	img, err := png.Decode(bytes.NewReader(assets.AppIconData))
	if err == nil {
		lines, err := ascii.RenderImage(img, ascii.ScalerCatmullRom, ascii.RT1PxPerCell, asciiArtWidth, asciiArtHeight)
		if err == nil {
			for row, line := range lines {
				line = strings.Replace(line, "█", "#", -1)
				Logger.Info(line, zap.Int("row", row+1))
			}

		}
	}

	Logger.Info("Initializing config and cache..")
	a.ConfigStore, err = storage.GetConfig(state.StoreAppName)
	if err != nil {
		Logger.Fatal("Initializing config storage failed", zap.Error(err))
	}

	a.CacheStore, err = storage.GetCache(state.StoreAppName)
	if err != nil {
		Logger.Fatal("Initializing cache storage failed", zap.Error(err))
	}

	a.JsonStore, err = storage.GetJsonStore()
	if err != nil {
		Logger.Fatal("Initializing json store failed", zap.Error(err))
	}

	tmpConf, err := a.ConfigStore.LoadConfig()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			Logger.Warn("Config file doesn't exist, using default")
			tmpConf = a.DefaultConfigReadOnly.DeepCopy()
		} else {
			Logger.Fatal("Unexpected error occurred during loading config", zap.Error(err))
		}
	}
	a.Config = *tmpConf
	a.SavedConfig = a.Config.DeepCopy()
	a.ProcessedConfig = a.Config.DeepCopy()

	Logger.Info("Config Loaded:")
	config.LogConfig(Logger, &a.Config)

	a.afterConfigLoadInit()

	Logger.Info("Raylib Window Initialization")
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagWindowTransparent)
	rl.InitWindow(DefaultScreenWidth, DefaultScreenHeight, "ED: Chat Is This Real")
	defer rl.CloseWindow()

	rl.SetWindowMinSize(MinimumScreenWidth, MinimumScreenHeight)
	rl.SetTargetFPS(int32(a.Config.FPS))

	rl.BeginDrawing()
	rl.ClearBackground(rl.Blank)
	//rl.SetConfigFlags()

	rl.EndDrawing()

	Logger.Debug("Setting Window Icon")
	windowIcon := rl.LoadImageFromMemory(".png", assets.AppIconData, int32(len(assets.AppIconData)))
	rl.SetWindowIcon(*windowIcon)
	rl.UnloadImage(windowIcon)

	Logger.Info("Raylib Window Initialization Completed")

	raygui.LoadStyleFromMemory(utils.UnsafeLoadData(assets.StylesFS, "styles/Elite.rgs"))
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_SIZE, UIFontSize)
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_SPACING, UIFontSpacing)

	a.State.Fonts.UI = utils.UnsafeLoadFont(assets.FontsFS, "fonts/Dosis-SpaceMono-Bold.ttf", UIFontSize, codepoints)
	raygui.SetFont(a.State.Fonts.UI)

	a.Render.s.CanvasSizeX, a.Render.s.CanvasSizeY = rl.GetRenderWidth(), rl.GetRenderHeight()

	a.State.Fonts.Main = utils.UnsafeLoadFont(assets.FontsFS, "fonts/Dosis-Bold.ttf", int32(a.Config.Fonts.Size), codepoints)
	a.State.Fonts.MainMono = utils.UnsafeLoadFont(assets.FontsFS, "fonts/Dosis-SpaceMono-Bold.ttf", int32(a.Config.Fonts.Size), codepoints)

	a.PFPDownloader, a.PFPResultsChan = inara.GetPFPDownloader(8)

	a.InaraClient = inara.GetClient(time.Second * 10)

	if a.Config.InaraSettings.Enabled {
		cmdr, fid, ok := pickupRecentCMDrAndFID()
		if !ok {
			Logger.Warn("CMDR and FID not found, awaiting full login")
		} else {
			Logger.Info("CMDR and FID found", zap.String("CMDR", cmdr), zap.String("FID", fid))
			err := a.InaraClient.SetHeader(
				inara.NewHeader(version.Release, state.APIAppName, a.Config.InaraSettings.Token, cmdr, fid),
			)
			if err != nil {
				Logger.Error("Failed to set Inara client header", zap.Error(err))
			}
			err = a.PFPDownloader.SetHeader(state.APIAppName, version.Release)
			if err != nil {
				Logger.Error("Failed to set PFPDownloader header", zap.Error(err))
			}
		}
	}

	Logger.Info("Initializing Audio Device")
	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()

	a.loadSoundList()
	a.loadSounds()

	a.prepareJobs()

	a.State.LoadGeneralGraphics()
	defer a.State.UnloadGeneralGraphics()

	a.State.LoadInaraGraphics()
	defer a.State.UnloadInaraGraphics()

	a.Render.windowLoop()
}

func (a *App) afterConfigLoadInit() {
	Logger.Debug("Initializing color settings HSV values")
	initColorSettingsHSV(a.Render.colorSettings)

	Logger.Debug("Initializing The Holy Global State...", zap.Int("buffer_size", a.Config.Viewports.BufferSize))
	a.State = *state.InitializeTheHolyGlobalState(&a.Config)

	zaputils.ConnectLogBuffer(a.State.LogsRVB)
	Logger.Debug("Log buffer attached")

	a.Render.s.UnloadPeriod.Store(a.Config.InaraSettings.UnloadPeriod.Nanoseconds())
}

func (a *App) prepareJobs() {
	generateEvents := &tasks.GenerateEvents{
		JsonStore: a.JsonStore,
		Logger:    Logger,
		Config:    &a.Config,
		State:     &a.State.ExportStateEvents,
	}

	generateNotifications := &tasks.GenerateNotifications{
		JsonStore: a.JsonStore,
		Logger:    Logger,
		Config:    &a.Config,
		State:     &a.State.ExportStateNotifications,
	}

	generateNPCMessages := &tasks.GenerateNPCMessages{
		JsonStore: a.JsonStore,
		Logger:    Logger,
		Config:    &a.Config,
		State:     &a.State.ExportStateNPCMessages,
	}

	inaraService := &tasks.InaraService{
		Config:         &a.Config,
		State:          &a.State,
		InaraClient:    a.InaraClient,
		UIState:        &a.Render.s,
		CacheStore:     a.CacheStore,
		PFPDownloader:  a.PFPDownloader,
		PFPResultsChan: a.PFPResultsChan,
		Logger:         Logger,
	}

	serverStatus := &tasks.ServerStatus{
		Logger: Logger,
		Config: &a.Config,
		State:  &a.State,
	}

	wakeupDetector := &tasks.WakeupDetector{
		Logger: Logger,
		State:  &a.State,
	}

	checkForUpdates := &tasks.CheckForUpdates{
		Logger:     Logger,
		State:      &a.State,
		RepoOwner:  state.RepoOwner,
		RepoName:   state.RepoName,
		APIAppName: state.APIAppName,
	}

	journalProcessing := tasks.NewJournalProcessingJob(
		Logger,
		&a.State,
		&a.Render.s,
		&a.Config,
		a.ProcessedConfig,
		a.CacheStore,
		a.InaraClient,
		a.PFPDownloader,
		MaxSounds,
		*profileCatchup,
	)

	parserSettings := parser.Settings{
		MaxConcurrentFiles:      a.Config.Parser.MaxConcurrentFiles,
		ProcessorWorkers:        a.Config.Parser.ProcessorWorkers,
		ProcessorWorkerChanSize: a.Config.Parser.ProcessorWorkerChanSize,
		ProcessorResultChanSize: a.Config.Parser.ProcessorResultChanSize,
	}
	parser.Configure(parserSettings)

	a.State.Tasks.Journal = task.NewTaskManager(journalProcessing)
	a.State.Tasks.Inara = task.NewTaskManager(inaraService)
	a.State.Tasks.ServerStatus = task.NewTaskManager(serverStatus)
	a.State.Tasks.UpdateChecker = task.NewTaskManager(checkForUpdates)
	a.State.Tasks.Wakeup = task.NewTaskManager(wakeupDetector)
	a.State.Tasks.GenJsons = task.NewTaskManager(generateEvents)
	a.State.Tasks.GenNotifications = task.NewTaskManager(generateNotifications)
	a.State.Tasks.GenNPCMessages = task.NewTaskManager(generateNPCMessages)

	// optional jobs
	if a.Config.InaraSettings.Enabled {
		a.State.Tasks.Inara.Start()
	}
	if a.Config.ServerStatus.Enabled {
		a.State.Tasks.ServerStatus.Start()
	}
	if a.Config.StatusBar.CheckForUpdates {
		a.State.Tasks.UpdateChecker.Start()
	}

	// required jobs
	a.State.Tasks.Wakeup.Start()
	a.State.Tasks.Journal.Start()
}
