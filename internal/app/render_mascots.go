package app

import (
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gethiox/EDCITR/internal/utils"
)

var (
	GopherLeftEyeCenter  = rl.Vector2{131, 84}
	GopherRightEyeCenter = rl.Vector2{262, 79}
	FSLeftEyeCenter      = rl.Vector2{206, 171}
	FSRightEyeCenter     = rl.Vector2{308, 171}
)

func (r *Renderer) renderMascot() {
	if r.app.Config.IsPonyEnjoyer {
		if r.drawFS() {
			r.app.Config.IsPonyEnjoyer = false
			r.s.SoundListReloadRequired = true
			rl.PlaySound(r.s.GopherSound)
		}
	} else {
		if r.drawGopher() {
			r.app.Config.IsPonyEnjoyer = true
			r.s.SoundListReloadRequired = true
			rl.PlaySound(r.s.FSSound)
		}
	}
}

func (r *Renderer) getSmoothedMouseTarget(mascotPos, leftEye, rightEye rl.Vector2) rl.Vector2 {
	mousePosMiddle := rl.Vector2Add(mascotPos, rl.Vector2Scale(rl.Vector2Add(leftEye, rightEye), 0.5))
	target := r.s.MousePosition

	if !r.s.MouseOverWindow {
		since := float32(min(time.Since(r.s.MouseOverWindowDisengaged).Milliseconds(), 400)) / 400.0
		progress := float32(utils.EaseOutExpo(float64(since)))
		return rl.Vector2{
			X: rl.Remap(progress, 0, 1, target.X, mousePosMiddle.X),
			Y: rl.Remap(progress, 0, 1, target.Y, mousePosMiddle.Y),
		}
	}

	since := float32(min(time.Since(r.s.MouseOverWindowEngaged).Milliseconds(), 400)) / 400.0
	progress := float32(utils.EaseInExpo(float64(since)))
	return rl.Vector2{
		X: rl.Remap(progress, 0, 1, mousePosMiddle.X, target.X),
		Y: rl.Remap(progress, 0, 1, mousePosMiddle.Y, target.Y),
	}
}

func (r *Renderer) checkBoop(boopCenter, boopSize rl.Vector2) bool {
	rect := rl.Rectangle{
		X:      boopCenter.X - boopSize.X/2,
		Y:      boopCenter.Y - boopSize.Y/2,
		Width:  boopSize.X,
		Height: boopSize.Y,
	}
	return rl.IsMouseButtonPressed(rl.MouseButtonLeft) && rl.CheckCollisionPointRec(r.s.MousePosition, rect)
}

func (r *Renderer) getEyePosOffset(lookTarget, mascotPos, eyeCenter rl.Vector2, hRange, vRange, texWidth, texHeight, resize float32) rl.Vector2 {
	eyeOrigin := rl.Vector2Add(mascotPos, eyeCenter)
	diff := rl.Vector2Subtract(lookTarget, eyeOrigin)
	norm := rl.Vector2Normalize(diff)

	dist := rl.Vector2Distance(eyeOrigin, lookTarget)
	maxDist := rl.Vector2Distance(rl.Vector2{}, rl.Vector2{float32(r.s.CanvasSizeX), float32(r.s.CanvasSizeY)})
	nDist := float32(utils.EaseOutExpo(utils.EaseOutExpo(float64(dist / maxDist))))

	offset := rl.Vector2Multiply(norm, rl.Vector2{X: hRange * nDist, Y: vRange * nDist})
	absolutePos := rl.Vector2Add(eyeOrigin, offset)

	return rl.Vector2Subtract(absolutePos, rl.Vector2{X: (texWidth / 2) * resize, Y: (texHeight / 2) * resize})
}

func (r *Renderer) drawFS() (boopPerformed bool) {
	pos := rl.Vector2{float32(r.s.CanvasSizeX/2) - float32(r.app.State.Textures.FS.FS.Width/2) - 530, float32((int32(r.s.CanvasSizeY) - r.app.State.Textures.FS.FS.Height) / 2)}
	drawColor := rl.ColorAlpha(rl.White, r.s.GuiCurrentAlphaImage)
	eyeResize := float32(0.5)

	// Body
	rl.DrawTexture(r.app.State.Textures.FS.FS, int32(pos.X), int32(pos.Y), drawColor)

	// Interaction
	lookTarget := r.getSmoothedMouseTarget(pos, FSLeftEyeCenter, FSRightEyeCenter)
	boopPerformed = r.checkBoop(rl.Vector2Add(pos, rl.Vector2{X: 256, Y: 215}), rl.Vector2{X: 50, Y: 40})

	// Iris
	texIris := r.app.State.Textures.FS.FSEyeIris
	w, h := float32(texIris.Width), float32(texIris.Height)

	for i, tex := range []rl.Texture2D{texIris, r.app.State.Textures.FS.FSEyeDot} {
		hRange, vRange := 6.0+(4.0*float32(i)), 12.0+(10.0*float32(i))

		pL := r.getEyePosOffset(lookTarget, pos, FSLeftEyeCenter, hRange, vRange, w, h, eyeResize)
		pR := r.getEyePosOffset(lookTarget, pos, FSRightEyeCenter, hRange, vRange, w, h, eyeResize)

		rl.DrawTextureEx(tex, pL, 0, eyeResize, drawColor)
		rl.DrawTextureEx(tex, pR, 0, eyeResize, drawColor)
	}

	// Highlights (using 0.5 multiplier for ranges)
	baseL := r.getEyePosOffset(lookTarget, pos, FSLeftEyeCenter, 6.0+(4.0*0.5), 12.0+(10.0*0.5), w, h, eyeResize)
	baseR := r.getEyePosOffset(lookTarget, pos, FSRightEyeCenter, 6.0+(4.0*0.5), 12.0+(10.0*0.5), w, h, eyeResize)
	r.drawFSHighlights(baseL, baseR, eyeResize, drawColor)

	return boopPerformed
}

func (r *Renderer) drawFSHighlights(baseL, baseR rl.Vector2, resize float32, color rl.Color) {
	tex := r.app.State.Textures.FS.FSEyeHl
	srcRect := rl.Rectangle{0, 0, float32(tex.Width), float32(tex.Height)}

	drawHighlight := func(base rl.Vector2, scaleX, scaleY, offsetX, offsetY, rotation float32) {
		dest := rl.Rectangle{
			X:      base.X + (float32(tex.Width)*resize)/2 + offsetX,
			Y:      base.Y + (float32(tex.Height)*resize)/2 + offsetY,
			Width:  float32(tex.Width) * scaleX * resize,
			Height: float32(tex.Height) * scaleY * resize,
		}
		origin := rl.Vector2{X: dest.Width / 2, Y: dest.Height / 2}
		rl.DrawTexturePro(tex, srcRect, dest, origin, rotation, color)
	}

	// Big highlights
	drawHighlight(baseL, 0.8, 0.8, 7, -14, -22.5)
	drawHighlight(baseR, 0.8, 0.8, 7, -14, -22.5)

	// Small highlights
	drawHighlight(baseL, 0.3, 0.3, 0, 0, -34.0)
	drawHighlight(baseR, 0.3, 0.3, 0, 0, -34.0)
}

func (r *Renderer) drawGopher() (boopPerformed bool) {
	pos := rl.Vector2{float32(r.s.CanvasSizeX/2) - float32(r.app.State.Textures.Gopher.Gopher.Width/2) - 530, float32((int32(r.s.CanvasSizeY) - r.app.State.Textures.Gopher.Gopher.Height) / 2)}
	drawColor := rl.ColorAlpha(rl.White, r.s.GuiCurrentAlphaImage)

	// Body
	rl.DrawTexture(r.app.State.Textures.Gopher.Gopher, int32(pos.X), int32(pos.Y), drawColor)

	handRect := rl.Rectangle{0, 0, float32(r.app.State.Textures.Gopher.GopherHands.Width), float32(r.app.State.Textures.Gopher.GopherHands.Height)}
	handDest := rl.Rectangle{pos.X + handRect.Width/2, pos.Y + handRect.Height/2, handRect.Width, handRect.Height}
	handOrigin := rl.Vector2{handRect.Width / 2, handRect.Height / 2}
	rl.DrawTexturePro(r.app.State.Textures.Gopher.GopherHands, handRect, handDest, handOrigin, float32(math.Sin(float64(time.Now().UnixMilli())/500)*5), drawColor)

	// Interaction
	lookTarget := r.getSmoothedMouseTarget(pos, GopherLeftEyeCenter, GopherRightEyeCenter)
	boopPerformed = r.checkBoop(rl.Vector2Add(pos, rl.Vector2{X: 202, Y: 128}), rl.Vector2{X: 50, Y: 30})

	// Eyes
	w, h := float32(r.app.State.Textures.Gopher.GopherEye.Width), float32(r.app.State.Textures.Gopher.GopherEye.Height)
	posL := r.getEyePosOffset(lookTarget, pos, GopherLeftEyeCenter, 33, 33, w, h, 1.0)
	posR := r.getEyePosOffset(lookTarget, pos, GopherRightEyeCenter, 33, 33, w, h, 1.0)

	rl.DrawTextureV(r.app.State.Textures.Gopher.GopherEye, posL, drawColor)
	rl.DrawTextureV(r.app.State.Textures.Gopher.GopherEye, posR, drawColor)

	return boopPerformed
}
