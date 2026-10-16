package app

import "time"

// TODO: todo
const (
	EdLoaderFrames      = 30
	EdLoaderFrameWidth  = 400
	EdLoaderFrameHeight = 320
	EDLoaderPFPSize     = 0.8

	PeekDuration     = time.Millisecond * 600
	GuiPeekAlpha     = 0.15
	GuiPeekBgAlpha   = 0.0
	GuiNormalAlpha   = 1.0
	GuiNormalBgAlpha = 0.6

	HoverBrighten float64 = 0.2
	ClickDarken   float64 = 0.15
	StatusDarken  float64 = 0.4

	SymmetricOffset          = 330
	TabRectWidth             = 100
	FilterLabelWidth         = 390
	FilterListOffsetX        = 250
	ScrollSpeedMultiplier    = 30.0
	AdditionalIconLabelWidth = 200
	ActionIconScale          = 6
	ActionIconEdge           = float32(ActionIconScale * 16)
	OptionWidth              = 200
	OptionHeight             = UIFontSize
	OptionHeightWithPad      = OptionHeight + 4
	CheckboxWidth            = 28
	CheckboxSpace            = 4
	ViewportSettingsOffset   = 330
	IconRow                  = 16
	LineCount                = 16 // lines of ui elements
	SoundListRows            = 16

	ColorsPerRow       = 8
	ColorHeight        = OptionHeight * 1.5
	ColorWidth         = OptionHeight * 2.0
	ColorHeightWithPad = OptionHeightWithPad * 2.0
	ColorBorder        = 4
	OutlineBorder      = 2
	ColorLabelWidth    = 400
	PickerOffsetX      = 50
	PickerWidth        = 300
	PickerHeight       = 300
	PickerAlphaPadding = 10
)
