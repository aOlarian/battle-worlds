package main

import rl "github.com/gen2brain/raylib-go/raylib"

func (cb *CityBuilder) buildBatDecor() {
	pink := rl.NewColor(250, 102, 153, 255)
	orange := rl.NewColor(255, 148, 72, 255)
	purple := rl.NewColor(166, 104, 222, 255)
	cream := rl.NewColor(255, 224, 187, 255)
	aqua := rl.NewColor(123, 224, 233, 255)
	yellow := rl.NewColor(255, 216, 79, 255)
	colors := []rl.Color{pink, orange, purple}
	// The aquarium uses an open glass framework so its inhabitants remain visible.
	candidates := [][2]float32{{-190, 137}, {-132, 185}, {-188, 185}, {-130, 135}, {-160, 195}, {-195, 160}}
	for dx := -48; dx <= 48; dx += 3 {
		for dz := -48; dz <= 48; dz += 3 {
			if dx*dx+dz*dz < 48*48 {
				candidates = append(candidates, [2]float32{float32(-160 + dx), float32(160 + dz)})
			}
		}
	}
	dome := false
	for _, p := range candidates {
		x, z := p[0], p[1]
		if !cb.heroClear(x, z, 7) {
			continue
		}
		y := cb.W.LandHeight(x, z)
		m := NewMB()
		m.Cylinder(v3(x, y, z), v3(x, y+.6, z), 6, 6, 24, cream, true)
		m.Tube(v3(x, y+.6, z), 5.7, 6, .45, 24, aqua)
		// Meridians and latitude rings suggest transparent glass without hiding fish.
		for j := 0; j < 12; j++ {
			a := float32(j) * 6.2831853 / 12
			for k := 0; k < 8; k++ {
				t := float32(k) * 1.5707963 / 8
				u := float32(k+1) * 1.5707963 / 8
				m.Cylinder(v3(x+5.8*cosf(a)*cosf(t), y+1+6*sinf(t), z+5.8*sinf(a)*cosf(t)), v3(x+5.8*cosf(a)*cosf(u), y+1+6*sinf(u), z+5.8*sinf(a)*cosf(u)), .08, .08, 5, aqua, true)
			}
		}
		for k := 0; k < 3; k++ {
			t := float32(k) * .38
			m.Tube(v3(x, y+1+6*sinf(t), z), 5.8*cosf(t)-.06, 5.8*cosf(t)+.06, .10, 24, aqua)
		}
		for i := 0; i < 7; i++ {
			a := float32(i) * 6.2831853 / 7
			fx := x + 3*cosf(a)
			fz := z + 3*sinf(a)
			fy := y + 2 + float32(i%3)
			m.Ellipsoid(v3(fx, fy, fz), v3(.8, .45, .3), 8, 12, colors[i%3])
			m.Tri(v3(fx-.6, fy, fz), v3(fx-1.2, fy+.5, fz), v3(fx-1.2, fy-.5, fz), v3(0, 0, 1), yellow)
			m.Sphere(v3(fx+.4, fy+.12, fz+.28), .09, rl.Black)
		}
		cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 6, Top: y + 7})
		cb.addChunk(m, true)
		dome = true
		break
	}
	_ = dome
	for gx := -2; gx <= 2; gx++ {
		for gz := -2; gz <= 2; gz++ {
			m := NewMB()
			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					x := float32(-160 + gx*22 + i*6 - 9)
					z := float32(160 + gz*22 + j*6 - 9)
					if sq(x+160)+sq(z-160) > 55*55 {
						continue
					}
					k := (i*7 + j*3 + (gx+2)*5 + (gz+2)*11) % 7
					r := float32(1.2)
					if k < 3 {
						r = 2
					}
					if !cb.heroClear(x, z, r) {
						continue
					}
					y := cb.W.LandHeight(x, z)
					c := colors[(i+j+gx+gz+20)%3]
					m.Disc(v3(x, y+.05, z), 0, 2.2, 16, []rl.Color{cream, rl.NewColor(255, 192, 172, 255), rl.NewColor(202, 184, 235, 255)}[(i+j+20)%3])
					switch {
					case k < 3:
						m.Cylinder(v3(x, y, z), v3(x, y+4.5, z), .65, .42, 9, c, true)
						for b := 0; b < 5; b++ {
							a := float32(b) * 6.2831853 / 5
							dx, dz := cosf(a), sinf(a)
							m.Cylinder(v3(x, y+1.4+float32(b%2), z), v3(x+dx*1.5, y+4.2, z+dz*1.5), .38, .23, 8, c, true)
							m.Cylinder(v3(x+dx*1.5, y+4.2, z+dz*1.5), v3(x+dx*1.65, y+5.4+float32(b%2), z+dz*1.65), .23, .14, 8, c, true)
							m.Sphere(v3(x+dx*1.65, y+5.4+float32(b%2), z+dz*1.65), .17, c)
						}
						cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: .8, Top: y + 5.5})
					case k == 3:
						m.Ellipsoid(v3(x, y+.65, z), v3(1.6, .65, 1.1), 8, 14, cream)
						for b := -2; b <= 2; b++ {
							dx := float32(b) * .48
							m.Cylinder(v3(x, y+.72, z-1), v3(x+dx, y+1.1, z+.65), .07, .07, 6, pink, true)
						}
						m.Sphere(v3(x, y+1, z), .35, rl.White)
						cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 1.2, Top: y + 1.4})
					case k == 4:
						for b := 0; b < 5; b++ {
							a := float32(b) * 6.2831853 / 5
							m.Ellipsoid(v3(x+cosf(a)*.65, y+.18, z+sinf(a)*.65), v3(.7, .18, .35), 6, 8, orange)
						}
						m.Sphere(v3(x, y+.2, z), .38, orange)
					default:
						m.Cylinder(v3(x, y, z), v3(x, y+3.8, z), .11, .09, 7, aqua, true)
						m.Emis = .3
						m.Ellipsoid(v3(x, y+4, z), v3(.9, .55, .9), 8, 12, c)
						m.Emis = 0
						for b := 0; b < 6; b++ {
							a := float32(b) * 6.2831853 / 6
							dx, dz := cosf(a)*.6, sinf(a)*.6
							m.Cylinder(v3(x+dx, y+3.9, z+dz), v3(x+dx*1.2, y+2.7, z+dz*1.2), .045, .035, 5, cream, true)
						}
						cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: .2, Top: y + 4.5})
					}
				}
			}
			cb.addChunk(m, true)
		}
	}
}
