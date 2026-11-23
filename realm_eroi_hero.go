package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

func init() {
	registerRealmChar(RealmCharCode(3, RoleHero), eroiHero)
	registerRealmChar(RealmCharCode(3, RoleBoss), eroiBoss)
	registerRealmVehicle("eroi", eroiCar)
}

func eroiColor(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }

func eroiStar(b *MeshBuilder, x, y, z, radius float32, col rl.Color) {
	for i := 0; i < 10; i++ {
		a := math.Pi/2 + float64(i)*math.Pi/5
		n := math.Pi/2 + float64(i+1)*math.Pi/5
		r1, r2 := radius, radius*0.43
		if i%2 == 1 {
			r1, r2 = r2, r1
		}
		b.Tri(v3(x, y, z), v3(x+r1*float32(math.Cos(a)), y+r1*float32(math.Sin(a)), z), v3(x+r2*float32(math.Cos(n)), y+r2*float32(math.Sin(n)), z), v3(0, 0, 1), col)
	}
}

func eroiPerson(s CharStyle, boss bool) *CharMeshes {
	white := eroiColor(242, 247, 243)
	teal := eroiColor(12, 171, 163)
	gold := eroiColor(255, 198, 49)
	skin := eroiColor(224, 169, 128)
	dark := eroiColor(28, 29, 37)
	suit, trim := white, teal
	if boss {
		suit = eroiColor(112, 105, 135)
		trim = eroiColor(71, 64, 92)
	}
	b := NewMB()
	b.Ellipsoid(v3(0, 1.46, 0), v3(.48, .53, .27), 10, 16, suit)
	b.Ellipsoid(v3(0, 1.05, 0), v3(.40, .18, .25), 8, 14, trim)
	b.Cylinder(v3(0, 1.06, 0), v3(0, 1.19, 0), .37, .38, 16, gold, true)
	b.BeveledBox(v3(0, 1.125, .37), v3(.19, .14, .08), .03, gold)
	if boss {
		b.BeveledBox(v3(0, 1.52, .25), v3(.52, .40, .12), .07, eroiColor(155, 148, 169))
		b.Emis = .8
		b.Ellipsoid(v3(0, 1.55, .33), v3(.10, .10, .035), 8, 12, gold)
		b.Emis = 0
		b.Ellipsoid(v3(0, 2.32, 0), v3(.43, .48, .37), 10, 16, trim)
		b.BeveledBox(v3(0, 2.35, .31), v3(.68, .23, .16), .07, eroiColor(37, 42, 62))
		b.Emis = .8
		for _, x := range []float32{-.16, .16} {
			b.Ellipsoid(v3(x, 2.36, .40), v3(.10, .055, .025), 6, 10, gold)
		}
		b.Emis = 0
		for _, x := range []float32{-.49, .49} {
			b.Ellipsoid(v3(x, 1.91, 0), v3(.26, .20, .3), 8, 12, suit)
			b.Emis = 1
			sign := float32(1)
			if x < 0 {
				sign = -1
			}
			b.Quad(v3(x-.09, 2.0, .26), v3(x+.08, 2.0, .26), v3(x+sign*.04, 2.27, .26), v3(x-sign*.09, 2.27, .26), v3(0, 0, 1), gold)
			b.Tri(v3(x-sign*.09, 2.16, .28), v3(x+sign*.16, 2.16, .28), v3(x-sign*.10, 1.94, .28), v3(0, 0, 1), gold)
			b.Emis = 0
		}
	} else {
		b.Cylinder(v3(0, 1.83, 0), v3(0, 2.03, 0), .16, .16, 12, skin, true)
		b.Ellipsoid(v3(0, 2.30, 0), v3(.39, .43, .34), 12, 16, skin)
		b.Ellipsoid(v3(0, 2.58, -.04), v3(.40, .20, .33), 10, 16, dark)
		b.Ellipsoid(v3(.12, 2.66, .18), v3(.19, .17, .17), 8, 12, dark)
		for _, x := range []float32{-.15, .15} {
			b.Ellipsoid(v3(x, 2.34, .305), v3(.10, .12, .055), 8, 12, rl.White)
			b.Ellipsoid(v3(x, 2.34, .355), v3(.045, .065, .025), 8, 10, dark)
		}
		b.Ellipsoid(v3(0, 2.17, .32), v3(.13, .025, .025), 6, 10, eroiColor(139, 74, 66))
		eroiStar(b, 0, 1.57, .274, .25, gold)
	}
	if s.Crown {
		AddCrown(b, 2.75)
	}
	cm := &CharMeshes{Body: b.Build(), HeadTop: 2.75}
	leg := NewMB()
	leg.Ellipsoid(v3(0, -.22, 0), v3(.19, .25, .20), 8, 12, suit)
	cm.Leg = leg.Build()
	shin := NewMB()
	shin.Ellipsoid(v3(0, -.20, 0), v3(.18, .25, .19), 8, 12, trim)
	shin.Ellipsoid(v3(0, -.40, .09), v3(.19, .10, .28), 8, 12, trim)
	cm.Shin = shin.Build()
	arm := NewMB()
	arm.Ellipsoid(v3(0, -.18, 0), v3(.19, .23, .20), 8, 12, suit)
	cm.ArmL = arm.Build()
	cm.ArmR = cm.ArmL
	fore := NewMB()
	fore.Ellipsoid(v3(0, -.17, 0), v3(.15, .21, .16), 8, 12, trim)
	fore.Ellipsoid(v3(0, -.37, .02), v3(.16, .14, .16), 8, 12, trim)
	cm.ForeL = fore.Build()
	cm.ForeR = cm.ForeL
	if !boss {
		cape := NewMB()
		orange := eroiColor(245, 126, 32)
		cape.Quad(v3(-.32, 0, 0), v3(.32, 0, 0), v3(.52, -1.65, -.12), v3(-.52, -1.65, -.12), v3(0, 0, -1), orange)
		cape.Quad(v3(-.32, 0, -.01), v3(.32, 0, -.01), v3(.52, -1.65, -.13), v3(-.52, -1.65, -.13), v3(0, 0, 1), orange)
		cm.Cape = cape.Build()
		cm.HasCape = true
	}
	return cm
}
func eroiHero(s CharStyle) *CharMeshes { return eroiPerson(s, false) }
func eroiBoss(s CharStyle) *CharMeshes { return eroiPerson(s, true) }

func eroiCar(col rl.Color) rl.Mesh {
	b := NewMB()
	chrome := eroiColor(205, 216, 218)
	rubber := eroiColor(33, 35, 42)
	seat := eroiColor(120, 70, 48)
	lamp := eroiColor(255, 235, 166)
	b.BeveledBox(v3(0, .72, 0), v3(2, .65, 4.1), .23, col)
	b.BeveledBox(v3(0, 1.02, 1.32), v3(1.9, .32, 1.35), .14, col)
	b.BeveledBox(v3(0, 1.05, -1.62), v3(1.9, .28, .65), .12, col)
	for _, x := range []float32{-.51, .51} {
		b.BeveledBox(v3(x, 1.12, -.55), v3(.8, .42, .6), .12, seat)
		b.BeveledBox(v3(x, 1.38, -.91), v3(.8, .65, .18), .08, seat)
	}
	b.BeveledBox(v3(0, 1.48, .56), v3(1.72, .58, .08), .04, eroiColor(140, 213, 215))
	for _, x := range []float32{-.9, .9} {
		b.Cylinder(v3(x, 1.17, .56), v3(x, 1.82, .56), .045, .045, 8, chrome, true)
	}
	b.Cylinder(v3(-.9, 1.82, .56), v3(.9, 1.82, .56), .045, .045, 8, chrome, true)
	for _, x := range []float32{-1.01, 1.01} {
		for _, z := range []float32{-1.35, 1.35} {
			b.Cylinder(v3(x-.12, .45, z), v3(x+.12, .45, z), .43, .43, 16, rubber, true)
			outer := x
			if x < 0 {
				outer -= .13
			} else {
				outer += .13
			}
			b.Cylinder(v3(outer-.01, .45, z), v3(outer+.01, .45, z), .23, .23, 12, chrome, true)
		}
	}
	for _, x := range []float32{-.68, .68} {
		b.Cylinder(v3(x, .90, 2.03), v3(x, .90, 2.12), .22, .22, 16, chrome, true)
		b.Emis = .7
		b.Cylinder(v3(x, .90, 2.12), v3(x, .90, 2.14), .16, .16, 16, lamp, true)
		b.Emis = 0
	}
	b.BeveledBox(v3(0, .54, 2.08), v3(1.85, .13, .16), .05, chrome)
	b.BeveledBox(v3(0, .54, -2.08), v3(1.85, .13, .16), .05, chrome)
	return b.Build()
}
