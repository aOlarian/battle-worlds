package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() {
	for _, role := range []int{RoleEnemy1, RoleEnemy2, RoleEnemy3, RoleEnemy4, RoleNPC1, RoleNPC2, RoleNPC3} {
		r := role
		registerRealmChar(RealmCharCode(4, r), func(s CharStyle) *CharMeshes { return christmasCast(r, s) })
	}
}
func christmasCast(role int, s CharStyle) *CharMeshes {
	c := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	white := c(245, 248, 246)
	dark := c(36, 40, 48)
	red := c(190, 34, 44)
	skin := c(239, 182, 147)
	gold := c(247, 192, 53)
	green := c(88, 162, 74)
	purple := c(132, 73, 171)
	blue := c(43, 155, 192)
	brown := c(112, 70, 45)
	bodyCol, headCol, limb, hand, shoe := red, skin, red, skin, dark
	switch role {
	case RoleEnemy1:
		bodyCol, headCol, limb, hand, shoe = green, green, green, green, c(56, 106, 57)
	case RoleEnemy2:
		bodyCol, headCol, limb, hand, shoe = white, white, brown, brown, white
	case RoleEnemy3:
		bodyCol, headCol, limb, hand, shoe = dark, dark, dark, dark, c(240, 162, 43)
	case RoleEnemy4:
		bodyCol, headCol, limb, hand, shoe = purple, skin, purple, skin, dark
	case RoleNPC3:
		bodyCol, headCol, limb, hand, shoe = c(47, 144, 97), skin, c(47, 144, 97), skin, brown
	}
	m := NewMB()
	ell := func(x, y, z, rx, ry, rz float32, col rl.Color) { m.Ellipsoid(v3(x, y, z), v3(rx, ry, rz), 8, 12, col) }
	ell(0, 1.42, 0, .55, .62, .35, bodyCol)
	if role == RoleEnemy2 {
		ell(0, 1.04, 0, .62, .52, .44, white)
		ell(0, 1.64, 0, .46, .43, .36, white)
	}
	if role == RoleEnemy3 {
		ell(0, 1.42, .25, .39, .52, .17, white)
	}
	ell(0, 2.24, 0, .42, .43, .36, headCol)
	if role == RoleEnemy3 {
		ell(0, 2.22, .25, .33, .29, .15, white)
	}
	if role == RoleEnemy4 {
		ell(0, 2.31, .31, .40, .14, .08, dark)
	}
	eyeWhite := white
	if role == RoleEnemy1 {
		eyeWhite = c(250, 222, 70)
	}
	for _, x := range []float32{-.16, .16} {
		ell(x, 2.30, .33, .105, .12, .065, eyeWhite)
		ell(x, 2.30, .39, .043, .065, .035, dark)
		ell(x-.015, 2.33, .42, .015, .02, .012, white)
	}
	if role == RoleEnemy2 || role == RoleEnemy3 {
		m.Cylinder(v3(0, 2.19, .34), v3(0, 2.16, .68), .105, .01, 10, c(244, 139, 40), true)
	} else {
		nose := skin
		if role == RoleEnemy1 {
			nose = green
		}
		ell(0, 2.15, .35, .09, .075, .08, nose)
		for i := 0; i < 7; i++ {
			x := float32(i-3) * .055
			y := float32(2.07) + .025*float32((i-3)*(i-3))/9
			ell(x, y, .365, .035, .022, .025, dark)
		}
	}
	if role == RoleNPC1 {
		ell(0, 1.98, .27, .39, .33, .23, white)
		ell(-.14, 2.11, .43, .18, .08, .08, white)
		ell(.14, 2.11, .43, .18, .08, .08, white)
	}
	if role == RoleNPC2 {
		ell(0, 1.05, 0, .58, .35, .38, red)
		ell(0, 2.43, -.1, .43, .27, .32, c(220, 220, 213))
		for _, x := range []float32{-.16, .16} {
			m.Cylinder(v3(x, 2.30, .405), v3(x, 2.30, .42), .13, .13, 12, gold, true)
			ell(x, 2.30, .43, .10, .10, .02, skin)
			ell(x, 2.30, .455, .035, .05, .02, dark)
		}
		m.BeveledBox(v3(0, 1.37, .32), v3(.72, .65, .10), .08, white)
	}
	if role == RoleNPC3 {
		m.BeveledBox(v3(0, 1.35, .32), v3(.69, .69, .10), .07, c(197, 145, 91))
		m.BeveledBox(v3(0, 1.25, .39), v3(.35, .18, .06), .03, brown)
	}
	if role == RoleNPC1 {
		m.Cylinder(v3(0, 1.25, 0), v3(0, 1.39, 0), .55, .55, 14, dark, true)
		m.BeveledBox(v3(0, 1.32, .53), v3(.19, .17, .06), .025, gold)
	}
	if role == RoleEnemy2 || role == RoleEnemy3 {
		scarf := red
		if role == RoleEnemy3 {
			scarf = blue
		}
		m.Cylinder(v3(0, 1.92, 0), v3(0, 2.03, 0), .43, .39, 12, scarf, true)
		ell(.22, 1.72, .34, .10, .28, .07, scarf)
	}
	if role == RoleEnemy4 {
		ell(.22, 1.57, -.37, .42, .50, .32, c(184, 132, 83))
		m.Cylinder(v3(.2, 1.93, -.4), v3(.2, 2.09, -.4), .14, .09, 8, brown, true)
	}
	if role == RoleEnemy1 {
		for _, x := range []float32{-.25, 0, .25} {
			m.Cylinder(v3(x, 2.52, 0), v3(x*1.2, 2.77, -.04), .10, .01, 7, green, true)
		}
	}
	if role == RoleNPC1 || role == RoleNPC3 || role == RoleEnemy4 {
		hat := red
		if role == RoleNPC3 {
			hat = c(47, 144, 97)
		}
		if role == RoleEnemy4 {
			hat = purple
		}
		m.Cylinder(v3(0, 2.52, 0), v3(.16, 2.90, -.04), .39, .025, 12, hat, true)
		ell(.16, 2.9, -.04, .10, .10, .10, white)
		m.Cylinder(v3(0, 2.50, 0), v3(0, 2.58, 0), .40, .38, 12, white, true)
	}
	if role == RoleEnemy2 {
		m.Cylinder(v3(0, 2.56, 0), v3(0, 2.78, 0), .32, .27, 12, blue, true)
		m.Cylinder(v3(0, 2.51, 0), v3(0, 2.58, 0), .39, .39, 12, dark, true)
	}
	if role == RoleEnemy3 {
		ell(0, 2.52, -.02, .42, .23, .37, blue)
		m.Cylinder(v3(0, 2.48, 0), v3(0, 2.54, 0), .42, .42, 12, c(100, 214, 226), true)
	}
	if s.Crown {
		AddCrown(m, 2.92)
	}
	cm := &CharMeshes{Body: m.Build(), HeadTop: 2.92}
	legCol := limb
	if role == RoleEnemy2 {
		legCol = white
	}
	leg := NewMB()
	leg.Ellipsoid(v3(0, -.22, 0), v3(.20, .27, .22), 6, 10, legCol)
	cm.Leg = leg.Build()
	leg = NewMB()
	leg.Ellipsoid(v3(0, -.22, 0), v3(.19, .25, .21), 6, 10, legCol)
	leg.Ellipsoid(v3(0, -.42, .10), v3(.22, .12, .30), 6, 10, shoe)
	cm.Shin = leg.Build()
	for _, side := range []int{-1, 1} {
		a := NewMB()
		a.Ellipsoid(v3(0, -.18, 0), v3(.16, .23, .17), 6, 10, limb)
		upper := a.Build()
		a = NewMB()
		a.Ellipsoid(v3(0, -.18, 0), v3(.14, .22, .15), 6, 10, limb)
		a.Ellipsoid(v3(0, -.38, .04), v3(.15, .13, .14), 6, 10, hand)
		if role == RoleEnemy2 {
			a.Cylinder(v3(0, -.30, 0), v3(float32(side)*.25, -.46, 0), .04, .015, 6, brown, true)
		}
		if role == RoleNPC3 && side == 1 {
			a.Cylinder(v3(0, -.38, .04), v3(0, -.63, .04), .045, .045, 8, brown, true)
			a.BeveledBox(v3(0, -.64, .04), v3(.32, .14, .14), .03, c(160, 166, 170))
		}
		fore := a.Build()
		if side < 0 {
			cm.ArmL, cm.ForeL = upper, fore
		} else {
			cm.ArmR, cm.ForeR = upper, fore
		}
	}
	return cm
}
