package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Poduri reale: tablier cu sosea si trotuare, parapeti, felinare, pile in apa
// si structura specifica fiecarui tip (suspendat, piatra cu arce, hobanat, arc de otel, grinda).

var (
	ColConcrete  = rl.NewColor(196, 196, 190, 255)
	ColConcrete2 = rl.NewColor(170, 170, 166, 255)
	ColSteelRed  = rl.NewColor(214, 74, 50, 255)
	ColStone     = rl.NewColor(214, 186, 144, 255)
	ColStoneDark = rl.NewColor(176, 146, 108, 255)
	ColCable     = rl.NewColor(232, 232, 228, 255)
)

func (cb *CityBuilder) bridgePoint(b *Bridge, t, lat, dy float32) rl.Vector3 {
	x, z := b.Point(t, lat)
	return v3(x, b.DeckY(t)+dy, z)
}

// tablierul comun tuturor podurilor
func (cb *CityBuilder) bridgeDeck(mb *MeshBuilder, b *Bridge, fascia rl.Color, solidParapet bool) {
	steps := int(b.Length / 2)
	if steps < 8 {
		steps = 8
	}
	hw := b.Width / 2
	dir3 := v3(b.Dir.X, 0, b.Dir.Y)
	per3 := v3(b.Perp.X, 0, b.Perp.Y)
	for i := 0; i < steps; i++ {
		t0 := float32(i) / float32(steps)
		t1 := float32(i+1) / float32(steps)
		P := func(t, lat, dy float32) rl.Vector3 { return cb.bridgePoint(b, t, lat, dy) }
		up := v3(0, 1, 0)
		// sosea
		mb.Quad(P(t0, -4.5, 0.05), P(t1, -4.5, 0.05), P(t1, 4.5, 0.05), P(t0, 4.5, 0.05), up, ColAsphalt)
		if i%3 == 0 {
			mb.Quad(P(t0, -0.12, 0.07), P(t1, -0.12, 0.07), P(t1, 0.12, 0.07), P(t0, 0.12, 0.07), up, ColStripe)
		}
		for _, s := range []float32{-1, 1} {
			mb.Quad(P(t0, s*4.1, 0.065), P(t1, s*4.1, 0.065), P(t1, s*4.25, 0.065), P(t0, s*4.25, 0.065), up, ColStripe)
			// trotuar
			mb.Quad(P(t0, s*4.5, 0.25), P(t1, s*4.5, 0.25), P(t1, s*hw, 0.25), P(t0, s*hw, 0.25), up, ColSidewalk)
			mb.Quad(P(t0, s*4.5, 0.05), P(t1, s*4.5, 0.05), P(t1, s*4.5, 0.25), P(t0, s*4.5, 0.25), rl.Vector3Scale(per3, -s), ColCurb)
			// fata laterala a tablierului
			mb.Quad(P(t0, s*hw, -1.2), P(t1, s*hw, -1.2), P(t1, s*hw, 0.25), P(t0, s*hw, 0.25), rl.Vector3Scale(per3, s), fascia)
			// parapet
			if solidParapet {
				in := s * (hw - 0.5)
				mb.Quad(P(t0, in, 0.25), P(t1, in, 0.25), P(t1, in, 1.25), P(t0, in, 1.25), rl.Vector3Scale(per3, -s), ColStone)
				mb.Quad(P(t0, s*hw, 0.25), P(t1, s*hw, 0.25), P(t1, s*hw, 1.25), P(t0, s*hw, 1.25), rl.Vector3Scale(per3, s), ColStone)
				mb.Quad(P(t0, in, 1.25), P(t1, in, 1.25), P(t1, s*(hw+0.1), 1.35), P(t0, s*(hw+0.1), 1.35), up, ColStoneDark)
			} else {
				lat := s * (hw - 0.15)
				a := P(t0, lat, 1.15)
				c := P(t1, lat, 1.15)
				mb.Cylinder(a, c, 0.07, 0.07, 5, ColLampPole, false)
				mb.Cylinder(P(t0, lat, 0.65), P(t1, lat, 0.65), 0.04, 0.04, 4, ColLampPole, false)
				mb.Box(P(t0, lat, 0.7), v3(0.12, 0.9, 0.12), ColLampPole)
			}
		}
		// dedesubt
		mb.Quad(P(t0, -hw, -1.2), P(t1, -hw, -1.2), P(t1, hw, -1.2), P(t0, hw, -1.2), v3(0, -1, 0), shade(ColConcrete2, 0.8))
	}
	// capete (culee) pana in sol
	for _, t := range []float32{0, 1} {
		s := float32(1)
		if t == 0 {
			s = -1
		}
		c := b.A
		if t == 1 {
			c = b.B
		}
		g := cb.W.LandHeight(c.X, c.Y)
		y0 := minf(g, -1.5)
		_ = y0
		p := func(lat, y float32) rl.Vector3 { x, z := b.Point(t, lat); return v3(x, y, z) }
		mb.Quad(p(-hw, -1.4), p(hw, -1.4), p(hw, 0.05), p(-hw, 0.05), rl.Vector3Scale(dir3, s), ColConcrete2)
	}
	// felinare pe pod
	k := 0
	for d := float32(12); d < b.Length-8; d += 24 {
		t := d / b.Length
		s := float32(1)
		if k%2 == 1 {
			s = -1
		}
		k++
		p := cb.bridgePoint(b, t, s*(hw-0.45), 0.25)
		cb.streetLamp(mb, p.X, p.Y, p.Z, -b.Perp.X*s, -b.Perp.Y*s)
	}
}

// pila (stalp in apa): cutie rotunjita cu "pinteni" care despica apa
func (cb *CityBuilder) pier(mb *MeshBuilder, b *Bridge, t float32, alongW, acrossW, topY float32, col rl.Color, cutwater bool) {
	x, z := b.Point(t, 0)
	bot := SeabedY - 0.5
	yaw := atan2f(b.Dir.X, b.Dir.Y) // axa locala z = directia podului
	mb.SetMatrix(matTRS(v3(x, 0, z), yaw, 1))
	h := topY - bot
	mb.BeveledBox(v3(0, bot+h/2, 0), v3(acrossW, h, alongW), 0.25, col)
	if cutwater {
		for _, s := range []float32{-1, 1} {
			x0 := s * acrossW / 2
			x1 := s * (acrossW/2 + alongW*0.7)
			top := minf(topY, WaterY+2.2)
			a := v3(x0, bot, -alongW/2)
			bb := v3(x0, bot, alongW/2)
			c := v3(x1, bot, 0)
			a2 := v3(x0, top, -alongW/2)
			b2 := v3(x0, top, alongW/2)
			c2 := v3(x1, top, 0)
			mb.Quad(a, c, c2, a2, v3(s, 0, -1), shade(col, 0.95))
			mb.Quad(bb, c, c2, b2, v3(s, 0, 1), shade(col, 0.95))
			mb.Tri(a2, b2, c2, v3(0, 1, 0), shade(col, 1.05))
		}
	}
	mb.ClearMatrix()
	// pila e solida (pentru inot): un sir de cercuri pe latime
	r := alongW / 2
	n := int(acrossW/alongW) + 1
	for k := 0; k < n; k++ {
		off := float32(0)
		if n > 1 {
			off = -acrossW/2 + r + (acrossW-2*r)*float32(k)/float32(n-1)
		}
		px, pz := b.Point(t, off)
		cb.W.Circles = append(cb.W.Circles, CircleCollider{X: px, Z: pz, R: r + 0.2, Top: topY})
	}
}

// ---------------- POD SUSPENDAT ----------------

func (cb *CityBuilder) suspensionBridge(b *Bridge) {
	mb := NewMB()
	cb.bridgeDeck(mb, b, ColSteelRed, false)
	hw := b.Width / 2
	legLat := hw + 1.4
	topY := float32(44)
	towersT := []float32{0.25, 0.75}
	for _, tt := range towersT {
		// fundatie
		x, z := b.Point(tt, 0)
		cb.pier(mb, b, tt, 6, b.Width+7, WaterY+0.8, ColConcrete, true)
		for _, s := range []float32{-1, 1} {
			bx, bz := b.Point(tt, s*legLat)
			tx, tz := b.Point(tt, s*(legLat-0.6))
			mb.CylinderC(v3(bx, WaterY+0.8, bz), v3(tx, topY, tz), 1.35, 1.0, 4, ColSteelRed, shade(ColSteelRed, 1.08), true)
		}
		// grinzi transversale (portaluri)
		for _, y := range []float32{b.DeckY(tt) - 1.6, 24, 34, topY - 0.8} {
			l := legLat - 0.6*(y-WaterY)/(topY-WaterY)
			p0 := cb.bridgePoint(b, tt, -l, 0)
			p1 := cb.bridgePoint(b, tt, l, 0)
			p0.Y, p1.Y = y, y
			mb.Cylinder(p0, p1, 0.9, 0.9, 4, shade(ColSteelRed, 0.95), true)
		}
		cb.W.Circles = append(cb.W.Circles,
			CircleCollider{X: x + b.Perp.X*legLat, Z: z + b.Perp.Y*legLat, R: 1.4, Top: topY},
			CircleCollider{X: x - b.Perp.X*legLat, Z: z - b.Perp.Y*legLat, R: 1.4, Top: topY})
	}
	// cablurile principale (parabola intre turnuri, aproape drepte spre ancoraje)
	cableY := func(t float32) float32 {
		yLow := b.DeckY(0.5) + 2.2
		switch {
		case t < 0.25:
			s := t / 0.25
			return lerpf(2.6, topY, s) - 3.0*sinf(math.Pi*s)
		case t > 0.75:
			s := (1 - t) / 0.25
			return lerpf(2.6, topY, s) - 3.0*sinf(math.Pi*s)
		}
		u := (t - 0.5) / 0.25
		return yLow + (topY-yLow)*u*u
	}
	cableLat := func(t float32) float32 {
		switch {
		case t < 0.25:
			return lerpf(hw+3.2, legLat-0.6, t/0.25)
		case t > 0.75:
			return lerpf(hw+3.2, legLat-0.6, (1-t)/0.25)
		}
		return legLat - 0.6
	}
	for _, s := range []float32{-1, 1} {
		n := 120
		prev := rl.Vector3{}
		for i := 0; i <= n; i++ {
			t := float32(i) / float32(n)
			x, z := b.Point(t, s*cableLat(t))
			p := v3(x, cableY(t), z)
			if i > 0 {
				mb.Cylinder(prev, p, 0.42, 0.42, 6, ColSteelRed, false)
			}
			prev = p
		}
		// tiranti verticali
		for d := float32(4); d < b.Length-2; d += 4 {
			t := d / b.Length
			if absf(t-0.25) < 0.012 || absf(t-0.75) < 0.012 {
				continue
			}
			lat := s * cableLat(t)
			x, z := b.Point(t, lat)
			dy := b.DeckY(t)
			cy := cableY(t)
			if cy-dy < 1.2 {
				continue
			}
			x2, z2 := b.Point(t, s*(hw-0.05))
			mb.Cylinder(v3(x2, dy+0.3, z2), v3(x, cy, z), 0.06, 0.06, 4, ColSteelRed, false)
		}
		// ancoraje masive la capete
		for _, t := range []float32{0, 1} {
			x, z := b.Point(t, s*(hw+3.2))
			mb.BeveledBox(v3(x, 1.2, z), v3(4, 3.4, 4), 0.3, ColConcrete)
			cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 2.4, Top: 4})
		}
	}
	cb.addChunk(mb, true)
}

// ---------------- POD DE PIATRA CU ARCE ----------------

func (cb *CityBuilder) stoneArchBridge(b *Bridge) {
	mb := NewMB()
	cb.bridgeDeck(mb, b, ColStone, true)
	hw := b.Width / 2
	L := b.Length
	t0, t1 := float32(0.12), float32(0.88)
	nSpans := 6
	pierW := float32(4.4)
	pierT := make([]float32, nSpans+1)
	for k := 0; k <= nSpans; k++ {
		pierT[k] = t0 + (t1-t0)*float32(k)/float32(nSpans)
	}
	springY := WaterY - 0.35
	type arch struct {
		a, bb, m, hs, rise float32
	}
	arches := []arch{}
	for k := 0; k < nSpans; k++ {
		a := pierT[k]*L + pierW/2
		bb := pierT[k+1]*L - pierW/2
		m := (a + bb) / 2
		crown := b.DeckY(m/L) - 1.9
		rise := crown - springY
		if rise < 2.0 {
			continue
		}
		arches = append(arches, arch{a, bb, m, (bb - a) / 2, rise})
	}
	intrados := func(d float32) (float32, bool) {
		for _, ar := range arches {
			if d > ar.a && d < ar.bb {
				u := (d - ar.m) / ar.hs
				return springY + ar.rise*sqrtf(maxf(1-u*u, 0)), true
			}
		}
		return 0, false
	}
	ground := func(d float32) float32 {
		x, z := b.Point(d/L, 0)
		return minf(cb.W.LandHeight(x, z), -0.3)
	}
	stoneCol := func(d, y float32, ring bool) rl.Color {
		ix := int32(math.Floor(float64(d / 1.6)))
		iy := int32(math.Floor(float64(y / 0.8)))
		if iy&1 == 1 {
			ix = int32(math.Floor(float64((d + 0.8) / 1.6)))
		}
		v := hash2(ix, iy)
		c := shade(ColStone, 0.88+0.2*v)
		if ring {
			c = shade(ColStoneDark, 0.92+0.18*float32((ix)&1))
		}
		return c
	}
	per3 := v3(b.Perp.X, 0, b.Perp.Y)
	step := float32(1.0)
	for d := float32(0); d < L-0.001; d += step {
		d1 := minf(d+step, L)
		dm := (d + d1) / 2
		top0 := b.DeckY(d/L) - 1.2
		top1 := b.DeckY(d1/L) - 1.2
		ib0, in0 := intrados(d)
		ib1, in1 := intrados(d1)
		_, inm := intrados(dm)
		bot0, bot1 := ground(d), ground(d1)
		if in0 && inm {
			bot0 = ib0
		}
		if in1 && inm {
			bot1 = ib1
		}
		if !inm {
			bot0, bot1 = ground(d), ground(d1)
		}
		for _, s := range []float32{-1, 1} {
			// facem fatada in benzi de 0.8 m (blocuri de piatra)
			ylo := minf(bot0, bot1)
			for y := float32(math.Floor(float64(ylo/0.8))) * 0.8; y < maxf(top0, top1); y += 0.8 {
				ya0 := clampf(y, bot0, top0)
				yb0 := clampf(y+0.8, bot0, top0)
				ya1 := clampf(y, bot1, top1)
				yb1 := clampf(y+0.8, bot1, top1)
				if yb0-ya0 < 0.01 && yb1-ya1 < 0.01 {
					continue
				}
				ring := inm && (y-(bot0+bot1)/2) < 1.2
				c := stoneCol(dm, y+0.4, ring)
				p := func(dd, yy float32) rl.Vector3 {
					x, z := b.Point(dd/L, s*hw)
					return v3(x, yy, z)
				}
				mb.Quad(p(d, ya0), p(d1, ya1), p(d1, yb1), p(d, yb0), rl.Vector3Scale(per3, s), c)
			}
		}
		// intradosul (bolta vazuta de dedesubt)
		if inm {
			p := func(dd, lat, yy float32) rl.Vector3 {
				x, z := b.Point(dd/L, lat)
				return v3(x, yy, z)
			}
			mb.Quad(p(d, -hw, bot0), p(d1, -hw, bot1), p(d1, hw, bot1), p(d, hw, bot0), v3(0, -1, 0), shade(ColStoneDark, 0.85))
		}
	}
	// fetele pilelor spre deschiderea arcelor + pintenii
	for _, ar := range arches {
		for _, d := range []float32{ar.a, ar.bb} {
			s := float32(1)
			if d == ar.a {
				s = -1
			}
			p := func(lat, yy float32) rl.Vector3 {
				x, z := b.Point(d/L, lat)
				return v3(x, yy, z)
			}
			n := v3(-b.Dir.X*s, 0, -b.Dir.Y*s)
			mb.Quad(p(-hw, SeabedY-0.5), p(hw, SeabedY-0.5), p(hw, springY+0.05), p(-hw, springY+0.05), n, ColStoneDark)
		}
	}
	for k := 0; k <= nSpans; k++ {
		x, z := b.Point(pierT[k], 0)
		if cb.W.LandHeight(x, z) > WaterY+0.3 {
			continue
		}
		yaw := atan2f(b.Dir.X, b.Dir.Y)
		mb.SetMatrix(matTRS(v3(x, 0, z), yaw, 1))
		for _, s := range []float32{-1, 1} {
			x0 := s * hw
			x1 := s * (hw + 2.6)
			top := WaterY + 2.0
			bot := SeabedY - 0.5
			a := v3(x0, bot, -pierW/2)
			bb := v3(x0, bot, pierW/2)
			c := v3(x1, bot, 0)
			a2 := v3(x0, top, -pierW/2)
			b2 := v3(x0, top, pierW/2)
			c2 := v3(x1, top, 0)
			mb.Quad(a, c, c2, a2, v3(s, 0, -1), ColStone)
			mb.Quad(bb, c, c2, b2, v3(s, 0, 1), ColStone)
			// capac in forma de piramida
			mb.Tri(a2, c2, v3(x0, top+1.5, 0), v3(s, 1, -1), ColStoneDark)
			mb.Tri(b2, c2, v3(x0, top+1.5, 0), v3(s, 1, 1), ColStoneDark)
		}
		mb.ClearMatrix()
	}
	cb.addChunk(mb, true)
}

// ---------------- POD HOBANAT ----------------

func (cb *CityBuilder) cableStayedBridge(b *Bridge) {
	mb := NewMB()
	cb.bridgeDeck(mb, b, LegoBlue, false)
	hw := b.Width / 2
	white := rl.NewColor(240, 242, 245, 255)
	tm := float32(0.5)
	x, z := b.Point(tm, 0)
	deckY := b.DeckY(tm)
	join := deckY + 18
	topY := float32(60)
	cb.pier(mb, b, tm, 8, 26, WaterY+0.6, ColConcrete, true)
	for _, s := range []float32{-1, 1} {
		bx, bz := b.Point(tm, s*10.5)
		jx, jz := b.Point(tm, s*1.4)
		mb.CylinderC(v3(bx, WaterY+0.6, bz), v3(jx, join, jz), 1.3, 0.9, 6, white, white, true)
		cb.W.Circles = append(cb.W.Circles, CircleCollider{X: bx, Z: bz, R: 1.5, Top: join})
	}
	mb.Cylinder(v3(x, join-1, z), v3(x, topY, z), 1.6, 1.0, 8, white, true)
	// grinda sub tablier
	p0 := cb.bridgePoint(b, tm, -8.6, -1.7)
	p1 := cb.bridgePoint(b, tm, 8.6, -1.7)
	mb.Cylinder(p0, p1, 0.8, 0.8, 4, white, true)
	// lumina rosie de avertizare
	mb.Emis = 1
	mb.Sphere(v3(x, topY+0.4, z), 0.5, LegoRed)
	mb.Emis = 0
	// hobanele (evantai)
	for k := 0; k < 9; k++ {
		ay := join + 6 + float32(k)*3.6
		dist := 12 + float32(k)*9.5
		for _, dirS := range []float32{-1, 1} {
			t := tm + dirS*dist/b.Length
			if t < 0.03 || t > 0.97 {
				continue
			}
			for _, s := range []float32{-1, 1} {
				anchor := v3(x+b.Perp.X*s*0.7, ay, z+b.Perp.Y*s*0.7)
				dp := cb.bridgePoint(b, t, s*(hw-0.05), 0.3)
				mb.Cylinder(anchor, dp, 0.09, 0.09, 4, ColCable, false)
			}
		}
	}
	// pile intermediare
	for _, t := range []float32{0.14, 0.86} {
		cb.pier(mb, b, t, 3, 9, b.DeckY(t)-1.2, ColConcrete, true)
	}
	cb.addChunk(mb, true)
}

// ---------------- POD CU ARC DE OTEL ----------------

func (cb *CityBuilder) steelArchBridge(b *Bridge) {
	mb := NewMB()
	cb.bridgeDeck(mb, b, rl.NewColor(60, 70, 90, 255), false)
	hw := b.Width / 2
	ta, tb := float32(0.2), float32(0.8)
	H := float32(27)
	archCol := rl.NewColor(255, 200, 30, 255)
	ribY := func(t float32) float32 {
		u := (t - 0.5) / ((tb - ta) / 2)
		base := lerpf(b.DeckY(ta), b.DeckY(tb), (t-ta)/(tb-ta))
		return base + 0.4 + H*(1-u*u)
	}
	ribLat := func(t float32) float32 {
		u := (t - 0.5) / ((tb - ta) / 2)
		return lerpf(hw+0.8, hw-0.6, 1-u*u)
	}
	n := 60
	for _, s := range []float32{-1, 1} {
		var prev rl.Vector3
		for i := 0; i <= n; i++ {
			t := ta + (tb-ta)*float32(i)/float32(n)
			x, z := b.Point(t, s*ribLat(t))
			p := v3(x, ribY(t), z)
			if i > 0 {
				mb.Cylinder(prev, p, 0.85, 0.85, 8, archCol, false)
			}
			prev = p
		}
		// tiranti
		for d := float32(5); d < b.Length; d += 5 {
			t := d / b.Length
			if t <= ta+0.01 || t >= tb-0.01 {
				continue
			}
			x, z := b.Point(t, s*ribLat(t))
			x2, z2 := b.Point(t, s*(hw-0.05))
			mb.Cylinder(v3(x2, b.DeckY(t)+0.3, z2), v3(x, ribY(t)-0.6, z), 0.08, 0.08, 4, ColCable, false)
		}
	}
	// contravantuiri intre arce, sus (doar unde e loc liber deasupra soselei)
	for i := 1; i < n; i += 4 {
		t := ta + (tb-ta)*float32(i)/float32(n)
		if ribY(t)-b.DeckY(t) < 9 {
			continue
		}
		x0, z0 := b.Point(t, -ribLat(t))
		x1, z1 := b.Point(t, ribLat(t))
		y := ribY(t)
		mb.Cylinder(v3(x0, y, z0), v3(x1, y, z1), 0.45, 0.45, 6, archCol, false)
		t2 := ta + (tb-ta)*float32(i+4)/float32(n)
		if t2 < tb && ribY(t2)-b.DeckY(t2) >= 9 {
			x2, z2 := b.Point(t2, ribLat(t2))
			mb.Cylinder(v3(x0, y, z0), v3(x2, ribY(t2), z2), 0.3, 0.3, 6, archCol, false)
		}
	}
	// blocuri de reazem + pile
	for _, t := range []float32{ta, tb} {
		cb.pier(mb, b, t, 7, b.Width+3, b.DeckY(t)-1.2, ColConcrete, true)
		for _, s := range []float32{-1, 1} {
			x, z := b.Point(t, s*(hw+0.8))
			mb.BeveledBox(v3(x, b.DeckY(t)+0.3, z), v3(2.6, 2.4, 2.6), 0.3, ColConcrete)
			cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 1.2, Top: b.DeckY(t) + 3})
		}
	}
	for _, t := range []float32{0.09, 0.91} {
		cb.pier(mb, b, t, 3, 9, b.DeckY(t)-1.2, ColConcrete, true)
	}
	cb.addChunk(mb, true)
}

// ---------------- POD SCURT PE GRINZI ----------------

func (cb *CityBuilder) beamBridge(b *Bridge) {
	mb := NewMB()
	cb.bridgeDeck(mb, b, ColConcrete, false)
	for _, t := range []float32{0.36, 0.68} {
		for _, s := range []float32{-1, 1} {
			x, z := b.Point(t, s*3.5)
			mb.Cylinder(v3(x, SeabedY-0.5, z), v3(x, b.DeckY(t)-1.2, z), 1.0, 0.9, 12, ColConcrete, false)
		}
		p0 := cb.bridgePoint(b, t, -5.5, -1.6)
		p1 := cb.bridgePoint(b, t, 5.5, -1.6)
		mb.Cylinder(p0, p1, 0.6, 0.6, 4, ColConcrete2, true)
	}
	cb.addChunk(mb, true)
}

func (cb *CityBuilder) buildBridges() {
	for i := range cb.W.Bridges {
		b := &cb.W.Bridges[i]
		switch b.Kind {
		case BridgeSuspension:
			cb.suspensionBridge(b)
		case BridgeStoneArch:
			cb.stoneArchBridge(b)
		case BridgeCableStayed:
			cb.cableStayedBridge(b)
		case BridgeSteelArch:
			cb.steelArchBridge(b)
		default:
			cb.beamBridge(b)
		}
	}
}
