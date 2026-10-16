package app

import (
	"fmt"
	"path/filepath"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	journal_app "github.com/gethiox/EDCITR/internal/app/journal"
	"go.uber.org/zap"
)

func (r *Renderer) renderTabAudio(startPosY float32) {
	modifiedPosY := startPosY + OptionHeightWithPad*0
	activeRect := rl.Rectangle{float32(r.s.CanvasMiddleX - 150), modifiedPosY, 300, OptionHeight}

	volume := r.app.Config.Audio.MasterVolume
	if gui.Slider(
		activeRect,
		"Master Volume ",
		fmt.Sprintf(" %6.2f", r.app.Config.Audio.MasterVolume*100),
		&volume,
		0,
		1,
	) { // TODO: verify renders properly
		r.app.Config.Audio.MasterVolume = volume
		r.s.VolumeReloadRequired = true
	}

	modifiedPosY = startPosY + OptionHeightWithPad*1
	activeRect = rl.Rectangle{float32(r.s.CanvasMiddleX - 150), modifiedPosY, 300, OptionHeight}
	if gui.Button(activeRect, "restart audio") {
		rl.CloseAudioDevice()
		rl.InitAudioDevice()
	}

	// Show volume controls
	save := gui.GetStyle(gui.LABEL, gui.TEXT_ALIGNMENT)
	colorSave := gui.GetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL)
	for i, ch := range journal_app.ChannelAllList {
		modifiedPosY = startPosY + OptionHeightWithPad*3 + OptionHeightWithPad*float32(i)
		activeRect = rl.Rectangle{float32(r.s.CanvasMiddleX - 150), modifiedPosY, 300, OptionHeight}

		soundPath, builtin := resolveConfigPath(r.app.Config.Audio.ChannelSoundPaths[ch])
		text := filepath.Base(soundPath)
		if text == "." {
			text = "[sound not set]"
		}
		if !builtin {
			text = fmt.Sprintf("[custom] %s", text)
		}

		gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_RIGHT)
		gui.Label(rl.Rectangle{float32(r.s.CanvasMiddleX - 150 - 300), modifiedPosY, 300 - 8, OptionHeight}, journal_app.MessageChannelToString[ch])
		gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)

		if !builtin {
			gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, gui.NewColorPropertyValue(r.app.Config.Colors.GeneralOK.AsRGBA()))
		}
		if r.s.CurrentlyEditingSound == ch {
			gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, gui.NewColorPropertyValue(ColorUIHLFdev.AsRGBA()))
		}

		gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
		gui.LabelButton(activeRect, text)
		gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, colorSave)

		var iconColor = ColorUI
		if rl.CheckCollisionPointRec(r.s.MousePosition, rl.Rectangle{float32(r.s.CanvasMiddleX) + 150 + 8, modifiedPosY - 3 + 2, 16 * 2, 16*2 - 4}) {
			iconColor = ColorUIHLFdev
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && soundPath != "" {
				rl.PlaySound(r.s.LoadedSounds[ch][0])
			}
		}
		gui.DrawIcon(gui.ICON_PLAYER_PLAY, int32(r.s.CanvasMiddleX+150+8), int32(modifiedPosY)-3, 2, iconColor.AsRGBA())

		if rl.CheckCollisionPointRec(r.s.MousePosition, activeRect) {
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
				if r.s.CurrentlyEditingSound == ch {
					r.s.CurrentlyEditingSound = journal_app.ChannelUndefined
				} else {
					r.s.CurrentlyEditingSound = ch
				}
			}
			if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
				r.app.Config.Audio.ChannelSoundPaths[ch] = ""
				r.app.unloadSound(ch)
			}
		}
	}
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, save)

	if r.s.CurrentlyEditingSound == journal_app.ChannelUndefined { // show per-channel volume sliders when no sound selected
		modifiedPosY = startPosY + OptionHeightWithPad*1
		activeRect = rl.Rectangle{float32(r.s.CanvasMiddleX + 150 + 60), modifiedPosY, 400, OptionHeight}
		txt := fmt.Sprintf("per-channel volume:")
		gui.Label(activeRect, txt)

		for i, ch := range journal_app.ChannelAllList {
			modifiedPosY = startPosY + OptionHeightWithPad*3 + OptionHeightWithPad*float32(i)
			activeRect = rl.Rectangle{float32(r.s.CanvasMiddleX - 150), modifiedPosY, 300, OptionHeight}

			//selectedVolume := gui.Slider(
			//	rl.Rectangle{float32(r.s.CanvasMiddleX + 150 + 60), modifiedPosY, 300, OptionHeight},
			//	"",
			//	fmt.Sprintf(" %6.2f", r.app.Config.Audio.ChannelVolume[ch]*100),
			//	r.app.Config.Audio.ChannelVolume[ch]*100,
			//	0,
			//	100,
			//) / 100
			//if selectedVolume != r.app.Config.Audio.ChannelVolume[ch] {
			//	r.app.Config.Audio.ChannelVolume[ch] = selectedVolume
			//	r.s.VolumeReloadRequired = true
			//}

			volume := r.app.Config.Audio.ChannelVolume[ch]
			if gui.Slider(
				rl.Rectangle{float32(r.s.CanvasMiddleX + 150 + 60), modifiedPosY, 300, OptionHeight},
				"",
				fmt.Sprintf(" %6.2f", r.app.Config.Audio.ChannelVolume[ch]*100),
				&volume,
				0,
				1,
			) { // TODO: verify renders properly
				r.app.Config.Audio.ChannelVolume[ch] = volume
				r.s.VolumeReloadRequired = true
			}
		}
	} else { // show sound list otherwise
		modifiedPosY = startPosY + OptionHeightWithPad*3 + OptionHeightWithPad*float32(len(journal_app.ChannelAllList)+1)
		activeRect = rl.Rectangle{float32(r.s.CanvasMiddleX - 200), modifiedPosY, 400, OptionHeight}

		colorSave := gui.GetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL)
		gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, gui.NewColorPropertyValue(r.app.Config.Colors.GeneralOK.AsRGBA()))
		gui.Label(activeRect, fmt.Sprintf("drop custom sound file (%s)", journal_app.MessageChannelToString[r.s.CurrentlyEditingSound]))
		gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, colorSave)

		modifiedPosY = startPosY + OptionHeightWithPad*3 + OptionHeightWithPad*float32(len(journal_app.ChannelAllList)+2)
		gui.DrawIcon(gui.ICON_AUDIO, int32(r.s.CanvasMiddleX-(16*6)/2), int32(modifiedPosY), 6, r.app.Config.Colors.GeneralOK.AsRGBA())

		if rl.IsFileDropped() {
			dropped := rl.LoadDroppedFiles()
			Logger.Info("Files dropped!", zap.Any("dropped", dropped))
			if len(dropped) > 1 {
				//gui.WindowBox()
			}
			if len(dropped) == 1 {
				r.app.Config.Audio.ChannelSoundPaths[r.s.CurrentlyEditingSound] = dropped[0]
				if r.app.Config.Audio.ChannelSoundPaths[r.s.CurrentlyEditingSound] != "" {
					r.app.unloadSound(r.s.CurrentlyEditingSound)
				}
				r.app.loadSound(r.s.CurrentlyEditingSound, r.app.Config.Audio.ChannelSoundPaths[r.s.CurrentlyEditingSound])
				rl.PlaySound(r.s.LoadedSounds[r.s.CurrentlyEditingSound][0])
			}
		}

		modifiedPosY = startPosY + OptionHeightWithPad*1
		activeRect = rl.Rectangle{float32(r.s.CanvasMiddleX + 150 + 60), modifiedPosY, 400, OptionHeight}
		txt := fmt.Sprintf("Built-in Sounds:")
		gui.Label(activeRect, txt)

		modifiedPosY = startPosY + OptionHeightWithPad*2
		activeRect = rl.Rectangle{float32(r.s.CanvasMiddleX + 150 + 60), modifiedPosY, 400, OptionHeight}
		if r.s.SoundListCurrentPos == 0 {
			gui.Label(activeRect, fmt.Sprintf(`============== (%3d ) ==============`, r.s.SoundListCurrentPos))
		} else {
			gui.Label(activeRect, fmt.Sprintf(` = = = = = = = (%3d ) = = = = = = = `, r.s.SoundListCurrentPos))
		}

		entireListRect := rl.Rectangle{float32(r.s.CanvasMiddleX + 150 + 60), modifiedPosY, 400, OptionHeightWithPad * float32(SoundListRows+2)}
		if rl.CheckCollisionPointRec(r.s.MousePosition, entireListRect) {
			r.s.SoundListCurrentPos -= int(r.getMouseWheelV().Y)
			if r.s.SoundListCurrentPos < 0 {
				r.s.SoundListCurrentPos = 0
			}
			if r.s.SoundListCurrentPos > len(r.s.BuiltInSoundList)-SoundListRows {
				r.s.SoundListCurrentPos = len(r.s.BuiltInSoundList) - SoundListRows
			}
		}

		save := gui.GetStyle(gui.LABEL, gui.TEXT_ALIGNMENT)
		gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_LEFT)
		for i, nameAndPath := range r.s.BuiltInSoundList[r.s.SoundListCurrentPos : r.s.SoundListCurrentPos+SoundListRows] {
			modifiedPosY = startPosY + OptionHeightWithPad*3 + OptionHeightWithPad*float32(i)
			activeRect = rl.Rectangle{float32(r.s.CanvasMiddleX + 150 + 60), modifiedPosY, 400, OptionHeight}

			if gui.LabelButton(activeRect, nameAndPath[0]) {
				r.app.Config.Audio.ChannelSoundPaths[r.s.CurrentlyEditingSound] = nameAndPath[1]
				if r.app.Config.Audio.ChannelSoundPaths[r.s.CurrentlyEditingSound] != "" {
					r.app.unloadSound(r.s.CurrentlyEditingSound)
				}
				r.app.loadSound(r.s.CurrentlyEditingSound, nameAndPath[1])
				rl.PlaySound(r.s.LoadedSounds[r.s.CurrentlyEditingSound][0])
			}
		}
		gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, save)

		modifiedPosY = startPosY + OptionHeightWithPad*3 + (OptionHeightWithPad * float32(SoundListRows))
		activeRect = rl.Rectangle{float32(r.s.CanvasMiddleX + 150 + 60), modifiedPosY, 400, OptionHeight}
		if r.s.SoundListCurrentPos < len(r.s.BuiltInSoundList)-SoundListRows {
			gui.Label(activeRect, fmt.Sprintf(` = = = = = = = (%3d ) = = = = = = = `, len(r.s.BuiltInSoundList)-SoundListRows-r.s.SoundListCurrentPos))
		} else {
			gui.Label(activeRect, fmt.Sprintf(`============== (%3d ) ==============`, len(r.s.BuiltInSoundList)-SoundListRows-r.s.SoundListCurrentPos))
		}
	}
}
