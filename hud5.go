package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 5.0: HUD nou: stelele lumii, harta de dezvaluire la trecerea nivelului.
// ====================================================================

func (g *Game) levelBannerText() string {
	wi := g.levelWorld(g.Level)
	if wi >= 0 {
		return fmt.Sprintf("NIVELUL %d  •  %s: învinge monștrii și Șeful!", g.Level, Worlds[wi].Name)
	}
	return fmt.Sprintf("NIVELUL %d  •  monștri pe toate insulele!", g.Level)
}

// hudWorld: lumea ale carei stele le aratam (cea in care e Aventurierul, altfel cea a nivelului).
func (g *Game) hudWorld() int {
	if wi := g.currentWorld(); wi >= 0 && g.worldOpen(wi) {
		return wi
	}
	if wi := g.levelWorld(g.Level); wi >= 0 {
		return wi
	}
	return 0
}

// drawStarsHUD: sub panoul jucatorului: culoarea lumii + 3 stele desenate.
func (g *Game) drawStarsHUD() {
	wi := g.hudWorld()
	wd := Worlds[wi]
	r := rl.NewRectangle(18, 160, 360, 64)
	panel(r, wd.Main)
	rl.DrawCircleV(vv(r.X+30, r.Y+32), 16, wd.Main)
	rl.DrawCircleLines(int32(r.X+30), int32(r.Y+32), 16, rl.White)
	drawIcon(IconIsland, r.X+30, r.Y+32, 22, rl.White)
	type st struct {
		ok   bool
		icon IconKind
	}
	stars := []st{{g.starMonsters(wi), IconCrown}, {g.starBuild(wi), IconBrick}, {g.starBricks(wi), IconGoldBrick}}
	switch wd.Sector {
	case SecWest:
		stars[1].icon = IconTrampoline
	case SecPark:
		stars[1].icon = IconCoaster
	}
	for k, s := range stars {
		x := r.X + 78 + float32(k)*72
		col := rl.NewColor(80, 86, 100, 255)
		if s.ok {
			col = LegoGoldBright
		}
		drawStar(x, r.Y+30, 22, col)
		drawIcon(s.icon, x+20, r.Y+44, 20, rl.White)
	}
	// piesele aurii gasite in lumea asta: 5 caramizi mici
	n := g.bricksFound(wi)
	for k := 0; k < BricksPerWorld; k++ {
		c := rl.NewColor(70, 74, 90, 255)
		if k < n {
			c = LegoGoldBright
		}
		drawBrick(r.X+300+float32(k%3)*18, r.Y+18+float32(k/3)*22, 14, c)
	}
	g.hint(fmt.Sprintf("%s  •  stele: %d (în total %d)", wd.Name, g.worldStars(wi), g.totalStars()), r.X+r.Width/2, r.Y+r.Height+4, 190)
}

// drawReveal: la trecerea nivelului: stea mare, apoi harta cu insula noua care iese din nor.
func (g *Game) drawReveal() {
	rv := &g.Prog.Reveal
	if !rv.Active {
		return
	}
	W, H := sw(), sh()
	if rv.T < revealConfetti+0.3 {
		a := uint8(255 * clampf(rv.T/0.3, 0, 1))
		cy := H * 0.3
		s := 70 + 12*sinf(g.Time*8)
		drawStar(W/2, cy, s, fade(LegoGoldBright, a))
		drawIcon(IconCheck, W/2, cy+4, s*0.8, fade(UIGreen, a))
		g.hint(fmt.Sprintf("Nivelul %d trecut!", g.Level-1), W/2, cy+s+20, a)
	}
	ma := g.revealMapAlpha()
	if ma <= 0 || rv.World < 0 {
		return
	}
	rl.DrawRectangle(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(10, 14, 30, uint8(200*ma)))
	sc := minf((W-120)/900, (H-170)/540)
	mw, mh := 900*sc, 540*sc
	x0, y0 := (W-mw)/2, (H-mh)/2+20
	g.drawMapContents(x0, y0, mw, mh, sc, 0, 0, false)
	msc := sc
	if curRealm > 0 {
		msc = sc * 0.84 // aceeasi scara ca drawRealmMap
	}
	toMap := func(x, z float32) rl.Vector2 { return vv(x0+mw/2+x*msc, y0+mh/2+z*msc) }
	wd := Worlds[rv.World]
	c := Sectors[wd.Sector].Center
	p := toMap(c.X, c.Y)
	// inelul auriu care pulseaza in jurul insulei noi + steluțe
	t := rv.T - revealConfetti - revealMapIn
	if t > 0 {
		rr := 70*sc + 8*sinf(g.Time*5)
		rl.DrawRing(p, rr, rr+6, 0, 360, 48, LegoGoldBright)
		for k := 0; k < 6; k++ {
			a := g.Time*1.5 + float32(k)*1.047
			drawStar(p.X+cosf(a)*(rr+22), p.Y+sinf(a)*(rr+22), 9, LegoWhite)
		}
		if t > 2 {
			drawIcon(IconOutfit, p.X+rr+40, p.Y-rr*0.4, 46, rl.White)
		}
	}
	// Aventurierul (capul galben) clipeste unde e acum
	pp := toMap(g.Player.Pos.X, g.Player.Pos.Z)
	if int(g.Time*4)%2 == 0 {
		rl.DrawCircleV(pp, 12, LegoYellow)
		rl.DrawCircleLines(int32(pp.X), int32(pp.Y), 12, icoDark)
	}
	// acoperim totul cu negru cand harta apare / dispare
	if ma < 1 {
		rl.DrawRectangle(int32(x0-4), int32(y0-4), int32(mw+8), int32(mh+8), rl.NewColor(10, 14, 30, uint8(255*(1-ma))))
	}
	g.hint("Insulă nouă: "+wd.Name+"! Urmează săgețile verzi.", W/2, y0-36, uint8(255*ma))
}

// drawSnow: fulgi de zapada in jurul camerei, in lumea de Craciun (deschisa).
func (g *Game) drawSnow() {
	// V 6.0: ninge doar in Satul de Craciun (parcul din lumea 0 e acum de vara)
	cam := g.Camera.Position
	if Realms[curRealm].ID != "craciun" || g.InTunnel >= 0 {
		return
	}
	for i := 0; i < 260; i++ {
		fi := float32(i)
		// pozitii pseudo-aleatoare intr-un cub de 40 m care urmeaza camera
		ox := float32(math.Mod(float64(fi*37.7+cam.X*0), 40)) - 20
		oz := float32(math.Mod(float64(fi*53.3), 40)) - 20
		fall := float32(math.Mod(float64(fi*11.1+g.Time*(1.4+float32(i%5)*0.25)), 22))
		x := float32(math.Floor(float64(cam.X/40)))*40 + ox
		z := float32(math.Floor(float64(cam.Z/40)))*40 + oz
		for x-cam.X > 20 {
			x -= 40
		}
		for cam.X-x > 20 {
			x += 40
		}
		for z-cam.Z > 20 {
			z -= 40
		}
		for cam.Z-z > 20 {
			z += 40
		}
		y := cam.Y + 12 - fall
		x += sinf(g.Time*1.3+fi) * 0.6
		s := 0.07 + float32(i%3)*0.03
		rl.DrawCube(v3(x, y, z), s, s, s, rl.NewColor(255, 255, 255, 230))
	}
}
