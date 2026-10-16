package app

import (
	"fmt"
	"math"
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	my_noire "github.com/gethiox/EDCITR/internal/noire"
)

func (r *Renderer) renderTabColors(startPosY float32) {
	startPosY += OptionHeightWithPad

	fdevColor := ColorUIHLFdev.Brighten(0.2).Shade(0.3)
	borderColor := rl.Black
	posXstart := float32(int32(r.s.CanvasMiddleX - ((int(ColorWidth) * ColorsPerRow) / 2)))

	labelX := float32(r.s.CanvasMiddleX) - (ColorLabelWidth / 2.0)

	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		r.s.CurrentlyPressedColor = nil
	}
	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
		r.s.CurrentlyReleasedColor = nil
	}

	var colorInfoPresented bool

	for i, setting := range r.colorSettings {
		if setting.Color == nil {
			continue
		}

		colRect := colorSettingRect(posXstart, startPosY, i)

		// Outer Highlight/Border
		if r.s.CurrentlyModifiedColor == setting {
			s := time.Since(r.app.startedAt).Seconds()
			color := my_noire.NewHCL(math.Mod(s*180, 360), 80.0, 80.0)
			rl.DrawRectangleLinesEx(colRect, ColorBorder, color.AsRGBA())
		} else {
			rl.DrawRectangleLinesEx(colRect, ColorBorder, fdevColor.AsRGBA())
		}

		// Inner Outline
		outlineRect := shrinkRect(colRect, ColorBorder)
		rl.DrawRectangleLinesEx(outlineRect, OutlineBorder, borderColor)

		// Color Segment
		innerRect := shrinkRect(outlineRect, OutlineBorder)
		segmentColor := (*setting.Color).AsRGBA()
		//if !keepAlphaIfNeeded(123, rl.Rectangle{}) { // Note: 123 remains here, assuming it's a specific control ID
		//	segmentColor.A = uint8(GuiCurrentAlpha * 255)
		//}
		rl.DrawRectangle(int32(innerRect.X), int32(innerRect.Y), int32(innerRect.Width), int32(innerRect.Height), segmentColor)

		// Mouse Interaction
		if rl.CheckCollisionPointRec(r.s.MousePosition, colRect) {
			gui.Label(rl.Rectangle{labelX, startPosY - OptionHeightWithPad, ColorLabelWidth, OptionHeight}, setting.Name)
			colorInfoPresented = true

			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
				r.s.CurrentlyPressedColor = setting
			}
			if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
				r.s.CurrentlyReleasedColor = setting
			}

			if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
				if setting.Color != nil {
					*setting.Color = setting.DefaultColor
					setting.UpdateHSV()
				}
			}
		}
	}

	// Drag and Drop, Color Picker toggle
	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
		if r.s.CurrentlyPressedColor != nil && r.s.CurrentlyReleasedColor != nil {
			if r.s.CurrentlyPressedColor != r.s.CurrentlyReleasedColor {
				*r.s.CurrentlyReleasedColor.Color = *r.s.CurrentlyPressedColor.Color
				r.s.CurrentlyReleasedColor.Hsv = r.s.CurrentlyPressedColor.Hsv
			} else {
				if r.s.CurrentlyModifiedColor == r.s.CurrentlyReleasedColor {
					r.s.CurrentlyModifiedColor = nil
				} else {
					r.s.CurrentlyModifiedColor = r.s.CurrentlyReleasedColor
				}
			}
		}
		r.s.CurrentlyPressedColor = nil
		r.s.CurrentlyReleasedColor = nil
	}

	// Default label if hovering over nothing
	if !colorInfoPresented {
		gui.Label(rl.Rectangle{labelX, startPosY - OptionHeightWithPad, ColorLabelWidth, OptionHeight}, "Colors")
	}

	// Render Picker side-panel
	if r.s.CurrentlyModifiedColor != nil {
		pickerX := float32(r.s.CanvasMiddleX) + (ColorWidth*float32(ColorsPerRow))/2 + PickerOffsetX

		gui.Label(rl.Rectangle{pickerX, startPosY - OptionHeightWithPad, PickerWidth, OptionHeight}, r.s.CurrentlyModifiedColor.Name)

		pickerRect := rl.Rectangle{pickerX, startPosY, PickerWidth, PickerHeight}

		before := r.s.CurrentlyModifiedColor.Hsv
		gui.ColorPickerHSV(pickerRect, "", &r.s.CurrentlyModifiedColor.Hsv)
		after := r.s.CurrentlyModifiedColor.Hsv

		if before != after {
			*r.s.CurrentlyModifiedColor.Color = my_noire.NewHSV(float64(after.X), float64(after.Y)*100, float64(after.Z)*100)
			r.s.CurrentlyModifiedColor.Hsv = after
		}

		alphaRect := rl.Rectangle{pickerX, startPosY + PickerHeight + PickerAlphaPadding, PickerWidth, OptionHeight}
		currentVal := float32(r.s.CurrentlyModifiedColor.Color.Alpha)
		_ = gui.ColorBarAlpha(alphaRect, "", &currentVal)
		r.s.CurrentlyModifiedColor.Color.Alpha = float64(currentVal)

		alphaLabelRect := rl.Rectangle{pickerX, alphaRect.Y + OptionHeightWithPad, PickerWidth, OptionHeight}
		gui.Label(alphaLabelRect, fmt.Sprintf("Alpha: %4.2f", currentVal))
	}
}
