package app

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/assets"
	"github.com/gethiox/EDCITR/internal/app/journal"
	"go.uber.org/zap"
)

func (a *App) unloadSound(ch journal.MessageChannel) {
	for i := 0; i < MaxSounds; i++ {
		rl.UnloadSoundAlias(a.Render.s.LoadedSounds[ch][i])
	}
	rl.UnloadSound(a.Render.s.LoadedSoundsSource[ch])
}

func (a *App) setVolume(ch journal.MessageChannel, volume float32) {
	for i := 0; i < MaxSounds; i++ {
		rl.SetSoundVolume(a.Render.s.LoadedSounds[ch][i], volume)
	}
}

func (a *App) loadSound(ch journal.MessageChannel, path string) error {
	var raw []byte
	var err error

	if strings.HasPrefix(path, "builtin:") {
		path, _ = strings.CutPrefix(path, "builtin:")
		raw, err = assets.SoundsFS.ReadFile(path)
		if err != nil {
			Logger.Warn("Failed to load sound file", zap.Error(err), zap.String("fs", "embed"), zap.String("path", path))
			return err
		}
	} else {
		raw, err = os.ReadFile(path)
		if err != nil {
			Logger.Warn("Failed to load sound file", zap.Error(err), zap.String("fs", "os"), zap.String("path", path))
			return err
		}
	}
	base := filepath.Base(path)
	parts := strings.Split(base, ".")
	ext := parts[len(parts)-1]

	wave := rl.LoadWaveFromMemory("."+ext, raw, int32(len(raw)))
	a.Render.s.LoadedSoundsSource[ch] = rl.LoadSoundFromWave(wave)
	rl.UnloadWave(wave)

	for i := 0; i < MaxSounds; i++ {
		a.Render.s.LoadedSounds[ch][i] = rl.LoadSoundAlias(a.Render.s.LoadedSoundsSource[ch])
	}
	return nil
}

func (a *App) loadSounds() {
	fsRaw, err := assets.SoundsFS.ReadFile("sounds/pony/Yay.ogg")
	if err != nil {
		panic(err)
	}

	wave := rl.LoadWaveFromMemory(".ogg", fsRaw, int32(len(fsRaw)))
	a.Render.s.FSSound = rl.LoadSoundFromWave(wave)
	rl.UnloadWave(wave)

	gopherRaw, err := assets.SoundsFS.ReadFile("sounds/random/Gnome.ogg")
	if err != nil {
		panic(err)
	}

	wave = rl.LoadWaveFromMemory(".ogg", gopherRaw, int32(len(gopherRaw)))
	a.Render.s.GopherSound = rl.LoadSoundFromWave(wave)
	rl.UnloadWave(wave)

	rl.SetMasterVolume(a.Config.Audio.MasterVolume)

	a.Render.s.LoadedSounds = make(map[journal.MessageChannel][]rl.Sound, len(journal.ChannelAllList))
	a.Render.s.LoadedSoundsSource = make(map[journal.MessageChannel]rl.Sound, len(journal.ChannelAllList))

	for _, ch := range journal.ChannelAllList {
		path := a.Config.Audio.ChannelSoundPaths[ch]

		a.Render.s.LoadedSounds[ch] = make([]rl.Sound, MaxSounds)

		if path == "" {
			continue
		}

		_ = a.loadSound(ch, path)
	}
}

func (a *App) loadSoundList() {
	a.Render.s.BuiltInSoundList = make([][2]string, 0)
	sounds, err := assets.SoundsFS.ReadDir("sounds")
	if err != nil {
		panic(err)
	}
	for _, entry := range sounds {
		if !entry.IsDir() {
			panic("wtf")
		}

		soundsCategoryName := entry.Name()

		if !a.Config.IsPonyEnjoyer && soundsCategoryName == "pony" {
			continue
		}

		soundsCategory, err := assets.SoundsFS.ReadDir(path.Join("sounds", soundsCategoryName))
		if err != nil {
			panic(err)
		}

		for _, category := range soundsCategory {
			if category.IsDir() {
				panic("wtf")
			}

			soundName := path.Join(soundsCategoryName, category.Name())
			soundPath := path.Join("sounds", soundsCategoryName, category.Name())

			a.Render.s.BuiltInSoundList = append(a.Render.s.BuiltInSoundList, [2]string{soundName, "builtin:" + soundPath})
		}
	}

}
