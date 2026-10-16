package app

import (
	"fmt"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/config"
)

type uiFilterGroup struct {
	DisplayName string
	Key         string
	SubTypes    []config.MsgType
}

func (r *Renderer) renderTabFilters(startPosY float32) {
	baseX := float32(r.s.CanvasMiddleX + FilterListOffsetX)

	// Setup Scroll Area & Mouse Wheel Logic
	viewRect := rl.Rectangle{
		X:      baseX - FilterLabelWidth - 100, // Extend left to cover long labels
		Y:      startPosY,
		Width:  (FilterLabelWidth * 2) + 200, // Extend right to cover subtypes
		Height: float32(rl.GetScreenHeight()) - startPosY - 20,
	}

	if rl.CheckCollisionPointRec(r.s.MousePosition, viewRect) {
		wheel := r.getMouseWheelV()
		// Apply our speed multiplier here
		r.s.FilterScrollY += wheel.Y * ScrollSpeedMultiplier
	}

	if r.s.FilterScrollY > 0 {
		r.s.FilterScrollY = 0
	}
	minScroll := -(r.s.LastTotalHeight - viewRect.Height)
	if minScroll > 0 {
		minScroll = 0
	}
	if r.s.FilterScrollY < minScroll {
		r.s.FilterScrollY = minScroll
	}

	rl.BeginScissorMode(int32(viewRect.X), int32(viewRect.Y), int32(viewRect.Width), int32(viewRect.Height))

	currentY := startPosY + r.s.FilterScrollY

	// 1. Render Local Notifications (Passing Local Blacklist Mode)
	var localGroups []uiFilterGroup
	for _, s := range config.LocalNotificationsAvailable {
		localGroups = append(localGroups, uiFilterGroup{DisplayName: s.WhoForUser, Key: s.Who, SubTypes: s.MsgTypes})
	}
	currentY = r.renderFilterSection("Local Notifications", currentY, localGroups, r.app.Config.Filters.Notifications, &r.app.Config.Filters.NotificationsBlacklistMode)

	currentY += OptionHeightWithPad

	// 2. Render NPC Messages (Passing NPC Blacklist Mode)
	var npcGroups []uiFilterGroup
	for _, s := range config.NPCMessagesAvailable {
		npcGroups = append(npcGroups, uiFilterGroup{DisplayName: s.WhoForUser, Key: s.Who, SubTypes: s.MsgTypes})
	}
	currentY = r.renderFilterSection("NPC Messages", currentY, npcGroups, r.app.Config.Filters.NpcMessages, &r.app.Config.Filters.NpcMessagesBlacklistMode)

	rl.EndScissorMode()

	r.s.LastTotalHeight = currentY - (startPosY + r.s.FilterScrollY)
}

func (r *Renderer) renderFilterSection(title string, startY float32, groups []uiFilterGroup, stateMap map[string]config.Filter, isBlacklist *bool) float32 {
	currentY := startY
	saveAlign := gui.GetStyle(gui.LABEL, gui.TEXT_ALIGNMENT)

	baseX := float32(r.s.CanvasMiddleX + FilterListOffsetX)

	// 1. Draw Section Title and Blacklist Toggle
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	gui.Label(rl.Rectangle{baseX - 200, currentY, 400, OptionHeight}, title)
	currentY += OptionHeightWithPad

	modeText := "Mode: Whitelist"
	if *isBlacklist {
		modeText = "Mode: Blacklist"
	}
	// Centered toggle under the title
	_ = gui.CheckBox(rl.Rectangle{baseX - 50, currentY, float32(CheckboxWidth), OptionHeight}, modeText, isBlacklist)
	currentY += OptionHeightWithPad * 1.5

	for _, group := range groups {
		isExpanded := r.s.FilterExpanded[group.Key]
		filter := stateMap[group.Key] // Grab current filter struct

		// 2. Count checked subtypes to determine 'All' or 'Partial' state
		totalCount := len(group.SubTypes)
		checkedCount := 0

		if totalCount == 0 {
			if filter.Base != nil && filter.Base.Enabled {
				checkedCount = 1
			}
			totalCount = 1
		} else {
			for _, t := range group.SubTypes {
				if filter.SubTypes != nil && filter.SubTypes[t.Value] != nil && filter.SubTypes[t.Value].Enabled {
					checkedCount++
				}
			}
		}

		allChecked := checkedCount == totalCount
		isPartial := checkedCount > 0 && checkedCount < totalCount

		// 3. Draw Category Label (Right-aligned, before the icons)
		gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_RIGHT)
		gui.Label(rl.Rectangle{baseX - FilterLabelWidth - 80, currentY, FilterLabelWidth, OptionHeight}, group.DisplayName)

		// 4. Draw Collapse/Expand Icon Button
		iconID := 119 // ICON_ARROW_RIGHT_FILL
		if isExpanded {
			iconID = 120 // ICON_ARROW_DOWN_FILL
		}
		iconRect := rl.Rectangle{baseX - 70, currentY, OptionHeight, OptionHeight}
		if gui.Button(iconRect, fmt.Sprintf("#%d#", iconID)) {
			r.s.FilterExpanded[group.Key] = !isExpanded
		}

		// 5. Draw Master Category Checkbox
		checkAllRect := rl.Rectangle{baseX - 40, currentY, float32(CheckboxWidth), OptionHeight}

		// Pass strictly allChecked. If partial, raygui draws an empty box.
		allCheckedChanged := gui.CheckBox(checkAllRect, "", &allChecked)

		// Draw Partial Overlay (a one-third-sized inner square) inside the empty box
		if isPartial {
			overlayRect := rl.Rectangle{
				checkAllRect.X + checkAllRect.Width/3,
				checkAllRect.Y + checkAllRect.Height/3,
				checkAllRect.Width / 3,
				checkAllRect.Height / 3,
			}
			rl.DrawRectangleRec(overlayRect, ColorUIHLFdev.AsRGBA())
		}

		// Apply "Check All" / "Uncheck All" if user clicked the master checkbox
		if allCheckedChanged {
			if len(group.SubTypes) == 0 {
				if filter.Base == nil {
					filter.Base = &config.FilterInfo{}
				}
				filter.Base.Enabled = allChecked
			} else {
				if filter.SubTypes == nil {
					filter.SubTypes = make(map[string]*config.FilterInfo)
				}
				for _, t := range group.SubTypes {
					if filter.SubTypes[t.Value] == nil {
						filter.SubTypes[t.Value] = &config.FilterInfo{}
					}
					filter.SubTypes[t.Value].Enabled = allChecked
				}
			}
			stateMap[group.Key] = filter // Save mutated struct back to map
		}

		currentY += OptionHeightWithPad

		// 6. Draw Subtypes (Only if Expanded)
		if isExpanded {
			gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_LEFT)

			if len(group.SubTypes) == 0 {
				rect := rl.Rectangle{baseX + 10, currentY, float32(CheckboxWidth), OptionHeight}
				currentState := filter.Base != nil && filter.Base.Enabled
				stateUpdated := gui.CheckBox(rect, "[no subtypes]", &currentState)

				if stateUpdated {
					if filter.Base == nil {
						filter.Base = &config.FilterInfo{}
					}
					filter.Base.Enabled = currentState
					stateMap[group.Key] = filter // Save back to map
				}
				currentY += OptionHeightWithPad
			} else {
				for _, t := range group.SubTypes {
					rect := rl.Rectangle{baseX + 10, currentY, float32(CheckboxWidth), OptionHeight}
					currentState := filter.SubTypes != nil && filter.SubTypes[t.Value] != nil && filter.SubTypes[t.Value].Enabled
					stateUpdated := gui.CheckBox(rect, t.ValueForUser, &currentState)

					if stateUpdated {
						if filter.SubTypes == nil {
							filter.SubTypes = make(map[string]*config.FilterInfo)
						}
						if filter.SubTypes[t.Value] == nil {
							filter.SubTypes[t.Value] = &config.FilterInfo{}
						}
						filter.SubTypes[t.Value].Enabled = currentState
						stateMap[group.Key] = filter // Save back to map
					}
					currentY += OptionHeightWithPad
				}
			}
		}
	}

	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, saveAlign)
	return currentY + OptionHeightWithPad
}
