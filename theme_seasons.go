package main

import rl "github.com/gen2brain/raylib-go/raylib"

func (cb *CityBuilder) buildSeasonDecor() {
	red := rl.NewColor(244, 70, 82, 255)
	yellow := rl.NewColor(255, 210, 55, 255)
	blue := rl.NewColor(35, 164, 230, 255)
	aqua := rl.NewColor(45, 200, 231, 255)
	white := rl.NewColor(255, 246, 220, 255)
	sand := rl.NewColor(231, 188, 112, 255)
	green := rl.NewColor(50, 180, 89, 255)
	pink := rl.NewColor(255, 110, 171, 255)
	colors := []rl.Color{red, yellow, blue, pink, green}
	for sector := 0; sector < 2; sector++ {
		center := float32(-356)
		if sector == 1 {
			center = 362
		}
		kiosk := false
		slide := false
		for gx := -1; gx <= 1; gx++ {
			for gz := -1; gz <= 1; gz++ {
				m := NewMB()
				for i := 0; i < 5; i++ {
					for j := 0; j < 5; j++ {
						x := center + float32(gx*48+i*9-18)
						z := float32(gz*48 + j*9 - 18)
						k := (i*7 + j*3 + (gx+1)*5 + (gz+1)*11) % 7
						r := float32(2.5)
						if sector == 0 && !kiosk {
							r = 5
						}
						if sector == 1 && !slide {
							r = 7
						}
						if !cb.heroClear(x, z, r) {
							continue
						}
						y := cb.W.LandHeight(x, z)
						c := colors[(i+j+gx+gz+8)%len(colors)]
						if sector == 0 {
							if !kiosk {
								kiosk = true
								m.Box(v3(x, y+2, z), v3(6, 4, 5), pink)
								m.Box(v3(x, y+4.2, z), v3(6.5, .4, 5.5), white)
								m.Box(v3(x, y+2, z+2.55), v3(3, 1.3, .12), aqua)
								m.Cylinder(v3(x, y+4.4, z), v3(x, y+6, z), .8, 0, 12, sand, true)
								m.Sphere(v3(x, y+6, z), .85, pink)
								m.Sphere(v3(x, y+6.8, z), .6, white)
								cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 4, Top: y + 7.4})
							} else {
								switch k {
								case 0, 1:
									m.Cylinder(v3(x, y, z), v3(x, y+2.8, z), .09, .09, 8, white, true)
									for a := 0; a < 8; a++ {
										a0 := float32(a) * .785398
										a1 := float32(a+1) * .785398
										col := c
										if a%2 == 0 {
											col = white
										}
										m.Tri(v3(x, y+3.25, z), v3(x+2*cosf(a0), y+2.65, z+2*sinf(a0)), v3(x+2*cosf(a1), y+2.65, z+2*sinf(a1)), v3(0, 1, 0), col)
									}
								case 2:
									m.Box(v3(x, y+.4, z), v3(1.3, .2, 2.8), c)
									m.Box(v3(x, y+.8, z-1.05), v3(1.3, .9, .2), white)
									for _, dx := range []float32{-.55, .55} {
										m.Box(v3(x+dx, y+.2, z), v3(.12, .4, 2.6), sand)
									}
								case 3:
									m.Box(v3(x, y+.55, z), v3(2, 1.1, 2), sand)
									for _, dx := range []float32{-.9, .9} {
										for _, dz := range []float32{-.9, .9} {
											m.Cylinder(v3(x+dx, y, z+dz), v3(x+dx, y+1.65, z+dz), .42, .42, 8, sand, true)
										}
									}
								case 4:
									for a := 0; a < 16; a++ {
										a0 := float32(a) * .392699
										a1 := float32(a+1) * .392699
										col := c
										if a%4 < 2 {
											col = white
										}
										m.Cylinder(v3(x+cosf(a0), y+.3, z+sinf(a0)), v3(x+cosf(a1), y+.3, z+sinf(a1)), .28, .28, 6, col, true)
									}
								case 5:
									m.Sphere(v3(x, y+.8, z), .8, c)
									m.Disc(v3(x, y+.02, z+2), 0, 1.3, 20, white)
								default:
									m.Box(v3(x, y+.04, z), v3(1.7, .08, 3), c)
									m.Box(v3(x, y+.09, z), v3(1.7, .03, .35), white)
								}
							}
						} else {
							if !slide {
								slide = true
								for _, dx := range []float32{-3, 3} {
									for _, dz := range []float32{-3, 3} {
										m.Cylinder(v3(x+dx, y, z+dz), v3(x+dx, y+18, z+dz), .32, .32, 8, yellow, true)
										cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x + dx, Z: z + dz, R: .4, Top: y + 18})
									}
								}
								m.Box(v3(x, y+18, z), v3(7, .5, 7), pink)
								for a := 0; a < 16; a++ {
									t := float32(a) / 16
									u := float32(a+1) / 16
									a0 := t * 9.42
									a1 := u * 9.42
									m.Cylinder(v3(x+4*cosf(a0), y+17-t*16, z+4*sinf(a0)), v3(x+4*cosf(a1), y+17-u*16, z+4*sinf(a1)), 1.1, 1.1, 10, colors[a%5], true)
								}
							} else {
								switch k {
								case 0, 1:
									m.Disc(v3(x, y+.05, z), 0, 3.2, 32, blue)
									m.Disc(v3(x, y+.06, z), 2.9, 3.2, 32, white)
								case 2:
									m.Cylinder(v3(x, y, z), v3(x, y+2.6, z), .3, .3, 10, white, true)
									m.Ellipsoid(v3(x, y+2.7, z), v3(2, .65, 2), 8, 16, c)
									for a := 0; a < 8; a++ {
										ang := float32(a) * .785398
										m.Cylinder(v3(x+1.65*cosf(ang), y+2.6, z+1.65*sinf(ang)), v3(x+1.9*cosf(ang), y+.3, z+1.9*sinf(ang)), .035, .035, 6, aqua, true)
									}
									// Nex: ciuperca de apa nu are coliziune (copiii alearga pe sub stropi)
								case 3:
									m.Disc(v3(x, y+.05, z), 0, 2.2, 24, blue)
									m.Cylinder(v3(x, y+.1, z), v3(x, y+3, z), .12, .12, 8, aqua, true)
									m.Sphere(v3(x, y+3, z), .35, white)
								default:
									m.Cylinder(v3(x, y, z), v3(x+.5, y+5, z), .35, .22, 8, sand, true)
									for a := 0; a < 7; a++ {
										ang := float32(a) * .8976
										m.Cylinder(v3(x+.5, y+5, z), v3(x+.5+2.4*cosf(ang), y+4.3, z+2.4*sinf(ang)), .55, .04, 7, green, true)
									}
									cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: .4, Top: y + 5})
								}
							}
						}
					}
				}
				cb.addChunk(m, true)
			}
		}
	}
}
