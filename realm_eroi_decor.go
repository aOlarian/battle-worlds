package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() { registerRealmDecor("eroi", eroiDecor) }

func eroiDecor(rb *RealmBuilder) {
	gold := rl.NewColor(244, 193, 76, 255)
	pale := rl.NewColor(238, 228, 194, 255)
	blue := rl.NewColor(55, 123, 183, 255)
	glass := rl.NewColor(113, 202, 228, 255)
	ice := rl.NewColor(189, 236, 250, 255)
	white := rl.NewColor(241, 249, 247, 255)
	// Find a clear site on the perimeter, keeping the whole landmark off roads.
	site := func(cx, cz, r float32, start int) (float32, float32, bool) {
		for k := 0; k < 32; k++ {
			a := float32((k+start)%32) * 0.19634954
			x, z := cx+cosf(a)*43, cz+sinf(a)*43
			if rb.Free(x, z, r) {
				rb.Reserve(x, z, r)
				return x, z, true
			}
		}
		return 0, 0, false
	}
	for zone := 0; zone < 3; zone++ {
		c := rb.Zones[zone].C
		m := rb.Mesh()
		x, z, ok := site(c.X, c.Y, 11, 3)
		if ok {
			y := rb.H(x, z)
			switch zone {
			case 0:
				// Fountain and a friendly, original caped figure on its plinth.
				m.Cylinder(v3(x, y, z), v3(x, y+1, z), 10, 10, 24, pale, true)
				m.Cylinder(v3(x, y+1, z), v3(x, y+1.15, z), 8.8, 8.8, 24, glass, true)
				m.Box(v3(x, y+3, z), v3(6, 4, 6), gold)
				rb.Solid(x, z, 6, 6, 5, y)
				m.Box(v3(x, y+7, z), v3(3.4, 4, 2), blue)
				for _, dx := range []float32{-1, 1} {
					m.Box(v3(x+dx, y+5.2, z), v3(1.2, 2, 1.4), pale)
					m.Cylinder(v3(x+dx*1.8, y+8.5, z), v3(x+dx*2.7, y+11, z), .65, .45, 8, gold, true)
				}
				m.Sphere(v3(x, y+10.2, z), 1.35, pale)
				m.Quad(v3(x-1.5, y+8.5, z-1), v3(x+1.5, y+8.5, z-1), v3(x+2.5, y+5, z-3), v3(x-2.5, y+5, z-3), v3(0, 0, -1), gold)
				m.Cylinder(v3(x, y+5, z), v3(x, y+11.6, z), .35, .2, 10, white, true)
			case 1:
				// Newspaper tower with a rooftop globe.
				m.Box(v3(x, y+15, z), v3(16, 30, 14), pale)
				rb.Solid(x, z, 16, 14, 30, y)
				m.Box(v3(x, y+31, z), v3(12, 2, 10), blue)
				rb.Solid(x, z, 12, 10, 32, y)
				m.Sphere(v3(x, y+36, z), 3.5, gold)
				for _, dx := range []float32{-6, -3, 0, 3, 6} {
					m.Box(v3(x+dx, y+15, z+7.1), v3(1.3, 24, .25), glass)
				}
			case 2:
				// A cluster of faceted luminous crystal spires.
				m.Box(v3(x, y+2, z), v3(18, 4, 16), ice)
				rb.Solid(x, z, 18, 16, 4, y)
				m.Emis = .22
				for _, p := range [][3]float32{{0, 0, 28}, {-6, -3, 18}, {6, 2, 21}, {-3, 5, 14}, {4, -5, 16}} {
					h := p[2]
					m.Cylinder(v3(x+p[0], y+4, z+p[1]), v3(x+p[0], y+4+h*.7, z+p[1]), 2.5, 2, 6, glass, true)
					m.Cylinder(v3(x+p[0], y+4+h*.7, z+p[1]), v3(x+p[0], y+4+h, z+p[1]), 2, 0, 6, white, true)
				}
				m.Emis = 0
			}
		}
		// Additional station in zone 1 with a visible clock.
		if zone == 1 {
			sx, sz, yes := site(c.X, c.Y, 10, 19)
			if yes {
				y := rb.H(sx, sz)
				m.Box(v3(sx, y+7, sz), v3(18, 14, 14), pale)
				rb.Solid(sx, sz, 18, 14, 14, y)
				m.Box(v3(sx, y+18, sz), v3(6, 8, 6), blue)
				rb.Solid(sx, sz, 6, 6, 22, y)
				m.Box(v3(sx, y+18, sz+3.1), v3(4, 4, .2), gold)
				m.Box(v3(sx, y+18, sz+3.25), v3(.2, 2, .1), white)
				m.Box(v3(sx+.7, y+18, sz+3.26), v3(1.5, .2, .1), white)
			}
		}
		// Tiered art-deco blocks, with roof terraces and vertical gold ribs.
		for i := 0; i < 24; i++ {
			a := float32(i) * 6.2831853 / 24
			d := float32(52)
			bx, bz := c.X+cosf(a)*d, c.Y+sinf(a)*d
			if !rb.Free(bx, bz, 8) {
				continue
			}
			rb.Reserve(bx, bz, 8)
			y := rb.H(bx, bz)
			h := float32(16 + (i%4)*5)
			col := pale
			if zone == 2 {
				col = ice
			}
			m.Box(v3(bx, y+h/2, bz), v3(11, h, 10), col)
			rb.Solid(bx, bz, 11, 10, h, y)
			m.Box(v3(bx, y+h+1, bz), v3(8, 2, 7), blue)
			rb.Solid(bx, bz, 8, 7, h+2, y)
			for _, dx := range []float32{-4, 0, 4} {
				m.Box(v3(bx+dx, y+h/2, bz+5.05), v3(1, h-3, .18), glass)
			}
			m.Box(v3(bx, y+h-.4, bz), v3(11.4, .8, 10.4), gold)
		}
		// Low lamps bring the theme closer to the fighting area.
		for i := 0; i < 16; i++ {
			a := float32(i) * 6.2831853 / 16
			lx, lz := c.X+cosf(a)*24, c.Y+sinf(a)*24
			if !rb.Free(lx, lz, 1) {
				continue
			}
			rb.Reserve(lx, lz, 1)
			y := rb.H(lx, lz)
			m.Cylinder(v3(lx, y, lz), v3(lx, y+3, lz), .16, .12, 8, gold, true)
			m.Emis = .25
			m.Sphere(v3(lx, y+3.2, lz), .45, white)
			m.Emis = 0
		}
		rb.Add(m, true)
	}
}
