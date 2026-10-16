package app

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/app/state"
)

func (r *Renderer) startGenerateJsons() {
	r.app.State.Tasks.GenJsons.Start()
}

func (r *Renderer) startGenerateNotifications() {
	r.app.State.Tasks.GenNotifications.Start()
}

func (r *Renderer) startGenerateNPCMessages() {
	r.app.State.Tasks.GenNPCMessages.Start()
}

func (r *Renderer) renderTabExport(startPosY float32) {
	for row, widget := range r.exportSettings {
		widget.Draw(r.optionRect(startPosY, row))
	}

	dst := rl.Rectangle{
		float32(r.s.CanvasMiddleX) - UIFontSize/2,
		startPosY + UIFontSize*14,
		UIFontSize,
		UIFontSize,
	}

	switch r.app.State.ExportStateEvents {
	case state.StateOK:
		rl.DrawRectangleRec(dst, r.app.Config.Colors.GeneralOK.AsRGBA())
	case state.StateError:
		rl.DrawRectangleRec(dst, r.app.Config.Colors.GeneralCritical.AsRGBA())
	case state.StateQueued:
		frame := getEdLoaderFrame()
		src := rl.Rectangle{
			float32(EdLoaderFrameWidth) * float32(frame),
			0,
			float32(EdLoaderFrameWidth),
			float32(EdLoaderFrameHeight),
		}
		rl.DrawTexturePro(
			r.app.State.Textures.EliteLoader,
			src,
			dst,
			rl.Vector2{},
			0,
			rl.White,
		)
	case 0:
		rl.DrawRectangleRec(dst, r.app.Config.Colors.GeneralDarkGray.AsRGBA())
	default:
		panic("aaaa")
	}
}
