package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Paleta lumii (stil jucarie, culori calde, fara negru/alb pur)
var (
	ColGrass     = rl.NewColor(108, 190, 78, 255)
	ColGrass2    = rl.NewColor(98, 178, 72, 255)
	ColSand      = rl.NewColor(240, 218, 166, 255)
	ColWetSand   = rl.NewColor(206, 184, 136, 255)
	ColSeabed    = rl.NewColor(160, 168, 140, 255)
	ColAsphalt   = rl.NewColor(72, 76, 86, 255)
	ColSidewalk  = rl.NewColor(205, 205, 200, 255)
	ColCurb      = rl.NewColor(170, 170, 168, 255)
	ColStripe    = rl.NewColor(246, 246, 238, 255)
	ColYellowLn  = rl.NewColor(250, 205, 60, 255)
	ColPaving    = rl.NewColor(226, 214, 190, 255)
	ColPaving2   = rl.NewColor(204, 186, 160, 255)
	ColTrunk     = rl.NewColor(128, 86, 54, 255)
	ColLeaf      = rl.NewColor(76, 170, 64, 255)
	ColLeaf2     = rl.NewColor(110, 196, 80, 255)
	ColPine      = rl.NewColor(46, 128, 72, 255)
	ColLampPole  = rl.NewColor(60, 66, 78, 255)
	ColLampLight = rl.NewColor(255, 236, 160, 255)
	ColGlass     = rl.NewColor(120, 180, 225, 255)
	ColGlassDark = rl.NewColor(70, 120, 170, 255)
	ColFrame     = rl.NewColor(240, 240, 235, 255)
)

type RoadSeg struct {
	A, B  rl.Vector2
	HalfW float32
}

// CityBuilder strange geometria statica si obstacolele.
type CityBuilder struct {
	W       *World
	Rng     *rand.Rand
	Roads   []RoadSeg
	Reserve []CircleCollider // zone rezervate (santiere blocuri, repere) - fara cladiri/copaci
	Chunks  []Chunk
	Trees   []rl.Vector3
}

func (cb *CityBuilder) addChunk(mb *MeshBuilder, shadows bool) {
	if mb.VertexCount() == 0 {
		return
	}
	// sfera de incadrare
	minX, minY, minZ := float32(1e9), float32(1e9), float32(1e9)
	maxX, maxY, maxZ := float32(-1e9), float32(-1e9), float32(-1e9)
	for i := 0; i < len(mb.V); i += 3 {
		x, y, z := mb.V[i], mb.V[i+1], mb.V[i+2]
		minX, minY, minZ = minf(minX, x), minf(minY, y), minf(minZ, z)
		maxX, maxY, maxZ = maxf(maxX, x), maxf(maxY, y), maxf(maxZ, z)
	}
	c := v3((minX+maxX)/2, (minY+maxY)/2, (minZ+maxZ)/2)
	r := rl.Vector3Distance(c, v3(maxX, maxY, maxZ))
	cb.Chunks = append(cb.Chunks, Chunk{Mesh: mb.Build(), Center: c, Radius: r, Shadows: shadows})
}

// ---------------------------------------------------------------
// TEREN
// ---------------------------------------------------------------

func (cb *CityBuilder) terrainColor(x, z, h float32) rl.Color {
	sd := cb.W.LandSD(x, z)
	var c rl.Color
	switch {
	case h < WaterY-0.2:
		c = mixColor(ColWetSand, ColSeabed, smoothstep(WaterY-0.2, -4, h))
	case sd > -GrassInset+0.6:
		c = mixColor(ColSand, ColWetSand, smoothstep(-1.5, 0.8, sd))
	default:
		// iarba cu "placi blocuri" de 8 m in doua nuante + variatie fina
		px := int(math.Floor(float64(x / 8)))
		pz := int(math.Floor(float64(z / 8)))
		c = ColGrass
		if (px+pz)&1 == 0 {
			c = ColGrass2
		}
		n := fbm(x*0.05, z*0.05, 3)
		c = shade(c, 0.92+0.16*n)
		c = mixColor(c, ColSand, smoothstep(-GrassInset-0.8, -GrassInset+0.6, sd))
		// plaje largi pe Insula Plajei si pe insulitele din mare
		if x < -290 {
			c = mixColor(c, ColSand, smoothstep(-290, -315, x))
		}
		if absf(absf(x)-VentX) < 16 && absf(z) < 16 {
			c = ColSand
		}
	}
	if x > 300 && absf(z) < 85 && sd < -GrassInset && h > WaterY {
		c = mixColor(c, rl.NewColor(236, 214, 170, 255), smoothstep(300, 325, x)) // V 6.0: dale calde de parc acvatic (nu zapada)
	}
	return c
}

// terrainHole: celula de teren peste santul deschis al unui tunel (acolo e golul).
func (w *World) terrainHole(x, z float32) bool {
	for i := range w.Corridors {
		c := &w.Corridors[i]
		s, lat := c.Local(x, z)
		if s < 0 || s > c.Len || absf(lat) > c.HalfW+1.5 {
			continue
		}
		if c.IsOpen(s) || c.IsOpen(clampf(s-1.5, 0, c.Len)) || c.IsOpen(clampf(s+1.5, 0, c.Len)) {
			return true
		}
	}
	return false
}

func (cb *CityBuilder) buildTerrain() {
	const step = 2.0
	const extX = 432.0
	const extZ = 252.0
	nx := int(2 * extX / step)
	nz := int(2 * extZ / step)
	hs := make([]float32, (nx+1)*(nz+1))
	for j := 0; j <= nz; j++ {
		for i := 0; i <= nx; i++ {
			x := -extX + float32(i)*step
			z := -extZ + float32(j)*step
			hs[j*(nx+1)+i] = cb.W.NaturalLand(x, z)
		}
	}
	// bucati de 64 m pentru eliminare in afara ecranului
	parts := map[int]*MeshBuilder{}
	normal := func(i, j int) rl.Vector3 {
		i0, i1 := maxInt(i-1, 0), minInt(i+1, nx)
		j0, j1 := maxInt(j-1, 0), minInt(j+1, nz)
		dx := hs[j*(nx+1)+i1] - hs[j*(nx+1)+i0]
		dz := hs[j1*(nx+1)+i] - hs[j0*(nx+1)+i]
		return rl.Vector3Normalize(v3(-dx/(float32(i1-i0)*step), 1, -dz/(float32(j1-j0)*step)))
	}
	// Fundul marii plat: DOAR unde chiar e mare deschisa (nu un plan unic pe sub toata
	// harta: acela taia Marele Tunel, scarile si lifturile la -7 m si parea apa).
	seaParts := map[int]*MeshBuilder{}
	seaY := SeabedY - 0.02
	seaQuad := func(x0, z0, x1, z1 float32) {
		key := int((x0+extX)/64) + 100*int((z0+extZ)/64)
		mb := seaParts[key]
		if mb == nil {
			mb = NewMB()
			mb.AO = 0.8
			seaParts[key] = mb
		}
		mb.Quad(v3(x0, seaY, z0), v3(x1, seaY, z0), v3(x1, seaY, z1), v3(x0, seaY, z1), v3(0, 1, 0), ColSeabed)
	}
	for j := 0; j < nz; j++ {
		run := -1
		flush := func(i int) {
			if run >= 0 {
				z0 := -extZ + float32(j)*step
				seaQuad(-extX+float32(run)*step, z0, -extX+float32(i)*step, z0+step)
				run = -1
			}
		}
		for i := 0; i < nx; i++ {
			h00 := hs[j*(nx+1)+i]
			h10 := hs[j*(nx+1)+i+1]
			h01 := hs[(j+1)*(nx+1)+i]
			h11 := hs[(j+1)*(nx+1)+i+1]
			if h00 <= SeabedY+0.05 && h10 <= SeabedY+0.05 && h01 <= SeabedY+0.05 && h11 <= SeabedY+0.05 {
				if run < 0 {
					run = i
				}
				// o bucata de fundul marii nu trece de marginea unei bucati de 64 m
				if (i+1)%32 == 0 {
					flush(i + 1)
				}
				continue
			}
			flush(i)
			x0 := -extX + float32(i)*step
			z0 := -extZ + float32(j)*step
			if cb.W.terrainHole(x0+step/2, z0+step/2) {
				continue
			}
			p00 := v3(x0, h00, z0)
			p10 := v3(x0+step, h10, z0)
			p01 := v3(x0, h01, z0+step)
			p11 := v3(x0+step, h11, z0+step)
			c00 := cb.terrainColor(p00.X, p00.Z, h00)
			c10 := cb.terrainColor(p10.X, p10.Z, h10)
			c01 := cb.terrainColor(p01.X, p01.Z, h01)
			c11 := cb.terrainColor(p11.X, p11.Z, h11)
			key := int((x0+extX)/64) + 100*int((z0+extZ)/64)
			mb := parts[key]
			if mb == nil {
				mb = NewMB()
				parts[key] = mb
			}
			n00, n10, n01, n11 := normal(i, j), normal(i+1, j), normal(i, j+1), normal(i+1, j+1)
			mb.TriS(p00, n00, p01, n01, p11, n11, c00, c01, c11)
			mb.TriS(p00, n00, p11, n11, p10, n10, c00, c11, c10)
		}
		flush(nx)
	}
	for _, p := range parts {
		cb.addChunk(p, true)
	}
	for _, p := range seaParts {
		cb.addChunk(p, false)
	}
	// fundul marii, departe: doar in afara terenului (o rama in jurul hartii)
	sea := NewMB()
	sea.AO = 0.8
	e := float32(1400)
	ex, ez := float32(extX), float32(extZ)
	rect := func(x0, z0, x1, z1 float32) {
		sea.Quad(v3(x0, seaY, z0), v3(x1, seaY, z0), v3(x1, seaY, z1), v3(x0, seaY, z1), v3(0, 1, 0), ColSeabed)
	}
	rect(-e, -e, e, -ez)
	rect(-e, ez, e, e)
	rect(-e, -ez, -ex, ez)
	rect(ex, -ez, e, ez)
	cb.addChunk(sea, false)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Apa: grila cu adancimea "coapta" in varfuri (pentru culoare si spuma).
func (cb *CityBuilder) buildWater() rl.Mesh {
	mb := NewMB()
	const step = 4.0
	const extX = 464.0
	const extZ = 300.0
	nx := int(2 * extX / step)
	nz := int(2 * extZ / step)
	depthAt := func(x, z float32) float32 {
		h := cb.W.NaturalLand(x, z)
		mh, _, _, _, _ := cb.W.mountainAt(x, z, 100)
		h += mh
		return WaterY - h
	}
	ds := make([]float32, (nx+1)*(nz+1))
	for j := 0; j <= nz; j++ {
		for i := 0; i <= nx; i++ {
			ds[j*(nx+1)+i] = depthAt(-extX+float32(i)*step, -extZ+float32(j)*step)
		}
	}
	wc := rl.NewColor(255, 255, 255, 255)
	add := func(x, z, d float32) {
		mb.V = append(mb.V, x, WaterY, z)
		mb.N = append(mb.N, 0, 1, 0)
		mb.C = append(mb.C, wc.R, wc.G, wc.B, wc.A)
		mb.T = append(mb.T, d, 0)
	}
	for j := 0; j < nz; j++ {
		for i := 0; i < nx; i++ {
			d00 := ds[j*(nx+1)+i]
			d10 := ds[j*(nx+1)+i+1]
			d01 := ds[(j+1)*(nx+1)+i]
			d11 := ds[(j+1)*(nx+1)+i+1]
			if d00 < -0.8 && d10 < -0.8 && d01 < -0.8 && d11 < -0.8 {
				continue
			}
			x0 := -extX + float32(i)*step
			z0 := -extZ + float32(j)*step
			// ordine CCW vazuta de sus
			add(x0, z0, d00)
			add(x0, z0+step, d01)
			add(x0+step, z0+step, d11)
			add(x0, z0, d00)
			add(x0+step, z0+step, d11)
			add(x0+step, z0, d10)
		}
	}
	// oceanul de dincolo de harta (inel mare)
	big := float32(1400)
	deep := float32(8)
	ring := [][4]float32{
		{-big, -big, big, -extZ},
		{-big, extZ, big, big},
		{-big, -extZ, -extX, extZ},
		{extX, -extZ, big, extZ},
	}
	for _, r := range ring {
		add(r[0], r[1], deep)
		add(r[0], r[3], deep)
		add(r[2], r[3], deep)
		add(r[0], r[1], deep)
		add(r[2], r[3], deep)
		add(r[2], r[1], deep)
	}
	return mb.Build()
}

// ---------------------------------------------------------------
// DRUMURI
// ---------------------------------------------------------------

// roadArm: sosea dreapta cu trotuare, borduri si marcaje.
func (cb *CityBuilder) roadArm(mb *MeshBuilder, a, b rl.Vector2, halfW float32, sidewalks bool, centerLine bool) {
	cb.Roads = append(cb.Roads, RoadSeg{A: a, B: b, HalfW: halfW + 2})
	d := rl.Vector2Subtract(b, a)
	L := rl.Vector2Length(d)
	if L < 0.1 {
		return
	}
	dir := rl.Vector2Scale(d, 1/L)
	per := rl.NewVector2(-dir.Y, dir.X)
	pt := func(t, lat, y float32) rl.Vector3 {
		return v3(a.X+dir.X*t+per.X*lat, y, a.Y+dir.Y*t+per.Y*lat)
	}
	up := v3(0, 1, 0)
	y := float32(0.05)
	mb.Quad(pt(0, -halfW, y), pt(L, -halfW, y), pt(L, halfW, y), pt(0, halfW, y), up, ColAsphalt)
	// marcaje
	if centerLine {
		for t := float32(3); t < L-3; t += 6 {
			mb.Quad(pt(t, -0.12, y+0.01), pt(t+3, -0.12, y+0.01), pt(t+3, 0.12, y+0.01), pt(t, 0.12, y+0.01), up, ColStripe)
		}
	}
	mb.Quad(pt(0, halfW-0.45, y+0.01), pt(L, halfW-0.45, y+0.01), pt(L, halfW-0.3, y+0.01), pt(0, halfW-0.3, y+0.01), up, ColStripe)
	mb.Quad(pt(0, -halfW+0.3, y+0.01), pt(L, -halfW+0.3, y+0.01), pt(L, -halfW+0.45, y+0.01), pt(0, -halfW+0.45, y+0.01), up, ColStripe)
	if sidewalks {
		sw := float32(2.2)
		top := float32(0.2)
		for _, s := range []float32{-1, 1} {
			in := halfW * s
			out := (halfW + sw) * s
			mb.Quad(pt(0, in, top), pt(L, in, top), pt(L, out, top), pt(0, out, top), up, ColSidewalk)
			// bordura (fata verticala spre sosea)
			mb.Quad(pt(0, in, 0.0), pt(L, in, 0.0), pt(L, in, top), pt(0, in, top), v3(-per.X*s, 0, -per.Y*s), ColCurb)
			mb.Quad(pt(0, out, 0.0), pt(L, out, 0.0), pt(L, out, top), pt(0, out, top), v3(per.X*s, 0, per.Y*s), ColCurb)
			// capetele
			mb.Quad(pt(0, in, 0), pt(0, out, 0), pt(0, out, top), pt(0, in, top), v3(-dir.X, 0, -dir.Y), ColCurb)
			mb.Quad(pt(L, in, 0), pt(L, out, 0), pt(L, out, top), pt(L, in, top), v3(dir.X, 0, dir.Y), ColCurb)
		}
	}
}

// trecere de pietoni (zebra)
func (cb *CityBuilder) zebra(mb *MeshBuilder, c rl.Vector2, dir rl.Vector2, halfW float32) {
	per := rl.NewVector2(-dir.Y, dir.X)
	for lat := -halfW + 0.8; lat < halfW-0.6; lat += 1.2 {
		p := func(t, l float32) rl.Vector3 {
			return v3(c.X+dir.X*t+per.X*(lat+l), 0.065, c.Y+dir.Y*t+per.Y*(lat+l))
		}
		mb.Quad(p(-1.6, 0), p(1.6, 0), p(1.6, 0.6), p(-1.6, 0.6), v3(0, 1, 0), ColStripe)
	}
}

func (cb *CityBuilder) roundabout(mb *MeshBuilder, c rl.Vector2, rOut, rIn float32, centerCol rl.Color) {
	up := v3(0, 1, 0)
	mb.Disc(v3(c.X, 0.048, c.Y), rIn, rOut, 40, ColAsphalt)
	// insula verde din mijloc cu bordura
	mb.Cylinder(v3(c.X, 0, c.Y), v3(c.X, 0.3, c.Y), rIn, rIn, 32, ColCurb, false)
	mb.Disc(v3(c.X, 0.3, c.Y), 0, rIn, 32, centerCol)
	_ = up
	// linie alba pe interior
	mb.Disc(v3(c.X, 0.058, c.Y), rIn+0.3, rIn+0.45, 40, ColStripe)
}

func (cb *CityBuilder) buildRoads() {
	mb := NewMB()
	for si, s := range Sectors[:4] {
		c := s.Center
		reach := IslandHalf - GrassInset
		ra := float32(12)
		cb.roundabout(mb, c, ra, 5.5, ColGrass)
		cb.Reserve = append(cb.Reserve, CircleCollider{X: c.X, Z: c.Y, R: 14})
		dirs := []rl.Vector2{{X: 1, Y: 0}, {X: -1, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: -1}}
		for _, d := range dirs {
			a := rl.NewVector2(c.X+d.X*ra, c.Y+d.Y*ra)
			b := rl.NewVector2(c.X+d.X*reach, c.Y+d.Y*reach)
			cb.roadArm(mb, a, b, 4.5, true, true)
			cb.zebra(mb, rl.NewVector2(c.X+d.X*(ra+4), c.Y+d.Y*(ra+4)), rl.NewVector2(d.Y, -d.X), 4.5)
		}
		// bulevardul diagonal spre munte si tunel
		in := rl.Vector2Normalize(s.Inner)
		a := rl.NewVector2(c.X+in.X*ra, c.Y+in.Y*ra)
		m := cb.W.Mountains[si]
		ex, ez := m.World(MountAuOut, 0)
		cb.roadArm(mb, a, rl.NewVector2(ex, ez), 4.5, true, true)
	}
	// insula centrala: inel + 4 spite spre poduri + piateta pavata
	cb.ringRoad(mb, 48, 4)
	for i := range cb.W.Mountains {
		d := cb.W.Mountains[i].Dir
		a := rl.Vector2Scale(d, 52)
		b := rl.Vector2Scale(d, CenterR-GrassInset)
		cb.roadArm(mb, a, b, 4.5, false, true)
	}
	cb.addChunk(mb, false)
}

func (cb *CityBuilder) ringRoad(mb *MeshBuilder, r, halfW float32) {
	segs := 72
	up := v3(0, 1, 0)
	for i := 0; i < segs; i++ {
		a0 := float32(i) / float32(segs) * 2 * math.Pi
		a1 := float32(i+1) / float32(segs) * 2 * math.Pi
		p := func(a, rr, y float32) rl.Vector3 { return v3(cosf(a)*rr, y, sinf(a)*rr) }
		mb.Quad(p(a0, r-halfW, 0.05), p(a1, r-halfW, 0.05), p(a1, r+halfW, 0.05), p(a0, r+halfW, 0.05), up, ColAsphalt)
		if i%2 == 0 {
			mb.Quad(p(a0, r-0.12, 0.06), p(a1, r-0.12, 0.06), p(a1, r+0.12, 0.06), p(a0, r+0.12, 0.06), up, ColStripe)
		}
		// bordura interioara + trotuar
		mb.Quad(p(a0, r-halfW-2, 0.2), p(a1, r-halfW-2, 0.2), p(a1, r-halfW, 0.2), p(a0, r-halfW, 0.2), up, ColSidewalk)
		mid := (a0 + a1) / 2
		mb.Quad(p(a0, r-halfW, 0), p(a1, r-halfW, 0), p(a1, r-halfW, 0.2), p(a0, r-halfW, 0.2), v3(cosf(mid), 0, sinf(mid)), ColCurb)
	}
	cb.Roads = append(cb.Roads, RoadSeg{A: rl.NewVector2(0, 0), B: rl.NewVector2(0, 0), HalfW: -1}) // inelul e tratat separat
}

// pe drum? (pentru plasarea cladirilor/copacilor)
func (cb *CityBuilder) nearRoad(x, z, margin float32) bool {
	for _, r := range cb.Roads {
		if r.HalfW < 0 {
			d := sqrtf(x*x + z*z)
			if absf(d-48) < 6+margin {
				return true
			}
			continue
		}
		if distToSeg(x, z, r.A, r.B) < r.HalfW+margin {
			return true
		}
	}
	return false
}

func (cb *CityBuilder) reserved(x, z, r float32) bool {
	for _, c := range cb.Reserve {
		dx := x - c.X
		dz := z - c.Z
		rr := c.R + r
		if dx*dx+dz*dz < rr*rr {
			return true
		}
	}
	for _, b := range cb.W.Boxes {
		if x > b.MinX-r && x < b.MaxX+r && z > b.MinZ-r && z < b.MaxZ+r {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------
// CLADIRI
// ---------------------------------------------------------------

type Theme struct {
	Walls     []rl.Color
	Roofs     []rl.Color
	MinFloors int
	MaxFloors int
	Pitched   float32 // probabilitatea acoperisului in doua ape
	Tower     bool
}

var themes = []Theme{
	{ // Dino Land
		Walls:     []rl.Color{{238, 214, 170, 255}, {214, 150, 100, 255}, {196, 206, 140, 255}, {240, 190, 120, 255}},
		Roofs:     []rl.Color{{176, 96, 52, 255}, {150, 110, 60, 255}, {200, 120, 50, 255}},
		MinFloors: 1, MaxFloors: 3, Pitched: 0.85,
	},
	{ // Cartoon City
		Walls:     []rl.Color{{210, 55, 65, 255}, {225, 86, 82, 255}, {54, 105, 191, 255}, {84, 139, 212, 255}},
		Roofs:     []rl.Color{{180, 40, 55, 255}, {42, 82, 170, 255}},
		MinFloors: 2, MaxFloors: 4, Pitched: 0.7,
	},
	{ // Turbo Speedway
		Walls:     []rl.Color{{180, 170, 195, 255}, {157, 149, 173, 255}, {197, 188, 207, 255}, {166, 160, 185, 255}},
		Roofs:     []rl.Color{{109, 91, 139, 255}, {127, 111, 151, 255}},
		MinFloors: 2, MaxFloors: 5, Pitched: 0.1,
	},
	{ // Mecha Arena
		Walls:     []rl.Color{{235, 244, 252, 255}, {210, 230, 250, 255}, {104, 164, 224, 255}, {245, 250, 255, 255}},
		Roofs:     []rl.Color{{55, 112, 190, 255}, {91, 155, 218, 255}},
		MinFloors: 4, MaxFloors: 9, Pitched: 0, Tower: true,
	},
}

func (cb *CityBuilder) building(mb *MeshBuilder, cx, cz, w, d float32, floors int, th Theme, doorDir rl.Vector2) {
	rng := cb.Rng
	wall := th.Walls[rng.Intn(len(th.Walls))]
	roof := th.Roofs[rng.Intn(len(th.Roofs))]
	fh := float32(3.2)
	h := float32(floors)*fh + 0.6
	pitched := rng.Float32() < th.Pitched
	// soclu
	mb.Box(v3(cx, 0.3, cz), v3(w+0.4, 0.6, d+0.4), shade(wall, 0.72))
	// corp
	mb.BeveledBox(v3(cx, 0.6+h/2-0.3, cz), v3(w, h-0.6, d), 0.22, wall)
	roofKind, roofH := RoofFlat, float32(0)
	if pitched {
		roofH = minf(w, d) * 0.42
		roofKind = RoofRidgeZ
		if w >= d {
			roofKind = RoofRidgeX
		}
	}
	cb.W.Boxes = append(cb.W.Boxes, BoxCollider{MinX: cx - w/2 - 0.2, MaxX: cx + w/2 + 0.2, MinZ: cz - d/2 - 0.2, MaxZ: cz + d/2 + 0.2, Top: h, Roof: roofKind, RoofH: roofH})

	// cornise intre etaje
	for f := 1; f < floors; f++ {
		y := 0.3 + float32(f)*fh
		mb.Box(v3(cx, y, cz), v3(w+0.25, 0.18, d+0.25), shade(wall, 0.85))
	}
	// ferestre pe toate fetele
	type facade struct {
		n    rl.Vector2
		span float32
		half float32
	}
	faces := []facade{{rl.NewVector2(1, 0), d, w / 2}, {rl.NewVector2(-1, 0), d, w / 2}, {rl.NewVector2(0, 1), w, d / 2}, {rl.NewVector2(0, -1), w, d / 2}}
	glass := ColGlass
	if th.Tower {
		glass = ColGlassDark
	}
	for _, f := range faces {
		tang := rl.NewVector2(-f.n.Y, f.n.X)
		count := int(f.span / 2.6)
		if count < 1 {
			count = 1
		}
		spacing := f.span / float32(count)
		isDoorFace := f.n.X == doorDir.X && f.n.Y == doorDir.Y
		for fl := 0; fl < floors; fl++ {
			baseY := 0.6 + float32(fl)*fh
			if th.Tower {
				// banda continua de sticla pe fiecare etaj
				o := f.half + 0.04
				p := func(t, y float32) rl.Vector3 {
					return v3(cx+f.n.X*o+tang.X*t, y, cz+f.n.Y*o+tang.Y*t)
				}
				hs := f.span/2 - 0.5
				if fl == 0 && isDoorFace {
					continue
				}
				mb.Quad(p(-hs, baseY+0.5), p(hs, baseY+0.5), p(hs, baseY+2.6), p(-hs, baseY+2.6), v3(f.n.X, 0, f.n.Y), glass)
				for t := -hs; t <= hs+0.01; t += spacing {
					mb.Box(v3(cx+f.n.X*(o+0.05)+tang.X*t, baseY+1.55, cz+f.n.Y*(o+0.05)+tang.Y*t), v3(0.18+absf(f.n.Y)*0.0, 2.1, 0.18), ColFrame)
				}
				continue
			}
			for k := 0; k < count; k++ {
				t := -f.span/2 + spacing*(float32(k)+0.5)
				if fl == 0 && isDoorFace && k == count/2 {
					// usa
					o := f.half + 0.03
					p := func(tt, y float32) rl.Vector3 {
						return v3(cx+f.n.X*o+tang.X*(t+tt), y, cz+f.n.Y*o+tang.Y*(t+tt))
					}
					mb.Quad(p(-0.8, 0.6), p(0.8, 0.6), p(0.8, 3.0), p(-0.8, 3.0), v3(f.n.X, 0, f.n.Y), shade(roof, 0.8))
					mb.Box(v3(cx+f.n.X*(o+0.5)+tang.X*t, 3.25, cz+f.n.Y*(o+0.5)+tang.Y*t), v3(2.4*absf(tang.X)+1.0*absf(f.n.X), 0.15, 2.4*absf(tang.Y)+1.0*absf(f.n.Y)), roof)
					continue
				}
				o := f.half + 0.03
				p := func(tt, y, off float32) rl.Vector3 {
					return v3(cx+f.n.X*(o+off)+tang.X*(t+tt), y, cz+f.n.Y*(o+off)+tang.Y*(t+tt))
				}
				nrm := v3(f.n.X, 0, f.n.Y)
				mb.Quad(p(-0.8, baseY+0.7, 0), p(0.8, baseY+0.7, 0), p(0.8, baseY+2.5, 0), p(-0.8, baseY+2.5, 0), nrm, ColFrame)
				mb.Quad(p(-0.62, baseY+0.86, 0.02), p(0.62, baseY+0.86, 0.02), p(0.62, baseY+2.34, 0.02), p(-0.62, baseY+2.34, 0.02), nrm, glass)
				// pervaz
				mb.Box(v3(cx+f.n.X*(o+0.12)+tang.X*t, baseY+0.66, cz+f.n.Y*(o+0.12)+tang.Y*t), v3(1.8*absf(tang.X)+0.25*absf(f.n.X), 0.1, 1.8*absf(tang.Y)+0.25*absf(f.n.Y)), shade(wall, 0.8))
			}
		}
	}
	top := 0.6 + h - 0.6
	if pitched {
		// acoperis in doua ape, pe latura lunga
		rh := minf(w, d) * 0.42
		ov := float32(0.5)
		if w >= d {
			x0, x1 := cx-w/2-ov, cx+w/2+ov
			z0, z1 := cz-d/2-ov, cz+d/2+ov
			ridge := top + rh
			mb.Quad(v3(x0, top, z0), v3(x1, top, z0), v3(x1, ridge, cz), v3(x0, ridge, cz), v3(0, 1, -1), roof)
			mb.Quad(v3(x0, top, z1), v3(x1, top, z1), v3(x1, ridge, cz), v3(x0, ridge, cz), v3(0, 1, 1), roof)
			mb.Tri(v3(cx-w/2, top, z0+ov), v3(cx-w/2, top, z1-ov), v3(cx-w/2, ridge-0.2, cz), v3(-1, 0, 0), wall)
			mb.Tri(v3(cx+w/2, top, z0+ov), v3(cx+w/2, top, z1-ov), v3(cx+w/2, ridge-0.2, cz), v3(1, 0, 0), wall)
			// streasina vazuta de jos
			mb.Quad(v3(x0, top-0.01, z0), v3(x1, top-0.01, z0), v3(x1, top-0.01, z1), v3(x0, top-0.01, z1), v3(0, -1, 0), shade(roof, 0.6))
		} else {
			x0, x1 := cx-w/2-ov, cx+w/2+ov
			z0, z1 := cz-d/2-ov, cz+d/2+ov
			ridge := top + rh
			mb.Quad(v3(x0, top, z0), v3(x0, top, z1), v3(cx, ridge, z1), v3(cx, ridge, z0), v3(-1, 1, 0), roof)
			mb.Quad(v3(x1, top, z0), v3(x1, top, z1), v3(cx, ridge, z1), v3(cx, ridge, z0), v3(1, 1, 0), roof)
			mb.Tri(v3(x0+ov, top, cz-d/2), v3(x1-ov, top, cz-d/2), v3(cx, ridge-0.2, cz-d/2), v3(0, 0, -1), wall)
			mb.Tri(v3(x0+ov, top, cz+d/2), v3(x1-ov, top, cz+d/2), v3(cx, ridge-0.2, cz+d/2), v3(0, 0, 1), wall)
			mb.Quad(v3(x0, top-0.01, z0), v3(x1, top-0.01, z0), v3(x1, top-0.01, z1), v3(x0, top-0.01, z1), v3(0, -1, 0), shade(roof, 0.6))
		}
		// cos
		if rng.Float32() < 0.6 {
			mb.Box(v3(cx+w*0.22, top+rh*0.7, cz+d*0.1), v3(0.9, rh*0.9, 0.9), shade(wall, 0.7))
		}
		return
	}
	// acoperis plat cu parapet si detalii
	pc := shade(roof, 1.0)
	mb.Box(v3(cx, top+0.35, cz-d/2+0.15), v3(w, 0.7, 0.3), pc)
	mb.Box(v3(cx, top+0.35, cz+d/2-0.15), v3(w, 0.7, 0.3), pc)
	mb.Box(v3(cx-w/2+0.15, top+0.35, cz), v3(0.3, 0.7, d), pc)
	mb.Box(v3(cx+w/2-0.15, top+0.35, cz), v3(0.3, 0.7, d), pc)
	mb.Box(v3(cx, top+0.02, cz), v3(w-0.6, 0.04, d-0.6), shade(roof, 0.8))
	if rng.Float32() < 0.7 {
		mb.BeveledBox(v3(cx-w*0.2, top+0.6, cz+d*0.15), v3(1.6, 1.2, 1.2), 0.1, LegoLightGrey)
	}
	if th.Tower {
		mb.Cylinder(v3(cx+w*0.25, top, cz-d*0.2), v3(cx+w*0.25, top+5, cz-d*0.2), 0.12, 0.05, 6, LegoLightGrey, false)
		mb.AO, mb.Emis = 1, 1
		mb.Sphere(v3(cx+w*0.25, top+5.1, cz-d*0.2), 0.22, LegoRed)
		mb.Emis = 0
	} else if rng.Float32() < 0.35 {
		// rezervor de apa pe acoperis
		x := cx + w*0.2
		z := cz - d*0.2
		for _, o := range [][2]float32{{-0.7, -0.7}, {0.7, -0.7}, {-0.7, 0.7}, {0.7, 0.7}} {
			mb.Cylinder(v3(x+o[0], top, z+o[1]), v3(x+o[0], top+1.6, z+o[1]), 0.08, 0.08, 5, ColTrunk, false)
		}
		mb.Cylinder(v3(x, top+1.6, z), v3(x, top+3.2, z), 1.1, 1.1, 12, rl.NewColor(150, 100, 70, 255), true)
		mb.Cylinder(v3(x, top+3.2, z), v3(x, top+3.9, z), 1.15, 0.1, 12, rl.NewColor(110, 80, 60, 255), false)
	}
}

func (cb *CityBuilder) buildBuildings() {
	for si, s := range Sectors[:4] {
		mb := NewMB()
		th := themes[si]
		in := s.Inner
		offs := []float32{16, 31, 46}
		for _, qx := range []float32{-1, 1} {
			for _, qz := range []float32{-1, 1} {
				outer := qx == -in.X && qz == -in.Y
				for _, ox := range offs {
					for _, oz := range offs {
						x := s.Center.X + qx*ox
						z := s.Center.Y + qz*oz
						// pe diagonala bulevardului
						if qx == in.X && qz == in.Y && absf(ox-oz) < 13 {
							continue
						}
						if outer && ox+oz < 70 {
							continue // locul reperului insulei
						}
						w := 9 + cb.Rng.Float32()*3
						d := 9 + cb.Rng.Float32()*3
						if cb.reserved(x, z, maxf(w, d)/2+1) || cb.nearRoad(x, z, maxf(w, d)/2+2.2) {
							continue
						}
						if cb.inMountain(x, z, maxf(w, d)/2+1) {
							continue
						}
						if cb.W.LandSD(x, z) > -GrassInset-maxf(w, d)/2-1 {
							continue
						}
						floors := th.MinFloors + cb.Rng.Intn(th.MaxFloors-th.MinFloors+1)
						// usa spre cel mai apropiat drum
						var door rl.Vector2
						if ox < oz {
							door = rl.NewVector2(-qx, 0)
						} else {
							door = rl.NewVector2(0, -qz)
						}
						cb.building(mb, x, z, w, d, floors, th, door)
					}
				}
			}
		}
		cb.addChunk(mb, true)
	}
}

// ---------------------------------------------------------------
// COPACI, FELINARE, BANCI
// ---------------------------------------------------------------

func (cb *CityBuilder) roundTree(mb *MeshBuilder, x, z, s float32, col rl.Color) {
	y := cb.W.LandHeight(x, z)
	mb.Cylinder(v3(x, y, z), v3(x, y+2.2*s, z), 0.32*s, 0.22*s, 7, ColTrunk, false)
	leaf := func(n rl.Vector3) rl.Color {
		// partea de jos a coroanei mai inchisa (ocluzie falsa)
		return shade(col, 0.78+0.3*(n.Y*0.5+0.5))
	}
	mb.EllipsoidF(v3(x, y+3.3*s, z), v3(1.9*s, 1.6*s, 1.9*s), 7, 10, leaf)
	mb.EllipsoidF(v3(x+0.9*s, y+2.8*s, z+0.4*s), v3(1.2*s, 1.0*s, 1.2*s), 6, 8, leaf)
	mb.EllipsoidF(v3(x-0.7*s, y+3.0*s, z-0.6*s), v3(1.2*s, 1.1*s, 1.2*s), 6, 8, leaf)
	cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 0.45 * s, Top: y + 6})
	cb.Trees = append(cb.Trees, v3(x, y, z))
}

func (cb *CityBuilder) pineTree(mb *MeshBuilder, x, y, z, s float32) {
	mb.Cylinder(v3(x, y-0.3, z), v3(x, y+1.4*s, z), 0.25*s, 0.2*s, 6, ColTrunk, false)
	col := ColPine
	for i := 0; i < 3; i++ {
		fy := y + (1.0+float32(i)*1.3)*s
		r := (1.7 - float32(i)*0.45) * s
		mb.CylinderC(v3(x, fy, z), v3(x, fy+2.0*s, z), r, 0.05, 9, shade(col, 0.8), shade(col, 1.15), true)
	}
}

func (cb *CityBuilder) palmTree(mb *MeshBuilder, x, z, s float32) {
	y := cb.W.LandHeight(x, z)
	prev := v3(x, y, z)
	lean := cb.Rng.Float32() * 2 * math.Pi
	for i := 1; i <= 6; i++ {
		t := float32(i) / 6
		p := v3(x+cosf(lean)*t*t*1.4*s, y+t*6*s, z+sinf(lean)*t*t*1.4*s)
		mb.Cylinder(prev, p, 0.3*s*(1.1-t*0.4), 0.3*s*(1.1-t*0.45), 7, shade(ColTrunk, 0.9+0.2*float32(i%2)), false)
		prev = p
	}
	for k := 0; k < 7; k++ {
		a := float32(k) / 7 * 2 * math.Pi
		dir := v3(cosf(a), 0, sinf(a))
		mid := rl.Vector3Add(prev, v3(dir.X*1.6*s, 0.5*s, dir.Z*1.6*s))
		end := rl.Vector3Add(prev, v3(dir.X*3.2*s, -0.9*s, dir.Z*3.2*s))
		mb.Cylinder(prev, mid, 0.35*s, 0.45*s, 4, ColLeaf, true)
		mb.Cylinder(mid, end, 0.45*s, 0.05*s, 4, ColLeaf2, true)
	}
	mb.Sphere(prev, 0.45*s, rl.NewColor(120, 80, 40, 255))
	cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 0.4 * s, Top: y + 7})
}

func (cb *CityBuilder) streetLamp(mb *MeshBuilder, x, y, z float32, faceX, faceZ float32) {
	mb.Cylinder(v3(x, y, z), v3(x, y+0.3, z), 0.22, 0.18, 8, ColLampPole, true)
	mb.Cylinder(v3(x, y+0.3, z), v3(x, y+5.2, z), 0.1, 0.08, 6, ColLampPole, false)
	tip := v3(x+faceX*1.3, y+5.4, z+faceZ*1.3)
	mb.Cylinder(v3(x, y+5.2, z), tip, 0.07, 0.07, 5, ColLampPole, false)
	mb.Box(v3(tip.X, tip.Y-0.05, tip.Z), v3(0.7, 0.2, 0.7), ColLampPole)
	mb.Emis = 1
	mb.Box(v3(tip.X, tip.Y-0.2, tip.Z), v3(0.5, 0.1, 0.5), ColLampLight)
	mb.Emis = 0
}

func (cb *CityBuilder) bench(mb *MeshBuilder, x, z, yaw float32) {
	y := cb.W.LandHeight(x, z)
	mb.SetMatrix(matTRS(v3(x, y, z), yaw, 1))
	wood := rl.NewColor(176, 110, 60, 255)
	mb.Box(v3(0, 0.5, 0), v3(2.0, 0.1, 0.6), wood)
	mb.Box(v3(0, 0.95, -0.28), v3(2.0, 0.5, 0.08), wood)
	for _, sx := range []float32{-0.85, 0.85} {
		mb.Box(v3(sx, 0.25, 0), v3(0.1, 0.5, 0.55), ColLampPole)
	}
	mb.ClearMatrix()
}

func (cb *CityBuilder) flowerBed(mb *MeshBuilder, x, z float32) {
	y := cb.W.LandHeight(x, z)
	mb.Cylinder(v3(x, y, z), v3(x, y+0.45, z), 1.5, 1.5, 14, rl.NewColor(190, 140, 100, 255), false)
	mb.Disc(v3(x, y+0.45, z), 0, 1.4, 14, rl.NewColor(110, 70, 40, 255))
	cols := []rl.Color{LegoRed, LegoYellow, LegoPink, LegoPurple, LegoWhite, LegoOrange}
	for i := 0; i < 9; i++ {
		a := float32(i) / 9 * 2 * math.Pi
		r := float32(0.4 + 0.6*float32(i%2))
		mb.Sphere(v3(x+cosf(a)*r, y+0.65, z+sinf(a)*r), 0.28, cols[(i+int(x))&0xff%len(cols)])
	}
	mb.Sphere(v3(x, y+0.75, z), 0.35, ColLeaf)
	cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 1.5, Top: y + 0.45})
}

func (cb *CityBuilder) buildVegetationAndProps() {
	mb := NewMB()
	treeCount := 0
	// copaci in parcuri si pe margini, cate 22 pe insula
	for si, s := range Sectors[:4] {
		placed := 0
		for tries := 0; tries < 900 && placed < 22; tries++ {
			x := s.Center.X + (cb.Rng.Float32()*2-1)*57
			z := s.Center.Y + (cb.Rng.Float32()*2-1)*57
			if cb.W.LandSD(x, z) > -GrassInset-1.5 {
				continue
			}
			if cb.nearRoad(x, z, 2.2) || cb.reserved(x, z, 3.2) {
				continue
			}
			if mh, _, _, _, _ := cb.W.mountainAt(x, z, 100); mh > 0.2 {
				continue
			}
			ok := true
			for _, t := range cb.Trees {
				dx, dz := t.X-x, t.Z-z
				if dx*dx+dz*dz < 30 {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			if si == 0 && cb.Rng.Float32() < 0.6 {
				cb.palmTree(mb, x, z, 0.9+cb.Rng.Float32()*0.3)
				cb.Trees = append(cb.Trees, v3(x, 0, z))
			} else {
				cols := []rl.Color{ColLeaf, ColLeaf2, rl.NewColor(255, 150, 190, 255), rl.NewColor(250, 190, 60, 255)}
				col := cols[0]
				r := cb.Rng.Float32()
				if r < 0.45 {
					col = cols[0]
				} else if r < 0.75 {
					col = cols[1]
				} else if r < 0.88 {
					col = cols[2]
				} else {
					col = cols[3]
				}
				cb.roundTree(mb, x, z, 0.9+cb.Rng.Float32()*0.4, col)
			}
			placed++
			treeCount++
		}
	}
	// insula centrala: inel de copaci si straturi de flori
	for i := 0; i < 14; i++ {
		a := float32(i)/14*2*math.Pi + 0.1
		x := cosf(a) * 38
		z := sinf(a) * 38
		if z < -6 && absf(x) < 44 {
			continue // spatiul literelor D-A-V-I-D
		}
		cb.roundTree(mb, x, z, 1.0, []rl.Color{ColLeaf, rl.NewColor(255, 150, 190, 255)}[i%2])
	}
	for i := 0; i < 8; i++ {
		a := float32(i)/8*2*math.Pi + 0.4
		x := cosf(a) * 56
		z := sinf(a) * 56
		if !cb.nearRoad(x, z, 3) && !cb.reserved(x, z, 3) {
			cb.flowerBed(mb, x, z)
		}
	}
	cb.addChunk(mb, true)

	// felinare si banci de-a lungul soselelor
	lm := NewMB()
	for _, r := range cb.Roads {
		if r.HalfW < 0 {
			continue
		}
		d := rl.Vector2Subtract(r.B, r.A)
		L := rl.Vector2Length(d)
		if L < 10 {
			continue
		}
		dir := rl.Vector2Scale(d, 1/L)
		per := rl.NewVector2(-dir.Y, dir.X)
		for t := float32(8); t < L-4; t += 22 {
			for _, s := range []float32{-1, 1} {
				lat := (r.HalfW - 0.6) * s
				x := r.A.X + dir.X*t + per.X*lat
				z := r.A.Y + dir.Y*t + per.Y*lat
				if cb.W.LandSD(x, z) > -1 || cb.W.InTunnel(x, z, 0) >= 0 {
					continue
				}
				if mh, _, _, _, _ := cb.W.mountainAt(x, z, 100); mh > 0.3 {
					continue
				}
				cb.streetLamp(lm, x, 0.2, z, -per.X*s, -per.Y*s)
			}
		}
	}
	// felinare pe inelul central
	for i := 0; i < 16; i++ {
		a := float32(i)/16*2*math.Pi + 0.2
		x, z := cosf(a)*42.5, sinf(a)*42.5
		cb.streetLamp(lm, x, 0.2, z, cosf(a), sinf(a))
	}
	// banci in piata centrala
	for i := 0; i < 8; i++ {
		a := float32(i)/8*2*math.Pi + 0.39
		x, z := cosf(a)*30, sinf(a)*30
		if z < -6 && absf(x) < 44 {
			continue
		}
		cb.bench(lm, x, z, -a+math.Pi/2)
	}
	cb.addChunk(lm, true)
}

// ---------------------------------------------------------------
// PIATA CENTRALA + FANTANA
// ---------------------------------------------------------------

func (cb *CityBuilder) buildPlaza() {
	mb := NewMB()
	up := v3(0, 1, 0)
	// pavaj cu model radial
	rings := []float32{0, 6, 14, 22, 30, 40, 42}
	for ri := 0; ri < len(rings)-1; ri++ {
		segs := 48
		for i := 0; i < segs; i++ {
			a0 := float32(i) / float32(segs) * 2 * math.Pi
			a1 := float32(i+1) / float32(segs) * 2 * math.Pi
			c := ColPaving
			if (i+ri)%2 == 0 {
				c = ColPaving2
			}
			if ri == len(rings)-2 {
				c = shade(ColPaving2, 0.9)
			}
			r0, r1 := rings[ri], rings[ri+1]
			y := float32(0.04)
			p := func(a, r float32) rl.Vector3 { return v3(cosf(a)*r, y, sinf(a)*r) }
			if r0 == 0 {
				mb.Tri(v3(0, y, 0), p(a0, r1), p(a1, r1), up, c)
				continue
			}
			// placi mici, ca gura scarilor sa ramana deschisa (se vad treptele)
			na := int((a1-a0)*r1/0.7) + 1
			nr := int((r1-r0)/0.7) + 1
			for ia := 0; ia < na; ia++ {
				for ir := 0; ir < nr; ir++ {
					b0 := a0 + (a1-a0)*float32(ia)/float32(na)
					b1 := a0 + (a1-a0)*float32(ia+1)/float32(na)
					q0 := r0 + (r1-r0)*float32(ir)/float32(nr)
					q1 := r0 + (r1-r0)*float32(ir+1)/float32(nr)
					bm, qm := (b0+b1)/2, (q0+q1)/2
					if cb.W.stairMouth(cosf(bm)*qm, sinf(bm)*qm, 0.35) {
						continue
					}
					mb.Quad(p(b0, q0), p(b1, q0), p(b1, q1), p(b0, q1), up, c)
				}
			}
		}
	}
	// fantana in sudul pietei
	fx, fz := float32(0), float32(20)
	stone := rl.NewColor(232, 226, 214, 255)
	mb.Tube(v3(fx, 0, fz), 4.6, 5.2, 0.8, 36, stone)
	mb.Disc(v3(fx, 0.55, fz), 0, 4.6, 36, rl.NewColor(80, 190, 230, 255))
	mb.Cylinder(v3(fx, 0.5, fz), v3(fx, 2.2, fz), 0.7, 0.5, 14, stone, false)
	mb.Cylinder(v3(fx, 2.2, fz), v3(fx, 2.5, fz), 2.0, 2.0, 24, stone, true)
	mb.Disc(v3(fx, 2.51, fz), 0, 1.8, 24, rl.NewColor(90, 200, 235, 255))
	mb.Cylinder(v3(fx, 2.5, fz), v3(fx, 3.6, fz), 0.35, 0.25, 10, stone, false)
	mb.Sphere(v3(fx, 3.8, fz), 0.45, LegoGold)
	cb.W.Circles = append(cb.W.Circles, CircleCollider{X: fx, Z: fz, R: 5.3, Top: 3})
	cb.addChunk(mb, true)
}

// ---------------------------------------------------------------
// PLACI LEGO PENTRU SANTIERE + TONOMATE
// ---------------------------------------------------------------

func (cb *CityBuilder) baseplate(mb *MeshBuilder, x, z float32, col rl.Color) {
	size := float32(10)
	mb.BeveledBox(v3(x, 0.12, z), v3(size, 0.24, size), 0.08, col)
	stud := shade(col, 1.06)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			sx := x - size/2 + 0.625 + float32(i)*1.25
			sz := z - size/2 + 0.625 + float32(j)*1.25
			mb.Cylinder(v3(sx, 0.24, sz), v3(sx, 0.4, sz), 0.3, 0.3, 8, stud, true)
		}
	}
}

func (cb *CityBuilder) vendorKiosk(mb *MeshBuilder, x, z float32, accent rl.Color) {
	y := float32(0.0)
	mb.BeveledBox(v3(x, y+0.25, z), v3(2.2, 0.5, 2.2), 0.12, LegoDarkGrey)
	mb.BeveledBox(v3(x, y+1.6, z), v3(1.8, 2.3, 1.6), 0.18, LegoGold)
	// ecranul (straluceste)
	mb.Emis = 1
	for _, s := range []float32{-1, 1} {
		mb.Box(v3(x, y+1.9, z+s*0.82), v3(1.2, 0.9, 0.04), LegoCyan)
	}
	mb.Emis = 0
	// butoane colorate
	for k, c := range []rl.Color{LegoRed, LegoGreen, LegoBlue} {
		for _, s := range []float32{-1, 1} {
			mb.Cylinder(v3(x-0.4+float32(k)*0.4, y+1.15, z+s*0.8), v3(x-0.4+float32(k)*0.4, y+1.15, z+s*0.9), 0.12, 0.12, 8, c, true)
		}
	}
	mb.Box(v3(x, y+2.85, z), v3(2.0, 0.2, 1.8), accent)
	// proiectorul hologramei
	mb.Cylinder(v3(x, y+2.95, z), v3(x, y+3.15, z), 0.8, 0.6, 16, LegoDarkGrey, true)
	mb.Emis = 1
	mb.Disc(v3(x, y+3.16, z), 0.3, 0.55, 16, LegoCyan)
	mb.Emis = 0
	cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 1.3, Top: 3.2})
}

// ---------------------------------------------------------------
// REPERE PE INSULE (statica)
// ---------------------------------------------------------------

func (cb *CityBuilder) landmarkPos(si int) rl.Vector2 {
	s := Sectors[si]
	return rl.NewVector2(s.Center.X-s.Inner.X*36, s.Center.Y-s.Inner.Y*36)
}

func (cb *CityBuilder) buildLandmarks(cw *CityWorld) {
	mb := NewMB()
	// 0. Dino Land: statuie urias T-Rex din "caramizi" rotunjite
	{
		p := cb.landmarkPos(0)
		cb.Reserve = append(cb.Reserve, CircleCollider{X: p.X, Z: p.Y, R: 14})
		g := rl.NewColor(80, 170, 70, 255)
		g2 := rl.NewColor(130, 205, 90, 255)
		mb.SetMatrix(matTRS(v3(p.X, 0, p.Y), 0.7, 1))
		mb.BeveledBox(v3(0, 0.6, 0), v3(9, 1.2, 9), 0.3, rl.NewColor(150, 140, 120, 255))
		mb.BeveledBox(v3(-1.2, 3.2, 0), v3(1.4, 4.0, 1.6), 0.4, g)
		mb.BeveledBox(v3(1.2, 3.2, 0), v3(1.4, 4.0, 1.6), 0.4, g)
		mb.BeveledBox(v3(0, 6.3, 0.5), v3(3.6, 3.4, 5.0), 0.9, g)
		mb.BeveledBox(v3(0, 6.0, -3.6), v3(1.6, 1.6, 4.0), 0.6, g)
		mb.BeveledBox(v3(0, 5.6, -6.4), v3(0.9, 0.9, 2.6), 0.4, g)
		mb.BeveledBox(v3(0, 8.6, 2.8), v3(2.2, 2.4, 2.2), 0.6, g)
		mb.BeveledBox(v3(0, 9.8, 4.4), v3(2.8, 2.2, 3.8), 0.7, g2)
		mb.BeveledBox(v3(0, 8.9, 4.9), v3(2.4, 0.8, 3.0), 0.3, g)
		for _, s := range []float32{-1, 1} {
			mb.Sphere(v3(s*1.3, 10.4, 5.0), 0.38, LegoWhite)
			mb.Sphere(v3(s*1.42, 10.45, 5.2), 0.2, LegoBlack)
			mb.BeveledBox(v3(s*1.8, 6.6, 2.8), v3(0.5, 0.5, 1.5), 0.2, g2)
		}
		for k := 0; k < 6; k++ {
			mb.Tri(v3(-0.9+float32(k)*0.36, 9.3, 6.3), v3(-0.75+float32(k)*0.36, 9.3, 6.3), v3(-0.82+float32(k)*0.36, 9.0, 6.3), v3(0, 0, 1), LegoWhite)
		}
		for k := 0; k < 5; k++ {
			mb.Cylinder(v3(0, 8.0-float32(k)*0.2, 1.6-float32(k)*1.3), v3(0, 8.9-float32(k)*0.25, 1.6-float32(k)*1.3), 0.5, 0.05, 6, LegoOrange, true)
		}
		mb.ClearMatrix()
		cb.W.Circles = append(cb.W.Circles, CircleCollider{X: p.X, Z: p.Y, R: 5.2, Top: 12})
	}
	// 1. Cartoon City: baza rotii mari (roata se misca, e desenata separat)
	{
		p := cb.landmarkPos(1)
		cb.Reserve = append(cb.Reserve, CircleCollider{X: p.X, Z: p.Y, R: 16})
		if len(cw.Wheels) > 0 {
			cb.wheelStand(mb, &cw.Wheels[0])
		}
	}
	// 2. Turbo Speedway: pista ovala cu linie de start in carouri si tribuna
	{
		p := cb.landmarkPos(2)
		cb.Reserve = append(cb.Reserve, CircleCollider{X: p.X, Z: p.Y, R: 20})
		segs := 64
		ax, az := float32(16), float32(10)
		for i := 0; i < segs; i++ {
			a0 := float32(i) / float32(segs) * 2 * math.Pi
			a1 := float32(i+1) / float32(segs) * 2 * math.Pi
			pt := func(a, off, y float32) rl.Vector3 {
				return v3(p.X+cosf(a)*(ax+off), y, p.Y+sinf(a)*(az+off))
			}
			c := rl.NewColor(64, 66, 74, 255)
			mb.Quad(pt(a0, -2.5, 0.06), pt(a1, -2.5, 0.06), pt(a1, 2.5, 0.06), pt(a0, 2.5, 0.06), v3(0, 1, 0), c)
			kerb := LegoRed
			if i%2 == 0 {
				kerb = LegoWhite
			}
			mb.Quad(pt(a0, 2.5, 0.08), pt(a1, 2.5, 0.08), pt(a1, 3.2, 0.08), pt(a0, 3.2, 0.08), v3(0, 1, 0), kerb)
			mb.Quad(pt(a0, -3.2, 0.08), pt(a1, -3.2, 0.08), pt(a1, -2.5, 0.08), pt(a0, -2.5, 0.08), v3(0, 1, 0), kerb)
		}
		for i := 0; i < 8; i++ {
			for j := 0; j < 2; j++ {
				c := LegoWhite
				if (i+j)%2 == 0 {
					c = LegoBlack
				}
				x0 := p.X + ax - 2.5 + float32(i)*0.625
				z0 := p.Y + float32(j)*0.625 - 0.625
				mb.Quad(v3(x0, 0.075, z0), v3(x0+0.625, 0.075, z0), v3(x0+0.625, 0.075, z0+0.625), v3(x0, 0.075, z0+0.625), v3(0, 1, 0), c)
			}
		}
		// tribuna
		for k := 0; k < 4; k++ {
			mb.Box(v3(p.X, 0.4+float32(k)*0.8, p.Y-az-6-float32(k)*1.0), v3(18, 0.8+float32(k)*1.6, 1.2), []rl.Color{LegoRed, LegoYellow, LegoBlue, LegoWhite}[k])
		}
		mb.Box(v3(p.X, 6.0, p.Y-az-8), v3(19, 0.3, 4.5), LegoWhite)
		for _, s := range []float32{-9, 9} {
			mb.Cylinder(v3(p.X+s, 0, p.Y-az-6.2), v3(p.X+s, 6, p.Y-az-6.2), 0.15, 0.15, 6, LegoLightGrey, false)
		}
		cb.W.Boxes = append(cb.W.Boxes, BoxCollider{MinX: p.X - 9.5, MaxX: p.X + 9.5, MinZ: p.Y - az - 10.5, MaxZ: p.Y - az - 5.4, Top: 7})
	}
	// 3. Mecha Arena: arena rotunda + turnurile eolienelor
	{
		p := cb.landmarkPos(3)
		cb.Reserve = append(cb.Reserve, CircleCollider{X: p.X, Z: p.Y, R: 16})
		mb.Tube(v3(p.X, 0, p.Y), 8.5, 10, 4.5, 36, rl.NewColor(180, 190, 205, 255))
		mb.Emis = 1
		mb.Tube(v3(p.X, 4.5, p.Y), 9.2, 10.05, 0.3, 36, LegoCyan)
		mb.Emis = 0
		mb.Disc(v3(p.X, 0.05, p.Y), 0, 8.5, 36, rl.NewColor(90, 100, 120, 255))
		mb.Disc(v3(p.X, 0.07, p.Y), 3.0, 3.4, 36, LegoCyan)
		cb.W.Circles = append(cb.W.Circles, CircleCollider{X: p.X, Z: p.Y, R: 10, Top: 5})
	}
	cb.addChunk(mb, true)
}

// Pozitiile eolienelor pe Mecha Arena (desenate dinamic ca sa se roteasca)
func turbinePositions() []rl.Vector3 {
	c := Sectors[3].Center
	return []rl.Vector3{v3(c.X+52, 0, c.Y-26), v3(c.X-26, 0, c.Y+52), v3(c.X+53, 0, c.Y+20)}
}

// inMountain: ceva de raza r ar intra in munte (cladirile nu se construiesc acolo)
func (cb *CityBuilder) inMountain(x, z, r float32) bool {
	for _, o := range [][2]float32{{0, 0}, {r, r}, {-r, r}, {r, -r}, {-r, -r}} {
		if mh, _, _, _, _ := cb.W.mountainAt(x+o[0], z+o[1], 100); mh > 0.3 {
			return true
		}
	}
	return false
}
