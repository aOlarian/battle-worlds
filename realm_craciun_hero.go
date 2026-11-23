package main

import rl "github.com/gen2brain/raylib-go/raylib"

var christmasRed = rl.NewColor(190, 34, 43, 255)
var christmasGreen = rl.NewColor(35, 139, 75, 255)
var christmasCream = rl.NewColor(250, 239, 215, 255)
var christmasGold = rl.NewColor(238, 184, 47, 255)

func christmasCharacter(s CharStyle, boss bool) *CharMeshes {
	skin := rl.NewColor(244, 187, 157, 255)
	outfit := christmasGreen
	if boss {
		skin = rl.NewColor(132, 145, 68, 255)
		outfit = skin
	}
	dark := rl.NewColor(57, 43, 31, 255)
	mb := NewMB()
	if boss {
		mb.Ellipsoid(v3(0, 1.43, 0), v3(.65, .64, .47), 9, 16, skin)
		// Fur tufts give the giant a soft, shaggy outline.
		for _, side := range []float32{-1, 1} {
			for _, y := range []float32{1.15, 1.38, 1.61, 1.82} {
				mb.Cylinder(v3(side*.52, y, 0), v3(side*.72, y-.15, .02), .12, .01, 6, shade(skin, .85), true)
			}
		}
		for _, x := range []float32{-.48, -.32, 0, .32, .48} {
			mb.Cylinder(v3(x, 1.18, 0), v3(x*1.12, .94, .03), .14, .01, 6, skin, true)
		}
	} else {
		mb.Ellipsoid(v3(0, 1.48, 0), v3(.48, .55, .32), 9, 16, outfit)
		mb.Cylinder(v3(0, 1.12, 0), v3(0, 1.38, 0), .49, .43, 12, outfit, true)
		mb.Cylinder(v3(0, 1.32, 0), v3(0, 1.48, 0), .45, .44, 12, dark, true)
		mb.BeveledBox(v3(0, 1.4, .447), v3(.22, .2, .06), .025, christmasGold)
		mb.Box(v3(0, 1.4, .482), v3(.11, .10, .02), dark)
		for _, x := range []float32{-.28, -.14, 0, .14, .28} {
			mb.Tri(v3(x-.1, 1.95, .22), v3(x+.1, 1.95, .22), v3(x, 1.75, .34), v3(0, 0, 1), christmasCream)
		}
	}
	mb.Ellipsoid(v3(0, 2.29, 0), v3(.42, .44, .37), 10, 16, skin)
	for _, x := range []float32{-.16, .16} {
		mb.Ellipsoid(v3(x, 2.34, .33), v3(.095, .12, .055), 6, 10, christmasCream)
		mb.Ellipsoid(v3(x, 2.34, .38), v3(.045, .07, .025), 6, 8, dark)
		if boss {
			mb.Cylinder(v3(x-.11, 2.49, .35), v3(x+.1, 2.46, .36), .055, .045, 7, dark, true)
		} else {
			mb.Ellipsoid(v3(x*1.65, 2.18, .29), v3(.09, .055, .025), 5, 8, rl.NewColor(227, 104, 106, 255))
		}
	}
	mb.Ellipsoid(v3(0, 2.2, .38), v3(.09, .065, .07), 6, 8, skin)
	mb.Cylinder(v3(-.12, 2.09, .344), v3(.12, 2.09, .344), .025, .025, 8, dark, true)
	if boss {
		mb.Cylinder(v3(0, 1.93, 0), v3(0, 2.05, 0), .38, .38, 14, christmasRed, true)
		mb.BeveledBox(v3(.24, 1.68, .4), v3(.22, .56, .08), .04, christmasRed)
		for _, x := range []float32{-.32, -.17, 0, .17, .32} {
			mb.Cylinder(v3(x, 2.58, -.08), v3(x*1.2, 2.82, -.08), .12, .01, 6, shade(skin, .7), true)
		}
	} else {
		for _, x := range []float32{-1, 1} {
			mb.Cylinder(v3(x*.35, 2.28, 0), v3(x*.65, 2.48, 0), .13, .01, 8, skin, true)
		}
	}
	hat := christmasGreen
	if boss {
		hat = christmasRed
	}
	mb.Cylinder(v3(0, 2.61, 0), v3(0, 2.71, 0), .42, .4, 16, christmasCream, true)
	mb.Cylinder(v3(0, 2.68, 0), v3(.19, 3.06, -.04), .37, .17, 14, hat, true)
	mb.Cylinder(v3(.19, 3.06, -.04), v3(.36, 2.96, 0), .17, .02, 12, hat, true)
	mb.Sphere(v3(.36, 2.96, 0), .09, christmasGold)
	if s.Crown {
		AddCrown(mb, 3.08)
	}
	cm := &CharMeshes{Body: mb.Build(), HeadTop: 3.08}
	leg := NewMB()
	leg.Ellipsoid(v3(0, -.21, 0), v3(.18, .27, .18), 7, 12, christmasCream)
	if boss {
		leg.Ellipsoid(v3(0, -.21, 0), v3(.22, .27, .21), 7, 12, skin)
	} else {
		for _, y := range []float32{-.12, -.32} {
			leg.Cylinder(v3(0, y-.045, 0), v3(0, y+.045, 0), .185, .185, 12, christmasRed, true)
		}
	}
	cm.Leg = leg.Build()
	shin := NewMB()
	shin.Ellipsoid(v3(0, -.18, 0), v3(.16, .23, .16), 7, 12, christmasCream)
	if boss {
		shin.Ellipsoid(v3(0, -.18, 0), v3(.21, .24, .2), 7, 12, skin)
	} else {
		shin.Cylinder(v3(0, -.22, 0), v3(0, -.12, 0), .165, .165, 12, christmasRed, true)
	}
	shoe := dark
	if boss {
		shoe = skin
	}
	shin.Ellipsoid(v3(0, -.4, .11), v3(.2, .13, .3), 7, 12, shoe)
	if !boss {
		shin.Cylinder(v3(0, -.39, .3), v3(0, -.23, .43), .12, .025, 8, shoe, true)
	}
	cm.Shin = shin.Build()
	arm := NewMB()
	arm.Ellipsoid(v3(0, -.18, 0), v3(.2, .25, .2), 7, 12, outfit)
	cm.ArmL = arm.Build()
	cm.ArmR = cm.ArmL
	fore := NewMB()
	fore.Ellipsoid(v3(0, -.16, 0), v3(.16, .22, .16), 7, 12, outfit)
	fore.Sphere(v3(0, -.38, 0), .16, skin)
	cm.ForeL = fore.Build()
	cm.ForeR = cm.ForeL
	return cm
}

func christmasSleigh(_ rl.Color) rl.Mesh {
	b := NewMB()
	b.BeveledBox(v3(0, .68, 0), v3(1.9, .32, 3.7), .14, christmasRed)
	for _, x := range []float32{-.87, .87} {
		b.BeveledBox(v3(x, 1.05, -.15), v3(.18, .7, 3.25), .08, christmasRed)
		b.Cylinder(v3(x, .2, -1.8), v3(x, .2, 1.65), .09, .09, 10, christmasGold, true)
		pts := []rl.Vector3{v3(x, .2, 1.65), v3(x, .25, 1.94), v3(x, .42, 2.16), v3(x, .65, 2.22), v3(x, .78, 2.08)}
		for i := 0; i < len(pts)-1; i++ {
			b.Cylinder(pts[i], pts[i+1], .09, .09, 10, christmasGold, true)
		}
		b.Cylinder(v3(x, .2, -1.6), v3(x, .55, -1.5), .07, .07, 8, christmasGold, true)
		b.Cylinder(v3(x, .2, 1.2), v3(x, .55, 1.1), .07, .07, 8, christmasGold, true)
	}
	b.BeveledBox(v3(0, 1.05, -1.65), v3(1.8, .8, .18), .08, christmasRed)
	b.BeveledBox(v3(0, .99, .1), v3(1.4, .3, .7), .1, rl.NewColor(112, 55, 38, 255))
	b.Ellipsoid(v3(0, 1.35, -1.13), v3(.63, .65, .55), 9, 14, rl.NewColor(189, 138, 77, 255))
	for i, c := range []rl.Color{christmasGreen, rl.NewColor(74, 150, 220, 255), christmasRed} {
		x := float32(i-1) * .4
		b.BeveledBox(v3(x, 1.83, -1.13), v3(.38, .37, .4), .04, c)
		b.Box(v3(x, 2.025, -1.13), v3(.09, .025, .41), christmasGold)
		b.Box(v3(x, 2.025, -1.13), v3(.39, .025, .09), christmasGold)
	}
	return b.Build()
}
func init() {
	registerRealmChar(RealmCharCode(4, RoleHero), func(s CharStyle) *CharMeshes { return christmasCharacter(s, false) })
	registerRealmChar(RealmCharCode(4, RoleBoss), func(s CharStyle) *CharMeshes { return christmasCharacter(s, true) })
	registerRealmVehicle("craciun", christmasSleigh)
}
