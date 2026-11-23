package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

func init() { registerRealmDecor("eroi_2", eroiDecor2) }

func eroiDecor2(rb *RealmBuilder) {
	c := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	cream, blue, gold := c(232, 216, 181), c(54, 112, 155), c(245, 191, 65)
	dark, red, stone := c(56, 67, 78), c(226, 70, 57), c(140, 146, 152)
	// The two landmarks are tried at several positions on the outer rim.
	place := func(points [][2]float32, r float32, f func(*MeshBuilder, float32, float32, float32)) {
		for _, p := range points {
			x, z := p[0], p[1]
			if rb.Free(x, z, r) {
				y := rb.H(x, z)
				m := rb.Mesh()
				f(m, x, y, z)
				rb.Reserve(x, z, r)
				rb.Add(m, true)
				return
			}
		}
	}
	place([][2]float32{{-210, 185}, {-202, 191}, {-210, 115}, {-115, 195}}, 23, func(m *MeshBuilder, x, y, z float32) {
		// Moored vessel with hull, decks, bridge and two funnels.
		m.BeveledBox(v3(x, y+3, z), v3(14, 6, 37), 1.5, dark)
		m.Box(v3(x, y+6.5, z), v3(13, 1, 35), cream)
		m.Box(v3(x, y+9, z+5), v3(10, 5, 13), blue)
		m.Box(v3(x, y+12, z+5), v3(11, 1, 14), cream)
		for _, dz := range []float32{-10, -4, 2, 8, 14} {
			for _, s := range []float32{-1, 1} {
				m.Box(v3(x+s*7.05, y+3.5, z+dz), v3(.2, 1.1, 1.4), gold)
			}
		}
		for _, dz := range []float32{-5, 2} {
			m.Cylinder(v3(x, y+7, z+dz), v3(x, y+15, z+dz), 1.4, 1.2, 10, red, true)
		}
		rb.Solid(x, z, 14, 37, 13, y)
	})
	for _, p := range [][2]float32{{-200, 115}, {-118, 190}, {-203, 165}, {-125, 115}, {-190, 205}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 9) {
			continue
		}
		y := rb.H(x, z)
		m := rb.Mesh()
		m.Box(v3(x, y+5, z), v3(13, 10, 12), cream)
		m.Box(v3(x, y+10.6, z), v3(14, 1.2, 13), blue)
		for i := -1; i <= 1; i++ {
			m.Box(v3(x+float32(i)*3.5, y+4, z-6.1), v3(2, 5, .2), blue)
		}
		rb.Solid(x, z, 13, 12, 11, y)
		rb.Reserve(x, z, 9)
		rb.Add(m, true)
	}
	for _, p := range [][2]float32{{-195, 108}, {-112, 165}, {-190, 200}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 5) {
			continue
		}
		y := rb.H(x, z)
		m := rb.Mesh()
		m.Box(v3(x, y+10, z), v3(2, 20, 2), gold)
		m.Box(v3(x+5, y+20, z), v3(12, 1.2, 1.2), gold)
		m.Cylinder(v3(x+10, y+20, z), v3(x+10, y+12, z), .12, .12, 6, dark, false)
		rb.Pillar(x, z, 1.3, y+20)
		rb.Reserve(x, z, 5)
		rb.Add(m, true)
	}
	for _, p := range [][2]float32{{-180, 185}, {-177, 190}, {-130, 135}, {-135, 138}, {-190, 130}, {-185, 130}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 2.5) {
			continue
		}
		y := rb.H(x, z)
		m := rb.Mesh()
		m.Box(v3(x, y+1.5, z), v3(3, 3, 3), c(185, 116, 57))
		m.Box(v3(x, y+1.5, z-1.52), v3(3.1, .25, .1), gold)
		rb.Solid(x, z, 3, 3, 3, y)
		rb.Reserve(x, z, 2.5)
		rb.Add(m, true)
	}
	place([][2]float32{{205, 190}, {205, 110}, {115, 195}, {115, 110}}, 12, func(m *MeshBuilder, x, y, z float32) {
		m.Cylinder(v3(x, y, z), v3(x, y+10, z), 9, 9, 20, cream, true)
		m.Ellipsoid(v3(x, y+10, z), v3(9, 7, 9), 10, 20, blue)
		m.Box(v3(x, y+13, z), v3(1, 5, 15), dark)
		rb.Pillar(x, z, 9, y+17)
	})
	for _, p := range [][2]float32{{205, 117}, {210, 180}, {115, 190}, {120, 110}, {185, 205}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 5) {
			continue
		}
		y := rb.H(x, z)
		m := rb.Mesh()
		m.Cylinder(v3(x, y, z), v3(x, y+36, z), 2.3, .5, 8, stone, true)
		for h := float32(8); h < 36; h += 7 {
			m.Box(v3(x, y+h, z), v3(6, .5, .5), dark)
		}
		m.Emis = 1
		m.Sphere(v3(x, y+37, z), 1.3, red)
		m.Emis = 0
		rb.Pillar(x, z, 2.4, y+36)
		rb.Reserve(x, z, 5)
		rb.Add(m, true)
	}
	for _, p := range [][2]float32{{130, 185}, {180, 120}, {200, 135}, {130, 125}, {185, 190}, {115, 160}} {
		x, z := p[0], p[1]
		if !rb.Free(x, z, 4) {
			continue
		}
		y := rb.H(x, z)
		m := rb.Mesh()
		m.Ellipsoid(v3(x, y+2, z), v3(4, 3, 3), 6, 10, stone)
		rb.Pillar(x, z, 3, y+5)
		rb.Reserve(x, z, 4)
		rb.Add(m, true)
	}
	// Neighborhood blocks and parks, divided into small chunks for culling.
	for gx := -3; gx <= 3; gx++ {
		for gz := -3; gz <= 3; gz++ {
			m := rb.Mesh()
			for i := 0; i < 12; i++ {
				x := float32(gx*75) + float32(rb.Rng.Intn(65)-32)
				z := float32(gz*75) + float32(rb.Rng.Intn(65)-32)
				if rb.ZoneAt(x, z) != -1 {
					continue
				}
				park := (gx+gz+i)%3 == 0
				r := float32(8)
				if park {
					r = 3
				}
				if !rb.Free(x, z, r) {
					continue
				}
				y := rb.H(x, z)
				if park {
					m.Cylinder(v3(x, y, z), v3(x, y+6, z), .6, .4, 7, c(126, 85, 48), true)
					m.Ellipsoid(v3(x, y+8, z), v3(3.5, 4, 3.5), 7, 10, c(65, 155, 85))
					rb.Pillar(x, z, .7, y+6)
				} else {
					h := float32(13 + rb.Rng.Intn(15))
					m.Box(v3(x, y+h/2, z), v3(10, h, 10), cream)
					m.Box(v3(x, y+h+1, z), v3(8, 2, 8), blue)
					m.Box(v3(x, y+h+2.5, z), v3(5, 1, 5), gold)
					for row := float32(4); row < h-1; row += 4 {
						for k := -1; k <= 1; k++ {
							m.Box(v3(x+float32(k)*3, y+row, z-5.06), v3(1.6, 2, .12), blue)
						}
					}
					rb.Solid(x, z, 10, 10, h+3, y)
				}
				rb.Reserve(x, z, r)
			}
			rb.Add(m, true)
		}
	}
	// Small dockside cargo and marker lamps keep the combat areas legible.
	for _, center := range [][2]float32{{-160, 150}, {160, 150}} {
		m := rb.Mesh()
		for i := 0; i < 18; i++ {
			a := float64(i) * math.Pi * 2 / 18
			x := center[0] + float32(math.Cos(a))*27
			z := center[1] + float32(math.Sin(a))*27
			if !rb.Free(x, z, 1.2) {
				continue
			}
			y := rb.H(x, z)
			if center[0] < 0 {
				m.Box(v3(x, y+.65, z), v3(1.3, 1.3, 1.3), gold)
			} else {
				m.Cylinder(v3(x, y, z), v3(x, y+2, z), .15, .15, 6, dark, true)
				m.Emis = .8
				m.Sphere(v3(x, y+2.2, z), .35, red)
				m.Emis = 0
			}
			rb.Reserve(x, z, 1.2)
		}
		rb.Add(m, true)
	}
	// Airborne balloons and a dirigible do not occupy the ground.
	sky := rb.Mesh()
	for i, p := range [][3]float32{{-230, 58, -40}, {-70, 72, 220}, {80, 62, -210}, {230, 76, 30}, {-40, 48, 70}} {
		x, y, z := p[0], p[1], p[2]
		col := red
		if i%2 == 0 {
			col = gold
		}
		sky.Ellipsoid(v3(x, y, z), v3(5, 7, 5), 10, 16, col)
		sky.Box(v3(x, y-10, z), v3(3, 2, 3), cream)
		for _, s := range []float32{-1, 1} {
			sky.Cylinder(v3(x+s, y-9, z), v3(x+s*2, y-5, z), .1, .1, 5, dark, false)
		}
	}
	sky.Ellipsoid(v3(15, 80, 180), v3(18, 5, 5), 10, 20, cream)
	sky.Box(v3(15, 73, 180), v3(7, 3, 3), blue)
	sky.Box(v3(30, 80, 180), v3(4, 9, .6), red)
	rb.Add(sky, false)
}
