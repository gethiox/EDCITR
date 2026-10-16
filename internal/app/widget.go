package app

import (
	"fmt"
	"math"
	"time"

	"github.com/gen2brain/raylib-go/raygui"
	"github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/exp/constraints"
)

type Widget interface {
	Draw(rect rl.Rectangle)
}

type OptionWidget struct {
	drawFunc func(rect rl.Rectangle)
}

func (w OptionWidget) Draw(rect rl.Rectangle) {
	w.drawFunc(rect)
}

type ButtonWidget struct {
	drawFunc func(rect rl.Rectangle)
}

func (w ButtonWidget) Draw(rect rl.Rectangle) {
	w.drawFunc(rect)
}

// type sliderFunc func(bounds rl.Rectangle, textLeft, textRight string, value, minValue, maxValue float32) float32
type sliderFunc func(bounds rl.Rectangle, textLeft, textRight string, value *float32, minValue, maxValue float32) bool

type widgetConfig struct {
	onChange        func()
	onEdit          func(isEditing bool) // Added this
	acceptDragDrop  bool
	acceptClipboard bool
}

type WidgetOption func(*widgetConfig)

func OnEdit(callback func(isEditing bool)) WidgetOption {
	return func(c *widgetConfig) {
		c.onEdit = callback
	}
}

func OnChange(callback func()) WidgetOption {
	return func(c *widgetConfig) {
		c.onChange = callback
	}
}

func AcceptDragAndDrop() WidgetOption {
	return func(c *widgetConfig) {
		c.acceptDragDrop = true
	}
}

func AcceptClipboard() WidgetOption {
	return func(c *widgetConfig) {
		c.acceptClipboard = true
	}
}

type Number interface {
	constraints.Integer | constraints.Float
}

func makeSlider[T Number](target *T, label string, min, max float32, drawFunc sliderFunc, opts ...WidgetOption) OptionWidget {
	config := widgetConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return OptionWidget{func(rect rl.Rectangle) {
		currentVal := float32(*target)
		var textVal string

		val, isInt := any(*target).(int)
		if isInt {
			textVal = fmt.Sprintf("%d", val)
		} else {
			textVal = fmt.Sprintf("%.2f", float64(*target))
		}

		updated := drawFunc(rect, label, textVal, &currentVal, min, max)

		_, isInt = any(*target).(int)
		if isInt {
			*target = T(math.Round(float64(currentVal)))
		} else {
			*target = T(currentVal)
		}

		if updated && config.onChange != nil {
			config.onChange()
		}
	}}
}

// makeDurationSlider creates a slider specifically for time.Duration.
// The 'unit' parameter determines the scale of the float32 value (e.g., time.Second).
func makeDurationSlider(target *time.Duration, label string, unit time.Duration, min, max float32, drawFunc sliderFunc, opts ...WidgetOption) OptionWidget {
	config := widgetConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return OptionWidget{drawFunc: func(rect rl.Rectangle) {
		// Convert the underlying nanoseconds to a float based on the requested unit
		currentVal := float32(float64(*target) / float64(unit))

		// Format the text label to 2 decimal places (adjust if you want a different default)
		textVal := fmt.Sprintf("%.2f", currentVal)

		// Draw the slider
		changed := drawFunc(rect, label, textVal, &currentVal, min, max)

		// Convert the result back to time.Duration nanoseconds
		*target = time.Duration(float64(currentVal) * float64(unit))

		// Trigger onChange if the value was modified
		if changed && config.onChange != nil {
			config.onChange()
		}
	}}
}

func DurationSlider(target *time.Duration, label string, unit time.Duration, min, max float32, opts ...WidgetOption) OptionWidget {
	return makeDurationSlider(target, label, unit, min, max, raygui.Slider, opts...)
}

func DurationSliderBar(target *time.Duration, label string, unit time.Duration, min, max float32, opts ...WidgetOption) OptionWidget {
	return makeDurationSlider(target, label, unit, min, max, raygui.SliderBar, opts...)
}

func Slider[T Number](target *T, label string, min, max float32, opts ...WidgetOption) OptionWidget {
	return makeSlider(target, label, min, max, raygui.Slider, opts...)
}

func SliderBar[T Number](target *T, label string, min, max float32, opts ...WidgetOption) OptionWidget {
	return makeSlider(target, label, min, max, raygui.SliderBar, opts...)
}

func Checkbox(target *bool, label string, opts ...WidgetOption) ButtonWidget {
	config := widgetConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return ButtonWidget{func(rect rl.Rectangle) {
		changed := raygui.CheckBox(rect, label, target)

		if changed && config.onChange != nil {
			config.onChange()
		}
	}}
}

func Button(label string, action func()) OptionWidget {
	return OptionWidget{func(rect rl.Rectangle) {
		if raygui.Button(rect, label) {
			action()
		}
	}}
}

func Toggle(target *bool, label string, opts ...WidgetOption) ButtonWidget {
	config := widgetConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return ButtonWidget{drawFunc: func(rect rl.Rectangle) {
		// Assuming your gui wrapper returns the new bool state
		changed := raygui.Toggle(rect, label, target)

		if changed && config.onChange != nil {
			config.onChange()
		}
	}}
}

func LabelButton(label string, action func()) OptionWidget {
	return OptionWidget{drawFunc: func(rect rl.Rectangle) {
		// LabelButtons behave like Buttons but without the standard button border
		if raygui.LabelButton(rect, label) {
			action()
		}
	}}
}

func ComboBox[T constraints.Integer](target *T, options string, opts ...WidgetOption) OptionWidget {
	config := widgetConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return OptionWidget{drawFunc: func(rect rl.Rectangle) {
		currentVal := int32(*target)

		// ComboBox takes int32 and returns int32, no pointers or editMode needed
		changed := raygui.ComboBox(rect, options, &currentVal)

		*target = T(currentVal)

		if changed && config.onChange != nil {
			config.onChange()
		}
	}}
}

//func TextBox(target *string, editMode *bool, opts ...WidgetOption) OptionWidget {
//	config := widgetConfig{}
//	for _, opt := range opts {
//		opt(&config)
//	}
//
//	return OptionWidget{drawFunc: func(rect rl.Rectangle) {
//		oldValue := *target
//
//		if raygui.TextBox(rect, target, 256, *editMode) {
//			*editMode = !*editMode
//			// Fire the callback when edit mode changes
//			if config.onEdit != nil {
//				config.onEdit(*editMode)
//			}
//		}
//
//		if *target != oldValue && config.onChange != nil {
//			config.onChange()
//		}
//	}}
//}

func ValueBoxFloat[T Number](target *T, textBuffer *string, label string, editMode *bool, opts ...WidgetOption) OptionWidget {
	config := widgetConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return OptionWidget{drawFunc: func(rect rl.Rectangle) {
		oldValue := *target
		val32 := float32(*target)

		if raygui.ValueBoxFloat(rect, label, textBuffer, &val32, *editMode) {
			*editMode = !*editMode
			if config.onEdit != nil {
				config.onEdit(*editMode)
			}
		}

		*target = T(val32)

		if *target != oldValue && config.onChange != nil {
			config.onChange()
		}
	}}
}

func Spinner[T constraints.Integer](target *T, label string, min, max int, editMode *bool, opts ...WidgetOption) OptionWidget {
	config := widgetConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return OptionWidget{drawFunc: func(rect rl.Rectangle) {
		oldValue := *target
		val32 := int32(*target)

		if raygui.Spinner(rect, label, &val32, min, max, *editMode) {
			*editMode = !*editMode
			if config.onEdit != nil {
				config.onEdit(*editMode)
			}
		}

		*target = T(val32)

		if *target != oldValue && config.onChange != nil {
			config.onChange()
		}
	}}
}

func ValueBox[T constraints.Integer](target *T, label string, min, max int, editMode *bool, opts ...WidgetOption) OptionWidget {
	config := widgetConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return OptionWidget{drawFunc: func(rect rl.Rectangle) {
		oldValue := *target
		val32 := int32(*target)

		if raygui.ValueBox(rect, label, &val32, min, max, *editMode) {
			*editMode = !*editMode
			if config.onEdit != nil {
				config.onEdit(*editMode)
			}
		}

		*target = T(val32)

		if *target != oldValue && config.onChange != nil {
			config.onChange()
		}
	}}
}

func DropdownBox[T constraints.Integer](target *T, options string, editMode *bool, opts ...WidgetOption) OptionWidget {
	config := widgetConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return OptionWidget{drawFunc: func(rect rl.Rectangle) {
		oldValue := *target
		val32 := int32(*target)

		if raygui.DropdownBox(rect, options, &val32, *editMode) {
			*editMode = !*editMode
			if config.onEdit != nil {
				config.onEdit(*editMode)
			}
		}
		*target = T(val32)

		if *target != oldValue && config.onChange != nil {
			config.onChange()
		}
	}}
}

func TextBox(target *string, label string, editMode *bool, opts ...WidgetOption) OptionWidget {
	config := widgetConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	var buffer string
	var wasEditing bool

	return OptionWidget{drawFunc: func(rect rl.Rectangle) {
		// --- Draw Label ---
		if label != "" {
			// Measure the text using the current default font size
			fontSize := raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SIZE)
			textWidth := float32(rl.MeasureText(label, int32(fontSize)))

			// Position the label to the left of the textbox, adding a small 8px gap
			labelRect := rl.NewRectangle(rect.X-textWidth-8, rect.Y, textWidth, rect.Height)

			// Temporarily force right-alignment for the label
			prevAlign := raygui.GetStyle(raygui.LABEL, raygui.TEXT_ALIGNMENT)
			raygui.SetStyle(raygui.LABEL, raygui.TEXT_ALIGNMENT, raygui.TEXT_ALIGN_RIGHT)

			raygui.Label(labelRect, label)

			// Restore the previous alignment style immediately
			raygui.SetStyle(raygui.LABEL, raygui.TEXT_ALIGNMENT, prevAlign)
		}

		// --- Draw TextBox ---
		// Sync buffer with target when not editing, or right as edit begins
		if !*editMode || (*editMode && !wasEditing) {
			buffer = *target
		}

		// Draw the main TextBox
		if raygui.TextBox(rect, &buffer, 256, *editMode) {
			*editMode = !*editMode
			if config.onEdit != nil {
				config.onEdit(*editMode)
			}
			// Commit to target on edit complete
			if !*editMode && *target != buffer {
				*target = buffer
				if config.onChange != nil {
					config.onChange()
				}
			}
		}
		wasEditing = *editMode

		// Helper closure for Clipboard/Drag-Drop to update state safely
		applyText := func(newText string) {
			buffer = newText
			if !*editMode {
				*target = buffer
				if config.onChange != nil {
					config.onChange()
				}
			}
		}

		// --- Draw Icons ---
		iconOffsetX := rect.X + rect.Width + 4

		// Clipboard Paste Icon
		if config.acceptClipboard {
			pasteRect := rl.NewRectangle(iconOffsetX, rect.Y, rect.Height, rect.Height)
			iconOffsetX += rect.Height + 4 // Advance offset for the next icon

			if raygui.Button(pasteRect, "#17#") {
				if clip := rl.GetClipboardText(); clip != "" {
					applyText(clip)
				}
			}
		}

		// Drag and Drop Icon
		if config.acceptDragDrop {
			dropRect := rl.NewRectangle(iconOffsetX, rect.Y, rect.Height, rect.Height)
			raygui.Button(dropRect, "#5#") // Acts purely as a visual indicator

			if rl.IsFileDropped() {
				files := rl.LoadDroppedFiles()
				mPos := rl.GetMousePosition()

				// Check if dropped on the textbox OR the drop icon
				if len(files) > 0 && (rl.CheckCollisionPointRec(mPos, rect) || rl.CheckCollisionPointRec(mPos, dropRect)) {
					applyText(files[0])
				}
			}
		}
	}}
}
