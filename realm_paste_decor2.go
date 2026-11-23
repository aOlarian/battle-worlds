package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() { registerRealmDecor("paste_2", pasteDecor2) }

func pasteDecor2(rb *RealmBuilder) {
	c := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	cream, wood, pink := c(247, 224, 178), c(139, 91, 53), c(246, 135, 174)
	colors := []rl.Color{pink, c(255, 211, 75), c(166, 134, 231), c(107, 200, 219), c(250, 246, 222)}
	// The windmill stands at the outer edge; try several positions if another landmark occupies one.
	for _, p := range [][2]float32{{-220, 180}, {-225, 180}, {-200, 180}, {-195, 190}, {-205, 165}, {-180, 195}, {-215, 190}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 14) {
			continue
		}
		y := rb.H(x, z)
		m := rb.Mesh()
		m.Cylinder(v3(x, y, z), v3(x, y+20, z), 8, 5, 12, cream, true)
		m.Cylinder(v3(x, y+20, z), v3(x, y+24, z), 5.5, .1, 12, c(124, 82, 72), true)
		m.Box(v3(x, y+3, z+7.8), v3(3, 6, .4), wood)
		for _, h := range []float32{8, 14, 19} {
			m.Box(v3(x, y+h, z+6), v3(2.4, 2.4, .5), c(102, 172, 210))
		}
		hub := v3(x, y+17, z+8.5)
		m.Sphere(hub, 1.1, wood)
		for i := 0; i < 4; i++ {
			a := float32(i)*1.570796 + 0.35
			dx, dy := cosf(a), sinf(a)
			end := v3(x+dx*14, y+17+dy*14, z+8.5)
			m.Cylinder(hub, end, .35, .25, 6, wood, true)
			for j := 3; j <= 12; j += 3 {
				q := float32(j)
				px, py := x+dx*q, y+17+dy*q
				m.Cylinder(v3(px-dy*1.5, py+dx*1.5, z+8.5), v3(px+dy*1.5, py-dx*1.5, z+8.5), .12, .12, 5, cream, false)
			}
		}
		for i := 0; i < 3; i++ {
			m.Ellipsoid(v3(x-5+float32(i)*2, y+.8, z+9), v3(.9, 1.1, .7), 8, 10, c(207, 181, 130))
		}
		rb.Pillar(x, z, 8, y+24)
		rb.Reserve(x, z, 14)
		rb.Add(m, true)
		break
	}
	// Giant patterned eggs and a woven basket surround the fourth plaza.
	zn := rb.Zones[4]
	m := rb.Mesh()
	for i := 0; i < 12; i++ {
		a := float32(i)*6.283185/12 + .15
		radius := float32(39 + 6*(i%2))
		x, z := zn.C.X+cosf(a)*radius, zn.C.Y+sinf(a)*radius
		s := (2.2 + float32(i%4)) * .8
		if !rb.Free(x, z, s+1) {
			continue
		}
		y := rb.H(x, z)
		col := colors[i%len(colors)]
		m.Ellipsoid(v3(x, y+s*1.2, z), v3(s*.8, s*1.2, s*.8), 12, 16, col)
		for j := 0; j < 8; j++ {
			t := float32(j) * .785398
			m.Sphere(v3(x+cosf(t)*s*.8, y+s*1.25, z+sinf(t)*s*.8), s*.16, colors[(i+2)%len(colors)])
		}
		rb.Pillar(x, z, s*.8, y+s*2.4)
		rb.Reserve(x, z, s+1)
	}
	for _, p := range [][2]float32{{200, 180}, {195, 190}, {210, 160}, {180, 195}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 8) {
			continue
		}
		y := rb.H(x, z)
		m.Cylinder(v3(x, y, z), v3(x, y+6, z), 6, 7, 16, c(193, 132, 67), true)
		m.Tube(v3(x, y+5.5, z), 6.3, 7.2, .7, 16, wood)
		for j := 0; j < 16; j++ {
			a := float32(j) * 6.283185 / 16
			m.Cylinder(v3(x+cosf(a)*6, y+.3, z+sinf(a)*6), v3(x+cosf(a)*7, y+6, z+sinf(a)*7), .14, .14, 5, cream, false)
		}
		for j := 0; j < 12; j++ {
			a := float32(j) * 3.141593 / 12
			b := float32(j+1) * 3.141593 / 12
			m.Cylinder(v3(x+cosf(a)*6.7, y+6+sinf(a)*7, z), v3(x+cosf(b)*6.7, y+6+sinf(b)*7, z), .5, .5, 8, wood, true)
		}
		m.Ellipsoid(v3(x, y+7, z), v3(2, 3, 2), 10, 14, pink)
		m.Box(v3(x, y+4, z+7), v3(5, .7, .25), pink)
		m.Ellipsoid(v3(x-2, y+4, z+7), v3(1.5, .9, .25), 6, 8, pink)
		m.Ellipsoid(v3(x+2, y+4, z+7), v3(1.5, .9, .25), 6, 8, pink)
		rb.Pillar(x, z, 7, y+7)
		rb.Reserve(x, z, 8)
		break
	}
	rb.Add(m, true)
	// Fields and orchards are chunked spatially, leaving clear routes and arenas.
	for gx := -3; gx <= 3; gx++ {
		for gz := -3; gz <= 3; gz++ {
			mb := rb.Mesh()
			for k := 0; k < 115; k++ {
				x := float32(gx)*80 + (rb.Rng.Float32()-.5)*78
				z := float32(gz)*80 + (rb.Rng.Float32()-.5)*78
				zi := rb.ZoneAt(x, z)
				if zi >= 0 && zi != 3 && zi != 4 {
					continue
				}
				if !rb.Free(x, z, 1.1) {
					continue
				}
				y := rb.H(x, z)
				if k%10 == 0 && zi < 0 && rb.Free(x, z, 3) {
					h := float32(4) + rb.Rng.Float32()*2
					mb.Cylinder(v3(x, y, z), v3(x, y+h, z), .45, .3, 7, wood, true)
					for j := 0; j < 5; j++ {
						a := float32(j) * 1.2566
						mb.Ellipsoid(v3(x+cosf(a)*1.5, y+h+1, z+sinf(a)*1.5), v3(1.8, 1.5, 1.8), 6, 8, colors[(j+k)%2*4])
					}
					rb.Pillar(x, z, .5, y+h+2)
					rb.Reserve(x, z, 3)
				} else if k%17 == 0 && rb.Free(x, z, 3) {
					for _, dx := range []float32{-2, 0, 2} {
						fy := rb.H(x+dx, z)
						mb.Box(v3(x+dx, fy+0.75, z), v3(.22, 1.5, .22), cream)
					}
					mb.Box(v3(x, y+0.65, z), v3(4, .18, .18), cream)
					mb.Box(v3(x, y+1.15, z), v3(4, .18, .18), cream)
					rb.Reserve(x, z, 3)
				} else if k%8 == 0 {
					mb.Ellipsoid(v3(x, y+.55, z), v3(1.1, .7, .9), 6, 8, c(99, 172, 77))
					rb.Reserve(x, z, 1.1)
				} else if k%13 == 0 {
					mb.Ellipsoid(v3(x, y+.35, z), v3(.8, .45, .65), 6, 8, c(180, 180, 165))
					rb.Reserve(x, z, 1.1)
				} else {
					for j := 0; j < 5; j++ {
						fx := x + float32(j%3)*.5 - .5
						fz := z + float32(j/3)*.6 - .3
						fy := rb.H(fx, fz)
						mb.Cylinder(v3(fx, fy, fz), v3(fx, fy+.85, fz), .06, .04, 5, c(68, 150, 65), false)
						mb.Ellipsoid(v3(fx, fy+.95, fz), v3(.25, .35, .25), 6, 7, colors[(k+j)%len(colors)])
					}
					rb.Reserve(x, z, 1.1)
				}
			}
			rb.Add(mb, true)
		}
	}
}
