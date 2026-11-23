package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

func init() { registerRealmDecor("panze", panzeDecor) }

func panzeDecor(rb *RealmBuilder) {
	c := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	steel, white := c(94, 111, 126), c(242, 246, 239)
	glass := c(107, 205, 224)
	colors := []rl.Color{c(165, 183, 196), c(197, 178, 165), c(133, 158, 184), c(200, 194, 170), c(184, 117, 94)}
	line := func(m *MeshBuilder, a, b rl.Vector3, r float32, col rl.Color) { m.Cylinder(a, b, r, r, 5, col, true) }
	for zone := 0; zone < 3; zone++ {
		zn := rb.Zones[zone]
		m := rb.Mesh()
		var px, pz, ph float32
		hasPrev := false
		// Silueta orașului, cu clădiri mari doar către marginea zonei.
		for i := 0; i < 24; i++ {
			a := float64(i)*2*math.Pi/24 + 0.09
			dist := float32(46 + i%3*8)
			x := zn.C.X + float32(math.Cos(a))*dist
			z := zn.C.Y + float32(math.Sin(a))*dist
			if zone == 2 {
				continue
			}
			w := float32(10 + i%3*2)
			d := float32(10 + (i+1)%3*2)
			radius := sqrtf(w*w+d*d)/2 + 1
			if !rb.Free(x, z, radius) {
				continue
			}
			y := rb.H(x, z)
			h := float32(25 + (i*13)%42)
			if zone == 1 {
				h = float32(12 + (i*7)%16)
			}
			if zone == 0 && i == 2 {
				h = 90
			}
			m.Box(v3(x, y+h/2, z), v3(w, h, d), colors[i%len(colors)])
			if hasPrev && sqrtf(sq(x-px)+sq(z-pz)) < 42 {
				for j := 0; j < 5; j++ {
					t := float32(j) / 4
					line(m, v3(px, y+ph*.55+t*3, pz), v3(x, y+h*.55+t*3, z), .035, white)
				}
				for j := 1; j < 5; j++ {
					t := float32(j) / 5
					xx := px + (x-px)*t
					zz := pz + (z-pz)*t
					yy := y + ph*.55 + (h-ph)*.55*t
					line(m, v3(xx, yy, zz), v3(xx, yy+3, zz), .035, white)
				}
			}
			px, pz, ph, hasPrev = x, z, h, true
			rb.Solid(x, z, w, d, h, y)
			rb.Reserve(x, z, radius)
			m.Box(v3(x, y+h+.18, z), v3(w+.5, .36, d+.5), steel)
			for wy := float32(3); wy < h-1; wy += 3.1 {
				for wx := -w/2 + 1.5; wx < w/2-1; wx += 2.5 {
					m.Box(v3(x+wx, y+wy, z+d/2+.04), v3(1.35, 1.8, .1), glass)
					m.Box(v3(x+wx, y+wy, z-d/2-.04), v3(1.35, 1.8, .1), glass)
				}
				for wz := -d/2 + 1.5; wz < d/2-1; wz += 2.5 {
					m.Box(v3(x+w/2+.04, y+wy, z+wz), v3(.1, 1.8, 1.35), glass)
					m.Box(v3(x-w/2-.04, y+wy, z+wz), v3(.1, 1.8, 1.35), glass)
				}
			}
			m.Box(v3(x-2, y+h+.8, z), v3(2, 1.6, 2), steel)
			m.Cylinder(v3(x+2, y+h, z), v3(x+2, y+h+3, z), 1, 1, 10, colors[1], true)
			line(m, v3(x, y+h, z-2), v3(x, y+h+6, z-2), .09, steel)
			if h == 90 {
				m.Emis = 1
				m.Sphere(v3(x, y+h+6, z-2), 1.1, c(255, 230, 115))
				m.Emis = 0
			}
			// Scară de incendiu pe fațada laterală.
			for k := float32(3); k < h-1; k += 3 {
				m.Box(v3(x+w/2+.65, y+k, z), v3(1.3, .16, 3), steel)
				line(m, v3(x+w/2+1.1, y+k, z-1), v3(x+w/2+1.1, y+k+3, z+1), .06, steel)
			}
		}
		// Mobilier urban, lăsând loc de luptă și circulație.
		for i := 0; i < 32; i++ {
			a := float64(i) * 2 * math.Pi / 32
			dist := float32(23 + i%3*3)
			x := zn.C.X + float32(math.Cos(a))*dist
			z := zn.C.Y + float32(math.Sin(a))*dist
			if !rb.Free(x, z, 1.4) {
				continue
			}
			y := rb.H(x, z)
			rb.Reserve(x, z, 1.4)
			if zone == 2 && i%3 == 0 {
				m.Cylinder(v3(x, y, z), v3(x, y+4, z), .35, .25, 8, c(123, 79, 47), true)
				m.Ellipsoid(v3(x, y+5, z), v3(2.4, 2.5, 2.4), 8, 10, c(69, 169, 87))
				rb.Pillar(x, z, .35, y+4)
				continue
			}
			switch i % 4 {
			case 0:
				m.Box(v3(x, y+.6, z), v3(2.5, .2, .8), c(174, 102, 55))
				m.Box(v3(x, y+1.15, z-.35), v3(2.5, 1, .15), c(174, 102, 55))
			case 1:
				m.Cylinder(v3(x, y, z), v3(x, y+1, z), .35, .3, 8, c(230, 63, 53), true)
				m.Sphere(v3(x, y+1, z), .32, c(230, 63, 53))
			case 2:
				line(m, v3(x, y, z), v3(x, y+4, z), .12, steel)
				m.Box(v3(x, y+3.5, z), v3(.7, 1.5, .6), steel)
				for k := 0; k < 3; k++ {
					m.Sphere(v3(x, y+3+float32(k)*.46, z+.32), .15, []rl.Color{c(239, 65, 55), c(255, 206, 65), c(71, 215, 119)}[k])
				}
			case 3:
				m.Box(v3(x, y+1.2, z), v3(2, 2.4, 1.7), c(245, 190, 74))
				m.Box(v3(x, y+2.5, z), v3(2.4, .25, 2), c(227, 87, 94))
			}
		}
		if zone == 2 { // Arc de piatră cu deschidere, lac și carusel.
			for i := 0; i < 20; i++ {
				a := float64(i) * 2 * math.Pi / 20
				x := zn.C.X + float32(math.Cos(a))*43
				z := zn.C.Y + float32(math.Sin(a))*43
				if !rb.Free(x, z, 6) {
					continue
				}
				y := rb.H(x, z)
				for _, dx := range []float32{-4, 4} {
					m.Box(v3(x+dx, y+4, z), v3(2, 8, 2), c(205, 192, 164))
					rb.Solid(x+dx, z, 2, 2, 8, y)
				}
				for j := 0; j < 12; j++ {
					t := math.Pi * float64(j) / 11
					m.Box(v3(x+float32(math.Cos(t))*4, y+8+float32(math.Sin(t))*3, z), v3(1.4, 1.4, 2), c(205, 192, 164))
				}
				rb.Reserve(x, z, 6)
				break
			}
			for i := 0; i < 20; i++ {
				a := float64(i) * 2 * math.Pi / 20
				x := zn.C.X + float32(math.Cos(a))*34
				z := zn.C.Y + float32(math.Sin(a))*34
				if !rb.Free(x, z, 7) {
					continue
				}
				y := rb.H(x, z)
				m.Ellipsoid(v3(x, y+.08, z), v3(6, .12, 4), 6, 16, c(80, 186, 218))
				rb.Reserve(x, z, 7)
				break
			}
			for i := 0; i < 20; i++ {
				a := float64(i) * 2 * math.Pi / 20
				x := zn.C.X + float32(math.Cos(a))*38
				z := zn.C.Y + float32(math.Sin(a))*38
				if !rb.Free(x, z, 5) {
					continue
				}
				y := rb.H(x, z)
				m.Cylinder(v3(x, y, z), v3(x, y+.4, z), 4, 4, 16, c(237, 153, 84), true)
				m.Cylinder(v3(x, y+.4, z), v3(x, y+5, z), .18, .18, 8, steel, true)
				m.Cylinder(v3(x, y+4, z), v3(x, y+5.8, z), 4, 0, 16, c(238, 91, 117), true)
				for j := 0; j < 8; j++ {
					t := float64(j) * math.Pi / 4
					xx := x + float32(math.Cos(t))*2.7
					zz := z + float32(math.Sin(t))*2.7
					line(m, v3(xx, y+.4, zz), v3(xx, y+4, zz), .08, white)
					m.Sphere(v3(xx, y+1.4, zz), .45, colors[j%len(colors)])
				}
				rb.Reserve(x, z, 5)
				break
			}
		}
		if zone == 1 {
			for i := 0; i < 24; i++ {
				a := float64(i) * math.Pi / 12
				x := zn.C.X + float32(math.Cos(a))*40
				z := zn.C.Y + float32(math.Sin(a))*40
				if !rb.Free(x, z, 10) {
					continue
				}
				y := rb.H(x, z)
				m.Box(v3(x, y+.04, z), v3(18, .1, 7), c(77, 177, 207))
				m.Box(v3(x, y+1.2, z), v3(3, .35, 12), c(179, 130, 91))
				for _, dx := range []float32{-1.8, 1.8} {
					for _, dz := range []float32{-5, 5} {
						line(m, v3(x+dx, y+1, z+dz), v3(x+dx, y+5, z+dz), .14, steel)
					}
					line(m, v3(x+dx, y+5, z-5), v3(x+dx, y+5, z+5), .10, steel)
					for j := -4; j <= 4; j += 2 {
						line(m, v3(x+dx, y+1.4, z+float32(j)), v3(x+dx, y+5, z+float32(j)), .05, white)
					}
				}
				rb.Reserve(x, z, 10)
				break
			}
		}
		if zone != 2 {
			for i := 0; i < 12; i++ {
				a := float64(i) * math.Pi / 6
				x := zn.C.X + float32(math.Cos(a))*34
				z := zn.C.Y + float32(math.Sin(a))*34
				if !rb.Free(x, z, 2) {
					continue
				}
				y := rb.H(x, z)
				line(m, v3(x, y, z), v3(x, y+6, z), .18, steel)
				m.Box(v3(x, y+5, z), v3(4, 2.4, .2), colors[(i+2)%len(colors)])
				m.Box(v3(x, y+5, z+.12), v3(3.4, 1.8, .1), c(241, 125, 165))
				rb.Reserve(x, z, 2)
			}
		}
		rb.Add(m, true)
	}
}
