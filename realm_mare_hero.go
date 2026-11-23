package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() {
	registerRealmChar(RealmCharCode(0, RoleHero), mareHero)
	registerRealmChar(RealmCharCode(0, RoleBoss), mareBoss)
	registerRealmVehicle("mare", mareBuggy)
	Realms[0].Vehicle = "mare"
}

func mareColor(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
func mareCharacter(s CharStyle, boss bool) *CharMeshes {
	skin := mareColor(211, 145, 99)
	dark := mareColor(45, 42, 43)
	white := mareColor(255, 250, 230)
	turquoise := mareColor(23, 185, 181)
	orange := mareColor(245, 125, 37)
	blue := mareColor(35, 82, 143)
	gold := mareColor(247, 196, 62)
	green := mareColor(42, 133, 76)
	m := NewMB()
	cm := &CharMeshes{}
	shirt := turquoise
	if boss {
		shirt = blue
	}
	m.Ellipsoid(v3(0, 1.48, 0), v3(.53, .57, .32), 10, 16, shirt)
	if boss {
		m.Ellipsoid(v3(0, 1.36, .16), v3(.59, .43, .37), 10, 16, blue)
		m.Ellipsoid(v3(0, 1.00, -.04), v3(.53, .32, .30), 8, 14, blue)
		for _, y := range []float32{1.2, 1.42, 1.64, 1.82} {
			m.Sphere(v3(.09, y, .49), .045, gold)
		}
		m.Box(v3(0, 1.02, .05), v3(1, .12, .58), dark)
	} else {
		// White crest of a curling wave on the vest.
		for i := 0; i < 8; i++ {
			x := float32(i)*.055 - .2
			y := float32(1.45) + .09*sinf(float32(i)*.55)
			m.Sphere(v3(x, y, .315), .045, white)
			if i > 3 {
				m.Sphere(v3(x, y-.065, .32), .032, white)
			}
		}
	}
	m.Cylinder(v3(0, 1.87, 0), v3(0, 2.08, 0), .16, .16, 10, skin, true)
	m.Ellipsoid(v3(0, 2.32, 0), v3(.38, .43, .34), 12, 18, skin)
	for _, x := range []float32{-.15, .15} {
		m.Ellipsoid(v3(x, 2.39, .307), v3(.07, .09, .035), 6, 10, white)
		m.Sphere(v3(x, 2.39, .344), .038, dark)
	}
	for i := 0; i < 7; i++ {
		x := float32(i-3) * .045
		y := float32(2.19) + .04*(x*x/.018)
		m.Sphere(v3(x, y, .331), .023, dark)
	}
	if boss {
		m.Ellipsoid(v3(0, 2.08, .19), v3(.34, .24, .22), 8, 14, green)
		for i := -2; i <= 2; i++ {
			x := float32(i) * .12
			m.Cylinder(v3(x, 2.17, .34), v3(x+.04, 1.88, .34), .075, .025, 8, green, true)
		}
		m.Ellipsoid(v3(-.15, 2.39, .35), v3(.11, .10, .035), 6, 10, dark)
		m.Cylinder(v3(-.34, 2.49, .19), v3(.34, 2.49, .19), .025, .025, 8, dark, true)
		m.Ellipsoid(v3(0, 2.72, 0), v3(.55, .09, .42), 6, 14, dark)
		m.Ellipsoid(v3(0, 2.83, 0), v3(.43, .22, .28), 8, 14, blue)
		m.Tri(v3(-.52, 2.74, .08), v3(-.36, 3.02, .05), v3(-.15, 2.76, .20), v3(0, 0, 1), dark)
		m.Tri(v3(.52, 2.74, .08), v3(.36, 3.02, .05), v3(.15, 2.76, .20), v3(0, 0, 1), dark)
		m.Ellipsoid(v3(.36, 3.06, -.04), v3(.10, .27, .045), 8, 12, white)
		if s.Crown {
			AddCrown(m, 3.12)
		}
		cm.HeadTop = 3.12
	} else {
		hair := mareColor(247, 204, 73)
		m.Ellipsoid(v3(0, 2.63, -.04), v3(.39, .20, .34), 8, 14, hair)
		for i := -2; i <= 2; i++ {
			x := float32(i) * .13
			m.Cylinder(v3(x, 2.63, .16), v3(x+.06, 2.84, .13), .10, .005, 7, hair, true)
		}
		for _, x := range []float32{-.16, .16} {
			m.Ellipsoid(v3(x, 2.71, .25), v3(.13, .07, .04), 6, 10, dark)
		}
		m.Cylinder(v3(-.04, 2.71, .27), v3(.04, 2.71, .27), .025, .025, 6, dark, true)
		if s.Crown {
			AddCrown(m, 2.85)
		}
		cm.HeadTop = 2.85
	}
	cm.Body = m.Build()
	leg := NewMB()
	leg.Ellipsoid(v3(0, -.2, 0), v3(.22, .27, .23), 8, 12, orange)
	if boss {
		leg = NewMB()
		leg.Ellipsoid(v3(0, -.2, 0), v3(.23, .27, .24), 8, 12, blue)
	}
	cm.Leg = leg.Build()
	shin := NewMB()
	shin.Ellipsoid(v3(0, -.2, 0), v3(.15, .25, .16), 8, 12, skin)
	shoe := mareColor(38, 128, 150)
	if boss {
		shoe = dark
	}
	shin.Ellipsoid(v3(0, -.43, .08), v3(.21, .08, .30), 6, 12, shoe)
	if !boss {
		shin.Cylinder(v3(-.13, -.36, .05), v3(0, -.34, .19), .025, .025, 6, white, true)
		shin.Cylinder(v3(.13, -.36, .05), v3(0, -.34, .19), .025, .025, 6, white, true)
	}
	cm.Shin = shin.Build()
	arm := NewMB()
	arm.Ellipsoid(v3(0, -.17, 0), v3(.18, .25, .18), 8, 12, skin)
	if boss {
		arm = NewMB()
		arm.Ellipsoid(v3(0, -.17, 0), v3(.21, .25, .2), 8, 12, blue)
	}
	cm.ArmL = arm.Build()
	cm.ArmR = cm.ArmL
	fore := NewMB()
	fore.Ellipsoid(v3(0, -.18, 0), v3(.14, .24, .15), 8, 12, skin)
	fore.Sphere(v3(0, -.37, 0), .16, skin)
	cm.ForeL = fore.Build()
	cm.ForeR = cm.ForeL
	return cm
}
func mareHero(s CharStyle) *CharMeshes { return mareCharacter(s, false) }
func mareBoss(s CharStyle) *CharMeshes { return mareCharacter(s, true) }
func mareBuggy(col rl.Color) rl.Mesh {
	m := NewMB()
	dark := mareColor(48, 51, 54)
	metal := mareColor(176, 184, 182)
	seat := mareColor(97, 69, 49)
	white := mareColor(255, 246, 211)
	aqua := mareColor(28, 176, 183)
	m.BeveledBox(v3(0, .67, 0), v3(1.8, .38, 3.7), .16, col)
	m.BeveledBox(v3(0, 1.02, 1.3), v3(1.7, .45, 1), .12, col)
	for _, x := range []float32{-.98, .98} {
		for _, z := range []float32{-1.35, 1.35} {
			m.Cylinder(v3(x-.19, .43, z), v3(x+.19, .43, z), .53, .53, 16, dark, true)
			m.Cylinder(v3(x-.21, .43, z), v3(x+.21, .43, z), .23, .23, 12, metal, true)
		}
		m.Cylinder(v3(x, .85, -1.4), v3(x, 2.05, -1.1), .055, .055, 8, metal, true)
		m.Cylinder(v3(x, .85, .7), v3(x, 2.05, .55), .055, .055, 8, metal, true)
	}
	for _, z := range []float32{-1.1, .55} {
		m.Cylinder(v3(-1, 2.05, z), v3(1, 2.05, z), .055, .055, 8, metal, true)
	}
	for _, x := range []float32{-.45, .45} {
		m.Ellipsoid(v3(x, 1.12, -.35), v3(.36, .43, .22), 8, 12, seat)
	}
	m.Cylinder(v3(-.95, .5, 2), v3(.95, .5, 2), .1, .1, 10, metal, true)
	m.Ellipsoid(v3(0, 2.18, -.25), v3(.42, .07, 1.65), 8, 16, white)
	m.Box(v3(0, 2.25, -.25), v3(.12, .025, 2.3), aqua)
	for _, z := range []float32{-.9, .4} {
		m.Box(v3(0, 2.26, z), v3(.82, .035, .06), dark)
	}
	return m.Build()
}
