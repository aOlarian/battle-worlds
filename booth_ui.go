package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 5.0: ECRANUL CABINEI TELEFONICE (totul desenat; text mic pentru parinte)
//   acasa: 3 desene mari -> Harta | Personaje | Arme
//   harta: apesi pe o insula deschisa -> sageti GPS sau teleport
//   personaje: costumele castigate (umbra cu lacat = inca nu)
//   arme: armele deblocate cu stele (lacat + cate stele trebuie)
// ====================================================================

const (
	BoothHome = iota
	BoothMap
	BoothOutfits
	BoothGuns
	BoothRealms // V 6.0: lumile mari (calatorie cu trenul magic)
)

type BoothUI struct {
	Booth  int
	Tab    int
	Sel    int // insula aleasa pe harta (-1 = niciuna)
	T      float32
	Shake  float32
	Action func()

	rtChar  []rl.RenderTexture2D
	rtGun   []rl.RenderTexture2D
	rtReady bool
	dirty   bool
}

func (g *Game) openBooth(i int) {
	g.Booth.Booth = i
	g.Booth.Tab = BoothHome
	g.Booth.Sel = -1
	g.Booth.T = 0
	g.Booth.dirty = true
	g.PrevScene = SceneMapExplore
	g.Scene = SceneBooth
	g.unlockCursor()
	g.Play(g.Audio.Ding)
}

func (g *Game) closeBooth() {
	g.Scene = SceneMapExplore
	g.SkipMouse = 3
	g.lockCursor()
	g.Play(g.Audio.Click)
}

func (g *Game) updateBooth() {
	b := &g.Booth
	b.T += g.FrameTime
	if b.Shake > 0 {
		b.Shake -= g.FrameTime
	}
	g.City.Update(g.FrameTime, g.Time, g.Player.Pos, g.Camera.Position, g.Audio)
	g.updateParticles()
	if b.Action != nil {
		a := b.Action
		b.Action = nil
		a()
		return
	}
	if rl.IsKeyPressed(rl.KeyEscape) || rl.IsKeyPressed(rl.KeyE) || rl.IsMouseButtonPressed(rl.MouseButtonRight) {
		if b.Tab != BoothHome {
			b.Tab = BoothHome
			b.Sel = -1
			g.Play(g.Audio.Click)
		} else {
			g.closeBooth()
		}
	}
}

// ---------------------------------------------------------------
// actiunile (le foloseste si testul automat)
// ---------------------------------------------------------------

func (g *Game) boothTeleport(wi int) {
	if !g.worldOpen(wi) || wi >= len(g.City.Booths) {
		return
	}
	g.closeBooth()
	g.stopGPS()
	g.Play(g.Audio.Spring)
	g.startFade(func() {
		bt := &g.City.Booths[wi]
		d := bt.DoorPos(2.8)
		gy, _ := g.City.W.GroundAt(d.X, d.Z, bt.Pos.Y+2)
		g.Player.Pos = v3(d.X, gy, d.Z)
		g.Player.VelY = 0
		g.Player.OnGround = true
		g.Swimming = false
		g.Player.Yaw = bt.Yaw
		g.CamYaw = bt.Yaw + math.Pi
		g.prevPos = g.Player.Pos
		g.sayIcon(IconTeleport, "Teleport: "+Worlds[wi].Name+"!")
		p := g.Player.Pos
		for n := 0; n < 60; n++ {
			a := float32(n) * 0.6
			g.spawnParticle(v3(p.X+cosf(a), p.Y+float32(n%12)*0.25, p.Z+sinf(a)), v3(cosf(a)*2, 3, sinf(a)*2), []rl.Color{LegoCyan, LegoWhite, LegoGoldBright}[n%3], 0.2, 1.2)
		}
	})
}

func (g *Game) boothNavigate(wi int) {
	if !g.worldOpen(wi) {
		return
	}
	g.closeBooth()
	if g.currentWorld() == wi {
		return
	}
	g.startGPSToWorld(wi)
	g.sayIcon(IconGPS, "Urmează săgețile verzi până la "+Worlds[wi].Name+"!")
}

func (g *Game) boothWear(oi int) bool {
	if oi < 0 || oi >= len(Outfits) || !g.ownOutfit(oi) {
		return false
	}
	g.wearOutfit(oi)
	g.Play(g.Audio.Achievement)
	return true
}

func (g *Game) boothGun(k int) bool {
	if !g.gunUnlocked(k) {
		return false
	}
	g.Prog.Gun = k
	g.Player.Pose.Gun = k
	g.Play(g.Audio.DrawGun)
	return true
}

// ---------------------------------------------------------------
// previzualizari 3D (costume si arme), randate o data in texturi
// ---------------------------------------------------------------

const prevW, prevH = 256, 320

func (g *Game) renderBoothPreviews() {
	b := &g.Booth
	if b.rtReady && !b.dirty {
		return
	}
	if !b.rtReady {
		for range Outfits {
			b.rtChar = append(b.rtChar, rl.LoadRenderTexture(prevW, prevH))
		}
		for k := 0; k < NumGuns; k++ {
			b.rtGun = append(b.rtGun, rl.LoadRenderTexture(prevW, prevW*3/4))
		}
		b.rtReady = true
	}
	b.dirty = false
	r := g.R
	studio := v3(0, 600, 0) // departe de lume: fara umbre si fara ceata
	lit := func(mesh rl.Mesh, m rl.Matrix, tint rl.Color) { r.DrawLit(mesh, m, tint) }
	setF(r.Lit, r.locLit["shadowOn"], 0)
	r.SetIndoor(0)
	for i, o := range Outfits {
		cam := rl.Camera3D{Position: v3(studio.X+2.2, studio.Y+2.6, studio.Z+5.6), Target: v3(studio.X, studio.Y+1.45, studio.Z), Up: v3(0, 1, 0), Fovy: 40, Projection: rl.CameraPerspective}
		rl.BeginTextureMode(b.rtChar[i])
		rl.ClearBackground(rl.Blank)
		r.BeginFrame(cam, g.Time)
		setF(r.Lit, r.locLit["shadowOn"], 0)
		rl.BeginMode3D(cam)
		cm := GetCharMeshes(o.Style)
		pose := Pose{Breathe: g.Time * 2, Weapon: WeaponNone}
		DrawCharacter(lit, cm, studio, 0.35, 1, pose, rl.White)
		rl.EndMode3D()
		rl.EndTextureMode()
	}
	for k := 0; k < NumGuns; k++ {
		cam := rl.Camera3D{Position: v3(studio.X+1.7, studio.Y+0.7, studio.Z+0.55), Target: v3(studio.X, studio.Y, studio.Z+0.55), Up: v3(0, 1, 0), Fovy: 36, Projection: rl.CameraPerspective}
		rl.BeginTextureMode(b.rtGun[k])
		rl.ClearBackground(rl.Blank)
		r.BeginFrame(cam, g.Time)
		setF(r.Lit, r.locLit["shadowOn"], 0)
		rl.BeginMode3D(cam)
		if weaponsReady {
			m := rl.MatrixMultiply(rl.MatrixRotateY(-0.35), rl.MatrixTranslate(studio.X, studio.Y-0.1, studio.Z))
			r.DrawLit(gunMeshes[k], m, rl.White)
		}
		rl.EndMode3D()
		rl.EndTextureMode()
	}
}

func drawRT(rt rl.RenderTexture2D, dst rl.Rectangle, tint rl.Color) {
	src := rl.NewRectangle(0, 0, float32(rt.Texture.Width), -float32(rt.Texture.Height))
	rl.DrawTexturePro(rt.Texture, src, dst, rl.NewVector2(0, 0), 0, tint)
}

// ---------------------------------------------------------------
// desen
// ---------------------------------------------------------------

func (g *Game) boothCard(r rl.Rectangle, col rl.Color, hot bool) bool {
	m := rl.GetMousePosition()
	hover := rl.CheckCollisionPointRec(m, r)
	lift := float32(0)
	if hover {
		lift = -6
	}
	rr := rl.NewRectangle(r.X, r.Y+lift, r.Width, r.Height)
	rl.DrawRectangleRounded(rl.NewRectangle(rr.X+5, rr.Y+8, rr.Width, rr.Height), 0.12, 10, rl.NewColor(0, 0, 0, 90))
	rl.DrawRectangleRounded(rr, 0.12, 10, col)
	bc := rl.NewColor(255, 255, 255, 140)
	if hover || hot {
		bc = LegoGoldBright
	}
	rl.DrawRectangleRoundedLinesEx(rr, 0.12, 10, 5, bc)
	return hover && rl.IsMouseButtonPressed(rl.MouseButtonLeft)
}

func (g *Game) drawBooth() {
	g.renderBoothPreviews()
	b := &g.Booth
	rl.DrawRectangle(0, 0, ScreenWidth, ScreenHeight, rl.NewColor(20, 26, 48, 175))
	W, H := sw(), sh()
	// sus: telefonul + inchide
	iconBubble(IconPhone, 56, 52, 64, LegoRed, 255)
	closeR := rl.NewRectangle(W-92, 18, 70, 70)
	if g.boothCard(closeR, rl.NewColor(0, 0, 0, 0), false) {
		b.Action = g.closeBooth
	}
	drawIcon(IconClose, closeR.X+35, closeR.Y+35, 60, rl.White)
	if b.Tab != BoothHome {
		backR := rl.NewRectangle(W-178, 18, 70, 70)
		if g.boothCard(backR, rl.NewColor(40, 50, 80, 230), false) {
			b.Action = func() { b.Tab = BoothHome; b.Sel = -1; g.Play(g.Audio.Click) }
		}
		drawIcon(IconArrowLeft, backR.X+35, backR.Y+35, 46, rl.White)
	}
	switch b.Tab {
	case BoothHome:
		g.drawBoothHome(W, H)
	case BoothMap:
		g.drawBoothMap(W, H)
	case BoothOutfits:
		g.drawBoothOutfits(W, H)
	case BoothGuns:
		g.drawBoothGuns(W, H)
	case BoothRealms:
		g.drawBoothRealms(W, H)
	}
	g.hint("E / ESC / click dreapta = înapoi", W/2, H-34, 200)
}

func (g *Game) drawBoothHome(W, H float32) {
	b := &g.Booth
	cw := minf(W*0.21, H*0.46)
	ch := cw * 1.2
	gap := cw * 0.1
	x0 := (W - (4*cw + 3*gap)) / 2
	y0 := (H-ch)/2 + 10
	labels := []string{"Harta: alegi insula, mergi sau te teleportezi", "Personaje: îți schimbi costumul", "Arme: alegi arma", "Lumi: pleci cu trenul magic în altă lume"}
	cols := []rl.Color{rl.NewColor(60, 150, 210, 255), rl.NewColor(250, 200, 70, 255), rl.NewColor(120, 90, 200, 255), rl.NewColor(230, 90, 120, 255)}
	for i := 0; i < 4; i++ {
		r := rl.NewRectangle(x0+float32(i)*(cw+gap), y0, cw, ch)
		pulse := 0.5 + 0.5*sinf(b.T*3+float32(i))
		if g.boothCard(r, cols[i], false) {
			tab := []int{BoothMap, BoothOutfits, BoothGuns, BoothRealms}[i]
			b.Action = func() { b.Tab = tab; b.Sel = -1; g.Play(g.Audio.Click) }
		}
		inner := rl.NewRectangle(r.X+14, r.Y+14, r.Width-28, r.Height-28)
		switch i {
		case 0:
			sc := inner.Width / 900
			g.drawMapContents(inner.X, inner.Y+inner.Height*0.12, inner.Width, inner.Height*0.76, sc, 0, 0, false)
		case 1:
			drawRT(b.rtChar[g.Prog.Outfit], rl.NewRectangle(inner.X+inner.Width*0.05, inner.Y, inner.Width*0.9, inner.Width*0.9*1.25), rl.White)
		case 2:
			drawRT(b.rtGun[g.Prog.Gun], rl.NewRectangle(inner.X, inner.Y+inner.Height*0.2, inner.Width, inner.Width*0.75), rl.White)
		case 3:
			drawIcon(IconWorld, inner.X+inner.Width/2, inner.Y+inner.Height*0.35, inner.Width*0.6, rl.White)
			drawIcon(IconTrain, inner.X+inner.Width/2, inner.Y+inner.Height*0.78, inner.Width*0.5, rl.White)
		}
		rl.DrawRectangleRoundedLinesEx(r, 0.12, 10, 2+3*pulse, rl.NewColor(255, 255, 255, uint8(60+80*pulse)))
		g.hint(labels[i], r.X+r.Width/2, r.Y+r.Height+16, 210)
	}
}

// boothMapLayout: unde se deseneaza harta mare in cabina
func boothMapLayout(W, H float32) (x0, y0, mw, mh, sc float32) {
	sc = minf((W-120)/900, (H-190)/540)
	mw, mh = 900*sc, 540*sc
	return (W - mw) / 2, (H-mh)/2 + 10, mw, mh, sc
}

func (g *Game) drawBoothMap(W, H float32) {
	b := &g.Booth
	x0, y0, mw, mh, sc := boothMapLayout(W, H)
	rl.DrawRectangleRounded(rl.NewRectangle(x0-10, y0-10, mw+20, mh+20), 0.04, 8, rl.NewColor(255, 255, 255, 220))
	g.drawMapContents(x0, y0, mw, mh, sc, 0, 0, false)
	msc := sc
	if curRealm > 0 {
		msc = sc * 0.84 // aceeasi scara ca drawRealmMap
	}
	toMap := func(x, z float32) rl.Vector2 { return vv(x0+mw/2+x*msc, y0+mh/2+z*msc) }
	m := rl.GetMousePosition()
	click := rl.IsMouseButtonPressed(rl.MouseButtonLeft)
	for wi, wd := range Worlds {
		c := Sectors[wd.Sector].Center
		p := toMap(c.X, c.Y)
		r := 62 * sc
		if wd.Sector == 4 {
			r = 50 * sc
		}
		hover := rl.Vector2Distance(m, p) < r
		open := g.worldOpen(wi)
		if open {
			ring := rl.NewColor(255, 255, 255, 160)
			if hover || b.Sel == wi {
				ring = LegoGoldBright
			}
			rl.DrawRing(p, r, r+5, 0, 360, 48, ring)
			rl.DrawCircleV(vv(p.X, p.Y-r*0.55), 13, wd.Main)
			rl.DrawCircleLines(int32(p.X), int32(p.Y-r*0.55), 13, rl.White)
			g.hint(wd.Name, p.X, p.Y+r*0.25, 190)
		} else {
			drawIcon(IconLock, p.X, p.Y-6, 38, rl.White)
			drawStar(p.X+30, p.Y+18, 14, LegoGoldBright)
			g.textCenter(itoa(wi+1), p.X+30, p.Y+9, 16, icoDark, true)
		}
		if hover && click {
			if open {
				sel := wi
				b.Action = func() { b.Sel = sel; g.Play(g.Audio.Click) }
			} else {
				b.Action = func() { b.Shake = 0.4; g.Play(g.Audio.Error) }
			}
		}
	}
	// Aventurierul pe harta (capul lui)
	pp := toMap(g.Player.Pos.X, g.Player.Pos.Z)
	if int(b.T*3)%2 == 0 {
		rl.DrawCircleV(pp, 11, LegoYellow)
		rl.DrawCircleLines(int32(pp.X), int32(pp.Y), 11, icoDark)
	}
	if b.Sel >= 0 {
		c := Sectors[Worlds[b.Sel].Sector].Center
		p := toMap(c.X, c.Y)
		bw := float32(96)
		bx := clampf(p.X-bw-8, x0, x0+mw-2*bw-16)
		by := clampf(p.Y+40, y0, y0+mh-bw)
		gr := rl.NewRectangle(bx, by, bw, bw)
		tr := rl.NewRectangle(bx+bw+16, by, bw, bw)
		sel := b.Sel
		if g.boothCard(gr, rl.NewColor(30, 120, 60, 240), true) {
			b.Action = func() { g.boothNavigate(sel) }
		}
		drawIcon(IconGPS, gr.X+bw/2, gr.Y+bw/2, bw*0.62, rl.White)
		g.hint("săgeți până acolo", gr.X+bw/2, gr.Y+bw+6, 220)
		if g.boothCard(tr, rl.NewColor(30, 90, 160, 240), true) {
			b.Action = func() { g.boothTeleport(sel) }
		}
		drawIcon(IconTeleport, tr.X+bw/2, tr.Y+bw/2, bw*0.62, rl.White)
		g.hint("teleport", tr.X+bw/2, tr.Y+bw+6, 220)
	}
	if b.Shake > 0 {
		g.hint("Insula se deschide când treci nivelul (steaua cu cifra)", W/2, y0-40, 230)
	}
}

func (g *Game) drawBoothOutfits(W, H float32) {
	b := &g.Booth
	n := len(Outfits)
	cols := 4
	rows := (n + cols - 1) / cols
	cw := minf((W-160)/float32(cols), (H-200)/float32(rows)/1.25)
	ch := cw * 1.25
	gap := cw * 0.08
	cw -= gap
	ch = cw * 1.25
	x0 := (W - (float32(cols)*cw + float32(cols-1)*gap)) / 2
	y0 := (H-(float32(rows)*ch+float32(rows-1)*gap))/2 + 10
	for i := 0; i < n; i++ {
		r := rl.NewRectangle(x0+float32(i%cols)*(cw+gap), y0+float32(i/cols)*(ch+gap), cw, ch)
		own := g.ownOutfit(i)
		bg := rl.NewColor(60, 70, 100, 235)
		wi := -1
		for k, wd := range Worlds {
			if wd.Outfit == i {
				wi = k
			}
		}
		if own {
			bg = rl.NewColor(110, 170, 230, 245)
			if wi >= 0 {
				bg = mixColor(Worlds[wi].Main, rl.White, 0.35)
			}
		}
		cur := g.Prog.Outfit == i
		if g.boothCard(r, bg, cur) {
			oi := i
			b.Action = func() {
				if !g.boothWear(oi) {
					b.Shake = 0.4
					g.Play(g.Audio.Error)
				}
			}
		}
		tint := rl.White
		if !own {
			tint = rl.NewColor(20, 24, 40, 255) // umbra
		}
		drawRT(b.rtChar[i], rl.NewRectangle(r.X+4, r.Y+4, r.Width-8, (r.Width-8)*1.25), tint)
		if !own {
			drawIcon(IconLock, r.X+r.Width/2, r.Y+r.Height/2, r.Width*0.32, rl.White)
			if wi >= 0 {
				rl.DrawCircleV(vv(r.X+r.Width-24, r.Y+24), 14, Worlds[wi].Main)
				rl.DrawCircleLines(int32(r.X+r.Width-24), int32(r.Y+24), 14, rl.White)
			}
		}
		if cur {
			rl.DrawCircleV(vv(r.X+26, r.Y+26), 18, UIGreen)
			drawIcon(IconCheck, r.X+26, r.Y+26, 24, rl.White)
		}
		if own {
			g.hint(Outfits[i].Name, r.X+r.Width/2, r.Y+r.Height-28, 200)
		}
	}
}

func (g *Game) drawBoothGuns(W, H float32) {
	b := &g.Booth
	cols, rows := 3, 2
	cw := minf((W-160)/float32(cols), (H-220)/float32(rows)/0.85)
	gap := cw * 0.08
	cw -= gap
	ch := cw * 0.85
	x0 := (W - (float32(cols)*cw + float32(cols-1)*gap)) / 2
	y0 := (H-(float32(rows)*ch+float32(rows-1)*gap))/2 + 10
	stars := g.totalStars()
	// cate stele are Aventurierul (sus, langa telefon)
	drawStar(130, 52, 26, LegoGoldBright)
	g.text(itoa(stars), 162, 34, 36, rl.White, true)
	for k := 0; k < NumGuns; k++ {
		r := rl.NewRectangle(x0+float32(k%cols)*(cw+gap), y0+float32(k/cols)*(ch+gap), cw, ch)
		ok := g.gunUnlocked(k)
		bg := rl.NewColor(60, 70, 100, 235)
		if ok {
			bg = rl.NewColor(140, 120, 220, 245)
		}
		cur := g.Prog.Gun == k
		if g.boothCard(r, bg, cur) {
			kk := k
			b.Action = func() {
				if !g.boothGun(kk) {
					b.Shake = 0.4
					g.Play(g.Audio.Error)
				}
			}
		}
		tint := rl.White
		if !ok {
			tint = rl.NewColor(30, 34, 55, 255)
		}
		drawRT(b.rtGun[k], rl.NewRectangle(r.X+6, r.Y+6, r.Width-12, (r.Width-12)*0.75), tint)
		if !ok {
			drawIcon(IconLock, r.X+r.Width/2, r.Y+r.Height*0.42, r.Width*0.22, rl.White)
			drawStar(r.X+r.Width/2-14, r.Y+r.Height-44, 20, LegoGoldBright)
			g.text(itoa(GunStars[k]), r.X+r.Width/2+10, r.Y+r.Height-60, 30, rl.White, true)
		} else {
			g.hint(GunNames[k], r.X+r.Width/2, r.Y+r.Height-32, 200)
		}
		if cur {
			rl.DrawCircleV(vv(r.X+26, r.Y+26), 18, UIGreen)
			drawIcon(IconCheck, r.X+26, r.Y+26, 24, rl.White)
		}
	}
}

// drawBoothRealms: lumile mari. Cele deschise sunt colorate, cu eroul lor;
// apesi pe una si trenul magic te duce acolo.
func (g *Game) drawBoothRealms(W, H float32) {
	b := &g.Booth
	n := len(Realms)
	cols := 3
	rows := (n + cols - 1) / cols
	cw := minf((W-160)/float32(cols), (H-200)/float32(rows)/1.05)
	gap := cw * 0.08
	cw -= gap
	ch := cw * 1.05
	x0 := (W - (float32(cols)*cw + float32(cols-1)*gap)) / 2
	y0 := (H-(float32(rows)*ch+float32(rows-1)*gap))/2 + 10
	for i := 0; i < n; i++ {
		R := &Realms[i]
		r := rl.NewRectangle(x0+float32(i%cols)*(cw+gap), y0+float32(i/cols)*(ch+gap), cw, ch)
		open := g.realmOpen(i)
		bg := rl.NewColor(60, 70, 100, 235)
		if open {
			bg = mixColor(R.Main, rl.White, 0.3)
		}
		here := i == curRealm
		if g.boothCard(r, bg, here) {
			ri := i
			b.Action = func() {
				switch {
				case !open:
					b.Shake = 0.4
					g.Play(g.Audio.Error)
				case ri == curRealm:
					g.Play(g.Audio.Click)
				default:
					g.closeBooth()
					g.travelNow(ri)
				}
			}
		}
		oi := outfitOfRealm(i)
		tint := rl.White
		if !open {
			tint = rl.NewColor(20, 24, 40, 255)
		}
		if oi >= 0 && oi < len(b.rtChar) {
			s := r.Height * 0.78
			drawRT(b.rtChar[oi], rl.NewRectangle(r.X+r.Width/2-s*0.4, r.Y+6, s*0.8, s), tint)
		}
		if !open {
			drawIcon(IconLock, r.X+r.Width/2, r.Y+r.Height*0.45, r.Width*0.3, rl.White)
		}
		if here {
			rl.DrawCircleV(vv(r.X+26, r.Y+26), 18, UIGreen)
			drawIcon(IconCheck, r.X+26, r.Y+26, 24, rl.White)
		}
		// stelele lumii (cate s-au strans)
		st := 0
		if i < len(g.RealmStates) {
			st = g.RealmStates[i].Stars
		}
		if i == curRealm {
			st = g.realmStars()
		}
		if open {
			drawStar(r.X+r.Width-30, r.Y+28, 16, LegoGoldBright)
			g.textCenter(itoa(st), r.X+r.Width-30, r.Y+46, 18, rl.White, true)
		}
		g.hint(R.Name, r.X+r.Width/2, r.Y+r.Height-26, 210)
	}
}
