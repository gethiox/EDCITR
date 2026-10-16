package app

import (
	"fmt"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/app/journal"
	"go.uber.org/zap"
)

func (r *Renderer) renderViewportBorders(totalColumnWidth float32, statusBarHeight float32) {
	for mainColumn := 1; mainColumn < r.app.Config.Viewports.Count; mainColumn++ {
		rl.DrawRectangle(
			int32(totalColumnWidth)*int32(mainColumn)+int32(r.app.Config.Viewports.BorderSize*(mainColumn-1)+r.app.Config.Viewports.Appearance.LeftPad*mainColumn+r.app.Config.Viewports.Appearance.RightPad*mainColumn),
			0,
			int32(r.app.Config.Viewports.BorderSize),
			int32(r.s.CanvasSizeY)-int32(statusBarHeight),
			r.app.Config.Colors.ViewportBorder.AsRGBA(),
		)
	}
}

func (r *Renderer) renderViewports(totalColumnWidth float32, statusBarHeight float32) {
	var refreshAvailable = r.app.State.ProcessingLiveEvents.Load() && r.app.State.Inara.CMDRNamesRefreshAvailable.CompareAndSwap(true, false)

	var vars SomeVariables

	vars.ROWidth = totalColumnWidth
	vars.ROUserScroll = r.getMouseWheelV()
	vars.ROHeight = float32(r.s.CanvasSizeY) - statusBarHeight
	vars.ROstopY = 0

	toRenderCountEstimated := int(vars.ROHeight / (float32(r.app.Config.Fonts.Size) + float32(r.app.Config.Viewports.Appearance.MessageSpacing)))
	if toRenderCountEstimated < 0 {
		Logger.Warn("estimation negative", zap.Int("value", toRenderCountEstimated), zap.Float32("value", vars.ROHeight), zap.Int("value", r.app.Config.Fonts.Size), zap.Int("value", r.app.Config.Viewports.Appearance.MessageSpacing))
		toRenderCountEstimated = 0
	}

	for ViewportID := 0; ViewportID < r.app.Config.Viewports.Count; ViewportID++ {
		vars.ROstartX = float32(int(totalColumnWidth*float32(ViewportID)) + r.app.Config.Viewports.BorderSize*ViewportID + r.app.Config.Viewports.Appearance.LeftPad*(ViewportID) + r.app.Config.Viewports.Appearance.RightPad*(ViewportID))
		vars.ROstopX = vars.ROstartX + totalColumnWidth
		vars.ROstartY = float32(r.s.CanvasSizeY) - statusBarHeight // TODO: why being modified
		vars.RWoffsetYcum = 0
		vars.RWProfilePictureSize = 0

		// render Viewport content
		minimumMessagessToSee := 4
		allItems := r.app.State.ViewportBuffers[ViewportID].Buffer.GetCurrentSize()

		var viewChanged bool

		rectCheck := rl.Rectangle{vars.ROstartX, vars.ROstopY, vars.ROWidth, vars.ROHeight}
		if !r.s.OptionViewActive && r.app.State.ProcessingLiveEvents.Load() && rl.CheckCollisionPointRec(r.s.MousePosition, rectCheck) {
			scrollPos := r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollPos

			if rl.IsMouseButtonPressed(rl.MouseButtonMiddle) {
				if scrollPos != 0 {
					r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollCheckpoint = r.app.State.ViewportBuffers[ViewportID].Buffer.GetCheckpoint()
					r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollPos = 0
					viewChanged = true
				}
				goto yeetOut
			}

			scrollPosNew := min(max(scrollPos+int(vars.ROUserScroll.Y), 0), allItems-minimumMessagessToSee)
			r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollPos = scrollPosNew

			if scrollPos != scrollPosNew {
				viewChanged = true
			}

			if scrollPosNew == 0 {
				r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollCheckpoint = r.app.State.ViewportBuffers[ViewportID].Buffer.GetCheckpoint()
				goto yeetOut
			}

			// brighten terminator
			currentValue := r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollBrightenByUser
			newValue := currentValue + max(float64(vars.ROUserScroll.Y), 0)
			if newValue > 16 {
				additional := 16 - newValue
				r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollBrightenByUser = newValue + additional/1.5
			} else {
				r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollBrightenByUser = newValue
			}
		}
	yeetOut:

		var (
			scrollPos = r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollPos
			cp        = r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollCheckpoint
		)

		if scrollPos > 0 {
			newItems := r.app.State.ViewportBuffers[ViewportID].Buffer.NewItemsSince(cp)
			vars.ROstartY -= r.renderFollowingDisabled(vars, newItems, allItems)
		}

		vars.ROstartY -= float32(r.app.Config.Viewports.Appearance.MessageSpacing)

		var toRender []journal.MessageEvent
		if scrollPos > 0 {
			toRender, _ = r.app.State.ViewportBuffers[ViewportID].Buffer.ReadNewFromCheckpoint(cp, r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollPos, toRenderCountEstimated)
		} else {
			toRender = r.app.State.ViewportBuffers[ViewportID].Buffer.ReadNew(toRenderCountEstimated)
		}

		vars.RWProfilePictureSize = float32(r.app.Config.Fonts.Size*2) * r.app.Config.GenericImages.ImageSize

		for _, messageUntyped := range toRender {
			if vars.RWoffsetYcum > vars.ROHeight {
				break
			}

			switch message := messageUntyped.(type) {
			case *journal.ReceivedCMDRMessage:
				r.renderReceived(message, &vars)
				if refreshAvailable || (viewChanged && r.app.State.Inara.CMDRNamesStreamActive.Load()) {
					if message.To != "" {
						r.app.State.Inara.CMDRNamesStream <- message.To
					}
					if message.From != "" {
						r.app.State.Inara.CMDRNamesStream <- message.From
					}
				}
			case *journal.SentMessage:
				r.renderSent(message, &vars)
				if refreshAvailable || (viewChanged && r.app.State.Inara.CMDRNamesStreamActive.Load()) {
					if message.To != "" {
						r.app.State.Inara.CMDRNamesStream <- message.To
					}
					if message.From != "" {
						r.app.State.Inara.CMDRNamesStream <- message.From
					}
				}
			case *journal.NotificationMessage:
				r.renderReceivedNotification(message, &vars)
			case *journal.ReceivedNPCMessage:
				r.renderReceived(message, &vars)
			case *journal.GenericEvent:
				r.renderGenericJournalMessage(message, &vars)
			case *journal.FriendInviteMessage:
				r.renderInvite(message, &vars)
				if refreshAvailable || (viewChanged && r.app.State.Inara.CMDRNamesStreamActive.Load()) {
					if message.From != "" {
						r.app.State.Inara.CMDRNamesStream <- message.From
					}
				}
			case *journal.WingInviteMessage:
				r.renderInvite(message, &vars)
				if refreshAvailable || (viewChanged && r.app.State.Inara.CMDRNamesStreamActive.Load()) {
					if message.From != "" {
						r.app.State.Inara.CMDRNamesStream <- message.From
					}
				}
			default:
				Logger.Error("Unknown Message", zap.Any("message", message))
			}
		}

		if len(toRender) < toRenderCountEstimated {
			r.renderTerminator(&vars, r.app.State.ViewportBuffers[ViewportID].Scroll.ScrollBrightenByUser)
		}

		channels := r.app.Config.Viewports.ViewportChannelBitmask[ViewportID]
		r.renderViewportHeader(vars, channels.ExtractList())
	}
}

func (r *Renderer) getHeaderHeight() float32 {
	return float32(r.app.Config.Fonts.Size) * 1.5
}

func (r *Renderer) renderViewportHeader(vars SomeVariables, channelList []journal.MessageChannel) float32 {
	headerHeight := r.getHeaderHeight()
	rect := rl.Rectangle{vars.ROstartX, vars.ROstopY, vars.ROWidth, headerHeight}
	rl.DrawRectangleGradientEx(
		rect,
		r.app.Config.Colors.Header1.AsRGBA(),
		r.app.Config.Colors.Header2.AsRGBA(),
		r.app.Config.Colors.Header2.AsRGBA(),
		r.app.Config.Colors.Header1.AsRGBA(),
	)

	var cumWidth float32
	for _, channel := range channelList {
		txt := journal.MessageChannelToString[channel]
		m := rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		cumWidth += m.X
	}
	if len(channelList) == 0 {
		txt := "UNDEFINED"
		m := rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		cumWidth += m.X
	}
	if len(channelList) > 1 {
		m := rl.MeasureTextEx(r.app.State.Fonts.Main, ", ", float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		cumWidth += m.X * float32(len(channelList)-1)
		m2 := rl.MeasureTextEx(r.app.State.Fonts.Main, "[", float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		cumWidth += m2.X * 2
	}

	mh := rl.MeasureTextEx(r.app.State.Fonts.Main, "[]", float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))

	offsetX := vars.ROstartX + vars.ROWidth/2 - cumWidth/2
	offsetY := headerHeight/2 - mh.Y/2

	m := rl.MeasureTextEx(r.app.State.Fonts.Main, "[", float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
	rl.DrawTextEx(r.app.State.Fonts.Main, "[", rl.Vector2{offsetX, offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), r.app.Config.Colors.HeaderDefault.AsRGBA())
	offsetX += m.X
	for i, channel := range channelList {
		txt := journal.MessageChannelToString[channel]
		m := rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		rl.DrawTextEx(r.app.State.Fonts.Main, txt, rl.Vector2{offsetX, offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), (*r.ChannelToColor[channel]).AsRGBA())
		offsetX += m.X
		if i < len(channelList)-1 {
			m := rl.MeasureTextEx(r.app.State.Fonts.Main, ", ", float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
			rl.DrawTextEx(r.app.State.Fonts.Main, ", ", rl.Vector2{offsetX, offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), r.app.Config.Colors.HeaderDefault.AsRGBA())
			offsetX += m.X
		}
	}
	if len(channelList) == 0 {
		txt := "NOT DEFINED"
		m := rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		rl.DrawTextEx(r.app.State.Fonts.Main, txt, rl.Vector2{offsetX, offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), r.app.Config.Colors.HighlightWord.AsRGBA())
		offsetX += m.X
	}
	m = rl.MeasureTextEx(r.app.State.Fonts.Main, "]", float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
	rl.DrawTextEx(r.app.State.Fonts.Main, "]", rl.Vector2{offsetX, offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), r.app.Config.Colors.HeaderDefault.AsRGBA())
	offsetX += m.X

	return headerHeight
}

func (r *Renderer) renderFollowingDisabled(vars SomeVariables, newItems, allItems int) float32 {
	headerHeight := float32(r.app.Config.Fonts.Size) * 1.5
	up := r.app.Config.Colors.FollowingDisabled
	down := r.app.Config.Colors.FollowingDisabled

	rect := rl.Rectangle{vars.ROstartX, vars.ROstartY - headerHeight, vars.ROWidth, headerHeight}
	rl.DrawRectangleGradientV(
		int32(rect.X),
		int32(rect.Y),
		int32(rect.Width),
		int32(rect.Height),
		up.AsRGBA(),
		down.AsRGBA(),
	)

	// shadows
	rl.DrawRectangleGradientV(
		int32(rect.X),
		int32(rect.Y),
		int32(rect.Width),
		int32(rect.Height/8),
		up.Lighten(0.1).AsRGBA(),
		down.AsRGBA(),
	)
	rl.DrawRectangleGradientV(
		int32(rect.X),
		int32(rect.Y+(rect.Height/8)*7),
		int32(rect.Width),
		int32(rect.Height/8),
		up.AsRGBA(),
		down.Darken(0.3).AsRGBA(),
	)

	strips := rl.Black
	strips.A = 64
	rl.DrawTextureRec(
		r.app.State.Textures.Terminator.Texture,
		rl.Rectangle{
			vars.ROstartX + float32(math.Mod(float64(time.Now().UnixMilli())/2000, 1))*60,
			0,
			float32(vars.ROstopX - vars.ROstartX),
			-headerHeight,
		},
		rl.Vector2{vars.ROstartX, vars.ROstartY - headerHeight},
		strips,
	)

	text2Color := r.app.Config.Colors.GeneralGray.AsRGBA()
	if newItems > 0 {
		text2Color = r.app.Config.Colors.GeneralOK.AsRGBA()
	}

	text0 := fmt.Sprintf("(%d entries) ", allItems)
	text1 := "[following disabled] "
	text2 := fmt.Sprintf("(%d new)", newItems)

	m0 := rl.MeasureTextEx(r.app.State.Fonts.Main, text0, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
	m1 := rl.MeasureTextEx(r.app.State.Fonts.Main, text1, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
	m2 := rl.MeasureTextEx(r.app.State.Fonts.Main, text2, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))

	var offsetY = headerHeight/2 - m1.Y/2

	switch {
	case vars.ROWidth > m0.X+m1.X+m2.X:
		var offsetX = vars.ROstartX + vars.ROWidth/2 - m0.X/2 - m1.X/2 - m2.X/2
		rl.DrawTextEx(r.app.State.Fonts.Main, text0, rl.Vector2{offsetX, vars.ROstartY - headerHeight + offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), r.app.Config.Colors.GeneralGray.AsRGBA())
		rl.DrawTextEx(r.app.State.Fonts.Main, text1, rl.Vector2{offsetX + m0.X, vars.ROstartY - headerHeight + offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), r.app.Config.Colors.GeneralWarning.AsRGBA())
		rl.DrawTextEx(r.app.State.Fonts.Main, text2, rl.Vector2{offsetX + m0.X + m1.X, vars.ROstartY - headerHeight + offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), text2Color)
	case vars.ROWidth > m0.X+m2.X:
		var offsetX = vars.ROstartX + vars.ROWidth/2 - m0.X/2 - m2.X/2
		rl.DrawTextEx(r.app.State.Fonts.Main, text0, rl.Vector2{offsetX, vars.ROstartY - headerHeight + offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), r.app.Config.Colors.GeneralGray.AsRGBA())
		rl.DrawTextEx(r.app.State.Fonts.Main, text2, rl.Vector2{offsetX + m0.X, vars.ROstartY - headerHeight + offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), text2Color)
	default:
		var offsetX = vars.ROstartX + vars.ROWidth/2 - m2.X/2
		rl.DrawTextEx(r.app.State.Fonts.Main, text2, rl.Vector2{offsetX, vars.ROstartY - headerHeight + offsetY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), text2Color)
	}
	return headerHeight
}
