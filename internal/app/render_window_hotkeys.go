package app

import rl "github.com/gen2brain/raylib-go/raylib"

func (r *Renderer) handleKeyboardShortcuts() (shouldExit bool) {
	if r.s.IsEditing {
		return false
	}

	for _, action := range r.keyActions {
		if !rl.IsKeyPressed(action.key) {
			continue
		}
		action.action()
	}

	return rl.IsKeyPressed(rl.KeyQ)
}

func (r *Renderer) toggleSettingsView() {
	r.s.OptionViewActive = !r.s.OptionViewActive
}

func (r *Renderer) toggleMaximize() {
	if !r.s.IsMaximized {
		rl.RestoreWindow()
		rl.SetWindowState(rl.FlagWindowUndecorated)
		rl.MaximizeWindow()
		r.s.IsMaximized = true
	} else {
		rl.RestoreWindow()
		// double call hack to bring back window decoration (title bar)
		rl.ToggleBorderlessWindowed()
		rl.ToggleBorderlessWindowed()
		r.s.IsMaximized = false
	}
}
