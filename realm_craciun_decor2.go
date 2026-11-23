package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() { registerRealmDecor("craciun_2", craciunDecor2) }

func craciunDecor2(rb *RealmBuilder) {
	snow := rl.NewColor(245, 250, 255, 255)
	ice := rl.NewColor(145, 220, 245, 255)
	dark := rl.NewColor(58, 75, 86, 255)
	wood := rl.NewColor(130, 77, 44, 255)
	green := rl.NewColor(30, 112, 71, 255)
	gold := rl.NewColor(255, 218, 91, 255)
	red := rl.NewColor(205, 48, 52, 255)
	// Rink sits away from the plaza and the road.
	m := rb.Mesh()
	x, z := float32(-210), float32(110)
	if rb.Free(x, z, 20) {
		y := rb.H(x, z)
		m.Disc(v3(x, y+.05, z), 0, 19, 48, ice)
		m.Disc(v3(x, y+.08, z), 18.4, 19, 48, snow)
		rb.Reserve(x, z, 20)
	}
	// Pavilion: open sides and a snow-capped dome.
	x, z = -120, 195
	if rb.Free(x, z, 9) {
		y := rb.H(x, z)
		m.Cylinder(v3(x, y, z), v3(x, y+1, z), 8, 8, 12, wood, true)
		for i := 0; i < 8; i++ {
			a := float32(i) * .785398
			px, pz := x+6*cosf(a), z+6*sinf(a)
			m.Cylinder(v3(px, y+1, pz), v3(px, y+9, pz), .35, .35, 8, red, true)
			m.Emis = 1
			m.Sphere(v3(px, y+8.5, pz), .48, gold)
			m.Emis = 0
		}
		m.Cylinder(v3(x, y+9, z), v3(x, y+10, z), 7.4, 7.4, 12, red, true)
		m.Ellipsoid(v3(x, y+10, z), v3(7.4, 4.5, 7.4), 8, 16, snow)
		rb.Reserve(x, z, 9)
		rb.Pillar(x, z, 7, y+10)
	}
	// Benches, lanterns and a fire on the lake shore.
	for _, p := range [][2]float32{{-217, 170}, {-213, 192}, {-180, 204}, {-175, 180}, {-188, 109}, {-211, 133}} {
		x, z = p[0], p[1]
		if !rb.Free(x, z, 2.5) {
			continue
		}
		y := rb.H(x, z)
		m.Box(v3(x, y+1, z), v3(3, .4, 1), wood)
		m.Box(v3(x, y+1.8, z+.45), v3(3, 1.4, .25), wood)
		for _, dx := range []float32{-1, 1} {
			m.Box(v3(x+dx, y+.5, z), v3(.3, 1, .5), dark)
		}
		rb.Reserve(x, z, 2.5)
	}
	for _, p := range [][2]float32{{-220, 180}, {-203, 205}, {-174, 194}, {-207, 110}, {-177, 113}} {
		x, z = p[0], p[1]
		if !rb.Free(x, z, 1) {
			continue
		}
		y := rb.H(x, z)
		m.Cylinder(v3(x, y, z), v3(x, y+4, z), .18, .18, 6, dark, true)
		m.Emis = 1
		m.Box(v3(x, y+4, z), v3(.8, 1, .8), gold)
		m.Emis = 0
		rb.Reserve(x, z, 1)
	}
	x, z = -221, 150
	if rb.Free(x, z, 4) {
		y := rb.H(x, z)
		m.Cylinder(v3(x-2, y+.5, z), v3(x+2, y+.5, z), .6, .6, 8, wood, true)
		m.Cylinder(v3(x, y+.2, z), v3(x, y+3, z), 1.2, 0, 8, red, true)
		m.Emis = 1
		m.Cylinder(v3(x, y+.4, z), v3(x, y+2.4, z), .7, 0, 8, gold, true)
		m.Emis = 0
		rb.Reserve(x, z, 4)
	}
	rb.Add(m, true)
	// Cave is an arch of stone, leaving its dark entrance visibly open.
	m = rb.Mesh()
	x, z = 210, 110
	if rb.Free(x, z, 12) {
		y := rb.H(x, z)
		for _, dx := range []float32{-7, 7} {
			m.Ellipsoid(v3(x+dx, y+6, z), v3(5, 7, 5), 8, 12, dark)
			rb.Pillar(x+dx, z, 4, y+12)
		}
		m.Ellipsoid(v3(x, y+12, z), v3(12, 5, 6), 8, 16, dark)
		m.Ellipsoid(v3(x, y+15, z), v3(10, 2, 5), 6, 12, snow)
		m.Box(v3(x, y+5, z+3), v3(10, 9, .4), rl.NewColor(37, 59, 75, 255))
		for i := -3; i <= 3; i++ {
			px := x + float32(i)*2.6
			m.Cylinder(v3(px, y+11, z-4), v3(px, y+8-float32(i*i%3), z-4), .7, 0, 5, ice, true)
		}
		rb.Reserve(x, z, 12)
	}
	for _, p := range [][2]float32{{180, 190}, {217, 189}, {181, 165}, {220, 164}, {200, 206}} {
		x, z = p[0], p[1]
		if !rb.Free(x, z, 3) {
			continue
		}
		y := rb.H(x, z)
		for i := 0; i < 3; i++ {
			xx := x + float32(i-1)*1.1
			m.Cylinder(v3(xx, y, z), v3(xx, y+4+float32(i%2)*2, z), 1.1, 0, 5, ice, true)
		}
		rb.Reserve(x, z, 3)
		rb.Pillar(x, z, 1.5, y+4)
	}
	for _, p := range [][2]float32{{184, 134}, {190, 132}, {196, 130}, {202, 132}, {208, 134}} {
		x, z = p[0], p[1]
		if !rb.Free(x, z, 2) {
			continue
		}
		y := rb.H(x, z)
		m.Box(v3(x, y+1, z), v3(2, 2, 2), red)
		m.Box(v3(x, y+1.04, z), v3(.35, 2.1, 2.1), gold)
		m.Box(v3(x, y+2.06, z), v3(2.1, .15, .35), gold)
		rb.Reserve(x, z, 2)
	}
	for _, x = range []float32{220, 226} {
		z = 145
		if !rb.Free(x, z, 3) {
			continue
		}
		y := rb.H(x, z)
		m.BeveledBox(v3(x, y+2, z), v3(3, 4, 3), .4, green)
		m.BeveledBox(v3(x, y+.8, z-1.6), v3(3, 1.6, 5), .4, green)
		rb.Reserve(x, z, 3)
		rb.Solid(x, z, 3, 5, 4, y)
	}
	rb.Add(m, true)
	// Snowy forest, cabins, snowmen, reindeer and mountain peaks between zones.
	for gx := -3; gx <= 3; gx++ {
		for gz := -3; gz <= 3; gz++ {
			m = rb.Mesh()
			for j := 0; j < 14; j++ {
				x = float32(gx*80) + rb.Rng.Float32()*70 - 35
				z = float32(gz*80) + rb.Rng.Float32()*70 - 35
				if rb.ZoneAt(x, z) >= 0 && rb.ZoneAt(x, z) != 3 && rb.ZoneAt(x, z) != 4 {
					continue
				}
				kind := rb.Rng.Intn(14)
				r := float32(3.5)
				if kind == 0 {
					r = 6
				}
				if !rb.Free(x, z, r) {
					continue
				}
				y := rb.H(x, z)
				switch kind {
				case 0:
					m.Box(v3(x, y+3, z), v3(8, 6, 7), wood)
					m.Cylinder(v3(x, y+6, z), v3(x, y+10, z), 6, 0, 4, snow, true)
					m.Box(v3(x, y+2, z-3.6), v3(2, 4, .2), red)
					rb.Solid(x, z, 8, 7, 6, y)
				case 1:
					m.Sphere(v3(x, y+1.5, z), 1.5, snow)
					m.Sphere(v3(x, y+3.4, z), 1.1, snow)
					m.Sphere(v3(x, y+4.9, z), .8, snow)
					m.Box(v3(x, y+5.8, z), v3(1.5, .35, 1.5), dark)
					rb.Pillar(x, z, 1.4, y+5)
				case 2:
					m.Ellipsoid(v3(x, y+1.5, z), v3(2.5, 1.5, 1.8), 6, 8, wood)
					m.Ellipsoid(v3(x, y+3, z-1.5), v3(.8, 1, .8), 6, 8, wood)
					for _, dx := range []float32{-1.4, 1.4} {
						for _, dz := range []float32{-1, 1} {
							m.Cylinder(v3(x+dx, y, z+dz), v3(x+dx, y+1.5, z+dz), .2, .2, 5, wood, true)
						}
						m.Cylinder(v3(x+dx*.4, y+3.5, z-1.5), v3(x+dx, y+5, z-1.5), .15, .1, 5, wood, true)
					}
				case 3:
					m.Ellipsoid(v3(x, y+1.4, z), v3(3, 2, 2.6), 6, 10, snow)
					rb.Pillar(x, z, 2, y+3)
				default:
					h := float32(9) + rb.Rng.Float32()*6
					m.Cylinder(v3(x, y, z), v3(x, y+h*.45, z), .7, .5, 7, wood, true)
					for k := 0; k < 3; k++ {
						base := y + 2 + float32(k)*h*.21
						rad := 4 - float32(k)*.7
						m.Cylinder(v3(x, base, z), v3(x, base+h*.52, z), rad, 0, 9, green, true)
						m.Cylinder(v3(x, base+h*.12, z), v3(x, base+h*.53, z), rad*.78, 0, 9, snow, true)
					}
					rb.Pillar(x, z, 1, y+h)
				}
				rb.Reserve(x, z, r)
			}
			rb.Add(m, true)
		}
	}
	// Snowy peaks along the playable edge.
	m = rb.Mesh()
	for i := -3; i <= 3; i++ {
		for _, side := range []int{-1, 1} {
			for axis := 0; axis < 2; axis++ {
				x, z = float32(i*75), float32(side*265)
				if axis == 1 {
					x, z = z, x
				}
				if !rb.Free(x, z, 12) {
					continue
				}
				y := rb.H(x, z)
				m.Cylinder(v3(x, y, z), v3(x, y+25, z), 12, 0, 7, dark, true)
				m.Cylinder(v3(x, y+14, z), v3(x, y+25.3, z), 5.4, 0, 7, snow, true)
				rb.Reserve(x, z, 12)
				rb.Pillar(x, z, 8, y+20)
			}
		}
	}
	rb.Add(m, true)
}
