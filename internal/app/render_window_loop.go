package app

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/raylib-go/raygui"
	"github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/assets"
	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/storage"
	"github.com/gethiox/EDCITR/internal/utils"
	"go.uber.org/zap"
)

func (r *Renderer) windowLoop() {
	var (
		deltaTime       float64
		prevTime        float64
		currentTime     float64
		prevCanvasSizeX int
		prevCanvasSizeY int
		cogCurrentValue float32

		normalRender  = true
		RenderTexture = rl.LoadRenderTexture(int32(r.s.CanvasSizeX+RenderPad), int32(r.s.CanvasSizeY+RenderPad))
		GxShader      = rl.LoadShader("", "cmd/gui/shaders/gx.fs")
	)

	//s := rl.LoadMusicStream("assets/tracks/unreeeal_superhero_3.xm")
	//rl.PlayMusicStream(s)

	Logger.Info("config status", zap.Bool("dirty", r.app.Config.IsDirty(r.app.SavedConfig)))

ring0:
	for !rl.WindowShouldClose() {
		//rl.UpdateMusicStream(s)
		r.s.CurrentTime.Store(time.Now().UnixNano())

		prevCanvasSizeX, prevCanvasSizeY = r.s.CanvasSizeX, r.s.CanvasSizeY
		r.s.CanvasSizeX, r.s.CanvasSizeY = rl.GetRenderWidth(), rl.GetRenderHeight()
		r.s.CanvasMiddleX, r.s.CanvasMiddleY = r.s.CanvasSizeX/2, r.s.CanvasSizeY/2

		prevTime = currentTime
		currentTime = rl.GetTime()
		deltaTime = currentTime - prevTime

		r.s.IsCurrentlyHoveringOverCmdrName = false
		r.s.IsCurrentlyHoveringOverExtraParameters = false

		for ViewportID := 0; ViewportID < r.app.Config.Viewports.Count; ViewportID++ {
			r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollBrightenByUser *= math.Pow(0.95, deltaTime*float64(r.app.Config.FPS))
		}

		for column := range r.app.State.ViewportBuffers {
			r.app.State.ViewportBuffers[column].Scroll.ScrollBrightenByUser *= math.Pow(0.95, deltaTime*float64(r.app.Config.FPS))
		}

		if r.s.CanvasSizeX != prevCanvasSizeX || r.s.CanvasSizeY != prevCanvasSizeY {
			Logger.Debug("Resizing render texture")
			rl.UnloadRenderTexture(RenderTexture)
			RenderTexture = rl.LoadRenderTexture(int32(r.s.CanvasSizeX+RenderPad), int32(r.s.CanvasSizeY+RenderPad))

			// TODO: handle it better
			r.app.State.LogsVerticalPeek = min(
				r.app.State.LogsVerticalPeek,
				float32(r.s.CanvasSizeY)-float32(r.s.CanvasSizeY)*0.2,
			)

			r.app.State.Inara.CMDRNamesRefreshAvailable.Store(true)
		}

		if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
			r.app.State.ActivelyGrabbingStatusbar = false
		}

		if r.app.State.ActivelyGrabbingStatusbar {
			r.app.State.LogsVerticalPeek = min(
				max(r.app.State.LogsVerticalPeek-r.s.MouseFrameDelta.Y, 0),
				float32(r.s.CanvasSizeY)-float32(r.s.CanvasSizeY)*0.2,
			)
		}

		if r.s.FontReloadRequired {
			type setting struct {
				rawPath     string
				destination *rl.Font
			}

			for _, s := range []setting{
				{r.app.Config.Fonts.MainPath, &r.app.State.Fonts.Main},
				{r.app.Config.Fonts.MonoPath, &r.app.State.Fonts.MainMono},
			} {
				fontPath, builtin := resolveConfigPath(s.rawPath)
				nameParts := strings.Split(filepath.Base(fontPath), ".")
				ext := nameParts[len(nameParts)-1]

				var fontData []byte
				var err error

				if builtin {
					fontData, err = assets.FontsFS.ReadFile(fontPath)
				} else {
					fontData, err = os.ReadFile(fontPath)
				}

				if err != nil {
					Logger.Error("Load font failed", zap.Bool("builtin", builtin), zap.String("path", fontPath), zap.Error(err))
				}

				rl.UnloadFont(*s.destination)
				*s.destination = rl.LoadFontFromMemory("."+ext, fontData, int32(r.app.Config.Fonts.Size), codepoints)
				Logger.Debug("Font Loaded", zap.Bool("builtin", builtin), zap.String("path", fontPath))
			}

			r.s.FontReloadRequired = false
		}

		if r.s.FPSReloadRequired {
			rl.SetTargetFPS(int32(r.app.Config.FPS))
			r.s.FPSReloadRequired = false
		}

		if r.s.VolumeReloadRequired {
			rl.SetMasterVolume(r.app.Config.Audio.MasterVolume)
			for ch, soundName := range r.app.Config.Audio.ChannelSoundPaths {
				if soundName == "" {
					continue
				}

				for i := 0; i < MaxSounds; i++ {
					r.app.setVolume(ch, r.app.Config.Audio.ChannelVolume[ch])
				}
			}

			r.s.VolumeReloadRequired = false
		}

		if r.s.SoundListReloadRequired {
			r.s.SoundListCurrentPos = 0
			r.app.loadSoundList()
			r.s.SoundListReloadRequired = false
		}

		if r.app.State.AudioReloadRequired {
			r.app.State.AudioReloadRequired = false
			rl.CloseAudioDevice()
			rl.InitAudioDevice()
		}

		r.s.MouseFrameDelta = rl.GetMouseDelta()
		r.s.MousePosition = rl.GetMousePosition()

		//pad := float32(40)
		//boundsCheck := rl.Rectangle{0 + pad, 0 + pad, float32(r.s.CanvasSizeX) - pad*2, float32(r.s.CanvasSizeY) - pad*2}
		//switch {
		//case !r.s.MouseOverWindow && rl.CheckCollisionPointRec(r.s.MousePosition, boundsCheck):
		//	r.s.MouseOverWindow = true
		//	r.s.MouseOverWindowEngaged = time.Now()
		//case r.s.MouseOverWindow && !rl.CheckCollisionPointRec(r.s.MousePosition, boundsCheck):
		//	r.s.MouseOverWindow = false
		//	r.s.MouseOverWindowDisengaged = time.Now()
		//}

		switch {
		case !r.s.MouseOverWindow && rl.IsCursorOnScreen():
			r.s.MouseOverWindow = true
			r.s.MouseOverWindowEngaged = time.Now()
		case r.s.MouseOverWindow && !rl.IsCursorOnScreen():
			r.s.MouseOverWindow = false
			r.s.MouseOverWindowDisengaged = time.Now()
		}

		toLoadNumber := 120 / float32(r.app.Config.FPS)
		if toLoadNumber < 1 {
			toLoadNumber = 1
		}

		ok := r.s.ProfilePicturesToLoadMtx.TryLock()
		if ok {
			if len(r.s.ProfilePicturesToLoad) > 0 {
				iters := min(len(r.s.ProfilePicturesToLoad), int(toLoadNumber))
				for i := 0; i < iters; i++ {
					info := r.s.ProfilePicturesToLoad[i]
					rlImg := rl.NewImageFromImage(info.Img)
					tex := rl.LoadTextureFromImage(rlImg)
					rl.SetTextureFilter(tex, rl.FilterPoint)
					r.app.State.Inara.CMDRTextures.Set(info.CmdrName, tex)
					r.app.State.Inara.CMDRTexturesState.Set(info.CmdrName, state.StateOK)
					r.app.State.Inara.CMDRTexturesKeepAlive.Set(info.CmdrName, r.s.CurrentTime.Load())
					rl.UnloadImage(rlImg)
				}
				r.s.ProfilePicturesToLoad = r.s.ProfilePicturesToLoad[iters:]
			}
			r.s.ProfilePicturesToLoadMtx.Unlock()
		}

		// DRAWING /////////////////////////////////////////////////////////////////////////////////////////////////
		rl.BeginTextureMode(RenderTexture) // rl.BeginDrawing()

		rl.ClearBackground(r.app.Config.Colors.Background.AsRGBA())

		if r.s.ShowLogs {
			r.renderLogs(float32(r.s.CanvasSizeY))
		} else {
			horizontalRealEstate := r.s.CanvasSizeX
			horizontalRealEstate -= r.renderStats()

			r.renderLogs(r.app.State.LogsVerticalPeek)
			reserved := r.renderStatusBar()

			totalColumnWidth := float32(horizontalRealEstate-
				r.app.Config.Viewports.BorderSize*(r.app.Config.Viewports.Count-1)-
				r.app.Config.Viewports.Appearance.LeftPad*(r.app.Config.Viewports.Count)-
				r.app.Config.Viewports.Appearance.RightPad*(r.app.Config.Viewports.Count)) / float32(r.app.Config.Viewports.Count)

			r.renderViewports(totalColumnWidth, reserved)
			r.renderViewportBorders(totalColumnWidth, reserved)

			if r.app.State.ProcessingLiveEvents.Load() {
				if r.s.IsCurrentlyHoveringOverExtraParameters {
					r.renderHoverNotification()
				}

				if r.s.IsCurrentlyHoveringOverCmdrName {
					r.renderHoverProfileInfo()
				}

				r.renderBubbleMessage()

			}

			if r.app.State.ForceShowPing || r.app.State.ShowPingHover {
				r.renderServerStatusHistogram(reserved)
			}
		}

		// display loading overlay
		if !r.app.State.ProcessingLiveEvents.Load() {
			frame := getEdLoaderFrame()
			rect := rl.Rectangle{float32(EdLoaderFrameWidth) * float32(frame), 0, float32(EdLoaderFrameWidth), float32(r.app.State.Textures.EliteLoader.Height)}
			posTexture := rl.Vector2{float32(r.s.CanvasSizeX/2) - rect.Width/2, float32(r.s.CanvasSizeY/2) - rect.Height/2}

			modifiedNormalBg := GuiNormalBg
			modifiedNormalBg.A = uint8(255 / 2)
			rl.DrawRectangle(0, 0, int32(r.s.CanvasSizeX), int32(r.s.CanvasSizeY), modifiedNormalBg)
			rl.DrawTextureRec(r.app.State.Textures.EliteLoader, rect, posTexture, rl.White)

			text := fmt.Sprintf("Loading previous journal files, please wait")
			m := rl.MeasureTextEx(r.app.State.Fonts.UI, text, float32(UIFontSize), float32(UIFontSpacing))
			posText := rl.Vector2{float32(r.s.CanvasSizeX/2) - m.X/2, float32(r.s.CanvasSizeY/2) + rect.Height/2 + OptionHeightWithPad*1}
			rl.DrawTextEx(r.app.State.Fonts.UI, text, posText, UIFontSize, UIFontSpacing, rl.White)

			text = r.s.PreloadMessage.Load()
			m = rl.MeasureTextEx(r.app.State.Fonts.UI, text, float32(UIFontSize), float32(UIFontSpacing))
			posText = rl.Vector2{float32(r.s.CanvasSizeX/2) - m.X/2, float32(r.s.CanvasSizeY/2) + rect.Height/2 + OptionHeightWithPad*2}
			progress := float64(r.s.PreloadCurrent.Load()) / float64(r.s.PreloadMax)
			//c := rl.ColorFromHSV(progress*(360/3), 0.7, 1)
			rr, gg, bb := GetGradientColor(progress).Clamped().RGB255()

			rl.DrawTextEx(r.app.State.Fonts.UI, text, posText, UIFontSize, UIFontSpacing, rl.Color{rr, gg, bb, 255})

			posProgress := rl.Rectangle{float32(r.s.CanvasSizeX/2) - rect.Width/2, float32(r.s.CanvasSizeY/2) + rect.Height/2 + OptionHeightWithPad*3, rect.Width, OptionHeightWithPad}

			progressValue := float32(progress)
			raygui.ProgressBar(posProgress, "", "", &progressValue, 0, 1)
		}

		cogColor := r.app.Config.Colors.GeneralGray
		cogColor.Alpha = 0.4
		iconSize := max(r.getHeaderHeight(), 40)
		//goland:noinspection ALL
		cogRect := rl.Rectangle{
			float32(r.s.CanvasSizeX) - iconSize,
			0,
			iconSize,
			iconSize,
		}
		cogRectForDraw := rl.Rectangle{
			float32(r.s.CanvasSizeX) - iconSize/2,
			iconSize / 2,
			iconSize,
			iconSize,
		}
		if rl.CheckCollisionPointRec(r.s.MousePosition, cogRect) {
			if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				r.toggleSettingsView()
			}
			cogColor.Alpha = 0.8
			cogColor = cogColor.Lighten(HoverBrighten)
		}

		if r.s.OptionViewActive {
			cogCurrentValue += 45 * rl.GetFrameTime()
			if cogCurrentValue > 360 {
				cogCurrentValue -= 360
			}
			cogColor.Alpha = 0.8

			r.renderOptionView()
		}

		rl.DrawTexturePro(
			r.app.State.Textures.Cog.Texture,
			rl.Rectangle{0, 0, float32(r.app.State.Textures.Cog.Texture.Width), float32(r.app.State.Textures.Cog.Texture.Height)},
			cogRectForDraw,
			rl.Vector2{cogRect.Width / 2, cogRect.Height / 2},
			cogCurrentValue, // Just use the value directly
			cogColor.AsRGBA(),
		)

		rl.EndTextureMode() // rl.EndDrawing()

		if !r.s.DisableShortcuts {
			shouldExit := r.handleKeyboardShortcuts()
			if shouldExit {
				break ring0
			}
		}

		rl.BeginDrawing()
		if rl.IsKeyPressed(rl.KeyUp) {
			rl.UnloadShader(GxShader)

			// language=file-reference
			GxShader = rl.LoadShader("", "cmd/gui/shaders/gx.fs")
			//GxShader = rl.LoadShaderFromMemory("", string(unsafeLoadData(Shaders, "shaders/gx.fs")))
		}
		if rl.IsKeyPressed(rl.KeyDown) {
			normalRender = !normalRender
		}
		rl.ClearBackground(rl.Black)
		if normalRender {
			rl.DrawTextureRec(
				RenderTexture.Texture,
				rl.Rectangle{0, 0, float32(RenderTexture.Texture.Width), -float32(RenderTexture.Texture.Height)},
				rl.Vector2{},
				rl.White,
			)
		} else {
			rl.BeginShaderMode(GxShader)
			rl.DrawTextureRec(
				RenderTexture.Texture,
				rl.Rectangle{0, 0, float32(RenderTexture.Texture.Width), -float32(RenderTexture.Texture.Height)},
				rl.Vector2{},
				rl.White,
			)
			rl.EndShaderMode()
		}

		rl.EndDrawing()

		r.cleanupUnusedTextures()
		r.cleanupUnusedProfiles()
		//r.s.CurrentFrame++
	}
}

func (r *Renderer) cleanupUnusedTextures() {
	var toRemove []string

	r.app.State.Inara.CMDRTextures.IterCb(func(cmdrName string, tex rl.Texture2D) {
		lastDrawn, ok := r.app.State.Inara.CMDRTexturesKeepAlive.Get(cmdrName)
		if !ok {
			return
		}

		if r.s.CurrentTime.Load()-lastDrawn > r.s.UnloadPeriod.Load() {
			toRemove = append(toRemove, cmdrName)
		}
	})

	for _, cmdrName := range toRemove {
		tex, ok := r.app.State.Inara.CMDRTextures.Get(cmdrName)
		if !ok {
			Logger.Warn("failed to get texture for removal", zap.String("CMDR", cmdrName))
			continue
		}

		r.app.State.Inara.CMDRTexturesKeepAlive.Remove(cmdrName)
		r.app.State.Inara.CMDRTextures.Remove(cmdrName)
		r.app.State.Inara.CMDRTexturesState.Remove(cmdrName)
		rl.UnloadTexture(tex)
		Logger.Debug("Unloaded unused texture", zap.String("CMDR", cmdrName))

	}
}

func (r *Renderer) cleanupUnusedProfiles() {
	var toRemove []string

	r.app.State.Inara.CMDRInfo.IterCb(func(cmdrName string, info storage.InfoWithTime) {
		lastDrawn, ok := r.app.State.Inara.CMDRInfoKeepAlive.Get(cmdrName)
		if !ok {
			return
		}

		if r.s.CurrentTime.Load()-lastDrawn > r.s.UnloadPeriod.Load() {
			toRemove = append(toRemove, cmdrName)
		}
	})

	for _, cmdrName := range toRemove {
		_, ok := r.app.State.Inara.CMDRInfo.Get(cmdrName)
		if !ok {
			Logger.Warn("failed to get info for removal", zap.String("CMDR", cmdrName))
			continue
		}

		r.app.State.Inara.CMDRInfoKeepAlive.Remove(cmdrName)
		r.app.State.Inara.CMDRInfo.Remove(cmdrName)
		r.app.State.Inara.CMDRInfoState.Remove(cmdrName)

		Logger.Debug("Unloaded profile info", zap.String("CMDR", cmdrName))

	}
}

func (r *Renderer) renderBubbleMessage() {
	since := time.Since(r.app.State.BubbleMessage.EventTime)
	if since > time.Second*3 {
		return
	}

	fontSize := float32(r.app.Config.Fonts.Size)

	m := rl.MeasureTextEx(r.app.State.Fonts.Main, r.app.State.BubbleMessage.Text, fontSize, float32(r.app.Config.Fonts.Spacing))
	pos := rl.Vector2{
		r.app.State.BubbleMessage.MousePos.X - m.X/2,
		r.app.State.BubbleMessage.MousePos.Y - fontSize,
	}
	bgRect := rl.Rectangle{
		pos.X - fontSize/2,
		pos.Y - fontSize/2,
		m.X + fontSize,
		fontSize + fontSize,
	}

	var (
		bgColor = r.app.Config.Colors.Background
		fgColor = r.app.Config.Colors.GeneralNormal
	)

	if since > time.Second {
		v := since.Seconds() - 1
		v = max(v, 0)
		v = min(v, 2)
		v = utils.EaseOutExpo(v / 2)
		bgColor.Alpha *= 1 - v
		fgColor.Alpha *= 1 - v
	}

	rl.DrawRectangleRec(bgRect, bgColor.AsRGBA())
	rl.DrawTextEx(r.app.State.Fonts.Main, r.app.State.BubbleMessage.Text, pos, fontSize, float32(r.app.Config.Fonts.Spacing), fgColor.AsRGBA())
	rl.DrawRectangleLinesEx(bgRect, 2, fgColor.AsRGBA())
}
