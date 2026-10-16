// It's a noire package wrapper
// the main purpose of this package is to define compact JSON representation

package my_noire

import (
	"fmt"
	"image/color"
	"strconv"

	"github.com/teacat/noire"
)

type Color struct {
	noire.Color
}

func FromRGBA(c color.RGBA) Color {
	return Color{noire.Color{
		Red:   float64(c.R),
		Green: float64(c.G),
		Blue:  float64(c.B),
		Alpha: float64(c.A),
	}}
}

func (c Color) AsRGBA() color.RGBA {
	return color.RGBA{
		R: uint8(c.Red),
		G: uint8(c.Green),
		B: uint8(c.Blue),
		A: uint8(c.Alpha * 255),
	}
}

//goland:noinspection GoMixedReceiverTypes
func (c *Color) UnmarshalJSON(bytes []byte) error {
	s := string(bytes)

	if len(s) != len(`"#4080ff7f"`) {
		return fmt.Errorf("invalid color data: %s", s)
	}
	rs, gs, bs, as := s[2:4], s[4:6], s[6:8], s[8:10]

	r, err := strconv.ParseUint(rs, 16, 8)
	if err != nil {
		return err
	}
	g, err := strconv.ParseUint(gs, 16, 8)
	if err != nil {
		return err
	}
	b, err := strconv.ParseUint(bs, 16, 8)
	if err != nil {
		return err
	}
	a, err := strconv.ParseUint(as, 16, 8)
	if err != nil {
		return err
	}

	*c = NewRGBA(float64(r), float64(g), float64(b), float64(a)/255)

	return nil
}

//goland:noinspection GoMixedReceiverTypes
func (c *Color) MarshalJSON() ([]byte, error) {
	rgba := c.AsRGBA()
	return []byte(fmt.Sprintf(`"#%02x%02x%02x%02x"`, rgba.R, rgba.G, rgba.B, rgba.A)), nil
}

func CMYKToRGB(c, m, y, k float64) (r, g, b float64) { return noire.CMYKToRGB(c, m, y, k) }
func RGBToCMYK(r, g, b float64) (c, m, y, k float64) { return noire.RGBToCMYK(r, g, b) }
func HueToRGB(p, q, t float64) float64               { return noire.HueToRGB(p, q, t) }
func RGBToHSL(r, g, b float64) (h, s, l float64)     { return noire.RGBToHSL(r, g, b) }
func HSLToRGB(h, s, l float64) (r, g, b float64)     { return noire.HSLToRGB(h, s, l) }
func HSVToRGB(h, s, v float64) (r, g, b float64)     { return noire.HSVToRGB(h, s, v) }
func RGBToHSV(r, g, b float64) (h, s, v float64)     { return noire.RGBToHSV(r, g, b) }
func RGBToHex(r, g, b float64) string                { return noire.RGBToHex(r, g, b) }
func HexToRGB(h string) (r, g, b float64)            { return noire.HexToRGB(h) }
func HTMLToRGB(h string) (r, g, b float64)           { return noire.HTMLToRGB(h) }
func RGBToHTML(r, g, b float64) (h string)           { return noire.RGBToHTML(r, g, b) }

func NewHTML(color string) Color             { return Color{noire.NewHTML(color)} }
func NewHTMLA(color string, a float64) Color { return Color{noire.NewHTMLA(color, a)} }
func NewHex(color string) Color              { return Color{noire.NewHex(color)} }
func NewHexA(color string, a float64) Color  { return Color{noire.NewHexA(color, a)} }
func NewHSL(h, s, l float64) Color           { return Color{noire.NewHSL(h, s, l)} }
func NewHSLA(h, s, l, a float64) Color       { return Color{noire.NewHSLA(h, s, l, a)} }
func NewHSV(h, s, v float64) Color           { return Color{noire.NewHSV(h, s, v)} }
func NewHSVA(h, s, v, a float64) Color       { return Color{noire.NewHSVA(h, s, v, a)} }
func NewRGB(r, g, b float64) Color           { return Color{noire.NewRGB(r, g, b)} }
func NewRGBA(r, g, b, a float64) Color       { return Color{noire.NewRGBA(r, g, b, a)} }
func NewCMYK(c, m, y, k float64) Color       { return Color{noire.NewCMYK(c, m, y, k)} }
func NewCMYKA(c, m, y, k, a float64) Color   { return Color{noire.NewCMYKA(c, m, y, k, a)} }

func (c Color) Mix(color Color, weight float64) Color      { return Color{c.Color.Mix(color.Color, weight)} }
func (c Color) Hue() float64                               { return c.Color.Hue() }
func (c Color) Saturation() float64                        { return c.Color.Saturation() }
func (c Color) Lightness() float64                         { return c.Color.Lightness() }
func (c Color) AdjustHue(degrees float64) Color            { return Color{c.Color.AdjustHue(degrees)} }
func (c Color) Lighten(percent float64) Color              { return Color{c.Color.Lighten(percent)} }
func (c Color) Darken(percent float64) Color               { return Color{c.Color.Darken(percent)} }
func (c Color) Saturate(percent float64) Color             { return Color{c.Color.Saturate(percent)} }
func (c Color) Desaturate(percent float64) Color           { return Color{c.Color.Desaturate(percent)} }
func (c Color) Grayscale() Color                           { return Color{c.Color.Grayscale()} }
func (c Color) Complement() Color                          { return Color{c.Color.Complement()} }
func (c Color) Tint(percent float64) Color                 { return Color{c.Color.Tint(percent)} }
func (c Color) Shade(percent float64) Color                { return Color{c.Color.Shade(percent)} }
func (c Color) Invert() Color                              { return Color{c.Color.Invert()} }
func (c Color) LuminanaceWCAG() float64                    { return c.Color.LuminanaceWCAG() }
func (c Color) Luminanace() float64                        { return c.Color.Luminanace() }
func (c Color) Foreground() Color                          { return Color{c.Color.Foreground()} }
func (c Color) Brighten(percent float64) Color             { return Color{c.Color.Brighten(percent)} }
func (c Color) Contrast(color Color) float64               { return c.Color.Contrast(color.Color) }
func (c Color) IsLight() bool                              { return c.Color.IsLight() }
func (c Color) IsDark() bool                               { return c.Color.IsDark() }
func (c Color) HSV() (float64, float64, float64)           { return c.Color.HSV() }
func (c Color) HSVA() (float64, float64, float64, float64) { return c.Color.HSVA() }
func (c Color) HSL() (float64, float64, float64)           { return c.Color.HSL() }
func (c Color) HSLA() (float64, float64, float64, float64) { return c.Color.HSLA() }
func (c Color) RGB() (float64, float64, float64)           { return c.Color.RGB() }
func (c Color) RGBA() (float64, float64, float64, float64) { return c.Color.RGBA() }
func (c Color) CMYK() (float64, float64, float64, float64) { return c.Color.CMYK() }
func (c Color) Hex() string                                { return c.Color.Hex() }
func (c Color) HTML() string                               { return c.Color.HTML() }
