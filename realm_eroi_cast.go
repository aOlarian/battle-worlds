package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() {
	for _, role := range []int{RoleEnemy1, RoleEnemy2, RoleEnemy3, RoleEnemy4, RoleNPC1, RoleNPC2, RoleNPC3} {
		r := role
		registerRealmChar(RealmCharCode(3, r), func(s CharStyle) *CharMeshes { return eroiCast(r, s) })
	}
}
func eroiCast(role int, s CharStyle) *CharMeshes {
	c := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	dark := c(35, 48, 61)
	metal := c(150, 177, 186)
	light := c(219, 232, 231)
	cyan := c(95, 255, 235)
	skin := c(224, 165, 125)
	hair := c(71, 42, 33)
	shirt := c(215, 102, 87)
	pants := c(54, 86, 111)
	hand := skin
	if role <= RoleEnemy4 {
		shirt = metal
		pants = dark
		hand = metal
	}
	if role == RoleEnemy1 {
		shirt = c(80, 141, 157)
	}
	if role == RoleEnemy2 {
		shirt = c(130, 160, 175)
	}
	if role == RoleEnemy3 {
		shirt = c(230, 150, 65)
	}
	if role == RoleEnemy4 {
		shirt = light
		hand = c(103, 204, 113)
	}
	if role == RoleNPC2 {
		shirt = c(54, 91, 146)
		pants = c(37, 59, 96)
	}
	if role == RoleNPC3 {
		shirt = c(130, 200, 135)
		pants = c(95, 105, 164)
	}
	cm := &CharMeshes{}
	b := NewMB()
	if role == RoleEnemy1 {
		b.Ellipsoid(v3(0, 1.48, 0), v3(.27, .56, .22), 8, 12, shirt)
	} else if role == RoleEnemy3 {
		b.BeveledBox(v3(0, 1.48, 0), v3(.76, .88, .48), .16, shirt)
	} else {
		b.Ellipsoid(v3(0, 1.48, 0), v3(.46, .56, .28), 8, 12, shirt)
	}
	top := float32(2.78)
	switch role {
	case RoleEnemy1:
		b.BeveledBox(v3(0, 2.35, 0), v3(.9, .72, .55), .13, metal)
		b.BeveledBox(v3(0, 2.35, .29), v3(.74, .51, .08), .07, dark)
		b.Emis = 1
		for _, x := range []float32{-.19, .19} {
			b.Box(v3(x, 2.4, .34), v3(.13, .13, .025), cyan)
		}
		b.Emis = 0
		b.Cylinder(v3(-.3, 2.73, 0), v3(-.48, 2.91, 0), .035, .025, 8, metal, true)
		b.Sphere(v3(-.48, 2.91, 0), .07, cyan)
	case RoleEnemy2:
		b.Ellipsoid(v3(0, 1.95, 0), v3(.69, .72, .48), 10, 16, metal)
		b.Ellipsoid(v3(0, 2.12, .45), v3(.27, .27, .09), 8, 12, dark)
		b.Emis = 1
		b.Sphere(v3(0, 2.12, .53), .14, cyan)
		b.Emis = 0
		top = 2.67
	case RoleEnemy3:
		b.BeveledBox(v3(0, 2.32, 0), v3(.78, .7, .57), .14, shirt)
		b.BeveledBox(v3(0, 2.35, .3), v3(.61, .32, .05), .06, dark)
		b.Emis = 1
		for _, x := range []float32{-.16, .16} {
			b.Sphere(v3(x, 2.36, .34), .065, cyan)
		}
		b.Emis = 0
		b.Cylinder(v3(0, 2.67, 0), v3(0, 2.94, 0), .035, .035, 8, metal, true)
		b.Sphere(v3(0, 2.95, 0), .075, c(255, 92, 90))
		top = 3.02
	case RoleEnemy4:
		green := c(105, 205, 116)
		b.Ellipsoid(v3(0, 2.36, 0), v3(.53, .46, .37), 9, 14, green)
		for _, x := range []float32{-.3, 0, .3} {
			b.Ellipsoid(v3(x, 2.4, .32), v3(.115, .14, .07), 6, 8, light)
			b.Sphere(v3(x, 2.4, .38), .055, dark)
		}
		b.Ellipsoid(v3(0, 2.16, .35), v3(.12, .035, .025), 4, 8, dark)
		top = 2.82
	default:
		b.Ellipsoid(v3(0, 2.35, 0), v3(.38, .43, .34), 9, 14, skin)
		b.Ellipsoid(v3(0, 2.64, -.04), v3(.39, .19, .35), 7, 12, hair)
		if role == RoleNPC1 {
			b.Ellipsoid(v3(0, 2.32, -.23), v3(.38, .42, .16), 7, 12, hair)
		}
		for _, x := range []float32{-.14, .14} {
			b.Ellipsoid(v3(x, 2.39, .31), v3(.075, .09, .045), 5, 8, light)
			b.Sphere(v3(x, 2.39, .35), .04, dark)
		}
		b.Ellipsoid(v3(0, 2.19, .33), v3(.12, .035, .025), 4, 8, c(150, 70, 66))
		if role == RoleNPC2 {
			b.Ellipsoid(v3(0, 2.7, 0), v3(.4, .16, .36), 6, 12, shirt)
			b.BeveledBox(v3(0, 2.65, .33), v3(.55, .06, .3), .025, shirt)
			b.Cylinder(v3(0, 1.68, .28), v3(0, 1.45, .28), .018, .018, 8, light, true)
			b.BeveledBox(v3(0, 1.43, .31), v3(.16, .09, .08), .025, c(245, 211, 104))
		}
		if role == RoleNPC1 {
			b.BeveledBox(v3(0, 1.4, .33), v3(.43, .27, .19), .04, dark)
			b.Cylinder(v3(0, 1.4, .42), v3(0, 1.4, .54), .11, .11, 12, metal, true)
			b.Cylinder(v3(0, 1.4, .54), v3(0, 1.4, .57), .075, .075, 12, dark, true)
		}
	}
	if s.Crown {
		AddCrown(b, top)
	}
	cm.Body = b.Build()
	l := NewMB()
	l.Ellipsoid(v3(0, -.22, 0), v3(.19, .27, .2), 6, 10, pants)
	cm.Leg = l.Build()
	l = NewMB()
	l.Ellipsoid(v3(0, -.22, 0), v3(.18, .27, .19), 6, 10, pants)
	l.Ellipsoid(v3(0, -.42, .09), v3(.2, .1, .29), 5, 10, dark)
	cm.Shin = l.Build()
	for side := -1; side <= 1; side += 2 {
		a := NewMB()
		a.Ellipsoid(v3(0, -.18, 0), v3(.16, .23, .17), 6, 10, shirt)
		upper := a.Build()
		a = NewMB()
		a.Ellipsoid(v3(0, -.17, 0), v3(.13, .22, .14), 6, 10, shirt)
		a.Sphere(v3(0, -.37, .04), .14, hand)
		if side == 1 && role == RoleNPC3 {
			a.Cylinder(v3(0, -.35, .1), v3(0, -.08, .1), .04, .13, 10, c(215, 166, 94), true)
			a.Sphere(v3(0, -.03, .1), .14, c(255, 147, 175))
		}
		fore := a.Build()
		if side < 0 {
			cm.ArmL = upper
			cm.ForeL = fore
		} else {
			cm.ArmR = upper
			cm.ForeR = fore
		}
	}
	return cm
}
