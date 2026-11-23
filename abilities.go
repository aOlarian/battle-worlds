package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 6.0: ABILITATILE EROILOR (le are Aventurierul cand poarta costumul eroului)
//   Surferul        - pe apa sta pe placa de surf si merge de 2 ori mai repede
//   Micul Vampir    - in aer, SPATIU tinut = planeaza (cade incet)
//   Baiatul-Paianjen- Q = se trage cu panza spre locul tintit (acoperisuri!);
//                     pistolul trage ghemotoace de panza care opresc monstrii
//   Eroul Zburator  - SPATIU tinut = zboara in sus; in aer se misca repede
//   Spiridusul      - alearga mai repede (aluneca pe zapada); arunca bulgari
//   Iepurasul       - sare de 2 ori mai sus
// ====================================================================

type ZipState struct {
	Active bool
	Target rl.Vector3
	T      float32
}

func (g *Game) ability() int {
	oi := g.Prog.Outfit
	if oi < 0 || oi >= len(Outfits) {
		return AbilNone
	}
	return Outfits[oi].Ability
}

// projKind: cum arata glontul pistolului (dupa eroul purtat).
func (g *Game) projKind() int {
	oi := g.Prog.Outfit
	if oi < 0 || oi >= len(Outfits) || Outfits[oi].Realm < 0 {
		return ProjLaser
	}
	return Realms[Outfits[oi].Realm].Proj
}

func (g *Game) abilitySpeed(onGround bool) float32 {
	switch g.ability() {
	case AbilSurf:
		if g.Swimming {
			return 2.2
		}
	case AbilSled:
		if onGround {
			return 1.3
		}
	case AbilFly:
		if !onGround {
			return 1.8
		}
	case AbilGlide:
		if !onGround {
			return 1.3
		}
	}
	return 1
}

func (g *Game) jumpVel() float32 {
	if g.ability() == AbilHop {
		return 17
	}
	return 10
}

// abilityAir: dupa gravitatie (planare, zbor).
func (g *Game) abilityAir(dt float32) {
	p := &g.Player
	switch g.ability() {
	case AbilGlide:
		if g.jumpHeld() && p.VelY < -2.2 {
			p.VelY = -2.2
		}
	case AbilFly:
		if g.jumpHeld() {
			p.VelY = minf(p.VelY+34*dt, 10)
			if p.Pos.Y > 150 {
				p.VelY = minf(p.VelY, 0)
			}
		} else if p.VelY < -7 {
			p.VelY = -7 // coboara lin, ca un erou
		}
	}
}

// flyTakeoff: eroul zburator se ridica de pe sol tinand SPATIU apasat.
func (g *Game) flyTakeoff() bool {
	p := &g.Player
	if g.ability() == AbilFly && p.OnGround && g.jumpHeld() && !g.Swimming {
		p.OnGround = false
		p.VelY = 8
		g.AirStartY = p.Pos.Y
		return true
	}
	return false
}

// surfing: Surferul pe apa sta in picioare pe placa.
func (g *Game) surfing() bool { return g.Swimming && g.ability() == AbilSurf }

// ---------------- panza (Q) ----------------

func (g *Game) updateZip(dt float32) bool {
	p := &g.Player
	z := &g.Zip
	if !z.Active {
		if g.ability() == AbilWeb && (rl.IsKeyPressed(rl.KeyQ) || rl.IsKeyPressed(rl.KeyF)) && g.Ride.Kind == RideNone && !g.Drive.Active {
			t := g.AimPoint
			d := rl.Vector3Distance(p.Pos, t)
			if d > 4 && d < 90 {
				*z = ZipState{Active: true, Target: t}
				g.Swimming = false
				g.PlayAt(g.Audio.Shoot, p.Pos, 1, 30)
			} else {
				g.Play(g.Audio.Error)
			}
		}
		return false
	}
	z.T += dt
	to := rl.Vector3Subtract(z.Target, p.Pos)
	d := rl.Vector3Length(to)
	step := minf(34*dt, d)
	if d < 1.6 || z.T > 4 {
		z.Active = false
		p.VelY = 3
		p.OnGround = false
		g.AirStartY = p.Pos.Y
		return false
	}
	dir := rl.Vector3Scale(to, 1/d)
	nx, nz := p.Pos.X+dir.X*step, p.Pos.Z+dir.Z*step
	ny := p.Pos.Y + dir.Y*step
	w := g.City.W
	if w.Solid(nx, nz, ny+0.6, 0.35) {
		// a ajuns la perete: urca pe el pana deasupra
		ny = p.Pos.Y + 20*dt
		nx, nz = p.Pos.X, p.Pos.Z
		if gy, _ := w.GroundAt(p.Pos.X+dir.X*1.2, p.Pos.Z+dir.Z*1.2, p.Pos.Y+40); gy < ny-0.5 {
			nx, nz = p.Pos.X+dir.X*1.2, p.Pos.Z+dir.Z*1.2
			ny = gy + 0.2
			z.Active = false
			p.OnGround = true
			p.VelY = 0
		}
	}
	p.Pos = v3(nx, ny, nz)
	p.VelY = 0
	p.OnGround = false
	p.Pose.Air = true
	return true
}

func (g *Game) drawZip() {
	if !g.Zip.Active {
		return
	}
	p := g.Player.Pos
	hand := v3(p.X, p.Y+1.7, p.Z)
	rl.DrawCylinderEx(hand, g.Zip.Target, 0.05, 0.05, 5, rl.NewColor(250, 250, 255, 255))
	rl.DrawSphereEx(g.Zip.Target, 0.35, 6, 6, rl.NewColor(250, 250, 255, 220))
}

// drawSurfboard: placa de surf sub Surfer, pe apa.
func (g *Game) drawSurfboard() {
	if !g.surfing() {
		return
	}
	p := &g.Player
	fx, fz := sinf(p.Yaw), cosf(p.Yaw)
	a := v3(p.Pos.X-fx*1.2, WaterY+0.05, p.Pos.Z-fz*1.2)
	b := v3(p.Pos.X+fx*1.3, WaterY+0.12, p.Pos.Z+fz*1.3)
	rl.DrawCylinderEx(a, b, 0.42, 0.25, 10, rl.NewColor(255, 150, 60, 255))
	rl.DrawCylinderEx(v3(a.X, a.Y+0.08, a.Z), v3(b.X, b.Y+0.08, b.Z), 0.12, 0.08, 6, rl.NewColor(40, 190, 220, 255))
}

// drawThemedShot: glontul pistolului in forma lumii eroului (panza, liliac, stea...).
func drawThemedShot(pr *Projectile3D) bool {
	switch pr.Kind {
	case ProjWeb:
		rl.DrawSphereEx(pr.Position, 0.32, 8, 8, rl.NewColor(250, 250, 255, 255))
		d := rl.Vector3Normalize(pr.Velocity)
		for k := 0; k < 4; k++ {
			a := float32(k) * 1.5708
			o := v3(cosf(a)*0.5, sinf(a)*0.5, 0)
			o = rl.Vector3Subtract(o, rl.Vector3Scale(d, rl.Vector3DotProduct(o, d)))
			rl.DrawCylinderEx(pr.Position, rl.Vector3Add(pr.Position, o), 0.03, 0.03, 4, rl.NewColor(250, 250, 255, 255))
		}
	case ProjBat:
		rl.DrawSphereEx(pr.Position, 0.22, 8, 8, rl.NewColor(120, 60, 170, 255))
		rl.DrawSphereEx(pr.Position, 0.42, 8, 8, rl.NewColor(190, 120, 255, 110))
	case ProjStar:
		rl.DrawSphereEx(pr.Position, 0.22, 8, 8, rl.NewColor(255, 240, 120, 255))
		rl.DrawSphereEx(pr.Position, 0.5, 8, 8, rl.NewColor(255, 210, 60, 100))
	case ProjSnow:
		rl.DrawSphereEx(pr.Position, 0.3, 8, 8, rl.NewColor(250, 252, 255, 255))
	case ProjEgg:
		cols := []rl.Color{rl.NewColor(255, 150, 200, 255), rl.NewColor(150, 220, 255, 255), rl.NewColor(255, 230, 100, 255)}
		c := cols[int(pr.Position.X*3+pr.Position.Z*7)&0x7fffffff%3]
		rl.DrawSphereEx(pr.Position, 0.3, 8, 8, c)
	case ProjCandy:
		rl.DrawSphereEx(pr.Position, 0.28, 8, 8, rl.NewColor(255, 140, 30, 255))
		rl.DrawSphereEx(pr.Position, 0.5, 8, 8, rl.NewColor(255, 200, 80, 90))
	case ProjBubble:
		rl.DrawSphereEx(pr.Position, 0.32, 10, 10, rl.NewColor(140, 220, 255, 170))
		rl.DrawSphereEx(v3(pr.Position.X-0.1, pr.Position.Y+0.1, pr.Position.Z), 0.08, 6, 6, rl.NewColor(255, 255, 255, 230))
	default:
		return false
	}
	return true
}
