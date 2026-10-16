package app

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/utils"
)

func (r *Renderer) renderLogs(available float32) {
	rect := rl.Rectangle{
		0,
		float32(r.s.CanvasSizeY) - available,
		float32(r.s.CanvasSizeX),
		available,
	}

	if !r.s.OptionViewActive && rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
		scroll := r.getMouseWheelV()

		scrollPos := r.s.ShowLogsPos
		if scrollPos == 0 {
			r.s.ShowLogsCheckpoint = r.app.State.LogsRVB.GetCheckpoint()
		}

		r.s.ShowLogsPos += int(scroll.Y)
		if r.s.ShowLogsPos < 0 {
			r.s.ShowLogsPos = 0
		}
		if rl.IsMouseButtonPressed(rl.MouseButtonMiddle) {
			r.s.ShowLogsPos = 0
		}
	}

	number := int(available/UIFontSize) + 1

	var isScrolling bool
	var toRender []string
	if r.s.ShowLogsPos > 0 {
		toRender, _ = r.app.State.LogsRVB.ReadNewFromCheckpoint(r.s.ShowLogsCheckpoint, r.s.ShowLogsPos, number)
		isScrolling = true
	} else {
		toRender = r.app.State.LogsRVB.ReadNew(number)
	}

	for i, log := range toRender {
		pos := rl.Vector2{0, float32(r.s.CanvasSizeY) - float32(UIFontSize*(i+1))}
		var highlighted bool
		rect := rl.Rectangle{
			pos.X,
			pos.Y,
			float32(r.s.CanvasSizeX),
			UIFontSize,
		}
		if !r.s.OptionViewActive && rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
			if !rl.IsMouseButtonDown(rl.MouseRightButton) {
				highlighted = true
			}

			if rl.IsMouseButtonReleased(rl.MouseRightButton) {
				rl.SetClipboardText(ansi.Strip(log))
			}
		}
		utils.DrawAnsiTextEx(r.app.State.Fonts.UI, log, pos, UIFontSize, UIFontSpacing, r.app.Config.Colors.GeneralNormal, highlighted, HoverBrighten)
	}

	if isScrolling {
		n := r.app.State.LogsRVB.NewItemsSince(r.s.ShowLogsCheckpoint)
		color := r.app.Config.Colors.GeneralGray
		txt2 := " 0 New "
		if n > 0 {
			color = r.app.Config.Colors.GeneralCritical
			txt2 = fmt.Sprintf(" !!! %d New !!! ", n)
		}

		txt1 := " [SCROLLING MODE] "
		m1 := rl.MeasureTextEx(r.app.State.Fonts.UI, txt1, UIFontSize, UIFontSpacing)
		m2 := rl.MeasureTextEx(r.app.State.Fonts.UI, txt2, UIFontSize, UIFontSpacing)
		pos1 := rl.Vector2{float32(r.s.CanvasSizeX) - m1.X, float32(r.s.CanvasSizeY) - float32(UIFontSize)}
		pos2 := rl.Vector2{float32(r.s.CanvasSizeX) - m1.X - m2.X, float32(r.s.CanvasSizeY) - float32(UIFontSize)}
		rec := rl.Rectangle{
			float32(r.s.CanvasSizeX) - m1.X - m2.X,
			float32(r.s.CanvasSizeY) - float32(UIFontSize),
			m1.X + m2.X,
			UIFontSize,
		}
		rl.DrawRectangleRec(rec, r.app.Config.Colors.Background.AsRGBA())
		rl.DrawTextEx(r.app.State.Fonts.UI, txt1, pos1, UIFontSize, UIFontSpacing, r.app.Config.Colors.GeneralWarning.AsRGBA())
		rl.DrawTextEx(r.app.State.Fonts.UI, txt2, pos2, UIFontSize, UIFontSpacing, color.AsRGBA())
	}
}
