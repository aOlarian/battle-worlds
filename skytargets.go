package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// TINTE IN MARE SI PE CER: barcile, avioanele si pescarusii se pot
// impusca (dau aur). Laserul zboara repede si urmareste tinta, ca sa
// ajunga si la avionul de sus. Nimeni nu patimeste: pilotul avionului
// coboara cu parasuta, pescarusul face „puf” si se intoarce mai tarziu.
// ====================================================================

const (
	TgtNone  = 0
	TgtBoat  = 1
	TgtPlane = 2
	TgtBird  = 3

	BoatMaxHP    = float32(4)
	PlaneMaxHP   = float32(6)
	BoatRespawn  = float32(10)
	PlaneRespawn = float32(16)
	BirdRespawn  = float32(12)
	boatSinkTime = float32(3)
	laserFast    = float32(150) // viteza laserului spre tintele departate
)

type Plane struct {
	CX, CZ, RX, RZ float32
	H              float32
	Ang, AngSpd    float32
	Pos            rl.Vector3
	Yaw, Roll      float32
	HP, Dead       float32
	Flash          float32
	Timer          float32
}

type Flock struct {
	C       rl.Vector2
	R, H    float32
	Ang     float32
	Spd     float32
	Yaw     float32
	Center  rl.Vector3
	Phase   float32
	Squawk  float32
	visible bool
}

type Bird struct {
	Flock int
	Off   rl.Vector3 // locul in formatia in "V"
	Phase float32
	Pos   rl.Vector3
	Yaw   float32
	Dead  float32
}

type Parachute struct {
	Pos    rl.Vector3
	Yaw    float32
	Active bool
}

// ---------------- construire ----------------

func (cw *CityWorld) buildSkyTargets() {
	cw.Planes = []Plane{
		// avionul de jos: trece des peste oras, usor de nimerit
		{CX: 0, CZ: 0, RX: 210, RZ: 140, H: 40, Ang: 0.5, AngSpd: 0.17},
		// avionul de sus: ocoleste toata harta
		{CX: 0, CZ: 0, RX: 380, RZ: 240, H: 66, Ang: 3.4, AngSpd: -0.1},
	}
	for i := range cw.Planes {
		cw.Planes[i].HP = PlaneMaxHP
	}
	flocks := []Flock{
		{C: rl.NewVector2(0, 0), R: 46, H: 22},
		{C: rl.NewVector2(Sectors[0].Center.X, Sectors[0].Center.Y), R: 42, H: 26},
		{C: rl.NewVector2(Sectors[1].Center.X, Sectors[1].Center.Y), R: 40, H: 24},
		{C: rl.NewVector2(Sectors[2].Center.X, Sectors[2].Center.Y), R: 44, H: 28},
		{C: rl.NewVector2(Sectors[3].Center.X, Sectors[3].Center.Y), R: 48, H: 34},
		{C: rl.NewVector2(-366, 0), R: 40, H: 16},
		{C: rl.NewVector2(360, -20), R: 50, H: 30},
		{C: rl.NewVector2(-160, -110), R: 60, H: 14},
		{C: rl.NewVector2(170, 110), R: 60, H: 14},
	}
	for fi := range flocks {
		f := &flocks[fi]
		f.Ang = cw.Rng.Float32() * 6.28
		f.Spd = (7 + cw.Rng.Float32()*3) / f.R
		if fi%2 == 1 {
			f.Spd = -f.Spd
		}
		f.Phase = cw.Rng.Float32() * 6.28
		n := 5
		for k := 0; k < n; k++ {
			// formatie in V: primul in fata, ceilalti in spate, pe laturi
			row := float32((k + 1) / 2)
			side := float32(1)
			if k%2 == 0 {
				side = -1
			}
			off := v3(side*row*2.6, (cw.Rng.Float32()-0.5)*1.2, -row*2.4)
			cw.Birds = append(cw.Birds, Bird{Flock: fi, Off: off, Phase: cw.Rng.Float32() * 6.28})
		}
	}
	cw.Flocks = flocks
	cw.BirdBody, cw.BirdWingL, cw.BirdWingR = buildBirdMeshes()
	cw.ChuteMesh = buildChuteMesh()
}

func buildBirdMeshes() (rl.Mesh, rl.Mesh, rl.Mesh) {
	white := rl.NewColor(250, 250, 246, 255)
	grey := rl.NewColor(150, 160, 176, 255)
	body := NewMB()
	body.Ellipsoid(v3(0, 0, 0), v3(0.36, 0.34, 0.78), 6, 10, white)
	body.Ellipsoid(v3(0, 0.22, 0.66), v3(0.27, 0.26, 0.3), 6, 8, white)
	body.CylinderC(v3(0, 0.18, 0.92), v3(0, 0.12, 1.28), 0.09, 0.02, 6, LegoOrange, LegoOrange, true)
	body.Ellipsoid(v3(0.17, 0.3, 0.82), v3(0.06, 0.07, 0.05), 3, 5, rl.NewColor(30, 30, 36, 255))
	body.Ellipsoid(v3(-0.17, 0.3, 0.82), v3(0.06, 0.07, 0.05), 3, 5, rl.NewColor(30, 30, 36, 255))
	body.Box(v3(0, 0.02, -0.82), v3(0.5, 0.06, 0.4), grey)
	wing := func(dir float32) rl.Mesh {
		mb := NewMB()
		mb.Box(v3(dir*0.75, 0, 0), v3(1.5, 0.08, 0.62), white)
		mb.Box(v3(dir*1.7, 0, -0.08), v3(0.6, 0.07, 0.48), grey)
		return mb.Build()
	}
	return body.Build(), wing(1), wing(-1)
}

func buildChuteMesh() rl.Mesh {
	mb := NewMB()
	cols := []rl.Color{LegoRed, LegoWhite, LegoYellow, LegoWhite}
	// cupola din felii colorate
	n := 12
	for k := 0; k < n; k++ {
		a0 := float32(k) / float32(n) * 2 * math.Pi
		a1 := float32(k+1) / float32(n) * 2 * math.Pi
		top := v3(0, 6.4, 0)
		p0 := v3(cosf(a0)*3.2, 4.9, sinf(a0)*3.2)
		p1 := v3(cosf(a1)*3.2, 4.9, sinf(a1)*3.2)
		q0 := v3(cosf(a0)*3.6, 4.0, sinf(a0)*3.6)
		q1 := v3(cosf(a1)*3.6, 4.0, sinf(a1)*3.6)
		c := cols[k%len(cols)]
		mb.Tri(top, p1, p0, v3(cosf(a0), 1, sinf(a0)), c)
		mb.Tri(top, p0, p1, v3(cosf(a0), -1, sinf(a0)), c)
		mb.Quad(p0, p1, q1, q0, v3(cosf(a0), 0.4, sinf(a0)), c)
		mb.Quad(q0, q1, p1, p0, v3(-cosf(a0), -0.4, -sinf(a0)), c)
		if k%3 == 0 {
			mb.Cylinder(q0, v3(0, 1.9, 0), 0.03, 0.03, 3, LegoWhite, false)
		}
	}
	return mb.Build()
}

// ---------------- miscare ----------------

func (cw *CityWorld) updateSkyTargets(dt, t float32, playerPos, listener rl.Vector3, audio *GameAudio) {
	for i := range cw.Planes {
		pl := &cw.Planes[i]
		pl.Ang += pl.AngSpd * dt
		if pl.Flash > 0 {
			pl.Flash -= dt
		}
		if pl.Dead > 0 {
			pl.Dead -= dt
			if pl.Dead <= 0 {
				pl.HP = PlaneMaxHP
				pl.Ang += math.Pi // revine din partea cealalta a hartii
			}
		}
		a := pl.Ang
		pl.Pos = v3(pl.CX+cosf(a)*pl.RX, pl.H+sinf(a*2)*4, pl.CZ+sinf(a)*pl.RZ)
		sgn := sign32(pl.AngSpd)
		dx, dz := -sinf(a)*pl.RX*sgn, cosf(a)*pl.RZ*sgn
		pl.Yaw = atan2f(dx, dz)
		pl.Roll = -0.22 * sgn
		pl.Timer += dt
		if pl.Dead <= 0 && audio != nil && pl.Timer > 12 && rl.Vector3Distance(pl.Pos, playerPos) < 170 {
			audio.PlayAt(audio.Airplane, pl.Pos, listener, 1, 240)
			pl.Timer = 0
		}
	}
	if cw.Chute.Active {
		cw.Chute.Pos.Y -= dt * 3.2
		cw.Chute.Pos.X += sinf(t*0.7) * dt * 0.8
		cw.Chute.Yaw += dt * 0.3
		if g, _ := cw.W.GroundAt(cw.Chute.Pos.X, cw.Chute.Pos.Z, cw.Chute.Pos.Y); cw.Chute.Pos.Y < maxf(g, WaterY) {
			cw.Chute.Active = false
		}
	}
	for fi := range cw.Flocks {
		f := &cw.Flocks[fi]
		f.Ang += f.Spd * dt
		cx := f.C.X + cosf(f.Ang)*f.R
		cz := f.C.Y + sinf(f.Ang)*f.R
		f.Center = v3(cx, f.H+sinf(t*0.5+f.Phase)*2.5, cz)
		sgn := sign32(f.Spd)
		f.Yaw = atan2f(-sinf(f.Ang)*sgn, cosf(f.Ang)*sgn)
	}
	for i := range cw.Birds {
		b := &cw.Birds[i]
		if b.Dead > 0 {
			b.Dead -= dt
		}
		f := &cw.Flocks[b.Flock]
		sy, cy := sinf(f.Yaw), cosf(f.Yaw)
		ox := b.Off.X*cy + b.Off.Z*sy
		oz := -b.Off.X*sy + b.Off.Z*cy
		b.Pos = v3(f.Center.X+ox, f.Center.Y+b.Off.Y+sinf(t*1.3+b.Phase)*0.4, f.Center.Z+oz)
		b.Yaw = f.Yaw
	}
}

func (b *Boat) Hittable() bool { return b.Dead <= 0 }

// ---------------- desen ----------------

func (cw *CityWorld) drawSkyTargets(r *Renderer, draw DrawFn, t float32, shadow bool, near func(rl.Vector3, float32) bool) {
	for i := range cw.Planes {
		pl := &cw.Planes[i]
		if pl.Dead > 0 {
			continue
		}
		// avioanele se vad si de departe (nu doar pana la 300 m)
		if shadow || !r.SphereVisible(pl.Pos, 14) {
			continue
		}
		m := rl.MatrixMultiply(rl.MatrixRotateZ(pl.Roll), rl.MatrixRotateY(pl.Yaw))
		m = rl.MatrixMultiply(m, rl.MatrixTranslate(pl.Pos.X, pl.Pos.Y, pl.Pos.Z))
		tint := rl.White
		if pl.HP <= PlaneMaxHP*0.5 {
			tint = rl.NewColor(170, 160, 155, 255)
		}
		if pl.Flash > 0 {
			r.SetFlash(0.5)
		}
		draw(cw.PlaneMesh, m, tint)
		if pl.Flash > 0 {
			r.SetFlash(0)
		}
	}
	if cw.Chute.Active && near(cw.Chute.Pos, 8) {
		c := &cw.Chute
		draw(cw.ChuteMesh, rl.MatrixMultiply(rl.MatrixRotateY(c.Yaw), rl.MatrixTranslate(c.Pos.X, c.Pos.Y, c.Pos.Z)), rl.White)
		DrawCharacter(draw, cw.CitizenMeshes[3%len(cw.CitizenMeshes)], c.Pos, c.Yaw, 0.85, Pose{HandsUp: true}, rl.White)
	}
	for fi := range cw.Flocks {
		cw.Flocks[fi].visible = near(cw.Flocks[fi].Center, 14)
	}
	for i := range cw.Birds {
		b := &cw.Birds[i]
		if b.Dead > 0 || !cw.Flocks[b.Flock].visible {
			continue
		}
		base := rl.MatrixMultiply(rl.MatrixRotateY(b.Yaw), rl.MatrixTranslate(b.Pos.X, b.Pos.Y, b.Pos.Z))
		draw(cw.BirdBody, base, rl.White)
		flap := sinf(t*7+b.Phase) * 0.65
		wl := rl.MatrixMultiply(rl.MatrixRotateZ(flap), rl.MatrixTranslate(0.3, 0.08, 0.05))
		wr := rl.MatrixMultiply(rl.MatrixRotateZ(-flap), rl.MatrixTranslate(-0.3, 0.08, 0.05))
		draw(cw.BirdWingL, rl.MatrixMultiply(wl, base), rl.White)
		draw(cw.BirdWingR, rl.MatrixMultiply(wr, base), rl.White)
	}
}

// ---------------- tintire ----------------

// boatFrame: punctul (x,z) in sistemul barcii.
func boatFrame(b *Boat, x, z float32) (along, side float32) {
	dx, dz := x-b.Pos.X, z-b.Pos.Z
	sy, cy := sinf(b.Yaw), cosf(b.Yaw)
	return dx*sy + dz*cy, dx*cy - dz*sy
}

// thingCenter: centrul tintei (unde zboara laserul).
func (g *Game) thingCenter(kind, idx int) (rl.Vector3, bool) {
	cw := g.City
	switch kind {
	case TgtBoat:
		if idx >= 0 && idx < len(cw.Boats) && cw.Boats[idx].Hittable() {
			b := &cw.Boats[idx]
			return v3(b.Pos.X, b.Pos.Y+1.1, b.Pos.Z), true
		}
	case TgtPlane:
		if idx >= 0 && idx < len(cw.Planes) && cw.Planes[idx].Dead <= 0 {
			return cw.Planes[idx].Pos, true
		}
	case TgtBird:
		if idx >= 0 && idx < len(cw.Birds) && cw.Birds[idx].Dead <= 0 {
			return cw.Birds[idx].Pos, true
		}
	}
	return rl.Vector3{}, false
}

// aimThings: barca, avionul sau pescarusul din mijlocul ecranului (cu tinta "lipicioasa").
// hitD = unde loveste raza un perete/sol (maxD = nimic in cale, deci cer liber).
func (g *Game) aimThings(camPos, dir rl.Vector3, hitD, maxD float32) {
	g.AimKind, g.AimIdx = TgtNone, -1
	best := float32(1e9)
	try := func(kind, idx int, ctr rl.Vector3, rad, grow, far float32) {
		to := rl.Vector3Subtract(ctr, camPos)
		along := rl.Vector3DotProduct(to, dir)
		if along < 2 || along > far {
			return
		}
		if hitD < maxD && along > hitD+rad {
			return // e ceva in cale
		}
		perp := rl.Vector3Length(rl.Vector3Subtract(to, rl.Vector3Scale(dir, along)))
		allow := rad + along*grow
		if perp < allow {
			if sc := perp/allow + along*0.0005; sc < best {
				best = sc
				g.AimKind, g.AimIdx = kind, idx
			}
		}
	}
	cw := g.City
	for i := range cw.Planes {
		if c, ok := g.thingCenter(TgtPlane, i); ok {
			try(TgtPlane, i, c, 9, 0.03, 650)
		}
	}
	for i := range cw.Birds {
		if c, ok := g.thingCenter(TgtBird, i); ok {
			try(TgtBird, i, c, 1.6, 0.05, 220)
		}
	}
	for i := range cw.Boats {
		if c, ok := g.thingCenter(TgtBoat, i); ok {
			try(TgtBoat, i, c, 3.2, 0.035, 260)
		}
	}
}

// thingAt: proiectilul (in p) a lovit o barca / un avion / un pescarus?
func (g *Game) thingAt(p rl.Vector3) (int, int) {
	cw := g.City
	for i := range cw.Planes {
		pl := &cw.Planes[i]
		if pl.Dead > 0 {
			continue
		}
		if rl.Vector3Distance(p, pl.Pos) < 7 {
			return TgtPlane, i
		}
	}
	for i := range cw.Birds {
		b := &cw.Birds[i]
		if b.Dead <= 0 && rl.Vector3Distance(p, b.Pos) < 1.4 {
			return TgtBird, i
		}
	}
	for i := range cw.Boats {
		b := &cw.Boats[i]
		if !b.Hittable() || absf(p.X-b.Pos.X) > 5 || absf(p.Z-b.Pos.Z) > 5 {
			continue
		}
		dy := p.Y - b.Pos.Y
		if dy < -0.9 || dy > 3.8 {
			continue
		}
		a, s := boatFrame(b, p.X, p.Z)
		if a > -3.0 && a < 3.6 && absf(s) < 1.6 {
			return TgtBoat, i
		}
	}
	return TgtNone, -1
}

// hitThing: laserul a nimerit barca / avionul / pescarusul.
func (g *Game) hitThing(kind, idx int, at rl.Vector3) {
	cw := g.City
	switch kind {
	case TgtBoat:
		b := &cw.Boats[idx]
		b.HP -= 1
		b.Flash = 0.14
		g.addGold(5, v3(b.Pos.X, b.Pos.Y+3, b.Pos.Z))
		g.PlayAt(g.Audio.ShotHit, b.Pos, 0.9, 120)
		g.PlayAt(g.Audio.Clack, b.Pos, 0.7, 60)
		g.spray(at, 10, []rl.Color{LegoYellow, LegoWhite, LegoOrange}, 5)
		if b.HP <= 0 {
			b.Dead = BoatRespawn
			g.addGold(40, v3(b.Pos.X, b.Pos.Y+4.5, b.Pos.Z))
			g.PlayAt(g.Audio.Boom, b.Pos, 1, 220)
			g.PlayAt(g.Audio.Splash, b.Pos, 1, 120)
			g.splash(v3(b.Pos.X, WaterY, b.Pos.Z), 40, 4)
			g.spray(v3(b.Pos.X, b.Pos.Y+1.2, b.Pos.Z), 30, []rl.Color{rl.NewColor(255, 200, 60, 255), rl.NewColor(255, 120, 30, 255), LegoWhite}, 12)
			g.spray(v3(b.Pos.X, b.Pos.Y+1.2, b.Pos.Z), 10, []rl.Color{boatColors[b.Mesh%len(boatColors)]}, 14)
			g.Achieve.Unlock("boat_sink")
			if g.Debug != nil {
				g.Debug.ThingsHit[TgtBoat]++
			}
		}
	case TgtPlane:
		pl := &cw.Planes[idx]
		pl.HP -= 1
		pl.Flash = 0.14
		g.addGold(10, v3(pl.Pos.X, pl.Pos.Y+4, pl.Pos.Z))
		g.PlayAt(g.Audio.ShotHit, pl.Pos, 1, 400)
		g.spray(at, 14, []rl.Color{LegoYellow, LegoWhite, LegoCyan}, 8)
		if pl.HP <= 0 {
			pl.Dead = PlaneRespawn
			g.addGold(100, v3(pl.Pos.X, pl.Pos.Y+6, pl.Pos.Z))
			g.PlayAt(g.Audio.Boom, pl.Pos, 1, 600)
			if d := rl.Vector3Distance(pl.Pos, g.Player.Pos); d < 120 {
				g.CamShake = maxf(g.CamShake, 0.6*(1-d/120))
			}
			g.spray(pl.Pos, 60, []rl.Color{rl.NewColor(255, 200, 60, 255), rl.NewColor(255, 120, 30, 255), rl.NewColor(255, 240, 160, 255)}, 22)
			g.spray(pl.Pos, 30, []rl.Color{rl.NewColor(80, 80, 86, 200)}, 8)
			g.spray(pl.Pos, 24, []rl.Color{LegoWhite, LegoBlue, LegoRed}, 18)
			// pilotul coboara cu parasuta (nu patimeste nimeni)
			cw.Chute = Parachute{Pos: v3(pl.Pos.X, pl.Pos.Y-2, pl.Pos.Z), Yaw: pl.Yaw, Active: true}
			g.Achieve.Unlock("plane_boom")
			if g.Debug != nil {
				g.Debug.ThingsHit[TgtPlane]++
			}
		}
	case TgtBird:
		b := &cw.Birds[idx]
		b.Dead = BirdRespawn
		g.addGold(10, v3(b.Pos.X, b.Pos.Y+1.5, b.Pos.Z))
		g.PlayAt(g.Audio.ShotHit, b.Pos, 1, 160)
		// „puf”: un nor de pene
		g.spray(b.Pos, 26, []rl.Color{LegoWhite, rl.NewColor(230, 234, 240, 255), rl.NewColor(150, 160, 176, 255)}, 5)
		g.Achieve.Unlock("bird_hunter")
		if g.Debug != nil {
			g.Debug.ThingsHit[TgtBird]++
		}
	}
}

var boatColors = []rl.Color{LegoRed, LegoBlue, LegoYellow, LegoGreen, LegoOrange}

// spray: particule in toate directiile.
func (g *Game) spray(at rl.Vector3, n int, cols []rl.Color, speed float32) {
	for k := 0; k < n; k++ {
		v := v3((rand.Float32()-0.5)*speed, rand.Float32()*speed*0.7+speed*0.15, (rand.Float32()-0.5)*speed)
		g.spawnParticle(at, v, cols[k%len(cols)], 0.22+rand.Float32()*0.2, 0.6+rand.Float32()*0.7)
	}
}

// aimThingHP: viata tintei de pe cer / din mare (pentru bara de sub tinta), -1 = fara bara.
func (g *Game) aimThingHP() float32 {
	switch g.AimKind {
	case TgtBoat:
		return clampf(g.City.Boats[g.AimIdx].HP/BoatMaxHP, 0, 1)
	case TgtPlane:
		return clampf(g.City.Planes[g.AimIdx].HP/PlaneMaxHP, 0, 1)
	}
	return -1
}
