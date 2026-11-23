package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() { registerRealmDecor("monstri_2", monstriDecor2) }

func monstriDecor2(rb *RealmBuilder) {
	wood := rl.NewColor(111, 69, 55, 255)
	roof := rl.NewColor(83, 55, 98, 255)
	stone := rl.NewColor(125, 120, 133, 255)
	gold := rl.NewColor(255, 193, 83, 255)
	pine := rl.NewColor(37, 76, 64, 255)
	moss := rl.NewColor(91, 129, 82, 255)
	// Satul: case asezate in jurul pietei.
	m := rb.Mesh()
	house := func(x, z float32, tower bool) {
		r := float32(8)
		if !rb.Free(x, z, r) {
			return
		}
		y := rb.H(x, z)
		h := float32(7)
		if tower {
			h = 18
		}
		m.Box(v3(x, y+h/2, z), v3(10, h, 8), wood)
		m.Cylinder(v3(x, y+h, z), v3(x, y+h+6, z), 7, 0, 4, roof, true)
		m.Emis = .8
		for _, dx := range []float32{-3, 3} {
			m.Box(v3(x+dx, y+3, z+4.05), v3(1.7, 2.2, .15), gold)
		}
		if tower {
			m.Box(v3(x, y+13, z+4.05), v3(2, 3, .15), gold)
		}
		m.Emis = 0
		m.Box(v3(x, y+1.6, z+4.1), v3(2, 3.2, .2), roof)
		rb.Solid(x, z, 10, 8, h+5, y)
		rb.Reserve(x, z, r)
	}
	for _, p := range [][2]float32{{-200, 130}, {-198, 168}, {-180, 190}, {-142, 193}, {-120, 170}, {-122, 130}, {-150, 110}} {
		house(p[0], p[1], false)
	}
	house(-210, 205, true)
	// Fantana si obiecte marunte.
	if rb.Free(-180, 150, 4) {
		x, z := float32(-180), float32(150)
		y := rb.H(x, z)
		m.Tube(v3(x, y, z), 2, 3, 2, 12, stone)
		m.Cylinder(v3(x-2, y+1, z), v3(x-2, y+5, z), .25, .25, 6, wood, true)
		m.Cylinder(v3(x+2, y+1, z), v3(x+2, y+5, z), .25, .25, 6, wood, true)
		m.Box(v3(x, y+5, z), v3(5, .5, 1), roof)
		rb.Pillar(x, z, 3, y+5)
		rb.Reserve(x, z, 4)
	}
	for _, p := range [][2]float32{{-177, 177}, {-143, 175}, {-176, 122}, {-135, 139}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 3) {
			continue
		}
		y := rb.H(x, z)
		m.Box(v3(x, y+1, z), v3(4, 1.3, 2), wood)
		for _, dx := range []float32{-1.5, 1.5} {
			m.Cylinder(v3(x+dx, y+.8, z-1.2), v3(x+dx, y+.8, z+1.2), .8, .8, 10, roof, true)
		}
		rb.Reserve(x, z, 3)
	}
	for _, p := range [][2]float32{{-185, 140}, {-135, 183}, {-150, 120}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 2) {
			continue
		}
		y := rb.H(x, z)
		for i := 0; i < 4; i++ {
			m.Cylinder(v3(x-1, y+.3+float32(i/2)*.5, z+float32(i%2)*.6), v3(x+1, y+.3+float32(i/2)*.5, z+float32(i%2)*.6), .3, .3, 6, wood, true)
		}
		rb.Reserve(x, z, 2)
	}
	if rb.Free(-150, 132, 4) {
		x, z := float32(-150), float32(132)
		y := rb.H(x, z)
		for _, dx := range []float32{-3, 3} {
			m.Cylinder(v3(x+dx, y, z), v3(x+dx, y+4, z), .12, .12, 6, wood, true)
		}
		m.Cylinder(v3(x-3, y+3.8, z), v3(x+3, y+3.8, z), .06, .06, 6, wood, true)
		for i := 0; i < 3; i++ {
			m.Box(v3(x-2+float32(i)*2, y+3, z), v3(1.4, 1.4, .08), []rl.Color{gold, stone, roof}[i])
		}
		rb.Reserve(x, z, 4)
	}
	rb.Add(m, true)
	// Lacul este un disc decorativ amplasat departe de drum si piateta.
	m = rb.Mesh()
	x, z := float32(191), float32(180)
	if rb.Free(x, z, 22) {
		y := rb.H(x, z)
		m.Disc(v3(x, y+.05, z), 0, 21, 48, rl.NewColor(104, 143, 163, 255))
		rb.Reserve(x, z, 22)
		for i := 0; i < 10; i++ {
			a := float32(i) * .6283
			px, pz := x+cosf(a)*19, z+sinf(a)*19
			for j := 0; j < 3; j++ {
				m.Cylinder(v3(px+float32(j)*.3, y, pz), v3(px+float32(j)*.3, y+2+float32(j)*.3, pz), .07, .035, 5, moss, true)
			}
		}
		m.Box(v3(x-15, y+.35, z), v3(12, .4, 3), wood)
		for i := 0; i < 5; i++ {
			m.Box(v3(x-19+float32(i)*2, y+.58, z), v3(.1, .08, 3), stone)
		}
		m.Ellipsoid(v3(x-10, y+.5, z+5), v3(3, .6, 1.2), 6, 10, wood)
		for i := 0; i < 6; i++ {
			a := float32(i) * 1.05
			m.Ellipsoid(v3(x+cosf(a)*11, y+1.5, z+sinf(a)*11), v3(4, 1.2, 2.5), 6, 10, rl.NewColor(205, 196, 230, 100))
		}
	}
	// Farul lacului.
	x, z = 210, 115
	if rb.Free(x, z, 7) {
		y := rb.H(x, z)
		m.Cylinder(v3(x, y, z), v3(x, y+19, z), 4, 2.8, 12, stone, true)
		m.Emis = .85
		m.Cylinder(v3(x, y+19, z), v3(x, y+22, z), 2.6, 2.6, 12, gold, true)
		m.Emis = 0
		m.Cylinder(v3(x, y+22, z), v3(x, y+25, z), 3.5, 0, 12, roof, true)
		rb.Pillar(x, z, 4, y+25)
		rb.Reserve(x, z, 7)
	}
	rb.Add(m, true)
	// Padurea dintre zone si de la margine, impartita in bucati.
	for gx := -3; gx <= 3; gx++ {
		for gz := -3; gz <= 3; gz++ {
			m = rb.Mesh()
			for i := 0; i < 34; i++ {
				x := float32(gx*80-40) + rb.Rng.Float32()*80
				z := float32(gz*80-40) + rb.Rng.Float32()*80
				if rb.ZoneAt(x, z) != -1 || !rb.Free(x, z, 3.2) {
					continue
				}
				y := rb.H(x, z)
				h := float32(9) + rb.Rng.Float32()*8
				switch i % 7 {
				case 0:
					m.Ellipsoid(v3(x, y+1.4, z), v3(3, 1.8, 2.4), 6, 8, stone)
					m.Ellipsoid(v3(x, y+2.6, z), v3(2, .5, 1.5), 5, 8, moss)
					rb.Pillar(x, z, 2.3, y+3)
				case 1:
					m.Cylinder(v3(x, y, z), v3(x+1, y+h*.65, z), .65, .35, 7, wood, true)
					m.Ellipsoid(v3(x+1, y+h*.7, z), v3(3, 3, 3), 6, 8, pine)
					rb.Pillar(x, z, .7, y+h)
				default:
					m.Cylinder(v3(x, y, z), v3(x, y+h*.45, z), .65, .5, 7, wood, true)
					for j := 0; j < 3; j++ {
						base := y + 3 + float32(j)*h*.19
						m.Cylinder(v3(x, base, z), v3(x, base+h*.5, z), 3.7-float32(j)*.55, 0, 8, pine, true)
					}
					rb.Pillar(x, z, .7, y+h)
				}
				rb.Reserve(x, z, 3.2)
				if i%3 == 0 {
					m.Cylinder(v3(x+2, y, z), v3(x+2, y+.8, z), .16, .12, 6, stone, true)
					m.Ellipsoid(v3(x+2, y+.85, z), v3(.6, .3, .6), 5, 8, roof)
				}
			}
			rb.Add(m, true)
		}
	}
	// Siluete pe coama, departe de zonele de lupta.
	m = rb.Mesh()
	for _, p := range [][2]float32{{-270, -260}, {270, -260}, {-270, 260}, {270, 260}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 7) {
			continue
		}
		y := rb.H(x, z)
		m.Box(v3(x, y+5, z), v3(9, 10, 6), stone)
		for _, dx := range []float32{-4, 4} {
			m.Cylinder(v3(x+dx, y, z), v3(x+dx, y+13, z), 2, 2, 8, stone, true)
			m.Cylinder(v3(x+dx, y+13, z), v3(x+dx, y+17, z), 2.5, 0, 8, roof, true)
		}
		rb.Solid(x, z, 9, 6, 13, y)
		rb.Reserve(x, z, 7)
	}
	rb.Add(m, true)
}
