package config

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	my_noire "github.com/gethiox/EDCITR/internal/noire"
	"go.uber.org/zap"
)

// LogConfig prints the configuration in categorized sections
func LogConfig(logger *zap.Logger, cfg *EDCITRConfig) {
	Logger := logger.Named("Config")
	// 1. General Settings
	Logger.Info("General",
		zap.String("Version", cfg.Version.FormatFull()),
		zap.Int("BufferSize", cfg.Viewports.BufferSize),
		zap.Int("FPS", cfg.FPS),
	)

	Logger.Info("Fonts",
		zap.String("MainFont", cfg.Fonts.MainPath),
		zap.String("MonoFont", cfg.Fonts.MonoPath),
		zap.Int("FontSize", cfg.Fonts.Size),
		zap.Int("MainFontSpacing", cfg.Fonts.Spacing),
		zap.Int("MonoFontSpacing", cfg.Fonts.SpacingMono),
	)

	Logger.Info("details",
		zap.Int("MessageSpacing", cfg.Viewports.Appearance.MessageSpacing),
		zap.Int("LeftPad", cfg.Viewports.Appearance.LeftPad),
		zap.Int("RightPad", cfg.Viewports.Appearance.RightPad),
		zap.Int("MessageIndentPad", cfg.Viewports.Appearance.MessageIndentPad),
		zap.Int("MarkerWidth", cfg.Viewports.Appearance.MarkerWidth),
		zap.Float32("MarkerHoldTime", cfg.Viewports.Appearance.MarkerHoldTime),
		zap.Float32("MarkerFadeoutTime", cfg.Viewports.Appearance.MarkerFadeoutTime),
		zap.Strings("HighlightWords", cfg.Viewports.Appearance.HighlightWords),

		zap.Int("Count", cfg.Viewports.Count),
		zap.Int("BorderSize", cfg.Viewports.BorderSize),
	)

	// 2. Audio & Session
	Logger.Info("Audio & Session",
		zap.Float32("MasterVolume", cfg.Audio.MasterVolume),
		zap.Float32("WheelSpeed", cfg.WheelSpeed),
	)

	// 3. Toggles
	Logger.Info("Toggles",
		zap.Bool("PrefixCMDR", cfg.Viewports.Appearance.Toggles.PrefixCMDR),
		zap.Bool("PrefixNPC", cfg.Viewports.Appearance.Toggles.PrefixNPC),
		zap.Bool("EliteTime", cfg.Viewports.Appearance.Toggles.TimestampEliteFormat),
		zap.Bool("LocalTime", cfg.Viewports.Appearance.Toggles.Timestamps),
		zap.Bool("AlwaysShowTimestamps", cfg.Viewports.Appearance.Toggles.AlwaysShowTimestamps),
	)

	Logger.Info("Inara",
		zap.Bool("Enabled", cfg.InaraSettings.Enabled),
		zap.Bool("CMDRImages", cfg.InaraSettings.CMDRImages),
		zap.Bool("ProfileIndicator", cfg.InaraSettings.ProfileIndicator),
		zap.Duration("UnloadPeriod", cfg.InaraSettings.UnloadPeriod),
		zap.Duration("WaitBeforeRequest", cfg.InaraSettings.WaitBeforeRequest),
		zap.String("Token", cfg.InaraSettings.Token),
	)

	Logger.Info("Integrations",
		zap.Bool("Ping", cfg.ServerStatus.Enabled),
	)

	Logger.Info("Parser",
		zap.Int("PreloadLogs", cfg.Parser.PreloadNLogfiles),
		zap.String("ArchiveDirectory", cfg.Parser.ArchiveDirectory),

		zap.Int("MaxConcurrentFiles", cfg.Parser.MaxConcurrentFiles),
		zap.Int("ProcessorWorkers", cfg.Parser.ProcessorWorkers),
		zap.Int("ProcessorWorkerChanSize", cfg.Parser.ProcessorWorkerChanSize),
		zap.Int("ProcessorResultChanSize", cfg.Parser.ProcessorResultChanSize),
	)

	// 5. Colors (Using Reflection)
	// We use reflection here because mapping 50+ color fields manually is tedious.
	Logger.Info("Colors", zap.Int("count", reflect.ValueOf(cfg.Colors).NumField()))
	logColors(Logger, cfg.Colors)
}

// logColors Iterates over the Colors struct and logs them in perfectly aligned columns
func logColors(logger *zap.Logger, colors Colors) {
	val := reflect.ValueOf(colors)
	typ := val.Type()

	type colorEntry struct {
		Name string
		Hex  string
		R    float64
		G    float64
		B    float64
		A    float64
	}

	var entries []colorEntry

	// 1. Extract all colors
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name

		c := field.Interface().(my_noire.Color)
		r, g, b, a := c.RGBA()

		// Assuming noire.Hex() doesn't include the '#', if it does, remove the '#' here
		hex := c.Hex()

		entries = append(entries, colorEntry{
			Name: fieldName,
			Hex:  hex,
			R:    r,
			G:    g,
			B:    b,
			A:    a * 255,
		})
	}

	cols := 4

	// 2. Calculate maximum key length PER COLUMN
	colWidths := make([]int, cols)
	for i, entry := range entries {
		colIdx := i % cols
		if len(entry.Name) > colWidths[colIdx] {
			colWidths[colIdx] = len(entry.Name)
		}
	}

	// 3. Build and log the aligned rows
	for i := 0; i < len(entries); i += cols {
		var rowBuilder strings.Builder

		for j := 0; j < cols && i+j < len(entries); j++ {
			entry := entries[i+j]
			coloredHex := formatANSIColor(entry.Hex, entry.R, entry.G, entry.B, true)
			coloredAlpha := formatANSIAlpha(entry.A, true)

			// Pad the field name to colWidths[j] using positive width (%*s) to right-align it
			// Format: "        KeyName = #COLOR"
			item := fmt.Sprintf("%*s=[#%s%s]", colWidths[j], entry.Name, coloredHex, coloredAlpha)
			rowBuilder.WriteString(item)

			// Add spacing between columns (except for the last column)
			if j < cols-1 {
				rowBuilder.WriteString(" ")
			}
		}

		logger.Info(rowBuilder.String())
	}
}

func formatANSIAlpha(a float64, trueColor bool) string {
	if trueColor {
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%02x\x1b[0m", int(a), int(a), int(a), uint8(a))
	}
	return fmt.Sprintf("\x1b[38;5;%dm%02x\x1b[0m", rgbTo256(a, a, a), uint8(a))
}

// formatANSIColor wraps a string in terminal color codes based on RGB values
func formatANSIColor(text string, r, g, b float64, trueColor bool) string {
	if trueColor {
		// 24-bit TrueColor Mode: \x1b[38;2;R;G;Bm
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", int(r), int(g), int(b), text)
	}

	// 8-bit 256-Color Mode: \x1b[38;5;Nm
	return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", rgbTo256(r, g, b), text)
}

// rgbTo256 converts a 24-bit RGB value into the nearest 8-bit ANSI 256-color index
func rgbTo256(r, g, b float64) int {
	// Exact Grayscale match (indices 232-255)
	if r == g && g == b {
		if r < 8 {
			return 16 // Black
		}
		if r > 248 {
			return 231 // White
		}
		return int(math.Round(((r-8)/247)*24)) + 232
	}

	// 6x6x6 Color Cube (indices 16-231)
	// Map 0-255 range into 0-5 steps
	r5 := int(math.Round((r / 255.0) * 5))
	g5 := int(math.Round((g / 255.0) * 5))
	b5 := int(math.Round((b / 255.0) * 5))

	return 16 + (36 * r5) + (6 * g5) + b5
}
