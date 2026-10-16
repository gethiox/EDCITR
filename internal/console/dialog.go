package console

import "github.com/ncruces/zenity"

func ShowCrashDialog(title, msg string) {
	_ = zenity.Error(msg, zenity.Title(title), zenity.Pulsate())
}
