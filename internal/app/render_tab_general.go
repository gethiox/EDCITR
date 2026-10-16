package app

import (
	"fmt"

	gui "github.com/gen2brain/raylib-go/raygui"
	journal_app "github.com/gethiox/EDCITR/internal/app/journal"
)

func (r *Renderer) renderTabGeneral(startPosY float32) {
	//modifiedPosY := startPosY + optionHeightWithPad*0

	for row, widget := range r.settingsGeneral {
		switch widget.(type) {
		case OptionWidget:
			widget.Draw(r.optionRect(startPosY, row))
		case ButtonWidget:
			widget.Draw(r.buttonRect(startPosY, row))
		}
	}

	r.renderChannelConfiguration(startPosY)

	//modifiedPosY = startPosY + optionHeightWithPad*14
	//r.app.Config.ViewportBuffers.Appearance.Toggles.TimestampEliteFormat = gui.CheckBox(rl.Rectangle{float32(r.s.CanvasMiddleX - 210), modifiedPosY, 200 / 5, optionHeight},
	//	"[elite time]",
	//	r.app.Config.ViewportBuffers.Appearance.Toggles.TimestampEliteFormat,
	//)
	//
	//modifiedPosY = startPosY + optionHeightWithPad*15
	//r.app.Config.ViewportBuffers.Appearance.Toggles.Timestamps = gui.CheckBox(rl.Rectangle{float32(r.s.CanvasMiddleX - 210), modifiedPosY, 200 / 5, optionHeight},
	//	"(local time)",
	//	r.app.Config.ViewportBuffers.Appearance.Toggles.Timestamps,
	//)
	//
	//modifiedPosY = startPosY + optionHeightWithPad*14
	//r.app.Config.ViewportBuffers.Appearance.Toggles.PrefixCMDR = gui.CheckBox(rl.Rectangle{float32(r.s.CanvasMiddleX + 10), modifiedPosY, 200 / 5, optionHeight},
	//	"CMDR Prefix",
	//	r.app.Config.ViewportBuffers.Appearance.Toggles.PrefixCMDR,
	//)
	//
	//modifiedPosY = startPosY + optionHeightWithPad*15
	//r.app.Config.ViewportBuffers.Appearance.Toggles.PrefixNPC = gui.CheckBox(rl.Rectangle{float32(r.s.CanvasMiddleX + 10), modifiedPosY, 200 / 5, optionHeight},
	//	"NPC Prefix",
	//	r.app.Config.ViewportBuffers.Appearance.Toggles.PrefixNPC,
	//)
	//
	//modifiedPosY = startPosY + optionHeightWithPad*16
	//labelFontRect := rl.Rectangle{float32(r.s.CanvasMiddleX - 300), modifiedPosY, 600, optionHeight}
	//labelFontRect.X -= 400
	//labelFontRect.Width = 400
	//gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_RIGHT)
	//gui.Label(labelFontRect, "Font ")
	//gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	//
	//modifiedPosY = startPosY + optionHeightWithPad*17
	//wordListRect := rl.Rectangle{float32(r.s.CanvasMiddleX - 300), modifiedPosY, 600, optionHeight}
	//ok := r.renderEditField(wordListRect, &r.app.State.WordListEditField, &r.s.HighlightWordsEditing, &r.s.HighlightWordsEditingWarn, "Highlight Word List")
	//if ok {
	//	wordList := strings.Split(strings.ToLower(r.app.State.WordListEditField), ",")
	//	r.app.Config.ViewportBuffers.Appearance.HighlightWords = wordList
	//	r.app.State.WordList = make(map[string]bool, len(wordList))
	//	for _, word := range wordList {
	//		r.app.State.WordList[word] = true
	//	}
	//}
}

func (r *Renderer) renderChannelConfiguration(startPosY float32) {
	save := gui.GetStyle(gui.LABEL, gui.TEXT_ALIGNMENT)
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_RIGHT)
	for row, channel := range journal_app.ChannelAllList {
		gui.Label(
			r.channelLabelRect(startPosY, row),
			journal_app.MessageChannelToString[channel],
		)
	}
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, save)

	for column := 0; column < r.app.Config.Viewports.Count; column++ {
		gui.Label( // viewport label number
			r.viewportLabelRect(startPosY, column),
			fmt.Sprintf("%d", column+1),
		)

		for row, channel := range journal_app.ChannelAllList {
			checked := r.app.Config.Viewports.ViewportChannelBitmask[column]&channel == channel
			updated := gui.CheckBox(
				r.viewportCheckboxRect(startPosY, row, column),
				"",
				&checked,
			)
			if updated {
				if checked {
					r.app.Config.Viewports.ViewportChannelBitmask[column] |= channel
				} else {
					r.app.Config.Viewports.ViewportChannelBitmask[column] ^= channel
				}
			}

		}
	}
}
