package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ---------------- MODELE ----------------

func buildCarMesh(body rl.Color, race bool) rl.Mesh {
	mb := NewMB()
	glass := rl.NewColor(60, 90, 130, 255)
	if race {
		mb.BeveledBox(v3(0, 0.55, 0), v3(2.1, 0.55, 4.4), 0.2, body)
		mb.BeveledBox(v3(0, 0.95, -0.4), v3(1.1, 0.45, 1.6), 0.2, glass)
		mb.BeveledBox(v3(0, 1.25, -2.05), v3(2.2, 0.12, 0.6), 0.04, LegoBlack)
		for _, s := range []float32{-0.8, 0.8} {
			mb.Box(v3(s, 0.95, -2.05), v3(0.1, 0.6, 0.4), LegoBlack)
		}
		mb.Box(v3(0, 0.84, 1.2), v3(0.5, 0.02, 1.6), LegoWhite)
	} else {
		mb.BeveledBox(v3(0, 0.68, 0), v3(2.0, 0.72, 4.2), 0.25, body)
		mb.BeveledBox(v3(0, 1.32, -0.25), v3(1.72, 0.66, 2.3), 0.22, body)
		// geamuri
		mb.Quad(v3(-0.72, 1.08, 0.92), v3(0.72, 1.08, 0.92), v3(0.62, 1.55, 0.78), v3(-0.62, 1.55, 0.78), v3(0, 0.4, 1), glass)
		mb.Quad(v3(-0.72, 1.08, -1.42), v3(0.72, 1.08, -1.42), v3(0.62, 1.55, -1.3), v3(-0.62, 1.55, -1.3), v3(0, 0.4, -1), glass)
		for _, s := range []float32{-1, 1} {
			mb.Quad(v3(s*0.875, 1.1, 0.75), v3(s*0.875, 1.1, -1.25), v3(s*0.875, 1.52, -1.15), v3(s*0.875, 1.52, 0.65), v3(s, 0, 0), glass)
		}
		mb.Box(v3(0, 0.45, 2.12), v3(1.9, 0.22, 0.12), LegoLightGrey)
		mb.Box(v3(0, 0.45, -2.12), v3(1.9, 0.22, 0.12), LegoLightGrey)
	}
	mb.Emis = 1
	for _, s := range []float32{-0.65, 0.65} {
		mb.Box(v3(s, 0.72, 2.1), v3(0.42, 0.2, 0.06), rl.NewColor(255, 248, 210, 255))
		mb.Box(v3(s, 0.72, -2.1), v3(0.42, 0.18, 0.06), rl.NewColor(240, 40, 40, 255))
	}
	mb.Emis = 0
	for _, w := range [][2]float32{{-0.98, 1.35}, {0.98, 1.35}, {-0.98, -1.35}, {0.98, -1.35}} {
		mb.Cylinder(v3(w[0]-0.17, 0.42, w[1]), v3(w[0]+0.17, 0.42, w[1]), 0.42, 0.42, 14, rl.NewColor(40, 40, 46, 255), true)
		s := float32(1)
		if w[0] < 0 {
			s = -1
		}
		mb.Cylinder(v3(w[0]+s*0.17, 0.42, w[1]), v3(w[0]+s*0.19, 0.42, w[1]), 0.24, 0.24, 10, LegoLightGrey, true)
	}
	return mb.Build()
}

func buildBoatMesh(col rl.Color) rl.Mesh {
	mb := NewMB()
	white := rl.NewColor(244, 244, 240, 255)
	// cocă ascutita in fata
	hw := float32(1.3)
	pts := []rl.Vector3{v3(-hw, 0, -2.6), v3(hw, 0, -2.6), v3(hw, 0, 1.6), v3(0, 0, 3.4), v3(-hw, 0, 1.6)}
	bot := []rl.Vector3{v3(-hw*0.6, -0.8, -2.4), v3(hw*0.6, -0.8, -2.4), v3(hw*0.6, -0.8, 1.3), v3(0, -0.6, 3.0), v3(-hw*0.6, -0.8, 1.3)}
	topY := float32(0.9)
	for i := 0; i < 5; i++ {
		j := (i + 1) % 5
		a, b := pts[i], pts[j]
		at := v3(a.X, topY, a.Z)
		bt := v3(b.X, topY, b.Z)
		mid := v3((a.X+b.X)/2, 0, (a.Z+b.Z)/2)
		mb.Quad(bot[i], bot[j], b, a, mid, col)
		mb.Quad(a, b, bt, at, mid, white)
	}
	// punte
	c := v3(0, topY, 0)
	for i := 0; i < 5; i++ {
		j := (i + 1) % 5
		mb.Tri(c, v3(pts[i].X, topY, pts[i].Z), v3(pts[j].X, topY, pts[j].Z), v3(0, 1, 0), rl.NewColor(200, 150, 100, 255))
		mb.Tri(v3(0, -0.8, 0), bot[i], bot[j], v3(0, -1, 0), col)
	}
	mb.BeveledBox(v3(0, 1.5, -0.6), v3(1.8, 1.2, 2.0), 0.15, white)
	mb.Quad(v3(-0.8, 1.35, 0.42), v3(0.8, 1.35, 0.42), v3(0.8, 1.85, 0.42), v3(-0.8, 1.85, 0.42), v3(0, 0, 1), rl.NewColor(70, 110, 160, 255))
	mb.BeveledBox(v3(0, 2.15, -0.6), v3(2.0, 0.12, 2.2), 0.04, col)
	mb.Cylinder(v3(0, 2.2, -1.2), v3(0, 3.6, -1.2), 0.05, 0.05, 5, LegoLightGrey, false)
	mb.Tri(v3(0, 3.6, -1.2), v3(0, 3.1, -1.2), v3(0, 3.35, -0.5), v3(1, 0, 0), LegoRed)
	mb.Tri(v3(0, 3.6, -1.2), v3(0, 3.1, -1.2), v3(0, 3.35, -0.5), v3(-1, 0, 0), LegoRed)
	return mb.Build()
}

func buildPlaneMesh() rl.Mesh {
	mb := NewMB()
	white := rl.NewColor(244, 246, 250, 255)
	mb.Cylinder(v3(0, 0, -7), v3(0, 0, 6), 1.35, 1.35, 16, white, false)
	mb.Ellipsoid(v3(0, 0, 6), v3(1.35, 1.35, 2.2), 8, 16, white)
	mb.CylinderC(v3(0, 0, -7), v3(0, 0.6, -11), 1.35, 0.3, 16, white, white, true)
	mb.Box(v3(0, -0.2, 0), v3(19, 0.25, 2.8), white)
	mb.Box(v3(0, 0.5, -10), v3(6.5, 0.18, 1.6), white)
	mb.BeveledBox(v3(0, 2.2, -10.2), v3(0.25, 3.2, 2.2), 0.08, LegoBlue)
	mb.Box(v3(1.36, 0, -0.5), v3(0.02, 0.35, 12), LegoBlue)
	mb.Box(v3(-1.36, 0, -0.5), v3(0.02, 0.35, 12), LegoBlue)
	for k := 0; k < 12; k++ {
		z := 4.5 - float32(k)*1.0
		for _, s := range []float32{-1, 1} {
			mb.Quad(v3(s*1.355, 0.35, z), v3(s*1.355, 0.35, z+0.45), v3(s*1.355, 0.72, z+0.45), v3(s*1.355, 0.72, z), v3(s, 0, 0), rl.NewColor(50, 70, 100, 255))
		}
	}
	for _, s := range []float32{-4.5, 4.5} {
		mb.Cylinder(v3(s, -0.9, 1.5), v3(s, -0.9, -1.2), 0.55, 0.45, 12, LegoLightGrey, true)
	}
	return mb.Build()
}

func buildCloudMesh(seed int) rl.Mesh {
	mb := NewMB()
	mb.Emis = 0.55
	n := 5 + seed%3
	for i := 0; i < n; i++ {
		fx := (hash2(int32(seed), int32(i))*2 - 1) * 14
		fz := (hash2(int32(seed+7), int32(i))*2 - 1) * 6
		r := 5 + hash2(int32(seed+3), int32(i))*5
		y := r * 0.25
		mb.EllipsoidF(v3(fx, y, fz), v3(r*1.3, r*0.8, r), 7, 12, func(n rl.Vector3) rl.Color {
			return shade(rl.NewColor(255, 255, 255, 255), 0.86+0.14*(n.Y*0.5+0.5))
		})
	}
	return mb.Build()
}

func buildFerrisWheel() (wheel rl.Mesh, cabin rl.Mesh) {
	mb := NewMB()
	R := float32(11)
	n := 40
	for _, z := range []float32{-1.2, 1.2} {
		for i := 0; i < n; i++ {
			a0 := float32(i) / float32(n) * 2 * math.Pi
			a1 := float32(i+1) / float32(n) * 2 * math.Pi
			mb.Cylinder(v3(cosf(a0)*R, sinf(a0)*R, z), v3(cosf(a1)*R, sinf(a1)*R, z), 0.22, 0.22, 6, LegoWhite, false)
		}
		for i := 0; i < 12; i++ {
			a := float32(i) / 12 * 2 * math.Pi
			mb.Cylinder(v3(0, 0, z), v3(cosf(a)*R, sinf(a)*R, z), 0.09, 0.09, 5, LegoWhite, false)
		}
	}
	mb.Emis = 1
	for i := 0; i < 24; i++ {
		a := float32(i) / 24 * 2 * math.Pi
		cols := []rl.Color{LegoRed, LegoYellow, LegoCyan, LegoPink}
		mb.Sphere(v3(cosf(a)*(R+0.25), sinf(a)*(R+0.25), 1.35), 0.18, cols[i%4])
	}
	mb.Emis = 0
	wheel = mb.Build()
	cb := NewMB()
	cb.Cylinder(v3(0, 0, -0.1), v3(0, -0.8, -0.1), 0.06, 0.06, 4, LegoLightGrey, false)
	cb.BeveledBox(v3(0, -1.6, 0), v3(1.6, 1.4, 1.4), 0.3, rl.NewColor(255, 255, 255, 255))
	cb.BeveledBox(v3(0, -0.9, 0), v3(1.8, 0.12, 1.6), 0.05, rl.NewColor(255, 255, 255, 255))
	cabin = cb.Build()
	return
}

func buildTurbine() (tower rl.Mesh, rotor rl.Mesh) {
	mb := NewMB()
	mb.Cylinder(v3(0, 0, 0), v3(0, 26, 0), 1.0, 0.55, 14, LegoWhite, false)
	mb.BeveledBox(v3(0, 26.6, -0.4), v3(1.4, 1.3, 3.2), 0.4, LegoWhite)
	tower = mb.Build()
	rb := NewMB()
	rb.Ellipsoid(v3(0, 0, 0), v3(0.6, 0.6, 0.9), 6, 10, LegoRed)
	for k := 0; k < 3; k++ {
		a := float32(k) / 3 * 2 * math.Pi
		dir := v3(cosf(a), sinf(a), 0)
		rb.SetMatrix(rl.MatrixRotateZ(a))
		rb.BeveledBox(v3(5.5, 0, 0), v3(10, 0.75, 0.18), 0.06, LegoWhite)
		rb.Box(v3(9.8, 0, 0), v3(1.2, 0.6, 0.2), LegoRed)
		rb.ClearMatrix()
		_ = dir
	}
	rotor = rb.Build()
	return
}

// ---------------- RUTE DE TRAFIC ----------------

type Route struct {
	Pts  []rl.Vector3 // y precalculat (poduri, tunele)
	Cum  []float32
	Len  float32
	Name string
}

type routeBuilder struct {
	pts []rl.Vector2
}

func (rb *routeBuilder) add(p rl.Vector2) {
	if n := len(rb.pts); n > 0 {
		if rl.Vector2Distance(rb.pts[n-1], p) < 0.5 {
			return
		}
	}
	rb.pts = append(rb.pts, p)
}

// linie pe banda din dreapta (circulatie pe dreapta)
func (rb *routeBuilder) line(a, b rl.Vector2, lane float32) {
	d := rl.Vector2Normalize(rl.Vector2Subtract(b, a))
	right := rl.NewVector2(-d.Y, d.X)
	off := rl.Vector2Scale(right, lane)
	L := rl.Vector2Distance(a, b)
	n := int(L/4) + 1
	for i := 0; i <= n; i++ {
		t := float32(i) / float32(n)
		rb.add(rl.Vector2Add(rl.Vector2Lerp(a, b, t), off))
	}
}

func screenAngle(d rl.Vector2) float32 { return atan2f(-d.Y, d.X) }

// arc in sens trigonometric vazut de sus (sensul giratiilor la circulatia pe dreapta)
func (rb *routeBuilder) arc(c rl.Vector2, r, th0, th1 float32) {
	for th1 <= th0 {
		th1 += 2 * math.Pi
	}
	n := int((th1-th0)*r/3) + 2
	for i := 0; i <= n; i++ {
		th := th0 + (th1-th0)*float32(i)/float32(n)
		rb.add(rl.NewVector2(c.X+r*cosf(th), c.Y-r*sinf(th)))
	}
}

// trecere prin giratie: intrare din bratul dIn, iesire pe bratul dOut
func (rb *routeBuilder) roundabout(c, dIn, dOut rl.Vector2) {
	r := float32(8.75)
	d := float32(0.26)
	rb.arc(c, r, screenAngle(dIn)+d, screenAngle(dOut)-d)
}

func (w *World) finishRoute(rb *routeBuilder, name string) *Route {
	r := &Route{Name: name}
	y := float32(0.0)
	for _, p := range rb.pts {
		g, k := w.GroundAt(p.X, p.Y, y+0.6)
		// urmarim suprafata pas cu pas (pod / tunel)
		y = g
		yy := maxf(g, 0)
		if k == SurfTunnel {
			yy = g // in tunelul de sub mare drumul coboara sub nivelul apei
		}
		r.Pts = append(r.Pts, v3(p.X, yy+0.05, p.Y))
	}
	// a doua trecere: netezire pe inaltime
	for k := 0; k < 3; k++ {
		for i := 1; i < len(r.Pts)-1; i++ {
			r.Pts[i].Y = (r.Pts[i-1].Y + 2*r.Pts[i].Y + r.Pts[i+1].Y) / 4
		}
	}
	r.Cum = make([]float32, len(r.Pts)+1)
	for i := 0; i < len(r.Pts); i++ {
		j := (i + 1) % len(r.Pts)
		r.Cum[i+1] = r.Cum[i] + rl.Vector3Distance(r.Pts[i], r.Pts[j])
	}
	r.Len = r.Cum[len(r.Pts)]
	return r
}

// Sample: pozitie si directie la distanta s de-a lungul rutei (bucla inchisa)
func (r *Route) Sample(s float32) (rl.Vector3, rl.Vector3) {
	s = float32(math.Mod(float64(s), float64(r.Len)))
	if s < 0 {
		s += r.Len
	}
	lo, hi := 0, len(r.Pts)
	for hi-lo > 1 {
		mid := (lo + hi) / 2
		if r.Cum[mid] <= s {
			lo = mid
		} else {
			hi = mid
		}
	}
	a := r.Pts[lo]
	b := r.Pts[(lo+1)%len(r.Pts)]
	seg := r.Cum[lo+1] - r.Cum[lo]
	t := float32(0)
	if seg > 0 {
		t = (s - r.Cum[lo]) / seg
	}
	p := rl.Vector3Lerp(a, b, t)
	d := rl.Vector3Normalize(rl.Vector3Subtract(b, a))
	return p, d
}

func islandArm(si int, d rl.Vector2, r float32) rl.Vector2 {
	c := Sectors[si].Center
	return rl.NewVector2(c.X+d.X*r, c.Y+d.Y*r)
}

var (
	dirE = rl.NewVector2(1, 0)
	dirW = rl.NewVector2(-1, 0)
	dirN = rl.NewVector2(0, -1)
	dirS = rl.NewVector2(0, 1)
)

// bucla mare pe cele 4 poduri
func (w *World) BuildRingRoute(reverse bool) *Route {
	type hop struct {
		si      int
		in, out rl.Vector2
	}
	hops := []hop{
		{0, dirS, dirE}, // NV: vine dinspre podul de vest, pleaca spre podul de nord
		{1, dirW, dirS},
		{3, dirN, dirW},
		{2, dirE, dirN},
	}
	if reverse {
		rev := []hop{}
		for i := len(hops) - 1; i >= 0; i-- {
			h := hops[i]
			rev = append(rev, hop{h.si, h.out, h.in})
		}
		hops = rev
	}
	rb := &routeBuilder{}
	lane := float32(2.2)
	for i, h := range hops {
		c := Sectors[h.si].Center
		rb.roundabout(c, h.in, h.out)
		nx := hops[(i+1)%len(hops)]
		rb.line(islandArm(h.si, h.out, 12), islandArm(nx.si, nx.in, 12), lane)
	}
	name := "inel"
	if reverse {
		name = "inel-invers"
	}
	return w.finishRoute(rb, name)
}

// bucla prin toate cele 4 tuneluri si piata centrala
func (w *World) BuildTunnelRoute(reverse bool) *Route {
	inner := func(si int) rl.Vector2 { return rl.Vector2Normalize(Sectors[si].Inner) }
	out := func(si int) rl.Vector2 { return rl.Vector2Scale(inner(si), -1) } // de la centru spre insula
	type step struct {
		kind    int // 0 = giratie pe insula, 1 = inelul central
		si      int
		in, out rl.Vector2
	}
	// NV -> centru -> SE -> (pod E) -> NE -> centru -> SV -> (pod V) -> NV
	steps := []step{
		{0, 0, dirS, inner(0)},
		{1, -1, out(0), out(3)},
		{0, 3, inner(3), dirN},
		{0, 1, dirS, inner(1)},
		{1, -1, out(1), out(2)},
		{0, 2, inner(2), dirN},
	}
	if reverse {
		rev := []step{}
		for i := len(steps) - 1; i >= 0; i-- {
			s := steps[i]
			rev = append(rev, step{s.kind, s.si, s.out, s.in})
		}
		steps = rev
	}
	rb := &routeBuilder{}
	lane := float32(2.2)
	endPoint := func(s step, leaving bool) rl.Vector2 {
		d := s.out
		if !leaving {
			d = s.in
		}
		if s.kind == 1 {
			return rl.Vector2Scale(d, 52)
		}
		return islandArm(s.si, d, 12)
	}
	for i, s := range steps {
		if s.kind == 0 {
			rb.roundabout(Sectors[s.si].Center, s.in, s.out)
		} else {
			r := float32(50)
			dd := float32(2.2) / r
			rb.arc(rl.NewVector2(0, 0), r, screenAngle(s.in)+dd, screenAngle(s.out)-dd)
		}
		nx := steps[(i+1)%len(steps)]
		rb.line(endPoint(s, true), endPoint(nx, false), lane)
	}
	name := "tuneluri"
	if reverse {
		name = "tuneluri-invers"
	}
	return w.finishRoute(rb, name)
}

// ---------------- VEHICULE IN MISCARE ----------------

type Car struct {
	Route  *Route
	S      float32
	Speed  float32
	MaxSpd float32
	Mesh   int
	Pos    rl.Vector3
	Dir    rl.Vector3
	Yaw    float32
	Pitch  float32
	Honk   float32
	Whoosh float32
	HP     float32 // viata (se strica la impuscaturi)
	Flash  float32
	Dead   float32 // >0: a explodat, reapare dupa atatea secunde
	Driven bool    // o conduce Aventurierul
	Parked bool    // Aventurierul a lasat-o parcata
	Hidden bool    // V 5.0: e intr-o lume inca inchisa (sub nori)
}

type Boat struct {
	C      rl.Vector2
	RX, RZ float32
	Phase  float32
	Speed  float32
	Mesh   int
	Pos    rl.Vector3
	Yaw    float32
	Roll   float32
	HP     float32
	Dead   float32 // >0: s-a scufundat, reapare dupa atatea secunde
	Flash  float32
}

func (b *Boat) Update(dt, t float32) {
	if b.Flash > 0 {
		b.Flash -= dt
	}
	if b.Dead > 0 {
		b.Dead -= dt
		if b.Dead <= 0 {
			b.HP = BoatMaxHP
		}
	}
	b.Phase += dt * b.Speed / ((b.RX + b.RZ) / 2)
	x := b.C.X + cosf(b.Phase)*b.RX
	z := b.C.Y + sinf(b.Phase)*b.RZ
	dx := -sinf(b.Phase) * b.RX
	dz := cosf(b.Phase) * b.RZ
	b.Pos = v3(x, WaterY-0.15+sinf(t*1.7+b.Phase*3)*0.12, z)
	b.Yaw = atan2f(dx, dz)
	b.Roll = sinf(t*1.3+b.Phase) * 0.05
}

// bucla prin Marele Tunel: Insula Plajei -> pe sub mare -> Parcul de Distractii si inapoi
func (w *World) BuildGreatTunnelRoute() *Route {
	rb := &routeBuilder{}
	lane := float32(2.2)
	cW := rl.NewVector2(-401, 0)
	cE := rl.NewVector2(401, 0)
	rb.line(rl.NewVector2(-392, 0), rl.NewVector2(392, 0), lane)
	rb.arc(cE, 6.5, screenAngle(dirW)+0.4, screenAngle(dirW)-0.4)
	rb.line(rl.NewVector2(392, 0), rl.NewVector2(-392, 0), lane)
	rb.arc(cW, 6.5, screenAngle(dirE)+0.4, screenAngle(dirE)-0.4)
	return w.finishRoute(rb, "marele-tunel")
}
