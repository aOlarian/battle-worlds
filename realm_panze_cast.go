package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() {
	for _, role := range []int{RoleEnemy1, RoleEnemy2, RoleEnemy3, RoleEnemy4, RoleNPC1, RoleNPC2, RoleNPC3} {
		r := role
		registerRealmChar(RealmCharCode(2, r), func(s CharStyle) *CharMeshes { return panzeCast(r, s) })
	}
}
func panzeCast(role int, s CharStyle) *CharMeshes {
	c := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	dark, steel, green, purple := c(42, 48, 59), c(143, 160, 174), c(64, 167, 104), c(139, 80, 177)
	skin, hair, white := c(229, 174, 135), c(73, 47, 36), c(245, 243, 229)
	red, gold, blue := c(243, 48, 56), c(246, 191, 50), c(65, 122, 166)
	robot := role <= RoleEnemy4
	shirt, pants := blue, dark
	if role == RoleNPC2 {
		shirt = c(235, 116, 157)
		pants = c(83, 104, 158)
	}
	if role == RoleNPC3 {
		shirt = c(138, 110, 76)
		pants = c(75, 78, 87)
	}
	if robot {
		shirt = green
		pants = steel
	}
	if role == RoleEnemy2 {
		shirt = purple
	}
	if role == RoleEnemy3 {
		shirt = c(67, 170, 190)
	}
	if role == RoleEnemy4 {
		shirt = steel
		pants = dark
	}
	m := NewMB()
	ell := func(x, y, z, a, b, d float32, col rl.Color) { m.Ellipsoid(v3(x, y, z), v3(a, b, d), 8, 12, col) }
	if robot {
		switch role {
		case RoleEnemy1:
			ell(0, 1.43, -.03, .56, .64, .45, green)
			ell(0, 1.53, -.30, .48, .53, .25, purple)
			ell(0, 2.12, .10, .43, .38, .38, green)
			for _, x := range []float32{-.25, .25} {
				m.Cylinder(v3(x, 2.34, 0), v3(x*1.7, 2.85, .05), .045, .025, 8, steel, true)
				ell(x*1.7, 2.85, .05, .08, .08, .08, purple)
			}
		case RoleEnemy2:
			ell(0, 1.4, 0, .53, .48, .48, purple)
			ell(0, 1.93, .13, .41, .33, .36, steel)
			pts := []rl.Vector3{v3(0, 1.35, -.35), v3(0, 1.55, -.75), v3(0, 2.05, -.95), v3(0, 2.48, -.72), v3(0, 2.53, -.34)}
			for i := 0; i < len(pts)-1; i++ {
				m.Cylinder(pts[i], pts[i+1], .13-float32(i)*.02, .11-float32(i)*.02, 9, purple, true)
			}
			m.Emis = 1
			ell(0, 2.53, -.3, .13, .13, .13, gold)
			m.Emis = 0
		case RoleEnemy3:
			ell(0, 1.65, 0, .56, .56, .52, shirt)
			m.Cylinder(v3(0, 2.14, 0), v3(0, 2.49, 0), .06, .06, 8, steel, true)
			ell(0, 2.49, 0, .13, .09, .13, gold)
			m.BeveledBox(v3(0, 2.51, 0), v3(1.55, .07, .18), .03, dark)
			m.BeveledBox(v3(0, 2.51, 0), v3(.18, .07, 1.05), .03, dark)
		case RoleEnemy4:
			ell(0, 1.48, 0, .66, .68, .45, steel)
			ell(0, 2.14, 0, .48, .43, .4, dark)
			ell(0, 2.03, .32, .34, .19, .15, steel)
			ell(0, 1.45, .40, .38, .38, .08, dark)
		}
		if role == RoleEnemy1 || role == RoleEnemy2 {
			for _, side := range []float32{-1, 1} {
				for _, y := range []float32{1.22, 1.52} {
					m.Cylinder(v3(side*.38, y, -.15), v3(side*.72, y-.12, -.28), .045, .04, 7, steel, true)
					m.Cylinder(v3(side*.72, y-.12, -.28), v3(side*.82, y-.38, -.18), .04, .025, 7, steel, true)
				}
			}
		}
		if role == RoleEnemy4 {
			for _, x := range []float32{-.3, .3} {
				m.BeveledBox(v3(x, 1.75, .34), v3(.26, .18, .10), .04, steel)
			}
		}
		m.Emis = 1
		for _, x := range []float32{-.16, .16} {
			ell(x, 2.12, .40, .085, .09, .06, red)
		}
		m.Emis = 0
	} else {
		ell(0, 1.48, 0, .46, .55, .3, shirt)
		ell(0, 2.29, 0, .37, .42, .34, skin)
		for _, x := range []float32{-.14, .14} {
			ell(x, 2.34, .30, .075, .095, .045, white)
			ell(x, 2.34, .342, .035, .05, .025, dark)
		}
		ell(0, 2.15, .32, .12, .035, .025, hair)
		if role == RoleNPC1 || role == RoleNPC3 {
			ell(0, 2.63, 0, .39, .16, .36, dark)
			m.BeveledBox(v3(0, 2.59, .31), v3(.55, .06, .3), .03, dark)
		} else {
			ell(0, 2.59, -.07, .39, .18, .34, hair)
			ell(-.31, 2.35, -.13, .14, .3, .14, hair)
			ell(.31, 2.35, -.13, .14, .3, .14, hair)
		}
		if role == RoleNPC1 {
			ell(-.08, 2.18, .35, .12, .045, .035, hair)
			ell(.08, 2.18, .35, .12, .045, .035, hair)
		}
		if role == RoleNPC3 {
			m.BeveledBox(v3(0, 1.49, .27), v3(.65, .72, .08), .04, c(99, 72, 49))
			for _, y := range []float32{1.25, 1.45, 1.65} {
				ell(0, y, .32, .035, .035, .025, gold)
			}
			for i := 0; i < 4; i++ {
				m.Box(v3(-.62, 1.15+float32(i)*.055, .35), v3(.52, .045, .35), white)
			}
		}
		if role == RoleNPC2 {
			m.Cylinder(v3(.63, 1.24, .1), v3(.85, 2.68, .1), .012, .012, 6, white, true)
			ell(.85, 2.95, .1, .28, .36, .27, red)
		}
	}
	top := float32(2.75)
	if role == RoleNPC2 {
		top = 3.31
	}
	if s.Crown {
		AddCrown(m, top)
	}
	cm := &CharMeshes{Body: m.Build(), HeadTop: top}
	limb := func(col rl.Color, radius float32, foot bool) rl.Mesh {
		b := NewMB()
		b.Ellipsoid(v3(0, -.22, 0), v3(radius, .25, radius), 7, 10, col)
		if foot {
			b.Ellipsoid(v3(0, -.42, .10), v3(radius*1.2, .12, radius*1.5), 6, 10, dark)
		}
		return b.Build()
	}
	rad := float32(.18)
	if role == RoleEnemy3 {
		rad = .055
	}
	if role == RoleEnemy4 {
		rad = .24
	}
	cm.Leg = limb(pants, rad, false)
	cm.Shin = limb(pants, rad, true)
	cm.ArmL = limb(shirt, rad, false)
	cm.ArmR = limb(shirt, rad, false)
	fore := func(side float32) rl.Mesh {
		b := NewMB()
		b.Ellipsoid(v3(0, -.2, 0), v3(rad, .24, rad), 7, 10, shirt)
		hand := skin
		if robot {
			hand = steel
		}
		size := float32(.15)
		if role == RoleEnemy4 {
			size = .31
		}
		b.Ellipsoid(v3(0, -.39, .06), v3(size, size, size), 7, 10, hand)
		if role == RoleEnemy2 {
			for _, x := range []float32{-.13, .13} {
				b.Ellipsoid(v3(x, -.48, .15), v3(.08, .16, .09), 6, 8, purple)
			}
		}
		return b.Build()
	}
	cm.ForeL = fore(-1)
	cm.ForeR = fore(1)
	return cm
}
