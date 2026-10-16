package app

func (r *Renderer) renderTabParser(startPosY float32) {
	for row, widget := range r.settingsParser {
		switch widget.(type) {
		case OptionWidget:
			widget.Draw(r.optionRect(startPosY, row))
		case ButtonWidget:
			widget.Draw(r.buttonRect(startPosY, row))
		}
	}
}
