package main

import (
	"fmt"
	"math"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// PARCUL DE DISTRACTII: trambuline, roti mari (cu cabine in care urci),
// montagne russe. Plus constructia Insulei Plajei.
// ====================================================================

// ---------------- TRAMBULINE ----------------

type Trampoline struct {
	Pos     rl.Vector3 // centrul, la sol
	R       float32
	MaxApex float32 // cat de sus poate arunca (m deasupra panzei)
	Name    string
	Color   rl.Color
	Dip     float32 // cat de adanc e apasata panza acum
}

func (t *Trampoline) MatY() float32 { return t.Pos.Y + 0.75 + t.R*0.1 }

var padSizes = []struct {
	Name string
	R    float32
	Apex float32
	Col  rl.Color
}{
	{"Trambulina mică", 1.6, 7, LegoLime},
	{"Trambulina medie", 2.2, 12, LegoCyan},
	{"Trambulina mare", 3.0, 20, LegoOrange},
	{"Trambulina uriașă", 3.8, 28, LegoPurple},
	{"Trambulina gigant", 4.6, 40, LegoRed},
}

func newPad(x, z float32, size int, w *World) Trampoline {
	s := padSizes[size]
	return Trampoline{Pos: v3(x, maxf(w.NaturalLand(x, z), 0), z), R: s.R, MaxApex: s.Apex, Name: s.Name, Color: s.Col}
}

var padMatMesh rl.Mesh

// cadrul (picioare, arcuri, perna colorata) intra in geometria statica
func (cb *CityBuilder) padFrame(mb *MeshBuilder, t *Trampoline) {
	c := t.Pos
	my := t.MatY()
	n := 8
	for k := 0; k < n; k++ {
		a := float32(k) / float32(n) * 2 * math.Pi
		x, z := c.X+cosf(a)*(t.R+0.25), c.Z+sinf(a)*(t.R+0.25)
		mb.Cylinder(v3(x, c.Y, z), v3(x, my-0.05, z), 0.09, 0.09, 6, LegoDarkGrey, false)
		mb.Cylinder(v3(x, c.Y, z), v3(x, c.Y+0.06, z), 0.2, 0.2, 8, LegoDarkGrey, true)
	}
	// arcuri
	ns := int(t.R * 9)
	for k := 0; k < ns; k++ {
		a := float32(k) / float32(ns) * 2 * math.Pi
		mb.Cylinder(v3(c.X+cosf(a)*(t.R-0.05), my-0.02, c.Z+sinf(a)*(t.R-0.05)), v3(c.X+cosf(a)*(t.R+0.32), my+0.02, c.Z+sinf(a)*(t.R+0.32)), 0.035, 0.035, 4, LegoLightGrey, false)
	}
	// perna de protectie (inel colorat)
	seg := 32
	for k := 0; k < seg; k++ {
		a0 := float32(k) / float32(seg) * 2 * math.Pi
		a1 := float32(k+1) / float32(seg) * 2 * math.Pi
		col := t.Color
		if k%2 == 0 {
			col = shade(t.Color, 0.85)
		}
		ri, ro := t.R+0.1, t.R+0.6
		p := func(a, r, y float32) rl.Vector3 { return v3(c.X+cosf(a)*r, y, c.Z+sinf(a)*r) }
		mb.Quad(p(a0, ri, my+0.12), p(a1, ri, my+0.12), p(a1, ro, my+0.12), p(a0, ro, my+0.12), v3(0, 1, 0), col)
		mb.Quad(p(a0, ro, my-0.12), p(a1, ro, my-0.12), p(a1, ro, my+0.12), p(a0, ro, my+0.12), v3(cosf(a0), 0, sinf(a0)), shade(col, 0.9))
		mb.Quad(p(a0, ri, my-0.12), p(a1, ri, my-0.12), p(a1, ri, my+0.12), p(a0, ri, my+0.12), v3(-cosf(a0), 0, -sinf(a0)), shade(col, 0.8))
	}
}

func buildPadMat() {
	mb := NewMB()
	seg := 32
	for k := 0; k < seg; k++ {
		a0 := float32(k) / float32(seg) * 2 * math.Pi
		a1 := float32(k+1) / float32(seg) * 2 * math.Pi
		// panza: centrul coboara cand e apasata (scara pe y la desenare)
		mb.TriS(v3(0, -1, 0), v3(0, 1, 0), v3(cosf(a0), 0, sinf(a0)), v3(0, 1, 0), v3(cosf(a1), 0, sinf(a1)), v3(0, 1, 0), rl.NewColor(40, 44, 60, 255), rl.NewColor(50, 54, 72, 255), rl.NewColor(50, 54, 72, 255))
		mb.Quad(v3(cosf(a0)*0.3, -0.7+0.005, sinf(a0)*0.3), v3(cosf(a1)*0.3, -0.7+0.005, sinf(a1)*0.3), v3(cosf(a1)*0.38, -0.62+0.005, sinf(a1)*0.38), v3(cosf(a0)*0.38, -0.62+0.005, sinf(a0)*0.38), v3(0, 1, 0), LegoWhite)
	}
	padMatMesh = mb.Build()
}

func (t *Trampoline) DrawMat(draw DrawFn) {
	// panza e un con foarte turtit: varful coboara cu Dip (minim 4 cm, sa nu fie plata)
	d := maxf(t.Dip, 0.04)
	m := rl.MatrixScale(t.R, d, t.R)
	m = rl.MatrixMultiply(m, rl.MatrixTranslate(t.Pos.X, t.MatY(), t.Pos.Z))
	draw(padMatMesh, m, rl.White)
}

// ---------------- ROATA MARE ----------------

type FerrisWheel struct {
	Name      string
	Pos       rl.Vector3 // butucul
	R         float32
	N         int
	Ang       float32
	Speed     float32
	Hang      float32 // cat de jos atarna podeaua cabinei sub axul ei
	Scale     float32
	Mesh      rl.Mesh
	Cabin     rl.Mesh
	Board     rl.Vector3
	CabinCols []rl.Color
}

// CabinAxle: axul de care atarna cabina k (pe janta)
func (fw *FerrisWheel) CabinAxle(k int) rl.Vector3 {
	a := fw.Ang + float32(k)/float32(fw.N)*2*math.Pi
	return v3(fw.Pos.X+cosf(a)*fw.R, fw.Pos.Y+sinf(a)*fw.R, fw.Pos.Z)
}

// CabinFloor: unde sta Aventurierul in cabina k
func (fw *FerrisWheel) CabinFloor(k int) rl.Vector3 {
	p := fw.CabinAxle(k)
	return v3(p.X, p.Y-fw.Hang, p.Z)
}

// LowestCabin: cabina cea mai apropiata de sol
func (fw *FerrisWheel) LowestCabin() int {
	best, by := 0, float32(1e9)
	for k := 0; k < fw.N; k++ {
		if y := fw.CabinAxle(k).Y; y < by {
			best, by = k, y
		}
	}
	return best
}

func buildWheelMesh(R float32, n int, half float32) rl.Mesh {
	mb := NewMB()
	segs := int(R * 4)
	for _, z := range []float32{-half, half} {
		for i := 0; i < segs; i++ {
			a0 := float32(i) / float32(segs) * 2 * math.Pi
			a1 := float32(i+1) / float32(segs) * 2 * math.Pi
			mb.Cylinder(v3(cosf(a0)*R, sinf(a0)*R, z), v3(cosf(a1)*R, sinf(a1)*R, z), 0.26, 0.26, 6, LegoWhite, false)
			mb.Cylinder(v3(cosf(a0)*R*0.62, sinf(a0)*R*0.62, z), v3(cosf(a1)*R*0.62, sinf(a1)*R*0.62, z), 0.14, 0.14, 5, LegoWhite, false)
		}
		for i := 0; i < n*2; i++ {
			a := float32(i) / float32(n*2) * 2 * math.Pi
			mb.Cylinder(v3(0, 0, z), v3(cosf(a)*R, sinf(a)*R, z), 0.1, 0.1, 5, LegoWhite, false)
		}
	}
	// axele cabinelor (intre cele doua jante)
	for k := 0; k < n; k++ {
		a := float32(k) / float32(n) * 2 * math.Pi
		mb.Cylinder(v3(cosf(a)*R, sinf(a)*R, -half), v3(cosf(a)*R, sinf(a)*R, half), 0.12, 0.12, 6, LegoLightGrey, false)
	}
	mb.Cylinder(v3(0, 0, -half-0.6), v3(0, 0, half+0.6), 1.0, 1.0, 16, LegoDarkGrey, true)
	mb.Emis = 1
	cols := []rl.Color{LegoRed, LegoYellow, LegoCyan, LegoPink, LegoLime}
	nl := int(R * 2.4)
	for i := 0; i < nl; i++ {
		a := float32(i) / float32(nl) * 2 * math.Pi
		mb.Sphere(v3(cosf(a)*(R+0.3), sinf(a)*(R+0.3), half+0.15), 0.2, cols[i%len(cols)])
		mb.Sphere(v3(cosf(a)*(R+0.3), sinf(a)*(R+0.3), -half-0.15), 0.2, cols[(i+2)%len(cols)])
	}
	mb.Emis = 0
	return mb.Build()
}

// gondola deschisa (podea, balustrada, stalpi, acoperis rotund), axul la y = 0
func buildGondola(s float32) rl.Mesh {
	mb := NewMB()
	hang := 3.2 * s
	floorY := -hang
	mb.Cylinder(v3(0, 0, 0), v3(0, -0.9*s, 0), 0.08*s, 0.08*s, 6, LegoLightGrey, false)
	mb.EllipsoidF(v3(0, -0.95*s, 0), v3(1.5*s, 0.35*s, 1.3*s), 5, 14, func(n rl.Vector3) rl.Color { return rl.White })
	for _, x := range []float32{-1.2, 1.2} {
		for _, z := range []float32{-1.0, 1.0} {
			mb.Cylinder(v3(x*s, floorY, z*s), v3(x*s, -0.95*s, z*s), 0.07*s, 0.07*s, 5, LegoLightGrey, false)
		}
	}
	mb.BeveledBox(v3(0, floorY-0.1*s, 0), v3(2.7*s, 0.2*s, 2.3*s), 0.06, rl.White)
	// balustrada joasa (culoarea vine din tint)
	for _, z := range []float32{-1.1, 1.1} {
		mb.BeveledBox(v3(0, floorY+0.45*s, z*s), v3(2.6*s, 0.9*s, 0.1*s), 0.04, rl.White)
	}
	for _, x := range []float32{-1.3, 1.3} {
		mb.BeveledBox(v3(x*s, floorY+0.45*s, 0), v3(0.1*s, 0.9*s, 2.2*s), 0.04, rl.White)
	}
	return mb.Build()
}

func newWheel(name string, hub rl.Vector3, R float32, n int, scale float32) FerrisWheel {
	cols := []rl.Color{LegoRed, LegoYellow, LegoCyan, LegoPink, LegoLime, LegoOrange, LegoPurple, LegoBlue}
	fw := FerrisWheel{Name: name, Pos: hub, R: R, N: n, Speed: 0.9 / R, Hang: 3.2 * scale, Scale: scale,
		Mesh: buildWheelMesh(R, n, 2.2*scale+0.2), Cabin: buildGondola(scale), CabinCols: cols}
	fw.Board = v3(hub.X, 0, hub.Z+2.2*scale+4)
	return fw
}

// suportul in A al rotii, in geometria statica
func (cb *CityBuilder) wheelStand(mb *MeshBuilder, fw *FerrisWheel) {
	h := fw.Pos
	spread := fw.R * 0.55
	half := 2.2*fw.Scale + 0.9
	for _, z := range []float32{-half, half} {
		for _, sx := range []float32{-1, 1} {
			mb.Cylinder(v3(h.X+sx*spread, 0, h.Z+z*1.6), v3(h.X, h.Y, h.Z+z), 0.5, 0.35, 10, LegoWhite, false)
			cb.W.Circles = append(cb.W.Circles, CircleCollider{X: h.X + sx*spread, Z: h.Z + z*1.6, R: 0.7, Top: h.Y})
		}
		mb.Cylinder(v3(h.X-spread*0.5, h.Y*0.5, h.Z+z*1.3), v3(h.X+spread*0.5, h.Y*0.5, h.Z+z*1.3), 0.25, 0.25, 8, LegoWhite, false)
	}
	// platforma de urcare
	b := fw.Board
	mb.Cylinder(v3(b.X, 0, b.Z-2), v3(b.X, 0.15, b.Z-2), 6, 6, 32, rl.NewColor(240, 200, 220, 255), true)
	mb.Emis = 1
	mb.Disc(v3(b.X, 0.16, b.Z-2), 5.2, 5.6, 32, LegoGold)
	mb.Emis = 0
}

func (fw *FerrisWheel) Draw(draw DrawFn) {
	m := rl.MatrixMultiply(rl.MatrixRotateZ(fw.Ang), rl.MatrixTranslate(fw.Pos.X, fw.Pos.Y, fw.Pos.Z))
	draw(fw.Mesh, m, rl.White)
	for k := 0; k < fw.N; k++ {
		p := fw.CabinAxle(k)
		draw(fw.Cabin, rl.MatrixTranslate(p.X, p.Y, p.Z), fw.CabinCols[k%len(fw.CabinCols)])
	}
}

// ---------------- MONTAGNE RUSSE ----------------

type Coaster struct {
	Ctrl     []rl.Vector3
	Pts      []rl.Vector3 // la fiecare 0.5 m
	Len      float32
	LiftA    float32
	LiftB    float32
	StationS float32 // unde opreste trenul
	S        float32
	V        float32
	Wait     float32
	Running  bool
	CarMesh  rl.Mesh
	Rider    bool
	ClackT   float32
	Laps     int
	Traveled float32
}

const coasterCars = 3
const coasterGap = float32(3.3)

func catmull(p0, p1, p2, p3 rl.Vector3, t float32) rl.Vector3 {
	t2, t3 := t*t, t*t*t
	f := func(a, b, c, d float32) float32 {
		return 0.5 * (2*b + (-a+c)*t + (2*a-5*b+4*c-d)*t2 + (-a+3*b-3*c+d)*t3)
	}
	return v3(f(p0.X, p1.X, p2.X, p3.X), f(p0.Y, p1.Y, p2.Y, p3.Y), f(p0.Z, p1.Z, p2.Z, p3.Z))
}

func newCoaster() *Coaster {
	// (x, y, z): statia pe latura de sud, deal de ridicare spre est, cadere spre nord-vest
	ctrl := []rl.Vector3{
		v3(318, 1.0, 22), v3(330, 1.0, 22), v3(342, 1.0, 22),
		v3(360, 3.5, 24), v3(378, 13, 28), v3(394, 24, 32), v3(404, 30, 40),
		v3(404, 26, 52), v3(397, 9, 62), v3(382, 3, 69),
		v3(364, 16, 71), v3(346, 6, 70), v3(328, 11, 67),
		v3(312, 6, 58), v3(309, 12, 44), v3(311, 4, 31), v3(312, 1.0, 24),
	}
	c := &Coaster{Ctrl: ctrl}
	n := len(ctrl)
	dense := []rl.Vector3{}
	for i := 0; i < n; i++ {
		p0, p1, p2, p3 := ctrl[(i-1+n)%n], ctrl[i], ctrl[(i+1)%n], ctrl[(i+2)%n]
		for k := 0; k < 40; k++ {
			dense = append(dense, catmull(p0, p1, p2, p3, float32(k)/40))
		}
	}
	// reesantionare la 0.5 m
	cum := []float32{0}
	for i := 1; i <= len(dense); i++ {
		cum = append(cum, cum[i-1]+rl.Vector3Distance(dense[i-1], dense[i%len(dense)]))
	}
	total := cum[len(cum)-1]
	j := 0
	for s := float32(0); s < total; s += 0.5 {
		for j < len(dense)-1 && cum[j+1] < s {
			j++
		}
		t := (s - cum[j]) / maxf(cum[j+1]-cum[j], 1e-4)
		c.Pts = append(c.Pts, rl.Vector3Lerp(dense[j], dense[(j+1)%len(dense)], t))
	}
	c.Len = float32(len(c.Pts)) * 0.5
	// deal de ridicare (lantul): de la ctrl[3] pana in varf (ctrl[6])
	c.LiftA = c.nearestS(ctrl[3])
	c.LiftB = c.nearestS(ctrl[6])
	c.StationS = c.nearestS(ctrl[1]) + 4
	c.S = c.StationS
	c.Wait = 3
	c.CarMesh = buildCoasterCar()
	return c
}

func (c *Coaster) nearestS(p rl.Vector3) float32 {
	best, bd := 0, float32(1e9)
	for i, q := range c.Pts {
		if d := rl.Vector3Distance(p, q); d < bd {
			best, bd = i, d
		}
	}
	return float32(best) * 0.5
}

func (c *Coaster) wrap(s float32) float32 {
	s = float32(math.Mod(float64(s), float64(c.Len)))
	if s < 0 {
		s += c.Len
	}
	return s
}

// At: pozitia pe sina si directia (tangenta)
func (c *Coaster) At(s float32) (rl.Vector3, rl.Vector3) {
	s = c.wrap(s)
	i := int(s / 0.5)
	f := s/0.5 - float32(i)
	a := c.Pts[i%len(c.Pts)]
	b := c.Pts[(i+1)%len(c.Pts)]
	p := rl.Vector3Lerp(a, b, f)
	t := rl.Vector3Normalize(rl.Vector3Subtract(c.Pts[(i+2)%len(c.Pts)], c.Pts[(i-1+len(c.Pts))%len(c.Pts)]))
	return p, t
}

func (c *Coaster) inRange(s, a, b float32) bool {
	s = c.wrap(s)
	if a <= b {
		return s >= a && s <= b
	}
	return s >= a || s <= b
}

// distanta de mers pana la oprirea din statie
func (c *Coaster) toStation() float32 {
	return c.wrap(c.StationS - c.S)
}

// Update: true in cadrul in care trenul s-a oprit in statie
func (c *Coaster) Update(dt float32, audio *GameAudio, listener rl.Vector3) bool {
	if !c.Running {
		c.Wait -= dt
		if c.Wait <= 0 {
			c.Running = true
			c.V = 1.5
		}
		return false
	}
	p, t := c.At(c.S)
	// din statie pana in varful dealului trenul e tras (roti motoare, apoi lantul)
	if c.inRange(c.S, c.StationS-2, c.LiftB) {
		if c.V < 5.5 {
			c.V = minf(c.V+dt*4, 5.5)
		}
		c.ClackT -= dt
		if c.ClackT <= 0 && audio != nil {
			c.ClackT = 0.22
			audio.PlayAt(audio.Clack, p, listener, 0.8, 60)
		}
	} else {
		c.V += (-9.8*t.Y - 0.015*c.V*c.V*0.1 - 0.12) * dt
		if c.V < 2 {
			c.V = 2
		}
	}
	// franele dinaintea statiei (doar dupa ce a facut aproape tot circuitul)
	nearEnd := c.Traveled > c.Len-60
	if d := c.toStation(); nearEnd && d < 40 {
		target := maxf(1.2, d*0.35)
		if c.V > target {
			c.V = maxf(target, c.V-dt*9)
		}
		if d < 0.3 {
			c.S = c.StationS
			c.V = 0
			c.Running = false
			c.Wait = 5
			c.Traveled = 0
			c.Laps++
			return true
		}
	}
	step := c.V * dt
	if d := c.toStation(); nearEnd && step > d && d < 40 {
		step = d
	}
	c.S = c.wrap(c.S + step)
	c.Traveled += step
	return false
}

// CarPose: pozitia si orientarea vagonului k
func (c *Coaster) CarPose(k int) (rl.Vector3, float32, float32) {
	p, t := c.At(c.S - float32(k)*coasterGap)
	yaw := atan2f(t.X, t.Z)
	pitch := -float32(math.Asin(float64(clampf(t.Y, -1, 1))))
	return p, yaw, pitch
}

func (c *Coaster) carMatrix(k int) rl.Matrix {
	p, yaw, pitch := c.CarPose(k)
	m := rl.MatrixMultiply(rl.MatrixRotateX(pitch), rl.MatrixRotateY(yaw))
	return rl.MatrixMultiply(m, rl.MatrixTranslate(p.X, p.Y, p.Z))
}

// SeatPos: unde sta Aventurierul (in primul vagon)
func (c *Coaster) SeatPos() rl.Vector3 {
	m := c.carMatrix(0)
	return rl.Vector3Transform(v3(0, 0.35, -0.2), m)
}

func (c *Coaster) Draw(draw DrawFn) {
	cols := []rl.Color{LegoRed, LegoYellow, LegoBlue}
	for k := 0; k < coasterCars; k++ {
		draw(c.CarMesh, c.carMatrix(k), cols[k])
	}
}

func buildCoasterCar() rl.Mesh {
	mb := NewMB()
	mb.BeveledBox(v3(0, 0.55, 0), v3(1.7, 0.7, 2.8), 0.2, rl.White)
	mb.BeveledBox(v3(0, 1.05, 1.2), v3(1.7, 0.5, 0.35), 0.12, rl.White)
	mb.BeveledBox(v3(0, 0.95, -0.2), v3(1.3, 0.15, 1.6), 0.05, LegoDarkGrey)
	mb.BeveledBox(v3(0, 1.35, -1.0), v3(1.3, 0.9, 0.2), 0.06, LegoDarkGrey)
	for _, x := range []float32{-0.6, 0.6} {
		for _, z := range []float32{-1.0, 1.0} {
			mb.Cylinder(v3(x-0.12, 0.15, z), v3(x+0.12, 0.15, z), 0.2, 0.2, 10, LegoDarkGrey, true)
		}
	}
	mb.Emis = 1
	mb.Sphere(v3(-0.55, 0.7, 1.45), 0.12, LegoYellow)
	mb.Sphere(v3(0.55, 0.7, 1.45), 0.12, LegoYellow)
	mb.Emis = 0
	return mb.Build()
}

// sina, traversele, stalpii si statia (geometrie statica)
func (cb *CityBuilder) coasterTrack(c *Coaster) {
	mb := NewMB()
	up := v3(0, 1, 0)
	nPts := len(c.Pts)
	for i := 0; i < nPts; i += 2 {
		a := c.Pts[i]
		b := c.Pts[(i+2)%nPts]
		t := rl.Vector3Normalize(rl.Vector3Subtract(b, a))
		side := rl.Vector3Normalize(rl.Vector3CrossProduct(t, up))
		for _, s := range []float32{-0.6, 0.6} {
			o := rl.Vector3Scale(side, s)
			mb.Cylinder(rl.Vector3Add(a, o), rl.Vector3Add(b, o), 0.1, 0.1, 6, LegoRed, false)
		}
		mb.Cylinder(rl.Vector3Add(a, v3(0, -0.45, 0)), rl.Vector3Add(b, v3(0, -0.45, 0)), 0.22, 0.22, 8, LegoYellow, false)
		if i%3 == 0 {
			mb.Cylinder(rl.Vector3Add(a, rl.Vector3Scale(side, -0.65)), rl.Vector3Add(a, rl.Vector3Scale(side, 0.65)), 0.06, 0.06, 4, LegoDarkGrey, false)
			mb.Cylinder(a, rl.Vector3Add(a, v3(0, -0.45, 0)), 0.06, 0.06, 4, LegoDarkGrey, false)
		}
	}
	// stalpi la fiecare ~7 m
	for i := 0; i < nPts; i += 14 {
		p := c.Pts[i]
		g := maxf(cb.W.NaturalLand(p.X, p.Z), 0)
		if p.Y-g < 1.6 {
			continue
		}
		mb.Cylinder(v3(p.X, g, p.Z), v3(p.X, p.Y-0.5, p.Z), 0.32, 0.26, 8, LegoWhite, false)
		mb.Cylinder(v3(p.X, g, p.Z), v3(p.X, g+0.25, p.Z), 0.6, 0.6, 10, ColConcrete, true)
		cb.W.Circles = append(cb.W.Circles, CircleCollider{X: p.X, Z: p.Z, R: 0.45, Top: p.Y - 0.6})
	}
	// statia: peron + acoperis
	s0 := c.Ctrl[0]
	s2 := c.Ctrl[2]
	pz := s0.Z - 2.3
	mb.BeveledBox(v3((s0.X+s2.X)/2, 0.25, pz), v3(s2.X-s0.X+4, 0.5, 2.6), 0.08, rl.NewColor(230, 214, 190, 255))
	cb.W.Boxes = append(cb.W.Boxes, BoxCollider{MinX: s0.X - 2, MaxX: s2.X + 2, MinZ: pz - 1.3, MaxZ: pz + 1.3, Top: 0.5, Roof: RoofFlat})
	for x := s0.X - 1; x <= s2.X+1; x += 6 {
		mb.Cylinder(v3(x, 0.5, pz-1.1), v3(x, 4.6, pz-1.1), 0.15, 0.15, 8, LegoRed, false)
		cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: pz - 1.1, R: 0.3, Top: 4.6})
	}
	mb.BeveledBox(v3((s0.X+s2.X)/2, 4.75, s0.Z-0.6), v3(s2.X-s0.X+5, 0.3, 6.4), 0.1, LegoYellow)
	mb.Emis = 1
	mb.Box(v3((s0.X+s2.X)/2, 5.4, pz-1.0), v3(8, 1.0, 0.2), LegoRed)
	mb.Box(v3((s0.X+s2.X)/2, 5.4, pz-0.88), v3(6.8, 0.3, 0.05), LegoYellow)
	mb.Emis = 0
	cb.addChunk(mb, true)
}

// ---------------- CONSTRUCTIA INSULELOR NOI ----------------

func (cb *CityBuilder) smallHouse(mb *MeshBuilder, x, z, w, d float32, floors int, wall, roof rl.Color, door rl.Vector2) {
	th := Theme{Walls: []rl.Color{wall}, Roofs: []rl.Color{roof}, MinFloors: floors, MaxFloors: floors, Pitched: 1}
	cb.building(mb, x, z, w, d, floors, th, door)
}

// Insula Plajei: plaja larga, far, case colorate, trambuline langa case
func (cb *CityBuilder) buildBeachIsland() {
	mb := NewMB()
	cx := WestX
	// sens giratoriu la capatul rampei Marelui Tunel
	cb.roundabout(mb, rl.NewVector2(-401, 0), 9, 4, ColGrass)
	cb.Reserve = append(cb.Reserve, CircleCollider{X: -401, Z: 0, R: 12}, CircleCollider{X: -366, Z: 0, R: 30})
	// farul
	fx, fz := float32(-398), float32(-50)
	for k := 0; k < 7; k++ {
		col := LegoRed
		if k%2 == 1 {
			col = LegoWhite
		}
		y0 := float32(k) * 3
		mb.Cylinder(v3(fx, y0, fz), v3(fx, y0+3, fz), 2.6-float32(k)*0.18, 2.6-float32(k+1)*0.18, 20, col, false)
	}
	mb.Cylinder(v3(fx, 21, fz), v3(fx, 21.4, fz), 2.2, 2.2, 20, LegoDarkGrey, true)
	mb.Emis = 1
	mb.Cylinder(v3(fx, 21.4, fz), v3(fx, 23.2, fz), 1.2, 1.2, 16, LegoGoldBright, false)
	mb.Emis = 0
	mb.Cylinder(v3(fx, 23.2, fz), v3(fx, 24.6, fz), 1.5, 0.1, 16, LegoRed, true)
	cb.W.Circles = append(cb.W.Circles, CircleCollider{X: fx, Z: fz, R: 2.7, Top: 24})
	cb.Reserve = append(cb.Reserve, CircleCollider{X: fx, Z: fz, R: 5})
	// case colorate de plaja (acoperis in doua ape, se poate sari pe ele)
	walls := []rl.Color{{255, 200, 120, 255}, {150, 220, 255, 255}, {255, 170, 200, 255}, {190, 240, 170, 255}, {255, 236, 150, 255}, {210, 190, 255, 255}}
	roofs := []rl.Color{{220, 70, 60, 255}, {60, 120, 220, 255}, {240, 140, 40, 255}}
	houses := [][3]float32{{-380, 26, 1}, {-360, 30, 2}, {-338, 26, 1}, {-320, 34, 1}, {-380, -26, 2}, {-356, -32, 1}, {-334, -26, 1}, {-318, -36, 2}}
	for i, h := range houses {
		door := rl.NewVector2(0, -1)
		if h[1] < 0 {
			door = rl.NewVector2(0, 1)
		}
		cb.smallHouse(mb, h[0], h[1], 9, 8, int(h[2]), walls[i%len(walls)], roofs[i%len(roofs)], door)
		cb.Reserve = append(cb.Reserve, CircleCollider{X: h[0], Z: h[1], R: 7})
	}
	// umbrele de soare si sezlonguri pe plaja (nord si sud)
	cols := []rl.Color{LegoRed, LegoYellow, LegoCyan, LegoPink, LegoOrange}
	k := 0
	for _, z := range []float32{56, -56} {
		for x := cx - 40; x < cx+44; x += 12 {
			px := x + float32(k%3)*1.5
			if rl.Vector2Distance(rl.NewVector2(px, z), WestBridgeB) < 12 {
				k++
				continue
			}
			y := cb.W.NaturalLand(px, z)
			mb.Cylinder(v3(px, y, z), v3(px, y+2.6, z), 0.07, 0.07, 5, LegoWhite, false)
			mb.CylinderC(v3(px, y+2.4, z), v3(px, y+3.1, z), 1.9, 0.05, 12, cols[k%len(cols)], LegoWhite, true)
			mb.BeveledBox(v3(px+1.6, y+0.3, z), v3(0.8, 0.12, 1.9), 0.04, cols[(k+2)%len(cols)])
			k++
		}
	}
	cb.addChunk(mb, true)
	// trambuline langa case
	cb.W.Pads = append(cb.W.Pads, newPad(-349, 24, 1, cb.W), newPad(-369, -22, 2, cb.W))
	for _, p := range cb.W.Pads[len(cb.W.Pads)-2:] {
		cb.Reserve = append(cb.Reserve, CircleCollider{X: p.Pos.X, Z: p.Pos.Z, R: p.R + 1.5})
	}
}

// Parcul de Distractii: roata uriasa, montagne russe, parcul de trambuline, turnuri de sarit
func (cb *CityBuilder) buildFunPark(cw *CityWorld) {
	mb := NewMB()
	cb.roundabout(mb, rl.NewVector2(401, 0), 9, 4, LegoPink)
	cb.Reserve = append(cb.Reserve, CircleCollider{X: 401, Z: 0, R: 12})
	// alei pavate colorate
	paths := [][4]float32{{310, -30, 420, -30}, {330, -30, 330, 14}, {360, 14, 360, -76}, {310, 12, 420, 12}}
	for _, p := range paths {
		a, b := rl.NewVector2(p[0], p[1]), rl.NewVector2(p[2], p[3])
		d := rl.Vector2Normalize(rl.Vector2Subtract(b, a))
		per := rl.NewVector2(-d.Y, d.X)
		L := rl.Vector2Distance(a, b)
		for t := float32(0); t < L; t += 2 {
			c := ColPaving
			if int(t/2)%2 == 0 {
				c = rl.NewColor(240, 200, 210, 255)
			}
			q := func(tt, l float32) rl.Vector3 {
				return v3(a.X+d.X*tt+per.X*l, 0.06, a.Y+d.Y*tt+per.Y*l)
			}
			if cb.W.nearOpenCut(a.X+d.X*t, a.Y+d.Y*t, 2) {
				continue
			}
			mb.Quad(q(t, -1.8), q(t+2, -1.8), q(t+2, 1.8), q(t, 1.8), v3(0, 1, 0), c)
		}
	}
	// roata uriasa
	cw.Wheels = append(cw.Wheels, newWheel("Roata Uriașă", v3(330, 34, -48), 30, 16, 1.0))
	fw := &cw.Wheels[len(cw.Wheels)-1]
	cb.wheelStand(mb, fw)
	cb.Reserve = append(cb.Reserve, CircleCollider{X: 330, Z: -48, R: 20})
	// turnuri colorate de pe care se sare (acoperis plat)
	towers := [][5]float32{{414, -50, 8, 8, 6}, {358, -66, 8, 8, 4}, {388, -38, 7, 7, 3}}
	tc := []rl.Color{{255, 160, 200, 255}, {160, 220, 255, 255}, {255, 230, 140, 255}}
	for i, t := range towers {
		th := Theme{Walls: []rl.Color{tc[i]}, Roofs: []rl.Color{LegoPurple}, MinFloors: int(t[4]), MaxFloors: int(t[4]), Pitched: 0}
		cb.building(mb, t[0], t[1], t[2], t[3], int(t[4]), th, rl.NewVector2(-1, 0))
	}
	// chioscuri cu inghetata si baloane
	for i, p := range [][2]float32{{318, 2 - 14}, {372, 22}, {400, 26}, {346, -20}} {
		x, z := p[0], p[1]
		mb.BeveledBox(v3(x, 1.2, z), v3(2.6, 2.4, 2.0), 0.15, LegoWhite)
		mb.CylinderC(v3(x, 2.4, z), v3(x, 3.3, z), 2.2, 0.1, 12, []rl.Color{LegoRed, LegoBlue, LegoPink, LegoOrange}[i], LegoWhite, true)
		for b := 0; b < 4; b++ {
			bx := x + 1.4 + float32(b%2)*0.5
			bz := z - 0.5 + float32(b/2)*0.6
			mb.Cylinder(v3(bx, 1.0, bz), v3(bx, 3.6+float32(b)*0.3, bz), 0.015, 0.015, 3, LegoWhite, false)
			mb.Ellipsoid(v3(bx, 3.9+float32(b)*0.3, bz), v3(0.32, 0.4, 0.32), 5, 8, []rl.Color{LegoRed, LegoYellow, LegoCyan, LegoLime}[(b+i)%4])
		}
		cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 1.6, Top: 3})
	}
	// felinare pe alei
	for i := 0; i < 10; i++ {
		x := float32(312 + i*11)
		cb.streetLamp(mb, x, 0, -27.5, 0, -1)
		cb.streetLamp(mb, x, 0, 9.5, 0, -1)
	}
	cb.addChunk(mb, true)
	// montagne russe
	cw.Coaster = newCoaster()
	cb.coasterTrack(cw.Coaster)
	// parcul de trambuline: de la cea mica la cea gigant
	pads := []struct {
		x, z float32
		s    int
	}{{372, -40, 0}, {384, -48, 1}, {398, -56, 2}, {376, -62, 3}, {400, -70, 4}, {346, -44, 2}}
	pm := NewMB()
	for _, p := range pads {
		t := newPad(p.x, p.z, p.s, cb.W)
		cb.W.Pads = append(cb.W.Pads, t)
		cb.padFrame(pm, &cb.W.Pads[len(cb.W.Pads)-1])
	}
	cb.addChunk(pm, true)
}

// trambuline langa blocurile din oras: puse DUPA cladiri, lipite de perete (1 m),
// fiecare destul de puternica sa-l arunce pe Aventurierul mai sus de acoperisul blocului.
func (cb *CityBuilder) reserveCityPads() {
	cb.W.Pads = append(cb.W.Pads, newPad(52, 24, 1, cb.W))
	cb.Reserve = append(cb.Reserve, CircleCollider{X: 52, Z: 24, R: 4})
}

func (cb *CityBuilder) placeRoofPads() {
	w := cb.W
	perSector := []int{3, 3, 3, 4}
	for si := 0; si < 4; si++ {
		type cand struct {
			x, z float32
			size int
			top  float32
		}
		var cands []cand
		var why [8]int
		for bi := range w.Boxes {
			b := &w.Boxes[bi]
			if b.Roof == RoofNone || b.Top < 5 || b.Top > 34 || (b.MaxX-b.MinX) < 6 {
				why[0]++
				continue
			}
			cx, cz := (b.MinX+b.MaxX)/2, (b.MinZ+b.MaxZ)/2
			if SectorAt(cx, cz) != si {
				continue
			}
			top := b.SurfaceY(cx, cz)
			size := -1
			for k := 1; k < len(padSizes); k++ {
				if padSizes[k].Apex >= top+4 {
					size = k
					break
				}
			}
			if size < 0 {
				continue
			}
			R := padSizes[size].R
			gap := R + 1.0 // marginea trambulinei la 1 m de perete
			// toate laturile, in 3 locuri pe latura (mijloc, apoi spre colturi)
			var sides [][2]float32
			for _, f := range []float32{0, -0.3, 0.3} {
				ox := (b.MaxX - b.MinX) * f
				oz := (b.MaxZ - b.MinZ) * f
				sides = append(sides, [2]float32{b.MaxX + gap, cz + oz}, [2]float32{b.MinX - gap, cz + oz}, [2]float32{cx + ox, b.MaxZ + gap}, [2]float32{cx + ox, b.MinZ - gap})
			}
			for _, sp := range sides {
				x, z := sp[0], sp[1]
				if w.NaturalLand(x, z) < 0 || w.LandSD(x, z) > -GrassInset-R-1 {
					why[1]++
					continue
				}
				if cb.nearRoad(x, z, R+0.2) || cb.inMountain(x, z, R+1) || w.nearOpenCut(x, z, R+2) {
					why[2]++
					continue
				}
				// nimic altceva in loc (alta cladire, santier, reper)
				if cb.reserved(x, z, R+0.4) || w.Solid(x, z, 0.5, R+0.45) {
					why[3]++
					continue
				}
				clash := false
				for _, p := range w.Pads {
					dx, dz := p.Pos.X-x, p.Pos.Z-z
					if dx*dx+dz*dz < (p.R+R+14)*(p.R+R+14) {
						clash = true
						break
					}
				}
				for _, c := range cands {
					dx, dz := c.x-x, c.z-z
					if dx*dx+dz*dz < 22*22 {
						clash = true
						break
					}
				}
				if clash {
					why[4]++
					continue
				}
				cands = append(cands, cand{x, z, size, top})
				break
			}
		}
		// intai blocurile mai inalte (mai spectaculos), dar pastram si unul mic
		for i := 1; i < len(cands); i++ {
			for j := i; j > 0 && cands[j].top > cands[j-1].top; j-- {
				cands[j], cands[j-1] = cands[j-1], cands[j]
			}
		}
		if os.Getenv("BATTLE_DEBUG") != "" {
			fmt.Println("trambuline bloc sector", si, "candidati", len(cands), "respinse", why)
		}
		n := minInt(perSector[si], len(cands))
		for k := 0; k < n; k++ {
			c := cands[k]
			if k == n-1 && len(cands) > n {
				c = cands[len(cands)-1] // cel mai jos bloc, pentru inceput
			}
			w.Pads = append(w.Pads, newPad(c.x, c.z, c.size, w))
			R := padSizes[c.size].R
			cb.Reserve = append(cb.Reserve, CircleCollider{X: c.x, Z: c.z, R: R + 1.5})
		}
	}
}

func (cb *CityBuilder) buildPadFrames(from int) {
	mb := NewMB()
	for i := from; i < len(cb.W.Pads); i++ {
		cb.padFrame(mb, &cb.W.Pads[i])
	}
	cb.addChunk(mb, true)
}
