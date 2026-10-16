package app

import (
	"context"
	"fmt"
	"math"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	my_noire "github.com/gethiox/EDCITR/internal/noire"
	parser2 "github.com/gethiox/EDCITR/internal/parser"
	"go.uber.org/zap"
)

func TraceLogCallbackFun(level int, message string) {
	switch rl.TraceLogLevel(level) {
	case rl.LogAll:
		RlLogger.Info(message)
	case rl.LogTrace:
		RlLogger.Error(message)
	case rl.LogDebug:
		RlLogger.Debug(message)
	case rl.LogInfo:
		RlLogger.Info(message)
	case rl.LogWarning:
		RlLogger.Warn(message)
	case rl.LogError:
		RlLogger.Error(message)
	case rl.LogFatal:
		RlLogger.Fatal(message)
	case rl.LogNone:
	default:
		message = "[unhandled loglevel] " + message
		RlLogger.Error(message)
	}
}

func getEdLoaderFrame() int {
	return int(math.Mod(float64(time.Now().UnixMilli())/1000*30, float64(EdLoaderFrames)))
}

func getRatio(x, y float32) float32 {
	return (x / x) / (x / y)
}

func howToResize(x, y float32) (float32, float32) {
	ratio := getRatio(x, y)

	if ratio >= 1 {
		return getRatio(y, x), 1
	}
	return 1, ratio
}

// resolveConfigPath accepts a raw string with potential "builtin:" prefix
// returns a clean path without a prefix and a boolean if given path had "builtin:" prefix.
func resolveConfigPath(s string) (path string, builtin bool) {
	if strings.HasPrefix(s, "builtin:") {
		builtin = true
		path = s[8:]
	} else {
		builtin = false
		path = s
	}

	return
}

func pickupRecentCMDrAndFID() (cmdr, fid string, ok bool) {
	entries, err := parser2.ListCurrentLogs()
	if err != nil {
		Logger.Error("Failed to list journal logs", zap.Error(err))
		return
	}

	if len(entries) == 0 {
		return
	}

	var filteredByAge = make([]parser2.EntryWithTime, 0, 16)
	var ageThreshold = time.Hour * 24 * 2
	now := time.Now()

	for _, entry := range entries {
		since := now.Sub(entry.Time)
		if since < 0 {
			Logger.Warn("Duration is negative", zap.Time("timestamp", entry.Time))
			continue
		}

		if since > ageThreshold {
			continue
		}

		filteredByAge = append(filteredByAge, entry)
	}

	slices.Reverse(filteredByAge) // read from newest to oldest

	ctx, cancel := context.WithCancel(context.Background())
	outputEntries := parser2.ReadLogs(filteredByAge, ctx)
	defer cancel()

	for _, entry := range outputEntries {
		for eventUntyped := range entry.Events {
			switch event := eventUntyped.(type) {
			case *parser2.Commander:
				cmdr = event.Name
				fid = event.FID
				ok = true
				return
			default:
				continue
			}
		}
	}
	return
}

func getCommitTime() (time.Time, bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return time.Time{}, false
	}

	for _, setting := range info.Settings {
		if setting.Key == "vcs.time" { // vcs.time is formatted as RFC3339
			commitTime, err := time.Parse(time.RFC3339, setting.Value)
			if err != nil {
				return time.Time{}, false
			}
			return commitTime, true
		}
	}
	return time.Time{}, false
}

func fmtDuration(d time.Duration) string {
	switch {
	case d < time.Second:
		return fmt.Sprintf("%5.1f ms", d.Seconds()*1000)
	case d < time.Minute:
		return fmt.Sprintf("%5.2f s ", d.Seconds())
	case d < time.Hour:
		return fmt.Sprintf("%5.2f m ", d.Minutes())
	default:
		return fmt.Sprintf("%5.2f h ", d.Hours())
	}
}

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}

func dimColorForOutdatedSession(color my_noire.Color, msgSession, currentSession int) my_noire.Color {
	if currentSession > msgSession {
		return color.Shade(0.2).Desaturate(0.2)
	}
	return color
}
