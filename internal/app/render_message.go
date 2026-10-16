package app

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	journal_app "github.com/gethiox/EDCITR/internal/app/journal"
	"github.com/gethiox/EDCITR/internal/app/state"
	"github.com/gethiox/EDCITR/internal/config"
	"github.com/gethiox/EDCITR/internal/inara"
	events "github.com/gethiox/EDCITR/internal/inara/inapi/events/commander_profile"
	"github.com/gethiox/EDCITR/internal/noire"
	"github.com/gethiox/EDCITR/internal/parser/message"
	"github.com/gethiox/EDCITR/internal/utils"
	"go.uber.org/zap"
	"mvdan.cc/xurls/v2"
)

var urlFinder = xurls.Strict()

type SomeVariables struct {
	ROstartY, ROstartX float32
	ROstopY, ROstopX   float32
	ROWidth, ROHeight  float32
	ROUserScroll       rl.Vector2

	RWProfilePictureSize float32
	RWoffsetYcum         float32
}

type TextDrawCommand struct {
	Text          string
	Position      rl.Vector2
	IsWord        bool
	IsUrl         bool
	IsHighlighted bool
	Width         float32
}

type ProfileInfoDrawCommand struct {
	Text    string
	TextCol rl.Color
	TextPos rl.Vector2

	HasIcon        bool
	SuffixIcon     rl.Texture2D
	SuffixIconTint rl.Color
	SuffixIconPos  rl.Vector2
}

type ProfileRankDrawCommand struct {
	Name    string
	NamePos rl.Vector2

	Value    string
	ValuePos rl.Vector2

	RankIcon rl.Texture2D
	IconPos  rl.Vector2
}

type NotificationDrawCommand struct {
	Name    string
	NamePos rl.Vector2

	Value    string
	ValuePos rl.Vector2
}

func (r *Renderer) renderHoverNotification() {
	notification := r.s.LastExtraParametersHover
	pos, altPos := r.s.HoverPos, r.s.AlternativeHoverPos

	if len(notification) == 0 {
		return
	}

	notificationKeys := make([]string, 0, len(notification))
	for k := range notification {
		notificationKeys = append(notificationKeys, k)
	}
	sort.Strings(notificationKeys)

	pos.X += float32(r.app.Config.Fonts.Size)
	altPos.X -= float32(r.app.Config.Fonts.Size)

	bw := float32(3)

	mainColor := r.app.Config.Colors.HeaderDefault

	var (
		m1, m2 rl.Vector2
		posY   float32
	)

	minimumValueWidth := float32(0)
	minimumKeyWidth := float32(0)

	var notificationDraw = make([]NotificationDrawCommand, 0)
	// Generate rank list information
	for i, key := range notificationKeys {
		value := notification[key]
		key = key + ": "

		m1 = rl.MeasureTextEx(r.app.State.Fonts.Main, key, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		m2 = rl.MeasureTextEx(r.app.State.Fonts.Main, value, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		minimumKeyWidth = max(minimumKeyWidth, float32(r.app.Config.Fonts.Size)/2+m1.X)
		minimumValueWidth = max(minimumValueWidth, float32(r.app.Config.Fonts.Size)/2+m2.X)
		posY = float32(r.app.Config.Fonts.Size) * float32(i)

		notificationDraw = append(notificationDraw, NotificationDrawCommand{
			Name:     key,
			NamePos:  rl.Vector2{-m1.X, posY},
			Value:    value,
			ValuePos: rl.Vector2{0, posY},
		})
	}

	minimumKeyWidth += float32(r.app.Config.Fonts.Size)
	minimumValueWidth += float32(r.app.Config.Fonts.Size)
	totalWidth := minimumKeyWidth + minimumValueWidth

	contentBodyHeight := float32(0)
	contentBodyHeight += float32(r.app.Config.Fonts.Size) * float32(len(notificationDraw))
	contentBodyHeight += float32(r.app.Config.Fonts.Size) * 2

	var finalPos rl.Vector2
	if pos.X+totalWidth+bw > float32(r.s.CanvasSizeX) {
		finalPos = altPos
		finalPos.X -= totalWidth + bw
	} else {
		finalPos = pos
	}

	if finalPos.Y+contentBodyHeight > float32(r.s.CanvasSizeY)-bw {
		finalPos.Y = float32(r.s.CanvasSizeY) - bw - contentBodyHeight
	}

	center := finalPos.X + minimumKeyWidth

	backgroundRect := rl.Rectangle{
		finalPos.X,
		finalPos.Y,
		totalWidth,
		float32(0), // extended below
	}
	backgroundRect.Height += float32(r.app.Config.Fonts.Size) * float32(len(notificationDraw))
	backgroundRect.Height += float32(r.app.Config.Fonts.Size) * 2

	borderRect := backgroundRect
	borderRect.X -= bw
	borderRect.Y -= bw
	borderRect.Width += bw * 2
	borderRect.Height += bw * 2

	rl.DrawRectangleRec(backgroundRect, r.app.Config.Colors.Background.AsRGBA())
	rl.DrawRectangleLinesEx(borderRect, bw, mainColor.AsRGBA())

	anchorY := finalPos.Y + float32(r.app.Config.Fonts.Size)

	// Render rank list information
	for _, draw := range notificationDraw {
		rl.DrawTextEx(r.app.State.Fonts.Main, draw.Name, rl.Vector2{center + draw.NamePos.X, anchorY + draw.NamePos.Y}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), rl.White)
		rl.DrawTextEx(r.app.State.Fonts.Main, draw.Value, rl.Vector2{center + draw.ValuePos.X, anchorY + draw.ValuePos.Y}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), rl.White)
	}

	if len(notificationDraw) > 0 {
		anchorY += float32(r.app.Config.Fonts.Size)
		anchorY += float32(r.app.Config.Fonts.Size) * float32(len(notificationDraw))
	}
}

func (r *Renderer) renderHoverProfileInfo() {
	cmdrName := r.s.LastCmdrNameHover
	pos, altPos := r.s.HoverPos, r.s.AlternativeHoverPos

	info, textureOk := r.app.State.Inara.CMDRInfo.Get(cmdrName)
	if !textureOk {
		return
	}

	pos.X += float32(r.app.Config.Fonts.Size)
	altPos.X -= float32(r.app.Config.Fonts.Size)

	bw := float32(3)
	picSize := r.app.Config.GenericImages.HoverSize
	iconSpacing := float32(5)

	mainColor := r.app.Config.Colors.HeaderCMDR
	switch {
	case cmdrName == r.app.State.CurrentThings.CMDR.Load():
		mainColor = r.app.Config.Colors.HeaderSent
	default:
		mainColor = r.CMDRColorFriendEnemy(cmdrName)
	}

	var (
		txt       string
		row       int
		hasIcon   bool
		m, m1, m2 rl.Vector2
		txtCol    my_noire.Color
		posY      float32
	)

	minimumWidth := picSize + float32(r.app.Config.Fonts.Size)*2

	var profileInfos = make([]ProfileInfoDrawCommand, 0)
	// Generate basic profile information
	{
		row = 0
		txt = fmt.Sprintf("Inara username: %s (ID: %d)", info.ProfileInfo.UserName, info.ProfileInfo.UserID)
		txtCol = r.app.Config.Colors.GeneralNormal
		m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		minimumWidth = max(minimumWidth, m.X)
		profileInfos = append(profileInfos, ProfileInfoDrawCommand{
			Text:    txt,
			TextCol: txtCol.AsRGBA(),
			TextPos: rl.Vector2{-m.X / 2, float32(r.app.Config.Fonts.Size) * float32(row)},
		})

		row = 1
		txt = "No CMDR Name"
		txtCol = r.app.Config.Colors.GeneralDarkGray
		if info.ProfileInfo.CommanderName != "" {
			txt = fmt.Sprintf("CMDR %s", info.ProfileInfo.CommanderName)
			txtCol = mainColor
		}
		m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		minimumWidth = max(minimumWidth, m.X)
		profileInfos = append(profileInfos, ProfileInfoDrawCommand{
			Text:    txt,
			TextCol: txtCol.AsRGBA(),
			TextPos: rl.Vector2{-m.X / 2, float32(r.app.Config.Fonts.Size) * float32(row)},
		})

		row = 2
		txt = "Squadron: None"
		txtCol = r.app.Config.Colors.GeneralDarkGray
		if info.ProfileInfo.CommanderSquadron.SquadronName != "" {
			txt = fmt.Sprintf("Squadron: %s (%s)", info.ProfileInfo.CommanderSquadron.SquadronName, info.ProfileInfo.CommanderSquadron.SquadronMemberRank)
			txtCol = r.app.Config.Colors.GeneralNormal
		}
		m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		minimumWidth = max(minimumWidth, m.X)
		profileInfos = append(profileInfos, ProfileInfoDrawCommand{
			Text:    txt,
			TextCol: txtCol.AsRGBA(),
			TextPos: rl.Vector2{-m.X / 2, float32(r.app.Config.Fonts.Size) * float32(row)},
		})

		row = 3
		txt = "Role: None"
		txtCol = r.app.Config.Colors.GeneralDarkGray
		if info.ProfileInfo.PreferredGameRole != "" {
			txt = fmt.Sprintf("Role: %s", info.ProfileInfo.PreferredGameRole)
			txtCol = r.app.Config.Colors.GeneralNormal
		}
		m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		minimumWidth = max(minimumWidth, m.X)
		profileInfos = append(profileInfos, ProfileInfoDrawCommand{
			Text:    txt,
			TextCol: txtCol.AsRGBA(),
			TextPos: rl.Vector2{-m.X / 2, float32(r.app.Config.Fonts.Size) * float32(row)},
		})

		row = 4
		txt = "Ship: None"
		txtCol = r.app.Config.Colors.GeneralDarkGray
		if info.ProfileInfo.CommanderMainShip.ShipType != "" {
			txt = fmt.Sprintf("Ship: %s", info.ProfileInfo.CommanderMainShip)
			txtCol = r.app.Config.Colors.GeneralNormal
		}
		m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		minimumWidth = max(minimumWidth, m.X)
		profileInfos = append(profileInfos, ProfileInfoDrawCommand{
			Text:    txt,
			TextCol: txtCol.AsRGBA(),
			TextPos: rl.Vector2{-m.X / 2, float32(r.app.Config.Fonts.Size) * float32(row)},
		})

		row = 5
		txt = "Power: None"
		txtCol = r.app.Config.Colors.GeneralDarkGray
		hasIcon = false
		if info.ProfileInfo.PreferredPowerName != "" {
			txt = fmt.Sprintf("Power: %s", info.ProfileInfo.PreferredPowerName)
			txtCol = my_noire.Color{inara.Powerplay2Colors[info.ProfileInfo.PreferredPowerName]}
			hasIcon = true
		}
		m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		minimumWidth = max(minimumWidth, m.X)
		profileInfos = append(profileInfos, ProfileInfoDrawCommand{
			Text:           txt,
			TextCol:        txtCol.AsRGBA(),
			TextPos:        rl.Vector2{-m.X / 2, float32(r.app.Config.Fonts.Size) * float32(row)},
			HasIcon:        hasIcon,
			SuffixIcon:     r.app.State.Inara.Powers[inara.PowerToID[info.ProfileInfo.PreferredPowerName]],
			SuffixIconTint: rl.White,
			SuffixIconPos:  rl.Vector2{+m.X/2 + float32(r.app.Config.Fonts.Size)/4, float32(r.app.Config.Fonts.Size) * float32(row)},
		})

		row = 6
		txt = "Power: None"
		txtCol = r.app.Config.Colors.GeneralDarkGray
		hasIcon = false
		if info.ProfileInfo.PreferredAllegianceName != "" {
			txt = fmt.Sprintf("Allegiance: %s", info.ProfileInfo.PreferredAllegianceName)
			txtCol = my_noire.Color{inara.SuperpowerColors[info.ProfileInfo.PreferredAllegianceName]}
			hasIcon = true
		}
		m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		minimumWidth = max(minimumWidth, m.X)
		profileInfos = append(profileInfos, ProfileInfoDrawCommand{
			Text:           txt,
			TextCol:        txtCol.AsRGBA(),
			TextPos:        rl.Vector2{-m.X / 2, float32(r.app.Config.Fonts.Size) * float32(row)},
			HasIcon:        hasIcon,
			SuffixIcon:     r.app.State.Inara.Superpowers[inara.SuperpowerToID[info.ProfileInfo.PreferredAllegianceName]],
			SuffixIconTint: txtCol.AsRGBA(),
			SuffixIconPos:  rl.Vector2{+m.X/2 + float32(r.app.Config.Fonts.Size)/4, float32(r.app.Config.Fonts.Size) * float32(row)},
		})
	}

	var profileRanks = make([]ProfileRankDrawCommand, 0)
	// Generate rank list information
	for i, rank := range info.ProfileInfo.CommanderRanksPilot {
		rankName, rankValue := events.ConvertToRankNameAndValue(rank)
		nameID, valueID := events.ConvertToInt(rank)

		m1 = rl.MeasureTextEx(r.app.State.Fonts.Main, rankName, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		m2 = rl.MeasureTextEx(r.app.State.Fonts.Main, rankValue, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
		minimumWidth = max(minimumWidth, float32(r.app.Config.Fonts.Size)/2+iconSpacing+m1.X)
		minimumWidth = max(minimumWidth, float32(r.app.Config.Fonts.Size)/2+iconSpacing+m2.X)
		posY = float32(r.app.Config.Fonts.Size) * float32(i)

		profileRanks = append(profileRanks, ProfileRankDrawCommand{
			Name:     rankName,
			NamePos:  rl.Vector2{-float32(r.app.Config.Fonts.Size)/2 - iconSpacing - m1.X, posY},
			Value:    rankValue,
			ValuePos: rl.Vector2{+float32(r.app.Config.Fonts.Size)/2 + iconSpacing, posY},
			RankIcon: r.app.State.Inara.Ranks[nameID][valueID],
			IconPos:  rl.Vector2{-float32(r.app.Config.Fonts.Size) / 2, posY},
		})
	}

	updateInfo := fmt.Sprintf("[profile updated %s]", utils.Ago.Format(info.ProfileInfoTime))
	mUpdateInfo := rl.MeasureTextEx(r.app.State.Fonts.Main, updateInfo, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
	minimumWidth = max(minimumWidth, mUpdateInfo.X)

	minimumWidth += float32(r.app.Config.Fonts.Size) * 2

	contentHeadHeight := float32(r.app.Config.Fonts.Size) + picSize
	contentBodyHeight := float32(r.app.Config.Fonts.Size) * float32(len(profileInfos))
	if len(profileRanks) > 0 {
		contentBodyHeight += float32(r.app.Config.Fonts.Size) // section separator
		contentBodyHeight += float32(r.app.Config.Fonts.Size) * float32(len(profileRanks))
	}
	contentBodyHeight += float32(r.app.Config.Fonts.Size) // space for update times
	contentBodyHeight += float32(r.app.Config.Fonts.Size) // space for update times

	var finalPos rl.Vector2
	if pos.X+minimumWidth+bw > float32(r.s.CanvasSizeX) {
		finalPos = altPos
		finalPos.X -= minimumWidth + bw
	} else {
		finalPos = pos
	}

	if finalPos.Y+contentBodyHeight > float32(r.s.CanvasSizeY)-bw {
		finalPos.Y = float32(r.s.CanvasSizeY) - bw - contentBodyHeight
	}
	if finalPos.Y-contentHeadHeight < 0+bw {
		finalPos.Y = bw + contentHeadHeight
	}

	center := finalPos.X + minimumWidth/2

	backgroundRect := rl.Rectangle{
		finalPos.X,
		finalPos.Y - picSize - float32(r.app.Config.Fonts.Size),
		minimumWidth,
		float32(r.app.Config.Fonts.Size) + picSize, // extended below
	}
	backgroundRect.Height += float32(r.app.Config.Fonts.Size) * float32(len(profileInfos))
	if len(profileRanks) > 0 {
		backgroundRect.Height += float32(r.app.Config.Fonts.Size) // section separator
		backgroundRect.Height += float32(r.app.Config.Fonts.Size) * float32(len(profileRanks))
	}
	backgroundRect.Height += float32(r.app.Config.Fonts.Size) // space for update times
	backgroundRect.Height += float32(r.app.Config.Fonts.Size) // space for update times

	borderRect := backgroundRect
	borderRect.X -= bw
	borderRect.Y -= bw
	borderRect.Width += bw * 2
	borderRect.Height += bw * 2

	rl.DrawRectangleRec(backgroundRect, r.app.Config.Colors.Background.AsRGBA())
	rl.DrawRectangleLinesEx(borderRect, bw, mainColor.AsRGBA())

	anchorY := finalPos.Y
	// Render basic profile information
	for _, inf := range profileInfos {
		rl.DrawTextEx(r.app.State.Fonts.Main, inf.Text, rl.Vector2{center + inf.TextPos.X, anchorY + inf.TextPos.Y}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), inf.TextCol)
		if inf.HasIcon {
			rl.DrawTexturePro(
				inf.SuffixIcon,
				rl.Rectangle{0, 0, float32(inf.SuffixIcon.Width), float32(inf.SuffixIcon.Height)},
				rl.Rectangle{center + inf.SuffixIconPos.X, anchorY + inf.SuffixIconPos.Y, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Size)},
				rl.Vector2{},
				0,
				inf.SuffixIconTint,
			)
		}
	}

	anchorY += float32(r.app.Config.Fonts.Size) * (float32(len(profileInfos)))
	anchorY += float32(r.app.Config.Fonts.Size)
	// Render rank list information
	for _, rank := range profileRanks {
		rl.DrawTextEx(r.app.State.Fonts.Main, rank.Name, rl.Vector2{center + rank.NamePos.X, anchorY + rank.NamePos.Y}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), rl.White)
		rl.DrawTextEx(r.app.State.Fonts.Main, rank.Value, rl.Vector2{center + rank.ValuePos.X, anchorY + rank.ValuePos.Y}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), rl.White)

		rl.DrawTexturePro(
			rank.RankIcon,
			rl.Rectangle{0, 0, float32(rank.RankIcon.Width), float32(rank.RankIcon.Height)},
			rl.Rectangle{center + rank.IconPos.X, anchorY + rank.IconPos.Y, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Size)},
			rl.Vector2{},
			0,
			rl.White,
		)
	}

	if len(profileRanks) > 0 {
		anchorY += float32(r.app.Config.Fonts.Size)
		anchorY += float32(r.app.Config.Fonts.Size) * float32(len(profileRanks))
	}
	rl.DrawTextEx(r.app.State.Fonts.Main, updateInfo, rl.Vector2{center - mUpdateInfo.X/2, anchorY}, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), r.app.Config.Colors.GeneralDarkGray.AsRGBA())

	texture, textureOk := r.app.State.Inara.CMDRTextures.Get(cmdrName)

	if textureOk {
		anchorX := center - picSize/2
		anchorY := finalPos.Y - picSize

		sizeX, sizeY := howToResize(float32(texture.Width), float32(texture.Height))
		ofstX := (picSize - (picSize * sizeX)) / 2
		ofstY := (picSize - (picSize * sizeY)) / 2

		dst := rl.Rectangle{
			anchorX + ofstX,
			anchorY + ofstY,
			picSize * sizeX,
			picSize * sizeY,
		}
		pfp := rl.Rectangle{
			anchorX + ofstX + bw,
			anchorY + ofstY + bw,
			picSize*sizeX - bw*2,
			picSize*sizeY - bw*2,
		}

		rl.DrawTexturePro(
			texture,
			rl.Rectangle{0, 0, float32(texture.Width), float32(texture.Height)},
			pfp,
			rl.Vector2{},
			0,
			rl.White,
		)
		rl.DrawRectangleLinesEx(dst, bw, mainColor.AsRGBA())

		txt := fmt.Sprintf("%dx%d px", texture.Width, texture.Height)
		if texture.Width != int32(picSize*sizeX) && texture.Height != int32(picSize*sizeY) {
			txt += fmt.Sprintf(" (%dx%d px)", int(picSize*sizeX), int(picSize*sizeY))
		}
		m := rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))

		rl.DrawTextEx(
			r.app.State.Fonts.Main,
			txt,
			rl.Vector2{dst.X + picSize/2 - m.X/2, dst.Y - float32(r.app.Config.Fonts.Size)},
			float32(r.app.Config.Fonts.Size),
			float32(r.app.Config.Fonts.Spacing),
			r.app.Config.Colors.GeneralDarkGray.AsRGBA(),
		)
	} else {

		for _, thing := range []struct {
			texture rl.Texture2D
			color   my_noire.Color
		}{
			{r.app.State.Inara.AvatarPlaceholderCMDRShoulders, GenerateSuitColor(cmdrName, 0)},
			{r.app.State.Inara.AvatarPlaceholderCMDRHead, GenerateSkinTone(cmdrName, 0)},
			{r.app.State.Inara.AvatarPlaceholderCMDRQuestion, r.app.Config.Colors.Background},
		} {
			texture = thing.texture

			anchorX := center - picSize/2
			anchorY := finalPos.Y - picSize

			sizeX, sizeY := howToResize(float32(texture.Width), float32(texture.Height))
			ofstX := (picSize - (picSize * sizeX)) / 2
			ofstY := (picSize - (picSize * sizeY)) / 2

			pfp := rl.Rectangle{
				anchorX + ofstX + bw,
				anchorY + ofstY + bw,
				picSize*sizeX - bw*2,
				picSize*sizeY - bw*2,
			}

			rl.DrawTexturePro(
				texture,
				rl.Rectangle{0, 0, float32(texture.Width), float32(texture.Height)},
				pfp,
				rl.Vector2{},
				0,
				thing.color.AsRGBA(),
			)
		}
		anchorX := center - picSize/2
		anchorY := finalPos.Y - picSize

		sizeX, sizeY := howToResize(float32(texture.Width), float32(texture.Height))
		ofstX := (picSize - (picSize * sizeX)) / 2
		ofstY := (picSize - (picSize * sizeY)) / 2

		dst := rl.Rectangle{
			anchorX + ofstX,
			anchorY + ofstY,
			picSize * sizeX,
			picSize * sizeY,
		}
		rl.DrawRectangleLinesEx(dst, bw, mainColor.AsRGBA())

		// loading overlay
		texState, _ := r.app.State.Inara.CMDRTexturesState.Get(cmdrName)
		switch texState {
		case state.StateDownloading, state.StateQueued:
			anchorX := center - picSize/2
			anchorY := finalPos.Y - picSize

			sizeX, sizeY := howToResize(float32(EdLoaderFrameWidth), float32(EdLoaderFrameHeight))
			sizeX, sizeY = sizeX*EDLoaderPFPSize, sizeY*EDLoaderPFPSize
			ofstX := (picSize - (picSize * sizeX)) / 2
			ofstY := (picSize - (picSize * sizeY)) / 2

			frame := getEdLoaderFrame()
			rect := rl.Rectangle{float32(EdLoaderFrameWidth) * float32(frame), 0, float32(EdLoaderFrameWidth), float32(EdLoaderFrameHeight)}
			pfp := rl.Rectangle{
				anchorX + ofstX + bw,
				anchorY + ofstY + bw,
				picSize*sizeX - bw*2,
				picSize*sizeY - bw*2,
			}

			rl.DrawTexturePro(
				r.app.State.Textures.EliteLoader,
				rect,
				pfp,
				rl.Vector2{},
				0,
				rl.White,
			)
		default:
			break
		}
	}
}

func (r *Renderer) renderInvitationHeader(msgUntyped journal_app.MessageEvent, vars *SomeVariables) float32 {
	switch msg := msgUntyped.(type) {
	case *journal_app.FriendInviteMessage:
		return r.renderHeader(msgUntyped, vars)
	case *journal_app.WingInviteMessage:
		return r.renderHeader(msgUntyped, vars)
	case *journal_app.SquadronInviteMessage:
		return r.renderSquadronInviteHeader(msg, vars)
	default:
		Logger.Error("Unrecognized journal event", zap.Any("msg", msgUntyped))
		return 0
	}
}

func (r *Renderer) CMDRColorFriendEnemy(cmdrName string) my_noire.Color {
	if r.app.State.FriendList.Has(cmdrName) {
		return r.app.Config.Colors.HeaderFriend
	}
	if r.app.State.EnemyList.Has(cmdrName) {
		return r.app.Config.Colors.HeaderEnemy
	}
	return r.app.Config.Colors.HeaderCMDR
}

func (r *Renderer) renderSquadronInviteHeader(msg *journal_app.SquadronInviteMessage, vars *SomeVariables) float32 {
	var (
		currentSessionCounter = r.app.State.SessionCounter.Load()
		sessionCounter        = msg.Session()
		text                  = msg.Squadron
	)

	if len(text) > 24 {
		text = text[:19] + "(...)"
	}

	headerColor := r.app.Config.Colors.HeaderDefault
	color := r.app.Config.Colors.HeaderDefault
	color = dimColorForOutdatedSession(color, sessionCounter, currentSessionCounter)

	xOffset := float32(0) + float32(r.app.Config.Viewports.Appearance.MarkerWidth)
	if r.app.Config.GenericImages.Enabled {
		xOffset += vars.RWProfilePictureSize
	}

	var fontSize = float32(r.app.Config.Fonts.Size)
	var fontSpacing = float32(r.app.Config.Fonts.Spacing)

	txt := " ["
	m := rl.MeasureTextEx(r.app.State.Fonts.Main, txt, fontSize, fontSpacing)
	rl.DrawTextEx(r.app.State.Fonts.Main, txt, rl.Vector2{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize}, fontSize, fontSpacing, headerColor.AsRGBA())
	xOffset += m.X

	srcM := rl.MeasureTextEx(r.app.State.Fonts.Main, text, fontSize, fontSpacing)

	if !r.s.OptionViewActive && rl.CheckCollisionPointRec(r.s.MousePosition, rl.Rectangle{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize, srcM.X, fontSize}) {
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) || rl.IsMouseButtonDown(rl.MouseButtonRight) {
			color = color.Darken(ClickDarken)
		} else {
			color = color.Lighten(HoverBrighten)
		}
		if rl.IsMouseButtonReleased(rl.MouseButtonRight) {
			rl.SetClipboardText(msg.Squadron)
			r.app.State.BubbleMessage.SetMessage("name copied!", r.s.MousePosition)
		}
	}

	txt = "] "
	m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, fontSize, fontSpacing)
	rl.DrawTextEx(r.app.State.Fonts.Main, txt, rl.Vector2{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize}, fontSize, fontSpacing, headerColor.AsRGBA())
	xOffset += m.X

	r.renderTime(msg.Timestamp, vars, sessionCounter)

	vars.RWoffsetYcum += fontSize
	return fontSize
}

func (r *Renderer) renderHeader(msgUntyped journal_app.MessageEvent, vars *SomeVariables) float32 {
	var (
		txt            = ""
		timestamp      = msgUntyped.TimestampTime()
		sessionCounter = msgUntyped.Session()
		channel        = msgUntyped.MessageChannel()

		from, to               string
		fromDisplay, toDisplay string
		fromNPC                bool
		inara1, inara2         state.DataState
		srcColor, dstColor     my_noire.Color
	)

	switch msg := msgUntyped.(type) {
	case *journal_app.SentMessage:
		from, to = msg.From, msg.To
		fromDisplay, toDisplay = msg.From, msg.To

		inara1, _ = r.app.State.Inara.CMDRInfoState.Get(msg.From)
		inara2, _ = r.app.State.Inara.CMDRInfoState.Get(msg.To)

		srcColor = r.app.Config.Colors.HeaderSent
		dstColor = r.CMDRColorFriendEnemy(msg.To)
	case *journal_app.ReceivedCMDRMessage:
		from, to = msg.From, msg.To
		fromDisplay, toDisplay = msg.From, msg.To

		inara1, _ = r.app.State.Inara.CMDRInfoState.Get(msg.From)
		inara2, _ = r.app.State.Inara.CMDRInfoState.Get(msg.To)

		srcColor = r.CMDRColorFriendEnemy(msg.From)
		dstColor = r.app.Config.Colors.HeaderSent
	case *journal_app.ReceivedNPCMessage:
		from, to = msg.FromLocalised, msg.To
		fromDisplay, toDisplay = msg.FromLocalised, msg.To
		fromNPC = true

		srcColor = r.app.Config.Colors.HeaderNPC
		dstColor = r.app.Config.Colors.HeaderSent
	case *journal_app.WingInviteMessage:
		from = msg.From
		fromDisplay = msg.From

		srcColor = r.CMDRColorFriendEnemy(msg.From)
		dstColor = r.app.Config.Colors.HeaderSent
	case *journal_app.FriendInviteMessage:
		from = msg.From
		fromDisplay = msg.From

		srcColor = r.CMDRColorFriendEnemy(msg.From)
		dstColor = r.app.Config.Colors.HeaderSent
	default:
		Logger.Error("nothing ever happens", zap.String("type", fmt.Sprintf("%T", msg)))
		return 0
	}

	if fromNPC {
		if r.app.Config.Viewports.Appearance.Toggles.PrefixNPC {
			fromDisplay = "NPC " + fromDisplay
		}
	} else {
		if r.app.Config.Viewports.Appearance.Toggles.PrefixCMDR {
			fromDisplay = "CMDR " + fromDisplay
		}
	}
	if r.app.Config.Viewports.Appearance.Toggles.PrefixCMDR {
		toDisplay = "CMDR " + toDisplay
	}

	currentSessionCounter := r.app.State.SessionCounter.Load()

	srcColorHover := dimColorForOutdatedSession(srcColor, sessionCounter, currentSessionCounter)
	dstColorHover := dimColorForOutdatedSession(dstColor, sessionCounter, currentSessionCounter)
	headerDefault := dimColorForOutdatedSession(r.app.Config.Colors.HeaderDefault, sessionCounter, currentSessionCounter)

	xOffset := float32(0) + float32(r.app.Config.Viewports.Appearance.MarkerWidth)
	if r.app.Config.GenericImages.Enabled {
		xOffset += vars.RWProfilePictureSize
	}

	var fontSize = float32(r.app.Config.Fonts.Size)
	var fontSpacing = float32(r.app.Config.Fonts.Spacing)

	txt = " ["
	m := rl.MeasureTextEx(r.app.State.Fonts.Main, txt, fontSize, fontSpacing)
	rl.DrawTextEx(r.app.State.Fonts.Main, txt, rl.Vector2{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize}, fontSize, fontSpacing, headerDefault.AsRGBA())
	xOffset += m.X

	srcM := rl.MeasureTextEx(r.app.State.Fonts.Main, fromDisplay, fontSize, fontSpacing)

	if !r.s.OptionViewActive && rl.CheckCollisionPointRec(r.s.MousePosition, rl.Rectangle{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize, srcM.X, fontSize}) {
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) || rl.IsMouseButtonDown(rl.MouseButtonRight) {
			srcColorHover = srcColorHover.Darken(ClickDarken)
		} else {
			srcColorHover = srcColorHover.Lighten(HoverBrighten)
		}
		if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
			info, ok := r.app.State.Inara.CMDRInfo.Get(from)
			if ok {
				rl.OpenURL(info.ProfileInfo.InaraURL)
			} else {
				rl.OpenURL(fmt.Sprintf("https://inara.cz/elite/cmdrs/?search=%s", url.QueryEscape(from)))
			}
		}
		if rl.IsMouseButtonReleased(rl.MouseButtonRight) {
			rl.SetClipboardText(from)
			r.app.State.BubbleMessage.SetMessage("name copied!", r.s.MousePosition)
		}
		r.s.IsCurrentlyHoveringOverCmdrName = true
		r.s.LastCmdrNameHover = from
		r.s.HoverPos = rl.Vector2{vars.ROstartX + xOffset + srcM.X, vars.ROstartY - vars.RWoffsetYcum - fontSize}
		r.s.AlternativeHoverPos = rl.Vector2{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize}
	}

	txt = fromDisplay
	m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, fontSize, fontSpacing)
	rl.DrawTextEx(r.app.State.Fonts.Main, txt, rl.Vector2{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize}, fontSize, fontSpacing, srcColorHover.AsRGBA())
	xOffset += m.X

	inaraIcons := []state.DataState{inara1}
	if inara1 == state.StateOK {
		r.app.State.Inara.CMDRInfoKeepAlive.Set(from, r.s.CurrentTime.Load())
	}

	if channel == journal_app.ChannelPlayer {
		inaraIcons = append(inaraIcons, inara2)
		if inara2 == state.StateOK {
			r.app.State.Inara.CMDRInfoKeepAlive.Set(to, r.s.CurrentTime.Load())
		}

		txt = " > "
		m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, fontSize, fontSpacing)
		rl.DrawTextEx(r.app.State.Fonts.Main, txt, rl.Vector2{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize}, fontSize, fontSpacing, headerDefault.AsRGBA())
		xOffset += m.X

		dstM := rl.MeasureTextEx(r.app.State.Fonts.Main, toDisplay, fontSize, fontSpacing)
		if !r.s.OptionViewActive && rl.CheckCollisionPointRec(r.s.MousePosition, rl.Rectangle{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize, dstM.X, fontSize}) {
			dstColorHover = adjustForClickOrHover(dstColorHover)
			if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
				info, ok := r.app.State.Inara.CMDRInfo.Get(to)
				if ok {
					rl.OpenURL(info.ProfileInfo.InaraURL)
				} else {
					rl.OpenURL(fmt.Sprintf("https://inara.cz/elite/cmdrs/?search=%s", url.QueryEscape(to)))
				}
			}
			if rl.IsMouseButtonReleased(rl.MouseButtonRight) {
				rl.SetClipboardText(to)
				r.app.State.BubbleMessage.SetMessage("name copied!", r.s.MousePosition)
			}
			r.s.IsCurrentlyHoveringOverCmdrName = true
			r.s.LastCmdrNameHover = to
			r.s.HoverPos = rl.Vector2{vars.ROstartX + xOffset + dstM.X, vars.ROstartY - vars.RWoffsetYcum - fontSize}
			r.s.AlternativeHoverPos = rl.Vector2{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize}
		}

		txt = toDisplay
		m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, fontSize, fontSpacing)
		rl.DrawTextEx(r.app.State.Fonts.Main, txt, rl.Vector2{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize}, fontSize, fontSpacing, dstColorHover.AsRGBA())
		xOffset += m.X
	}

	txt = "] "
	m = rl.MeasureTextEx(r.app.State.Fonts.Main, txt, fontSize, fontSpacing)
	rl.DrawTextEx(r.app.State.Fonts.Main, txt, rl.Vector2{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize}, fontSize, fontSpacing, headerDefault.AsRGBA())
	xOffset += m.X

	var color = r.app.Config.Colors.InaraLogo.Lighten(0.2)

	if r.app.Config.InaraSettings.ProfileIndicator {
		for _, s := range inaraIcons {
			iconRect := rl.Rectangle{vars.ROstartX + xOffset, vars.ROstartY - vars.RWoffsetYcum - fontSize, fontSize, fontSize}

			if s == state.StateOK || s == state.StateOKButMissing {
				rl.DrawTexturePro(
					r.app.State.Inara.InaraLogo,
					rl.Rectangle{0, 0, float32(r.app.State.Inara.InaraLogo.Width), float32(r.app.State.Inara.InaraLogo.Height)},
					iconRect,
					rl.Vector2{},
					0,
					color.AsRGBA(),
				)
			}

			if s == state.StateOKButMissing {
				type pair struct {
					a, b rl.Vector2
				}

				x1, x2 := vars.ROstartX+xOffset, vars.ROstartX+xOffset+fontSize
				y1, y2 := vars.ROstartY-vars.RWoffsetYcum-fontSize, vars.ROstartY-vars.RWoffsetYcum
				z := fontSize / 8

				for _, p := range []pair{
					{
						rl.Vector2{x1 + z, y1 + z},
						rl.Vector2{x2 - z, y2 - z},
					},
					{
						rl.Vector2{x2 - z, y1 + z},
						rl.Vector2{x1 + z, y2 - z},
					},
				} {
					rl.DrawLineEx(p.a, p.b, fontSize/6, r.app.Config.Colors.GeneralCritical.Desaturate(0.4).AsRGBA())
				}
			}

			if s == state.StateQueued || s == state.StateDownloading {
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
					iconRect,
					rl.Vector2{},
					0,
					rl.White,
				)
			}

			xOffset += fontSize
		}
	}

	r.renderTime(timestamp, vars, sessionCounter)

	vars.RWoffsetYcum += fontSize
	return fontSize
}

// drawWithSplit renders a message text at given position with specified maximum width.
// it calculates how much vertical space is needed for given maxWidth value, then
// renders it and returns how much of vertical space is occupied by a drawn text.
func (r *Renderer) drawWithSplit(message string, position rl.Vector2, maxWidth float32, bodyColor my_noire.Color) (bodyHeight float32, hoveredOver bool) {
	var commands []TextDrawCommand

	// Base Color Calculation
	hlColor := r.app.Config.Colors.HighlightWord
	txtColor := bodyColor

	fontSize := float32(r.app.Config.Fonts.Size)
	fontSpacing := float32(r.app.Config.Fonts.Spacing)
	cursorX := float32(0)
	cursorY := float32(0)
	spaceWidth := rl.MeasureTextEx(r.app.State.Fonts.Main, " ", fontSize, fontSpacing).X
	hyphenWidth := rl.MeasureTextEx(r.app.State.Fonts.Main, "-", fontSize, fontSpacing).X

	words := strings.Split(message, " ")

	for _, word := range words {
		//isUrl := urlFinder.MatchString(word)
		isUrl := false
		isHighlighted := false

		lowerText := strings.ToLower(word)
		if r.app.State.WordList[lowerText] || lowerText == r.app.State.CurrentThings.CMDR.Load() {
			isHighlighted = true
		}

		wordWidth := rl.MeasureTextEx(r.app.State.Fonts.Main, word, fontSize, fontSpacing).X

		// Case A: Word fits on current line
		if cursorX+wordWidth <= maxWidth {
			commands = append(commands, TextDrawCommand{
				Text:          word,
				Position:      rl.Vector2{X: cursorX, Y: cursorY},
				IsWord:        true,
				IsUrl:         isUrl,
				IsHighlighted: isHighlighted,
				Width:         wordWidth,
			})
			cursorX += wordWidth + spaceWidth
			continue
		}

		// Case B: Wrap to new line if not already at start
		if cursorX > 0 {
			cursorX = 0
			cursorY += fontSize
		}

		// Case C: Word fits on new line
		if wordWidth <= maxWidth {
			commands = append(commands, TextDrawCommand{
				Text:          word,
				Position:      rl.Vector2{X: cursorX, Y: cursorY},
				IsWord:        true,
				IsUrl:         isUrl,
				IsHighlighted: isHighlighted,
				Width:         wordWidth,
			})
			cursorX += wordWidth + spaceWidth
			continue
		}

		// Case D: Hard Split (Word wider than line)
		runes := []rune(word)
		processedIdx := 0

		for processedIdx < len(runes) {
			remainder := string(runes[processedIdx:])
			remainderWidth := rl.MeasureTextEx(r.app.State.Fonts.Main, remainder, fontSize, fontSpacing).X

			// Subcase D1: Remainder fits
			if remainderWidth <= maxWidth {
				commands = append(commands, TextDrawCommand{
					Text:          remainder,
					Position:      rl.Vector2{X: 0, Y: cursorY},
					IsWord:        true,
					IsUrl:         isUrl,
					IsHighlighted: isHighlighted,
					Width:         remainderWidth,
				})
				cursorX += remainderWidth + hyphenWidth
				break
			}

			// Subcase D2: Must cut
			currentWidth := float32(0)
			cutIdx := -1

			for i := processedIdx; i < len(runes); i++ {
				charW := rl.MeasureTextEx(r.app.State.Fonts.Main, string(runes[i]), fontSize, fontSpacing).X
				if currentWidth+charW+hyphenWidth > maxWidth {
					cutIdx = i
					break
				}
				currentWidth += charW
			}

			if cutIdx == -1 || cutIdx == processedIdx {
				cutIdx = processedIdx + 1
			}

			chunkText := string(runes[processedIdx:cutIdx]) + "-"
			chunkWidth := rl.MeasureTextEx(r.app.State.Fonts.Main, chunkText, fontSize, fontSpacing).X // Measure exact chunk width

			commands = append(commands, TextDrawCommand{
				Text:          chunkText,
				Position:      rl.Vector2{X: 0, Y: cursorY},
				IsWord:        false,
				IsUrl:         isUrl,
				IsHighlighted: isHighlighted,
				Width:         chunkWidth,
			})

			cursorY += fontSize
			processedIdx = cutIdx
		}
	}

	bodyHeight = cursorY + fontSize

	for _, cmd := range commands {
		pos := rl.Vector2{position.X + cmd.Position.X, position.Y - bodyHeight + cmd.Position.Y}

		if !r.s.OptionViewActive && rl.CheckCollisionPointRec(r.s.MousePosition, rl.Rectangle{
			pos.X,
			pos.Y,
			cmd.Width + spaceWidth,
			fontSize,
		}) {
			hoveredOver = true
			break
		}
	}

	if hoveredOver {
		txtColor = adjustForClickOrHover(txtColor)
	}

	for _, cmd := range commands {
		pos := rl.Vector2{position.X + cmd.Position.X, position.Y - bodyHeight + cmd.Position.Y}

		switch {
		case cmd.IsUrl:
			rl.DrawTextEx(r.app.State.Fonts.Main, cmd.Text, pos, fontSize, fontSpacing, r.app.Config.Colors.HighlightUrl.AsRGBA())
		case cmd.IsHighlighted:
			rl.DrawTextEx(r.app.State.Fonts.Main, cmd.Text, pos, fontSize, fontSpacing, hlColor.AsRGBA())
		default:
			rl.DrawTextEx(r.app.State.Fonts.Main, cmd.Text, pos, fontSize, fontSpacing, txtColor.AsRGBA())
		}
	}
	return bodyHeight, hoveredOver
}

func (r *Renderer) renderBody(msgUntyped journal_app.MessageEvent, vars *SomeVariables) float32 {
	var (
		txt            string
		sessionCounter int
		bodyColor      my_noire.Color

		extraParams map[string]string
		txtNpc      string
		npcMessage  bool
	)

	switch msg := msgUntyped.(type) {
	case *journal_app.SentMessage:
		txt = msg.Message
		bodyColor = *r.ChannelToColor[msg.Channel]
		sessionCounter = msg.SessionCounter
	case *journal_app.ReceivedCMDRMessage:
		txt = msg.Message
		bodyColor = *r.ChannelToColor[msg.Channel]
		sessionCounter = msg.SessionCounter
	case *journal_app.ReceivedNPCMessage:
		bodyColor = r.app.Config.Colors.NPCNormal
		sessionCounter = msg.SessionCounter
		txt = msg.MessageLocalised
		txtNpc = msg.Message
		npcMessage = true
		extraParams = msg.DecodedMessage.Parameters
	default:
		Logger.Error("not supported!", zap.String("type", fmt.Sprintf("%T", msg)))
		return 0
	}

	imageOffset := float32(0)
	if r.app.Config.GenericImages.Enabled {
		imageOffset = vars.RWProfilePictureSize
	}
	startX := vars.ROstartX + imageOffset + float32(r.app.Config.Viewports.Appearance.MessageIndentPad) + float32(r.app.Config.Viewports.Appearance.MarkerWidth)
	maxWidth := vars.ROstopX - startX
	startY := vars.ROstartY - vars.RWoffsetYcum

	pos := rl.Vector2{startX, startY}

	bodyColor = dimColorForOutdatedSession(bodyColor, sessionCounter, r.app.State.SessionCounter.Load())

	bodyHeight, hover := r.drawWithSplit(txt, pos, maxWidth, bodyColor)
	if hover {
		if rl.IsMouseButtonReleased(rl.MouseButtonRight) {
			if npcMessage {
				rl.SetClipboardText(txt + " (" + txtNpc + ")")
			} else {
				rl.SetClipboardText(txt)
			}
			r.app.State.BubbleMessage.SetMessage("message copied!", r.s.MousePosition)
		}
		if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
			urls := urlFinder.FindAllString(txt, -1)
			if len(urls) > 0 {
				rl.OpenURL(urls[0])
				r.app.State.BubbleMessage.SetMessage("url opened!!", r.s.MousePosition)
			}
		}

		if len(extraParams) != 0 {
			r.s.LastExtraParametersHover = extraParams
			r.s.IsCurrentlyHoveringOverExtraParameters = true
			r.s.HoverPos = rl.Vector2{startX + maxWidth, startY - bodyHeight}
			r.s.AlternativeHoverPos = rl.Vector2{startX, startY - bodyHeight}
		}
	}

	if r.app.Config.GenericImages.Enabled {
		bodyHeight = max(bodyHeight, vars.RWProfilePictureSize-float32(r.app.Config.Fonts.Size))
	}

	vars.RWoffsetYcum += bodyHeight

	return bodyHeight
}

func (r *Renderer) renderNewMessageMarker(totalHeight float32, timestamp time.Time, vars *SomeVariables) {
	since := time.Since(timestamp)
	fadeoutDuraton := since.Milliseconds()

	mhold := int64(r.app.Config.Viewports.Appearance.MarkerHoldTime * 1000)
	mfadeout := int64(r.app.Config.Viewports.Appearance.MarkerFadeoutTime * 1000)

	markerColor := r.app.Config.Colors.MewMessageMarker

	switch {
	case fadeoutDuraton < mhold:
		markerColor.Alpha = r.app.Config.Colors.MewMessageMarker.Alpha
	case fadeoutDuraton >= mhold && fadeoutDuraton < mhold+mfadeout:
		progress := float64(float64(fadeoutDuraton-mhold) / float64(mfadeout))
		markerColor.Alpha = r.app.Config.Colors.MewMessageMarker.Alpha * (1 - progress)
	default:
		markerColor.Alpha = 0
	}

	rl.DrawRectangle(int32(vars.ROstartX)+2, int32(vars.ROstartY-vars.RWoffsetYcum)+2, int32(r.app.Config.Viewports.Appearance.MarkerWidth)-4, int32(totalHeight)-2, markerColor.AsRGBA())
}

func (r *Renderer) renderTerminator(vars *SomeVariables, brighten float64) {
	up := r.app.Config.Colors.NoMoreMessages
	down := r.app.Config.Colors.NoMoreMessages.Lighten(brighten / 16).Saturate(brighten / 16)
	up.Alpha = 0

	strips := rl.Black
	strips.A = 64

	rl.DrawRectangleGradientV(int32(vars.ROstartX), int32(vars.ROstartY-vars.RWoffsetYcum-200), int32(vars.ROstopX-vars.ROstartX), 200, up.AsRGBA(), down.AsRGBA())
	rl.DrawTextureRec(
		r.app.State.Textures.Terminator.Texture,
		rl.Rectangle{vars.ROstartX + float32(math.Mod(float64(time.Now().UnixMilli())/2000, 1))*60, 0, float32(vars.ROstopX - vars.ROstartX), -float32(r.app.State.Textures.Terminator.Texture.Height)},
		rl.Vector2{vars.ROstartX, vars.ROstartY - vars.RWoffsetYcum - 200},
		strips,
	)

	text := "[no older messages]"
	m := rl.MeasureTextEx(r.app.State.Fonts.Main, text, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))
	rl.DrawTextEx(
		r.app.State.Fonts.Main,
		text,
		rl.Vector2{vars.ROstartX + (vars.ROstopX-vars.ROstartX)/2 - m.X/2, vars.ROstartY - vars.RWoffsetYcum - float32(r.app.Config.Fonts.Size)*1.5},
		float32(r.app.Config.Fonts.Size),
		float32(r.app.Config.Fonts.Spacing),
		r.app.Config.Colors.NoMoreMessagesText.AsRGBA(),
	)
}

func (r *Renderer) renderSent(msg *journal_app.SentMessage, vars *SomeVariables) {
	var h float32

	h += r.renderBody(msg, vars)
	h += r.renderHeader(msg, vars)
	r.renderProfilePicture(msg, vars)
	r.renderNewMessageMarker(h, msg.GeneratedAt, vars)

	vars.RWoffsetYcum += float32(r.app.Config.Viewports.Appearance.MessageSpacing)
}

func (r *Renderer) renderReceived(msgUntyped journal_app.MessageEvent, vars *SomeVariables) {
	var h float32

	h += r.renderBody(msgUntyped, vars)
	h += r.renderHeader(msgUntyped, vars)
	r.renderProfilePicture(msgUntyped, vars)
	var timestamp time.Time
	switch msg := msgUntyped.(type) {
	case *journal_app.ReceivedCMDRMessage:
		timestamp = msg.GeneratedAt
	case *journal_app.ReceivedNPCMessage:
		timestamp = msg.GeneratedAt
	}
	r.renderNewMessageMarker(h, timestamp, vars)

	vars.RWoffsetYcum += float32(r.app.Config.Viewports.Appearance.MessageSpacing)
}

func (r *Renderer) renderProfilePicture(msgUntyped journal_app.MessageEvent, vars *SomeVariables) float32 {
	if !r.app.Config.GenericImages.Enabled {
		return 0
	}
	var session = msgUntyped.Session()

	var img rl.Texture2D
	var imgOk bool
	var cmdrName string
	var imgColor my_noire.Color

	switch msg := msgUntyped.(type) {
	case *journal_app.SentMessage:
		imgColor = r.app.Config.Colors.HeaderSent
		img, imgOk = r.app.State.Inara.CMDRTextures.Get(msg.From)
		cmdrName = msg.From
	case *journal_app.ReceivedCMDRMessage:
		img, imgOk = r.app.State.Inara.CMDRTextures.Get(msg.From)
		imgColor = r.CMDRColorFriendEnemy(msg.From)
		cmdrName = msg.From
	case *journal_app.ReceivedNPCMessage:
		imgColor = r.app.Config.Colors.HeaderNPC
		cmdrName = msg.FromLocalised
	case *journal_app.WingInviteMessage:
		img, imgOk = r.app.State.Inara.CMDRTextures.Get(msg.From)
		imgColor = r.CMDRColorFriendEnemy(msg.From)
		cmdrName = msg.From
	case *journal_app.FriendInviteMessage:
		img, imgOk = r.app.State.Inara.CMDRTextures.Get(msg.From)
		imgColor = r.CMDRColorFriendEnemy(msg.From)
		cmdrName = msg.From
	default:
		Logger.Error("Unknown message type received", zap.Any("msg", msgUntyped))
		return 0
	}

	imgColor = dimColorForOutdatedSession(imgColor, session, r.app.State.SessionCounter.Load())
	white := dimColorForOutdatedSession(my_noire.NewHex("ffffffff"), session, r.app.State.SessionCounter.Load())

	xOffset := float32(0) + float32(r.app.Config.Viewports.Appearance.MarkerWidth)
	bw := float32(r.app.Config.GenericImages.BorderWidth)

	if imgOk {
		sizeX, sizeY := howToResize(float32(img.Width), float32(img.Height))
		ofstX := (vars.RWProfilePictureSize - (vars.RWProfilePictureSize * sizeX)) / 2
		ofstY := (vars.RWProfilePictureSize - (vars.RWProfilePictureSize * sizeY)) / 2
		dst := rl.Rectangle{
			ofstX + bw + vars.ROstartX + xOffset,
			ofstY + bw + vars.ROstartY - vars.RWoffsetYcum,
			vars.RWProfilePictureSize*sizeX - bw*2,
			vars.RWProfilePictureSize*sizeY - bw*2,
		}
		rl.DrawTexturePro(
			img,
			rl.Rectangle{0, 0, float32(img.Width), float32(img.Height)},
			dst,
			rl.Vector2{},
			0,
			white.AsRGBA(),
		)
		r.app.State.Inara.CMDRTexturesKeepAlive.Set(cmdrName, r.s.CurrentTime.Load())
	} else {
		for _, thing := range []struct {
			texture rl.Texture2D
			color   my_noire.Color
		}{
			{r.app.State.Inara.AvatarPlaceholderCMDRShoulders, dimColorForOutdatedSession(GenerateSuitColor(cmdrName, 0), session, r.app.State.SessionCounter.Load())},
			{r.app.State.Inara.AvatarPlaceholderCMDRHead, dimColorForOutdatedSession(GenerateSkinTone(cmdrName, 0), session, r.app.State.SessionCounter.Load())},
			{r.app.State.Inara.AvatarPlaceholderCMDRQuestion, r.app.Config.Colors.Background},
		} {
			img = thing.texture

			sizeX, sizeY := howToResize(float32(img.Width), float32(img.Height))
			ofstX := (vars.RWProfilePictureSize - (vars.RWProfilePictureSize * sizeX)) / 2
			ofstY := (vars.RWProfilePictureSize - (vars.RWProfilePictureSize * sizeY)) / 2

			dst := rl.Rectangle{
				ofstX + bw + vars.ROstartX + xOffset,
				ofstY + bw + vars.ROstartY - vars.RWoffsetYcum,
				vars.RWProfilePictureSize*sizeX - bw*2,
				vars.RWProfilePictureSize*sizeY - bw*2,
			}

			rl.DrawTexturePro(
				img,
				rl.Rectangle{0, 0, float32(img.Width), float32(img.Height)},
				dst,
				rl.Vector2{},
				0,
				thing.color.AsRGBA(),
			)
			//r.app.State.Inara.CMDRTexturesLastDrawnFrame.Set(cmdrName, r.s.CurrentFrame)
		}
	}

	borderRect := rl.Rectangle{
		vars.ROstartX + xOffset,
		vars.ROstartY - vars.RWoffsetYcum,
		vars.RWProfilePictureSize,
		vars.RWProfilePictureSize,
	}
	rl.DrawRectangleLinesEx(borderRect, bw, imgColor.AsRGBA())

	texState, _ := r.app.State.Inara.CMDRTexturesState.Get(cmdrName)
	switch texState {
	case state.StateDownloading, state.StateQueued:
		sizeX, sizeY := howToResize(float32(EdLoaderFrameWidth), float32(EdLoaderFrameHeight))
		sizeX, sizeY = sizeX*EDLoaderPFPSize, sizeY*EDLoaderPFPSize

		ofstX := (vars.RWProfilePictureSize - (vars.RWProfilePictureSize * sizeX)) / 2
		ofstY := (vars.RWProfilePictureSize - (vars.RWProfilePictureSize * sizeY)) / 2

		frame := getEdLoaderFrame()
		src := rl.Rectangle{
			float32(EdLoaderFrameWidth) * float32(frame),
			0,
			float32(EdLoaderFrameWidth),
			float32(EdLoaderFrameHeight),
		}
		dst := rl.Rectangle{
			ofstX + bw + vars.ROstartX + xOffset,
			ofstY + bw + vars.ROstartY - vars.RWoffsetYcum,
			vars.RWProfilePictureSize*sizeX - bw*2,
			vars.RWProfilePictureSize*sizeY - bw*2,
		}

		rl.DrawTexturePro(
			r.app.State.Textures.EliteLoader,
			src,
			dst,
			rl.Vector2{},
			0,
			white.AsRGBA(),
		)
	default:
		break
	}

	return vars.RWProfilePictureSize
}

func colorForNotification(decoded message.DecodedMessage) {
}

func (r *Renderer) renderReceivedNotification(msg *journal_app.NotificationMessage, vars *SomeVariables) {
	var level config.Level
	var notificationColor my_noire.Color

	who := msg.DecodedMessage.What
	subType := msg.DecodedMessage.SubWhat

	if msg.DecodedMessage.SubWhat == "" {
		filter, ok := r.app.ProcessedConfig.Filters.Notifications[who]
		if ok {
			level = filter.Base.Level
		}
	} else {
		filter, ok := r.app.ProcessedConfig.Filters.Notifications[who]
		if ok {
			subFilter, ok := filter.SubTypes[subType]
			if ok {
				level = subFilter.Level
			}
		}
	}

	switch level {
	case config.LevelUseless:
		notificationColor = r.app.Config.Colors.NotificationNormal
	case config.LevelInteresting:
		notificationColor = r.app.Config.Colors.NotificationInteresting
	case config.LevelGood:
		notificationColor = r.app.Config.Colors.NotificationPositive
	case config.LevelWarning:
		notificationColor = r.app.Config.Colors.NotificationWarning
	case config.LevelCritical:
		notificationColor = r.app.Config.Colors.NotificationCritical
	default:
		notificationColor = r.app.Config.Colors.NotificationNormal
	}

	notificationColor = dimColorForOutdatedSession(notificationColor, msg.SessionCounter, r.app.State.SessionCounter.Load())

	txt := "> "
	m := rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))

	pos := rl.Vector2{
		vars.ROstartX + float32(r.app.Config.Viewports.Appearance.MarkerWidth) + m.X,
		vars.ROstartY - vars.RWoffsetYcum,
	}
	maxWidth := vars.ROstopX - pos.X
	bodyHeight, hover := r.drawWithSplit(msg.MessageLocalised, pos, maxWidth, notificationColor)

	txtPos := rl.Vector2{
		vars.ROstartX + float32(r.app.Config.Viewports.Appearance.MarkerWidth),
		vars.ROstartY - vars.RWoffsetYcum - bodyHeight,
	}
	rl.DrawTextEx(r.app.State.Fonts.Main, txt, txtPos, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), notificationColor.AsRGBA())
	if hover {
		r.renderTime(msg.Timestamp, vars, msg.SessionCounter)

		if rl.IsMouseButtonReleased(rl.MouseRightButton) {
			rl.SetClipboardText(msg.MessageLocalised + " (" + msg.Message + ")")
			r.app.State.BubbleMessage.SetMessage("notification copied!", r.s.MousePosition)
		}
		r.s.IsCurrentlyHoveringOverExtraParameters = true
		r.s.LastExtraParametersHover = msg.DecodedMessage.Parameters

		p1 := rl.Vector2{
			vars.ROstopX,
			vars.ROstartY - vars.RWoffsetYcum - bodyHeight,
		}
		p2 := rl.Vector2{
			vars.ROstartX,
			vars.ROstartY - vars.RWoffsetYcum - bodyHeight,
		}

		r.s.HoverPos = p1
		r.s.AlternativeHoverPos = p2
	}

	vars.RWoffsetYcum += bodyHeight
	r.renderNewMessageMarker(bodyHeight, msg.GeneratedAt, vars)
	vars.RWoffsetYcum += float32(r.app.Config.Viewports.Appearance.MessageSpacing)
}

func (r *Renderer) renderGenericJournalMessage(msg *journal_app.GenericEvent, vars *SomeVariables) {
	headerColor := dimColorForOutdatedSession(r.app.Config.Colors.HeaderDefault, msg.SessionCounter, r.app.State.SessionCounter.Load())

	txt := fmt.Sprintf("[%s]", msg.MessageBody)

	pos := rl.Vector2{
		vars.ROstartX + float32(r.app.Config.Viewports.Appearance.MarkerWidth),
		vars.ROstartY - vars.RWoffsetYcum,
	}
	maxWidth := vars.ROstopX - pos.X
	bodyHeight, hover := r.drawWithSplit(txt, pos, maxWidth, headerColor)
	if hover {
		r.renderTime(msg.Timestamp, vars, msg.SessionCounter)
		if rl.IsMouseButtonReleased(rl.MouseRightButton) {
			txt := msg.MessageBody
			rl.SetClipboardText(txt)
			r.app.State.BubbleMessage.SetMessage("entry info copied!", r.s.MousePosition)
		}
	}

	vars.RWoffsetYcum += bodyHeight
	r.renderNewMessageMarker(bodyHeight, msg.GeneratedAt, vars)
	vars.RWoffsetYcum += float32(r.app.Config.Viewports.Appearance.MessageSpacing)
}

func (r *Renderer) renderInvite(msgUntyped journal_app.MessageEvent, vars *SomeVariables) {
	var body string

	notificationColor := dimColorForOutdatedSession(r.app.Config.Colors.NotificationPositive, msgUntyped.Session(), r.app.State.SessionCounter.Load())

	txt := " > "
	m := rl.MeasureTextEx(r.app.State.Fonts.Main, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing))

	switch msgUntyped.(type) {
	case *journal_app.FriendInviteMessage:
		body = "friend invitation"
	case *journal_app.WingInviteMessage:
		body = "wing invitation"
	case *journal_app.SquadronInviteMessage:
		body = "squadron invitation"
	}

	imageOffset := float32(0)
	if r.app.Config.GenericImages.Enabled {
		imageOffset = vars.RWProfilePictureSize
	}

	pos := rl.Vector2{
		vars.ROstartX + imageOffset + float32(r.app.Config.Viewports.Appearance.MarkerWidth) + m.X,
		vars.ROstartY - vars.RWoffsetYcum,
	}
	maxWidth := vars.ROstopX - pos.X
	bodyHeight, hover := r.drawWithSplit(body, pos, maxWidth, notificationColor)
	txtPos := rl.Vector2{
		vars.ROstartX + imageOffset + float32(r.app.Config.Viewports.Appearance.MarkerWidth),
		vars.ROstartY - vars.RWoffsetYcum - bodyHeight,
	}

	rl.DrawTextEx(r.app.State.Fonts.Main, txt, txtPos, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.Spacing), notificationColor.AsRGBA())
	if r.app.Config.GenericImages.Enabled {
		bodyHeight = max(bodyHeight, vars.RWProfilePictureSize-float32(r.app.Config.Fonts.Size))
	}

	if hover {
		//r.renderTime(msg.Timestamp, vars, msg.SessionCounter)
		if rl.IsMouseButtonReleased(rl.MouseRightButton) {
			rl.SetClipboardText(body)
			r.app.State.BubbleMessage.SetMessage("notification copied!", r.s.MousePosition)
		}
	}
	vars.RWoffsetYcum += bodyHeight

	bodyHeight += r.renderInvitationHeader(msgUntyped, vars)
	r.renderProfilePicture(msgUntyped, vars)
	r.renderNewMessageMarker(bodyHeight, msgUntyped.GeneratedTime(), vars)

	vars.RWoffsetYcum += float32(r.app.Config.Viewports.Appearance.MessageSpacing)
}

func (r *Renderer) renderTime(timestamp time.Time, vars *SomeVariables, sessionCounter int) (xOffset float32) {
	if !r.app.Config.Viewports.Appearance.Toggles.Timestamps {
		return xOffset
	}

	var color my_noire.Color
	var txt string
	var imageOffset float32

	if r.app.Config.Viewports.Appearance.Toggles.TimestampEliteFormat {
		txt = fmt.Sprintf("[%s]", utils.EliteTimestamp(timestamp).Format(r.app.Config.Viewports.Appearance.TimestampFormat))
		color = dimColorForOutdatedSession(r.app.Config.Colors.TimeFormatElite, sessionCounter, r.app.State.SessionCounter.Load())
	} else {
		txt = fmt.Sprintf("(%s)", timestamp.Local().Format(r.app.Config.Viewports.Appearance.TimestampFormat))
		color = dimColorForOutdatedSession(r.app.Config.Colors.TimeFormatLocal, sessionCounter, r.app.State.SessionCounter.Load())
	}

	m := rl.MeasureTextEx(r.app.State.Fonts.MainMono, txt, float32(r.app.Config.Fonts.Size), float32(r.app.Config.Fonts.SpacingMono))
	pos := rl.Vector2{vars.ROstopX - xOffset - m.X - imageOffset, vars.ROstartY - vars.RWoffsetYcum - float32(r.app.Config.Fonts.Size)}
	rect := rl.Rectangle{
		pos.X,
		pos.Y,
		m.X,
		float32(r.app.Config.Fonts.Size),
	}
	rl.DrawRectangleRec(rect, r.app.Config.Colors.Background.AsRGBA())

	rl.DrawTextEx(
		r.app.State.Fonts.MainMono,
		txt,
		pos,
		float32(r.app.Config.Fonts.Size),
		float32(r.app.Config.Fonts.SpacingMono),
		color.AsRGBA(),
	)

	xOffset += m.X

	return xOffset
}
