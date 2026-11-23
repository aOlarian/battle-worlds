package main

import rl "github.com/gen2brain/raylib-go/raylib"

func halloweenDecor2(rb *RealmBuilder) {
	pink := rl.NewColor(245, 90, 160, 255)
	cream := rl.NewColor(255, 225, 180, 255)
	purple := rl.NewColor(105, 55, 130, 255)
	wood := rl.NewColor(95, 55, 55, 255)
	orange := rl.NewColor(244, 124, 28, 255)
	gold := rl.NewColor(255, 204, 75, 255)
	green := rl.NewColor(75, 120, 65, 255)
	// Candy market, arranged around the outer edge of zone 3.
	m := rb.Mesh()
	cx, cz := rb.Zones[3].C.X, rb.Zones[3].C.Y
	for i := 0; i < 12; i++ {
		a := float32(i) * 6.283185 / 12
		x, z := cx+cosf(a)*43, cz+sinf(a)*43
		if !rb.Free(x, z, 5) {
			continue
		}
		y := rb.H(x, z)
		m.Box(v3(x, y+1.1, z), v3(6, 2.2, 3.5), wood)
		m.Box(v3(x, y+2.3, z), v3(6.5, .35, 4), cream)
		for j := -1; j <= 1; j += 2 {
			for k := -1; k <= 1; k += 2 {
				m.Cylinder(v3(x+float32(j)*2.6, y, z+float32(k)*1.4), v3(x+float32(j)*2.6, y+5, z+float32(k)*1.4), .16, .16, 6, wood, true)
			}
		}
		for j := 0; j < 6; j++ {
			c := pink
			if j%2 == 0 {
				c = cream
			}
			m.Box(v3(x-2.5+float32(j), y+5, z), v3(1, .35, 4.5), c)
		}
		m.Emis = 1
		m.Box(v3(x, y+5.2, z), v3(6.5, .16, 4.5), pink)
		for j := -2; j <= 2; j++ {
			m.Sphere(v3(x+float32(j), y+4.7, z+2.2), .18, gold)
		}
		m.Emis = 0
		rb.Solid(x, z, 6, 3.5, 2.5, y)
		rb.Reserve(x, z, 5)
	}
	// Monumental spiral lollipop, vertical disc made from luminous spiral beads.
	for _, p := range [][2]float32{{-202, 183}, {-202, 117}, {-120, 190}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 5) {
			continue
		}
		y := rb.H(x, z)
		m.Cylinder(v3(x, y, z), v3(x, y+10, z), .5, .5, 10, cream, true)
		m.Emis = 1
		m.Ellipsoid(v3(x, y+11, z), v3(4, 4, .55), 12, 20, cream)
		m.Emis = .8
		for j := 0; j < 100; j++ {
			a := float32(j) * .27
			r := float32(j) * .037
			m.Sphere(v3(x+cosf(a)*r, y+11+sinf(a)*r, z+.58), .24, pink)
		}
		m.Emis = 0
		rb.Pillar(x, z, 1, y+15)
		rb.Reserve(x, z, 5)
	}
	for i := 0; i < 16; i++ {
		a := float32(i) * 6.283185 / 16
		x, z := cx+cosf(a)*28, cz+sinf(a)*28
		if !rb.Free(x, z, 2) {
			continue
		}
		y := rb.H(x, z)
		c := pink
		if i%2 == 0 {
			c = gold
		}
		m.Ellipsoid(v3(x, y+1.2, z), v3(1.8, 1, 1), 8, 12, c)
		m.Cylinder(v3(x-2.5, y+1.2, z), v3(x-1.4, y+1.2, z), 1, .5, 7, cream, true)
		m.Cylinder(v3(x+1.4, y+1.2, z), v3(x+2.5, y+1.2, z), .5, 1, 7, cream, true)
		rb.Reserve(x, z, 2)
	}
	rb.Add(m, true)
	// Manor placed towards the outside of zone 4.
	m = rb.Mesh()
	x, z := float32(205), float32(115)
	if rb.Free(x, z, 23) {
		y := rb.H(x, z)
		m.Box(v3(x, y+8, z), v3(30, 16, 20), purple)
		m.Emis = 1
		m.Box(v3(x, y+17, z), v3(33, .4, 22), gold)
		m.Emis = 0
		m.Box(v3(x, y+16.5, z), v3(33, 1, 22), wood)
		m.Box(v3(x, y+19, z), v3(24, 3, 16), purple)
		for _, dx := range []float32{-12, -6, 0, 6, 12} {
			for _, dy := range []float32{4, 10, 15} {
				m.Emis = .95
				m.Box(v3(x+dx, y+dy, z-10.1), v3(2.6, 3, .3), gold)
				m.Emis = 0
				m.Box(v3(x+dx, y+dy, z-10.3), v3(.18, 3, .2), wood)
			}
		}
		for _, dx := range []float32{-9, -3, 3, 9} {
			m.Cylinder(v3(x+dx, y, z-13), v3(x+dx, y+9, z-13), .55, .45, 10, cream, true)
		}
		m.Box(v3(x, y+9, z-12), v3(23, 1, 5), cream)
		m.Box(v3(x, y+10.5, z-14), v3(23, .3, .3), wood)
		for j := 0; j < 4; j++ {
			m.Box(v3(x, y+float32(j)*.35+.18, z-15+float32(j)), v3(12, .35, 3), cream)
		}
		rb.Solid(x, z, 30, 20, 20, y)
		rb.Reserve(x, z, 23)
	}
	rb.Add(m, true)
	// A little illuminated wheel and clouds of cotton candy.
	m = rb.Mesh()
	wx, wz := float32(-205), float32(145)
	if rb.Free(wx, wz, 5) {
		y := rb.H(wx, wz)
		m.Cylinder(v3(wx, y, wz), v3(wx, y+6, wz), .4, .4, 8, wood, true)
		m.Emis = .9
		for i := 0; i < 24; i++ {
			a := float32(i) * 6.283185 / 24
			m.Sphere(v3(wx+cosf(a)*4, y+6+sinf(a)*4, wz), .23, gold)
		}
		m.Emis = 0
		for i := 0; i < 8; i++ {
			a := float32(i) * 6.283185 / 8
			m.Cylinder(v3(wx, y+6, wz), v3(wx+cosf(a)*4, y+6+sinf(a)*4, wz), .1, .1, 6, cream, true)
		}
		rb.Reserve(wx, wz, 5)
	}
	for i := 0; i < 8; i++ {
		a := float32(i) * 6.283185 / 8
		x, z := cx+cosf(a)*35, cz+sinf(a)*35
		if !rb.Free(x, z, 2) {
			continue
		}
		y := rb.H(x, z)
		m.Cylinder(v3(x, y, z), v3(x, y+3, z), .14, .14, 6, cream, true)
		m.Ellipsoid(v3(x, y+4, z), v3(1.4, 1.8, 1.2), 8, 10, pink)
		rb.Reserve(x, z, 2)
	}
	rb.Add(m, true)
	// Friendly ghost-shaped topiary around the manor garden.
	m = rb.Mesh()
	for i := 0; i < 10; i++ {
		a := float32(i) * 6.283185 / 10
		x, z := rb.Zones[4].C.X+cosf(a)*31, rb.Zones[4].C.Y+sinf(a)*31
		if !rb.Free(x, z, 2) {
			continue
		}
		y := rb.H(x, z)
		m.Ellipsoid(v3(x, y+1.5, z), v3(1.1, 1.5, .8), 8, 10, green)
		m.Sphere(v3(x, y+2.8, z), .8, green)
		m.Sphere(v3(x-.3, y+2.9, z-.7), .12, cream)
		m.Sphere(v3(x+.3, y+2.9, z-.7), .12, cream)
		rb.Reserve(x, z, 2)
	}
	rb.Add(m, true)
	// Roadside lanterns, kept outside the road clearance.
	m = rb.Mesh()
	for _, road := range rb.CB.Roads {
		dx, dz := road.B.X-road.A.X, road.B.Y-road.A.Y
		length := sqrtf(dx*dx + dz*dz)
		if length < 1 {
			continue
		}
		ux, uz := dx/length, dz/length
		for t := float32(12); t < length-12; t += 12 {
			for _, side := range []float32{-1, 1} {
				x := road.A.X + ux*t - uz*side*(road.HalfW+4)
				z := road.A.Y + uz*t + ux*side*(road.HalfW+4)
				if !rb.Free(x, z, .6) {
					continue
				}
				y := rb.H(x, z)
				m.Cylinder(v3(x, y, z), v3(x, y+4, z), .16, .16, 6, wood, true)
				m.Emis = 1
				m.Box(v3(x, y+4, z), v3(.9, 1, .9), gold)
				// Festoon bulbs suspended between neighbouring posts.
				for j := 1; j < 6; j++ {
					f := float32(j) / 6
					xx, zz := x+ux*12*f, z+uz*12*f
					if rb.Free(xx, zz, .1) {
						m.Sphere(v3(xx, y+4.1-1.2*4*f*(1-f), zz), .14, orange)
					}
				}
				m.Emis = 0
				rb.Reserve(x, z, .6)
			}
		}
	}
	rb.Add(m, true)
	// Candle groups, candy baskets and mossy stones across all five settlements.
	for zone := 0; zone < 5; zone++ {
		m = rb.Mesh()
		for i := 0; i < 40; i++ {
			a := float32(i) * 2.39996
			r := float32(20 + (i%10)*3)
			x, z := rb.Zones[zone].C.X+cosf(a)*r, rb.Zones[zone].C.Y+sinf(a)*r
			if !rb.Free(x, z, 1) {
				continue
			}
			y := rb.H(x, z)
			if i%3 == 0 {
				m.Cylinder(v3(x, y, z), v3(x, y+.6, z), .65, .8, 8, wood, true)
				m.Emis = 1
				for j := -1; j <= 1; j++ {
					m.Sphere(v3(x+float32(j)*.35, y+.7, z), .22, []rl.Color{pink, gold, orange}[j+1])
				}
				m.Emis = 0
			} else if i%3 == 1 {
				for j := -1; j <= 1; j++ {
					xx := x + float32(j)*.35
					m.Cylinder(v3(xx, y, z), v3(xx, y+.9+float32(j+1)*.15, z), .15, .15, 6, cream, true)
					m.Emis = 1
					m.Sphere(v3(xx, y+1.1+float32(j+1)*.15, z), .14, gold)
					m.Emis = 0
				}
			} else {
				m.Ellipsoid(v3(x, y+.3, z), v3(.8, .4, .6), 5, 7, green)
			}
			rb.Reserve(x, z, 1)
		}
		rb.Add(m, true)
	}
	// Trees, lanterns, pumpkins and fireflies in the spaces between zones.
	for gx := -260; gx <= 260; gx += 65 {
		for gz := -260; gz <= 260; gz += 65 {
			m = rb.Mesh()
			for j := 0; j < 16; j++ {
				x := float32(gx) + rb.Rng.Float32()*60
				z := float32(gz) + rb.Rng.Float32()*60
				if rb.ZoneAt(x, z) >= 0 || !rb.Free(x, z, 3) {
					continue
				}
				y := rb.H(x, z)
				h := float32(6) + rb.Rng.Float32()*5
				bend := float32(2)
				m.Cylinder(v3(x, y, z), v3(x+bend, y+h, z), .7, .25, 7, wood, true)
				for k := -1; k <= 1; k += 2 {
					m.Cylinder(v3(x+bend*.6, y+h*.6, z), v3(x+bend+float32(k)*3, y+h+2, z+float32(k)), .28, .06, 6, wood, true)
				}
				rb.Pillar(x, z, .7, y+h)
				rb.Reserve(x, z, 3)
				m.Ellipsoid(v3(x+2, y+.55, z+2), v3(1, .65, .9), 8, 10, orange)
				m.Cylinder(v3(x+2, y+1, z+2), v3(x+2, y+1.5, z+2), .12, .1, 6, green, true)
				m.Emis = .9
				m.Sphere(v3(x-2, y+2+rb.Rng.Float32()*3, z+1), .16, gold)
				m.Emis = 0
			}
			rb.Add(m, true)
		}
	}
}
