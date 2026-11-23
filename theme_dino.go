package main

import rl "github.com/gen2brain/raylib-go/raylib"

var dinoHead rl.Mesh
var dinoHeadPos rl.Vector3
var dinoHeadReady bool

func (cb *CityBuilder) buildDinoDecor() {
	dinoHeadReady = false
	green := rl.NewColor(33, 143, 74, 255)
	lime := rl.NewColor(85, 192, 78, 255)
	brown := rl.NewColor(135, 86, 44, 255)
	rock := rl.NewColor(126, 131, 105, 255)
	blue := rl.NewColor(35, 180, 227, 255)
	foam := rl.NewColor(202, 248, 250, 255)
	pink := rl.NewColor(255, 75, 143, 255)
	yellow := rl.NewColor(255, 209, 53, 255)
	colors := []rl.Color{pink, yellow, rl.NewColor(241, 92, 54, 255), rl.NewColor(168, 89, 226, 255)}
	// Reperul este ales din mai multe locuri, cu tot bazinul ferit de drumuri.
	for _, p := range [][2]float32{{-187, -190}, {-190, -145}, {-140, -193}, {-135, -135}, {-185, -175}, {-145, -180}, {-180, -180}, {-175, -185}, {-195, -170}, {-145, -145}, {-175, -120}, {-125, -175}, {-200, -150}, {-150, -200}} {
		x, z := p[0], p[1]
		if !cb.heroClear(x, z, 5) {
			continue
		}
		y := cb.W.LandHeight(x, z)
		m := NewMB()
		m.Ellipsoid(v3(x, y+5, z-2), v3(6, 8, 4), 8, 12, rock)
		m.Box(v3(x, y+6, z+1.8), v3(3.2, 11, .35), blue)
		m.Ellipsoid(v3(x, y+.16, z+4), v3(5, .16, 4), 6, 12, blue)
		for i := 0; i < 7; i++ {
			a := float32(i) * .9
			m.Sphere(v3(x+cosf(a)*3.5, y+.3, z+4+sinf(a)*2.8), .3, foam)
		}
		for i := 0; i < 4; i++ {
			m.Box(v3(x-1.2+float32(i)*.8, y+5, z+2.01), v3(.12, 8, .05), foam)
		}
		cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z - 2, R: 4.5, Top: y + 13})
		cb.addChunk(m, true)
		break
	}
	for gx := -1; gx <= 1; gx++ {
		for gz := -1; gz <= 1; gz++ {
			m := NewMB()
			for i := 0; i < 6; i++ {
				for j := 0; j < 6; j++ {
					x := float32(-160 + gx*36 + i*6 - 15)
					z := float32(-160 + gz*36 + j*6 - 15)
					if sq(x+160)+sq(z+160) > sq(59) {
						continue
					}
					k := (i*7 + j*11 + (gx+1)*13 + (gz+1)*17) % 12
					r := float32(1.3)
					if k == 0 {
						r = 4
					}
					if k == 1 {
						r = 3
					}
					if !cb.heroClear(x, z, r) {
						continue
					}
					y := cb.W.LandHeight(x, z)
					switch {
					case k == 0: // coliba pe piloni, cu acoperis de frunze
						for _, dx := range []float32{-2, 2} {
							for _, dz := range []float32{-2, 2} {
								m.Cylinder(v3(x+dx, y, z+dz), v3(x+dx, y+3, z+dz), .18, .18, 6, brown, true)
							}
						}
						m.Box(v3(x, y+2.7, z), v3(5, .3, 5), brown)
						m.Box(v3(x, y+4, z-.9), v3(4, 2.5, .18), brown)
						m.Box(v3(x-1.9, y+4, z), v3(.2, 2.5, 2), brown)
						m.Box(v3(x+1.9, y+4, z), v3(.2, 2.5, 2), brown)
						m.Cylinder(v3(x, y+5, z), v3(x, y+6.8, z), 3.9, 0, 8, green, true)
						cb.W.Boxes = append(cb.W.Boxes, BoxCollider{MinX: x - 2.5, MaxX: x + 2.5, MinZ: z - 2.5, MaxZ: z + 2.5, Top: y + 5})
					case k <= 4: // palmier, cu papagal pe o creanga
						h := float32(6 + (i+j)%3)
						m.Cylinder(v3(x, y, z), v3(x+.5, y+h, z), .4, .25, 7, brown, true)
						for a := 0; a < 8; a++ {
							ang := float32(a) * .785398
							dx, dz := cosf(ang), sinf(ang)
							m.Ellipsoid(v3(x+.5+dx*1.35, y+h-.4, z+dz*1.35), v3(.75, .22, .75), 5, 8, green)
							m.Cylinder(v3(x+.5, y+h, z), v3(x+.5+dx*2.8, y+h-.9, z+dz*2.8), .23, .025, 5, lime, true)
						}
						if k == 2 {
							bx := x + 1.1
							by := y + h - 1.1
							m.Cylinder(v3(x, y+h-1.8, z), v3(x+2, y+h-1, z), .13, .08, 6, brown, true)
							m.Ellipsoid(v3(bx, by, z), v3(.42, .65, .35), 6, 8, colors[(i+j)%4])
							m.Sphere(v3(bx, by+.55, z), .32, colors[(i+j)%4])
							m.Cylinder(v3(bx+.2, by+.5, z), v3(bx+.6, by+.35, z), .18, 0, 6, yellow, true)
							m.Sphere(v3(bx+.18, by+.65, z+.25), .06, rl.White)
							m.Cylinder(v3(bx-.2, by-.3, z), v3(bx-.8, by-1.1, z), .18, .01, 5, blue, true)
						}
						cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: .5, Top: y + h})
					case k <= 6: // bambus
						for a := 0; a < 5; a++ {
							ang := float32(a) * 1.2566
							xx, zz := x+cosf(ang)*.65, z+sinf(ang)*.65
							h := float32(4 + a%3)
							m.Cylinder(v3(xx, y, zz), v3(xx, y+h, zz), .12, .09, 6, lime, true)
							for n := 1; n < 4; n++ {
								yy := y + float32(n)*h/4
								m.Cylinder(v3(xx-.14, yy, zz), v3(xx+.14, yy, zz), .045, .045, 5, green, true)
							}
						}
						cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: .8, Top: y + 5})
					case k <= 9: // ferigi uriase
						for a := 0; a < 9; a++ {
							ang := float32(a) * .698
							dx, dz := cosf(ang), sinf(ang)
							m.Cylinder(v3(x, y+.1, z), v3(x+dx*2, y+1.7, z+dz*2), .12, .02, 5, green, true)
							for n := 1; n <= 3; n++ {
								q := float32(n) / 4
								for _, s := range []float32{-1, 1} {
									m.Cylinder(v3(x+dx*2*q, y+1.7*q, z+dz*2*q), v3(x+dx*2*q-dz*s*.6, y+1.7*q+.15, z+dz*2*q+dx*s*.6), .18, .01, 5, lime, true)
								}
							}
						}
					default: // flori mari
						m.Cylinder(v3(x, y, z), v3(x, y+1.5, z), .09, .07, 5, green, true)
						for a := 0; a < 6; a++ {
							ang := float32(a) * 1.047
							m.Sphere(v3(x+cosf(ang)*.5, y+1.5, z+sinf(ang)*.5), .32, colors[(i+j)%4])
						}
						m.Sphere(v3(x, y+1.6, z), .25, yellow)
					}
				}
			}
			cb.addChunk(m, true)
		}
	}
	// Poduri suspendate decorative, amplasate numai unde toata amprenta e libera.
	for _, p := range [][2]float32{{-180, -135}, {-140, -180}, {-185, -170}, {-135, -150}} {
		x, z := p[0], p[1]
		if !cb.heroClear(x, z, 4) {
			continue
		}
		y := cb.W.LandHeight(x, z)
		m := NewMB()
		for n := -5; n <= 5; n++ {
			xx := x + float32(n)*.8
			sag := float32(n*n) * .025
			m.Box(v3(xx, y+2+sag, z), v3(.7, .16, 2), brown)
			for _, s := range []float32{-1, 1} {
				m.Cylinder(v3(xx, y+2+sag, z+s), v3(xx, y+3.2+sag, z+s), .035, .035, 5, green, true)
				if n < 5 {
					m.Cylinder(v3(xx, y+3.2+sag, z+s), v3(xx+.8, y+3.2+float32((n+1)*(n+1))*.025, z+s), .07, .07, 5, green, true)
				}
			}
		}
		cb.addChunk(m, true)
	}
}
func drawDinoAnimated(draw DrawFn, t float32, shadow bool, near func(rl.Vector3, float32) bool) {}
