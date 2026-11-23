package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() {
	for _, role := range []int{RoleEnemy1, RoleEnemy2, RoleEnemy3, RoleEnemy4, RoleNPC1, RoleNPC2, RoleNPC3} {
		r := role
		registerRealmChar(RealmCharCode(1, r), func(s CharStyle) *CharMeshes { return monstriCast(r, s) })
	}
}
func monstriCast(role int, s CharStyle) *CharMeshes {
	col := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	white := col(245, 243, 225)
	dark := col(48, 40, 52)
	green := col(88, 160, 85)
	stone := col(146, 152, 161)
	brown := col(139, 87, 55)
	orange := col(242, 137, 36)
	body := NewMB()
	leg := NewMB()
	shin := NewMB()
	al := NewMB()
	ar := NewMB()
	fl := NewMB()
	fr := NewMB()
	ell := func(b *MeshBuilder, x, y, z, rx, ry, rz float32, c rl.Color) {
		b.Ellipsoid(v3(x, y, z), v3(rx, ry, rz), 8, 12, c)
	}
	cyl := func(b *MeshBuilder, x, y, z, X, Y, Z, r float32, c rl.Color) {
		b.Cylinder(v3(x, y, z), v3(X, Y, Z), r, r, 8, c, true)
	}
	eye := func(x, y, z, sz float32) {
		ell(body, x, y, z, sz, sz*1.2, .055, white)
		ell(body, x, y, z+.055, sz*.45, sz*.55, .035, dark)
	}
	smile := func(y, z float32) {
		for i := -3; i <= 3; i++ {
			x := float32(i) * .065
			ell(body, x, y+float32(i*i)*.009, z, .035, .025, .025, dark)
		}
	}
	base := green
	clothes := col(96, 91, 133)
	head := green
	switch role {
	case RoleEnemy1:
		base = white
		clothes = white
		head = white
	case RoleEnemy2:
		base = green
		clothes = green
		head = orange
	case RoleEnemy3:
		base = white
		clothes = white
		head = white
	case RoleEnemy4:
		base = stone
		clothes = stone
		head = stone
	case RoleNPC2:
		base = brown
		clothes = col(145, 70, 61)
		head = brown
	case RoleNPC3:
		base = col(230, 223, 203)
		clothes = col(75, 104, 127)
		head = clothes
	}
	if role == RoleEnemy1 {
		ell(body, 0, 1.42, 0, .66, 1.04, .46, white)
		for i := -2; i <= 2; i++ {
			ell(body, float32(i)*.23, .56, 0, .19, .18, .43, white)
		}
		eye(-.22, 2.02, .39, .14)
		eye(.22, 2.02, .39, .14)
		ell(body, 0, 1.72, .46, .105, .14, .035, dark)
	} else {
		ell(body, 0, 1.45, 0, .52, .61, .34, clothes)
		if role != RoleNPC3 {
			ell(body, 0, 2.26, 0, .49, .48, .40, head)
		}
		if role != RoleEnemy3 && role != RoleNPC3 {
			eye(-.19, 2.32, .36, .105)
			eye(.19, 2.32, .36, .105)
			smile(2.08, .39)
		}
	}
	switch role {
	case RoleEnemy2:
		for _, x := range []float32{-.32, 0, .32} {
			ell(body, x, 2.25, .02, .23, .47, .39, orange)
		}
		eye(-.19, 2.32, .39, .105)
		eye(.19, 2.32, .39, .105)
		smile(2.08, .42)
		cyl(body, 0, 2.68, 0, .05, 2.86, 0, .065, brown)
		ell(body, -.35, 1.55, .31, .28, .10, .04, col(110, 200, 75))
		ell(body, .35, 1.25, .31, .28, .10, .04, col(110, 200, 75))
	case RoleEnemy3:
		for _, y := range []float32{1.10, 1.30, 1.50, 1.70, 1.90, 2.05, 2.48, 2.62} {
			cyl(body, -.40, y, .28, .40, y+.06, .28, .045, col(205, 194, 164))
		}
		for _, y := range []float32{1.10, 1.30, 1.50, 1.70, 1.90, 2.05, 2.48, 2.62} {
			cyl(body, -.40, y, -.28, .40, y+.06, -.28, .045, col(205, 194, 164))
		}
		eye(0, 2.31, .39, .18)
	case RoleEnemy4:
		for _, x := range []float32{-.35, .35} {
			body.Cylinder(v3(x, 2.61, 0), v3(x*1.15, 2.85, 0), .12, .01, 8, stone, true)
		}
		for _, x := range []float32{-1, 1} {
			body.Tri(v3(x*.36, 1.75, -.2), v3(x*.85, 2.03, -.3), v3(x*.62, 1.40, -.3), v3(0, 0, -1), stone)
		}
		ell(body, 0, 1.35, .30, .38, .36, .10, col(188, 190, 190))
	case RoleNPC1:
		for _, x := range []float32{-.53, .53} {
			cyl(body, x, 2.03, 0, x*1.25, 2.03, 0, .09, stone)
		}
		cyl(body, -.32, 2.48, .31, .25, 2.51, .31, .025, dark)
		for _, x := range []float32{-.2, 0, .2} {
			cyl(body, x, 2.45, .34, x, 2.55, .32, .018, dark)
		}
		for _, y := range []float32{1.25, 1.5, 1.75} {
			ell(body, 0, y, .34, .045, .045, .025, col(240, 205, 112))
		}
	case RoleNPC2:
		for _, x := range []float32{-.38, .38} {
			body.Cylinder(v3(x, 2.48, 0), v3(x*1.3, 2.92, 0), .19, .015, 8, brown, true)
		}
		ell(body, 0, 2.13, .39, .27, .15, .18, col(193, 142, 95))
		ell(body, 0, 2.20, .56, .09, .06, .05, dark)
		for _, x := range []float32{-.3, 0, .3} {
			cyl(body, x, 1.03, .32, x, 1.9, .32, .025, col(237, 180, 99))
		}
		for _, y := range []float32{1.2, 1.45, 1.7} {
			cyl(body, -.45, y, .32, .45, y, .32, .025, col(237, 180, 99))
		}
	case RoleNPC3:
		body.Cylinder(v3(0, 2.55, 0), v3(0, 2.62, 0), .55, .55, 16, dark, true)
		body.Cylinder(v3(0, 2.62, 0), v3(0, 2.90, 0), .32, .29, 16, clothes, true)
		for _, x := range []float32{-.22, .22} {
			ell(body, x, 2.30, .12, .17, .17, .035, dark)
			ell(body, x, 2.30, .16, .12, .12, .02, col(174, 213, 220))
		}
		cyl(body, -.07, 2.30, .13, .07, 2.30, .13, .025, dark)
		for _, y := range []float32{1.25, 1.5, 1.75} {
			ell(body, 0, y, .34, .04, .04, .025, white)
		}
	}
	if role == RoleEnemy1 {
		ell(leg, 0, 0, 0, .025, .025, .025, white)
		ell(shin, 0, .46, 0, .025, .025, .025, white)
	} else {
		ell(leg, 0, -.22, 0, .19, .25, .20, base)
		ell(shin, 0, -.20, 0, .17, .24, .18, base)
		ell(shin, 0, -.42, .08, .20, .11, .27, base)
	}
	for _, b := range []*MeshBuilder{al, ar} {
		ell(b, 0, -.18, 0, .17, .23, .18, clothes)
	}
	for _, b := range []*MeshBuilder{fl, fr} {
		ell(b, 0, -.18, 0, .15, .22, .16, base)
		ell(b, 0, -.38, .03, .19, .15, .18, base)
	}
	if role == RoleEnemy3 {
		band := col(205, 194, 164)
		for _, b := range []*MeshBuilder{leg, shin, al, ar, fl, fr} {
			for _, y := range []float32{-.12, -.29} {
				b.Cylinder(v3(0, y, 0), v3(0, y+.045, 0), .19, .19, 10, band, false)
			}
		}
		for _, b := range []*MeshBuilder{al, ar} {
			ell(b, 0, -.12, .24, .15, .15, .30, white)
		}
		for _, b := range []*MeshBuilder{fl, fr} {
			ell(b, 0, .18, .48, .16, .15, .25, white)
		}
	}
	if s.Crown {
		AddCrown(body, 2.90)
	}
	return &CharMeshes{Body: body.Build(), Leg: leg.Build(), Shin: shin.Build(), ArmL: al.Build(), ArmR: ar.Build(), ForeL: fl.Build(), ForeR: fr.Build(), HeadTop: 2.9}
}
