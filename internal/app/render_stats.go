package app

import (
	"fmt"
	"slices"
	"time"

	"github.com/gen2brain/raylib-go/raylib"
	my_noire "github.com/gethiox/EDCITR/internal/noire"
	"github.com/gethiox/EDCITR/internal/utils"
)

type Separator struct{}

type Label struct {
	name string
}

type Info struct {
	name, value string
	modified    time.Time
}

type Stat struct {
	name, value string
	modified    time.Time
}

type MultiStat []Stat

func (r *Renderer) getStats() []any {
	var stats []any

	if r.app.Config.Stats.JumpsAndDockings {
		stats = append(stats,
			MultiStat{
				Stat{"Jumps", fmt.Sprintf("%d", r.app.State.Stats.Jumps.Load()), r.app.State.Stats.Jumps.Modified()},
				Stat{"Dockings", fmt.Sprintf("%d", r.app.State.Stats.Dockings.Load()), r.app.State.Stats.Dockings.Modified()},
			},
		)
	}

	if r.app.Config.Stats.Signals {
		stats = append(stats,
			Stat{"Discovered Signals", fmt.Sprintf("%d", r.app.State.Stats.SignalsDetected.Load()), r.app.State.Stats.SignalsDetected.Modified()},
		)
	}

	if r.app.Config.Stats.Credits {
		stats = append(stats,
			Separator{},
			Stat{"Income", utils.FormatAsCredits(r.app.State.Stats.CreditsIncome.Load()), r.app.State.Stats.CreditsIncome.Modified()},
			Stat{"Expenses", utils.FormatAsCredits(r.app.State.Stats.CreditsExpenses.Load()), r.app.State.Stats.CreditsExpenses.Modified()},
			Stat{"Net Income", utils.FormatAsCredits(r.app.State.Stats.CreditsNetIncome.Load()), r.app.State.Stats.CreditsNetIncome.Modified()},
			Stat{"Crew Wage", utils.FormatAsCredits(r.app.State.Stats.CrewWage.Load()), r.app.State.Stats.CrewWage.Modified()},
		)
	}

	if r.app.Config.Stats.Commodities {
		stats = append(stats,
			Separator{},
			Stat{"Commodity Buy", utils.FormatAsCredits(r.app.State.Stats.CommodityBuy.Load()), r.app.State.Stats.CommodityBuy.Modified()},
			Stat{"Commodity Sell", utils.FormatAsCredits(r.app.State.Stats.CommoditySell.Load()), r.app.State.Stats.CommoditySell.Modified()},
			Stat{"Commodity Net", utils.FormatAsCredits(r.app.State.Stats.CommodityNet.Load()), r.app.State.Stats.CommodityNet.Modified()},
		)
	}

	if r.app.Config.Stats.Playtime {
		stats = append(stats,
			Separator{},
			Stat{"Open Playtime", utils.FormatDuration(r.app.State.Stats.PlaytimeOpen.Load()), r.app.State.Stats.PlaytimeOpen.Modified()},
			Stat{"Solo Playtime", utils.FormatDuration(r.app.State.Stats.PlaytimeSolo.Load()), r.app.State.Stats.PlaytimeSolo.Modified()},
			Stat{"Group Playtime", utils.FormatDuration(r.app.State.Stats.PlaytimeGroup.Load()), r.app.State.Stats.PlaytimeGroup.Modified()},
		)
	}

	if r.app.Config.Stats.Kills {
		stats = append(stats,
			Separator{},
			Stat{"Deaths", fmt.Sprintf("%d (%d PVP)", r.app.State.Stats.Deaths.Load(), r.app.State.Stats.DeathsByPlayer.Load()), r.app.State.Stats.DeathsByPlayer.Modified()},
			MultiStat{
				Stat{"PVP Kills", fmt.Sprintf("%d", r.app.State.Stats.KillsPlayer.Load()), r.app.State.Stats.KillsPlayer.Modified()},
				Stat{"NPC Kills", fmt.Sprintf("%d", r.app.State.Stats.KillsNPC.Load()), r.app.State.Stats.KillsNPC.Modified()},
			},
			Stat{"Combat Zone Kills", fmt.Sprintf("%d", r.app.State.Stats.KillsCombatZone.Load()), r.app.State.Stats.KillsCombatZone.Modified()},
		)
	}

	if r.app.Config.Stats.Journal {
		stats = append(stats,
			Separator{},
			Label{"=== Journal Stats ==="},
			Stat{"Produced Messages", fmt.Sprintf("%d", r.app.State.Stats.ProducedMessages.Load()), r.app.State.Stats.ProducedMessages.Modified()},
			Stat{"Processed Events", fmt.Sprintf("%d", r.app.State.Stats.ProcessedEvents.Load()), r.app.State.Stats.ProcessedEvents.Modified()},
			Stat{"Not supported (skipped) Events", fmt.Sprintf("%d", r.app.State.Stats.UnprocessedEvents.Load()), r.app.State.Stats.UnprocessedEvents.Modified()},
		)
	}

	if r.app.Config.Stats.Sessions {
		stats = append(stats,
			Separator{},
			Label{"=== Sessions ==="},
			MultiStat{
				Stat{"Open", fmt.Sprintf("%d", r.app.State.Stats.SessionsOpen.Load()), r.app.State.Stats.SessionsOpen.Modified()},
				Stat{"Solo", fmt.Sprintf("%d", r.app.State.Stats.SessionsSolo.Load()), r.app.State.Stats.SessionsSolo.Modified()},
				Stat{"Group", fmt.Sprintf("%d", r.app.State.Stats.SessionsGroup.Load()), r.app.State.Stats.SessionsGroup.Modified()},
			},
			Stat{"Startups", fmt.Sprintf("%d", r.app.State.SessionCounter.Load()), r.app.State.SessionCounter.Modified()},
		)
	}

	if r.app.Config.Stats.CurrentVariables {
		stats = append(stats,
			Separator{},
			Label{"=== Current Variables ==="},
			Info{"CMDR", r.app.State.CurrentThings.CMDR.Load(), r.app.State.CurrentThings.CMDR.Modified()},
			Info{"Mode", r.app.State.CurrentThings.Mode.Load(), r.app.State.CurrentThings.Mode.Modified()},
			Info{"Music", r.app.State.CurrentThings.Music.Load(), r.app.State.CurrentThings.Music.Modified()},
		)
	}

	if r.app.Config.Stats.CurrentLocation {
		stats = append(stats,
			Separator{},
			Label{"=== Current Location ==="},
			Info{"System", r.app.State.CurrentThings.LocationSystem.Load(), r.app.State.CurrentThings.LocationSystem.Modified()},
			Info{"Body", r.app.State.CurrentThings.LocationBody.Load(), r.app.State.CurrentThings.LocationBody.Modified()},
			Info{"Body Type", r.app.State.CurrentThings.LocationBodyType.Load(), r.app.State.CurrentThings.LocationBodyType.Modified()},
			Info{"Population", r.app.State.CurrentThings.LocationPopulation.Load(), r.app.State.CurrentThings.LocationPopulation.Modified()},
		)
	}

	return stats
}

func (r *Renderer) renderStatsHeader(width float32) float32 {
	headerHeight := float32(r.app.Config.Fonts.Size) * 1.5
	rect := rl.Rectangle{float32(r.s.CanvasSizeX) - width, 0, width, headerHeight}
	rl.DrawRectangleGradientEx(
		rect,
		r.app.Config.Colors.Header1.AsRGBA(),
		r.app.Config.Colors.Header2.AsRGBA(),
		r.app.Config.Colors.Header2.AsRGBA(),
		r.app.Config.Colors.Header1.AsRGBA(),
	)

	txt := "[=Statistics=] (WIP)"
	m := rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
	offsetX := rect.X + rect.Width/2 - m.X/2
	offsetY := headerHeight/2 - m.Y/2

	rectCheck := rl.Rectangle{
		offsetX,
		offsetY,
		m.X,
		m.Y,
	}
	var color = r.app.Config.Colors.GeneralNormal
	if !r.s.OptionViewActive && rl.CheckCollisionPointRec(r.s.MousePosition, rectCheck) {
		color = adjustForClickOrHover(color)
		if rl.IsMouseButtonReleased(rl.MouseButtonRight) {
			txt := ""
			stats := r.getStats()

			for _, stat := range stats {
				switch s := stat.(type) {
				case Info, Separator, Label:
					continue
				case Stat:
					txt += fmt.Sprintf("%s: %s\n", s.name, s.value)
				case MultiStat:
					for _, s := range s {
						txt += fmt.Sprintf("%s: %s\n", s.name, s.value)
					}
				default:
					panic(fmt.Sprintf("unknown stat type %T", s))
				}

			}
			rl.SetClipboardText(txt)
			r.app.State.BubbleMessage.SetMessage("stats copied!", r.s.MousePosition)
		}
	}

	rl.DrawTextEx(r.app.State.Fonts.Main, txt, rl.Vector2{offsetX, offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), color.AsRGBA())
	return headerHeight
}

type DrawCall struct {
	text             string
	offsetX, offsetY float32
	size             rl.Vector2
	font             *rl.Font
	fontSize         float32
	fontSpacing      float32
}

type ASDFASDF struct {
	DrawCalls [2]DrawCall
	Color     my_noire.Color
}

func (r *Renderer) renderStats() (reservedSpace int) {
	if !r.app.Config.Stats.Enabled {
		return 0
	}

	fontSize := float32(r.app.Config.Fonts.Size)
	statusBarHeight := r.getRealStatusBarHeight()

	stats := r.getStats()

	var drawCalls = make([]ASDFASDF, 0, len(stats))
	minimumWidthRequired := float32(0)

	var now = time.Now()
	var occuppied, free int

	for _, stat := range stats {
		var key, value string
		var modified time.Time
		var multiStat MultiStat

		switch s := stat.(type) {
		case Info:
			key = fmt.Sprintf("%s:", s.name)
			value = fmt.Sprintf(" %s", s.value)
			modified = s.modified
		case Stat:
			key = fmt.Sprintf("%s:", s.name)
			value = fmt.Sprintf(" %s", s.value)
			modified = s.modified
		case MultiStat:
			key = fmt.Sprintf("%s:", s[0].name)
			value = fmt.Sprintf(" %s", s[0].value)
			multiStat = s[1:]
			modified = s[0].modified
		case Separator:
			free++
			continue
		case Label:
			key = s.name
			value = ""
		default:
			panic(fmt.Sprintf("unknown stat type %T", s))
		}

		offsetX := float32(0)

	Multi:

		mk := rl.MeasureTextEx(r.app.State.Fonts.Main, key, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		mv := rl.MeasureTextEx(r.app.State.Fonts.MainMono, value, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.SpacingMono))

		minimumWidthRequired = max(minimumWidthRequired, offsetX+mk.X+mv.X)

		calls := [2]DrawCall{
			{
				text:        key,
				offsetX:     offsetX,
				offsetY:     fontSize*float32(occuppied) + (fontSize/2)*float32(free),
				size:        rl.Vector2{mk.X, mk.Y},
				font:        &r.app.State.Fonts.Main,
				fontSize:    float32(r.app.Config.Fonts.Size),
				fontSpacing: float32(r.app.Config.Fonts.Spacing),
			},
			{
				text:        value,
				offsetX:     offsetX + mk.X,
				offsetY:     fontSize*float32(occuppied) + (fontSize/2)*float32(free),
				size:        rl.Vector2{mv.X, mv.Y},
				font:        &r.app.State.Fonts.MainMono,
				fontSize:    float32(r.app.Config.Fonts.Size),
				fontSpacing: float32(r.app.Config.Fonts.SpacingMono),
			},
		}

		desaturate := min(now.Sub(modified).Seconds(), 4.0) / 4.0

		drawCalls = append(drawCalls, ASDFASDF{
			DrawCalls: calls,
			Color:     r.app.Config.Colors.GeneralOK.Desaturate(desaturate),
		})

		if len(multiStat) > 0 {
			key = fmt.Sprintf("%s:", multiStat[0].name)
			value = fmt.Sprintf(" %s", multiStat[0].value)
			modified = multiStat[0].modified
			multiStat = multiStat[1:]
			offsetX += mk.X + mv.X + fontSize/2

			goto Multi
		}
		occuppied++

	}

	additionalSpace := "______"
	txt1 := "Open Playtime:"
	txt2 := " 15d 3h 32m 32s" + additionalSpace
	m1 := rl.MeasureTextEx(r.app.State.Fonts.Main, txt1, fontSize, float32(r.app.Config.Fonts.Spacing))
	m2 := rl.MeasureTextEx(r.app.State.Fonts.MainMono, txt2, fontSize, float32(r.app.Config.Fonts.SpacingMono))

	minimumWidthRequired = max(minimumWidthRequired, m1.X+m2.X)
	minimumWidthRequired += fontSize

	reservedSpace = int(minimumWidthRequired + float32(r.app.Config.Viewports.BorderSize))

	rl.DrawRectangleRec( // border
		rl.Rectangle{
			float32(r.s.CanvasSizeX) - minimumWidthRequired - float32(r.app.Config.Viewports.BorderSize),
			0,
			float32(r.app.Config.Viewports.BorderSize),
			float32(r.s.CanvasSizeY) - statusBarHeight - r.app.State.LogsVerticalPeek,
		},
		r.app.Config.Colors.ViewportBorder.AsRGBA(),
	)
	headerHeight := r.renderStatsHeader(minimumWidthRequired)

	anchor := rl.Vector2{
		float32(r.s.CanvasSizeX) - minimumWidthRequired + fontSize/2,
		headerHeight,
	}

	resetText := "[Reset]"
	m := rl.MeasureTextEx(r.app.State.Fonts.Main, resetText, fontSize, float32(r.app.Config.Fonts.Spacing))

	resetRect := rl.Rectangle{
		anchor.X,
		anchor.Y,
		m.X,
		m.Y,
	}
	color := r.app.Config.Colors.GeneralNormal
	if rl.CheckCollisionPointRec(r.s.MousePosition, resetRect) {
		color = adjustForClickOrHover(color)
		if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			r.app.State.Stats.ResetAll()
		}
	}
	rl.DrawTextEx(r.app.State.Fonts.Main, resetText, rl.Vector2{resetRect.X, resetRect.Y}, fontSize, float32(r.app.Config.Fonts.Spacing), color.AsRGBA())

	anchor.Y += fontSize

	for _, asdf := range drawCalls {
		call := asdf.DrawCalls

		rect := rl.Rectangle{
			anchor.X + call[0].offsetX,
			anchor.Y + call[0].offsetY,
			call[1].offsetX + call[1].size.X,
			fontSize,
		}
		color := asdf.Color
		if !r.s.OptionViewActive && rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
			color = adjustForClickOrHover(color)
			if rl.IsMouseButtonReleased(rl.MouseButtonRight) {
				txt := call[0].text + call[1].text
				rl.SetClipboardText(txt)
				r.app.State.BubbleMessage.SetMessage("stat copied!", r.s.MousePosition)
			}
		}

		for _, subCall := range call {
			pos := anchor
			pos.X += subCall.offsetX
			pos.Y += subCall.offsetY
			// TODO: Avoid copy-paste
			if pos.Y > float32(r.s.CanvasSizeY)-statusBarHeight-r.app.State.LogsVerticalPeek {
				return
			}
			rl.DrawTextEx(*subCall.font, subCall.text, pos, subCall.fontSize, subCall.fontSpacing, color.AsRGBA())
		}
	}

	friends := r.app.State.FriendList.Items()
	var friendList []string

	for friend, online := range friends {
		if !online {
			continue
		}
		friendList = append(friendList, friend)
	}
	slices.Sort(friendList)

	pos := rl.Vector2{
		anchor.X,
		anchor.Y + float32(occuppied)*fontSize + float32(free+1)*(fontSize/2),
	}
	// TODO: Avoid copy-paste
	if pos.Y > float32(r.s.CanvasSizeY)-statusBarHeight-r.app.State.LogsVerticalPeek {
		return
	}
	txt := fmt.Sprintf("Online Friends: (%d/%d)", len(friendList), len(friends))
	rl.DrawTextEx(r.app.State.Fonts.Main, txt, pos, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), r.app.Config.Colors.GeneralNormal.AsRGBA())
	for i, friend := range friendList {
		var color = r.app.Config.Colors.HeaderFriend
		pos := pos
		pos.Y += float32(i+1) * fontSize

		// TODO: Avoid copy-paste
		if pos.Y > float32(r.s.CanvasSizeY)-statusBarHeight-r.app.State.LogsVerticalPeek {
			return
		}

		txt := friend
		m := rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		rect := rl.Rectangle{
			pos.X,
			pos.Y,
			m.X,
			m.Y,
		}
		if !r.s.OptionViewActive && rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
			color = adjustForClickOrHover(color)

			if rl.IsMouseButtonReleased(rl.MouseRightButton) {
				rl.SetClipboardText(friend)
				r.app.State.BubbleMessage.SetMessage("friend name copied!", r.s.MousePosition)
			}
		}
		rl.DrawTextEx(r.app.State.Fonts.Main, txt, pos, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), color.AsRGBA())
	}

	return
}
