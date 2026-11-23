package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// PICTOGRAME: Aventurierul are 5 ani si nu citeste inca. Tot ce e important pe
// ecran se vede ca desen (tasta de apasat, ce face, cat mai e), iar textul
// ramane mic si discret, pentru parintele care ii explica.
// ====================================================================

type IconKind int

const (
	IconNone IconKind = iota
	IconCar
	IconLiftUp
	IconLiftDown
	IconWheel
	IconCoaster
	IconTrampoline
	IconSword
	IconBrick
	IconCoin
	IconNoCoin
	IconTunnel
	IconBridge
	IconIsland
	IconStar
	IconCamera
	IconSave
	IconHeart
	IconExit
	IconMouse
	IconWait
	IconGrow
	IconTrophy
	IconMap
	IconHelp
	IconMonster
	IconCheck
	IconEye
	IconHorn
	IconTurbo
	IconArrowUp
	IconArrowDown
	IconArrowLeft
	IconArrowRight
	IconLock
	IconOutfit
	IconGun
	IconGPS
	IconTeleport
	IconPhone
	IconCrown
	IconGoldBrick
	IconClose
	IconTrain // V 6.0
	IconWorld
	IconWings
)

var (
	icoDark  = rl.NewColor(16, 20, 32, 235)
	icoLight = rl.NewColor(250, 250, 252, 255)
)

// tri: triunghi desenat in ambele sensuri (raylib il ascunde pe cel "intors")
func tri(a, b, c rl.Vector2, col rl.Color) {
	rl.DrawTriangle(a, b, c, col)
	rl.DrawTriangle(a, c, b, col)
}

func vv(x, y float32) rl.Vector2 { return rl.NewVector2(x, y) }

func fade(c rl.Color, a uint8) rl.Color {
	c.A = uint8(uint16(c.A) * uint16(a) / 255)
	return c
}

// drawKeycap: o tasta desenata (ca pe tastatura), cu litera pe ea.
// lit = apasa acum (verde, pulseaza).
func (g *Game) drawKeycap(label string, cx, cy, s float32, lit bool, a uint8) {
	w := s
	if len([]rune(label)) > 1 {
		w = s * (0.55 + 0.32*float32(len([]rune(label))))
	}
	if lit {
		k := 1 + 0.06*sinf(g.Time*9)
		w *= k
		s *= k
	}
	body := rl.NewColor(236, 238, 244, 255)
	edge := rl.NewColor(150, 156, 172, 255)
	txt := rl.NewColor(30, 34, 48, 255)
	if lit {
		body = rl.NewColor(120, 230, 130, 255)
		edge = rl.NewColor(50, 150, 70, 255)
	}
	r := rl.NewRectangle(cx-w/2, cy-s/2, w, s)
	rl.DrawRectangleRounded(rl.NewRectangle(r.X-3, r.Y-3, r.Width+6, r.Height+9), 0.3, 6, fade(icoDark, a))
	rl.DrawRectangleRounded(rl.NewRectangle(r.X, r.Y+s*0.1, r.Width, r.Height), 0.3, 6, fade(edge, a))
	rl.DrawRectangleRounded(r, 0.3, 6, fade(body, a))
	fs := s * 0.62
	if len([]rune(label)) > 1 {
		fs = s * 0.42
	}
	g.textCenter(label, cx, cy-fs*0.55, fs, fade(txt, a), true)
}

// drawSpaceKey: bara de spatiu (lunga), cu sageti in sus pe ea.
func (g *Game) drawSpaceKey(cx, cy, s float32, lit bool, a uint8) {
	w := s * 3.6
	if lit {
		k := 1 + 0.07*sinf(g.Time*10)
		w *= k
		s *= k
	}
	body := rl.NewColor(236, 238, 244, 255)
	edge := rl.NewColor(150, 156, 172, 255)
	if lit {
		body = rl.NewColor(120, 230, 130, 255)
		edge = rl.NewColor(50, 150, 70, 255)
	}
	r := rl.NewRectangle(cx-w/2, cy-s/2, w, s)
	rl.DrawRectangleRounded(rl.NewRectangle(r.X-3, r.Y-3, r.Width+6, r.Height+9), 0.35, 6, fade(icoDark, a))
	rl.DrawRectangleRounded(rl.NewRectangle(r.X, r.Y+s*0.1, r.Width, r.Height), 0.35, 6, fade(edge, a))
	rl.DrawRectangleRounded(r, 0.35, 6, fade(body, a))
	for _, dx := range []float32{-w * 0.25, 0, w * 0.25} {
		drawIcon(IconArrowUp, cx+dx, cy, s*0.55, fade(rl.NewColor(40, 60, 90, 255), a))
	}
}

// drawIcon: pictograma kind, centrata in (cx, cy), cam de marimea s.
func drawIcon(kind IconKind, cx, cy, s float32, col rl.Color) {
	h := s / 2
	switch kind {
	case IconCar:
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h, cy-h*0.15, s, h*0.7), 0.4, 6, col)
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.55, cy-h*0.62, s*0.55, h*0.6), 0.5, 6, col)
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.42, cy-h*0.52, s*0.2, h*0.38), 0.3, 4, rl.NewColor(150, 210, 255, col.A))
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.0, cy-h*0.52, s*0.2, h*0.38), 0.3, 4, rl.NewColor(150, 210, 255, col.A))
		for _, wx := range []float32{-h * 0.55, h * 0.55} {
			rl.DrawCircleV(vv(cx+wx, cy+h*0.55), h*0.27, fade(icoDark, col.A))
			rl.DrawCircleV(vv(cx+wx, cy+h*0.55), h*0.12, rl.NewColor(200, 200, 210, col.A))
		}
	case IconLiftUp, IconLiftDown:
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.75, cy-h, s*0.75, s), 0.15, 4, rl.NewColor(120, 130, 150, col.A))
		rl.DrawRectangle(int32(cx-1), int32(cy-h*0.9), 2, int32(s*0.9), fade(icoDark, col.A))
		k := IconArrowUp
		if kind == IconLiftDown {
			k = IconArrowDown
		}
		drawIcon(k, cx+h*0.7, cy, s*0.6, col)
	case IconWheel:
		rl.DrawRing(vv(cx, cy-h*0.1), h*0.72, h*0.86, 0, 360, 32, col)
		for k := 0; k < 8; k++ {
			a := float32(k) * math.Pi / 4
			rl.DrawLineEx(vv(cx, cy-h*0.1), vv(cx+cosf(a)*h*0.75, cy-h*0.1+sinf(a)*h*0.75), 2, col)
			rl.DrawCircleV(vv(cx+cosf(a)*h*0.82, cy-h*0.1+sinf(a)*h*0.82), h*0.14, []rl.Color{LegoRed, LegoYellow, LegoBlue, LegoLime}[k%4])
		}
		rl.DrawLineEx(vv(cx, cy-h*0.1), vv(cx-h*0.6, cy+h), 3, col)
		rl.DrawLineEx(vv(cx, cy-h*0.1), vv(cx+h*0.6, cy+h), 3, col)
	case IconCoaster:
		pts := []rl.Vector2{}
		for k := 0; k <= 16; k++ {
			t := float32(k) / 16
			pts = append(pts, vv(cx-h+t*s, cy+h*0.4-sinf(t*math.Pi*1.5)*h*0.9))
		}
		for k := 0; k+1 < len(pts); k++ {
			rl.DrawLineEx(pts[k], pts[k+1], s*0.08, col)
			if k%3 == 0 {
				rl.DrawLineEx(pts[k], vv(pts[k].X, cy+h), 2, fade(col, 160))
			}
		}
		rl.DrawRectangleRounded(rl.NewRectangle(pts[5].X-h*0.3, pts[5].Y-h*0.42, h*0.6, h*0.36), 0.4, 4, LegoRed)
	case IconTrampoline:
		rl.DrawEllipse(int32(cx), int32(cy+h*0.1), h, h*0.32, fade(icoDark, col.A))
		rl.DrawEllipse(int32(cx), int32(cy+h*0.05), h*0.9, h*0.25, col)
		for _, lx := range []float32{-h * 0.8, h * 0.8} {
			rl.DrawLineEx(vv(cx+lx, cy+h*0.15), vv(cx+lx, cy+h*0.8), 3, col)
		}
		drawIcon(IconArrowUp, cx, cy-h*0.55, s*0.5, LegoGoldBright)
	case IconSword:
		rl.DrawLineEx(vv(cx-h*0.65, cy+h*0.65), vv(cx+h*0.75, cy-h*0.75), s*0.14, rl.NewColor(130, 240, 255, col.A))
		rl.DrawLineEx(vv(cx-h*0.65, cy+h*0.65), vv(cx+h*0.75, cy-h*0.75), s*0.05, icoLight)
		rl.DrawLineEx(vv(cx-h*0.75, cy+h*0.15), vv(cx-h*0.15, cy+h*0.75), s*0.11, LegoGold)
		rl.DrawLineEx(vv(cx-h*0.5, cy+h*0.5), vv(cx-h*0.9, cy+h*0.9), s*0.1, rl.NewColor(110, 70, 40, col.A))
	case IconBrick:
		drawBrick(cx, cy, s, col)
	case IconCoin, IconNoCoin:
		drawCoin(cx, cy, h*0.8, LegoGold)
		if kind == IconNoCoin {
			rl.DrawLineEx(vv(cx-h*0.8, cy-h*0.8), vv(cx+h*0.8, cy+h*0.8), s*0.12, UIRed)
			rl.DrawLineEx(vv(cx+h*0.8, cy-h*0.8), vv(cx-h*0.8, cy+h*0.8), s*0.12, UIRed)
		}
	case IconTunnel:
		rl.DrawCircleSector(vv(cx, cy+h*0.3), h, 180, 360, 24, rl.NewColor(130, 120, 110, col.A))
		rl.DrawRectangle(int32(cx-h), int32(cy+h*0.3), int32(s), int32(h*0.7), rl.NewColor(130, 120, 110, col.A))
		rl.DrawCircleSector(vv(cx, cy+h*0.5), h*0.6, 180, 360, 20, fade(icoDark, col.A))
		rl.DrawRectangle(int32(cx-h*0.6), int32(cy+h*0.5), int32(h*1.2), int32(h*0.5), fade(icoDark, col.A))
		rl.DrawCircleV(vv(cx, cy+h*0.3), h*0.12, LegoGoldBright)
	case IconBridge:
		rl.DrawRectangle(int32(cx-h), int32(cy), int32(s), int32(h*0.22), col)
		rl.DrawRing(vv(cx, cy+h*0.9), h*0.75, h*0.88, 180, 360, 20, col)
		for _, px := range []float32{-h * 0.7, h * 0.7} {
			rl.DrawRectangle(int32(cx+px-2), int32(cy-h*0.8), 4, int32(h*0.8), col)
		}
		rl.DrawRectangle(int32(cx-h), int32(cy+h*0.7), int32(s), int32(h*0.3), rl.NewColor(80, 160, 230, col.A))
	case IconIsland:
		rl.DrawEllipse(int32(cx), int32(cy+h*0.55), h, h*0.35, rl.NewColor(80, 160, 230, col.A))
		rl.DrawEllipse(int32(cx), int32(cy+h*0.45), h*0.7, h*0.22, rl.NewColor(240, 220, 150, col.A))
		rl.DrawLineEx(vv(cx, cy+h*0.4), vv(cx+h*0.15, cy-h*0.6), 4, rl.NewColor(120, 80, 40, col.A))
		for k := 0; k < 4; k++ {
			a := float32(k)*0.9 + 3.6
			rl.DrawLineEx(vv(cx+h*0.15, cy-h*0.6), vv(cx+h*0.15+cosf(a)*h*0.6, cy-h*0.6+sinf(a)*h*0.35+h*0.2), 5, LegoGreen)
		}
	case IconStar:
		drawStar(cx, cy, h, col)
	case IconCamera:
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h, cy-h*0.5, s, h*1.3), 0.3, 6, col)
		rl.DrawRectangle(int32(cx-h*0.3), int32(cy-h*0.75), int32(h*0.6), int32(h*0.3), col)
		rl.DrawCircleV(vv(cx, cy+h*0.12), h*0.42, fade(icoDark, col.A))
		rl.DrawCircleV(vv(cx, cy+h*0.12), h*0.25, rl.NewColor(120, 190, 255, col.A))
	case IconSave:
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.8, cy-h*0.8, s*0.8, s*0.8), 0.15, 4, LegoBlue)
		rl.DrawRectangle(int32(cx-h*0.45), int32(cy-h*0.8), int32(h*0.9), int32(h*0.55), icoLight)
		rl.DrawRectangle(int32(cx-h*0.5), int32(cy+h*0.15), int32(h), int32(h*0.6), rl.NewColor(220, 225, 235, col.A))
	case IconHeart:
		drawHeart(cx, cy, s*0.9, UIRed)
	case IconExit:
		rl.DrawRectangleRoundedLinesEx(rl.NewRectangle(cx-h*0.9, cy-h*0.8, h*1.0, s*0.8), 0.2, 4, 3, col)
		drawIcon(IconArrowRight, cx+h*0.35, cy, s*0.6, col)
	case IconMouse:
		rl.DrawRectangleRounded(rl.NewRectangle(cx-h*0.55, cy-h*0.85, h*1.1, s*0.85), 0.9, 10, col)
		rl.DrawLineEx(vv(cx, cy-h*0.85), vv(cx, cy-h*0.25), 2, fade(icoDark, col.A))
		rl.DrawLineEx(vv(cx-h*0.55, cy-h*0.25), vv(cx+h*0.55, cy-h*0.25), 2, fade(icoDark, col.A))
	case IconWait:
		tri(vv(cx-h*0.6, cy-h*0.85), vv(cx+h*0.6, cy-h*0.85), vv(cx, cy), col)
		tri(vv(cx-h*0.6, cy+h*0.85), vv(cx+h*0.6, cy+h*0.85), vv(cx, cy), col)
		rl.DrawRectangle(int32(cx-h*0.75), int32(cy-h*0.95), int32(h*1.5), 3, col)
		rl.DrawRectangle(int32(cx-h*0.75), int32(cy+h*0.9), int32(h*1.5), 3, col)
	case IconGrow:
		drawBrick(cx-h*0.55, cy+h*0.35, s*0.35, col)
		drawIcon(IconArrowRight, cx-h*0.05, cy+h*0.2, s*0.3, icoLight)
		drawBrick(cx+h*0.5, cy, s*0.75, col)
	case IconTrophy:
		rl.DrawCircleSector(vv(cx, cy-h*0.35), h*0.6, 0, 180, 16, LegoGoldBright)
		rl.DrawRectangle(int32(cx-h*0.6), int32(cy-h*0.75), int32(h*1.2), int32(h*0.4), LegoGoldBright)
		rl.DrawRectangle(int32(cx-h*0.12), int32(cy+h*0.2), int32(h*0.24), int32(h*0.4), LegoGold)
		rl.DrawRectangle(int32(cx-h*0.45), int32(cy+h*0.6), int32(h*0.9), int32(h*0.25), LegoGold)
		drawStar(cx, cy-h*0.45, h*0.25, icoLight)
	case IconMap:
		rl.DrawRectangle(int32(cx-h*0.9), int32(cy-h*0.7), int32(s*0.9), int32(s*0.7), rl.NewColor(80, 160, 230, col.A))
		rl.DrawEllipse(int32(cx-h*0.25), int32(cy-h*0.1), h*0.4, h*0.3, LegoGreen)
		rl.DrawEllipse(int32(cx+h*0.45), int32(cy+h*0.3), h*0.25, h*0.18, LegoGreen)
		rl.DrawCircleV(vv(cx+h*0.5, cy-h*0.35), h*0.14, UIRed)
	case IconHelp:
		rl.DrawCircleV(vv(cx, cy), h, LegoBlue)
	case IconMonster:
		rl.DrawCircleV(vv(cx, cy), h*0.8, col)
		rl.DrawCircleV(vv(cx-h*0.3, cy-h*0.15), h*0.22, icoLight)
		rl.DrawCircleV(vv(cx+h*0.3, cy-h*0.15), h*0.22, icoLight)
		rl.DrawCircleV(vv(cx-h*0.28, cy-h*0.12), h*0.1, fade(icoDark, col.A))
		rl.DrawCircleV(vv(cx+h*0.32, cy-h*0.12), h*0.1, fade(icoDark, col.A))
		tri(vv(cx-h*0.65, cy-h*0.5), vv(cx-h*0.35, cy-h*0.75), vv(cx-h*0.75, cy-h*1.0), col)
		tri(vv(cx+h*0.65, cy-h*0.5), vv(cx+h*0.35, cy-h*0.75), vv(cx+h*0.75, cy-h*1.0), col)
		rl.DrawRectangle(int32(cx-h*0.35), int32(cy+h*0.3), int32(h*0.7), int32(h*0.14), fade(icoDark, col.A))
	case IconCheck:
		rl.DrawLineEx(vv(cx-h*0.7, cy), vv(cx-h*0.15, cy+h*0.55), s*0.16, col)
		rl.DrawLineEx(vv(cx-h*0.15, cy+h*0.55), vv(cx+h*0.75, cy-h*0.6), s*0.16, col)
	case IconEye:
		rl.DrawEllipse(int32(cx), int32(cy), h, h*0.55, col)
		rl.DrawCircleV(vv(cx, cy), h*0.4, rl.NewColor(70, 140, 220, col.A))
		rl.DrawCircleV(vv(cx, cy), h*0.2, fade(icoDark, col.A))
	case IconHorn:
		tri(vv(cx-h*0.7, cy-h*0.25), vv(cx+h*0.6, cy-h*0.8), vv(cx+h*0.6, cy+h*0.8), col)
		rl.DrawRectangle(int32(cx-h*0.9), int32(cy-h*0.25), int32(h*0.3), int32(h*0.5), col)
	case IconTurbo:
		tri(vv(cx+h*0.25, cy-h), vv(cx-h*0.55, cy+h*0.15), vv(cx+h*0.05, cy+h*0.15), LegoGoldBright)
		tri(vv(cx-h*0.2, cy+h), vv(cx+h*0.55, cy-h*0.15), vv(cx-h*0.05, cy-h*0.15), LegoGoldBright)
	case IconLock, IconOutfit, IconGun, IconGPS, IconTeleport, IconPhone, IconCrown, IconGoldBrick, IconClose, IconTrain, IconWorld, IconWings:
		drawIcon5(kind, cx, cy, s, col)
	case IconArrowUp, IconArrowDown, IconArrowLeft, IconArrowRight:
		dx, dy := float32(0), float32(-1)
		switch kind {
		case IconArrowDown:
			dy = 1
		case IconArrowLeft:
			dx, dy = -1, 0
		case IconArrowRight:
			dx, dy = 1, 0
		}
		px, py := -dy, dx
		tip := vv(cx+dx*h, cy+dy*h)
		bl := vv(cx-dx*h*0.05+px*h*0.75, cy-dy*h*0.05+py*h*0.75)
		br := vv(cx-dx*h*0.05-px*h*0.75, cy-dy*h*0.05-py*h*0.75)
		tri(tip, bl, br, col)
		rl.DrawLineEx(vv(cx, cy), vv(cx-dx*h, cy-dy*h), s*0.28, col)
	}
}

// drawBrick: caramida blocuri 2x2 vazuta din fata (cu nopi).
func drawBrick(cx, cy, s float32, col rl.Color) {
	h := s / 2
	rl.DrawRectangleRounded(rl.NewRectangle(cx-h, cy-h*0.35, s, h*1.1), 0.2, 4, shade(col, 0.8))
	rl.DrawRectangleRounded(rl.NewRectangle(cx-h, cy-h*0.45, s, h*0.95), 0.2, 4, col)
	for _, sx := range []float32{-h * 0.5, h * 0.5} {
		rl.DrawRectangleRounded(rl.NewRectangle(cx+sx-h*0.28, cy-h*0.72, h*0.56, h*0.32), 0.3, 4, shade(col, 1.12))
	}
}

func drawStar(cx, cy, r float32, col rl.Color) {
	pts := make([]rl.Vector2, 10)
	for k := 0; k < 10; k++ {
		a := float32(k)*math.Pi/5 - math.Pi/2
		rr := r
		if k%2 == 1 {
			rr = r * 0.45
		}
		pts[k] = vv(cx+cosf(a)*rr, cy+sinf(a)*rr)
	}
	for k := 0; k < 10; k++ {
		tri(vv(cx, cy), pts[k], pts[(k+1)%10], col)
	}
}

// iconBubble: pictograma pe un cerc inchis (se vede pe orice fundal).
func iconBubble(kind IconKind, cx, cy, s float32, ring rl.Color, a uint8) {
	rl.DrawCircleV(vv(cx, cy), s*0.62, fade(icoDark, a))
	rl.DrawRing(vv(cx, cy), s*0.58, s*0.64, 0, 360, 36, fade(ring, a))
	drawIcon(kind, cx, cy, s*0.72, fade(icoLight, a))
}

// hint: text mic si discret (pentru parinte), cu contur ca sa se citeasca pe orice fundal.
func (g *Game) hint(text string, cx, y float32, a uint8) {
	if text == "" {
		return
	}
	tw := g.measure(text, 16, false) + 18
	rl.DrawRectangleRounded(rl.NewRectangle(cx-tw/2, y-2, tw, 22), 0.5, 6, rl.NewColor(16, 20, 32, uint8(float32(a)*0.55)))
	g.textCenter(text, cx, y, 16, rl.NewColor(235, 240, 250, uint8(float32(a)*0.85)), false)
}

// drawBrickBar: bara de progres din caramizi blocuri (pusa / ramase), latime w.
func drawBrickBar(x, y, w, hgt float32, done, total int, col rl.Color, t float32) {
	if total <= 0 {
		return
	}
	rl.DrawRectangleRounded(rl.NewRectangle(x-4, y-4, w+8, hgt+8), 0.35, 6, icoDark)
	if total <= 24 {
		gap := float32(3)
		cw := (w - gap*float32(total-1)) / float32(total)
		for k := 0; k < total; k++ {
			bx := x + float32(k)*(cw+gap)
			c := rl.NewColor(70, 76, 96, 255)
			if k < done {
				c = col
			}
			r := rl.NewRectangle(bx, y, cw, hgt)
			rl.DrawRectangleRounded(r, 0.25, 4, c)
			if k < done {
				rl.DrawRectangle(int32(bx+cw*0.25), int32(y-hgt*0.18), int32(maxf(cw*0.5, 2)), int32(hgt*0.22), shade(c, 1.15))
			}
			if k == done {
				// urmatoarea piesa: pulseaza (aici pune sabia)
				a := uint8(120 + 120*sinf(t*6))
				rl.DrawRectangleRoundedLinesEx(r, 0.25, 4, 3, rl.NewColor(255, 240, 140, a))
			}
		}
		return
	}
	k := clampf(float32(done)/float32(total), 0, 1)
	rl.DrawRectangleRounded(rl.NewRectangle(x, y, w, hgt), 0.4, 6, rl.NewColor(70, 76, 96, 255))
	if k > 0 {
		rl.DrawRectangleRounded(rl.NewRectangle(x, y, w*k, hgt), 0.4, 6, col)
	}
}

// drawDots: n buline, primele done pline (ex. 5 lovituri pentru marire).
func drawDots(cx, cy, r float32, done, n int, col rl.Color) {
	gap := r * 2.8
	x0 := cx - gap*float32(n-1)/2
	for k := 0; k < n; k++ {
		x := x0 + float32(k)*gap
		rl.DrawCircleV(vv(x, cy), r+3, icoDark)
		c := rl.NewColor(70, 76, 96, 255)
		if k < done {
			c = col
		}
		rl.DrawCircleV(vv(x, cy), r, c)
	}
}
