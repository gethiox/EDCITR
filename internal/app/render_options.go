package app

import (
	"fmt"
	"runtime"
	"runtime/metrics"
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/app/state"
	my_noire "github.com/gethiox/EDCITR/internal/noire"
	"github.com/gethiox/EDCITR/internal/utils"
	"go.uber.org/zap"
)

var metricList = []metrics.Sample{
	{Name: "/cgo/go-to-c-calls:calls"},
	{Name: "/cpu/classes/gc/pause:cpu-seconds"},
	{Name: "/cpu/classes/gc/total:cpu-seconds"},
	{Name: "/cpu/classes/idle:cpu-seconds"},
	{Name: "/cpu/classes/total:cpu-seconds"},
	{Name: "/cpu/classes/user:cpu-seconds"},
	{Name: "/gc/cycles/total:gc-cycles"},
	{Name: "/memory/classes/total:bytes"},
	{Name: "/sched/goroutines:goroutines"},
	{Name: "/sched/goroutines/running:goroutines"},
	{Name: "/sched/goroutines/waiting:goroutines"},
	{Name: "/sched/threads/total:threads"},
}

func (r *Renderer) renderOptionView() {
	r.renderAlphaAndBackground()
	r.renderMascot()
	r.renderDebugOverlay()

	startPosY := float32(r.s.CanvasMiddleY) - (LineCount*OptionHeightWithPad)/2

	r.renderTabNavigation(startPosY)

	switch r.s.OptionsActiveTab {
	case state.TabGeneral:
		r.renderTabGeneral(startPosY)
	case state.TabColors:
		r.renderTabColors(startPosY)
	case state.TabFilters:
		r.renderTabFilters(startPosY)
	case state.TabAudio:
		r.renderTabAudio(startPosY)
	case state.TabNetwork:
		r.renderTabNetwork(startPosY)
	case state.TabExport:
		r.renderTabExport(startPosY)
	case state.TabParser:
		r.renderTabParser(startPosY)
	case state.TabStats:
		r.renderTabStats(startPosY)
	}

	r.renderActionButtons(startPosY)
}

func (r *Renderer) renderActionButtons(startPosY float32) {
	modifiedPosY := startPosY + OptionHeightWithPad*IconRow

	saveLabel, saveColor := "  Save Config  ", ColorUI
	if r.app.Config.IsDirty(r.app.SavedConfig) {
		saveLabel, saveColor = "!  Save Config !", ColorUIHLFdev
	}

	saveRect := rl.Rectangle{X: float32(r.s.CanvasMiddleX) - SymmetricOffset, Y: modifiedPosY, Width: ActionIconEdge, Height: ActionIconEdge}
	saveLabelRect := rl.Rectangle{X: saveRect.X - AdditionalIconLabelWidth/2, Y: saveRect.Y - OptionHeight, Width: ActionIconEdge + AdditionalIconLabelWidth, Height: OptionHeight}

	r.renderIconAction(saveRect, saveLabelRect, gui.ICON_FILE_SAVE_CLASSIC, saveLabel, saveColor,
		func() { // Left click: Save
			if err := r.app.ConfigStore.SaveConfig(&r.app.Config); err != nil {
				Logger.Error("save config failed", zap.Error(err))
			} else {
				Logger.Info("config saved!", zap.String("path", r.app.ConfigStore.Root()))
				r.app.SavedConfig = r.app.Config.DeepCopy()
			}
			r.playBoopSound()
		},
		func() { // Right click: Default
			r.app.Config = *r.app.DefaultConfigReadOnly.DeepCopy()
			r.playBoopSound()
			Logger.Info("default config loaded!", zap.String("path", r.app.ConfigStore.Root()))
		},
	)

	switch r.s.OptionsActiveTab {
	case state.TabGeneral, state.TabFilters, state.TabParser:
		break
	default:
		return
	}

	repLabel, repColor := "  reprocess journals  ", ColorUI
	if r.app.Config.IsReprocessingRequired(r.app.ProcessedConfig) {
		repLabel, repColor = "! reprocess journals !", r.app.Config.Colors.GeneralWarning
	}

	repRect := rl.Rectangle{X: float32(r.s.CanvasMiddleX) + SymmetricOffset, Y: modifiedPosY, Width: ActionIconEdge, Height: ActionIconEdge}
	repLabelRect := rl.Rectangle{X: repRect.X - AdditionalIconLabelWidth/2, Y: repRect.Y - OptionHeight, Width: ActionIconEdge + AdditionalIconLabelWidth, Height: OptionHeight}

	icon := gui.ICON_UNDO_FILL
	if !r.app.State.ProcessingLiveEvents.Load() {
		icon = gui.ICON_SAND_TIMER
	}

	r.renderIconAction(repRect, repLabelRect, icon, repLabel, repColor,
		func() { // Left click: Reprocess
			r.playBoopSound()
			r.app.State.Tasks.Journal.RestartAsync()
		},
		nil, // No right click action
	)

}

func (r *Renderer) renderAlphaAndBackground() {
	if r.s.StatePeek { // TODO: holy moly
		elapsedNormalized := rl.Clamp(float32(time.Since(r.s.StatePeekAt))/float32(PeekDuration), 0, 1)
		finalScale := float32(utils.EaseOutExpo(float64(elapsedNormalized)))
		r.s.GuiCurrentAlpha = GuiNormalAlpha - (float32(GuiNormalAlpha-GuiPeekAlpha) * finalScale)
		r.s.GuiCurrentAlphaImage = GuiNormalAlpha - (float32(GuiNormalAlpha-0.0) * finalScale)
		r.s.GuiCurrentBgAlpha = GuiNormalBgAlpha - (float32(GuiNormalBgAlpha-GuiPeekBgAlpha) * finalScale)
	} else {
		elapsedNormalized := rl.Clamp(float32(time.Since(r.s.StatePeekReleasedAt))/float32(PeekDuration), 0, 1)
		stateDeltaNormalized := rl.Clamp(float32(r.s.StatePeekReleasedAt.Sub(r.s.StatePeekAt))/float32(PeekDuration), 0, 1)
		finalNormalized := rl.Clamp((1.0-stateDeltaNormalized)+elapsedNormalized, 0, 1)
		finalScale := float32(utils.EaseOutExpo(float64(finalNormalized)))
		r.s.GuiCurrentAlpha = GuiPeekAlpha + (float32(GuiNormalAlpha-GuiPeekAlpha) * finalScale)
		r.s.GuiCurrentAlphaImage = 0.0 + (float32(GuiNormalAlpha-0.0) * finalScale)
		r.s.GuiCurrentBgAlpha = GuiPeekBgAlpha + (float32(GuiNormalBgAlpha-GuiPeekBgAlpha) * finalScale)
	}

	gui.SetAlpha(r.s.GuiCurrentAlpha)
	modifiedNormalBg := GuiNormalBg
	modifiedNormalBg.A = uint8(r.s.GuiCurrentBgAlpha * 255)
	rl.DrawRectangle(0, 0, int32(r.s.CanvasSizeX), int32(r.s.CanvasSizeY), modifiedNormalBg)
}

func (r *Renderer) renderTabNavigation(startPosY float32) {
	modifiedPosY := startPosY - (OptionHeightWithPad * 2)
	for i, tab := range state.TabListAll {
		activeBefore := r.s.OptionsActiveTab == tab
		activeAfter := activeBefore
		gui.Toggle(
			rl.Rectangle{
				float32(r.s.CanvasMiddleX) - (TabRectWidth*float32(len(state.TabListAll)))/2 + TabRectWidth*float32(i) + 1,
				modifiedPosY,
				TabRectWidth,
				OptionHeightWithPad,
			},
			tab.String(),
			&activeAfter,
		)
		if !activeBefore && activeAfter {
			r.s.OptionsActiveTab = tab
		}
	}
}

func (r *Renderer) renderIconAction(
	rect, labelRect rl.Rectangle,
	icon gui.IconID, label string, color my_noire.Color,
	onLeftClick, onRightClick func(),
) {
	savedColorStyle := gui.GetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL)
	savedAlignStyle := gui.GetStyle(gui.LABEL, gui.TEXT_ALIGNMENT)
	gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, gui.NewColorPropertyValue(color.AsRGBA()))
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	gui.Label(labelRect, label)
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, savedAlignStyle)
	gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, savedColorStyle)

	if rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
		color = adjustForClickOrHover(color)

		if onLeftClick != nil && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			onLeftClick()
		} else if onRightClick != nil && rl.IsMouseButtonPressed(rl.MouseRightButton) {
			onRightClick()
		}
	}

	gui.DrawIcon(icon, int32(rect.X), int32(rect.Y), ActionIconScale, color.AsRGBA())
}

//func (r *Renderer) renderActionButtons(startPosY float32) {
//	isDirty := r.app.Config.IsDirty(r.app.SavedConfig)
//
//	saveLabel := "Save Config  "
//	saveColor := ColorUI
//	if isDirty {
//		saveLabel = "Save Config *"
//		saveColor = ColorUIHLFdev
//	}
//
//	modifiedPosY := startPosY + optionHeightWithPad*12
//
//	saveRect := rl.Rectangle{float32(r.s.CanvasMiddleX) - symmetricOffset, modifiedPosY, actionIconEdge, actionIconEdge}
//	saveLabelRect := rl.Rectangle{saveRect.X - 60, saveRect.Y - optionHeight, actionIconEdge + 120, optionHeight}
//
//	savedStyle := gui.GetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL)
//	gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, gui.NewColorPropertyValue(saveColor.AsRGBA()))
//	gui.Label(saveLabelRect, saveLabel)
//	gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, savedStyle)
//
//	if rl.CheckCollisionPointRec(r.s.MousePosition, saveRect) {
//		saveColor = r.adjustColor(saveColor)
//
//		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
//			err := r.app.ConfigStore.SaveConfig(&r.app.Config)
//			if err != nil {
//				Logger.Error("save config failed", zap.Error(err))
//			} else {
//				Logger.Info("config saved!", zap.String("path", r.app.ConfigStore.Root()))
//				r.app.SavedConfig = r.app.Config.DeepCopy()
//			}
//			r.playBoopSound()
//		} else if rl.IsMouseButtonPressed(rl.MouseRightButton) {
//			r.app.Config = *DefaultConfigReadOnly.DeepCopy()
//			r.playBoopSound()
//			Logger.Info("default config loaded!", zap.String("path", r.app.ConfigStore.Root()))
//		}
//	}
//	gui.DrawIcon(gui.ICON_FILE_SAVE_CLASSIC, int32(saveRect.X), int32(saveRect.Y), actionIconScale, saveColor.AsRGBA())
//
//	switch r.s.OptionsActiveTab {
//	default:
//	case TabGeneral, TabFilters:
//		repLabel := "reprocessing not required"
//		repColor := ColorUI
//
//		reprocessingRequired := r.app.Config.IsReprocessingRequired(r.app.ProcessedConfig)
//
//		if reprocessingRequired {
//			repLabel = "! reprocessing required !"
//			repColor = r.app.Config.Colors.GeneralWarning
//		}
//
//		repRect := rl.Rectangle{float32(r.s.CanvasMiddleX) + symmetricOffset, modifiedPosY, actionIconEdge, actionIconEdge}
//		additionalLabelSpace := float32(100)
//		repLabelRect := rl.Rectangle{repRect.X - additionalLabelSpace, repRect.Y - optionHeight, actionIconEdge + additionalLabelSpace*2, optionHeight}
//		gui.Label(repLabelRect, repLabel)
//
//		if rl.CheckCollisionPointRec(r.s.MousePosition, repRect) {
//			repColor = r.adjustColor(repColor)
//
//			if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
//				r.playBoopSound()
//				r.app.State.Jobs.Journal.RestartAsync()
//			}
//		}
//
//		icon := gui.ICON_REDO_FILL
//		gui.DrawIcon(icon, int32(repRect.X), int32(repRect.Y), actionIconScale, repColor.AsRGBA())
//	}
//}

func (r *Renderer) playBoopSound() {
	if r.app.Config.IsPonyEnjoyer {
		rl.PlaySound(r.s.FSSound)
	} else {
		rl.PlaySound(r.s.GopherSound)
	}
}

func (r *Renderer) renderDebugOverlay() {
	debugLine1 := fmt.Sprintf(
		"[Mouse Position: %4.0f x %4.0f]   [Mouse Frame Delta: %4.0f x %4.0f]   [Mouse Over Canvas: %-5t]",
		r.s.MousePosition.X, r.s.MousePosition.Y, r.s.MouseFrameDelta.X, r.s.MouseFrameDelta.Y, r.s.MouseOverWindow,
	)
	currentMonitor := rl.GetCurrentMonitor()
	MonitorPosition := rl.GetMonitorPosition(currentMonitor)
	debugLine2 := fmt.Sprintf(
		"[Canvas Size: %d x %d]   [Monitor: %d x %d @ %d, Position: %.0f x %.0f]   [Rendered FPS: %3d]",
		r.s.CanvasSizeX, r.s.CanvasSizeY, rl.GetMonitorWidth(currentMonitor), rl.GetMonitorHeight(currentMonitor), rl.GetMonitorRefreshRate(currentMonitor),
		MonitorPosition.X, MonitorPosition.Y, rl.GetFPS(),
	)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	saveAlign := gui.GetStyle(gui.LABEL, gui.TEXT_ALIGNMENT)
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	gui.Label(rl.Rectangle{0, float32(UIFontSize) * 0.5, float32(r.s.CanvasSizeX), float32(UIFontSize)}, debugLine1)
	gui.Label(rl.Rectangle{0, float32(UIFontSize) * 1.5, float32(r.s.CanvasSizeX), float32(UIFontSize)}, debugLine2)

	debugBottom := fmt.Sprintf("Alloc=%.2f MiB, Sys=%.2f MiB", float64(m.Alloc)/1024/1024, float64(m.Sys)/1024/1024)
	gui.Label(rl.Rectangle{0, float32(r.s.CanvasSizeY) - float32(UIFontSize)*1.5, float32(r.s.CanvasSizeX), float32(UIFontSize)}, debugBottom)

	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, saveAlign)
}
