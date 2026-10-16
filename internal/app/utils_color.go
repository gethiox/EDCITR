package app

import (
	"hash/fnv"

	"github.com/gethiox/EDCITR/internal/noire"
	"github.com/lucasb-eyer/go-colorful"
)

var keypoints = GradientTable{
	{MustParseHex("#9e0142"), 0.0},
	{MustParseHex("#d53e4f"), 0.1},
	{MustParseHex("#f46d43"), 0.2},
	{MustParseHex("#fdae61"), 0.3},
	{MustParseHex("#fee090"), 0.4},
	{MustParseHex("#ffffbf"), 0.5},
	{MustParseHex("#e6f598"), 0.6},
	{MustParseHex("#abdda4"), 0.7},
	{MustParseHex("#66c2a5"), 0.8},
	{MustParseHex("#3288bd"), 0.9},
	{MustParseHex("#5e4fa2"), 1.0},
}

func MustParseHex(s string) colorful.Color {
	c, err := colorful.Hex(s)
	if err != nil {
		panic("MustParseHex: " + err.Error())
	}
	return c
}

type GradientTable []struct {
	Col colorful.Color
	Pos float64
}

func (gt GradientTable) GetInterpolatedColorFor(t float64) colorful.Color {
	for i := 0; i < len(gt)-1; i++ {
		c1 := gt[i]
		c2 := gt[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			t := (t - c1.Pos) / (c2.Pos - c1.Pos)
			return c1.Col.BlendHcl(c2.Col, t).Clamped()
		}
	}
	return gt[len(gt)-1].Col
}

func GetGradientColor(v float64) colorful.Color {
	return keypoints.GetInterpolatedColorFor(v)
}

func hashInput(input string, seed int64) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(input))

	_, _ = h.Write([]byte{
		byte(seed), byte(seed >> 8), byte(seed >> 16), byte(seed >> 24),
		byte(seed >> 32), byte(seed >> 40), byte(seed >> 48), byte(seed >> 56),
	})

	return h.Sum64()
}

func GenerateSkinTone(input string, seed int64) my_noire.Color {
	h := hashInput("habba"+input, seed)

	r1 := float64(h&0xFFFF) / 65535.0
	r2 := float64((h>>16)&0xFFFF) / 65535.0
	r3 := float64((h>>32)&0xFFFF) / 65535.0

	hue := 5.0 + (r1 * 25.0)
	sat := 0.1 + (r2 * 0.2)
	val := 0.4 + (r3 * 0.6)

	return my_noire.NewHSV(hue, sat*100, val*100)
}

func GenerateSuitColor(input string, seed int64) my_noire.Color {
	h := hashInput("babba"+input, seed)

	r1 := float64(h&0xFFFF) / 65535.0
	r2 := float64((h>>16)&0xFFFF) / 65535.0
	r3 := float64((h>>32)&0xFFFF) / 65535.0

	hue := r1 * 360.0
	sat := 0.4 + (r2 * 0.2)
	val := 0.6 + (r3 * 0.3)

	return my_noire.NewHSV(hue, sat*100, val*100)
}
