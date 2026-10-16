package app

import (
	"fmt"
	"time"

	"github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/elite_status"
	my_noire "github.com/gethiox/EDCITR/internal/noire"
	"github.com/gethiox/EDCITR/internal/utils"
	"github.com/gethiox/EDCITR/internal/version"
)

func (r *Renderer) getStatusBarHeight() float32 {
	return float32(r.app.Config.Fonts.Size) * 1.5
}

func (r *Renderer) getRealStatusBarHeight() float32 {
	if r.app.Config.StatusBar.Enabled {
		return r.getStatusBarHeight()
	}
	return 0
}

func (r *Renderer) renderStatusBar() (reserved float32) {
	// draw "scrolling mode" header

	if !r.app.Config.StatusBar.Enabled {
		return 0
	}

	fontSize := float32(r.app.Config.Fonts.Size)
	headerHeight := r.getStatusBarHeight()

	rect := rl.Rectangle{0, float32(r.s.CanvasSizeY) - headerHeight - r.app.State.LogsVerticalPeek, float32(r.s.CanvasSizeX), headerHeight}

	if !r.s.OptionViewActive && rl.CheckCollisionPointRec(r.s.MousePosition, rect) {
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			r.app.State.ActivelyGrabbingStatusbar = true
		}
		if rl.IsMouseButtonPressed(rl.MouseButtonMiddle) {
			r.app.State.LogsVerticalPeek = 0
		}
	}

	rl.DrawRectangleGradientV(
		int32(rect.X),
		int32(rect.Y),
		int32(rect.Width),
		int32(rect.Height),
		r.app.Config.Colors.StatusBar1.AsRGBA(),
		r.app.Config.Colors.StatusBar2.AsRGBA(),
	)

	var offsetY = headerHeight/2 - fontSize/2

	if r.app.Config.StatusBar.CurrentJournal {
		r.renderCurrentJournal(rect, offsetY)
	}

	if r.app.Config.StatusBar.InaraStats {
		r.renderInaraStats(rect, offsetY)
	}

	var reservedX float32

	if r.app.Config.ServerStatus.Enabled {
		reservedX += r.renderServerStatus(rect, offsetY, reservedX)
	}

	if r.app.Config.StatusBar.NextMaintenance {
		reservedX += r.renderNextMaintenance(rect, offsetY, reservedX)
	}

	if r.app.Config.StatusBar.CheckForUpdates {
		reservedX += r.renderUpdateChecker(rect, offsetY, reservedX)
	}

	return headerHeight + r.app.State.LogsVerticalPeek
}

func (r *Renderer) renderUpdateChecker(rect rl.Rectangle, offsetY, reservedX float32) float32 {
	fontSize := float32(r.app.Config.Fonts.Size)
	fontSpacing := float32(r.app.Config.Fonts.Spacing)
	posY := rect.Y + offsetY

	var updateText1 string
	var updateText2 string
	var color1 my_noire.Color
	var color2 my_noire.Color
	var color = r.app.Config.Colors.GeneralDarkGray

	switch r.app.State.UpdateCheckState.Load() {
	case state.StateOKButMissing:
		updateText1 = "not found "
		color1 = r.app.Config.Colors.GeneralDarkGray
	case state.StateError:
		updateText1 = "error     "
		color1 = r.app.Config.Colors.GeneralCritical
	case state.StateDownloading:
		updateText1 = "checking.."
		color1 = r.app.Config.Colors.GeneralDarkGray
	case state.StateOK:
		ver := r.app.State.UpdateCheckVersion.Load()
		if ver.GreaterThan(version.Release) {
			updateText1 = version.Release.FormatFull()
			updateText2 = ver.FormatFull()

			currentReleaseTime, ok := getCommitTime()
			if !ok {
				color1 = r.app.Config.Colors.GeneralWarning
			} else {
				age := time.Since(currentReleaseTime)
				switch {
				case age < 24*time.Hour*90:
					color1 = r.app.Config.Colors.GeneralWarning
				default:
					color1 = r.app.Config.Colors.GeneralCritical
				}
			}
			color2 = r.app.Config.Colors.GeneralOK
		} else {
			color1 = r.app.Config.Colors.GeneralDarkGray
			updateText1 = "up to date"
		}
	default:
		updateText1 = "not found "
		color1 = r.app.Config.Colors.GeneralDarkGray
	}

	var objects = make(RenderObjects, 0)
	objects = append(objects,
		RenderObject{" [", color, &r.app.State.Fonts.Main, fontSize, fontSpacing},
	)

	if updateText2 != "" {
		objects = append(objects,
			RenderObject{updateText1, color1, &r.app.State.Fonts.Main, fontSize, fontSpacing},
			RenderObject{" > ", color, &r.app.State.Fonts.Main, fontSize, fontSpacing},
			RenderObject{updateText2, color2, &r.app.State.Fonts.Main, fontSize, fontSpacing},
		)
	} else {
		objects = append(objects,
			RenderObject{updateText1, color1, &r.app.State.Fonts.Main, fontSize, fontSpacing},
		)
	}

	objects = append(objects,
		RenderObject{"] ", color, &r.app.State.Fonts.Main, fontSize, fontSpacing},
	)

	var totalWidth = objects.Width()

	pos := rl.Vector2{rect.X + rect.Width - reservedX - totalWidth, posY}

	objects.Draw(pos)

	return totalWidth
}

func (r *Renderer) renderCurrentJournal(rect rl.Rectangle, offsetY float32) {
	fontSize := float32(r.app.Config.Fonts.Size)
	posY := rect.Y + offsetY

	text1 := "current journal: "
	m1 := rl.MeasureTextEx(r.app.State.Fonts.Main, text1, fontSize, float32(r.app.Config.Fonts.Spacing))
	text2color := r.app.Config.Colors.GeneralDarkGray
	text2 := "None"
	curLogFile := r.app.State.CurrentThings.LogFile.Load()
	if curLogFile != "" {
		text2 = curLogFile
		text2color = r.app.Config.Colors.GeneralOK
	}

	var offsetX = fontSize / 2

	rl.DrawTextEx(r.app.State.Fonts.Main, text1, rl.Vector2{rect.X + offsetX, posY}, fontSize, float32(r.app.Config.Fonts.Spacing), r.app.Config.Colors.GeneralDarkGray.AsRGBA())
	rl.DrawTextEx(r.app.State.Fonts.Main, text2, rl.Vector2{rect.X + offsetX + m1.X, posY}, fontSize, float32(r.app.Config.Fonts.Spacing), text2color.AsRGBA())
}

func (r *Renderer) genASDF() []ASDF {
	var (
		infoCount        = r.app.State.Inara.CMDRInfo.Count()
		infoQueue        = r.app.State.Inara.CMDRInfoQueued.Load()
		infoCachePresent = r.app.CacheStore.GetPresentInfoCount()
		infoCacheMissing = r.app.CacheStore.GetMissingInfoCount()

		texCount        = r.app.State.Inara.CMDRTextures.Count()
		texQueue        = r.app.PFPDownloader.InQueue()
		texCachePresent = r.app.CacheStore.GetPresentPictureCount()
		texCacheMissing = r.app.CacheStore.GetMissingPictureCount()

		CountColor        = r.app.Config.Colors.GeneralOK.Shade(StatusDarken)
		QueueColor        = r.app.Config.Colors.GeneralWarning.Shade(StatusDarken)
		CachePresentColor = r.app.Config.Colors.GeneralOK.Shade(StatusDarken)
		CacheMissingColor = r.app.Config.Colors.GeneralWarning.Shade(StatusDarken)
	)

	ret := []ASDF{
		{"[", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{"Profiles", r.app.Config.Colors.GeneralGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{": ", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{fmt.Sprintf("%d", infoCount), CountColor, &r.app.State.Fonts.MainMono, float32(r.app.Config.Fonts.SpacingMono)},
		{" / ", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{fmt.Sprintf("%d", infoQueue), QueueColor, &r.app.State.Fonts.MainMono, float32(r.app.Config.Fonts.SpacingMono)},
		{" (", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{fmt.Sprintf("%d", infoCachePresent), CachePresentColor, &r.app.State.Fonts.MainMono, float32(r.app.Config.Fonts.SpacingMono)},
		{"/", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{fmt.Sprintf("%d", infoCacheMissing), CacheMissingColor, &r.app.State.Fonts.MainMono, float32(r.app.Config.Fonts.SpacingMono)},
		{")]   ", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},

		{"[", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{"Images", r.app.Config.Colors.GeneralGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{": ", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{fmt.Sprintf("%d", texCount), CountColor, &r.app.State.Fonts.MainMono, float32(r.app.Config.Fonts.SpacingMono)},
		{" / ", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{fmt.Sprintf("%d", texQueue), QueueColor, &r.app.State.Fonts.MainMono, float32(r.app.Config.Fonts.SpacingMono)},
		{" (", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{fmt.Sprintf("%d", texCachePresent), CachePresentColor, &r.app.State.Fonts.MainMono, float32(r.app.Config.Fonts.SpacingMono)},
		{"/", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		{fmt.Sprintf("%d", texCacheMissing), CacheMissingColor, &r.app.State.Fonts.MainMono, float32(r.app.Config.Fonts.SpacingMono)},
		{")]", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
	}

	clientReady := r.app.InaraClient.Ready()
	downloaderReady := r.app.PFPDownloader.Ready()

	if !clientReady || !downloaderReady {
		ret = append(ret,
			ASDF{"   [", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
			ASDF{"Awaiting full login", r.app.Config.Colors.GeneralWarning.Shade(StatusDarken), &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
			ASDF{"]", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		)
	} else {
		ret = append(ret,
			ASDF{"   [", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
			ASDF{"Client active", r.app.Config.Colors.GeneralOK.Shade(StatusDarken), &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
			ASDF{"]", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, float32(r.app.Config.Fonts.Spacing)},
		)
	}

	return ret
}

func (r *Renderer) renderInaraStats(rect rl.Rectangle, offsetY float32) {
	fontSize := float32(r.app.Config.Fonts.Size)
	posY := rect.Y + offsetY

	elements := r.genASDF()

	var widthX float32
	for _, element := range elements {
		widthX += rl.MeasureTextEx(*element.font, element.text, fontSize, element.spacing).X
	}

	var cumX float32
	for _, element := range elements {
		rl.DrawTextEx(*element.font, element.text, rl.Vector2{rect.Width/2 - widthX/2 + cumX, posY}, fontSize, element.spacing, element.color.AsRGBA())
		cumX += rl.MeasureTextEx(*element.font, element.text, fontSize, element.spacing).X
	}
}

func (r *Renderer) renderNextMaintenance(rect rl.Rectangle, offsetY, reservedX float32) float32 {
	fontSize := float32(r.app.Config.Fonts.Size)
	fontSpacing := float32(r.app.Config.Fonts.Spacing)
	posY := rect.Y + offsetY

	next, _ := utils.GetNextOccurrenceUTC(time.Now().Add(-time.Minute*60), "Thursday", "07:00")
	var nextColor my_noire.Color
	nextDuration := next.Sub(time.Now())

	switch {
	case nextDuration < 0:
		nextColor = r.app.Config.Colors.GeneralCritical
	case nextDuration < time.Hour:
		nextColor = r.app.Config.Colors.GeneralWarning
	default:
		nextColor = r.app.Config.Colors.GeneralDarkGray
	}

	nextText := fmt.Sprintf("next maintenance: %s", utils.Ago.Format(next))

	var objects = RenderObjects{
		{" [", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, fontSize, fontSpacing},
		{nextText, nextColor, &r.app.State.Fonts.Main, fontSize, fontSpacing},
		{"] ", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, fontSize, fontSpacing},
	}

	var totalWidth = objects.Width()

	pos := rl.Vector2{rect.X + rect.Width - reservedX - totalWidth, posY}

	objects.Draw(pos)

	return totalWidth
}

func (r *Renderer) renderServerStatus(rect rl.Rectangle, offsetY, reservedX float32) float32 {
	fontSize := float32(r.app.Config.Fonts.Size)
	fontSpacing := float32(r.app.Config.Fonts.Spacing)
	posY := rect.Y + offsetY

	var avgOk float64
	var avgDuration time.Duration

	var count int
	records := r.app.State.ServerStatusRVB.ReadNew(r.app.Config.ServerStatus.HistorySize)
	for _, record := range records {
		if record.Err == nil {
			count++
			avgOk += 1
		}
		avgDuration += record.Status.RequestDuration
	}
	if len(records) > 0 {
		avgOk /= float64(len(records))
		avgDuration /= time.Duration(len(records))
	}

	var lastPing time.Time
	var ledColor = r.app.Config.Colors.GeneralDarkGray
	var pingMsg = "unknown"
	if len(records) > 0 {
		record := records[0]
		lastPing = record.Status.Timestamp.Add(record.Status.RequestDuration)

		if record.Err != nil {
			ledColor = r.app.Config.Colors.GeneralWarning
			pingMsg = "timeout/error"
		} else {
			pingMsg = record.Status.EndpointData.Status
			switch record.Status.EndpointData.Code {
			case elite_status.CodeDown, elite_status.CodeMaintenance:
				ledColor = r.app.Config.Colors.GeneralCritical
			case elite_status.CodeGood:
				ledColor = r.app.Config.Colors.GeneralOK
			default:
				ledColor = r.app.Config.Colors.GeneralCritical
			}
		}
	}

	var brighten float64
	if time.Since(lastPing).Seconds() < 1.0 {
		brighten = 1 - float64(time.Since(lastPing).Seconds())
	}

	rightPad := fontSize / 2
	ledPosX := rect.X + rect.Width - fontSize - rightPad - reservedX
	ledRect := rl.Rectangle{ledPosX, posY, fontSize, fontSize}
	rl.DrawTexturePro(
		r.app.State.Textures.LED.Texture,
		rl.Rectangle{0, 0, float32(r.app.State.Textures.LED.Texture.Width), float32(r.app.State.Textures.LED.Texture.Height)},
		ledRect,
		rl.Vector2{},
		0,
		ledColor.Lighten(0.2+brighten*0.5).AsRGBA(),
	)
	if !r.s.OptionViewActive && rl.CheckCollisionPointRec(r.s.MousePosition, ledRect) {
		r.app.State.ShowPingHover = true
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			r.app.State.ForceShowPing = !r.app.State.ForceShowPing
		}
	} else {
		r.app.State.ShowPingHover = false
	}

	if !r.app.Config.StatusBar.ServerStatusText {
		return ledRect.Width + rightPad
	}

	statusText := fmt.Sprintf("ServerStatus: %s", pingMsg)

	var objects = RenderObjects{
		{" [", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, fontSize, fontSpacing},
		{statusText, ledColor, &r.app.State.Fonts.Main, fontSize, fontSpacing},
		{"] ", r.app.Config.Colors.GeneralDarkGray, &r.app.State.Fonts.Main, fontSize, fontSpacing},
	}

	var totalWidth = objects.Width()

	objects.Draw(rl.Vector2{ledPosX - totalWidth, posY})

	return ledRect.Width + rightPad + totalWidth
}

func (r *Renderer) renderServerStatusHistogram(reservedForStatusbar float32) {
	pings := r.app.State.ServerStatusTimestampsRVB.ReadNew(r.app.Config.ServerStatus.HistorySize)
	statuses := r.app.State.ServerStatusRVB.ReadNew(r.app.Config.ServerStatus.HistorySize)

	var separator float32 = 10.0
	var border float32 = 2.0
	barWidthWithSpace := float32(r.s.CanvasSizeX) / float32(r.app.Config.ServerStatus.HistorySize-1)
	barWidth := barWidthWithSpace - border*2

	fontSize := barWidth
	if fontSize > float32(r.app.Config.Fonts.Size) {
		fontSize = float32(r.app.Config.Fonts.Size)
	}

	var counter int
	for _, pingTs := range pings {
		var found state.ServerStatus
		var foundOk bool

		for _, status := range statuses {
			if status.Timestamp == pingTs {
				found = status
				foundOk = true
				break
			}
		}
		if time.Now().Before(pingTs) {
			continue
		}
		counter++

		testtxt := fmt.Sprintf(" Maintenance - %s - 12.345 s ", time.Now().Format(r.app.Config.Viewports.Appearance.TimestampFormat))
		minHeight := rl.MeasureTextEx(r.app.State.Fonts.MainMono, testtxt, fontSize, float32(r.app.Config.Fonts.SpacingMono)).X

		var brighten float64
		var v float64

		if foundOk {
			v = found.Status.RequestDuration.Seconds() / r.app.Config.ServerStatus.Timeout.Seconds()

			brighten = 1 - float64(time.Since(found.Status.Timestamp.Add(found.Status.RequestDuration)).Seconds())
			if brighten < 0 {
				brighten = 0
			}
		} else {
			brighten = 0

			v = time.Since(pingTs).Seconds() / r.app.Config.ServerStatus.Timeout.Seconds()
		}

		availableHeight := float32(r.s.CanvasSizeY) - reservedForStatusbar
		height := (availableHeight - minHeight) * float32(utils.EaseOutExpo(utils.EaseOutExpo(v)))

		var txt string
		var ledColor my_noire.Color

		switch {
		case !foundOk:
			ledColor = r.app.Config.Colors.GeneralDarkGray
			txt = " waiting for response"
		case found.Err != nil:
			ledColor = r.app.Config.Colors.GeneralCritical
			txt = fmt.Sprintf(
				" err/timeout - %s - %s",
				found.Timestamp.Format(r.app.Config.Viewports.Appearance.TimestampFormat),
				fmtDuration(found.Status.RequestDuration),
			)
		default:
			switch found.Status.EndpointData.Code {
			case elite_status.CodeDown:
				ledColor = r.app.Config.Colors.GeneralCritical
			case elite_status.CodeMaintenance:
				ledColor = r.app.Config.Colors.GeneralWarning
			case elite_status.CodeGood:
				ledColor = r.app.Config.Colors.GeneralOK
			default:
				ledColor = r.app.Config.Colors.GeneralCritical
			}
			txt = fmt.Sprintf(
				" %11s - %s - %s",
				found.Status.EndpointData.Status,
				found.Timestamp.Format(r.app.Config.Viewports.Appearance.TimestampFormat),
				fmtDuration(found.Status.RequestDuration),
			)
		}

		anchorX := float32(r.s.CanvasSizeX) - barWidthWithSpace*(float32(counter))

		rl.DrawRectangleLinesEx( // background outline for better readability
			rl.Rectangle{
				anchorX - border,
				availableHeight - minHeight - height - separator - border,
				barWidth + border*2,
				height + separator + minHeight + border*2,
			},
			border,
			r.app.Config.Colors.Background.AsRGBA(),
		)

		rl.DrawRectangleRec( // minimum bar height reserved for text, darkened
			rl.Rectangle{
				anchorX,
				availableHeight - minHeight,
				barWidth,
				minHeight,
			},
			ledColor.Darken(0.1).Lighten(brighten).AsRGBA(),
		)
		rl.DrawRectangleRec( // drawing visually distinct separator with background color
			rl.Rectangle{
				anchorX,
				availableHeight - minHeight - separator,
				barWidth,
				separator,
			},
			r.app.Config.Colors.Background.AsRGBA(),
		)
		rl.DrawRectangleRec( // main bar representing ping time duration
			rl.Rectangle{
				anchorX,
				availableHeight - minHeight - separator - height,
				barWidth,
				height,
			},
			ledColor.Lighten(brighten).AsRGBA(),
		)
		rl.DrawTextPro( // text overlay with background color
			r.app.State.Fonts.MainMono,
			txt,
			rl.Vector2{
				anchorX,
				float32(r.s.CanvasSizeY) - reservedForStatusbar,
			},
			rl.Vector2{},
			270,
			fontSize,
			float32(r.app.Config.Fonts.SpacingMono),
			r.app.Config.Colors.Background.AsRGBA(),
		)
	}
}
