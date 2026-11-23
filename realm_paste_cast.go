package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() {
	for role := RoleEnemy1; role <= RoleEnemy4; role++ {
		r := role
		registerRealmChar(RealmCharCode(5, r), func(s CharStyle) *CharMeshes { return pasteCast(s, r) })
	}
	for role := RoleNPC1; role <= RoleNPC3; role++ {
		r := role
		registerRealmChar(RealmCharCode(5, r), func(s CharStyle) *CharMeshes { return pasteCast(s, r) })
	}
}

func pasteCast(s CharStyle, role int) *CharMeshes {
	c := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	orange, cream, dark := c(231, 119, 42), c(255, 236, 199), c(39, 40, 51)
	white, pink, yellow := c(250, 247, 235), c(242, 134, 165), c(255, 211, 55)
	base := orange
	switch role {
	case RoleEnemy2:
		base = c(40, 57, 83)
	case RoleEnemy3:
		base = c(76, 76, 78)
	case RoleEnemy4:
		base = c(164, 119, 84)
	case RoleNPC1:
		base = yellow
	case RoleNPC2:
		base = white
	case RoleNPC3:
		base = c(224, 169, 105)
	}
	b := NewMB()
	ell := func(x, y, z, rx, ry, rz float32, col rl.Color) { b.Ellipsoid(v3(x, y, z), v3(rx, ry, rz), 8, 12, col) }
	ell(0, 1.38, 0, .52, .64, .38, base)
	ell(0, 2.12, 0, .51, .49, .39, base)
	if role == RoleEnemy1 || role == RoleEnemy3 || role == RoleEnemy4 || role == RoleNPC2 {
		ell(0, 1.38, .34, .34, .42, .07, cream)
	}
	if role == RoleEnemy3 {
		ell(0, 2.15, .35, .19, .40, .07, white)
		ell(0, 2.15, -.35, .17, .39, .08, white)
		ell(0, 1.45, -.35, .15, .47, .07, white)
	}
	if role == RoleNPC2 {
		for _, x := range []float32{-.4, -.2, 0, .2, .4} {
			ell(x, 2.48, 0, .18, .19, .29, white)
		}
		ell(0, 2.08, .36, .32, .27, .12, cream)
	}
	if role == RoleEnemy1 || role == RoleEnemy3 || role == RoleNPC2 {
		for _, x := range []float32{-.38, .38} {
			b.Cylinder(v3(x, 2.35, 0), v3(x*1.3, 2.85, 0), .19, .015, 8, base, true)
			b.Cylinder(v3(x, 2.40, .12), v3(x*1.25, 2.72, .08), .09, .01, 8, pink, true)
		}
	}
	if role == RoleEnemy4 {
		for row := 0; row < 4; row++ {
			for i := -3; i <= 3; i++ {
				x := float32(i) * .15
				y := float32(1.66 + float32(row)*.22)
				z := float32(-.28)
				b.Cylinder(v3(x, y, z), v3(x*1.35, y+.27, z-.35), .105, .005, 6, c(100, 68, 51), true)
			}
		}
	}
	if role == RoleEnemy2 || role == RoleNPC1 || role == RoleNPC3 {
		b.Cylinder(v3(0, 2.08, .30), v3(0, 2.02, .67), .20, .015, 8, c(239, 151, 40), true)
	} else {
		ell(0, 1.99, .40, .29, .18, .18, cream)
		ell(0, 2.09, .57, .09, .07, .06, dark)
	}
	for _, x := range []float32{-.21, .21} {
		ell(x, 2.23, .33, .14, .16, .08, white)
		ell(x, 2.23, .405, .065, .085, .04, dark)
		ell(x-.018, 2.27, .44, .022, .027, .01, white)
	}
	if role == RoleEnemy3 {
		for _, x := range []float32{-.21, .21} {
			b.Cylinder(v3(x-.12, 2.40, .39), v3(x+.12, 2.36, .39), .035, .035, 6, dark, true)
		}
	}
	if role == RoleEnemy1 {
		b.Cylinder(v3(0, 1.76, 0), v3(0, 1.88, 0), .43, .43, 12, c(76, 171, 171), true)
		ell(.32, 1.65, .32, .13, .26, .08, c(76, 171, 171))
		ell(0, 1.25, -.49, .35, .35, .40, orange)
		ell(0, 1.33, -.76, .23, .22, .17, white)
	}
	if role == RoleEnemy2 {
		ell(0, 1.42, .42, .25, .34, .20, c(255, 199, 222))
		for _, x := range []float32{-.11, .11} {
			ell(x, 1.23, .54, .09, .10, .08, base)
		}
	}
	if role == RoleNPC2 {
		ell(.31, 1.78, .36, .17, .12, .06, pink)
		ell(-.31, 1.78, .36, .17, .12, .06, pink)
		ell(0, 1.78, .42, .09, .09, .06, pink)
	}
	if role == RoleNPC3 {
		ell(0, 1.32, .38, .37, .40, .07, c(137, 202, 197))
		ell(0, 1.42, .45, .19, .12, .04, white)
		for _, x := range []float32{-.17, 0, .17} {
			ell(x, 2.58, 0, .13, .18, .13, c(213, 72, 60))
		}
		ell(0, 1.87, .4, .07, .12, .05, c(213, 72, 60))
	}
	if s.Crown {
		AddCrown(b, 2.75)
	}
	cm := &CharMeshes{Body: b.Build(), HeadTop: 2.8}
	limb := func(col rl.Color, foot bool) rl.Mesh {
		m := NewMB()
		m.Ellipsoid(v3(0, -.22, 0), v3(.16, .26, .16), 8, 10, col)
		if foot {
			m.Ellipsoid(v3(0, -.42, .09), v3(.19, .12, .24), 8, 10, col)
		}
		return m.Build()
	}
	legcol := base
	if role == RoleEnemy2 || role == RoleNPC1 || role == RoleNPC3 {
		legcol = c(220, 144, 42)
	}
	cm.Leg = limb(legcol, false)
	cm.Shin = limb(legcol, true)
	cm.ArmL = limb(base, false)
	cm.ArmR = limb(base, false)
	cm.ForeL = limb(base, false)
	cm.ForeR = limb(base, false)
	return cm
}
