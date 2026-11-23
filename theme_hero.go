package main

import rl "github.com/gen2brain/raylib-go/raylib"

func (cb *CityBuilder) heroClear(x, z, r float32) bool {
	if !cb.themeClear(x, z, r) {
		return false
	}
	for _, b := range cb.W.Boxes {
		if x+r > b.MinX && x-r < b.MaxX && z+r > b.MinZ && z-r < b.MaxZ {
			return false
		}
	}
	return true
}

func (cb *CityBuilder) buildHeroDecor() {
	white := rl.NewColor(245, 244, 222, 255)
	red := rl.NewColor(217, 55, 49, 255)
	dark := rl.NewColor(54, 66, 75, 255)
	wood := rl.NewColor(151, 102, 62, 255)
	gold := rl.NewColor(255, 220, 99, 255)
	blue := rl.NewColor(80, 181, 206, 255)
	rock := rl.NewColor(126, 139, 139, 255)
	m := NewMB()
	// Search the island for a footprint clear of existing structures and roads.
	placed := false
	for radius := float32(25); radius <= 53 && !placed; radius += 4 {
		for i := 0; i < 32 && !placed; i++ {
			a := float32(i) * .19634954
			x := 160 + radius*cosf(a)
			z := 160 + radius*sinf(a)
			if !cb.heroClear(x, z, 4.8) {
				continue
			}
			y := cb.W.LandHeight(x, z)
			m.Cylinder(v3(x, y, z), v3(x, y+.8, z), 4.5, 4.5, 16, dark, true)
			for j := 0; j < 6; j++ {
				y0 := y + .8 + float32(j)*4
				r0 := float32(3.8) - float32(j)*.28
				c := white
				if j%2 == 1 {
					c = red
				}
				m.Cylinder(v3(x, y0, z), v3(x, y0+4, z), r0, r0-.28, 16, c, true)
			}
			m.Cylinder(v3(x, y+24.8, z), v3(x, y+25.3, z), 2.9, 2.9, 16, dark, true)
			m.Emis = .95
			m.Cylinder(v3(x, y+25.3, z), v3(x, y+28.2, z), 1.9, 1.9, 16, gold, true)
			m.Emis = 0
			for k := 0; k < 8; k++ {
				a := float32(k) * .785398
				dx, dz := 2*cosf(a), 2*sinf(a)
				m.Cylinder(v3(x+dx, y+25.3, z+dz), v3(x+dx, y+28.2, z+dz), .10, .10, 6, dark, true)
			}
			m.Cylinder(v3(x, y+28.2, z), v3(x, y+30, z), 2.5, 0, 16, red, true)
			m.Box(v3(x, y+1.8, z+3.77), v3(1.2, 2.5, .18), dark)
			cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 4.5, Top: y + 30})
			placed = true
		}
	}
	cb.addChunk(m, true)
	// Small maritime groups, each checked against the actual existing world.
	palette := []rl.Color{rl.NewColor(239, 153, 104, 255), blue, rl.NewColor(244, 211, 111, 255), rl.NewColor(153, 202, 174, 255)}
	for gx := 0; gx < 4; gx++ {
		for gz := 0; gz < 4; gz++ {
			m = NewMB()
			for i := 0; i < 3; i++ {
				for j := 0; j < 3; j++ {
					x := float32(110 + gx*25 + i*8)
					z := float32(110 + gz*25 + j*8)
					d := sq(x-160) + sq(z-160)
					if d > sq(55) || d < sq(18) {
						continue
					}
					k := (gx*17 + gz*11 + i*5 + j*3) % 8
					r := float32(1.6)
					if k == 0 {
						r = 3.4
					}
					if !cb.heroClear(x, z, r) {
						continue
					}
					y := cb.W.LandHeight(x, z)
					switch k {
					case 0: // fisherman's cottage with a pitched roof
						c := palette[(gx+gz+i+j)%len(palette)]
						m.Box(v3(x, y+2, z), v3(5, 4, 4), c)
						p1, p2 := v3(x-2.8, y+4, z-2.3), v3(x-2.8, y+4, z+2.3)
						p3, p4 := v3(x+2.8, y+4, z-2.3), v3(x+2.8, y+4, z+2.3)
						ridge1, ridge2 := v3(x, y+5.8, z-2.3), v3(x, y+5.8, z+2.3)
						m.Quad(p1, p2, ridge2, ridge1, v3(-1, 1, 0), red)
						m.Quad(ridge1, ridge2, p4, p3, v3(1, 1, 0), red)
						m.Tri(p1, ridge1, p3, v3(0, 0, -1), c)
						m.Tri(p2, p4, ridge2, v3(0, 0, 1), c)
						m.Box(v3(x, y+1.25, z+2.03), v3(1, 2.5, .1), wood)
						for _, dx := range []float32{-1.6, 1.6} {
							m.Box(v3(x+dx, y+2.5, z+2.04), v3(.8, .9, .1), white)
							m.Box(v3(x+dx, y+2.5, z+2.1), v3(.58, .68, .08), blue)
						}
						cb.W.Boxes = append(cb.W.Boxes, BoxCollider{MinX: x - 2.5, MaxX: x + 2.5, MinZ: z - 2, MaxZ: z + 2, Top: y + 4, Roof: RoofRidgeZ, RoofH: 1.8})
					case 1: // beached boat, with gunwales and benches
						m.Ellipsoid(v3(x, y+.48, z), v3(1.15, .55, 2.5), 6, 12, red)
						m.Ellipsoid(v3(x, y+.85, z), v3(.9, .18, 2), 6, 12, wood)
						for _, dz := range []float32{-.7, .7} {
							m.Box(v3(x, y+1, z+dz), v3(1.6, .14, .35), white)
						}
						cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 1.2, Top: y + 1})
					case 2: // fishing net hung between poles
						for _, dx := range []float32{-1.5, 1.5} {
							m.Cylinder(v3(x+dx, y, z), v3(x+dx, y+3, z), .1, .08, 6, wood, true)
						}
						for n := 0; n <= 6; n++ {
							dx := float32(n)*.5 - 1.5
							m.Cylinder(v3(x+dx, y+.4, z), v3(x+dx, y+2.8, z), .025, .025, 4, white, false)
						}
						for n := 0; n < 5; n++ {
							h := y + .5 + float32(n)*.55
							m.Cylinder(v3(x-1.5, h, z), v3(x+1.5, h, z), .025, .025, 4, white, false)
						}
					case 3: // gull on a post
						m.Cylinder(v3(x, y, z), v3(x, y+3, z), .16, .12, 7, wood, true)
						m.Ellipsoid(v3(x, y+3.25, z), v3(.6, .25, .3), 6, 10, white)
						m.Sphere(v3(x+.5, y+3.4, z), .22, white)
						m.Cylinder(v3(x+.65, y+3.35, z), v3(x+.95, y+3.32, z), .11, 0, 6, gold, true)
					case 4: // buoy
						m.Sphere(v3(x, y+.6, z), .6, red)
						m.Cylinder(v3(x, y+.6, z), v3(x, y+1.6, z), .12, .06, 8, white, true)
					case 5: // anchor planted in sand
						m.Cylinder(v3(x, y+.2, z), v3(x, y+2.4, z), .15, .15, 8, dark, true)
						m.Cylinder(v3(x-.8, y+1.9, z), v3(x+.8, y+1.9, z), .12, .12, 8, dark, true)
						for _, s := range []float32{-1, 1} {
							m.Cylinder(v3(x, y+.3, z), v3(x+s*1.1, y+.8, z), .16, .12, 8, dark, true)
							m.Sphere(v3(x+s*1.1, y+.8, z), .2, dark)
						}
						m.Sphere(v3(x, y+2.5, z), .22, dark)
					default: // shoreline stones
						m.Ellipsoid(v3(x, y+.55, z), v3(1.5, .85, 1.1), 6, 9, rock)
						cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 1.2, Top: y + 1.4})
					}
				}
			}
			cb.addChunk(m, true)
		}
	}
}

func heroStar(m *MeshBuilder, x, y, z, r float32, col rl.Color) {
	c := v3(x, y, z)
	for i := 0; i < 10; i++ {
		a := float32(i)*.6283185 + 1.5707963
		b := float32(i+1)*.6283185 + 1.5707963
		ra, rb := r, r
		if i%2 == 1 {
			ra = r * .43
		}
		if (i+1)%2 == 1 {
			rb = r * .43
		}
		p := v3(x+cosf(a)*ra, y+sinf(a)*ra, z)
		q := v3(x+cosf(b)*rb, y+sinf(b)*rb, z)
		m.Tri(c, p, q, v3(0, 0, 1), col)
		m.Tri(c, p, q, v3(0, 0, -1), col)
	}
}
