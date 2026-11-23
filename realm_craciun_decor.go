package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() { registerRealmDecor("craciun", christmasDecor) }

func christmasDecor(rb *RealmBuilder) {
	white := rl.NewColor(247, 250, 255, 255)
	green := rl.NewColor(28, 112, 66, 255)
	red := rl.NewColor(196, 44, 50, 255)
	gold := rl.NewColor(255, 202, 65, 255)
	brown := rl.NewColor(159, 92, 51, 255)
	blue := rl.NewColor(65, 145, 200, 255)
	pink := rl.NewColor(240, 113, 147, 255)
	dark := rl.NewColor(65, 42, 35, 255)
	colors := []rl.Color{red, gold, blue, pink}
	tree := func(m *MeshBuilder, x, y, z, h float32) {
		m.Cylinder(v3(x, y, z), v3(x, y+h*.18, z), h*.07, h*.06, 8, brown, true)
		for j := 0; j < 3; j++ {
			a := float32(j)
			m.Cylinder(v3(x, y+h*(.12+a*.20), z), v3(x, y+h*(.60+a*.20), z), h*(.26-a*.045), 0, 12, green, true)
		}
		m.Emis = 1
		for j := 0; j < 30; j++ {
			a := float32(j) * 2.399
			yy := float32(j) / 30
			r := h * (.24 - yy*.18)
			m.Sphere(v3(x+cosf(a)*r, y+h*(.17+yy*.70), z+sinf(a)*r), h*.018, colors[j%4])
		}
		m.Emis = 0
	}
	gift := func(m *MeshBuilder, x, y, z, s float32, c rl.Color) {
		m.Box(v3(x, y+s/2, z), v3(s, s, s), c)
		m.Box(v3(x, y+s/2, z), v3(s*.16, s+.06, s+.06), gold)
		m.Box(v3(x, y+s/2, z), v3(s+.06, s+.07, s*.16), gold)
	}
	house := func(m *MeshBuilder, x, y, z, w, h, d float32, c rl.Color) {
		m.Box(v3(x, y+h/2, z), v3(w, h, d), c)
		m.Box(v3(x, y+h+.35, z), v3(w+1, .7, d+1), white)
		m.Emis = 1
		for _, dx := range []float32{-.28, .28} {
			m.Box(v3(x+dx*w, y+h*.58, z+d/2+.04), v3(w*.22, h*.28, .12), gold)
		}
		m.Emis = 0
		m.Box(v3(x, y+h*.22, z+d/2+.08), v3(w*.17, h*.44, .15), dark)
		for i := 0; i < 9; i++ {
			m.Emis = 1
			m.Sphere(v3(x-w*.48+float32(i)*w*.12, y+h*.94, z+d/2+.16), .18, colors[i%4])
			m.Emis = 0
		}
	}
	snowman := func(m *MeshBuilder, x, y, z float32) {
		m.Sphere(v3(x, y+1, z), 1, white)
		m.Sphere(v3(x, y+2.4, z), .72, white)
		m.Sphere(v3(x, y+3.4, z), .5, white)
		m.Cylinder(v3(x, y+3.3, z+.42), v3(x, y+3.3, z+1), .15, 0, 8, rl.NewColor(245, 125, 35, 255), true)
		m.Cylinder(v3(x, y+3.8, z), v3(x, y+4.35, z), .43, .4, 10, dark, true)
	}
	for zone := 0; zone < 3; zone++ {
		zn := rb.Zones[zone]
		m := rb.Mesh()
		// Landmarks are placed by searching the outer ring, respecting all reservations.
		for k := 0; k < 72; k++ {
			a := float32(k) * .087266
			x := zn.C.X + cosf(a)*45
			z := zn.C.Y + sinf(a)*45
			radius := float32(9)
			if zone == 1 {
				radius = 14
			}
			if !rb.Free(x, z, radius) {
				continue
			}
			y := rb.H(x, z)
			if zone == 0 {
				tree(m, x, y, z, 35)
				rb.Pillar(x, z, 3, y+35)
				m.Emis = 1
				heroStar(m, x, y+35, z, 2.5, gold)
				m.Emis = 0
				for j := 0; j < 6; j++ {
					a := float32(j) * 1.047
					gift(m, x+cosf(a)*6, y, z+sinf(a)*6, 2.4, colors[j%4])
				}
			}
			if zone == 1 {
				house(m, x, y, z, 20, 13, 16, red)
				rb.Solid(x, z, 20, 16, 13.7, y)
				for _, dx := range []float32{-6, 6} {
					m.Cylinder(v3(x+dx, y+13, z), v3(x+dx, y+23, z), 1.3, 1.3, 10, brown, true)
					m.Cylinder(v3(x+dx, y+23, z), v3(x+dx, y+23.7, z), 1.6, 1.6, 10, white, true)
				}
			}
			if zone == 2 {
				house(m, x, y, z, 13, 10, 11, brown)
				rb.Solid(x, z, 13, 11, 10.7, y)
				for j := 0; j < 8; j++ {
					m.Sphere(v3(x-5.5+float32(j)*1.6, y+10, z+5.6), .55, white)
				}
			}
			rb.Reserve(x, z, radius)
			break
		}
		for k := 0; k < 32; k++ {
			a := float32(k) * .19635
			r := float32(36 + (k%3)*8)
			x := zn.C.X + cosf(a)*r
			z := zn.C.Y + sinf(a)*r
			size := float32(3.5)
			if k%4 == 0 {
				size = 5
			}
			if !rb.Free(x, z, size) {
				continue
			}
			y := rb.H(x, z)
			switch {
			case k%4 == 0:
				house(m, x, y, z, 6, 5, 5, []rl.Color{red, blue, brown}[zone])
				rb.Solid(x, z, 6, 5, 5.7, y)
			case k%4 == 1:
				tree(m, x, y, z, 9)
				rb.Pillar(x, z, 1, y+9)
			case k%4 == 2:
				snowman(m, x, y, z)
			default:
				gift(m, x, y, z, 2.8, colors[k%4])
			}
			rb.Reserve(x, z, size)
		}
		// Candy canes and oversized lollipops mark the gingerbread village.
		if zone == 2 {
			for k := 0; k < 16; k++ {
				a := float32(k) * .3927
				x := zn.C.X + cosf(a)*27
				z := zn.C.Y + sinf(a)*27
				if !rb.Free(x, z, 1.5) {
					continue
				}
				y := rb.H(x, z)
				m.Cylinder(v3(x, y, z), v3(x, y+5, z), .3, .3, 8, white, true)
				for j := 0; j < 5; j++ {
					m.Cylinder(v3(x, y+float32(j), z), v3(x, y+float32(j)+.35, z), .32, .32, 8, red, true)
				}
				m.Sphere(v3(x, y+5.5, z), 1.3, colors[k%4])
				rb.Reserve(x, z, 1.5)
			}
		}
		if zone == 1 {
			for k := 0; k < 3; k++ {
				x := zn.C.X - 25 + float32(k)*8
				z := zn.C.Y + 25
				if !rb.Free(x, z, 3) {
					continue
				}
				y := rb.H(x, z)
				if k == 0 {
					m.Sphere(v3(x, y+3, z), 2.3, brown)
					m.Sphere(v3(x, y+6, z), 1.6, brown)
					for _, dx := range []float32{-1.3, 1.3} {
						m.Sphere(v3(x+dx, y+7, z), .65, brown)
					}
				} else if k == 1 {
					m.Box(v3(x, y+3, z), v3(3, 4, 2), blue)
					m.Sphere(v3(x, y+6, z), 1.3, brown)
					m.Box(v3(x, y+7.4, z), v3(3, 1, 2.5), red)
				} else {
					m.Box(v3(x, y+2, z), v3(5, 3, 2.5), red)
					for _, dx := range []float32{-1.5, 1.5} {
						m.Sphere(v3(x+dx, y+.7, z+1.3), .7, dark)
					}
				}
				rb.Pillar(x, z, 2, y+7)
				rb.Reserve(x, z, 3)
			}
		}
		// A small sleigh or toy conveyor sits at the edge of each square.
		for k := 0; k < 24; k++ {
			a := float32(k) * .261799
			x, z := zn.C.X+cosf(a)*24, zn.C.Y+sinf(a)*24
			if !rb.Free(x, z, 3) {
				continue
			}
			y := rb.H(x, z)
			m.Box(v3(x, y+1, z), v3(4, .5, 2), red)
			for _, dz := range []float32{-1, 1} {
				m.Cylinder(v3(x-2, y+.3, z+dz), v3(x+2, y+.3, z+dz), .13, .13, 6, gold, true)
			}
			if zone == 1 {
				for j := 0; j < 3; j++ {
					gift(m, x-1+float32(j), y+1.3, z, .6, colors[j])
				}
			} else {
				m.Box(v3(x-1.5, y+1.7, z), v3(.3, 1.4, 2), red)
			}
			rb.Reserve(x, z, 3)
			break
		}
		rb.Add(m, true)
	}
	for gx := -240; gx <= 240; gx += 80 {
		for gz := -240; gz <= 240; gz += 80 {
			m := rb.Mesh()
			for i := 0; i < 18; i++ {
				x := float32(gx) + rb.Rng.Float32()*75
				z := float32(gz) + rb.Rng.Float32()*75
				if rb.ZoneAt(x, z) >= 0 || !rb.Free(x, z, 3.5) {
					continue
				}
				y := rb.H(x, z)
				h := float32(7) + rb.Rng.Float32()*5
				tree(m, x, y, z, h)
				rb.Pillar(x, z, 1, y+h)
				rb.Reserve(x, z, 3.5)
			}
			rb.Add(m, true)
		}
	}

}
