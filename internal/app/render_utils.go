package app

import (
	"time"

	"github.com/gen2brain/raylib-go/raygui"
	"github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/noire"
)

type RenderObject struct {
	text        string
	color       my_noire.Color
	font        *rl.Font
	fontSize    float32
	fontSpacing float32
}

type RenderObjects []RenderObject

func (o RenderObjects) Width() float32 {
	var textWidth float32

	for _, object := range o {
		textWidth += rl.MeasureTextEx(*object.font, object.text, object.fontSize, object.fontSpacing).X
	}

	return textWidth
}

func (o RenderObjects) Draw(pos rl.Vector2) {
	for _, object := range o {
		rl.DrawTextEx(*object.font, object.text, pos, object.fontSize, object.fontSpacing, object.color.AsRGBA())
		pos.X += rl.MeasureTextEx(*object.font, object.text, object.fontSize, object.fontSpacing).X
	}
}

func (r *Renderer) getMouseWheelV() rl.Vector2 {
	v := rl.GetMouseWheelMoveV()
	v.Y *= r.app.Config.WheelSpeed
	return v
}

func (r *Renderer) keepAlphaIfNeeded(id int, rect rl.Rectangle) bool {
	if rect.Width == 0 && rect.Height == 0 {
		return false
	}

	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
		r.s.OverrideAlpha = id
	}

	if r.s.OverrideAlpha == id {
		raygui.SetAlpha(GuiNormalAlpha)
		return true
	}
	return false
}

func (r *Renderer) keepAlphaIfNeededOnHover(id int, rect rl.Rectangle) bool {
	if rect.Width == 0 && rect.Height == 0 {
		return false
	}

	if rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
		r.s.OverrideAlpha = id
	}

	if r.s.OverrideAlpha == id {
		raygui.SetAlpha(GuiNormalAlpha)
		return true
	}
	return false
}

func (r *Renderer) peakOnHover(rect rl.Rectangle) {
	if !r.s.StateHoverPeek && rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
		r.s.StateHoverPeek = true
		r.s.StatePeek = true
		r.s.StatePeekAt = time.Now()
		r.s.StatePeekChangedAt = r.s.StatePeekCurrent
		// StatePeekReleasedAt // not modified
	}
	if r.s.StateHoverPeek && !rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
		r.s.StateHoverPeek = false
		r.s.StatePeek = false
		r.s.StatePeekReleasedAt = time.Now()
		r.s.StatePeekChangedAt = r.s.StatePeekCurrent
		// StatePeekAt // not modified
	}

	raygui.SetAlpha(r.s.GuiCurrentAlpha)
}

func (r *Renderer) peekIfNecessary(rect rl.Rectangle) {
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
		r.s.StatePeek = true
		r.s.StatePeekAt = time.Now()
		r.s.StatePeekChangedAt = r.s.StatePeekCurrent
		// StatePeekReleasedAt // not modified
	}
	if r.s.StatePeek && rl.IsMouseButtonReleased(rl.MouseButtonLeft) && !r.s.StateHoverPeek {
		r.s.StatePeek = false
		r.s.StatePeekReleasedAt = time.Now()
		r.s.StatePeekChangedAt = r.s.StatePeekCurrent
		// StatePeekAt // not modified
	}

	raygui.SetAlpha(r.s.GuiCurrentAlpha)
}

func (r *Renderer) drawScrollIndicators(rect rl.Rectangle, currentScroll, maxScroll float32) {
	indicatorHeight := float32(20)
	color := r.app.Config.Colors.HeaderDefault.AsRGBA() // Or another contrasting color

	if currentScroll > 0 {
		// Top indicator (can scroll up)
		rl.DrawRectangleGradientV(int32(rect.X), int32(rect.Y), int32(rect.Width), int32(indicatorHeight), color, rl.Blank)
		rl.DrawText("▲", int32(rect.X+rect.Width/2-5), int32(rect.Y+2), 10, rl.RayWhite)
	}

	if currentScroll < maxScroll {
		// Bottom indicator (can scroll down)
		rl.DrawRectangleGradientV(int32(rect.X), int32(rect.Y+rect.Height-indicatorHeight), int32(rect.Width), int32(indicatorHeight), rl.Blank, color)
		rl.DrawText("▼", int32(rect.X+rect.Width/2-5), int32(rect.Y+rect.Height-12), 10, rl.RayWhite)
	}
}

// TODO: what to do with this shit
func (r *Renderer) renderEditField(rect rl.Rectangle, editField *string, stateEditing, stateWarning *bool, label string) bool {
	txt := "Clear Value?"
	m := rl.MeasureTextEx(r.app.State.Fonts.UI, " "+txt+" ", UIFontSize, UIFontSpacing)

	buttonWidth := float32(100)
	middle := rect.X + rect.Width - rect.Width/2

	titleRect, yesRect, noRect := rect, rect, rect

	titleRect.Width = m.X
	titleRect.X = middle - m.X/2

	yesRect.Width = buttonWidth
	yesRect.X = middle - m.X/2 - buttonWidth

	noRect.Width = buttonWidth
	noRect.X = middle + m.X/2

	if *stateWarning {
		raygui.Label(titleRect, txt)
		yes := raygui.Button(yesRect, "Yes")
		no := raygui.Button(noRect, "No")
		if no {
			*stateWarning = false
			return false
		}
		if yes {
			*stateWarning = false
			*editField = ""
			return true
		}
	} else {
		if rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
				*stateEditing = true
				r.s.DisableShortcuts = true // todo: better handle r.s.DisableShortcuts
			}
			if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
				*stateEditing = false
				r.s.DisableShortcuts = false
				*stateWarning = true
			}
		}

		labelRect := rect
		labelRect.X -= 400
		labelRect.Width = 400
		style := raygui.GetStyle(raygui.LABEL, raygui.TEXT_ALIGNMENT)
		raygui.SetStyle(raygui.LABEL, raygui.TEXT_ALIGNMENT, raygui.TEXT_ALIGN_RIGHT)
		raygui.Label(labelRect, label+" ")
		raygui.SetStyle(raygui.LABEL, raygui.TEXT_ALIGNMENT, style)

		editComplete := raygui.TextBox(rect,
			editField,
			255*255,
			*stateEditing,
		)

		if editComplete {
			*stateEditing = false
			r.s.DisableShortcuts = false
		}
		return editComplete
	}

	return false
}

func adjustForClickOrHover(color my_noire.Color) my_noire.Color {
	if rl.IsMouseButtonDown(rl.MouseLeftButton) || rl.IsMouseButtonDown(rl.MouseRightButton) {
		return color.Darken(ClickDarken)
	}
	return color.Lighten(HoverBrighten)
}
