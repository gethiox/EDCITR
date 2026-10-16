package my_noire

import (
	"github.com/lucasb-eyer/go-colorful"
	"github.com/teacat/noire"
)

func NewHCL(h, c, l float64) Color {
	hcl := colorful.Hcl(h, c/100, l/100).Clamped()
	return Color{
		noire.Color{
			Red:   hcl.R * 255,
			Green: hcl.G * 255,
			Blue:  hcl.B * 255,
			Alpha: 1.0,
		},
	}

}
