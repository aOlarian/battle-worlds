package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() {
	for _, role := range []int{RoleEnemy1, RoleEnemy2, RoleEnemy3, RoleEnemy4, RoleNPC1, RoleNPC2, RoleNPC3} {
		r := role
		registerRealmChar(RealmCharCode(6, r), func(s CharStyle) *CharMeshes { return halloweenCast(r, s) })
	}
}

func halloweenCast(role int, s CharStyle) *CharMeshes {
	c := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	orange, yellow, green := c(239, 112, 23), c(255, 222, 55), c(73, 137, 61)
	dark, white, skin := c(35, 30, 43), c(242, 239, 225), c(231, 180, 148)
	purple, brown, pink := c(114, 69, 150), c(139, 88, 49), c(240, 122, 151)
	body, head, limb, hand, shoe := purple, skin, purple, skin, dark
	switch role {
	case RoleEnemy1:
		body, head, limb, hand, shoe = orange, orange, green, green, green
	case RoleEnemy2:
		body, head, limb, hand, shoe = dark, white, white, white, white
	case RoleEnemy3:
		body, head, limb, hand, shoe = dark, dark, dark, dark, dark
	case RoleEnemy4:
		body, head, limb, hand, shoe = purple, purple, purple, purple, purple
	case RoleNPC1:
		body, head, limb, hand, shoe = white, white, white, white, c(106, 99, 142)
	case RoleNPC2:
		body, head, limb, hand, shoe = c(59, 133, 151), skin, c(59, 133, 151), skin, brown
	case RoleNPC3:
		body, head, limb, hand, shoe = purple, skin, purple, skin, brown
	}
	m := NewMB()
	ell := func(x, y, z, rx, ry, rz float32, col rl.Color) { m.Ellipsoid(v3(x, y, z), v3(rx, ry, rz), 8, 12, col) }
	if role != RoleEnemy2 {
		ell(0, 1.42, 0, .52, .59, .37, body)
	}
	if role == RoleEnemy1 {
		for _, x := range []float32{-.37, -.19, 0, .19, .37} {
			ell(x, 1.55, 0, .24, .61, .39, orange)
		}
	}
	if role == RoleEnemy2 {
		ell(0, 1.42, 0, .16, .48, .15, white)
		for _, y := range []float32{1.32, 1.48, 1.64, 1.78} {
			ell(0, y, .08, .39, .065, .19, white)
		}
		ell(-.16, 1.91, .32, .17, .10, .06, orange)
		ell(.16, 1.91, .32, .17, .10, .06, orange)
		ell(0, 1.91, .39, .07, .07, .06, yellow)
	}
	if role == RoleNPC1 {
		ell(0, 1.10, 0, .61, .47, .42, white)
	}
	ell(0, 2.25, 0, .43, .44, .38, head)
	if role == RoleEnemy3 || role == RoleEnemy4 {
		for _, x := range []float32{-.29, .29} {
			m.Cylinder(v3(x, 2.46, 0), v3(x*1.35, 2.94, 0), .18, .015, 9, head, true)
		}
	}
	if role == RoleEnemy1 {
		m.Cylinder(v3(0, 2.60, 0), v3(.08, 2.86, 0), .10, .07, 8, green, true)
	}
	if role == RoleNPC2 {
		ell(0, 1.15, .34, .42, .07, .06, brown)
		ell(0, 2.52, 0, .45, .18, .39, brown)
		m.Cylinder(v3(0, 2.55, 0), v3(0, 2.78, 0), .36, .28, 12, brown, true)
		ell(0, 2.57, .30, .33, .07, .12, orange)
	}
	if role == RoleNPC3 {
		ell(0, 2.39, -.20, .46, .37, .30, c(195, 194, 186))
		m.Cylinder(v3(0, 2.53, 0), v3(0, 2.59, 0), .55, .55, 14, dark, true)
		m.Cylinder(v3(0, 2.57, 0), v3(.13, 3.08, 0), .38, .02, 14, purple, true)
		ell(0, 1.36, .35, .39, .45, .08, green)
		ell(0, 1.19, .42, .20, .12, .035, c(119, 190, 88))
	}
	eye := dark
	if role == RoleEnemy1 || role == RoleEnemy3 {
		eye = yellow
		m.Emis = 1
	}
	for _, x := range []float32{-.16, .16} {
		ell(x, 2.32, .345, .105, .13, .055, eye)
		if role == RoleEnemy3 {
			m.Emis = 0
			ell(x, 2.32, .40, .025, .08, .02, dark)
			m.Emis = 1
		}
	}
	m.Emis = 0
	if role != RoleNPC1 {
		if role == RoleEnemy3 || role == RoleEnemy4 {
			ell(0, 2.18, .36, .17, .10, .08, white)
			ell(0, 2.23, .43, .07, .045, .04, pink)
		}
		smile := dark
		if role == RoleEnemy1 {
			smile = yellow
			m.Emis = 1
		}
		for i := -3; i <= 3; i++ {
			x := float32(i) * .055
			ell(x, 2.08+float32(i*i)*.009, .36, .035, .025, .03, smile)
		}
		m.Emis = 0
		if role == RoleEnemy4 {
			for _, x := range []float32{-.10, .10} {
				m.Cylinder(v3(x, 2.12, .43), v3(x, 2.05, .43), .035, .003, 6, white, true)
			}
		}
	}
	if s.Crown {
		AddCrown(m, 3.1)
	}
	cm := &CharMeshes{Body: m.Build(), HeadTop: 3.1}
	l := NewMB()
	l.Ellipsoid(v3(0, -.22, 0), v3(.16, .25, .17), 6, 10, limb)
	cm.Leg = l.Build()
	l = NewMB()
	l.Ellipsoid(v3(0, -.20, 0), v3(.15, .23, .16), 6, 10, limb)
	l.Ellipsoid(v3(0, -.42, .08), v3(.19, .12, .25), 6, 10, shoe)
	cm.Shin = l.Build()
	for _, side := range []float32{-1, 1} {
		a := NewMB()
		a.Ellipsoid(v3(0, -.18, 0), v3(.14, .23, .15), 6, 10, limb)
		if role == RoleEnemy4 {
			a.Ellipsoid(v3(side*.34, -.12, -.03), v3(.48, .19, .12), 6, 10, purple)
			for i := 0; i < 3; i++ {
				x := side * (.2 + float32(i)*.24)
				a.Ellipsoid(v3(x, -.28, 0), v3(.18, .20, .07), 6, 10, purple)
			}
		}
		upper := a.Build()
		a = NewMB()
		a.Ellipsoid(v3(0, -.19, 0), v3(.13, .22, .14), 6, 10, limb)
		a.Ellipsoid(v3(0, -.39, .05), v3(.15, .13, .14), 6, 10, hand)
		if (side == -1 && role == RoleNPC1) || (side == 1 && (role == RoleNPC2 || role == RoleNPC3)) {
			a.Cylinder(v3(.10, -.68, .10), v3(.10, -.43, .10), .24, .29, 12, brown, true)
			a.Cylinder(v3(-.16, -.43, .10), v3(-.16, -.18, .10), .025, .025, 6, brown, true)
			a.Cylinder(v3(.36, -.43, .10), v3(.36, -.18, .10), .025, .025, 6, brown, true)
			a.Cylinder(v3(-.16, -.18, .10), v3(.36, -.18, .10), .025, .025, 6, brown, true)
			for i := 0; i < 3; i++ {
				col := pink
				if role == RoleNPC3 {
					col = c(220, 61, 53)
				} else if i == 1 {
					col = yellow
				}
				a.Sphere(v3(-.06+float32(i)*.16, -.40, .10), .105, col)
			}
		}
		fore := a.Build()
		if side < 0 {
			cm.ArmL, cm.ForeL = upper, fore
		} else {
			cm.ArmR, cm.ForeR = upper, fore
		}
	}
	if role == RoleEnemy3 {
		t := NewMB()
		t.Cylinder(v3(0, 0, 0), v3(.38, .35, -.35), .13, .11, 9, dark, true)
		t.Cylinder(v3(.38, .35, -.35), v3(.48, 1.05, -.40), .11, .06, 9, dark, true)
		t.Sphere(v3(.48, 1.05, -.40), .07, dark)
		cm.Tail = t.Build()
		cm.HasTail = true
	}
	return cm
}
