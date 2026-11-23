package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 5.0: PIESELE AURII ascunse (5 pe fiecare lume = steaua 3).
// Una pe un acoperis (se ajunge cu trambulina), una in tunelul lumii,
// restul pe jos, departe de drumuri. Sclipesc si suna cand esti aproape.
// ====================================================================

var goldBrickPos [MaxZones][BricksPerWorld]rl.Vector3
var goldBrickMesh rl.Mesh
var goldBricksReady bool

func buildGoldBrickMesh() {
	mb := NewMB()
	mb.Emis = 1
	gold := rl.NewColor(255, 200, 40, 255)
	mb.BeveledBox(v3(0, 0, 0), v3(1.3, 0.62, 1.3), 0.08, gold)
	for _, x := range []float32{-0.32, 0.32} {
		for _, z := range []float32{-0.32, 0.32} {
			mb.Cylinder(v3(x, 0.31, z), v3(x, 0.5, z), 0.2, 0.2, 12, rl.NewColor(255, 225, 90, 255), true)
		}
	}
	mb.Emis = 0
	goldBrickMesh = mb.Build()
}

func (g *Game) placeGoldBricks() {
	w := g.City.W
	rng := rand.New(rand.NewSource(5050))
	for wi, wd := range Worlds {
		si := wd.Sector
		c := Sectors[si].Center
		placed := []rl.Vector3{}
		far := func(p rl.Vector3, d float32) bool {
			for _, q := range placed {
				if sq(p.X-q.X)+sq(p.Z-q.Z) < d*d {
					return false
				}
			}
			return true
		}
		// 1) pe un acoperis langa o trambulina din lume
		bestTop := float32(0)
		var roof rl.Vector3
		for _, pd := range w.Pads {
			if SectorAt(pd.Pos.X, pd.Pos.Z) != si {
				continue
			}
			for _, b := range w.Boxes {
				cx, cz := (b.MinX+b.MaxX)/2, (b.MinZ+b.MaxZ)/2
				if b.Roof == RoofNone || b.Top < 6 || b.Top > 45 || sq(cx-pd.Pos.X)+sq(cz-pd.Pos.Z) > 18*18 {
					continue
				}
				if b.MaxX-b.MinX < 4 || b.MaxZ-b.MinZ < 4 {
					continue
				}
				if b.Top > bestTop {
					bestTop = b.Top
					roof = v3(cx, b.SurfaceY(cx, cz)+0.9, cz)
				}
			}
		}
		if bestTop > 0 {
			placed = append(placed, roof)
		}
		// 2) in tunelul lumii (muntele ei, sau Marele Tunel pentru plaja / parc)
		switch {
		case curRealm > 0:
		case si < 4:
			m := &w.Mountains[si]
			x, z := m.World((PortalOut-PortalIn)/2, 0)
			gy, _ := w.GroundAt(x, z, 3)
			placed = append(placed, v3(x, gy+0.9, z))
		case si == SecWest || si == SecPark:
			sx := float32(-1)
			if si == SecPark {
				sx = 1
			}
			gt := &w.Corridors[0]
			s, _ := gt.Local(sx*170, 0)
			x, z := gt.Point(s, 3.5)
			placed = append(placed, v3(x, gt.FloorC(s)+0.9, z))
		}
		// 3) restul pe jos, pe insula
		rad := float32(54)
		switch si {
		case 4:
			rad = 52
		case SecWest, SecPark:
			rad = 62
		}
		if curRealm > 0 {
			rad = realmZoneR(si) - 6
		}
		for tries := 0; tries < 4000 && len(placed) < BricksPerWorld; tries++ {
			a := rng.Float32() * 2 * math.Pi
			d := 10 + rng.Float32()*(rad-10)
			x := c.X + cosf(a)*d
			z := c.Y + sinf(a)*d
			if w.LandSD(x, z) > -9 || w.onRoad(x, z) || !w.CanStand(x, z, 0.4, 0.9) {
				continue
			}
			gy, k := w.GroundAt(x, z, 1)
			if k == SurfWater || absf(gy) > 1.5 || SectorAt(x, z) != si {
				continue
			}
			p := v3(x, gy+0.9, z)
			if !far(p, 16) {
				continue
			}
			placed = append(placed, p)
		}
		for i := 0; i < BricksPerWorld; i++ {
			if i < len(placed) {
				goldBrickPos[wi][i] = placed[i]
			} else {
				goldBrickPos[wi][i] = v3(c.X+float32(i)*3, 0.9, c.Y+20)
			}
		}
	}
	goldBricksReady = true
}

func (g *Game) updateBricks(dt float32) {
	if !goldBricksReady {
		return
	}
	p := g.Player.Pos
	for wi := range Worlds {
		if !g.worldOpen(wi) {
			continue
		}
		for i, b := range goldBrickPos[wi] {
			if g.Prog.Bricks[wi][i] {
				continue
			}
			d := rl.Vector3Distance(v3(p.X, p.Y+1, p.Z), b)
			if d < 2.2 {
				g.Prog.Bricks[wi][i] = true
				g.Play(g.Audio.Coin)
				g.Play(g.Audio.Achievement)
				n := g.bricksFound(wi)
				g.sayIcon(IconGoldBrick, "Piesă aurie! "+itoa(n)+" din 5")
				for k := 0; k < 40; k++ {
					a := float32(k) * 0.5
					g.spawnParticle(b, v3(cosf(a)*4, 3+rand.Float32()*5, sinf(a)*4), []rl.Color{LegoGoldBright, LegoWhite}[k%2], 0.22, 1.3)
				}
				g.Score += 500
			}
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

func (g *Game) drawGoldBricks() {
	if !goldBricksReady {
		return
	}
	for wi := range Worlds {
		if !g.worldOpen(wi) {
			continue
		}
		for i, b := range goldBrickPos[wi] {
			if g.Prog.Bricks[wi][i] || !g.R.SphereVisible(b, 2) {
				continue
			}
			if sq(b.X-g.Camera.Position.X)+sq(b.Z-g.Camera.Position.Z) > 160*160 {
				continue
			}
			y := b.Y + sinf(g.Time*2.2+float32(i))*0.25
			m := rl.MatrixMultiply(rl.MatrixRotateY(g.Time*1.6+float32(i)), rl.MatrixTranslate(b.X, y, b.Z))
			g.R.DrawLit(goldBrickMesh, m, rl.White)
		}
	}
}

// sclipiri + stalp scurt de lumina (fara iluminare)
func (g *Game) drawGoldBrickGlow() {
	if !goldBricksReady {
		return
	}
	rl.DisableDepthMask()
	for wi := range Worlds {
		if !g.worldOpen(wi) {
			continue
		}
		for i, b := range goldBrickPos[wi] {
			if g.Prog.Bricks[wi][i] {
				continue
			}
			d2 := sq(b.X-g.Camera.Position.X) + sq(b.Z-g.Camera.Position.Z)
			if d2 > 120*120 || !g.R.SphereVisible(b, 6) {
				continue
			}
			pulse := 0.5 + 0.5*sinf(g.Time*4+float32(i))
			rl.DrawCylinder(v3(b.X, b.Y-0.6, b.Z), 0.5, 0.05, 6, 8, rl.NewColor(255, 220, 80, uint8(40+50*pulse)))
			for k := 0; k < 4; k++ {
				a := g.Time*2 + float32(k)*math.Pi/2 + float32(i)
				s := v3(b.X+cosf(a)*1.1, b.Y+0.4+sinf(g.Time*3+float32(k))*0.4, b.Z+sinf(a)*1.1)
				rl.DrawCube(s, 0.16, 0.16, 0.16, rl.NewColor(255, 255, 220, uint8(150+100*pulse)))
			}
		}
	}
	rl.EnableDepthMask()
}

// sunet discret cand Aventurierul e aproape de o piesa
func (g *Game) brickChime(dt float32) {
	if !goldBricksReady {
		return
	}
	g.brickChimeT -= dt
	if g.brickChimeT > 0 {
		return
	}
	p := g.Player.Pos
	for wi := range Worlds {
		if !g.worldOpen(wi) {
			continue
		}
		for i, b := range goldBrickPos[wi] {
			if g.Prog.Bricks[wi][i] {
				continue
			}
			if d := rl.Vector3Distance(p, b); d < 22 {
				g.PlayAt(g.Audio.Coin, b, 0.35, 30)
				g.brickChimeT = 0.7 + d/14
				return
			}
		}
	}
	g.brickChimeT = 0.5
}
