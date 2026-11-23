package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// MARELE TUNEL DE SUB MARE + scarile si lifturile lui
//
// Un tunel subteran e un "coridor" drept: o axa A->B, o latime, un profil al
// podelei (un punct pe metru) si o bolta. Unde tavanul ar iesi din pamant,
// coridorul e deschis spre cer (sant cu ziduri de sprijin): acolo solul
// e sapat. Fizica: in coridor podeaua e "solul", peretii sunt locurile unde
// pamantul ramane mai sus decat poate urca Aventurierul.
// ====================================================================

const (
	CorrRoad = iota
	CorrStair
)

type Opening struct {
	S0, S1 float32
	Side   float32 // +1 = peretele din partea lat > 0
}

type Corridor struct {
	Name     string
	Kind     int
	A, B     rl.Vector2
	Dir      rl.Vector2
	Perp     rl.Vector2
	Len      float32
	HalfW    float32
	WallH    float32
	VaultH   float32
	WallFrom float32 // de unde incep peretii si bolta (scarile pornesc din tunelul mare)
	Prof     []float32
	open     []bool
	Openings []Opening
	Color    rl.Color
}

type Lift struct {
	Name   string
	Bottom rl.Vector3 // in tunel, langa usa
	Top    rl.Vector3 // pe insulita
	DoorS  float32
}

func newCorridor(name string, kind int, a, b rl.Vector2, halfW, wallH, vaultH float32, col rl.Color) Corridor {
	d := rl.Vector2Subtract(b, a)
	L := rl.Vector2Length(d)
	dir := rl.Vector2Scale(d, 1/L)
	return Corridor{Name: name, Kind: kind, A: a, B: b, Dir: dir, Perp: rl.NewVector2(-dir.Y, dir.X), Len: L,
		HalfW: halfW, WallH: wallH, VaultH: vaultH, Color: col}
}

func (c *Corridor) Local(x, z float32) (s, lat float32) {
	dx := x - c.A.X
	dz := z - c.A.Y
	return dx*c.Dir.X + dz*c.Dir.Y, dx*c.Perp.X + dz*c.Perp.Y
}

func (c *Corridor) Point(s, lat float32) (x, z float32) {
	return c.A.X + c.Dir.X*s + c.Perp.X*lat, c.A.Y + c.Dir.Y*s + c.Perp.Y*lat
}

// FloorC: podeaua pe axa la distanta s.
func (c *Corridor) FloorC(s float32) float32 {
	s = clampf(s, 0, float32(len(c.Prof)-1))
	i := int(s)
	if i >= len(c.Prof)-1 {
		return c.Prof[len(c.Prof)-1]
	}
	f := s - float32(i)
	return c.Prof[i]*(1-f) + c.Prof[i+1]*f
}

// FloorLat: podeaua cu trotuarul (tunelul de sosea are trotuare ridicate).
func (c *Corridor) FloorLat(s, lat float32) float32 {
	f := c.FloorC(s)
	if c.Kind == CorrRoad && absf(lat) > 4.5 {
		f += 0.25
	}
	return f
}

func (c *Corridor) CeilY(s, lat float32) float32 {
	k := clampf(lat/c.HalfW, -1, 1)
	return c.FloorC(s) + c.WallH + c.VaultH*sqrtf(1-k*k)
}

func (c *Corridor) IsOpen(s float32) bool {
	i := int(s)
	if i < 0 {
		i = 0
	}
	if i >= len(c.open) {
		i = len(c.open) - 1
	}
	return c.open[i]
}

// FloorAt: podeaua coridorului in (x,z), daca punctul e in interior.
func (c *Corridor) FloorAt(x, z float32) (float32, bool) {
	s, lat := c.Local(x, z)
	if s < 0 || s > c.Len || absf(lat) > c.HalfW-0.25 {
		return 0, false
	}
	return c.FloorLat(s, lat), true
}

// setProfile: podeaua dintr-o lista de puncte (s, y), netezita cu o fereastra de +-r metri.
func (c *Corridor) setProfile(pts [][2]float32, r int) {
	n := int(math.Ceil(float64(c.Len))) + 1
	lin := func(s float32) float32 {
		if s <= pts[0][0] {
			return pts[0][1]
		}
		for i := 1; i < len(pts); i++ {
			if s <= pts[i][0] {
				t := (s - pts[i-1][0]) / (pts[i][0] - pts[i-1][0])
				return pts[i-1][1] + (pts[i][1]-pts[i-1][1])*t
			}
		}
		return pts[len(pts)-1][1]
	}
	c.Prof = make([]float32, n)
	for i := 0; i < n; i++ {
		sum := float32(0)
		for k := -r; k <= r; k++ {
			sum += lin(float32(i + k))
		}
		c.Prof[i] = sum / float32(2*r+1)
	}
}

// markOpen: unde tavanul (cu placa de 0.6 m) ar iesi din pamant, coridorul e sant deschis.
func (c *Corridor) markOpen(w *World) {
	c.open = make([]bool, len(c.Prof))
	for i := range c.Prof {
		s := float32(i) + 0.5
		top := c.FloorC(s) + c.WallH + c.VaultH + 0.6
		nat := float32(1e9)
		for _, lat := range []float32{-c.HalfW - 0.5, 0, c.HalfW + 0.5} {
			x, z := c.Point(s, lat)
			nat = minf(nat, w.NaturalLand(x, z))
		}
		c.open[i] = top > nat-0.1 && s > c.WallFrom
	}
}

func (w *World) buildCorridors() {
	// Marele Tunel: din Insula Plajei (vest) pana in Parcul de Distractii (est)
	gt := newCorridor("Marele Tunel de sub Mare", CorrRoad, rl.NewVector2(-392, 0), rl.NewVector2(392, 0), 6.5, 3.2, 2.6, rl.NewColor(40, 150, 240, 255))
	px := [][2]float32{{-392, 0}, {-380, 0}, {-296, -14}, {-266, -16}, {-85, -16}, {-55, -9}, {55, -9}, {85, -16}, {266, -16}, {296, -14}, {380, 0}, {392, 0}}
	for i := range px {
		px[i][0] += 392 // x -> s
	}
	gt.setProfile(px, 10)
	w.Corridors = append(w.Corridors, gt)

	// scarile din piata centrala (iesiri spre suprafata)
	for _, x0 := range []float32{-20, 20} {
		st := newCorridor("Scările spre piață", CorrStair, rl.NewVector2(x0, 4.6), rl.NewVector2(x0, 20.8), 2.8, 2.6, 0.8, LegoGreen)
		base := w.Corridors[0].FloorC(x0+392) + 0.25
		st.WallFrom = 1.9
		st.setProfile([][2]float32{{0, base}, {1.9, base}, {1.9 + (-base)/0.65, 0}, {16.2, 0}}, 0)
		w.Corridors = append(w.Corridors, st)
		w.Corridors[0].Openings = append(w.Corridors[0].Openings, Opening{S0: x0 + 392 - 2.8, S1: x0 + 392 + 2.8, Side: 1})
	}
	for i := range w.Corridors {
		w.Corridors[i].markOpen(w)
	}
	// lifturile de pe insulitele din mare
	for _, x := range []float32{-VentX, VentX} {
		g := &w.Corridors[0]
		s := x + 392
		f := g.FloorC(s) + 0.25
		w.Lifts = append(w.Lifts, Lift{Name: "Liftul de pe insulița", Bottom: v3(x, f, 5.2), Top: v3(x, 0, 4.6), DoorS: s})
		g.Openings = append(g.Openings, Opening{S0: s - 1.3, S1: s + 1.3, Side: 1})
	}
}

// openCut: in santul deschis al unui coridor solul e podeaua lui.
func (w *World) openCut(x, z float32) (float32, bool) {
	for i := range w.Corridors {
		c := &w.Corridors[i]
		s, lat := c.Local(x, z)
		if s < 0 || s > c.Len || absf(lat) > c.HalfW || !c.IsOpen(s) {
			continue
		}
		return c.FloorLat(s, lat), true
	}
	return 0, false
}

// nearOpenCut: punctul e in santul deschis sau la mai putin de margin de el (pentru teren).
func (w *World) nearOpenCut(x, z, margin float32) bool {
	for i := range w.Corridors {
		c := &w.Corridors[i]
		s, lat := c.Local(x, z)
		if s < -margin || s > c.Len+margin || absf(lat) > c.HalfW+margin {
			continue
		}
		if c.IsOpen(clampf(s, 0, c.Len)) || c.IsOpen(clampf(s-margin, 0, c.Len)) || c.IsOpen(clampf(s+margin, 0, c.Len)) {
			return true
		}
	}
	return false
}

// corridorVolume: indexul coridorului daca punctul e in golul lui (sub pamant).
func (w *World) corridorVolume(x, z, y float32) int {
	for i := range w.Corridors {
		c := &w.Corridors[i]
		s, lat := c.Local(x, z)
		if s < 0 || s > c.Len || absf(lat) > c.HalfW-0.05 {
			continue
		}
		f := c.FloorC(s)
		if y < f-1.5 {
			continue
		}
		if c.IsOpen(s) {
			if y < f+c.WallH && f < -0.3 {
				return i
			}
			continue
		}
		if y <= c.CeilY(s, lat) {
			return i
		}
	}
	return -1
}

// corridorCovered: ca mai sus, dar doar partea acoperita (tunelul propriu-zis).
func (w *World) corridorCovered(x, z, y float32) int {
	for i := range w.Corridors {
		c := &w.Corridors[i]
		s, lat := c.Local(x, z)
		if s < c.WallFrom || s > c.Len || absf(lat) > c.HalfW || c.IsOpen(s) {
			continue
		}
		f := c.FloorC(s)
		if y > f-1.5 && y <= c.CeilY(s, lat) {
			return i
		}
	}
	return -1
}

// inCorridorBand: punctul e in tunelul subteran sau lipit de peretii lui (pana la 1 m in perete).
func (w *World) inCorridorBand(x, z, y float32) bool {
	for i := range w.Corridors {
		c := &w.Corridors[i]
		s, lat := c.Local(x, z)
		if s < -1 || s > c.Len+1 || absf(lat) > c.HalfW+1 {
			continue
		}
		ss := clampf(s, 0, c.Len)
		if y > c.FloorC(ss)-2 && y < c.FloorC(ss)+c.WallH+c.VaultH+0.5 {
			return true
		}
	}
	return false
}

// underCorridorRoof: intre bolta tunelului si suprafata pamantului (plin).
func (w *World) underCorridorRoof(x, z, y float32) bool {
	for i := range w.Corridors {
		c := &w.Corridors[i]
		s, lat := c.Local(x, z)
		if s < 0 || s > c.Len || absf(lat) > c.HalfW+0.6 || c.IsOpen(s) {
			continue
		}
		if y > c.CeilY(s, clampf(lat, -c.HalfW, c.HalfW)) && y < w.NaturalLand(x, z)-0.05 {
			return true
		}
	}
	return false
}

// ====================================================================
// GEOMETRIA
// ====================================================================

var corridorRainbow = []rl.Color{LegoRed, LegoOrange, LegoYellow, LegoLime, LegoGreen, LegoCyan, LegoBlue, LegoPurple, LegoPink}

func (cb *CityBuilder) buildCorridorMeshes() {
	w := cb.W
	for ci := range w.Corridors {
		c := &w.Corridors[ci]
		if c.Kind == CorrRoad {
			cb.roadTunnelMesh(c)
		} else {
			cb.stairMesh(c)
		}
	}
	for i := range w.Lifts {
		cb.liftMeshes(&w.Lifts[i])
	}
}

// partea de sub mare a Marelui Tunel (perete albastru cu pesti)
func underSea(x float32) bool { ax := absf(x); return ax > 92 && ax < 290 }

func (c *Corridor) wallOpen(s, side float32) bool {
	for _, o := range c.Openings {
		if o.Side == side && s > o.S0-0.01 && s < o.S1+0.01 {
			return true
		}
	}
	return false
}

func (cb *CityBuilder) roadTunnelMesh(c *Corridor) {
	w := cb.W
	hw := c.HalfW
	up := v3(0, 1, 0)
	P := func(s, lat, y float32) rl.Vector3 {
		x, z := c.Point(s, lat)
		return v3(x, y, z)
	}
	per3 := v3(c.Perp.X, 0, c.Perp.Y)
	dir3 := v3(c.Dir.X, 0, c.Dir.Y)
	yaw := atan2f(c.Dir.X, c.Dir.Y)
	const ds = float32(2)
	const chunk = float32(48)
	vaultSeg := 14
	for c0 := float32(0); c0 < c.Len; c0 += chunk {
		mb := NewMB()
		for s := c0; s < minf(c0+chunk, c.Len)-0.001; s += ds {
			s1 := minf(s+ds, c.Len)
			sm := (s + s1) / 2
			f0, f1 := c.FloorC(s), c.FloorC(s1)
			open := c.IsOpen(sm)
			xm, _ := c.Point(sm, 0)
			sea := underSea(xm)
			stripe := corridorRainbow[int(sm/64)%len(corridorRainbow)]
			if sea {
				stripe = rl.NewColor(40, 180, 240, 255) // dunga albastra = suntem sub mare (tunelul e uscat si luminat)
			}
			glow := float32(0)
			if open {
				mb.AO = 1
			} else {
				mb.AO = 0.22
				glow = 0.42 + 0.1*cosf(2*math.Pi*sm/4)
			}
			// sosea
			mb.Emis = glow * 0.7
			mb.Quad(P(s, -4.5, f0+0.05), P(s1, -4.5, f1+0.05), P(s1, 4.5, f1+0.05), P(s, 4.5, f0+0.05), up, ColAsphalt)
			if int(s/ds)%3 == 0 {
				mb.Quad(P(s, -0.12, f0+0.06), P(s1, -0.12, f1+0.06), P(s1, 0.12, f1+0.06), P(s, 0.12, f0+0.06), up, ColStripe)
			}
			for _, sd := range []float32{-1, 1} {
				mb.Emis = glow * 0.7
				mb.Quad(P(s, sd*4.1, f0+0.065), P(s1, sd*4.1, f1+0.065), P(s1, sd*4.25, f1+0.065), P(s, sd*4.25, f0+0.065), up, ColYellowLn)
				mb.Quad(P(s, sd*4.5, f0+0.25), P(s1, sd*4.5, f1+0.25), P(s1, sd*hw, f1+0.25), P(s, sd*hw, f0+0.25), up, ColSidewalk)
				mb.Quad(P(s, sd*4.5, f0+0.05), P(s1, sd*4.5, f1+0.05), P(s1, sd*4.5, f1+0.25), P(s, sd*4.5, f0+0.25), rl.Vector3Scale(per3, -sd), ColCurb)
				in := rl.Vector3Scale(per3, -sd)
				if open {
					// zid de sprijin pana deasupra solului + parapet + fasie pavata (acopera marginea terenului)
					x0, z0 := c.Point(s, sd*(hw+0.6))
					x1, z1 := c.Point(s1, sd*(hw+0.6))
					t0 := w.NaturalLand(x0, z0)
					t1 := w.NaturalLand(x1, z1)
					top0, top1 := maxf(t0+1.0, f0+1.2), maxf(t1+1.0, f1+1.2)
					mb.Emis = 0
					mb.Quad(P(s, sd*hw, f0+0.25), P(s1, sd*hw, f1+0.25), P(s1, sd*hw, top1), P(s, sd*hw, top0), in, shade(ColConcrete, 0.92))
					mb.Quad(P(s, sd*(hw+0.35), t0), P(s1, sd*(hw+0.35), t1), P(s1, sd*(hw+0.35), top1), P(s, sd*(hw+0.35), top0), rl.Vector3Scale(per3, sd), ColConcrete2)
					mb.Quad(P(s, sd*hw, top0), P(s1, sd*hw, top1), P(s1, sd*(hw+0.35), top1), P(s, sd*(hw+0.35), top0), up, ColConcrete2)
					mb.Quad(P(s, sd*(hw+0.35), t0+0.04), P(s1, sd*(hw+0.35), t1+0.04), P(s1, sd*(hw+3.4), t1+0.04), P(s, sd*(hw+3.4), t0+0.04), up, ColPaving)
					// dunga colorata pe zid
					mb.Emis = 0.6
					mb.Quad(P(s, sd*(hw-0.02), f0+1.3), P(s1, sd*(hw-0.02), f1+1.3), P(s1, sd*(hw-0.02), f1+1.6), P(s, sd*(hw-0.02), f0+1.6), in, c.Color)
					mb.Emis = 0
					continue
				}
				if c.wallOpen(sm, sd) {
					continue
				}
				wallC := ColTile
				mb.Emis = glow
				mb.Quad(P(s, sd*hw, f0+0.25), P(s1, sd*hw, f1+0.25), P(s1, sd*hw, f0+1.0+(f1-f0)), P(s, sd*hw, f0+1.0), in, wallC)
				mb.Emis = glow + 0.35
				mb.Quad(P(s, sd*hw, f0+1.0), P(s1, sd*hw, f1+1.0), P(s1, sd*hw, f1+1.45), P(s, sd*hw, f0+1.45), in, stripe)
				mb.Emis = glow
				upper := ColTile
				if sea {
					upper = rl.NewColor(244, 240, 230, 255)
				}
				mb.Quad(P(s, sd*hw, f0+1.45), P(s1, sd*hw, f1+1.45), P(s1, sd*hw, f1+c.WallH), P(s, sd*hw, f0+c.WallH), in, upper)
			}
			if open {
				mb.Emis = 0
				continue
			}
			// bolta eliptica
			for k := 0; k < vaultSeg; k++ {
				a0 := float32(k) / float32(vaultSeg) * math.Pi
				a1 := float32(k+1) / float32(vaultSeg) * math.Pi
				la0, la1 := hw*cosf(a0), hw*cosf(a1)
				h0, h1 := c.VaultH*sinf(a0), c.VaultH*sinf(a1)
				am := (a0 + a1) / 2
				inward := rl.Vector3Normalize(v3(-per3.X*cosf(am)*c.VaultH, -sinf(am)*hw, -per3.Z*cosf(am)*c.VaultH))
				col := ColVault
				if k%2 == 0 {
					col = shade(col, 0.95)
				}
				mb.Emis = glow * 0.8
				mb.Quad(P(s, la0, f0+c.WallH+h0), P(s1, la0, f1+c.WallH+h0), P(s1, la1, f1+c.WallH+h1), P(s, la1, f0+c.WallH+h1), inward, col)
			}
			mb.Emis = 0
		}
		// dotari in interior: lampi, ventilatoare, indicatoare, pesti
		for s := c0 + 1; s < minf(c0+chunk, c.Len); s += 4 {
			if c.IsOpen(s) {
				continue
			}
			f := c.FloorC(s)
			mb.AO = 0.5
			for _, sd := range []float32{-1, 1} {
				a := float32(math.Pi/2) - sd*0.55
				lat := (hw - 0.3) * cosf(a)
				y := f + c.WallH + (c.VaultH-0.15)*sinf(a)
				mb.Emis = 1
				mb.SetMatrix(matTRS(P(s, lat, y), yaw, 1))
				mb.Box(v3(0, 0, 0), v3(0.35, 0.16, 2.2), rl.NewColor(255, 246, 214, 255))
				mb.ClearMatrix()
			}
			mb.Emis = 0
			si := int(s)
			if si%24 == 1 {
				// indicator verde de iesire de urgenta pe ambii pereti
				for _, sd := range []float32{-1, 1} {
					if c.wallOpen(s, sd) {
						continue
					}
					mb.Emis = 1
					mb.SetMatrix(matTRS(P(s, sd*(hw-0.06), f+2.25), yaw, 1))
					mb.Box(v3(0, 0, 0), v3(0.1, 0.36, 0.8), rl.NewColor(50, 220, 110, 255))
					mb.Box(v3(-sd*0.03, 0, 0.12), v3(0.06, 0.14, 0.3), LegoWhite)
					mb.ClearMatrix()
					mb.Emis = 0
				}
			}
			if si%60 == 29 {
				// pereche de ventilatoare sub bolta
				for _, sd := range []float32{-1, 1} {
					cc := P(s, sd*2.2, f+c.WallH+c.VaultH-1.15)
					mb.AO = 0.4
					mb.Cylinder(rl.Vector3Add(cc, rl.Vector3Scale(dir3, -1.5)), rl.Vector3Add(cc, rl.Vector3Scale(dir3, 1.5)), 0.6, 0.6, 14, LegoDarkGrey, true)
					mb.Cylinder(rl.Vector3Add(cc, v3(0, 0.55, 0)), rl.Vector3Add(cc, v3(0, 1.2, 0)), 0.08, 0.08, 4, LegoDarkGrey, false)
				}
			}
			xm, _ := c.Point(s, 0)
			if underSea(xm) && si%12 == 5 {
				// pesti pictati pe pereti (Aventurierul e sub mare!)
				for _, sd := range []float32{-1, 1} {
					cb.fishMural(mb, c, s+float32((si/9)%3), sd, f+1.7+0.5*hash2(int32(si), int32(sd*7)), si)
				}
			}
		}
		// panou verde mare deasupra iesirilor (scari / lift)
		for _, o := range c.Openings {
			sm := (o.S0 + o.S1) / 2
			if sm < c0 || sm >= c0+chunk {
				continue
			}
			f := c.FloorC(sm)
			mb.AO = 0.5
			mb.Emis = 1
			// rama usii
			for _, e := range []float32{o.S0, o.S1} {
				mb.Box(P(e, o.Side*(hw-0.1), f+1.6), v3(0.3+absf(c.Dir.X)*0.0, 3.2, 0.3), rl.NewColor(60, 210, 110, 255))
			}
			mb.SetMatrix(matTRS(P(sm, o.Side*(hw-0.2), f+c.WallH+0.1), yaw, 1))
			mb.Box(v3(0, 0, 0), v3(0.12, 0.7, o.S1-o.S0+0.6), rl.NewColor(50, 220, 110, 255))
			// sageata alba spre usa
			mb.Box(v3(-o.Side*0.07, 0, 0), v3(0.04, 0.12, 1.2), LegoWhite)
			mb.Box(v3(-o.Side*0.07, 0.12, 0.45), v3(0.04, 0.12, 0.3), LegoWhite)
			mb.Box(v3(-o.Side*0.07, -0.12, 0.45), v3(0.04, 0.12, 0.3), LegoWhite)
			mb.ClearMatrix()
			// panou suspendat in mijlocul tunelului, vizibil de departe
			for _, off := range []float32{-30, 30} {
				ss := sm + off
				if ss < 2 || ss > c.Len-2 || c.IsOpen(ss) {
					continue
				}
				fy := c.FloorC(ss)
				mb.SetMatrix(matTRS(P(ss, o.Side*2.6, fy+c.WallH+c.VaultH-1.0), yaw, 1))
				mb.Box(v3(0, 0, 0), v3(2.8, 0.8, 0.12), rl.NewColor(50, 220, 110, 255))
				mb.Box(v3(o.Side*0.6, 0, -off/30*0.08), v3(1.0, 0.14, 0.04), LegoWhite)
				mb.ClearMatrix()
			}
			mb.Emis = 0
		}
		cb.addChunk(mb, true)
	}
	// portaluri unde santul devine tunel
	for s := float32(1); s < c.Len-1; s++ {
		a, b := c.IsOpen(s-1), c.IsOpen(s)
		if a && !b {
			cb.corridorPortal(c, float32(int(s)), -1)
		} else if !a && b {
			cb.corridorPortal(c, float32(int(s)), 1)
		}
	}
}

func (cb *CityBuilder) fishMural(mb *MeshBuilder, c *Corridor, s, side, y float32, seed int) {
	hw := c.HalfW - 0.03
	P := func(ss, yy float32) rl.Vector3 {
		x, z := c.Point(ss, side*hw)
		return v3(x, yy, z)
	}
	n := rl.Vector3Scale(v3(c.Perp.X, 0, c.Perp.Y), -side)
	cols := []rl.Color{LegoOrange, LegoYellow, LegoPink, LegoLime, rl.NewColor(255, 120, 90, 255)}
	col := cols[seed%len(cols)]
	dirS := float32(1)
	if seed%2 == 0 {
		dirS = -1
	}
	sz := 0.95 + 0.35*float32(seed%3)
	mb.Emis = 0.75
	seg := 10
	for k := 0; k < seg; k++ {
		a0 := float32(k) / float32(seg) * 2 * math.Pi
		a1 := float32(k+1) / float32(seg) * 2 * math.Pi
		mb.Tri(P(s, y), P(s+cosf(a0)*sz, y+sinf(a0)*sz*0.55), P(s+cosf(a1)*sz, y+sinf(a1)*sz*0.55), n, col)
	}
	tx := s - dirS*sz*0.9
	mb.Tri(P(tx, y), P(tx-dirS*sz*0.6, y+sz*0.45), P(tx-dirS*sz*0.6, y-sz*0.45), n, shade(col, 0.85))
	// ochi
	ex := s + dirS*sz*0.5
	for k := 0; k < 6; k++ {
		a0 := float32(k) / 6 * 2 * math.Pi
		a1 := float32(k+1) / 6 * 2 * math.Pi
		e := func(a float32) rl.Vector3 {
			x, z := c.Point(ex+cosf(a)*0.09, side*(hw-0.01))
			return v3(x, y+0.1+sinf(a)*0.09, z)
		}
		mb.Tri(e(a0), e(a1), func() rl.Vector3 { x, z := c.Point(ex, side*(hw-0.01)); return v3(x, y+0.1, z) }(), n, rl.NewColor(30, 30, 40, 255))
	}
	// bule
	mb.Emis = 0.4
	for k := 0; k < 3; k++ {
		bx := s + dirS*(sz+0.3+float32(k)*0.15)
		by := y + 0.3 + float32(k)*0.35
		x, z := c.Point(bx, side*(hw-0.02))
		cc := v3(x, by, z)
		for q := 0; q < 6; q++ {
			a0 := float32(q) / 6 * 2 * math.Pi
			a1 := float32(q+1) / 6 * 2 * math.Pi
			r := float32(0.07 + 0.03*float32(k))
			p0x, p0z := c.Point(bx+cosf(a0)*r, side*(hw-0.02))
			p1x, p1z := c.Point(bx+cosf(a1)*r, side*(hw-0.02))
			mb.Tri(cc, v3(p0x, by+sinf(a0)*r, p0z), v3(p1x, by+sinf(a1)*r, p1z), n, rl.NewColor(210, 240, 255, 255))
		}
	}
	mb.Emis = 0
}

// portalul de beton (cu arc colorat) la trecerea din sant in tunel; dir = spre exterior
func (cb *CityBuilder) corridorPortal(c *Corridor, s, dir float32) {
	w := cb.W
	mb := NewMB()
	hw := c.HalfW
	f := c.FloorC(s)
	P := func(lat, y float32) rl.Vector3 {
		x, z := c.Point(s, lat)
		return v3(x, y, z)
	}
	outN := v3(c.Dir.X*dir, 0, c.Dir.Y*dir)
	topAt := func(lat float32) float32 {
		x, z := c.Point(s, lat)
		return maxf(w.NaturalLand(x, z)+1.3, f+c.WallH+c.VaultH+1.2)
	}
	N := 48
	// conturul golului: perete drept, bolta, perete drept
	hole := make([][2]float32, 0, N+1)
	total := 2*c.WallH + float32(math.Pi)*(hw+c.VaultH)/2
	for k := 0; k <= N; k++ {
		t := total * float32(k) / float32(N)
		switch {
		case t <= c.WallH:
			hole = append(hole, [2]float32{hw, f + t})
		case t <= total-c.WallH:
			a := (t - c.WallH) / (total - 2*c.WallH) * math.Pi
			hole = append(hole, [2]float32{hw * cosf(a), f + c.WallH + c.VaultH*sinf(a)})
		default:
			hole = append(hole, [2]float32{-hw, f + c.WallH - (t - (total - c.WallH))})
		}
	}
	ow := hw + 1.2
	outer := make([][2]float32, 0, N+1)
	for k := 0; k <= N; k++ {
		// contur exterior parcurs in acelasi sens: dreapta-jos, sus, stanga-jos
		t := float32(k) / float32(N)
		var lat, y float32
		switch {
		case t < 0.2:
			lat, y = ow, f+(topAt(ow)-f)*(t/0.2)
		case t < 0.8:
			lat = ow - 2*ow*(t-0.2)/0.6
			y = topAt(lat)
		default:
			lat, y = -ow, topAt(-ow)-(topAt(-ow)-f)*((t-0.8)/0.2)
		}
		outer = append(outer, [2]float32{lat, y})
	}
	stone := func(lat, y float32) rl.Color {
		iy := int32(math.Floor(float64(y / 0.8)))
		ix := int32(math.Floor(float64((lat + float32(iy&1)*0.6) / 1.2)))
		return shade(ColConcrete, 0.9+0.14*hash2(ix, iy+int32(s)))
	}
	for k := 0; k < N; k++ {
		a0, a1 := hole[k], hole[k+1]
		b0, b1 := outer[k], outer[k+1]
		mb.Quad(P(a0[0], a0[1]), P(a1[0], a1[1]), P(b1[0], b1[1]), P(b0[0], b0[1]), outN, stone((a0[0]+b0[0])/2, (a0[1]+b0[1])/2))
	}
	// arcul colorat luminos in jurul golului
	mb.Emis = 1
	for k := 4; k < N-4; k++ {
		a0, a1 := hole[k], hole[k+1]
		e := func(p [2]float32, d float32) rl.Vector3 {
			x, z := c.Point(s+dir*0.05, p[0]*(1+d/hw))
			return v3(x, p[1]+d*0.6, z)
		}
		mb.Quad(e(a0, 0.15), e(a1, 0.15), e(a1, 0.55), e(a0, 0.55), outN, c.Color)
	}
	mb.Emis = 0
	// streasina de sus
	tl := topAt(0)
	x0, z0 := c.Point(s+dir*0.6, -ow-0.3)
	x1, z1 := c.Point(s+dir*0.6, ow+0.3)
	x2, z2 := c.Point(s-dir*0.6, ow+0.3)
	x3, z3 := c.Point(s-dir*0.6, -ow-0.3)
	mb.Quad(v3(x0, tl+0.3, z0), v3(x1, tl+0.3, z1), v3(x2, tl+0.3, z2), v3(x3, tl+0.3, z3), v3(0, 1, 0), ColConcrete2)
	mb.Quad(v3(x0, tl-0.1, z0), v3(x1, tl-0.1, z1), v3(x1, tl+0.3, z1), v3(x0, tl+0.3, z0), outN, ColConcrete2)
	cb.addChunk(mb, true)
}

func (cb *CityBuilder) stairMesh(c *Corridor) {
	w := cb.W
	mb := NewMB()
	hw := c.HalfW
	P := func(s, lat, y float32) rl.Vector3 {
		x, z := c.Point(s, lat)
		return v3(x, y, z)
	}
	per3 := v3(c.Perp.X, 0, c.Perp.Y)
	dir3 := v3(c.Dir.X, 0, c.Dir.Y)
	up := v3(0, 1, 0)
	yaw := atan2f(c.Dir.X, c.Dir.Y)
	const run = float32(0.5)
	for s := c.WallFrom; s < c.Len-0.001; s += run {
		s1 := minf(s+run, c.Len)
		sm := (s + s1) / 2
		open := c.IsOpen(sm)
		if open {
			mb.AO = 1
		} else {
			mb.AO = 0.25
		}
		glow := float32(0.3)
		if open {
			glow = 0
		}
		// treapta: calcai orizontal + contratreapta
		yT := c.FloorC(s1)
		yP := c.FloorC(s)
		mb.Emis = glow * 0.6
		mb.Quad(P(s, -hw, yT), P(s1, -hw, yT), P(s1, hw, yT), P(s, hw, yT), up, ColPaving)
		if yT-yP > 0.01 {
			mb.Quad(P(s, -hw, yP), P(s, hw, yP), P(s, hw, yT), P(s, -hw, yT), rl.Vector3Scale(dir3, -1), ColPaving2)
			mb.Emis = 0.9
			mb.Quad(P(s-0.01, -hw, yT-0.06), P(s-0.01, hw, yT-0.06), P(s-0.01, hw, yT-0.01), P(s-0.01, -hw, yT-0.01), rl.Vector3Scale(dir3, -1), LegoYellow)
		}
		for _, sd := range []float32{-1, 1} {
			in := rl.Vector3Scale(per3, -sd)
			if open {
				x0, z0 := c.Point(s, sd*(hw+0.6))
				x1, z1 := c.Point(s1, sd*(hw+0.6))
				t0, t1 := w.NaturalLand(x0, z0), w.NaturalLand(x1, z1)
				top0, top1 := maxf(t0+1.1, yP+1.1), maxf(t1+1.1, yT+1.1)
				mb.Emis = 0
				mb.Quad(P(s, sd*hw, yP-0.6), P(s1, sd*hw, yT-0.6), P(s1, sd*hw, top1), P(s, sd*hw, top0), in, shade(ColConcrete, 0.95))
				mb.Quad(P(s, sd*(hw+0.3), t0), P(s1, sd*(hw+0.3), t1), P(s1, sd*(hw+0.3), top1), P(s, sd*(hw+0.3), top0), rl.Vector3Scale(per3, sd), ColConcrete2)
				mb.Quad(P(s, sd*hw, top0), P(s1, sd*hw, top1), P(s1, sd*(hw+0.3), top1), P(s, sd*(hw+0.3), top0), up, LegoGreen)
				mb.Quad(P(s, sd*(hw+0.3), t0+0.045), P(s1, sd*(hw+0.3), t1+0.045), P(s1, sd*(hw+3.4), t1+0.045), P(s, sd*(hw+3.4), t0+0.045), up, ColPaving)
				continue
			}
			mb.Emis = glow
			mb.Quad(P(s, sd*hw, yP-0.6), P(s1, sd*hw, yT-0.6), P(s1, sd*hw, yT+c.WallH), P(s, sd*hw, yP+c.WallH), in, ColTile)
			mb.Emis = glow + 0.4
			mb.Quad(P(s, sd*(hw-0.01), yP+0.9), P(s1, sd*(hw-0.01), yT+0.9), P(s1, sd*(hw-0.01), yT+1.1), P(s, sd*(hw-0.01), yP+1.1), in, c.Color)
		}
		if !open {
			for k := 0; k < 8; k++ {
				a0 := float32(k) / 8 * math.Pi
				a1 := float32(k+1) / 8 * math.Pi
				la0, la1 := hw*cosf(a0), hw*cosf(a1)
				h0, h1 := c.VaultH*sinf(a0), c.VaultH*sinf(a1)
				am := (a0 + a1) / 2
				inward := rl.Vector3Normalize(v3(-per3.X*cosf(am), -sinf(am)*2, -per3.Z*cosf(am)))
				mb.Emis = glow * 0.7
				mb.Quad(P(s, la0, yP+c.WallH+h0), P(s1, la0, yT+c.WallH+h0), P(s1, la1, yT+c.WallH+h1), P(s, la1, yP+c.WallH+h1), inward, ColVault)
			}
			if int(s/run)%6 == 0 {
				mb.Emis = 1
				mb.SetMatrix(matTRS(P(sm, 0, yT+c.WallH+c.VaultH-0.08), yaw, 1))
				mb.Box(v3(0, 0, 0), v3(1.4, 0.1, 0.4), rl.NewColor(255, 246, 214, 255))
				mb.ClearMatrix()
			}
		}
		mb.Emis = 0
	}
	// gura scarilor: prima treapta deschisa
	var sOpen float32 = -1
	for s := c.WallFrom; s < c.Len; s += 0.5 {
		if c.IsOpen(s) {
			sOpen = s
			break
		}
	}
	if sOpen > 0 {
		// intrare deschisa (fara acoperis, ca scarile sa se vada de departe):
		// poarta verde cu panou luminos si o sageata in jos, plus un stalp inalt cu lumina
		mb.AO = 1
		top := float32(4.4)
		for _, sd := range []float32{-1, 1} {
			p := P(c.Len+0.2, sd*(hw+0.45), 0)
			mb.Cylinder(p, v3(p.X, top, p.Z), 0.16, 0.16, 10, LegoGreen, true)
		}
		mb.SetMatrix(matTRS(P(c.Len+0.2, 0, top), yaw, 1))
		mb.BeveledBox(v3(0, 0, 0), v3(2*hw+1.3, 0.35, 0.35), 0.06, LegoGreen)
		mb.Emis = 1
		mb.Box(v3(0, 0.75, 0), v3(3.6, 1.15, 0.2), rl.NewColor(50, 220, 110, 255))
		for _, fz := range []float32{0.11, -0.11} {
			mb.Box(v3(0, 0.88, fz), v3(0.22, 0.6, 0.04), LegoWhite)
			mb.Box(v3(-0.2, 0.52, fz), v3(0.22, 0.22, 0.04), LegoWhite)
			mb.Box(v3(0.2, 0.52, fz), v3(0.22, 0.22, 0.04), LegoWhite)
			mb.Box(v3(0, 0.38, fz), v3(0.22, 0.16, 0.04), LegoWhite)
			mb.Box(v3(-1.3, 0.75, fz), v3(0.5, 0.5, 0.04), LegoYellow)
			mb.Box(v3(1.3, 0.75, fz), v3(0.5, 0.5, 0.04), LegoYellow)
		}
		mb.ClearMatrix()
		mb.Emis = 0
		// stalp inalt cu lampa (se vede din toata piata)
		tp := P(c.Len+1.2, hw+1.4, 0)
		mb.Cylinder(tp, v3(tp.X, 8.2, tp.Z), 0.14, 0.12, 8, LegoDarkGrey, false)
		mb.Emis = 1
		mb.BeveledBox(v3(tp.X, 8.8, tp.Z), v3(1.2, 1.2, 1.2), 0.15, rl.NewColor(50, 220, 110, 255))
		mb.Box(v3(tp.X, 8.8, tp.Z), v3(0.3, 0.8, 1.26), LegoWhite)
		mb.Box(v3(tp.X, 8.8, tp.Z), v3(1.26, 0.8, 0.3), LegoWhite)
		// becuri galbene pe balustrade
		for s := sOpen; s < c.Len; s += 1.5 {
			for _, sd := range []float32{-1, 1} {
				x, z := c.Point(s, sd*(hw+0.15))
				gy := w.NaturalLand(x, z)
				mb.Box(v3(x, maxf(gy+1.1, 1.1)+0.1, z), v3(0.22, 0.14, 0.22), LegoYellow)
			}
		}
		mb.Emis = 0
		// balustrade solide in jurul gurii (si pentru inamici)
		for _, sd := range []float32{-1, 1} {
			ax, az := c.Point(sOpen-0.5, sd*(hw+0.15))
			bx, bz := c.Point(c.Len, sd*(hw+0.15))
			w.Boxes = append(w.Boxes, BoxCollider{MinX: minf(ax, bx) - 0.2, MaxX: maxf(ax, bx) + 0.2, MinZ: minf(az, bz) - 0.2, MaxZ: maxf(az, bz) + 0.2, Top: 1.1})
		}
		ax, az := c.Point(sOpen-0.6, -hw-0.3)
		bx, bz := c.Point(sOpen-0.6, hw+0.3)
		w.Boxes = append(w.Boxes, BoxCollider{MinX: minf(ax, bx) - 0.2, MaxX: maxf(ax, bx) + 0.2, MinZ: minf(az, bz) - 0.2, MaxZ: maxf(az, bz) + 0.2, Top: 1.1})
		// parapetul din capatul de jos al gurii
		ex0, ez0 := c.Point(sOpen-0.6, 0)
		_ = ex0
		_ = ez0
		mb.Quad(P(sOpen-0.6, -hw, -0.2), P(sOpen-0.6, hw, -0.2), P(sOpen-0.6, hw, 1.1), P(sOpen-0.6, -hw, 1.1), rl.Vector3Scale(dir3, -1), ColConcrete2)
		mb.Quad(P(sOpen-0.6, -hw, 1.1), P(sOpen-0.6, hw, 1.1), P(sOpen-0.3, hw, 1.1), P(sOpen-0.3, -hw, 1.1), up, LegoGreen)
	}
	cb.addChunk(mb, true)
}

func (cb *CityBuilder) liftMeshes(l *Lift) {
	mb := NewMB()
	// usa in tunel
	bx, bz := l.Bottom.X, float32(6.45)
	f := l.Bottom.Y
	mb.AO = 0.5
	mb.BeveledBox(v3(bx, f+1.4, bz), v3(2.4, 2.8, 0.1), 0.04, rl.NewColor(190, 196, 206, 255))
	mb.Box(v3(bx, f+1.4, bz-0.06), v3(0.04, 2.7, 0.02), LegoDarkGrey)
	mb.Emis = 1
	mb.Box(v3(bx+1.6, f+1.3, bz-0.08), v3(0.3, 0.5, 0.1), LegoGold)
	mb.Box(v3(bx+1.6, f+1.38, bz-0.14), v3(0.1, 0.1, 0.04), LegoCyan)
	mb.Emis = 0
	// insulita: turnul de ventilatie si chioscul liftului
	tx, tz := l.Top.X, l.Top.Z
	mb.AO = 1
	mb.Cylinder(v3(tx, 0, tz-9), v3(tx, 16, tz-9), 3.2, 2.6, 24, rl.NewColor(232, 232, 226, 255), true)
	for k := 0; k < 4; k++ {
		y := 3 + float32(k)*3.4
		mb.Cylinder(v3(tx, y, tz-9), v3(tx, y+0.6, tz-9), 3.0, 2.8, 24, []rl.Color{LegoRed, LegoBlue, LegoRed, LegoBlue}[k], false)
	}
	mb.Cylinder(v3(tx, 16, tz-9), v3(tx, 17.2, tz-9), 2.2, 2.2, 20, LegoDarkGrey, true)
	mb.Emis = 1
	mb.Sphere(v3(tx, 17.8, tz-9), 0.45, LegoRed)
	mb.Emis = 0
	mb.BeveledBox(v3(tx, 1.7, tz-1.4), v3(3.4, 3.4, 3.0), 0.2, rl.NewColor(150, 210, 240, 255))
	mb.BeveledBox(v3(tx, 3.55, tz-1.4), v3(3.9, 0.35, 3.5), 0.1, LegoGreen)
	mb.Box(v3(tx, 1.4, tz+0.12), v3(1.8, 2.6, 0.05), rl.NewColor(190, 196, 206, 255))
	mb.Emis = 1
	mb.Box(v3(tx, 4.1, tz+0.1), v3(2.2, 0.7, 0.12), rl.NewColor(50, 220, 110, 255))
	mb.Box(v3(tx, 4.1, tz+0.18), v3(1.2, 0.14, 0.04), LegoWhite)
	mb.Emis = 0
	cb.W.Circles = append(cb.W.Circles, CircleCollider{X: tx, Z: tz - 9, R: 3.3, Top: 17})
	cb.W.Boxes = append(cb.W.Boxes, BoxCollider{MinX: tx - 1.8, MaxX: tx + 1.8, MinZ: tz - 3, MaxZ: tz + 0.15, Top: 3.7, Roof: RoofFlat})
	// palmieri si stanci
	cb.palmTree(mb, tx-7, tz+3, 0.9)
	cb.palmTree(mb, tx+7, tz-2, 1.0)
	cb.addChunk(mb, true)
}

// stairMouth: punctul e deasupra gurii deschise a unor scari (acolo pavajul pietei are gaura).
func (w *World) stairMouth(x, z, margin float32) bool {
	for i := range w.Corridors {
		c := &w.Corridors[i]
		if c.Kind != CorrStair {
			continue
		}
		s, lat := c.Local(x, z)
		if s > c.Len+0.05 || absf(lat) > c.HalfW+margin {
			continue
		}
		if s >= c.WallFrom && (c.IsOpen(s) || c.IsOpen(s+0.6)) {
			return true
		}
	}
	return false
}
