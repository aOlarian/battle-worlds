package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() {
	registerRealmChar(RealmCharCode(2, RoleHero), panzeHero)
	registerRealmChar(RealmCharCode(2, RoleBoss), panzeBoss)
	registerRealmVehicle("panze", panzeTaxi)
}

func panzeHero(s CharStyle) *CharMeshes { return panzePerson(s, false) }
func panzeBoss(s CharStyle) *CharMeshes { return panzePerson(s, true) }
func panzePerson(s CharStyle, boss bool) *CharMeshes {
	black := rl.NewColor(24, 28, 31, 255)
	lime := rl.NewColor(154, 238, 40, 255)
	white := rl.NewColor(245, 248, 244, 255)
	skin := rl.NewColor(238, 184, 151, 255)
	grey := rl.NewColor(95, 110, 119, 255)
	red := rl.NewColor(220, 48, 48, 255)
	suit := black
	if boss {
		suit = white
	}
	b := NewMB()
	b.Ellipsoid(v3(0, 1.47, 0), v3(.47, .55, .28), 10, 16, suit)
	b.Ellipsoid(v3(0, 1.02, 0), v3(.39, .19, .25), 8, 12, suit)
	head := black
	if boss {
		head = skin
	}
	b.Ellipsoid(v3(0, 2.32, 0), v3(.39, .44, .34), 12, 18, head)
	if !boss {
		for _, x := range []float32{-.18, .18} {
			b.Ellipsoid(v3(x, 2.37, .295), v3(.145, .19, .055), 8, 12, lime)
			b.Ellipsoid(v3(x, 2.37, .342), v3(.115, .155, .025), 8, 12, white)
		}
		b.Ellipsoid(v3(0, 1.55, .287), v3(.09, .13, .035), 8, 10, lime)
		for _, side := range []float32{-1, 1} {
			for i := 0; i < 4; i++ {
				y := 1.47 + float32(i)*.065
				b.Cylinder(v3(side*.06, y, .29), v3(side*.27, y+.10, .24), .018, .014, 5, lime, false)
			}
			b.Cylinder(v3(side*.34, 1.12, .18), v3(side*.4, 1.8, .12), .014, .014, 5, lime, false)
		}
	} else {
		for _, x := range []float32{-.18, .18} {
			b.Ellipsoid(v3(x, 2.37, .30), v3(.17, .17, .06), 8, 12, grey)
			b.Ellipsoid(v3(x, 2.37, .355), v3(.13, .13, .025), 8, 12, white)
			b.Sphere(v3(x, 2.37, .38), .045, black)
		}
		b.Cylinder(v3(-.10, 2.13, .32), v3(.10, 2.13, .32), .025, .025, 6, black, true)
		for i := 0; i < 9; i++ {
			x := float32(i-4) * .085
			b.Cylinder(v3(x, 2.58, -.02), v3(x*1.3, 2.82+float32(i%2)*.09, 0), .095, .005, 7, white, true)
		}
		for _, x := range []float32{-.15, .15} {
			b.Cylinder(v3(x, 1.05, .26), v3(x, 1.82, .24), .014, .014, 5, grey, false)
		}
		for _, y := range []float32{1.25, 1.48, 1.7} {
			b.Sphere(v3(0, y, .29), .028, grey)
		}
		// Horseshoe backpack, with small articulated arms.
		b.Cylinder(v3(-.48, 1.38, -.43), v3(-.48, 2.17, -.43), .16, .16, 10, red, true)
		b.Cylinder(v3(.48, 1.38, -.43), v3(.48, 2.17, -.43), .16, .16, 10, red, true)
		b.Cylinder(v3(-.48, 1.38, -.43), v3(.48, 1.38, -.43), .16, .16, 10, red, true)
		for _, side := range []float32{-1, 1} {
			b.Cylinder(v3(side*.48, 2.05, -.43), v3(side*.48, 2.21, -.43), .165, .165, 10, grey, true)
			b.Cylinder(v3(side*.48, 1.65, -.43), v3(side*.76, 1.8, -.40), .055, .05, 8, grey, true)
			b.Cylinder(v3(side*.76, 1.8, -.4), v3(side*.85, 1.57, -.25), .05, .04, 8, grey, true)
			b.Sphere(v3(side*.85, 1.57, -.25), .08, grey)
		}
	}
	if s.Crown {
		AddCrown(b, 2.78)
	}
	cm := &CharMeshes{Body: b.Build(), HeadTop: 2.78}
	leg := NewMB()
	leg.Ellipsoid(v3(0, -.23, 0), v3(.19, .25, .19), 8, 12, suit)
	cm.Leg = leg.Build()
	shin := NewMB()
	shin.Ellipsoid(v3(0, -.20, 0), v3(.17, .24, .17), 8, 12, suit)
	shin.Ellipsoid(v3(0, -.39, .09), v3(.19, .11, .27), 8, 12, black)
	cm.Shin = shin.Build()
	arm := func() rl.Mesh {
		m := NewMB()
		m.Ellipsoid(v3(0, -.18, 0), v3(.17, .23, .17), 8, 12, suit)
		return m.Build()
	}
	cm.ArmL = arm()
	cm.ArmR = arm()
	fore := func() rl.Mesh {
		m := NewMB()
		m.Ellipsoid(v3(0, -.17, 0), v3(.15, .22, .15), 8, 12, suit)
		hand := black
		if boss {
			hand = skin
		}
		m.Sphere(v3(0, -.38, 0), .14, hand)
		if !boss {
			m.Cylinder(v3(0, -.30, 0), v3(0, -.19, 0), .18, .18, 10, grey, true)
			m.Sphere(v3(0, -.24, .17), .06, lime)
		}
		return m.Build()
	}
	cm.ForeL = fore()
	cm.ForeR = fore()
	return cm
}
func panzeTaxi(col rl.Color) rl.Mesh {
	m := NewMB()
	dark := rl.NewColor(30, 34, 39, 255)
	white := rl.NewColor(245, 245, 230, 255)
	glass := rl.NewColor(92, 154, 183, 255)
	yellow := rl.NewColor(255, 210, 50, 255)
	m.BeveledBox(v3(0, .7, 0), v3(2, .72, 4.2), .22, col)
	m.BeveledBox(v3(0, 1.3, -.25), v3(1.7, .68, 2.3), .2, col)
	m.Quad(v3(-.7, 1.09, .91), v3(.7, 1.09, .91), v3(.62, 1.53, .92), v3(-.62, 1.53, .92), v3(0, 0, 1), glass)
	m.Quad(v3(-.7, 1.09, -1.41), v3(.7, 1.09, -1.41), v3(.62, 1.53, -1.42), v3(-.62, 1.53, -1.42), v3(0, 0, -1), glass)
	for _, side := range []float32{-1, 1} {
		m.Quad(v3(side*.86, 1.1, -1.2), v3(side*.86, 1.1, .7), v3(side*.86, 1.53, .6), v3(side*.86, 1.53, -1.1), v3(side, 0, 0), glass)
		for i := 0; i < 16; i++ {
			for j := 0; j < 2; j++ {
				c := white
				if (i+j)%2 == 0 {
					c = dark
				}
				m.Box(v3(side*1.003, .72+float32(j)*.1, -1.95+float32(i)*.25), v3(.02, .1, .25), c)
			}
		}
		for _, z := range []float32{-1.35, 1.35} {
			m.Cylinder(v3(side*.87, .42, z), v3(side*1.13, .42, z), .42, .42, 14, dark, true)
			m.Cylinder(v3(side*1.13, .42, z), v3(side*1.15, .42, z), .22, .22, 12, white, true)
		}
	}
	m.BeveledBox(v3(0, 1.79, -.25), v3(.85, .3, .4), .08, yellow)
	m.Box(v3(0, 1.8, -.035), v3(.52, .12, .02), dark)
	for _, x := range []float32{-.65, .65} {
		m.Box(v3(x, .72, 2.11), v3(.35, .18, .04), white)
		m.Box(v3(x, .72, -2.11), v3(.35, .18, .04), rl.NewColor(225, 55, 50, 255))
	}
	return m.Build()
}
