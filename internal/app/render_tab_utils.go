package app

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/parser"
	"go.uber.org/zap"
)

func colorSettingRect(posXstart, startPosY float32, index int) rl.Rectangle {
	row := float32(index / ColorsPerRow)
	col := float32(index % ColorsPerRow)
	return rl.Rectangle{
		posXstart + col*ColorWidth,
		startPosY + row*float32(int(ColorHeight)),
		ColorWidth,
		ColorHeight,
	}
}

func shrinkRect(rect rl.Rectangle, amount float32) rl.Rectangle {
	return rl.Rectangle{
		rect.X + amount,
		rect.Y + amount,
		rect.Width - amount*2,
		rect.Height - amount*2,
	}
}

func (r *Renderer) optionRect(startPosY float32, row int) rl.Rectangle {
	return rl.Rectangle{
		float32(r.s.CanvasMiddleX - OptionWidth/2),
		startPosY + float32(OptionHeightWithPad*row),
		OptionWidth,
		OptionHeight,
	}
}

func (r *Renderer) buttonRect(startPosY float32, row int) rl.Rectangle {
	return rl.Rectangle{
		float32(r.s.CanvasMiddleX - OptionWidth/2),
		startPosY + float32(OptionHeightWithPad*row),
		OptionWidth / 8,
		OptionHeight,
	}
}

func (r *Renderer) channelLabelRect(startPosY float32, row int) rl.Rectangle {
	return rl.Rectangle{
		float32(r.s.CanvasMiddleX + ViewportSettingsOffset - OptionWidth - CheckboxSpace),
		startPosY + float32(OptionHeightWithPad*row) + OptionHeightWithPad,
		OptionWidth,
		OptionHeight,
	}
}

func (r *Renderer) viewportLabelRect(startPosY float32, column int) rl.Rectangle {
	return rl.Rectangle{
		float32(r.s.CanvasMiddleX + ViewportSettingsOffset + (CheckboxWidth+CheckboxSpace)*column),
		startPosY,
		float32(CheckboxWidth),
		OptionHeight,
	}
}

func (r *Renderer) viewportCheckboxRect(startPosY float32, row, column int) rl.Rectangle {
	return rl.Rectangle{
		float32(r.s.CanvasMiddleX + ViewportSettingsOffset + (CheckboxWidth+CheckboxSpace)*column),
		startPosY + OptionHeightWithPad + (OptionHeightWithPad * float32(row)),
		float32(CheckboxWidth),
		OptionHeight,
	}
}

func (r *Renderer) togglePing() {
	if r.app.Config.ServerStatus.Enabled {
		r.app.State.Tasks.ServerStatus.Start()
	} else {
		r.app.State.Tasks.ServerStatus.Stop()
	}
}

func (r *Renderer) toggleInara() {
	if r.app.Config.InaraSettings.Enabled {
		r.app.State.Tasks.Inara.Start()
	} else {
		r.app.State.Tasks.Inara.Stop()
	}
}

func (r *Renderer) toggleUpdateChecker() {
	if r.app.Config.StatusBar.CheckForUpdates {
		r.app.State.Tasks.UpdateChecker.Start()
	} else {
		r.app.State.Tasks.UpdateChecker.Stop()
	}
}

func (r *Renderer) disableGlobalHotkeys(editing bool) {
	r.s.IsEditing = editing
}

func (r *Renderer) updateParserSettings() {
	settings := parser.Settings{
		MaxConcurrentFiles:      r.app.Config.Parser.MaxConcurrentFiles,
		ProcessorWorkers:        r.app.Config.Parser.ProcessorWorkers,
		ProcessorWorkerChanSize: r.app.Config.Parser.ProcessorWorkerChanSize,
		ProcessorResultChanSize: r.app.Config.Parser.ProcessorResultChanSize,
	}
	parser.Configure(settings)
}

func (r *Renderer) clearCache() {
	err := r.app.CacheStore.ClearCache()
	if err != nil {
		Logger.Error("Failed to clear cache", zap.Error(err))
		return
	}
	Logger.Info("Cache cleared")
}

func (r *Renderer) deleteConfig() {
	err := r.app.ConfigStore.RemoveConfig()
	if err != nil {
		Logger.Error("Failed to delete config", zap.Error(err))
		return
	}
	Logger.Info("Config deleted")
	r.app.Config = *r.app.DefaultConfigReadOnly.DeepCopy()
}
