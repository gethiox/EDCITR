package app

func (r *Renderer) renderTabNetwork(startPosY float32) {
	for row, widget := range r.settingsNetwork {
		switch widget.(type) {
		case OptionWidget:
			widget.Draw(r.optionRect(startPosY, row))
		case ButtonWidget:
			widget.Draw(r.buttonRect(startPosY, row))
		}
	}
}
