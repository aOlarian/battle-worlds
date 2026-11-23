package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// HARTA V 4.1
//   5 insule ca punctele de pe zar (NV Dino Land, NE Cartoon City,
//   SV Turbo Speedway, SE Mecha Arena, centru Royal Round Plaza)
//   + Insula Plajei (vest) + Parcul de Distractii (est).
// Legaturi:
//   - 4 poduri mari intre insulele exterioare;
//   - pe fiecare diagonala: pod scurt -> munte cu TUNEL LUNG -> insula;
//   - Marele Tunel de sub Mare: din vest pana in est, pe sub mare si pe sub
//     piata centrala, cu iesiri pe scari in piata si lifturi pe 2 insulite.
// ====================================================================

const (
	SecNW_X  = -160.0
	SecNW_Z  = -160.0
	SecNE_X  = 160.0
	SecNE_Z  = -160.0
	SecSW_X  = -160.0
	SecSW_Z  = 160.0
	SecSE_X  = 160.0
	SecSE_Z  = 160.0
	SecCTR_X = 0.0
	SecCTR_Z = 0.0

	WaterY  = float32(-1.0)
	SeabedY = float32(-7.0)

	IslandHalf   = float32(65)  // jumatate de latura insula exterioara (pana la marginea plajei)
	IslandCorner = float32(16)  // raza colturilor rotunjite
	CenterR      = float32(64)  // raza insulei centrale (marginea plajei)
	GrassInset   = float32(6)   // latimea plajei
	IsletD       = float32(112) // centrul tunelului de munte, pe diagonala
	PortalIn     = float32(16)  // portalul dinspre centru, la u = -16
	PortalOut    = float32(46)  // portalul dinspre insula, la u = +46 (tunel de 62 m)
	MountAuIn    = float32(34)  // muntele se intinde de la u = -34 ...
	MountAuOut   = float32(64)  // ... pana la u = +64
	TubeHalfW    = float32(6.5) // jumatate din latimea tunelului
	TrenchHalfW  = float32(7.0) // jumatate din latimea santului de acces
	TubeWallH    = float32(3.0) // inaltimea peretilor drepti
	MountainAv   = float32(26)

	WestX          = float32(-356) // Insula Plajei
	WestHalfU      = float32(60)
	WestHalfV      = float32(70)
	EastX          = float32(362) // Parcul de Distractii
	EastHalfU      = float32(64)
	EastHalfV      = float32(84)
	VentX          = float32(256) // insulitele cu lift ale Marelui Tunel (x = -256 si +256)
	VentR          = float32(13)
	mapEdgeIslandX = float32(446) // marginea lumii: pana aici poate inota Aventurierul (intre insule oriunde)
	mapEdgeIslandZ = float32(268)
)

type SectorInfo struct {
	Name   string
	Center rl.Vector2
	Inner  rl.Vector2 // directia spre centrul hartii (semne)
}

const (
	SecWest = 5
	SecPark = 6
)

var Sectors = []SectorInfo{
	{"Dino Land", rl.NewVector2(SecNW_X, SecNW_Z), rl.NewVector2(1, 1)},
	{"Cartoon City", rl.NewVector2(SecNE_X, SecNE_Z), rl.NewVector2(-1, 1)},
	{"Turbo Speedway", rl.NewVector2(SecSW_X, SecSW_Z), rl.NewVector2(1, -1)},
	{"Mecha Arena", rl.NewVector2(SecSE_X, SecSE_Z), rl.NewVector2(-1, -1)},
	{"Royal Round Plaza", rl.NewVector2(0, 0), rl.NewVector2(0, 0)},
	{"Insula Plajei", rl.NewVector2(WestX, 0), rl.NewVector2(1, 0)},
	{"Parcul de Distracții", rl.NewVector2(EastX, 0), rl.NewVector2(-1, 0)},
}

// --------------------------------------------------------------------
// Forme de uscat
// --------------------------------------------------------------------

type LandShape struct {
	Kind   int // 0 = dreptunghi rotunjit aliniat, 1 = cerc, 2 = dreptunghi rotit
	C      rl.Vector2
	HalfU  float32
	HalfV  float32
	Corner float32
	R      float32
	Dir    rl.Vector2 // axa u pentru Kind 2
}

func sdRoundRect(px, pz, hx, hz, r float32) float32 {
	qx := absf(px) - (hx - r)
	qz := absf(pz) - (hz - r)
	ox := qx
	if ox < 0 {
		ox = 0
	}
	oz := qz
	if oz < 0 {
		oz = 0
	}
	in := qx
	if qz > in {
		in = qz
	}
	if in > 0 {
		in = 0
	}
	return sqrtf(ox*ox+oz*oz) + in - r
}

func (s *LandShape) SD(x, z float32) float32 {
	dx := x - s.C.X
	dz := z - s.C.Y
	switch s.Kind {
	case 1:
		return sqrtf(dx*dx+dz*dz) - s.R
	case 2:
		u := dx*s.Dir.X + dz*s.Dir.Y
		v := -dx*s.Dir.Y + dz*s.Dir.X
		return sdRoundRect(u, v, s.HalfU, s.HalfV, s.Corner)
	default:
		return sdRoundRect(dx, dz, s.HalfU, s.HalfV, s.Corner)
	}
}

// profilul de inaltime in functie de distanta fata de mal (sd < 0 = pe uscat)
func landProfile(sd float32) float32 {
	switch {
	case sd <= -GrassInset:
		return 0
	case sd <= 0:
		t := (sd + GrassInset) / GrassInset
		return -0.55 * t * t
	case sd <= 6:
		return -0.55 - 2.6*(sd/6)
	case sd <= 16:
		return -3.15 - 3.85*smoothstep(0, 1, (sd-6)/10)
	}
	return SeabedY
}

// --------------------------------------------------------------------
// Munte cu tunel (asimetric: tunelul lung trece pe sub o creasta)
// --------------------------------------------------------------------

type Mountain struct {
	Index int
	C     rl.Vector2 // originea axei tunelului (pe diagonala)
	Dir   rl.Vector2 // u: de la centrul hartii spre insula exterioara
	Perp  rl.Vector2 // v
	Neon  rl.Color
	Seed  float32
	Peak1 float32
	Peak2 float32
	Name  string
}

func (m *Mountain) Local(x, z float32) (u, v float32) {
	dx := x - m.C.X
	dz := z - m.C.Y
	u = dx*m.Dir.X + dz*m.Dir.Y
	v = dx*m.Perp.X + dz*m.Perp.Y
	return
}

func (m *Mountain) World(u, v float32) (x, z float32) {
	x = m.C.X + m.Dir.X*u + m.Perp.X*v
	z = m.C.Y + m.Dir.Y*u + m.Perp.Y*v
	return
}

func bump(t float32) float32 {
	if t <= -1 || t >= 1 {
		return 0
	}
	a := 1 - t*t
	return a * a
}

func expf(x float32) float32 { return float32(math.Exp(float64(x))) }

func powf(a, b float32) float32 { return float32(math.Pow(float64(a), float64(b))) }

func sq(x float32) float32 { return x * x }

// uN: u normalizat in [-1, 1] (muntele e mai lung spre insula)
func uN(u float32) float32 {
	if u < 0 {
		return u / MountAuIn
	}
	return u / MountAuOut
}

// Raw da inaltimea naturala a muntelui (fara sant).
func (m *Mountain) Raw(u, v float32) float32 {
	t := absf(uN(u))
	if t >= 1 {
		return 0
	}
	ru := 1 - smoothstep(0.45, 1.0, t)
	lat := 14*bump(v/MountainAv) +
		m.Peak1*expf(-sq((v+14)/7))*(0.55+0.45*expf(-sq((u+2)/18))) +
		m.Peak2*expf(-sq((v-12)/7.5))*(0.55+0.45*expf(-sq((u-28)/20)))
	n := (fbm(u*0.09+m.Seed, v*0.09-m.Seed, 4) - 0.5) * 8 * bump(v/28)
	h := powf(ru, 0.6) * (lat + n)
	// acoperire garantata deasupra tunelului (tavanul e la 9.5 m)
	var mu float32
	if u < 0 {
		mu = 1 - smoothstep(PortalIn, PortalIn+6, -u)
	} else {
		mu = 1 - smoothstep(PortalOut, PortalOut+6, u)
	}
	mask := (1 - smoothstep(8.5, 13, absf(v))) * mu
	if c := 12.2 * mask; c > h {
		h = c
	}
	if h < 0 {
		h = 0
	}
	return h
}

// InTrench: in santul de acces (fara munte deasupra), inaintea portalurilor.
func (m *Mountain) InTrench(u, v float32) bool {
	return absf(v) < TrenchHalfW && (u > PortalOut || u < -PortalIn)
}

// Effective: inaltimea muntelui pentru fizica si randare (0 in sant).
func (m *Mountain) Effective(u, v float32) float32 {
	if m.InTrench(u, v) {
		return 0
	}
	return m.Raw(u, v)
}

// InVoid: punctul e in golul tunelului (acolo muntele nu e solid).
func (m *Mountain) InVoid(u, v, y float32) bool {
	return u >= -PortalIn-1.2 && u <= PortalOut+1.2 && absf(v) < TubeHalfW-0.2 && y < 8.5
}

// InTube: interiorul acoperit al tunelului.
func (m *Mountain) InTube(u, v, y float32) bool {
	return u >= -PortalIn && u <= PortalOut && absf(v) < TubeHalfW && y < 8
}

// --------------------------------------------------------------------
// Poduri
// --------------------------------------------------------------------

const (
	BridgeSuspension = iota
	BridgeStoneArch
	BridgeCableStayed
	BridgeSteelArch
	BridgeBeam
)

type Bridge struct {
	Kind   int
	Name   string
	A, B   rl.Vector2
	Width  float32
	Apex   float32
	Dir    rl.Vector2
	Perp   rl.Vector2
	Length float32
	Color  rl.Color
}

func NewBridge(kind int, name string, a, b rl.Vector2, width, apex float32, col rl.Color) Bridge {
	d := rl.Vector2Subtract(b, a)
	L := rl.Vector2Length(d)
	dir := rl.Vector2Scale(d, 1/L)
	return Bridge{Kind: kind, Name: name, A: a, B: b, Width: width, Apex: apex, Dir: dir, Perp: rl.NewVector2(-dir.Y, dir.X), Length: L, Color: col}
}

// Local: t in [0,1] de-a lungul podului, lat = distanta laterala.
func (b *Bridge) Local(x, z float32) (t, lat float32) {
	dx := x - b.A.X
	dz := z - b.A.Y
	along := dx*b.Dir.X + dz*b.Dir.Y
	lat = dx*b.Perp.X + dz*b.Perp.Y
	t = along / b.Length
	return
}

func (b *Bridge) Point(t, lat float32) (x, z float32) {
	x = b.A.X + b.Dir.X*t*b.Length + b.Perp.X*lat
	z = b.A.Y + b.Dir.Y*t*b.Length + b.Perp.Y*lat
	return
}

// DeckY: inaltimea soselei pe pod (urcare lina, ca la podurile reale).
func (b *Bridge) DeckY(t float32) float32 {
	t = clampf(t, 0, 1)
	if b.Kind == BridgeStoneArch {
		s := 2*t - 1
		e := smoothstep(0, 0.12, t) * smoothstep(0, 0.12, 1-t)
		return b.Apex * (1 - s*s) * (0.35 + 0.65*e)
	}
	return b.Apex * (0.5 - 0.5*cosf(2*math.Pi*t))
}

// --------------------------------------------------------------------
// Obstacole
// --------------------------------------------------------------------

// Tipul acoperisului unei cutii (pe ce se poate sta deasupra)
const (
	RoofNone   = 0 // nu se sta pe ea (doar platformele joase, sub 1 m)
	RoofFlat   = 1 // acoperis plat la Top
	RoofRidgeX = 2 // in doua ape, coama paralela cu axa X
	RoofRidgeZ = 3 // in doua ape, coama paralela cu axa Z
)

type BoxCollider struct {
	MinX, MinZ, MaxX, MaxZ float32
	Top                    float32
	Roof                   int
	RoofH                  float32
}

// SurfaceY: inaltimea fetei de sus in (x,z) (acoperisurile in doua ape urca spre coama).
func (b *BoxCollider) SurfaceY(x, z float32) float32 {
	switch b.Roof {
	case RoofRidgeX:
		hd := (b.MaxZ - b.MinZ) / 2
		cz := (b.MaxZ + b.MinZ) / 2
		return b.Top + b.RoofH*clampf(1-absf(z-cz)/hd, 0, 1)
	case RoofRidgeZ:
		hw := (b.MaxX - b.MinX) / 2
		cx := (b.MaxX + b.MinX) / 2
		return b.Top + b.RoofH*clampf(1-absf(x-cx)/hw, 0, 1)
	}
	return b.Top
}

type CircleCollider struct {
	X, Z, R, Top float32
}

// --------------------------------------------------------------------
// Lumea
// --------------------------------------------------------------------

type SurfaceKind int

const (
	SurfLand SurfaceKind = iota
	SurfWater
	SurfBridge
	SurfTunnel
	SurfMountain
	SurfRoof
	SurfPad // trambulina
)

type World struct {
	HeightFn  func(x, z float32) float32 // V 6.0: lumile de uscat (dealuri, munti pe margine); nil = insulele
	Lands     []LandShape
	Mountains []Mountain
	Bridges   []Bridge
	Boxes     []BoxCollider
	Circles   []CircleCollider
	Flatten   []Capsule
	Corridors []Corridor
	Lifts     []Lift
	Pads      []Trampoline
	Roads     []RoadSeg // pentru sunetul pasilor (asfalt / iarba)

	gridBoxes   map[int64][]int
	gridCircles map[int64][]int
}

type Capsule struct {
	A, B rl.Vector2
	R    float32
}

func NewWorldLayout() *World {
	w := &World{}
	for _, s := range Sectors[:4] {
		w.Lands = append(w.Lands, LandShape{Kind: 0, C: s.Center, HalfU: IslandHalf, HalfV: IslandHalf, Corner: IslandCorner})
	}
	w.Lands = append(w.Lands, LandShape{Kind: 1, C: rl.NewVector2(0, 0), R: CenterR})
	// insulele noi si insulitele cu lift
	w.Lands = append(w.Lands,
		LandShape{Kind: 0, C: rl.NewVector2(WestX, 0), HalfU: WestHalfU, HalfV: WestHalfV, Corner: 22},
		LandShape{Kind: 0, C: rl.NewVector2(EastX, 0), HalfU: EastHalfU, HalfV: EastHalfV, Corner: 24},
		LandShape{Kind: 1, C: rl.NewVector2(-VentX, 0), R: VentR},
		LandShape{Kind: 1, C: rl.NewVector2(VentX, 0), R: VentR},
	)

	neon := []rl.Color{LegoRed, LegoBlue, LegoGreen, LegoYellow}
	names := []string{"Tunelul Roșu", "Tunelul Albastru", "Tunelul Verde", "Tunelul Galben"}
	for i, s := range Sectors[:4] {
		d := rl.Vector2Normalize(rl.NewVector2(-s.Inner.X, -s.Inner.Y)) // de la centru spre insula
		c := rl.Vector2Scale(d, IsletD)
		m := Mountain{Index: i, C: c, Dir: d, Perp: rl.NewVector2(-d.Y, d.X), Neon: neon[i], Seed: float32(i)*17.3 + 3.1,
			Peak1: 16 + float32(i%2)*4, Peak2: 20 - float32(i%2)*3, Name: names[i]}
		w.Mountains = append(w.Mountains, m)
		mid := (MountAuOut - MountAuIn) / 2
		w.Lands = append(w.Lands, LandShape{Kind: 2, C: rl.Vector2Add(c, rl.Vector2Scale(d, mid)), HalfU: (MountAuOut + MountAuIn) / 2, HalfV: 28, Corner: 12, Dir: d})
		a := rl.Vector2Scale(d, 128)
		b := rl.Vector2Scale(d, 178)
		w.Flatten = append(w.Flatten, Capsule{A: a, B: b, R: 8})
	}

	gx := IslandHalf - GrassInset // 59: de la centrul insulei la marginea ierbii
	w.Bridges = []Bridge{
		NewBridge(BridgeSuspension, "Podul Suspendat", rl.NewVector2(SecNW_X+gx, SecNW_Z), rl.NewVector2(SecNE_X-gx, SecNE_Z), 13, 7, LegoRed),
		NewBridge(BridgeStoneArch, "Podul de Piatră", rl.NewVector2(SecSW_X+gx, SecSW_Z), rl.NewVector2(SecSE_X-gx, SecSE_Z), 12, 9, rl.NewColor(214, 184, 140, 255)),
		NewBridge(BridgeCableStayed, "Podul cu Hobane", rl.NewVector2(SecNW_X, SecNW_Z+gx), rl.NewVector2(SecSW_X, SecSW_Z-gx), 13, 6, LegoWhite),
		NewBridge(BridgeSteelArch, "Podul cu Arc", rl.NewVector2(SecNE_X, SecNE_Z+gx), rl.NewVector2(SecSE_X, SecSE_Z-gx), 13, 6, LegoYellow),
	}
	for i := range w.Mountains {
		m := &w.Mountains[i]
		a := rl.Vector2Scale(m.Dir, CenterR-GrassInset)
		b := rl.Vector2Scale(m.Dir, IsletD-28)
		w.Bridges = append(w.Bridges, NewBridge(BridgeBeam, "Pod scurt", a, b, 13, 0.8, rl.NewColor(200, 200, 205, 255)))
	}
	// V 5.0: podurile noi spre Insula Plajei (din SV) si spre Parcul de Iarna (din NE)
	w.Bridges = append(w.Bridges,
		NewBridge(BridgeCableStayed, "Podul Plajei", WestBridgeA, WestBridgeB, 13, 6, LegoCyan),
		NewBridge(BridgeSuspension, "Podul de Iarnă", EastBridgeA, EastBridgeB, 13, 7, LegoBlue),
	)
	w.buildCorridors()
	return w
}

// NaturalLand: uscatul / fundul marii fara sapaturi (pentru apa si culori).
func (w *World) NaturalLand(x, z float32) float32 {
	if w.HeightFn != nil {
		return w.HeightFn(x, z)
	}
	h := SeabedY
	for i := range w.Lands {
		if p := landProfile(w.Lands[i].SD(x, z)); p > h {
			h = p
		}
	}
	for _, c := range w.Flatten {
		if h < 0 && distToSeg(x, z, c.A, c.B) < c.R {
			h = 0
		}
	}
	return h
}

// LandHeight: solul pe care se calca (cu santurile deschise ale tunelurilor).
func (w *World) LandHeight(x, z float32) float32 {
	h := w.NaturalLand(x, z)
	if cy, ok := w.openCut(x, z); ok && cy < h {
		return cy
	}
	return h
}

// LandSD: distanta (cu semn) la cel mai apropiat mal.
func (w *World) LandSD(x, z float32) float32 {
	best := float32(1e9)
	for i := range w.Lands {
		if s := w.Lands[i].SD(x, z); s < best {
			best = s
		}
	}
	return best
}

func distToSeg(px, pz float32, a, b rl.Vector2) float32 {
	dx := b.X - a.X
	dz := b.Y - a.Y
	l2 := dx*dx + dz*dz
	t := float32(0)
	if l2 > 0 {
		t = clampf(((px-a.X)*dx+(pz-a.Y)*dz)/l2, 0, 1)
	}
	qx := a.X + dx*t - px
	qz := a.Y + dz*t - pz
	return sqrtf(qx*qx + qz*qz)
}

// mountainAt: cel mai inalt munte in (x,z) si daca punctul e in golul unui tunel.
func (w *World) mountainAt(x, z, y float32) (h float32, inVoid bool, inTrench bool, trenchWall float32, mi int) {
	mi = -1
	for i := range w.Mountains {
		m := &w.Mountains[i]
		u, v := m.Local(x, z)
		if u < -MountAuIn-1 || u > MountAuOut+1 || absf(v) > 30 {
			continue
		}
		mi = i
		if m.InVoid(u, v, y) {
			inVoid = true
		}
		if m.InTrench(u, v) {
			inTrench = true
			trenchWall = m.Raw(u, TrenchHalfW)
			if t2 := m.Raw(u, -TrenchHalfW); t2 > trenchWall {
				trenchWall = t2
			}
			if absf(v) > TrenchHalfW-0.45 {
				h = trenchWall
				return
			}
		}
		if e := m.Effective(u, v); e > h {
			h = e
		}
	}
	return
}

// InTunnel: in ce tunel e entitatea (0..3 = tunelurile de munte, 10+ = tunelurile subterane), -1 = afara.
func (w *World) InTunnel(x, z, y float32) int {
	for i := range w.Mountains {
		m := &w.Mountains[i]
		u, v := m.Local(x, z)
		if m.InTube(u, v, y) {
			return i
		}
	}
	if ci := w.corridorCovered(x, z, y); ci >= 0 {
		return 10 + ci
	}
	return -1
}

// TunnelName: numele tunelului dat de InTunnel.
func (w *World) TunnelName(t int) string {
	if t >= 0 && t < len(w.Mountains) {
		return w.Mountains[t].Name
	}
	if t >= 10 && t-10 < len(w.Corridors) {
		return w.Corridors[t-10].Name
	}
	return ""
}

// BridgeAt: indexul podului si inaltimea soselei, daca (x,z) e pe pod.
func (w *World) BridgeAt(x, z float32) (int, float32) {
	for i := range w.Bridges {
		b := &w.Bridges[i]
		t, lat := b.Local(x, z)
		if t < -0.01 || t > 1.01 || absf(lat) > b.Width/2-0.35 {
			continue
		}
		y := b.DeckY(t)
		if absf(lat) > 4.6 {
			y += 0.2 // trotuar
		}
		return i, y
	}
	return -1, 0
}

const StepUp = float32(0.65)

// GroundAt: suprafata pe care sta o entitate aflata la inaltimea y.
func (w *World) GroundAt(x, z, y float32) (float32, SurfaceKind) {
	best := float32(-1000)
	kind := SurfLand
	consider := func(h float32, k SurfaceKind) {
		if h <= y+StepUp && h > best {
			best = h
			kind = k
		}
	}
	land := w.NaturalLand(x, z)
	if cy, ok := w.openCut(x, z); ok && cy < land {
		consider(cy, SurfTunnel)
	} else {
		consider(land, SurfLand)
	}
	mh, inVoid, _, _, _ := w.mountainAt(x, z, y)
	if mh > 0.05 && !inVoid {
		consider(land+mh-0.1, SurfMountain)
	}
	if bi, by := w.BridgeAt(x, z); bi >= 0 {
		consider(by, SurfBridge)
	}
	for ci := range w.Corridors {
		if f, ok := w.Corridors[ci].FloorAt(x, z); ok {
			consider(f, SurfTunnel)
		}
	}
	for _, bi := range w.gridBoxes[gridKey(x, z)] {
		b := &w.Boxes[bi]
		if x > b.MinX && x < b.MaxX && z > b.MinZ && z < b.MaxZ {
			if b.Roof != RoofNone {
				consider(b.SurfaceY(x, z), SurfRoof)
			} else if b.Top < 1.0 {
				consider(b.Top, SurfLand)
			}
		}
	}
	// varful pilelor de pod, al soclurilor si al altor stalpi grosi: se poate sta pe el
	if key := gridKey(x, z); true {
		for _, k := range [...]int64{key, key + 1, key - 1, key + 1<<20, key - 1<<20, key + 1 + 1<<20, key - 1 + 1<<20, key + 1 - 1<<20, key - 1 - 1<<20} {
			for _, ci := range w.gridCircles[k] {
				c := &w.Circles[ci]
				if c.R < 0.95 {
					continue
				}
				dx, dz := x-c.X, z-c.Z
				if dx*dx+dz*dz < c.R*c.R {
					consider(c.Top, SurfRoof)
				}
			}
		}
	}
	for i := range w.Pads {
		p := &w.Pads[i]
		dx, dz := x-p.Pos.X, z-p.Pos.Z
		if dx*dx+dz*dz < p.R*p.R {
			consider(p.MatY(), SurfPad)
		}
	}
	if best < -999 {
		best = land
	}
	if kind == SurfLand && best < WaterY-0.3 {
		kind = SurfWater
	}
	if kind == SurfLand && w.InTunnel(x, z, y) >= 0 {
		kind = SurfTunnel
	}
	if kind == SurfWater && w.inCorridorBand(x, z, y) {
		// langa peretii Marelui Tunel nu e apa (tunelul e uscat, apa e deasupra)
		kind = SurfTunnel
	}
	return best, kind
}

// Solid: exista un obstacol in punctul (x,y,z), pentru o entitate de raza r.
func (w *World) Solid(x, z, y, r float32) bool {
	if w.corridorVolume(x, z, y) >= 0 {
		// in tunelul subteran conteaza doar forma tunelului (podea si pereti prin GroundAt)
		return false
	}
	if w.underCorridorRoof(x, z, y) {
		return true
	}
	key := gridKey(x, z)
	for _, k := range [...]int64{key, key + 1, key - 1, key + 1<<20, key - 1<<20, key + 1 + 1<<20, key - 1 + 1<<20, key + 1 - 1<<20, key - 1 - 1<<20} {
		for _, bi := range w.gridBoxes[k] {
			b := &w.Boxes[bi]
			if x > b.MinX-r && x < b.MaxX+r && z > b.MinZ-r && z < b.MaxZ+r {
				top := b.Top
				if b.Roof != RoofNone {
					top = b.SurfaceY(clampf(x, b.MinX, b.MaxX), clampf(z, b.MinZ, b.MaxZ))
				}
				if y < top {
					return true
				}
			}
		}
		for _, ci := range w.gridCircles[k] {
			c := &w.Circles[ci]
			dx := x - c.X
			dz := z - c.Z
			rr := c.R + r
			if y < c.Top && dx*dx+dz*dz < rr*rr {
				return true
			}
		}
	}
	for i := range w.Pads {
		p := &w.Pads[i]
		dx, dz := x-p.Pos.X, z-p.Pos.Z
		rr := p.R + 0.35 + r
		if y < p.MatY()-0.05 && dx*dx+dz*dz < rr*rr {
			return true
		}
	}
	mh, inVoid, _, _, _ := w.mountainAt(x, z, y)
	if mh > 0.3 && !inVoid && y < w.NaturalLand(x, z)+mh-0.25 {
		return true
	}
	return false
}

const gridCell = 16.0

func gridKey(x, z float32) int64 {
	gx := int64(math.Floor(float64(x/gridCell))) + 512
	gz := int64(math.Floor(float64(z/gridCell))) + 512
	return gx + gz<<20
}

func (w *World) RebuildGrid() {
	w.gridBoxes = map[int64][]int{}
	w.gridCircles = map[int64][]int{}
	for i, b := range w.Boxes {
		for x := b.MinX; x <= b.MaxX+gridCell; x += gridCell {
			for z := b.MinZ; z <= b.MaxZ+gridCell; z += gridCell {
				k := gridKey(minf(x, b.MaxX), minf(z, b.MaxZ))
				if !containsInt(w.gridBoxes[k], i) {
					w.gridBoxes[k] = append(w.gridBoxes[k], i)
				}
			}
		}
	}
	for i, c := range w.Circles {
		k := gridKey(c.X, c.Z)
		w.gridCircles[k] = append(w.gridCircles[k], i)
	}
}

func (w *World) AddCircle(x, z, r, top float32) int {
	w.Circles = append(w.Circles, CircleCollider{X: x, Z: z, R: r, Top: top})
	idx := len(w.Circles) - 1
	if w.gridCircles != nil {
		k := gridKey(x, z)
		w.gridCircles[k] = append(w.gridCircles[k], idx)
	}
	return idx
}

func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func containsInt(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// CanStand: pozitia (x,z) e accesibila (pe uscat) pentru cineva aflat la inaltimea y.
func (w *World) CanStand(x, z, y, r float32) bool {
	g, kind := w.GroundAt(x, z, y)
	if kind == SurfWater {
		return false
	}
	if g > y+StepUp {
		return false
	}
	if w.Solid(x, z, maxf(y, g)+StepUp, r) {
		return false
	}
	return true
}

// CanStandPlayer: ca CanStand, dar Aventurierul poate inota oriunde intre insule, pana la marginea lumii.
// (cx, cz) = unde e acum (daca e deja dincolo de margine, poate doar sa se intoarca).
func (w *World) CanStandPlayer(x, z, y, r, cx, cz float32) bool {
	g, kind := w.GroundAt(x, z, y)
	if g > y+StepUp {
		return false
	}
	if kind == SurfWater {
		if (absf(x) > MapEdgeX && absf(x) > absf(cx)) || (absf(z) > MapEdgeZ && absf(z) > absf(cz)) {
			return false
		}
	}
	if w.Solid(x, z, maxf(y, g)+StepUp, r) {
		return false
	}
	return true
}

// Move muta o entitate cu alunecare de-a lungul obstacolelor. Intoarce true daca s-a miscat.
func (w *World) Move(pos *rl.Vector3, dx, dz, r float32) bool {
	return w.move(pos, dx, dz, r, func(x, z, y float32) bool { return w.CanStand(x, z, y, r) })
}

// MovePlayer: miscarea aventurierului (poate inota aproape de mal).
func (w *World) MovePlayer(pos *rl.Vector3, dx, dz, r, testY float32) bool {
	cx, cz := pos.X, pos.Z
	return w.move(pos, dx, dz, r, func(x, z, y float32) bool { return w.CanStandPlayer(x, z, testY, r, cx, cz) })
}

func (w *World) move(pos *rl.Vector3, dx, dz, r float32, ok func(x, z, y float32) bool) bool {
	nx := pos.X + dx
	nz := pos.Z + dz
	// daca a ajuns cumva inauntrul unui obstacol (ex. o constructie care a crescut, o pila
	// de pod in mare), il lasam sa iasa in orice directie, si spre apa (inoata mai departe)
	if w.Solid(pos.X, pos.Z, pos.Y+StepUp, r) {
		g, _ := w.GroundAt(nx, nz, pos.Y+2)
		pos.X, pos.Z = nx, nz
		if g > pos.Y && g < pos.Y+2.5 {
			pos.Y = g
		}
		return true
	}
	if ok(nx, nz, pos.Y) {
		pos.X, pos.Z = nx, nz
		return true
	}
	if dx != 0 && ok(pos.X+dx, pos.Z, pos.Y) {
		pos.X += dx
		return true
	}
	if dz != 0 && ok(pos.X, pos.Z+dz, pos.Y) {
		pos.Z += dz
		return true
	}
	return false
}

// SectorAt: in ce zona e punctul.
func SectorAt(x, z float32) int {
	if curRealm > 0 {
		return realmSectorAt(x, z)
	}
	if x*x+z*z < (CenterR+8)*(CenterR+8) {
		return 4
	}
	best, bd := -1, float32(1e9)
	for i, s := range Sectors {
		if i == 4 {
			continue
		}
		dx := x - s.Center.X
		dz := z - s.Center.Y
		if d := dx*dx + dz*dz; d < bd {
			bd, best = d, i
		}
	}
	lim := float32(90)
	if best >= SecWest {
		lim = 100
	}
	if bd > lim*lim {
		return -1
	}
	return best
}

// CameraClear: verifica daca segmentul tinta->camera trece prin ceva solid si scurteaza distanta.
func (w *World) CameraClear(target, desired rl.Vector3) rl.Vector3 {
	d := rl.Vector3Subtract(desired, target)
	L := rl.Vector3Length(d)
	if L < 0.01 {
		return desired
	}
	dir := rl.Vector3Scale(d, 1/L)
	step := float32(0.3)
	last := target
	inside := w.corridorVolume(target.X, target.Z, target.Y) >= 0 || w.InTunnel(target.X, target.Z, target.Y) >= 0
	for s := step; s <= L; s += step {
		p := rl.Vector3Add(target, rl.Vector3Scale(dir, s))
		g, _ := w.GroundAt(p.X, p.Z, p.Y)
		if p.Y < g+0.35 || w.Solid(p.X, p.Z, p.Y, 0.2) || w.tunnelCeilingHit(p) {
			return last
		}
		// sub apa camera nu are ce cauta (afara de tuneluri)
		if !inside && p.Y < WaterY+0.45 && w.NaturalLand(p.X, p.Z) < WaterY {
			return last
		}
		last = p
	}
	return desired
}

func (w *World) tunnelCeilingHit(p rl.Vector3) bool {
	for i := range w.Mountains {
		m := &w.Mountains[i]
		u, v := m.Local(p.X, p.Z)
		if u >= -PortalIn-1 && u <= PortalOut+1 && absf(v) < TubeHalfW {
			ceil := TubeWallH + sqrtf(maxf(TubeHalfW*TubeHalfW-v*v, 0)) - 0.4
			if p.Y > ceil && p.Y < 11 {
				return true
			}
		}
	}
	for ci := range w.Corridors {
		c := &w.Corridors[ci]
		s, lat := c.Local(p.X, p.Z)
		if s < 0 || s > c.Len || absf(lat) > c.HalfW || c.IsOpen(s) {
			continue
		}
		if p.Y > c.CeilY(s, lat)-0.35 && p.Y < w.NaturalLand(p.X, p.Z) {
			return true
		}
	}
	return false
}

// V 5.0: capetele podurilor noi (bratul de vest al insulei SV -> coltul SE al
// Insulei Plajei; bratul de est al insulei NE -> coltul NV al Parcului).
var (
	WestBridgeA = rl.NewVector2(SecSW_X-(IslandHalf-GrassInset), SecSW_Z)
	WestBridgeB = rl.NewVector2(-309, 58)
	EastBridgeA = rl.NewVector2(SecNE_X+(IslandHalf-GrassInset), SecNE_Z)
	EastBridgeB = rl.NewVector2(311, -71)
)

const (
	BridgeWestIdx = 8
	BridgeEastIdx = 9
)
