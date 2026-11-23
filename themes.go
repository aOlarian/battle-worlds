package main

import rl "github.com/gen2brain/raylib-go/raylib"

var boothMesh rl.Mesh

func buildBoothMesh() {
	m := NewMB()
	red := rl.NewColor(220, 35, 40, 255)
	trim := rl.NewColor(165, 22, 32, 255)
	glass := rl.NewColor(154, 225, 245, 255)
	dark := rl.NewColor(38, 52, 66, 255)
	cream := rl.NewColor(255, 244, 185, 255)
	// Platform, back, and four corner posts leave the front open.
	m.BeveledBox(v3(0, .12, 0), v3(1.6, .24, 1.6), .08, trim)
	m.Box(v3(0, 1.35, -.72), v3(1.42, 2.25, .12), red)
	for _, x := range []float32{-.72, .72} {
		for _, z := range []float32{-.72, .72} {
			m.Box(v3(x, 1.36, z), v3(.16, 2.35, .16), red)
		}
		m.Box(v3(x, .48, 0), v3(.12, .52, 1.35), red)
		m.Emis = .3
		m.Box(v3(x, 1.48, 0), v3(.08, 1.43, 1.26), glass)
		m.Emis = 0
		m.Box(v3(x, 1.45, 0), v3(.14, .07, 1.35), red)
		m.Box(v3(x, 1.48, 0), v3(.14, 1.5, .07), red)
	}
	// Door frame and glass panes, with a clear central opening.
	for _, x := range []float32{-.53, .53} {
		m.Emis = .25
		m.Box(v3(x, 1.47, .74), v3(.26, 1.46, .055), glass)
		m.Emis = 0
		m.Box(v3(x, 1.47, .78), v3(.28, .07, .08), red)
	}
	m.Box(v3(0, 2.34, 0), v3(1.6, .20, 1.6), red)
	m.BeveledBox(v3(0, 2.55, 0), v3(1.68, .27, 1.68), .08, trim)
	m.Ellipsoid(v3(0, 2.67, 0), v3(.84, .23, .84), 8, 16, red)
	// Large toy telephone on the back wall, with handset and keypad.
	m.BeveledBox(v3(0, 1.35, -.61), v3(.83, 1.05, .18), .08, dark)
	m.Box(v3(0, 1.56, -.50), v3(.52, .18, .08), rl.NewColor(110, 210, 235, 255))
	for _, x := range []float32{-.18, 0, .18} {
		for _, y := range []float32{1.05, 1.19, 1.33} {
			m.Sphere(v3(x, y, -.48), .047, cream)
		}
	}
	m.Cylinder(v3(-.31, 1.65, -.45), v3(-.31, 1.21, -.45), .065, .065, 8, cream, true)
	m.Sphere(v3(-.31, 1.65, -.45), .095, cream)
	m.Sphere(v3(-.31, 1.21, -.45), .095, cream)
	// Illuminated sign and unmistakable curved receiver silhouette.
	m.Emis = .85
	m.BeveledBox(v3(0, 2.55, .86), v3(1.15, .27, .08), .04, cream)
	m.Emis = 0
	m.Cylinder(v3(-.25, 2.58, .915), v3(-.16, 2.51, .915), .038, .038, 8, dark, true)
	m.Cylinder(v3(-.16, 2.51, .915), v3(.16, 2.51, .915), .038, .038, 8, dark, true)
	m.Cylinder(v3(.16, 2.51, .915), v3(.25, 2.58, .915), .038, .038, 8, dark, true)
	m.Sphere(v3(-.25, 2.58, .915), .056, dark)
	m.Sphere(v3(.25, 2.58, .915), .056, dark)
	boothMesh = m.Build()
}

func (cb *CityBuilder) themeClear(x, z, r float32) bool {
	// Nex: drum liber spre trambuline (Aventurierul merge spre ele si se urca singur)
	for _, p := range cb.W.Pads {
		if sq(x-p.Pos.X)+sq(z-p.Pos.Z) < sq(p.R+r+6) {
			return false
		}
	}
	if cb.nearRoad(x, z, r+1) || cb.reserved(x, z, r+1) || cb.inMountain(x, z, r+1) || cb.W.stairMouth(x, z, r+1) || cb.W.nearOpenCut(x, z, r+1) || cb.W.LandSD(x, z) > -10 || cb.W.LandHeight(x, z) < WaterY+1 {
		return false
	}
	for _, c := range cb.W.Circles {
		dx, dz := x-c.X, z-c.Z
		if dx*dx+dz*dz < (r+c.R+1)*(r+c.R+1) {
			return false
		}
	}
	return true
}

func (cb *CityBuilder) buildThemeDecor() {
	cb.buildSpiderDecor()
	cb.buildDinoDecor()
	cb.buildHeroDecor()
	cb.buildBatDecor()
	cb.buildSeasonDecor()
	colors := []rl.Color{LegoRed, rl.NewColor(38, 110, 225, 255), rl.NewColor(255, 210, 45, 255), rl.NewColor(75, 190, 95, 255)}
	for si, s := range Sectors {
		if si > 6 {
			break
		}
		if si == 1 || si == 0 || si == 3 || si == 2 || si == 5 || si == 6 {
			continue
		}
		for gx := -1; gx <= 1; gx++ {
			for gz := -1; gz <= 1; gz++ {
				m := NewMB()
				for ix := 0; ix < 4; ix++ {
					for iz := 0; iz < 4; iz++ {
						x := s.Center.X + float32(gx*48+ix*11-16)
						z := s.Center.Y + float32(gz*48+iz*11-16)
						if !cb.themeClear(x, z, 2.1) {
							continue
						}
						if si == 4 && x*x+z*z < 55*55 {
							continue // Nex: piata ramane libera pentru lupta (doar pe iarba din jurul ei)
						}
						y := cb.W.LandHeight(x, z)
						k := (ix + iz + gx + gz + 20) % 4
						c := colors[k]
						switch si {
						case 4: // piese, turnuri si flori din caramizi
							if (ix+iz+gx+gz+20)%4 == 0 {
								// Turn cu piese intercalate si butoni vizibili.
								for level := 0; level < 4; level++ {
									cc := colors[(k+level)%4]
									cy := y + float32(level)*1.15
									m.Box(v3(x, cy+.55, z), v3(2.2, 1.1, 2.2), cc)
									for _, dx := range []float32{-.55, .55} {
										for _, dz := range []float32{-.55, .55} {
											m.Cylinder(v3(x+dx, cy+1.1, z+dz), v3(x+dx, cy+1.25, z+dz), .25, .25, 8, cc, true)
										}
									}
								}
							} else if (ix+iz+gx+gz+20)%4 == 1 {
								// Floare de jucarie: tulpina patrata si petale cubice.
								m.Box(v3(x, y+.75, z), v3(.24, 1.5, .24), colors[3])
								m.Box(v3(x+.35, y+.5, z), v3(.6, .16, .3), colors[3])
								for _, o := range [][2]float32{{-.48, 0}, {.48, 0}, {0, -.48}, {0, .48}} {
									m.Box(v3(x+o[0], y+1.65, z+o[1]), v3(.48, .3, .48), c)
								}
								m.Cylinder(v3(x, y+1.5, z), v3(x, y+1.9, z), .26, .26, 8, colors[2], true)
							} else {
								m.Box(v3(x, y+.65, z), v3(2.4, 1.3, 1.5), c)
								for _, dx := range []float32{-.6, .6} {
									m.Cylinder(v3(x+dx, y+1.3, z), v3(x+dx, y+1.55, z), .3, .3, 8, c, true)
								}
							}
						case 0: // oua si ferigi
							m.Ellipsoid(v3(x, y+.9, z), v3(.65, .9, .65), 7, 10, rl.NewColor(239, 225, 169, 255))
							m.Sphere(v3(x+.3, y+1.1, z+.48), .16, colors[3])
							for a := 0; a < 6; a++ {
								ang := float32(a) * 1.047
								m.Cylinder(v3(x+1.3, y, z), v3(x+1.3+cosf(ang), y+.8, z+sinf(ang)), .08, .2, 5, colors[3], false)
							}
						case 1: // panza pe un cadru independent
							white := rl.NewColor(235, 245, 250, 255)
							for a := 0; a < 8; a++ {
								ang := float32(a) * .7854
								m.Cylinder(v3(x, y+2, z), v3(x+cosf(ang)*1.5, y+2+sinf(ang)*1.5, z), .035, .035, 4, white, false)
							}
							for a := 0; a < 8; a++ {
								a0 := float32(a) * .7854
								a1 := float32(a+1) * .7854
								for _, r := range []float32{.6, 1.1, 1.5} {
									m.Cylinder(v3(x+cosf(a0)*r, y+2+sinf(a0)*r, z), v3(x+cosf(a1)*r, y+2+sinf(a1)*r, z), .03, .03, 4, white, false)
								}
							}
						case 2: // stalp galben si steag mov
							m.Cylinder(v3(x, y, z), v3(x, y+4, z), .12, .1, 6, ColLampPole, true)
							m.Box(v3(x+.65, y+3.5, z), v3(1.3, .8, .1), rl.NewColor(112, 76, 155, 255))
							m.Sphere(v3(x, y+4.2, z), .3, ColLampLight)
						case 3: // cristal si stea aurie
							m.Cylinder(v3(x, y, z), v3(x, y+3, z), .7, 0, 5, rl.NewColor(90, 190, 240, 255), true)
							m.Sphere(v3(x, y+3, z), .22, colors[2])
						case 5: // umbrela de plaja
							m.Cylinder(v3(x, y, z), v3(x, y+2.5, z), .09, .09, 6, ColTrunk, true)
							m.Cylinder(v3(x, y+2.3, z), v3(x, y+2.9, z), 1.6, 0, 8, c, true)
						case 6: // brad cu globuri si cadouri
							cb.pineTree(m, x, y, z, .8)
							for a := 0; a < 6; a++ {
								ang := float32(a) * 1.047
								m.Sphere(v3(x+cosf(ang)*.8, y+1.8, z+sinf(ang)*.8), .15, colors[a%4])
							}
							m.Sphere(v3(x, y+4.1, z), .25, colors[2])
							m.Box(v3(x+1.4, y+.35, z), v3(.7, .7, .7), c)
							m.Box(v3(x+1.4, y+.72, z), v3(.15, .06, .72), colors[2])
						}
						if si != 4 {
							cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: .8, Top: y + 3})
						} else if (ix+iz+gx+gz+20)%4 != 1 {
							cb.W.Circles = append(cb.W.Circles, CircleCollider{X: x, Z: z, R: 1.6, Top: y + 5})
						}
					}
				}
				cb.addChunk(m, true)
			}
		}
	}
}

func (cw *CityWorld) DrawThemeAnimated(draw DrawFn, t float32, shadow bool, near func(rl.Vector3, float32) bool) {
	drawDinoAnimated(draw, t, shadow, near)
}
