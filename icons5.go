package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 5.0: pictograme noi (lacat, costum, arma, GPS, teleport, telefon,
// coroana, piesa aurie, inchide) + desene mari pentru cabina.
// ====================================================================

func drawIcon5(kind IconKind, cx, cy, s float32, col rl.Color) {
	h := s / 2
	a := col.A
	if kind == IconTrain || kind == IconWorld || kind == IconWings {
		drawIcon6(kind, cx, cy, s, col)
		return
	}
	switch kind {
	case IconLock:
		rl.DrawRing(vv(cx, cy-h*0.25), h*0.32, h*0.5, 180, 360, 20, col)
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.65, cy-h*0.25, h*1.3, h*1.05), 0.25, 6, fade(LegoGoldBright, a))
		rl.DrawCircleV(vv(cx, cy+h*0.18), h*0.16, fade(icoDark, a))
		rl.DrawRectangle(int32(cx-h*0.06), int32(cy+h*0.18), int32(maxf(h*0.12, 2)), int32(h*0.35), fade(icoDark, a))
	case IconOutfit:
		// tricou
		body := fade(LegoRed, a)
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.5, cy-h*0.55, h, h*1.45), 0.2, 6, body)
		tri(vv(cx-h*0.5, cy-h*0.55), vv(cx-h*1.0, cy-h*0.05), vv(cx-h*0.5, cy+h*0.15), body)
		tri(vv(cx+h*0.5, cy-h*0.55), vv(cx+h*0.5, cy+h*0.15), vv(cx+h*1.0, cy-h*0.05), body)
		rl.DrawCircleV(vv(cx, cy-h*0.55), h*0.22, fade(icoDark, a))
		drawStar(cx, cy+h*0.2, h*0.3, fade(LegoGoldBright, a))
	case IconGun:
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.9, cy-h*0.45, s*0.85, h*0.5), 0.4, 6, col)
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.75, cy-h*0.05, h*0.42, h*0.85), 0.4, 6, col)
		rl.DrawCircleV(vv(cx+h*0.85, cy-h*0.2), h*0.2, fade(LegoCyan, a))
		rl.DrawLineEx(vv(cx+h*1.05, cy-h*0.2), vv(cx+h*1.35, cy-h*0.2), h*0.14, fade(LegoCyan, a))
	case IconGPS:
		for k := 0; k < 3; k++ {
			y := cy + h*0.7 - float32(k)*h*0.65
			c := fade(UIGreen, uint8(float32(a)*(0.45+0.27*float32(k))))
			tri(vv(cx, y-h*0.55), vv(cx-h*0.75, y), vv(cx-h*0.4, y), c)
			tri(vv(cx, y-h*0.55), vv(cx+h*0.4, y), vv(cx+h*0.75, y), c)
			tri(vv(cx, y-h*0.55), vv(cx-h*0.4, y), vv(cx, y-h*0.2), c)
			tri(vv(cx, y-h*0.55), vv(cx, y-h*0.2), vv(cx+h*0.4, y), c)
		}
	case IconTeleport:
		for k := 0; k < 3; k++ {
			r := h * (0.35 + 0.3*float32(k))
			rl.DrawRing(vv(cx, cy+h*0.55), r, r+h*0.1, 0, 360, 24, fade(LegoCyan, uint8(float32(a)*(1-0.25*float32(k)))))
		}
		drawStar(cx, cy-h*0.35, h*0.55, fade(LegoGoldBright, a))
		drawStar(cx-h*0.7, cy-h*0.75, h*0.22, fade(LegoWhite, a))
		drawStar(cx+h*0.72, cy-h*0.6, h*0.18, fade(LegoWhite, a))
	case IconPhone:
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.7, cy-h, h*1.4, s), 0.15, 6, fade(LegoRed, a))
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.5, cy-h*0.7, h, h*0.9), 0.1, 4, rl.NewColor(170, 220, 255, a))
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.35, cy+h*0.35, h*0.7, h*0.4), 0.4, 4, fade(icoDark, a))
	case IconCrown:
		c := fade(LegoGoldBright, a)
		base := cy + h*0.5
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.85, base-h*0.35, h*1.7, h*0.45), 0.3, 4, c)
		for k := -1; k <= 1; k++ {
			x := cx + float32(k)*h*0.62
			top := base - h*1.15
			if k == 0 {
				top = base - h*1.4
			}
			tri(vv(x, top), vv(x-h*0.32, base-h*0.3), vv(x+h*0.32, base-h*0.3), c)
			rl.DrawCircleV(vv(x, top), h*0.13, fade(LegoRed, a))
		}
	case IconGoldBrick:
		drawBrick(cx, cy+h*0.1, s*0.95, fade(LegoGoldBright, a))
		drawStar(cx+h*0.75, cy-h*0.75, h*0.25, fade(LegoWhite, a))
	case IconClose:
		rl.DrawCircleV(vv(cx, cy), h, fade(UIRed, a))
		rl.DrawLineEx(vv(cx-h*0.45, cy-h*0.45), vv(cx+h*0.45, cy+h*0.45), h*0.25, fade(LegoWhite, a))
		rl.DrawLineEx(vv(cx-h*0.45, cy+h*0.45), vv(cx+h*0.45, cy-h*0.45), h*0.25, fade(LegoWhite, a))
	}
}

// drawGunPicture: desenul armei k (pentru cabina si HUD), latime s.
func drawGunPicture(k int, cx, cy, s float32, a uint8) {
	h := s / 2
	body := []rl.Color{LegoDarkGrey, LegoRed, LegoOrange, LegoGreen, LegoPurple, LegoPink}[k%NumGuns]
	glow := []rl.Color{LegoCyan, rl.NewColor(255, 80, 80, 255), LegoYellow, LegoRed, rl.NewColor(220, 140, 255, 255), LegoCyan}[k%NumGuns]
	body, glow = fade(body, a), fade(glow, a)
	switch k {
	case 1: // laser rosu: lung si subtire
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h, cy-h*0.16, s*1.0, h*0.3), 0.5, 6, body)
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.8, cy, h*0.3, h*0.6), 0.4, 4, body)
		rl.DrawCircleV(vv(cx+h, cy-h*0.01), h*0.14, glow)
		rl.DrawLineEx(vv(cx+h, cy), vv(cx+h*1.5, cy), h*0.06, glow)
	case 2: // blaster mare
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.9, cy-h*0.45, s*0.8, h*0.75), 0.35, 6, body)
		rl.DrawRectangleRounded(rl.NewRectangle(cx+h*0.55, cy-h*0.55, h*0.45, h*0.95), 0.3, 6, shade(body, 0.8))
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.7, cy+h*0.2, h*0.4, h*0.7), 0.4, 4, body)
		rl.DrawCircleV(vv(cx+h*1.0, cy-h*0.08), h*0.28, glow)
	case 3: // rachete: tub cu racheta
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h, cy-h*0.35, s*0.85, h*0.6), 0.5, 6, body)
		tri(vv(cx+h*1.15, cy-h*0.05), vv(cx+h*0.7, cy-h*0.33), vv(cx+h*0.7, cy+h*0.23), glow)
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.3, cy-h*0.75, h*0.5, h*0.35), 0.4, 4, fade(icoDark, a))
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.55, cy+h*0.2, h*0.35, h*0.6), 0.4, 4, body)
	case 4: // raza continua: bobine + farfurie
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.9, cy-h*0.25, s*0.7, h*0.5), 0.5, 6, body)
		for i := 0; i < 3; i++ {
			x := cx - h*0.5 + float32(i)*h*0.35
			rl.DrawRectangleRounded(rl.NewRectangle(x, cy-h*0.42, h*0.16, h*0.84), 0.5, 4, glow)
		}
		rl.DrawEllipse(int32(cx+h*0.62), int32(cy), h*0.18, h*0.5, glow)
		rl.DrawLineEx(vv(cx+h*0.7, cy), vv(cx+h*1.5, cy), h*0.18, fade(glow, uint8(float32(a)*0.7)))
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.75, cy+h*0.15, h*0.32, h*0.6), 0.4, 4, body)
	case 5: // tun cu bule
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.85, cy-h*0.2, s*0.7, h*0.5), 0.5, 6, body)
		tri(vv(cx+h*0.5, cy+h*0.05), vv(cx+h*1.05, cy-h*0.45), vv(cx+h*1.05, cy+h*0.55), body)
		rl.DrawCircleV(vv(cx-h*0.3, cy-h*0.45), h*0.35, rl.NewColor(180, 230, 255, uint8(float32(a)*0.8)))
		for i, c := range []rl.Color{LegoCyan, LegoPink, LegoYellow} {
			rl.DrawCircleLines(int32(cx+h*1.25+float32(i)*h*0.18), int32(cy-h*0.4+float32(i)*h*0.35), h*0.16, fade(c, a))
		}
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.7, cy+h*0.2, h*0.32, h*0.6), 0.4, 4, body)
	default: // pistol laser albastru
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.8, cy-h*0.3, s*0.75, h*0.5), 0.4, 6, body)
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.65, cy+h*0.1, h*0.38, h*0.7), 0.4, 4, body)
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.5, cy-h*0.42, h*0.7, h*0.15), 0.4, 4, fade(LegoOrange, a))
		rl.DrawCircleV(vv(cx+h*0.75, cy-h*0.05), h*0.18, glow)
	}
}

// drawCloudBlob: nor desenat pe harta (peste insulele inchise).
func drawCloudBlob(cx, cy, r float32, a uint8, t float32) {
	if a == 0 {
		return
	}
	for k := 0; k < 9; k++ {
		ang := float32(k)/9*2*math.Pi + t*0.05
		rr := r * (0.55 + 0.12*float32(k%3))
		x := cx + cosf(ang)*r*0.5
		y := cy + sinf(ang)*r*0.42
		rl.DrawCircleV(vv(x, y+3), rr, rl.NewColor(150, 165, 190, uint8(float32(a)*0.55)))
		rl.DrawCircleV(vv(x, y), rr, rl.NewColor(236, 241, 250, a))
	}
	rl.DrawCircleV(vv(cx, cy), r*0.7, rl.NewColor(246, 248, 252, a))
}

// V 6.0: tren, lume (glob), aripi (abilitatea costumului)
func drawIcon6(kind IconKind, cx, cy, s float32, col rl.Color) {
	h := s / 2
	a := col.A
	switch kind {
	case IconTrain:
		body := fade(LegoRed, a)
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.95, cy-h*0.15, h*1.3, h*0.65), 0.3, 6, body)
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.95, cy-h*0.75, h*0.6, h*0.7), 0.2, 6, fade(LegoBlue, a))
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.85, cy-h*0.62, h*0.38, h*0.3), 0.3, 4, rl.NewColor(170, 220, 255, a))
		rl.DrawRectangleRounded(rl.NewRectangle(cx+h*0.12, cy-h*0.65, h*0.22, h*0.55), 0.3, 4, fade(icoDark, a))
		rl.DrawCircleV(vv(cx+h*0.4, cy-h*0.85), h*0.16, rl.NewColor(240, 240, 245, a))
		rl.DrawCircleV(vv(cx+h*0.6, cy-h*1.0), h*0.12, rl.NewColor(240, 240, 245, a))
		rl.DrawRectangleRounded(rl.NewRectangle(cx+h*0.42, cy-h*0.05, h*0.55, h*0.55), 0.2, 4, fade(LegoYellow, a))
		for _, wx := range []float32{-h * 0.65, -h * 0.05, h * 0.7} {
			rl.DrawCircleV(vv(cx+wx, cy+h*0.62), h*0.22, fade(icoDark, a))
		}
		drawStar(cx+h*0.95, cy-h*0.7, h*0.22, fade(LegoGoldBright, a))
	case IconWorld:
		rl.DrawCircleV(vv(cx, cy), h*0.9, fade(LegoCyan, a))
		rl.DrawCircleV(vv(cx-h*0.25, cy-h*0.2), h*0.38, fade(LegoGreen, a))
		rl.DrawCircleV(vv(cx+h*0.35, cy+h*0.3), h*0.28, fade(LegoGreen, a))
		rl.DrawRing(vv(cx, cy), h*0.9, h*1.0, 0, 360, 32, fade(icoDark, a))
		drawStar(cx+h*0.8, cy-h*0.8, h*0.28, fade(LegoGoldBright, a))
	case IconWings:
		c := fade(LegoWhite, a)
		for _, sd := range []float32{-1, 1} {
			tri(vv(cx, cy), vv(cx+sd*h*1.0, cy-h*0.7), vv(cx+sd*h*0.9, cy+h*0.1), c)
			tri(vv(cx, cy), vv(cx+sd*h*0.9, cy+h*0.1), vv(cx+sd*h*0.5, cy+h*0.5), c)
		}
		rl.DrawCircleV(vv(cx, cy), h*0.25, fade(LegoGoldBright, a))
	}
	_ = math.Pi
}
