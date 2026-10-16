package assets

import "embed"

//go:embed winres/icon.png
var AppIconData []byte

//go:embed styles/*.rgs
var StylesFS embed.FS

//go:embed fonts/*.ttf
var FontsFS embed.FS

//go:embed sounds/*/*.ogg
var SoundsFS embed.FS

//go:embed shaders/*.fs
//go:embed shaders/*/*.fs
//go:embed shaders/*/*.vs
var ShadersFS embed.FS

//go:embed graphics/loader/EDLoader_1-30.png
//go:embed graphics/inara/inara.png
//go:embed graphics/inara/powers/*.png
//go:embed graphics/inara/superpowers/*.png
//go:embed graphics/inara/ranks/*/*.png
//go:embed graphics/inara/avatars/*.png
//go:embed graphics/mascots/*/*.png
var GraphicsFS embed.FS

//go:embed tracks/*
var TracksFS embed.FS
