package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 6.0: RealmBuilder construieste harta unei lumi de uscat (lumile 1+).
//
// Harta: patrat de ~600 m, margine de munti/dealuri inalte (nu se iese).
// Zona 0 (centrul, gara + prima insula-nivel) si 4 zone la colturi,
// legate prin drumuri (radiale + un inel exterior prin centrele zonelor).
// Fiecare zona: piateta in centru (raza 13), cabina telefonica, 2 santiere.
// Intre zone: dealuri line; langa zone si drumuri terenul e plat (y = 0).
//
// API pentru decor (Terra), folosit in functia data la registerRealmDecor:
//   rb.H(x,z)            inaltimea solului
//   rb.Free(x,z,r)       loc liber (nu drum, nu piateta/cabina/santier/gara, in harta)
//   rb.Reserve(x,z,r)    ocupa locul (sa nu mai puna altcineva ceva acolo)
//   rb.Mesh()            *MeshBuilder nou; rb.Add(mb, umbre) il adauga in lume
//                        (fa cate un mesh pe zona / pe bucati de ~80 m)
//   rb.Solid(cx,cz,w,d,h,y0)   cutie solida (nu se trece prin ea; se poate sta pe acoperis)
//   rb.SolidRoof(...)          la fel, cu acoperis in doua ape (RoofRidgeX/RoofRidgeZ)
//   rb.Pillar(x,z,r,top)       obstacol rotund (trunchi, stalp, statuie)
//   rb.Zones[i].C, .R          centrul si raza zonei i (0 = centrul hartii)
//   rb.ZoneAt(x,z)             zona unui punct (-1 = intre zone)
//   rb.Rng                     numere aleatoare (determinist)
//   rb.Def                     tema (culori: Grass, Dirt, Path, Main, Accent)
// ====================================================================

const (
	RealmHalf   = float32(300) // marginea hartii (de aici incep muntii)
	RealmEdge   = float32(286) // pana aici poate merge Aventurierul
	RealmRoadW  = float32(5)   // jumatate din latimea drumului
	RealmPlazaR = float32(13)  // piateta din centrul zonei
)

type RealmZone struct {
	C rl.Vector2
	R float32
}

type RealmBuilder struct {
	CB    *CityBuilder
	W     *World
	CW    *CityWorld
	Def   *RealmDef
	RI    int
	Zones []RealmZone
	Rng   *rand.Rand
	seed  float32
}

var realmZoneLayout = []rl.Vector2{{X: 0, Y: 0}, {X: 160, Y: -150}, {X: -160, Y: -150}, {X: -160, Y: 150}, {X: 160, Y: 150}}

// ---------------- API pentru decor ----------------

func (rb *RealmBuilder) H(x, z float32) float32 { return rb.W.NaturalLand(x, z) }

func (rb *RealmBuilder) Free(x, z, r float32) bool {
	if absf(x)+r > RealmEdge-4 || absf(z)+r > RealmEdge-4 {
		return false
	}
	if rb.CB.nearRoad(x, z, r+1.5) || rb.CB.reserved(x, z, r+1) {
		return false
	}
	for _, zn := range rb.Zones {
		if sq(x-zn.C.X)+sq(z-zn.C.Y) < sq(RealmPlazaR+r+2) {
			return false
		}
	}
	return true
}

func (rb *RealmBuilder) Reserve(x, z, r float32) {
	rb.CB.Reserve = append(rb.CB.Reserve, CircleCollider{X: x, Z: z, R: r})
}

func (rb *RealmBuilder) Mesh() *MeshBuilder { return NewMB() }

func (rb *RealmBuilder) Add(mb *MeshBuilder, shadows bool) { rb.CB.addChunk(mb, shadows) }

func (rb *RealmBuilder) Solid(cx, cz, w, d, h, y0 float32) {
	rb.W.Boxes = append(rb.W.Boxes, BoxCollider{MinX: cx - w/2, MaxX: cx + w/2, MinZ: cz - d/2, MaxZ: cz + d/2, Top: y0 + h, Roof: RoofFlat})
}

func (rb *RealmBuilder) SolidRoof(cx, cz, w, d, h, y0 float32, roof int, roofH float32) {
	rb.W.Boxes = append(rb.W.Boxes, BoxCollider{MinX: cx - w/2, MaxX: cx + w/2, MinZ: cz - d/2, MaxZ: cz + d/2, Top: y0 + h, Roof: roof, RoofH: roofH})
}

func (rb *RealmBuilder) Pillar(x, z, r, top float32) {
	rb.W.Circles = append(rb.W.Circles, CircleCollider{X: x, Z: z, R: r, Top: top})
}

func (rb *RealmBuilder) ZoneAt(x, z float32) int {
	for i, zn := range rb.Zones {
		if sq(x-zn.C.X)+sq(z-zn.C.Y) < sq(zn.R+20) {
			return i
		}
	}
	return -1
}

// ---------------- teren ----------------

func vnoise(x, z float32) float32 {
	// zgomot neted ieftin (suma de sinusuri rotite)
	return (sinf(x*0.021+z*0.013) + sinf(x*0.008-z*0.019+1.7)*1.3 + sinf(x*0.047+z*0.041+4.1)*0.35 + sinf(-x*0.033+z*0.027+2.3)*0.5) / 3.15
}

// realmHeight: dealuri line intre zone, munti pe margine, plat langa zone si drumuri.
func (rb *RealmBuilder) height(x, z float32) float32 {
	d := maxf(absf(x), absf(z))
	border := smoothstep(RealmHalf-34, RealmHalf+30, d) * (34 + 10*vnoise(x*1.7+rb.seed, z*1.7))
	hills := (vnoise(x+rb.seed, z-rb.seed)*0.5 + 0.5) * 7
	m := float32(1)
	for _, zn := range rb.Zones {
		dz := sqrtf(sq(x-zn.C.X) + sq(z-zn.C.Y))
		m = minf(m, smoothstep(zn.R+2, zn.R+26, dz))
	}
	for _, r := range rb.CB.Roads {
		dr := distToSeg(x, z, r.A, r.B)
		m = minf(m, smoothstep(r.HalfW+3, r.HalfW+16, dr))
	}
	return hills*m + border
}

func (rb *RealmBuilder) groundColor(x, z, h float32, ny float32) rl.Color {
	D := rb.Def
	n := vnoise(x*3.1, z*3.1)*0.5 + 0.5
	c := lerpColor(D.Grass, shade(D.Grass, 0.86), n)
	if ny < 0.82 || h > 16 {
		k := clampf(maxf((0.82-ny)*5, (h-16)/8), 0, 1)
		c = lerpColor(c, D.Dirt, k)
	}
	for _, zn := range rb.Zones {
		dz := sqrtf(sq(x-zn.C.X) + sq(z-zn.C.Y))
		if dz < RealmPlazaR+1 {
			// piateta pavata, cu inele
			ring := float32(math.Mod(float64(dz), 3)) < 1.5
			p := D.Path
			if ring {
				p = shade(p, 1.08)
			}
			return p
		}
	}
	return c
}

func lerpColor(a, b rl.Color, t float32) rl.Color {
	t = clampf(t, 0, 1)
	return rl.NewColor(uint8(float32(a.R)+(float32(b.R)-float32(a.R))*t), uint8(float32(a.G)+(float32(b.G)-float32(a.G))*t), uint8(float32(a.B)+(float32(b.B)-float32(a.B))*t), 255)
}

func (rb *RealmBuilder) buildTerrain() {
	step := float32(3)
	ext := float32(340)
	n := int(2 * ext / step)
	hs := make([]float32, (n+1)*(n+1))
	for j := 0; j <= n; j++ {
		for i := 0; i <= n; i++ {
			hs[j*(n+1)+i] = rb.H(-ext+float32(i)*step, -ext+float32(j)*step)
		}
	}
	nrm := func(i, j int) rl.Vector3 {
		i0, i1 := maxInt(i-1, 0), minInt(i+1, n)
		j0, j1 := maxInt(j-1, 0), minInt(j+1, n)
		dx := hs[j*(n+1)+i1] - hs[j*(n+1)+i0]
		dz := hs[j1*(n+1)+i] - hs[j0*(n+1)+i]
		return rl.Vector3Normalize(v3(-dx/(float32(i1-i0)*step), 1, -dz/(float32(j1-j0)*step)))
	}
	parts := map[int]*MeshBuilder{}
	for j := 0; j < n; j++ {
		for i := 0; i < n; i++ {
			key := (i * 3 / 80) + 1000*(j*3/80)
			mb := parts[key]
			if mb == nil {
				mb = NewMB()
				parts[key] = mb
			}
			var p [4]rl.Vector3
			var nn [4]rl.Vector3
			var cc [4]rl.Color
			for k, o := range [4][2]int{{0, 0}, {1, 0}, {1, 1}, {0, 1}} {
				ii, jj := i+o[0], j+o[1]
				x, z := -ext+float32(ii)*step, -ext+float32(jj)*step
				h := hs[jj*(n+1)+ii]
				p[k] = v3(x, h, z)
				nn[k] = nrm(ii, jj)
				cc[k] = rb.groundColor(x, z, h, nn[k].Y)
			}
			mb.TriS(p[0], nn[0], p[2], nn[2], p[1], nn[1], cc[0], cc[2], cc[1])
			mb.TriS(p[0], nn[0], p[3], nn[3], p[2], nn[2], cc[0], cc[3], cc[2])
		}
	}
	for _, mb := range parts {
		rb.CB.addChunk(mb, false)
	}
}

// ---------------- drumuri ----------------

func (rb *RealmBuilder) layoutRoads() {
	for i := 1; i < len(rb.Zones); i++ {
		rb.CB.Roads = append(rb.CB.Roads, RoadSeg{A: rb.Zones[0].C, B: rb.Zones[i].C, HalfW: RealmRoadW})
	}
	ring := []int{1, 2, 3, 4}
	for k := range ring {
		a, b := rb.Zones[ring[k]].C, rb.Zones[ring[(k+1)%4]].C
		rb.CB.Roads = append(rb.CB.Roads, RoadSeg{A: a, B: b, HalfW: RealmRoadW})
	}
}

func (rb *RealmBuilder) buildRoadMeshes() {
	D := rb.Def
	edge := shade(D.Path, 0.75)
	for _, r := range rb.CB.Roads {
		mb := NewMB()
		d := rl.Vector2Normalize(rl.Vector2Subtract(r.B, r.A))
		side := rl.NewVector2(-d.Y, d.X)
		L := rl.Vector2Distance(r.A, r.B)
		seg := int(L/3) + 1
		for i := 0; i < seg; i++ {
			t0, t1 := float32(i)/float32(seg), float32(i+1)/float32(seg)
			a, b := rl.Vector2Lerp(r.A, r.B, t0), rl.Vector2Lerp(r.A, r.B, t1)
			// nu desenam drumul peste piateta
			if rb.inPlaza(a.X, a.Y, 0) && rb.inPlaza(b.X, b.Y, 0) {
				continue
			}
			for _, sd := range [][3]float32{{-r.HalfW, r.HalfW, 0}, {-r.HalfW - 0.6, -r.HalfW, 1}, {r.HalfW, r.HalfW + 0.6, 1}} {
				col := D.Path
				if sd[2] == 1 {
					col = edge
				} else if i%2 == 0 {
					col = shade(D.Path, 0.97)
				}
				q := func(p rl.Vector2, s float32) rl.Vector3 {
					x, z := p.X+side.X*s, p.Y+side.Y*s
					return v3(x, rb.H(x, z)+0.05, z)
				}
				mb.Quad(q(a, sd[0]), q(b, sd[0]), q(b, sd[1]), q(a, sd[1]), v3(0, 1, 0), col)
			}
		}
		rb.CB.addChunk(mb, false)
	}
}

func (rb *RealmBuilder) inPlaza(x, z, r float32) bool {
	for _, zn := range rb.Zones {
		if sq(x-zn.C.X)+sq(z-zn.C.Y) < sq(RealmPlazaR+r) {
			return true
		}
	}
	return false
}

// ---------------- decor de baza (pana vine decorul tematic) ----------------

func (rb *RealmBuilder) genericDecor() {
	D := rb.Def
	trees := NewMB()
	cnt := 0
	for tries := 0; tries < 9000 && cnt < 520; tries++ {
		x := (rb.Rng.Float32()*2 - 1) * (RealmEdge - 6)
		z := (rb.Rng.Float32()*2 - 1) * (RealmEdge - 6)
		if !rb.Free(x, z, 2.5) {
			continue
		}
		zi := rb.ZoneAt(x, z)
		if zi >= 0 && rb.Rng.Float32() < 0.7 {
			continue // zonele raman mai libere (lupta)
		}
		y := rb.H(x, z)
		s := 0.9 + rb.Rng.Float32()*0.8
		switch D.ID {
		case "craciun", "monstri":
			rb.CB.pineTree(trees, x, y, z, s*1.2)
			if D.ID == "craciun" {
				trees.Cylinder(v3(x, y+3.5*s*1.2, z), v3(x, y+4.6*s*1.2, z), 0.9*s, 0.05, 8, rl.NewColor(250, 252, 255, 255), true)
			}
			rb.Pillar(x, z, 0.4*s, y+6)
		default:
			rb.CB.roundTree(trees, x, z, s, shade(D.Grass, 0.75))
		}
		rb.Reserve(x, z, 2)
		cnt++
		if trees.VertexCount() > 40000 {
			rb.Add(trees, true)
			trees = NewMB()
		}
	}
	rb.Add(trees, true)
	// case simple la marginea fiecarei zone (culorile zonei)
	for zi, zn := range rb.Zones {
		zd := rb.Def.Zones[zi]
		mb := NewMB()
		for k := 0; k < 7; k++ {
			a := float32(k)/7*2*math.Pi + float32(zi)
			r := zn.R - 10
			x, z := zn.C.X+cosf(a)*r, zn.C.Y+sinf(a)*r
			w, d := float32(7+rb.Rng.Intn(4)), float32(7+rb.Rng.Intn(3))
			if !rb.Free(x, z, maxf(w, d)/2+1) {
				continue
			}
			y := rb.H(x, z)
			h := float32(4 + rb.Rng.Intn(5))
			mb.BeveledBox(v3(x, y+h/2, z), v3(w, h, d), 0.3, lerpColor(zd.Main, LegoWhite, 0.45))
			// acoperis in doua ape
			rh := float32(2.2)
			mb.Quad(v3(x-w/2-0.4, y+h, z-d/2-0.4), v3(x+w/2+0.4, y+h, z-d/2-0.4), v3(x+w/2+0.4, y+h+rh, z), v3(x-w/2-0.4, y+h+rh, z), v3(0, 0.7, -0.7), zd.Accent)
			mb.Quad(v3(x+w/2+0.4, y+h, z+d/2+0.4), v3(x-w/2-0.4, y+h, z+d/2+0.4), v3(x-w/2-0.4, y+h+rh, z), v3(x+w/2+0.4, y+h+rh, z), v3(0, 0.7, 0.7), zd.Accent)
			mb.Tri(v3(x-w/2, y+h, z-d/2), v3(x-w/2, y+h, z+d/2), v3(x-w/2, y+h+rh, z), v3(-1, 0, 0), lerpColor(zd.Main, LegoWhite, 0.45))
			mb.Tri(v3(x+w/2, y+h, z+d/2), v3(x+w/2, y+h, z-d/2), v3(x+w/2, y+h+rh, z), v3(1, 0, 0), lerpColor(zd.Main, LegoWhite, 0.45))
			// usa si ferestre spre centrul zonei
			fx, fz := (zn.C.X - x), (zn.C.Y - z)
			fl := sqrtf(fx*fx+fz*fz) + 0.01
			fx, fz = fx/fl, fz/fl
			mb.Box(v3(x+fx*(w/2+0.02), y+1.1, z+fz*(d/2+0.02)), v3(1.2+absf(fz)*0, 2.2, 0.1+absf(fx)*0), shade(zd.Accent, 0.6))
			rb.SolidRoof(x, z, w, d, h, y, RoofRidgeX, rh)
			rb.Reserve(x, z, maxf(w, d)/2+1)
		}
		// statuie / fantana in mijlocul pietetei
		y := rb.H(zn.C.X, zn.C.Y)
		mb.Cylinder(v3(zn.C.X, y, zn.C.Y), v3(zn.C.X, y+0.6, zn.C.Y), 4.2, 4.2, 20, shade(rb.Def.Path, 0.85), true)
		mb.Cylinder(v3(zn.C.X, y+0.6, zn.C.Y), v3(zn.C.X, y+3.2, zn.C.Y), 0.8, 0.6, 12, zd.Main, true)
		mb.Ellipsoid(v3(zn.C.X, y+3.8, zn.C.Y), v3(1.2, 1.2, 1.2), 6, 10, zd.Accent)
		rb.Pillar(zn.C.X, zn.C.Y, 4.2, y+0.6)
		rb.Add(mb, true)
	}
}

// ---------------- harta intreaga ----------------

// InitRealmWorld: construieste lumea ri (>0) intr-un CityWorld nou.
func InitRealmWorld(ri int, progress func(string)) *CityWorld {
	if progress == nil {
		progress = func(string) {}
	}
	D := &Realms[ri]
	w := &World{}
	w.Lands = []LandShape{{Kind: 0, C: rl.NewVector2(0, 0), HalfU: 420, HalfV: 420, Corner: 10}}
	cb := &CityBuilder{W: w, Rng: rand.New(rand.NewSource(int64(2026 + ri*31)))}
	cw := &CityWorld{W: w, Rng: rand.New(rand.NewSource(int64(7 + ri)))}
	rb := &RealmBuilder{CB: cb, W: w, CW: cw, Def: D, RI: ri, Rng: rand.New(rand.NewSource(int64(99 + ri*7))), seed: float32(ri) * 37.7}
	for i := range D.Zones {
		if i >= len(realmZoneLayout) {
			break
		}
		r := float32(62)
		if i == 0 {
			r = 58
		}
		rb.Zones = append(rb.Zones, RealmZone{C: realmZoneLayout[i], R: r})
	}
	w.HeightFn = rb.height
	progress("Desenăm drumurile...")
	rb.layoutRoads()

	// santiere (2 pe zona), cabine, gara
	blueprints := GetLegoBlueprints()
	addSite := func(pos, vend rl.Vector3, bp LegoItemDefinition, yaw float32) {
		idx := len(cw.LegoSites)
		cw.LegoSites = append(cw.LegoSites, ConstructionSite{
			Position: pos, Blueprint: bp, Meshes: BuildBlueprintMeshes(bp),
			TotalStages: len(bp.Bricks), UpgradeLevel: 1, TargetScale: 1, CurrentScale: 1, Yaw: yaw, ColliderIdx: -1,
		})
		cw.VendorPods = append(cw.VendorPods, VendorMachine{Position: vend, SlotIndex: idx})
		cb.Reserve = append(cb.Reserve, CircleCollider{X: pos.X, Z: pos.Z, R: 9}, CircleCollider{X: vend.X, Z: vend.Z, R: 3})
	}
	for zi, zn := range rb.Zones {
		// directiile libere de drum: intre drumuri (pe diagonale fata de drumurile zonei)
		base := float32(zi) * 0.9
		for k := 0; k < 2; k++ {
			placed := false
			for t := 0; t < 24 && !placed; t++ {
				a := base + float32(k)*math.Pi + float32(t)*0.26
				r := float32(30)
				x, z := zn.C.X+cosf(a)*r, zn.C.Y+sinf(a)*r
				vx, vz := zn.C.X+cosf(a)*(r-11), zn.C.Y+sinf(a)*(r-11)
				if cb.nearRoad(x, z, 8) || cb.nearRoad(vx, vz, 3) || cb.reserved(x, z, 9) {
					continue
				}
				yaw := atan2f(zn.C.X-x, zn.C.Y-z)
				addSite(v3(x, 0, z), v3(vx, 0, vz), blueprints[(ri*3+zi*2+k)%len(blueprints)], yaw)
				placed = true
			}
		}
		// cabina: langa piateta, intre drumuri
		for t := 0; t < 48; t++ {
			a := base + 0.8 + float32(t)*0.131
			r := RealmPlazaR + 7
			x, z := zn.C.X+cosf(a)*r, zn.C.Y+sinf(a)*r
			if cb.nearRoad(x, z, 2.5) || cb.reserved(x, z, 3) {
				continue
			}
			cw.Booths = append(cw.Booths, Booth{Pos: v3(x, 0, z), Yaw: atan2f(zn.C.X-x, zn.C.Y-z), Sector: zi, World: zi})
			cb.Reserve = append(cb.Reserve, CircleCollider{X: x, Z: z, R: 4.5})
			break
		}
	}
	// gara: in zona 0, langa piateta
	cw.Station = realmStationSpot(cb, rb.Zones[0].C)
	cb.Reserve = append(cb.Reserve, CircleCollider{X: cw.Station.X, Z: cw.Station.Z, R: 16})

	progress("Construim lumea " + D.Name + "...")
	for _, key := range []string{D.ID, D.ID + "_2"} {
		if f := realmDecor[key]; f != nil {
			f(rb)
		}
	}
	rb.genericDecor()
	{
		mb := NewMB()
		for i := range cw.LegoSites {
			s := &cw.LegoSites[i]
			cb.baseplate(mb, s.Position.X, s.Position.Z, lerpColor(D.Main, LegoWhite, 0.4))
			w.Boxes = append(w.Boxes, BoxCollider{MinX: s.Position.X - 5, MaxX: s.Position.X + 5, MinZ: s.Position.Z - 5, MaxZ: s.Position.Z + 5, Top: 0.4})
		}
		for _, v := range cw.VendorPods {
			cb.vendorKiosk(mb, v.Position.X, v.Position.Z, D.Main)
		}
		buildStationMesh(mb, cw.Station, D)
		stationColliders(w, cw.Station)
		cb.addChunk(mb, true)
	}
	progress("Modelăm dealurile...")
	rb.buildTerrain()
	rb.buildRoadMeshes()
	cw.Chunks = cb.Chunks
	cw.Roads = cb.Roads
	w.Roads = cb.Roads
	cw.finishBooths()
	w.RebuildGrid()
	for i := range cw.LegoSites {
		s := &cw.LegoSites[i]
		s.ColliderIdx = w.AddCircle(s.Position.X, s.Position.Z, 0, 0)
	}

	progress("Pornim trăsurile...")
	vf := realmVehicles[D.Vehicle]
	cols := []rl.Color{D.Main, D.Accent, LegoWhite, shade(D.Main, 0.7), LegoYellow}
	for _, c := range cols {
		if vf != nil {
			cw.CarMeshes = append(cw.CarMeshes, vf(c))
		} else {
			cw.CarMeshes = append(cw.CarMeshes, buildCarMesh(c, false))
		}
	}
	bootLog("realm: vehicule gata")
	cw.Routes = []*Route{rb.ringRoute(false), rb.ringRoute(true), rb.starRoute()}
	bootLog("realm: trasee gata")
	k := 0
	for ri2, r := range cw.Routes {
		n := []int{7, 7, 8}[ri2]
		for i := 0; i < n; i++ {
			spd := 9 + cw.Rng.Float32()*3
			cw.Cars = append(cw.Cars, Car{Route: r, S: r.Len * float32(i) / float32(n), Speed: spd, MaxSpd: spd, Mesh: k % len(cw.CarMeshes), HP: CarMaxHP})
			k++
		}
	}
	bootLog("realm: masini gata")
	// locuitorii lumii
	for i := 0; i < 3; i++ {
		cw.CitizenMeshes = append(cw.CitizenMeshes, GetCharMeshes(RealmCharStyle(ri, RoleNPC1+i)))
	}
	for i := 0; i < 36; i++ {
		zn := rb.Zones[i%len(rb.Zones)]
		hr := zn.R - 8
		var pos rl.Vector3
		for tries := 0; tries < 80; tries++ {
			x := zn.C.X + (cw.Rng.Float32()*2-1)*hr
			z := zn.C.Y + (cw.Rng.Float32()*2-1)*hr
			if w.CanStand(x, z, 0.3, 0.4) {
				g, _ := w.GroundAt(x, z, 2)
				pos = v3(x, g, z)
				break
			}
		}
		cw.Citizens = append(cw.Citizens, Citizen{Pos: pos, Yaw: cw.Rng.Float32() * 6.28, Speed: 1.5 + cw.Rng.Float32()*0.6,
			Style: i % 3, Home: zn.C, HomeR: hr, Pose: Pose{Walk: cw.Rng.Float32() * 6}})
	}
	for i := 0; i < 6; i++ {
		cw.CloudMeshes = append(cw.CloudMeshes, buildCloudMesh(i*13+5+ri))
	}
	for i := 0; i < 18; i++ {
		cw.Clouds = append(cw.Clouds, Cloud{Pos: v3((cw.Rng.Float32()*2-1)*420, 95+cw.Rng.Float32()*40, (cw.Rng.Float32()*2-1)*420), Mesh: i % 6, S: 0.9 + cw.Rng.Float32()*0.8})
	}
	bootLog("realm: locuitori si nori gata")
	cw.Fountain = v3(0, 0, 0)
	cw.Realm = ri
	cw.Zones = rb.Zones
	buildRealmNav(cw, rb)
	return cw
}

// ---------------- traseele vehiculelor ----------------

func (rb *RealmBuilder) zoneArc(r *routeBuilder, c, dIn, dOut rl.Vector2, rad float32) {
	d := float32(0.3)
	r.arc(c, rad, screenAngle(dIn)+d, screenAngle(dOut)-d)
}

func (rb *RealmBuilder) ringRoute(reverse bool) *Route {
	order := []int{1, 2, 3, 4}
	if reverse {
		order = []int{4, 3, 2, 1}
	}
	r := &routeBuilder{}
	rad := RealmPlazaR + 4
	for k := range order {
		prev := rb.Zones[order[(k+3)%4]].C
		cur := rb.Zones[order[k]].C
		next := rb.Zones[order[(k+1)%4]].C
		dIn := rl.Vector2Normalize(rl.Vector2Subtract(prev, cur))
		dOut := rl.Vector2Normalize(rl.Vector2Subtract(next, cur))
		rb.zoneArc(r, cur, dIn, dOut, rad)
		r.line(rl.Vector2Add(cur, rl.Vector2Scale(dOut, rad+2)), rl.Vector2Add(next, rl.Vector2Scale(rl.Vector2Scale(dOut, -1), rad+2)), 2.2)
	}
	name := "inel"
	if reverse {
		name = "inel-invers"
	}
	return rb.W.finishRoute(r, name)
}

func (rb *RealmBuilder) starRoute() *Route {
	r := &routeBuilder{}
	rad := RealmPlazaR + 4
	hub := rb.Zones[0].C
	for k := 1; k < len(rb.Zones); k++ {
		zc := rb.Zones[k].C
		out := rl.Vector2Normalize(rl.Vector2Subtract(zc, hub))
		back := rl.Vector2Scale(out, -1)
		nextZ := rb.Zones[1+k%(len(rb.Zones)-1)].C
		nextOut := rl.Vector2Normalize(rl.Vector2Subtract(nextZ, hub))
		r.line(rl.Vector2Add(hub, rl.Vector2Scale(out, rad+2)), rl.Vector2Add(zc, rl.Vector2Scale(back, rad+2)), 2.2)
		rb.zoneArc(r, zc, back, rl.Vector2Rotate(back, 0.6), rad) // intoarcere in jurul pietetei
		r.line(rl.Vector2Add(zc, rl.Vector2Scale(back, rad+2)), rl.Vector2Add(hub, rl.Vector2Scale(out, rad+2)), 2.2)
		rb.zoneArc(r, hub, out, nextOut, rad)
	}
	return rb.W.finishRoute(r, "stea")
}

// ---------------- GPS ----------------

func buildRealmNav(cw *CityWorld, rb *RealmBuilder) {
	w := cw.W
	ng := &NavGraph{grid: map[int64][]int32{}}
	for _, r := range cw.Roads {
		ng.polyline(segPts(w, r.A, r.B, 0), false)
	}
	for _, zn := range rb.Zones {
		ring := []rl.Vector3{}
		for k := 0; k < 24; k++ {
			a := float32(k) / 24 * 2 * math.Pi
			x, z := zn.C.X+cosf(a)*(RealmPlazaR+3), zn.C.Y+sinf(a)*(RealmPlazaR+3)
			ring = append(ring, v3(x, rb.H(x, z), z))
		}
		ng.polyline(ring, true)
	}
	for i := range cw.Booths {
		b := &cw.Booths[i]
		d := b.DoorPos(2.5)
		ni := ng.add(v3(d.X, b.Pos.Y, d.Z))
		ng.joinNear(ni, 40, 3)
	}
	st := cw.StationDoor()
	ni := ng.add(st)
	ng.joinNear(ni, 40, 3)
	for i := range ng.N {
		ng.joinNear(int32(i), 5.5, 1.6)
	}
	cw.Nav = ng
}
