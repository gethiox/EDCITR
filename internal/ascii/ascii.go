package ascii

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	path2 "path"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

func init() {
	// Force Lipgloss to output TrueColor ANSI codes even if not in a TTY
	lipgloss.SetColorProfile(termenv.TrueColor)
}

var (
	ScalerNearestNeighbor draw.Scaler = draw.NearestNeighbor
	ScalerApproxBiLinear  draw.Scaler = draw.ApproxBiLinear
	ScalerBiLinear        draw.Scaler = draw.BiLinear
	ScalerCatmullRom      draw.Scaler = draw.CatmullRom
)

type RenderType int

const (
	RT4PxPerCell RenderType = iota
	RT2PxPerCell
	RT1PxPerCell
	RT1PxPer2Cells
)

type Decoder func(io.Reader) (image.Image, error)

func RenderImageFromFS(fs embed.FS, path string, scalar draw.Scaler, renderType RenderType, width, height int) ([]string, error) {
	//switch {
	//case term.IsTerminal(0):
	//	var err error
	//	width, height, err = term.GetSize(0)
	//	if err != nil {
	//		panic(err)
	//	}
	//}

	var decoder Decoder

	switch strings.ToLower(path2.Ext(path)) {
	case ".jpeg", ".jpg", ".jfif", ".pjpeg", ".pjp":
		decoder = jpeg.Decode
	case ".png":
		decoder = png.Decode
	case ".tiff":
		decoder = tiff.Decode
	case ".gif", ".gifv":
		decoder = gif.Decode
	case ".webp":
		decoder = webp.Decode
	case ".bmp":
		decoder = bmp.Decode
	default:
		return nil, fmt.Errorf("unsupported file type: %s", strings.ToLower(path2.Ext(path)))
	}

	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	img, err := decoder(f)
	if err != nil {
		return nil, err
	}

	return RenderImage(img, scalar, renderType, width, height)
}

func RenderImage(img image.Image, scalar draw.Scaler, renderType RenderType, width, height int) ([]string, error) {
	switch renderType {
	case RT4PxPerCell:
		return renderImage4PxPerCell(scalar, img, width, height), nil
	case RT2PxPerCell:
		return renderImage2PxPerCell(scalar, img, width, height), nil
	case RT1PxPerCell:
		return renderImage1PxPerCell(scalar, img, width, height), nil
	case RT1PxPer2Cells:
		return renderImage1PxPer2Cells(scalar, img, width, height), nil
	default:
		return nil, fmt.Errorf("unsupported render type: %d", renderType)
	}
}

func uint32ToUint8(x uint32) uint8 {
	return uint8(x)
	//return uint8(x >> 24)
	//return uint8(x / math.MaxUint32 * math.MaxUint8)
}

const (
	pixLeftUp    = '▘'
	pixRightUp   = '▝'
	pixLeftDown  = '▖'
	pixRightDown = '▗'
	pixLeft      = '▌'
	pixDiagonal  = '▚'
	pixUp        = '▀'
	pixFull      = '█'
	pixEmpty     = ' '
)

// colorDistance calculates the squared Euclidean distance in RGBA space.
// We use squared distance to avoid costly square root operations, as we only
// need to compare relative distances. It uses int64 to prevent overflow before squaring.
func colorDistance(c1, c2 color.RGBA) uint64 {
	dR := int64(c1.R) - int64(c2.R)
	dG := int64(c1.G) - int64(c2.G)
	dB := int64(c1.B) - int64(c2.B)
	dA := int64(c1.A) - int64(c2.A)
	return uint64(dR*dR + dG*dG + dB*dB + dA*dA)
}

// averageColor calculates the mean color from a slice of colors.
func averageColor(colors []color.RGBA) color.RGBA {
	if len(colors) == 0 {
		// Return black for an empty slice to avoid division by zero.
		return color.RGBA{}
	}
	var r, g, b, a uint64
	for _, c := range colors {
		r += uint64(c.R)
		g += uint64(c.G)
		b += uint64(c.B)
		a += uint64(c.A)
	}
	n := uint64(len(colors))
	return color.RGBA{
		R: uint8(r / n),
		G: uint8(g / n),
		B: uint8(b / n),
		A: uint8(a / n),
	}
}

// evaluateColors determines the best character, foreground, and background colors
// to represent a 2x2 pixel grid.
func evaluateColors(c00, c10, c01, c11 color.RGBA) (r rune, fg, bg color.RGBA) {
	colors := []color.RGBA{c00, c10, c01, c11}

	// --- Step 1: Cluster the 4 colors into 2 representative colors (fg and bg) ---
	// We use a simple K-Means clustering approach.

	// Initialization: Find the two most distant colors to use as initial centroids.
	// This provides a stable starting point for the algorithm.
	var maxDist uint64
	var centroidA, centroidB color.RGBA

	if len(colors) <= 1 { // Handle case with one color
		return pixFull, colors[0], colors[0]
	}

	for i := 0; i < len(colors); i++ {
		for j := i + 1; j < len(colors); j++ {
			dist := colorDistance(colors[i], colors[j])
			if dist > maxDist {
				maxDist = dist
				centroidA = colors[i]
				centroidB = colors[j]
			}
		}
	}

	// Edge case: If all four colors are identical, maxDist will be 0.
	if maxDist == 0 {
		return pixFull, colors[0], colors[0]
	}

	// Iteration: Refine the centroids. 5 iterations are more than enough for 4 points.
	for i := 0; i < 5; i++ {
		clusterA := []color.RGBA{}
		clusterB := []color.RGBA{}

		// Assign each color to the nearest centroid.
		for _, c := range colors {
			if colorDistance(c, centroidA) <= colorDistance(c, centroidB) {
				clusterA = append(clusterA, c)
			} else {
				clusterB = append(clusterB, c)
			}
		}

		// If a cluster becomes empty, it means the other cluster has captured all points.
		// We can stop early as further iterations won't change the result.
		if len(clusterA) == 0 || len(clusterB) == 0 {
			break
		}

		// Recalculate centroids by averaging the colors in each cluster.
		newCentroidA := averageColor(clusterA)
		newCentroidB := averageColor(clusterB)

		// If centroids have stabilized, we can stop.
		if newCentroidA == centroidA && newCentroidB == centroidB {
			break
		}
		centroidA = newCentroidA
		centroidB = newCentroidB
	}

	fg = centroidA
	bg = centroidB

	// --- Step 2: Create a binary pattern based on the clusters ---
	// We create a 4-bit integer where each bit represents a pixel.
	// Bit is 1 if the pixel is closer to the foreground color (centroidA), 0 otherwise.
	// Bit order: TopLeft, TopRight, BottomLeft, BottomRight
	pattern := 0
	if colorDistance(c00, fg) <= colorDistance(c00, bg) {
		pattern |= 1 << 3
	} // 1000
	if colorDistance(c10, fg) <= colorDistance(c10, bg) {
		pattern |= 1 << 2
	} // 0100
	if colorDistance(c01, fg) <= colorDistance(c01, bg) {
		pattern |= 1 << 1
	} // 0010
	if colorDistance(c11, fg) <= colorDistance(c11, bg) {
		pattern |= 1 << 0
	} // 0001

	// --- Step 3: Match the pattern to a character ---
	// We check the pattern directly. If no match is found, we check its inverse.
	// If the inverse matches, we use the same character but swap fg and bg.

	switch pattern {
	// 0 or 4 pixels are foreground
	case 0b0000:
		return pixEmpty, bg, fg
	case 0b1111:
		return pixFull, fg, bg

	// 1 pixel is foreground
	case 0b1000:
		return pixLeftUp, fg, bg
	case 0b0100:
		return pixRightUp, fg, bg
	case 0b0010:
		return pixLeftDown, fg, bg
	case 0b0001:
		return pixRightDown, fg, bg

	// 3 pixels are foreground (inverse of 1-pixel cases)
	case 0b0111:
		return pixLeftUp, bg, fg // Inverse of 1000
	case 0b1011:
		return pixRightUp, bg, fg // Inverse of 0100
	case 0b1101:
		return pixLeftDown, bg, fg // Inverse of 0010
	case 0b1110:
		return pixRightDown, bg, fg // Inverse of 0001

	// 2 pixels are foreground
	case 0b1100:
		return pixUp, fg, bg
	case 0b0011:
		return pixUp, bg, fg // Down is inverse of Up

	case 0b1010:
		return pixLeft, fg, bg
	case 0b0101:
		return pixLeft, bg, fg // Right is inverse of Left

	case 0b1001:
		return pixDiagonal, fg, bg
	case 0b0110:
		return pixDiagonal, bg, fg // Inverse diagonal
	}

	// This return should ideally not be reached, but is a safe fallback.
	return '?', fg, bg
}

func rgbaToLipgloss(color color.RGBA) lipgloss.Color {
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", color.R, color.G, color.B))
}

func renderImage4PxPerCell(scalar draw.Scaler, src image.Image, maxX, maxY int) []string {
	scaled := image.NewRGBA(image.Rect(0, 0, maxX*2, maxY*2))
	scalar.Scale(scaled, scaled.Rect, src, src.Bounds(), draw.Over, nil)

	type result struct {
		y   int
		row string
	}
	var results = make(chan result)

	wg := sync.WaitGroup{}
	for y := 0; y < maxY*2; y += 2 {
		row := ""
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			for x := 0; x < maxX*2; x += 2 {
				r00, g00, b00, _ := scaled.At(x+0, y+0).RGBA()
				r10, g10, b10, _ := scaled.At(x+1, y+0).RGBA()
				r01, g01, b01, _ := scaled.At(x+0, y+1).RGBA()
				r11, g11, b11, _ := scaled.At(x+1, y+1).RGBA()

				c00 := color.RGBA{uint32ToUint8(r00), uint32ToUint8(g00), uint32ToUint8(b00), 0}
				c10 := color.RGBA{uint32ToUint8(r10), uint32ToUint8(g10), uint32ToUint8(b10), 0}
				c01 := color.RGBA{uint32ToUint8(r01), uint32ToUint8(g01), uint32ToUint8(b01), 0}
				c11 := color.RGBA{uint32ToUint8(r11), uint32ToUint8(g11), uint32ToUint8(b11), 0}

				r, fg, bg := evaluateColors(c00, c10, c01, c11)

				style := lipgloss.NewStyle().
					Foreground(rgbaToLipgloss(fg)).
					Background(rgbaToLipgloss(bg))
				row += style.Render(string(r))
			}
			results <- result{y: y / 2, row: row}
		}(y)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var ascii = make([]string, maxY)
	for result := range results {
		ascii[result.y] = result.row
	}

	return ascii
}

func renderImage2PxPerCell(scalar draw.Scaler, src image.Image, maxX, maxY int) []string {
	scaled := image.NewRGBA(image.Rect(0, 0, maxX, maxY*2))
	scalar.Scale(scaled, scaled.Rect, src, src.Bounds(), draw.Over, nil)
	var ascii = make([]string, 0, maxY)

	for y := 0; y < maxY*2; y += 2 {
		row := ""
		for x := 0; x < maxX; x++ {
			r1, g1, b1, _ := scaled.At(x, y+0).RGBA()
			r2, g2, b2, _ := scaled.At(x, y+1).RGBA()
			c1 := color.RGBA{uint32ToUint8(r1), uint32ToUint8(g1), uint32ToUint8(b1), 0}
			c2 := color.RGBA{uint32ToUint8(r2), uint32ToUint8(g2), uint32ToUint8(b2), 0}

			style := lipgloss.NewStyle().
				Foreground(rgbaToLipgloss(c1)).
				Background(rgbaToLipgloss(c2))
			row += style.Render(string(pixUp))
		}
		ascii = append(ascii, row)
	}
	return ascii
}

func renderImage1PxPerCell(scalar draw.Scaler, src image.Image, maxX, maxY int) []string {
	scaled := image.NewRGBA(image.Rect(0, 0, maxX, maxY))
	scalar.Scale(scaled, scaled.Rect, src, src.Bounds(), draw.Over, nil)
	var ascii = make([]string, 0, maxY)

	for y := 0; y < maxY; y++ {
		row := ""
		for x := 0; x < maxX; x++ {
			r, g, b, _ := scaled.At(x, y).RGBA()
			c := color.RGBA{uint32ToUint8(r), uint32ToUint8(g), uint32ToUint8(b), 0}

			style := lipgloss.NewStyle().
				Foreground(rgbaToLipgloss(c))
			row += style.Render(string(pixFull))
		}
		ascii = append(ascii, row)
	}
	return ascii
}

func renderImage1PxPer2Cells(scalar draw.Scaler, src image.Image, maxX, maxY int) []string {
	scaled := image.NewRGBA(image.Rect(0, 0, maxX/2, maxY))
	scalar.Scale(scaled, scaled.Rect, src, src.Bounds(), draw.Over, nil)
	var ascii = make([]string, 0, maxY)

	for y := 0; y < maxY; y += 1 {
		row := ""
		for x := 0; x < maxX/2; x++ {
			r, g, b, _ := scaled.At(x, y).RGBA()
			c := color.RGBA{uint32ToUint8(r), uint32ToUint8(g), uint32ToUint8(b), 0}

			style := lipgloss.NewStyle().
				Foreground(rgbaToLipgloss(c))
			row += style.Render(string(pixFull) + string(pixFull))
		}
		ascii = append(ascii, row)
	}
	return ascii
}
