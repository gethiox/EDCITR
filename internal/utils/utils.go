package utils

import (
	"embed"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unique"

	"github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/noire"
	"github.com/xeonx/timeago"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var RlTraceLevelToString = map[rl.TraceLogLevel]string{
	rl.LogAll:     "All",
	rl.LogTrace:   "Trace",
	rl.LogDebug:   "Debug",
	rl.LogInfo:    "Info",
	rl.LogWarning: "Warning",
	rl.LogError:   "Error",
	rl.LogFatal:   "Fatal",
	rl.LogNone:    "None",
}

// remaps given time to Elite Dangerous universe
// formula: current_gregorian_date + 1286, in UTC timezone
func EliteTimestamp(t time.Time) time.Time {
	return t.AddDate(1286, 0, 0).UTC()
}

var Ago = timeago.Config{
	PastPrefix:   "",
	PastSuffix:   " ago",
	FuturePrefix: "in ",
	FutureSuffix: "",
	Periods: []timeago.FormatPeriod{
		{time.Second, "a second", "%d seconds"},
		{time.Minute, "a minute", "%d minutes"},
		{time.Hour, "an hour", "%d hours"},
		{time.Hour * 24, "one day", "%d days"},
	},
	Zero:          "about a second",
	Max:           24 * time.Hour * 14,
	DefaultLayout: "2006-01-02",
}

func UnsafeLoadData(fs embed.FS, path string) []byte {
	data, err := fs.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return data
}

func UnsafeLoadFont(fs embed.FS, path string, fontSize int32, codepoints []rune) rl.Font {
	nameParts := strings.Split(path, ".")
	if len(nameParts) != 2 {
		panic(fmt.Sprintf("expected 2 parts of filename, name [dot] extension, got %d parts", len(nameParts)))
	}
	ext := nameParts[1]

	fontData, err := fs.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return rl.LoadFontFromMemory("."+ext, fontData, fontSize, codepoints)
}

// GetNextOccurrenceUTC calculates the next time.Time for a specific day and time
func GetNextOccurrenceUTC(now time.Time, sDay, sTime string) (time.Time, error) {
	// 1. Parse the target weekday
	targetWeekday, err := ParseWeekday(sDay)
	if err != nil {
		return time.Time{}, err
	}

	parsedTime, err := time.Parse("15:04", sTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format, use HH:MM: %v", err)
	}

	nowUtc := time.Now().UTC()

	daysUntil := int(targetWeekday - nowUtc.Weekday())
	if daysUntil < 0 {
		daysUntil += 7
	}

	nextTimestamp := time.Date(
		nowUtc.Year(), nowUtc.Month(), nowUtc.Day()+daysUntil,
		parsedTime.Hour(), parsedTime.Minute(), 0, 0,
		time.UTC,
	)

	if nextTimestamp.Before(now) {
		nextTimestamp = nextTimestamp.AddDate(0, 0, 7)
	}

	return nextTimestamp, nil
}

// Helper to map string days to time.Weekday
func ParseWeekday(day string) (time.Weekday, error) {
	switch strings.ToLower(strings.TrimSpace(day)) {
	case "1", "monday", "mon":
		return time.Monday, nil
	case "2", "tuesday", "tue":
		return time.Tuesday, nil
	case "3", "wednesday", "wed":
		return time.Wednesday, nil
	case "4", "thursday", "thu":
		return time.Thursday, nil
	case "5", "friday", "fri":
		return time.Friday, nil
	case "6", "saturday", "sat":
		return time.Saturday, nil
	case "7", "sunday", "sun":
		return time.Sunday, nil
	default:
		return 0, fmt.Errorf("invalid day: %s", day)
	}
}

var ansiPalette = []my_noire.Color{
	my_noire.NewRGBA(40, 44, 52, 1.0),    // 30: Black (Dark Slate)
	my_noire.NewRGBA(224, 108, 117, 1.0), // 31: Red (Soft Salmon)
	my_noire.NewRGBA(152, 195, 121, 1.0), // 32: Green (Sage)
	my_noire.NewRGBA(229, 192, 123, 1.0), // 33: Yellow (Gold)
	my_noire.NewRGBA(97, 175, 239, 1.0),  // 34: Blue (Cornflower)
	my_noire.NewRGBA(198, 120, 221, 1.0), // 35: Magenta (Orchid)
	my_noire.NewRGBA(86, 182, 194, 1.0),  // 36: Cyan (Teal)
	my_noire.NewRGBA(171, 178, 191, 1.0), // 37: White (Light Gray)

	my_noire.NewRGBA(92, 99, 112, 1.0),   // 90: Bright Black (Gray)
	my_noire.NewRGBA(240, 140, 140, 1.0), // 91: Bright Red
	my_noire.NewRGBA(180, 220, 150, 1.0), // 92: Bright Green
	my_noire.NewRGBA(245, 210, 140, 1.0), // 93: Bright Yellow
	my_noire.NewRGBA(120, 200, 255, 1.0), // 94: Bright Blue
	my_noire.NewRGBA(220, 140, 240, 1.0), // 95: Bright Magenta
	my_noire.NewRGBA(120, 220, 230, 1.0), // 96: Bright Cyan
	my_noire.NewRGBA(255, 255, 255, 1.0), // 97: Bright White
}

// DrawAnsiTextEx parses ANSI color codes and draws text using a custom Font.
// It supports standard (30-37), bright (90-97), 256-color (38;5;n), and TrueColor (38;2;r;g;b).
func DrawAnsiTextEx(font rl.Font, text string, position rl.Vector2, fontSize, spacing float32, defaultColor my_noire.Color, highlighted bool, lighten float64) {
	currentColor := defaultColor
	currentPos := position

	if highlighted {
		currentColor = currentColor.Lighten(lighten)
	}

	for len(text) > 0 {
		// Find the start of the next escape sequence
		escStart := strings.Index(text, "\x1b[")

		// Case 1: No more escape codes, draw the remaining text
		if escStart == -1 {
			rl.DrawTextEx(font, text, currentPos, fontSize, spacing, currentColor.AsRGBA())
			break
		}

		// Case 2: Text exists before the escape code, draw it first
		if escStart > 0 {
			segment := text[:escStart]
			rl.DrawTextEx(font, segment, currentPos, fontSize, spacing, currentColor.AsRGBA())

			// Measure the width of the drawn segment to advance X position
			size := rl.MeasureTextEx(font, segment, fontSize, spacing)
			currentPos.X += size.X
		}

		// Process the escape sequence
		remaining := text[escStart+2:] // Skip "\x1b["
		end := strings.Index(remaining, "m")

		if end == -1 {
			// Malformed sequence: print the rest as raw text and exit
			rl.DrawTextEx(font, text[escStart:], currentPos, fontSize, spacing, currentColor.AsRGBA())
			break
		}

		// Parse the color code
		codeStr := remaining[:end]
		currentColor = parseAnsiColor(codeStr, currentColor, defaultColor)

		if highlighted {
			currentColor = currentColor.Lighten(lighten)
		}

		// Advance the buffer past the "m"
		text = remaining[end+1:]
	}
}

func parseAnsiColor(codeStr string, current, defaultColor my_noire.Color) my_noire.Color {
	if codeStr == "" || codeStr == "0" {
		return defaultColor
	}

	parts := strings.Split(codeStr, ";")
	toInt := func(s string) int {
		v, _ := strconv.Atoi(s)
		return v
	}

	code := toInt(parts[0])

	switch {
	case code == 0:
		return defaultColor

	// Standard Colors (30-37) -> Map to Palette 0-7
	case code >= 30 && code <= 37:
		return ansiPalette[code-30]

	// Bright Colors (90-97) -> Map to Palette 8-15
	case code >= 90 && code <= 97:
		return ansiPalette[code-90+8]

	// Extended Colors (38;...)
	case code == 38:
		if len(parts) >= 3 && parts[1] == "5" {
			// 256-Color Mode: 38;5;n
			return getAnsi256Color(toInt(parts[2]))
		} else if len(parts) >= 5 && parts[1] == "2" {
			// TrueColor Mode: 38;2;r;g;b
			r := uint8(toInt(parts[2]))
			g := uint8(toInt(parts[3]))
			b := uint8(toInt(parts[4]))
			return my_noire.NewRGBA(float64(r), float64(g), float64(b), 255)
		}
	}

	return current
}

func getAnsi256Color(n int) my_noire.Color {
	if n < 0 {
		n = 0
	}
	if n > 255 {
		n = 255
	}

	switch {
	case n < 16:
		// Use our nice palette for the first 16 colors
		return ansiPalette[n]

	case n <= 231:
		// 216-color cube (Standard math, keeping as is)
		n -= 16
		b := uint8((n % 6) * 51)
		n /= 6
		g := uint8((n % 6) * 51)
		n /= 6
		r := uint8(n * 51)
		return my_noire.NewRGBA(float64(r), float64(g), float64(b), 1.0)
	default:
		// Grayscale (Standard math)
		val := uint8(10 + (n-232)*10)
		return my_noire.NewRGBA(float64(val), float64(val), float64(val), 1.0)
	}
}

var printer = message.NewPrinter(language.English)

func FormatAsCredits(value int) string {
	return printer.Sprintf("%d CR", value)
}

func FormatDurationPrecise(duration time.Duration) string {
	switch {
	case duration < time.Second:
		return fmt.Sprintf("%dms", duration.Milliseconds())
	case duration < time.Minute:
		return fmt.Sprintf("%.0fs %.0fms", math.Trunc(duration.Seconds()), math.Mod(float64(duration.Milliseconds()), 1000))
	case duration < time.Hour:
		return fmt.Sprintf("%.0fm %.0fs %.0fms", math.Trunc(duration.Minutes()), math.Mod(math.Trunc(duration.Seconds()), 60), math.Mod(float64(duration.Milliseconds()), 1000))
	case duration < time.Hour*24:
		return fmt.Sprintf("%.0fh %.0fm %.0fs %.0fms", math.Trunc(duration.Hours()), math.Mod(math.Trunc(duration.Minutes()), 60), math.Mod(math.Trunc(duration.Seconds()), 60), math.Mod(float64(duration.Milliseconds()), 1000))
	default:
		days := duration.Hours() / 24
		return fmt.Sprintf("%.0fd %.0fh %.0fm %.0fs %.0fms", math.Trunc(days), math.Mod(math.Trunc(duration.Hours()), 24), math.Mod(math.Trunc(duration.Minutes()), 60), math.Mod(math.Trunc(duration.Seconds()), 60), math.Mod(float64(duration.Milliseconds()), 1000))
	}
}

func FormatDuration(duration time.Duration) string {
	switch {
	case duration < time.Second:
		return fmt.Sprintf("%dms", duration.Milliseconds())
	case duration < time.Minute:
		return fmt.Sprintf("%.0fs", duration.Seconds())
	case duration < time.Hour:
		return fmt.Sprintf("%.0fm %.0fs", duration.Minutes(), math.Mod(duration.Seconds(), 60))
	case duration < time.Hour*24:
		return fmt.Sprintf("%.0fh %.0fm %.0fs", duration.Hours(), math.Mod(duration.Minutes(), 60), math.Mod(duration.Seconds(), 60))
	default:
		days := duration.Hours() / 24
		return fmt.Sprintf("%.0fd %.0fh %.0fm %.0fs", days, math.Mod(duration.Hours(), 24), math.Mod(duration.Minutes(), 60), math.Mod(duration.Seconds(), 60))
	}
}

// based on time.RFC3339, modified to be friendly for general filesystems
const timeLayoutForFilename = "2006-01-02_15-04-05_0700"

func GenerateTimestampForFilename() string {
	return time.Now().Format(timeLayoutForFilename)
}

func EaseOutQuart(x float32) float32 {
	return float32(1 - math.Pow(float64(1-x), 4))
}

func EaseInExpo(x float64) float64 {
	if x == 0 {
		return 0
	}
	return math.Pow(2, 10*x-10)
}

func EaseOutExpo(x float64) float64 {
	if x == 0 {
		return 0
	}
	return 1 - math.Pow(2, -10*x)
}

func UniqueString(s string) string {
	return unique.Make(s).Value()
}
