package main

import rl "github.com/gen2brain/raylib-go/raylib"

func (cb *CityBuilder) buildSpiderDecor() {
	wood := rl.NewColor(145, 82, 39, 255)
	dark := rl.NewColor(85, 48, 28, 255)
	gold := rl.NewColor(255, 199, 45, 255)
	cream := rl.NewColor(255, 238, 183, 255)
	green := rl.NewColor(48, 154, 79, 255)
	red := rl.NewColor(226, 75, 65, 255)
	blue := rl.NewColor(53, 167, 210, 255)
	clear := func(x, z, r float32) bool {
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
	box := func(x, z, w, d, top float32) {
		cb.W.Boxes = append(cb.W.Boxes, BoxCollider{MinX: x - w/2, MaxX: x + w/2, MinZ: z - d/2, MaxZ: z + d/2, Top: top, Roof: RoofFlat})
	}
	// The ship is placed in the first sufficiently clear part of the island rim.
	ship := false
	for z := float32(-210); z <= -110; z += 2 {
		for x := float32(110); x <= 210; x += 2 {
			if ship || !clear(x, z, 4.5) {
				continue
			}
			ship = true
			y := cb.W.LandHeight(x, z)
			m := NewMB()
			m.BeveledBox(v3(x, y+1.3, z), v3(8, 2.6, 4), .6, wood)
			m.Box(v3(x, y+2.65, z), v3(7.5, .25, 3.6), dark)
			for _, dz := range []float32{-1.8, 1.8} {
				m.Box(v3(x, y+3.15, z+dz), v3(7.5, 1, .35), wood)
			}
			for _, dx := range []float32{-2, 2} {
				m.Cylinder(v3(x+dx, y+2.7, z), v3(x+dx, y+15, z), .28, .18, 9, dark, true)
				m.Cylinder(v3(x+dx-1.8, y+12, z), v3(x+dx+1.8, y+12, z), .13, .13, 7, dark, true)
				for _, side := range []float32{-1, 1} {
					m.Quad(v3(x+dx-1.6, y+11.8, z), v3(x+dx+1.6, y+11.8, z), v3(x+dx+1.3, y+6, z), v3(x+dx-1.3, y+6, z), v3(0, 0, side), cream)
				}
				m.Box(v3(x+dx+1, y+14.4, z), v3(2, .8, .12), red)
			}
			box(x, z, 8, 4, y+3.7)
			cb.addChunk(m, true)
		}
	}
	// Small groups give the shore a playful nautical character.
	for gx := 0; gx < 4; gx++ {
		for gz := 0; gz < 4; gz++ {
			m := NewMB()
			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					x := float32(112 + gx*25 + i*6)
					z := float32(-208 + gz*25 + j*6)
					if sq(x-160)+sq(z+160) > 56*56 || !clear(x, z, 1.5) {
						continue
					}
					y := cb.W.LandHeight(x, z)
					k := (gx*7 + gz*3 + i*5 + j) % 6
					switch k {
					case 0: // palm
						m.Cylinder(v3(x, y, z), v3(x+.6, y+6, z), .38, .23, 7, wood, true)
						for a := 0; a < 7; a++ {
							t := float32(a) * .8976
							m.Cylinder(v3(x+.6, y+6, z), v3(x+.6+cosf(t)*2.6, y+5.2, z+sinf(t)*2.6), .38, .05, 5, green, true)
						}
						cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: .5, Top: y + 6})
					case 1: // treasure chest
						m.BeveledBox(v3(x, y+.55, z), v3(1.6, 1.1, 1), .12, wood)
						m.Box(v3(x, y+1.12, z), v3(1.7, .25, 1.1), dark)
						m.Box(v3(x, y+.65, z+.52), v3(.22, .7, .08), gold)
						m.Sphere(v3(x, y+1.3, z), .23, gold)
						box(x, z, 1.6, 1, y+1.25)
					case 2: // barrel and hoops
						m.Cylinder(v3(x, y, z), v3(x, y+1.5, z), .65, .65, 10, wood, true)
						for _, h := range []float32{.25, 1.2} {
							m.Cylinder(v3(x, y+h, z), v3(x, y+h+.1, z), .68, .68, 10, dark, true)
						}
						cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: .65, Top: y + 1.5})
					case 3: // toy cannon
						m.Box(v3(x, y+.45, z), v3(1.8, .5, 1), wood)
						m.Cylinder(v3(x-.6, y+1, z), v3(x+1.2, y+1.25, z), .4, .32, 9, dark, true)
						for _, dz := range []float32{-.65, .65} {
							m.Cylinder(v3(x, y+.45, z+dz-.12), v3(x, y+.45, z+dz+.12), .45, .45, 9, gold, true)
						}
						box(x, z, 2, 1.5, y+1.4)
					case 4: // fishing net stretched on posts
						for _, dx := range []float32{-1.2, 1.2} {
							m.Cylinder(v3(x+dx, y, z), v3(x+dx, y+2.4, z), .09, .09, 5, wood, true)
						}
						for n := 0; n <= 4; n++ {
							q := float32(n) * .6
							m.Cylinder(v3(x-1.2+q, y+.4, z), v3(x-1.2+q, y+2.3, z), .025, .025, 4, cream, false)
							m.Cylinder(v3(x-1.2, y+.4+q*.75, z), v3(x+1.2, y+.4+q*.75, z), .025, .025, 4, cream, false)
						}
					case 5: // bright pennant
						m.Cylinder(v3(x, y, z), v3(x, y+4, z), .1, .1, 6, wood, true)
						m.Tri(v3(x, y+3.9, z), v3(x+1.8, y+3.4, z), v3(x, y+2.9, z), v3(0, 0, 1), blue)
						m.Tri(v3(x, y+3.9, z), v3(x+1.8, y+3.4, z), v3(x, y+2.9, z), v3(0, 0, -1), red)
					}
				}
			}
			cb.addChunk(m, true)
		}
	}
	// A short wooden landing among the clear shore spaces.
	for _, z := range []float32{-195, -125, -185, -135} {
		found := false
		for _, x := range []float32{140, 180, 150, 170} {
			if !clear(x, z, 5) {
				continue
			}
			y := cb.W.LandHeight(x, z)
			m := NewMB()
			for n := 0; n < 8; n++ {
				dz := float32(n) - 3.5
				m.Box(v3(x, y+.55, z+dz), v3(3, .22, .9), wood)
			}
			for _, dx := range []float32{-1.2, 1.2} {
				for _, dz := range []float32{-3, 3} {
					m.Cylinder(v3(x+dx, y-.3, z+dz), v3(x+dx, y+.5, z+dz), .18, .18, 6, dark, true)
				}
			}
			box(x, z, 3, 8, y+.7)
			cb.addChunk(m, true)
			found = true
			break
		}
		if found {
			break
		}
	}
}
