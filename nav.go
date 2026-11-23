package main

import (
	"container/heap"
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 5.0: SAGETI GPS pe jos. Graful de drumuri se face din traseele
// masinilor (poduri, tuneluri) + cateva legaturi pietonale (scarile din
// piata, podurile noi, cabinele). Drumul se recalculeaza des.
// ====================================================================

type GPSState struct {
	Active     bool
	World      int // lumea tinta (-1 = punct oarecare)
	FromInside bool
	Target     rl.Vector3
	Path       []rl.Vector3
	RepathT    float32
	Flow       float32
}

type navNode struct {
	P    rl.Vector3
	Next []int32
}

type NavGraph struct {
	N    []navNode
	grid map[int64][]int32
}

var nav *NavGraph

func navKey(x, z float32) int64 {
	return int64(math.Floor(float64(x/8)))*100003 + int64(math.Floor(float64(z/8)))
}

func (ng *NavGraph) add(p rl.Vector3) int32 {
	i := int32(len(ng.N))
	ng.N = append(ng.N, navNode{P: p})
	k := navKey(p.X, p.Z)
	ng.grid[k] = append(ng.grid[k], i)
	return i
}

func (ng *NavGraph) link(a, b int32) {
	if a == b {
		return
	}
	for _, x := range ng.N[a].Next {
		if x == b {
			return
		}
	}
	ng.N[a].Next = append(ng.N[a].Next, b)
	ng.N[b].Next = append(ng.N[b].Next, a)
}

func (ng *NavGraph) polyline(pts []rl.Vector3, closed bool) {
	prev := int32(-1)
	first := int32(-1)
	for _, p := range pts {
		i := ng.add(p)
		if prev >= 0 {
			ng.link(prev, i)
		} else {
			first = i
		}
		prev = i
	}
	if closed && first >= 0 && prev != first {
		ng.link(prev, first)
	}
}

// segPts: puncte la ~4 m intre a si b, cu inaltimea luata de pe sol.
func segPts(w *World, a, b rl.Vector2, y0 float32) []rl.Vector3 {
	L := rl.Vector2Distance(a, b)
	n := int(L/4) + 1
	out := []rl.Vector3{}
	y := y0
	for i := 0; i <= n; i++ {
		t := float32(i) / float32(n)
		p := rl.Vector2Lerp(a, b, t)
		gy, _ := w.GroundAt(p.X, p.Y, y+1.5)
		y = gy
		out = append(out, v3(p.X, gy, p.Y))
	}
	return out
}

func buildNav(cw *CityWorld) {
	w := cw.W
	ng := &NavGraph{grid: map[int64][]int32{}}
	for _, r := range cw.Routes {
		ng.polyline(r.Pts, true)
	}
	// podurile noi (fara trafic): bratul insulei -> pod -> cabina lumii
	for _, bi := range []int{BridgeWestIdx, BridgeEastIdx} {
		b := &w.Bridges[bi]
		pts := []rl.Vector3{}
		for t := float32(0); t <= 1.001; t += 0.025 {
			x, z := b.Point(t, 0)
			pts = append(pts, v3(x, b.DeckY(t), z))
		}
		ng.polyline(pts, false)
	}
	sw := Sectors[2].Center
	ne := Sectors[1].Center
	ng.polyline(segPts(w, rl.NewVector2(sw.X-10, sw.Y), WestBridgeA, 0), false)
	ng.polyline(segPts(w, rl.NewVector2(ne.X+10, ne.Y), EastBridgeA, 0), false)
	// scarile Marelui Tunel (piata <-> tunel) + de la gura scarii pana la soseaua din jurul pietei
	for i := 1; i < len(w.Corridors); i++ {
		c := &w.Corridors[i]
		pts := []rl.Vector3{}
		for s := float32(0); s <= c.Len; s += 2 {
			x, z := c.Point(s, 0)
			pts = append(pts, v3(x, c.FloorC(s), z))
		}
		ng.polyline(pts, false)
		for _, s := range []float32{0, c.Len} {
			x, z := c.Point(s, 0)
			if c.FloorC(s) > -1 {
				d := rl.Vector2Normalize(rl.NewVector2(x, z))
				ng.polyline(segPts(w, rl.NewVector2(x, z), rl.Vector2Scale(d, 50), 0), false)
			}
		}
	}
	// pe jos: piata (inel), plaja (de la pod la giratia Marelui Tunel), parcul (aleile roz)
	ring := []rl.Vector3{}
	for k := 0; k < 48; k++ {
		a := float32(k) / 48 * 2 * math.Pi
		ring = append(ring, v3(cosf(a)*46, 0.05, sinf(a)*46))
	}
	ng.polyline(ring, true)
	walk := func(pts ...rl.Vector2) {
		for i := 0; i+1 < len(pts); i++ {
			ng.polyline(segPts(w, pts[i], pts[i+1], 0), false)
		}
	}
	walk(WestBridgeB, rl.NewVector2(-322, 47), rl.NewVector2(-392, 47), rl.NewVector2(-401, 10))
	walk(EastBridgeB, rl.NewVector2(330, -74), rl.NewVector2(360, -74), rl.NewVector2(360, -30), rl.NewVector2(401, -30), rl.NewVector2(401, -10))
	// cabinele: legate de cel mai apropiat punct de drum
	for i := range cw.Booths {
		b := &cw.Booths[i]
		d := b.DoorPos(2.5)
		ni := ng.add(v3(d.X, b.Pos.Y, d.Z))
		ng.joinNear(ni, 40, 3)
	}
	// lipim traseele care se ating (benzi opuse, intersectii)
	for i := range ng.N {
		ng.joinNear(int32(i), 5.5, 1.6)
	}
	nav = ng
}

// joinNear: leaga nodul de nodurile apropiate (pe orizontala < r, pe verticala < dy).
// Pentru r mare leaga doar cel mai apropiat nod.
func (ng *NavGraph) joinNear(i int32, r, dy float32) {
	p := ng.N[i].P
	best, bd := int32(-1), r*r
	kx := int64(math.Floor(float64(p.X / 8)))
	kz := int64(math.Floor(float64(p.Z / 8)))
	span := int64(r/8) + 1
	for dx := -span; dx <= span; dx++ {
		for dz := -span; dz <= span; dz++ {
			for _, j := range ng.grid[(kx+dx)*100003+kz+dz] {
				if j == i {
					continue
				}
				q := ng.N[j].P
				if absf(q.Y-p.Y) > dy {
					continue
				}
				d := sq(q.X-p.X) + sq(q.Z-p.Z)
				if r <= 6 {
					if d < r*r {
						ng.link(i, j)
					}
				} else if d < bd {
					bd, best = d, j
				}
			}
		}
	}
	if r > 6 && best >= 0 {
		ng.link(i, best)
	}
}

func (ng *NavGraph) nearest(p rl.Vector3, ok func(int32) bool) int32 {
	best, bd := int32(-1), float32(1e12)
	// cautam pe o raza fixa; diferenta de inaltime costa mult (sa nu alegem
	// tunelul de sub picioare cand Aventurierul e sus, in piata)
	for r := int64(5); r <= 30 && best < 0; r += 25 {
		kx := int64(math.Floor(float64(p.X / 8)))
		kz := int64(math.Floor(float64(p.Z / 8)))
		for dx := -r; dx <= r; dx++ {
			for dz := -r; dz <= r; dz++ {
				for _, j := range ng.grid[(kx+dx)*100003+kz+dz] {
					if !ok(j) {
						continue
					}
					q := ng.N[j].P
					d := sq(q.X-p.X) + sq(q.Z-p.Z) + sq(q.Y-p.Y)*60
					if d < bd {
						bd, best = d, j
					}
				}
			}
		}
	}
	return best
}

type pqItem struct {
	n int32
	f float32
}
type pq []pqItem

func (q pq) Len() int            { return len(q) }
func (q pq) Less(i, j int) bool  { return q[i].f < q[j].f }
func (q pq) Swap(i, j int)       { q[i], q[j] = q[j], q[i] }
func (q *pq) Push(x interface{}) { *q = append(*q, x.(pqItem)) }
func (q *pq) Pop() interface{} {
	o := *q
	it := o[len(o)-1]
	*q = o[:len(o)-1]
	return it
}

// path: drumul pe graf de la a la b (A*), doar prin noduri permise.
func (ng *NavGraph) path(a, b int32, ok func(int32) bool) []rl.Vector3 {
	n := len(ng.N)
	g := make([]float32, n)
	from := make([]int32, n)
	for i := range g {
		g[i] = 1e12
		from[i] = -1
	}
	g[a] = 0
	goal := ng.N[b].P
	h := &pq{{a, rl.Vector3Distance(ng.N[a].P, goal)}}
	for h.Len() > 0 {
		it := heap.Pop(h).(pqItem)
		if it.n == b {
			break
		}
		for _, j := range ng.N[it.n].Next {
			if !ok(j) {
				continue
			}
			c := g[it.n] + rl.Vector3Distance(ng.N[it.n].P, ng.N[j].P)
			if c < g[j] {
				g[j] = c
				from[j] = it.n
				heap.Push(h, pqItem{j, c + rl.Vector3Distance(ng.N[j].P, goal)})
			}
		}
	}
	if from[b] < 0 && a != b {
		return nil
	}
	out := []rl.Vector3{}
	for x := b; x >= 0; x = from[x] {
		out = append(out, ng.N[x].P)
		if x == a {
			break
		}
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// ---------------------------------------------------------------

func (g *Game) startGPS(target rl.Vector3, wi int) {
	gp := &g.Prog.GPS
	*gp = GPSState{Active: true, World: wi, Target: target}
	gp.FromInside = wi >= 0 && g.currentWorld() == wi
	g.repath()
}

func (g *Game) startGPSToWorld(wi int) {
	if wi < 0 || wi >= len(g.City.Booths) {
		return
	}
	b := &g.City.Booths[wi]
	g.startGPS(b.DoorPos(2.5), wi)
}

func (g *Game) stopGPS() {
	g.Prog.GPS = GPSState{World: -1}
}

func (g *Game) repath() {
	gp := &g.Prog.GPS
	gp.RepathT = 0.8
	if nav == nil {
		return
	}
	p := g.Player.Pos
	if g.Drive.Active {
		p = g.City.Cars[g.Drive.Car].Pos
	}
	ok := func(j int32) bool {
		q := nav.N[j].P
		return !g.lockedAt(q.X, q.Z, q.Y)
	}
	a := nav.nearest(p, ok)
	b := nav.nearest(gp.Target, ok)
	if a < 0 || b < 0 {
		gp.Path = []rl.Vector3{p, gp.Target}
		return
	}
	path := nav.path(a, b, ok)
	if path == nil {
		if g.Debug != nil {
			fmt.Printf("gps: fara drum intre nodul %d (%v) si %d (%v)\n", a, nav.N[a].P, b, nav.N[b].P)
		}
		gp.Path = []rl.Vector3{p, gp.Target}
		return
	}
	// scurtatura: sarim peste primele noduri daca sunt in spatele aventurierului
	for len(path) > 2 && rl.Vector3Distance(p, path[1]) < rl.Vector3Distance(path[0], path[1]) {
		path = path[1:]
	}
	gp.Path = append([]rl.Vector3{p}, path...)
	gp.Path = append(gp.Path, gp.Target)
}

func (g *Game) updateGPS(dt float32) {
	gp := &g.Prog.GPS
	if !gp.Active {
		return
	}
	gp.Flow += dt
	gp.RepathT -= dt
	if gp.RepathT <= 0 {
		g.repath()
	}
	p := g.Player.Pos
	if rl.Vector3Distance(p, gp.Target) < 4.5 {
		g.stopGPS()
		g.sayIcon(IconCheck, "Ai ajuns!")
		g.Play(g.Audio.Coin)
	}
}

// drawGPS: sageti verzi pe jos, pe urmatorii ~50 m ai drumului.
func (g *Game) drawGPS() {
	gp := &g.Prog.GPS
	if !gp.Active || len(gp.Path) < 2 || g.Scene != SceneMapExplore {
		return
	}
	rl.DisableDepthMask()
	rl.DisableBackfaceCulling()
	spacing := float32(3.2)
	off := float32(math.Mod(float64(gp.Flow*4), float64(spacing)))
	travel := float32(0)
	maxD := float32(55)
	k := 0
	for i := 0; i+1 < len(gp.Path) && travel < maxD; i++ {
		a, b := gp.Path[i], gp.Path[i+1]
		seg := rl.Vector3Distance(a, b)
		if seg < 0.01 {
			continue
		}
		d := rl.Vector3Normalize(rl.Vector3Subtract(b, a))
		for s := off; s < seg && travel+s < maxD; s += spacing {
			c := rl.Vector3Add(a, rl.Vector3Scale(d, s))
			gy, _ := g.City.W.GroundAt(c.X, c.Z, c.Y+1.2)
			if absf(gy-c.Y) < 2.5 {
				c.Y = gy
			}
			fadeIn := clampf((travel+s)/4, 0, 1) * clampf((maxD-travel-s)/12, 0, 1)
			pulse := 0.75 + 0.25*sinf(gp.Flow*6-float32(k)*0.6)
			col := rl.NewColor(40, 235, 80, uint8(255*fadeIn*pulse))
			drawChevron(c, d, 2.3, col)
			k++
		}
		travel += seg
		off = float32(math.Mod(float64(off-seg), float64(spacing)))
		if off < 0 {
			off += spacing
		}
	}
	rl.EnableBackfaceCulling()
	rl.EnableDepthMask()
}

func drawChevron(c, d rl.Vector3, s float32, col rl.Color) {
	f := rl.Vector3Normalize(v3(d.X, 0, d.Z))
	r := v3(-f.Z, 0, f.X)
	y := c.Y + 0.18
	at := func(fw, side float32) rl.Vector3 {
		return v3(c.X+f.X*fw+r.X*side, y, c.Z+f.Z*fw+r.Z*side)
	}
	tip := at(s*0.55, 0)
	inner := at(s*0.05, 0)
	for _, side := range []float32{-1, 1} {
		o := at(-s*0.45, side*s*0.75)
		oi := at(-s*0.95, side*s*0.75)
		rl.DrawTriangle3D(tip, o, oi, col)
		rl.DrawTriangle3D(tip, oi, inner, col)
	}
}
