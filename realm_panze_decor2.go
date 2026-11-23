package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() { registerRealmDecor("panze_2", panzeDecor2) }

func panzeDecor2(rb *RealmBuilder) {
	c := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	steel, glass, trim := c(91, 113, 135), c(70, 155, 204), c(220, 233, 239)
	cargo := []rl.Color{c(230, 94, 60), c(43, 143, 193), c(242, 179, 59), c(72, 174, 132)}
	box := func(m *MeshBuilder, x, y, z, w, h, d float32, col rl.Color) { m.Box(v3(x, y+h/2, z), v3(w, h, d), col) }
	// Port: individual stacks preserve space for walking and combat.
	port := rb.Mesh()
	for i := 0; i < 16; i++ {
		x := float32(-204 + (i%4)*17)
		z := float32(111 + (i/4)*25)
		if sq(x+207)+sq(z-185) < 23*23 || !rb.Free(x, z, 7) {
			continue
		}
		y := rb.H(x, z)
		n := 1 + i%3
		for j := 0; j < n; j++ {
			yy := y + float32(j)*4
			box(port, x, yy, z, 11, 4, 5, cargo[(i+j)%len(cargo)])
			for k := -4; k <= 4; k += 2 {
				box(port, x+float32(k), yy+0.3, z+2.53, 0.13, 3.4, 0.12, trim)
			}
		}
		rb.Solid(x, z, 11, 5, float32(n)*4, y)
		rb.Reserve(x, z, 7)
	}
	// A tall gantry crane, its collision is limited to the supporting legs.
	x, z := float32(-207), float32(185)
	if rb.Free(x, z, 16) {
		y := rb.H(x, z)
		for _, dx := range []float32{-9, 9} {
			box(port, x+dx, y, z, 2.5, 43, 3, cargo[2])
			rb.Solid(x+dx, z, 2.5, 3, 43, y)
		}
		box(port, x, y+40, z, 25, 4, 4, cargo[2])
		box(port, x+9, y+36, z, 4, 4, 5, glass)
		box(port, x-2, y+24, z, 0.35, 16, 0.35, steel)
		box(port, x-2, y+23, z, 3, 1, 2, steel)
		rb.Reserve(x, z, 16)
	}
	for _, p := range [][2]float32{{-232, 125}, {-230, 220}, {-120, 205}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 13) {
			continue
		}
		y := rb.H(x, z)
		box(port, x, y, z, 20, 10, 14, steel)
		box(port, x, y+9, z, 21, 1, 15, trim)
		box(port, x, y+1, z+7.1, 8, 6, 0.2, glass)
		rb.Solid(x, z, 20, 14, 10, y)
		rb.Reserve(x, z, 13)
	}
	rb.Add(port, true)
	// Laboratory landmark and satellite buildings.
	lab := rb.Mesh()
	x, z = 203, 185
	if rb.Free(x, z, 15) {
		y := rb.H(x, z)
		box(lab, x, y, z, 19, 48, 19, glass)
		rb.Solid(x, z, 19, 19, 48, y)
		for h := float32(5); h < 48; h += 7 {
			box(lab, x, y+h, z, 19.5, 0.65, 19.5, trim)
		}
		lab.Emis = 0.9
		lab.Tube(v3(x, y+37, z), 13, 14, 1.5, 32, c(100, 240, 255))
		lab.Emis = 0
		lab.Ellipsoid(v3(x, y+48, z), v3(9, 5, 9), 8, 16, glass)
		lab.Cylinder(v3(x, y+52, z), v3(x, y+64, z), 0.5, 0.2, 8, trim, true)
		lab.Emis = 0.8
		lab.Sphere(v3(x, y+64, z), 1.2, c(255, 150, 230))
		lab.Emis = 0
		rb.Reserve(x, z, 15)
	}
	for i := 0; i < 9; i++ {
		x := float32(115 + (i%3)*35)
		z := float32(110 + (i/3)*35)
		if !rb.Free(x, z, 10) {
			continue
		}
		y := rb.H(x, z)
		h := float32(14 + i%3*5)
		box(lab, x, y, z, 14, h, 14, glass)
		box(lab, x, y+h-1, z, 15, 1, 15, trim)
		for k := float32(4); k < h-2; k += 5 {
			box(lab, x, y+k, z, 14.3, 0.5, 14.3, steel)
		}
		rb.Solid(x, z, 14, 14, h, y)
		rb.Reserve(x, z, 10)
	}
	rb.Add(lab, true)
	// Urban blocks and occasional trees, in small chunks across the spaces between zones.
	for gx := -3; gx <= 3; gx++ {
		for gz := -3; gz <= 3; gz++ {
			m := rb.Mesh()
			for i := 0; i < 4; i++ {
				x := float32(gx*78 - 24 + (i%2)*35)
				z := float32(gz*78 - 24 + (i/2)*35)
				if rb.ZoneAt(x, z) >= 0 || !rb.Free(x, z, 11) {
					continue
				}
				y := rb.H(x, z)
				if (gx+gz+i+30)%5 == 0 {
					box(m, x, y, z, 15, 0.18, 11, c(107, 113, 120))
					for k := -2; k <= 2; k++ {
						box(m, x+float32(k)*3, y+0.2, z, 0.15, 0.04, 8, trim)
					}
					rb.Reserve(x, z, 9)
					continue
				}
				h := float32(16 + ((gx*gx+gz*gz+i*7)%5)*6)
				w := float32(13)
				col := []rl.Color{c(190, 174, 160), c(160, 181, 198), c(204, 190, 175), c(134, 164, 186)}[(gx*gx+gz*gz+i)%4]
				box(m, x, y, z, w, h, 13, col)
				box(m, x, y+h-1, z, 14, 1, 14, trim)
				for yy := float32(3); yy < h-2; yy += 4 {
					for k := -1; k <= 1; k++ {
						box(m, x+float32(k)*4, y+yy, z+6.55, 2.3, 2.2, 0.12, glass)
						box(m, x+6.55, y+yy, z+float32(k)*4, 0.12, 2.2, 2.3, glass)
					}
				}
				rb.Solid(x, z, 13, 13, h, y)
				rb.Reserve(x, z, 10)
			}
			rb.Add(m, true)
		}
	}
	// Distant skyline follows the mountainous edge.
	for side := 0; side < 4; side++ {
		m := rb.Mesh()
		for t := -240; t <= 240; t += 30 {
			x, z := float32(t), float32(265)
			if side == 1 {
				z = -265
			}
			if side == 2 {
				x, z = 265, float32(t)
			}
			if side == 3 {
				x, z = -265, float32(t)
			}
			if !rb.Free(x, z, 10) {
				continue
			}
			y := rb.H(x, z)
			h := float32(38 + (t * t % 37))
			box(m, x, y, z, 13, h, 13, steel)
			box(m, x, y+h-1, z, 14, 1, 14, trim)
			rb.Solid(x, z, 13, 13, h, y)
			rb.Reserve(x, z, 10)
		}
		rb.Add(m, true)
	}
	// Street trees fill smaller gaps while keeping the roads clear.
	trees := rb.Mesh()
	for x := float32(-255); x <= 255; x += 18 {
		for z := float32(-255); z <= 255; z += 18 {
			if rb.ZoneAt(x, z) >= 0 || !rb.Free(x, z, 3) {
				continue
			}
			y := rb.H(x, z)
			trees.Cylinder(v3(x, y, z), v3(x, y+4, z), 0.35, 0.25, 6, c(120, 83, 55), true)
			trees.Ellipsoid(v3(x, y+5, z), v3(2.3, 2.6, 2.3), 6, 8, c(77, 158, 89))
			rb.Pillar(x, z, 0.4, y+5)
			rb.Reserve(x, z, 3)
		}
	}
	rb.Add(trees, true)
}
