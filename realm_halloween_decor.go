package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() { registerRealmDecor("halloween", halloweenDecor) }

func halloweenDecor(rb *RealmBuilder) {
	halloweenDecor2(rb)
	c := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	orange := c(231, 105, 24)
	gold := c(255, 199, 55)
	wood := c(91, 54, 47)
	dark := c(45, 35, 54)
	green := c(85, 236, 82)
	straw := c(213, 169, 74)
	pumpkin := func(m *MeshBuilder, x, y, z, s float32) {
		m.Ellipsoid(v3(x, y+s*.75, z), v3(s, s*.75, s*.8), 6, 10, orange)
		m.Cylinder(v3(x, y+s*1.4, z), v3(x+.1*s, y+s*1.75, z), s*.13, s*.09, 6, wood, true)
		m.Emis = 1
		for _, dx := range []float32{-.37, .37} {
			m.Ellipsoid(v3(x+dx*s, y+s*.85, z+s*.76), v3(s*.16, s*.19, s*.06), 4, 6, gold)
		}
		m.Box(v3(x, y+s*.48, z+s*.78), v3(s*.65, s*.1, s*.08), gold)
		m.Emis = 0
	}
	house := func(m *MeshBuilder, x, z, w, h float32, col rl.Color) {
		y := rb.H(x, z)
		d := w * .8
		m.Box(v3(x, y+h/2, z), v3(w, h, d), col)
		a := v3(x-w*.6, y+h, z-d*.6)
		b := v3(x+w*.6, y+h, z-d*.6)
		p := v3(x+w*.12, y+h+w*.8, z)
		q := v3(x-w*.6, y+h, z+d*.6)
		t := v3(x+w*.6, y+h, z+d*.6)
		m.Tri(a, b, p, v3(0, 0, -1), dark)
		m.Tri(q, t, p, v3(0, 0, 1), dark)
		m.Tri(a, q, p, v3(-1, 1, 0), dark)
		m.Tri(b, t, p, v3(1, 1, 0), dark)
		m.Box(v3(x, y+2, z+d/2+.08), v3(w*.22, 4, .2), wood)
		m.Emis = 1
		for _, dx := range []float32{-.33, .33} {
			m.Box(v3(x+w*dx, y+h*.55, z+d/2+.1), v3(w*.18, h*.22, .2), gold)
		}
		// Warm windows on every side, with friendly cobwebs at the corners.
		for _, side := range []float32{-1, 1} {
			for _, dx := range []float32{-.3, .3} {
				m.Box(v3(x+w*dx, y+h*.55, z+side*(d/2+.12)), v3(w*.18, h*.22, .2), gold)
				m.Box(v3(x+side*(w/2+.12), y+h*.55, z+d*dx), v3(.2, h*.22, d*.18), gold)
			}
		}
		m.Emis = 0
		for _, side := range []float32{-1, 1} {
			for j := 0; j < 4; j++ {
				m.Cylinder(v3(x+side*w*.48, y+h*.95, z+d*.51), v3(x+side*(w*.48-float32(j)*w*.08), y+h*(.55+float32(j)*.1), z+d*.51), .035, .035, 4, c(240, 235, 220), false)
			}
			pumpkin(m, x+side*w*.28, y, z+d/2+1, .55)
		}
		rb.SolidRoof(x, z, w, d, h, y, RoofRidgeX, w*.65)
	}
	tree := func(m *MeshBuilder, x, z float32) {
		y := rb.H(x, z)
		m.Cylinder(v3(x, y, z), v3(x+1, y+9, z), .8, .42, 7, wood, true)
		for i := -1; i <= 1; i++ {
			f := float32(i)
			m.Cylinder(v3(x+.6, y+5, z), v3(x+f*3, y+10+f, z+f), .25, .08, 6, wood, false)
		}
		rb.Pillar(x, z, .8, y+9)
	}
	scare := func(m *MeshBuilder, x, z, s float32) {
		y := rb.H(x, z)
		m.Cylinder(v3(x, y, z), v3(x, y+5*s, z), .15*s, .12*s, 6, wood, true)
		m.Box(v3(x, y+3*s, z), v3(3*s, .25*s, .25*s), wood)
		m.Box(v3(x, y+3*s, z), v3(1.5*s, 1.8*s, .6*s), c(133, 93, 164))
		m.Ellipsoid(v3(x, y+4.4*s, z), v3(.65*s, .7*s, .5*s), 6, 8, straw)
		m.Cylinder(v3(x, y+5*s, z), v3(x, y+5.8*s, z), .9*s, .08*s, 8, dark, true)
		rb.Pillar(x, z, .5*s, y+5*s)
	}
	cauldron := func(m *MeshBuilder, x, z, s float32) {
		y := rb.H(x, z)
		m.Ellipsoid(v3(x, y+s*.65, z), v3(s, s*.65, s), 8, 12, dark)
		m.Emis = 1
		m.Cylinder(v3(x, y+s*1.05, z), v3(x, y+s*1.09, z), s*.78, s*.78, 12, green, true)
		for i := 0; i < 5; i++ {
			f := float32(i)
			m.Ellipsoid(v3(x+(f-2)*s*.28, y+s*(1.2+float32(i%2)*.3), z), v3(s*.15, s*.15, s*.15), 5, 6, green)
		}
		m.Emis = 0
		rb.Pillar(x, z, s, y+s*1.1)
	}
	// Landmarks are sought around the plaza so roads and reserved sites stay clear.
	for zone := 0; zone < 3; zone++ {
		zn := rb.Zones[zone]
		m := rb.Mesh()
		placed := false
		for i := 0; i < 32 && !placed; i++ {
			a := float32(i)*.1963495 + .15
			x := zn.C.X + cosf(a)*43
			z := zn.C.Y + sinf(a)*43
			r := float32(10)
			if !rb.Free(x, z, r) {
				continue
			}
			y := rb.H(x, z)
			switch zone {
			case 0:
				pumpkin(m, x, y, z, 9)
				m.Emis = 1
				m.Cylinder(v3(x, y+1, z), v3(x, y+1.15, z), 10, 10, 20, gold, true)
				m.Emis = 0
				rb.Pillar(x, z, 8, y+14)
			case 1:
				scare(m, x, z, 3.8)
				m.Emis = 1
				m.Box(v3(x, y+11.4, z+.4), v3(5, .2, .2), gold)
				m.Emis = 0
			case 2:
				house(m, x, z, 14, 13, c(118, 68, 143))
				m.Cylinder(v3(x+5, y, z-3), v3(x+5, y+21, z-3), 3, 2.5, 10, c(105, 65, 127), true)
				m.Cylinder(v3(x+5, y+21, z-3), v3(x+6, y+29, z-3), 3.5, 0, 10, dark, true)
				m.Emis = 1
				m.Ellipsoid(v3(x+5, y+17, z+.05), v3(1.5, 1.5, .15), 8, 12, gold)
				m.Emis = 0
				// Giant broom leaning beside the witch's tower.
				m.Cylinder(v3(x-8, y+1, z+5), v3(x-5, y+15, z+5), .3, .22, 7, wood, true)
				for j := -2; j <= 2; j++ {
					m.Cylinder(v3(x-8, y+1, z+5), v3(x-9+float32(j)*.5, y+.1, z+5), .14, .07, 5, straw, true)
				}
			}
			rb.Reserve(x, z, r)
			placed = true
		}
		for i := 0; i < 48; i++ {
			a := float32(i) * .392699
			radius := float32(33 + (i%3)*11)
			x := zn.C.X + cosf(a)*radius
			z := zn.C.Y + sinf(a)*radius
			if !rb.Free(x, z, 5) {
				continue
			}
			switch {
			case zone == 0 && i%3 == 0:
				house(m, x, z, 7, 7, []rl.Color{c(144, 74, 155), c(191, 91, 45), c(56, 105, 68)}[(i/3)%3])
				rb.Reserve(x, z, 5)
			case zone == 1 && i%8 == 0:
				scare(m, x, z, 1)
				rb.Reserve(x, z, 2)
			case zone == 1 && i%9 == 0:
				house(m, x, z, 9, 7, c(157, 70, 54))
				rb.Reserve(x, z, 5)
			case zone == 2 && i%7 == 0:
				cauldron(m, x, z, 2.5)
				rb.Reserve(x, z, 3)
			case i%4 == 0:
				tree(m, x, z)
				rb.Reserve(x, z, 2)
			default:
				pumpkin(m, x, rb.H(x, z), z, 1.1+float32(i%3)*.35)
				rb.Reserve(x, z, 1.5)
			}
		}
		rb.Add(m, true)
	}
	// Small illuminated details keep the fighting area open.
	for zone := 0; zone < 3; zone++ {
		zn := rb.Zones[zone]
		m := rb.Mesh()
		for i := 0; i < 80; i++ {
			a := float32(i) * 2.39996
			r := float32(19) + float32(i%14)*2.4
			x, z := zn.C.X+cosf(a)*r, zn.C.Y+sinf(a)*r
			if !rb.Free(x, z, .5) {
				continue
			}
			y := rb.H(x, z)
			switch zone {
			case 0:
				m.Cylinder(v3(x, y, z), v3(x, y+2.7, z), .08, .08, 5, wood, false)
				m.Emis = 1
				m.Box(v3(x, y+2.8, z), v3(.45, .6, .45), gold)
				m.Emis = 0
			case 1:
				pumpkin(m, x, y, z, .45+float32(i%4)*.18)
				if i%7 == 0 {
					m.Box(v3(x+1, y+.35, z), v3(1.4, .7, .8), straw)
				}
			case 2:
				m.Cylinder(v3(x, y, z), v3(x, y+.8, z), .12, .1, 6, straw, true)
				m.Emis = 1
				m.Ellipsoid(v3(x, y+.9, z), v3(.5, .25, .5), 5, 8, []rl.Color{green, orange, c(222, 112, 240)}[i%3])
				m.Emis = 0
			}
			rb.Reserve(x, z, .5)
		}
		rb.Add(m, true)
	}
	// Clusters on the field and scattered decorations between the settlements.
	for gx := -3; gx <= 3; gx++ {
		for gz := -3; gz <= 3; gz++ {
			m := rb.Mesh()
			for i := 0; i < 35; i++ {
				x := float32(gx*80) + rb.Rng.Float32()*78 - 39
				z := float32(gz*80) + rb.Rng.Float32()*78 - 39
				if !rb.Free(x, z, 1.5) {
					continue
				}
				y := rb.H(x, z)
				if rb.ZoneAt(x, z) == 1 || i%3 != 0 {
					pumpkin(m, x, y, z, .45+rb.Rng.Float32()*.75)
				} else {
					tree(m, x, z)
				}
				rb.Reserve(x, z, 1.5)
			}
			rb.Add(m, true)
		}
	}
}
