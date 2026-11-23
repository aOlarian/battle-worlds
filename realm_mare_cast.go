package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

func mareCastColor(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
func mareEll(b *MeshBuilder, x, y, z, rx, ry, rz float32, c rl.Color) {
	b.Ellipsoid(v3(x, y, z), v3(rx, ry, rz), 8, 12, c)
}
func mareCast(s CharStyle, role int) *CharMeshes {
	red := mareCastColor(224, 60, 51)
	blue := mareCastColor(70, 160, 210)
	purple := mareCastColor(164, 91, 196)
	green := mareCastColor(73, 183, 94)
	white := mareCastColor(250, 247, 230)
	dark := mareCastColor(42, 45, 55)
	gold := mareCastColor(250, 194, 56)
	skin := mareCastColor(224, 165, 127)
	main := red
	switch role {
	case 2:
		main = blue
	case 3:
		main = purple
	case 4:
		main = green
	case 6, 7, 8:
		main = skin
	}
	b := NewMB()
	headTop := float32(2.75)
	switch role {
	case 1:
		mareEll(b, 0, 1.63, 0, .69, .51, .39, red)
		mareEll(b, 0, 2.22, 0, .47, .40, .34, red)
		b.Cylinder(v3(0, 2.46, 0), v3(0, 2.62, 0), .43, .39, 16, dark, true)
		mareEll(b, .39, 2.47, -.1, .14, .12, .12, dark)
	case 2:
		mareEll(b, 0, 1.60, 0, .48, .66, .35, blue)
		mareEll(b, 0, 2.24, 0, .43, .48, .35, blue)
		mareEll(b, 0, 1.64, .32, .32, .40, .08, white)
		mareEll(b, 0, 1.03, 0, .48, .24, .35, white)
		for _, x := range []float32{-.34, -.17, 0, .17, .34} {
			b.Box(v3(x, 1.06, .335), v3(.065, .31, .035), red)
		}
		b.Cylinder(v3(0, 1.9, -.25), v3(0, 2.36, -.61), .22, .015, 8, blue, true)
		mareEll(b, 0, 2.15, .35, .31, .15, .13, white)
		for _, x := range []float32{-.18, -.09, 0, .09, .18} {
			mareEll(b, x, 2.13, .465, .035, .045, .025, white)
		}
	case 3:
		mareEll(b, 0, 1.55, 0, .48, .55, .35, purple)
		mareEll(b, 0, 2.22, 0, .48, .51, .37, purple)
		b.Cylinder(v3(0, 2.61, 0), v3(0, 2.76, 0), .35, .32, 16, white, true)
		b.Cylinder(v3(0, 2.59, 0), v3(0, 2.63, 0), .4, .4, 16, blue, true)
		headTop = 2.76
		for _, side := range []float32{-1, 1} {
			for k := 0; k < 2; k++ {
				y := float32(1.2) + float32(k)*.23
				b.Cylinder(v3(side*.38, y, -.1), v3(side*(.75+float32(k)*.08), y-.27, -.12), .1, .035, 8, purple, true)
			}
		}
	case 4:
		mareEll(b, 0, 1.56, 0, .49, .57, .35, green)
		mareEll(b, 0, 2.25, 0, .43, .43, .35, green)
		mareEll(b, 0, 1.54, .32, .32, .40, .07, red)
		mareEll(b, 0, 2.20, .39, .20, .15, .22, gold)
		b.Cylinder(v3(0, 2.57, 0), v3(0, 2.68, 0), .34, .27, 12, dark, true)
		b.Cylinder(v3(0, 2.55, 0), v3(0, 2.59, 0), .43, .43, 12, dark, true)
		for _, x := range []float32{-.2, 0, .2} {
			b.Cylinder(v3(x, 2.57, -.1), v3(x, 2.83, -.14), .10, .015, 8, red, true)
		}
		headTop = 2.83
	default:
		shirt := mareCastColor(65, 190, 193)
		if role == 7 {
			shirt = red
		}
		if role == 8 {
			shirt = mareCastColor(100, 180, 225)
		}
		mareEll(b, 0, 1.52, 0, .42, .51, .27, shirt)
		mareEll(b, 0, 2.27, 0, .37, .43, .32, skin)
		if role == 6 {
			mareEll(b, 0, 1.13, 0, .56, .39, .33, shirt)
			for _, x := range []float32{-.3, 0, .3} {
				for _, y := range []float32{1.08, 1.34, 1.58} {
					mareEll(b, x, y, .31, .07, .07, .025, white)
					mareEll(b, x, y, .34, .025, .025, .02, gold)
				}
			}
			b.Cylinder(v3(0, 2.58, 0), v3(0, 2.63, 0), .62, .62, 20, gold, true)
			mareEll(b, 0, 2.65, 0, .36, .16, .32, gold)
		}
		if role == 7 {
			b.Cylinder(v3(-.18, 1.84, .28), v3(0, 1.55, .3), .025, .025, 8, white, true)
			b.Cylinder(v3(.18, 1.84, .28), v3(0, 1.55, .3), .025, .025, 8, white, true)
			mareEll(b, 0, 1.52, .33, .07, .08, .04, gold)
			for i := 0; i < 16; i++ {
				a := float64(i) * math.Pi / 8
				z := float64(i+1) * math.Pi / 8
				c := white
				if i%4 == 0 {
					c = red
				}
				b.Cylinder(v3(.45+float32(math.Cos(a))*.24, 1.15+float32(math.Sin(a))*.24, .25), v3(.45+float32(math.Cos(z))*.24, 1.15+float32(math.Sin(z))*.24, .25), .07, .07, 8, c, true)
			}
		}
		if role == 8 {
			mareEll(b, 0, 1.40, .27, .34, .36, .08, white)
			for _, x := range []float32{-.23, -.11, .01, .13, .25} {
				b.Box(v3(x, 1.4, .35), v3(.045, .58, .025), red)
			}
			b.Cylinder(v3(0, 2.58, 0), v3(0, 2.72, 0), .33, .32, 16, white, true)
			mareEll(b, 0, 2.72, 0, .32, .15, .29, white)
			headTop = 2.87
		}
	}
	// Eyes and a small friendly smile, placed on the forward face.
	for _, x := range []float32{-.16, .16} {
		mareEll(b, x, 2.32, .325, .105, .12, .065, white)
		mareEll(b, x, 2.32, .383, .045, .057, .03, dark)
		mareEll(b, x-.012, 2.35, .409, .014, .017, .01, white)
	}
	if role != 2 && role != 4 {
		mareEll(b, 0, 2.12, .34, .13, .035, .025, dark)
	}
	if s.Crown {
		AddCrown(b, headTop)
	}
	cm := &CharMeshes{Body: b.Build(), HeadTop: headTop}
	limb := func(c rl.Color, lower bool, leg bool, side float32) rl.Mesh {
		m := NewMB()
		if leg {
			mareEll(m, 0, -.22, 0, .17, .29, .17, c)
			if lower {
				mareEll(m, 0, -.40, .10, .19, .10, .26, c)
			}
		} else {
			mareEll(m, 0, -.19, 0, .15, .26, .15, c)
			if lower {
				if role == 1 {
					mareEll(m, side*.08, -.30, .04, .30, .22, .20, red)
					mareEll(m, side*.23, -.16, .04, .12, .16, .12, red)
					mareEll(m, side*-.08, -.16, .04, .12, .16, .12, red)
				} else {
					mareEll(m, 0, -.38, .03, .14, .14, .14, c)
				}
			}
		}
		return m.Build()
	}
	legcol := main
	if role >= 6 {
		legcol = skin
	}
	if role == 4 {
		legcol = gold
	}
	cm.Leg = limb(legcol, false, true, 1)
	cm.Shin = limb(legcol, true, true, 1)
	cm.ArmL = limb(main, false, false, -1)
	cm.ArmR = limb(main, false, false, 1)
	cm.ForeL = limb(main, true, false, -1)
	cm.ForeR = limb(main, true, false, 1)
	return cm
}
func init() {
	for _, role := range []int{RoleEnemy1, RoleEnemy2, RoleEnemy3, RoleEnemy4, RoleNPC1, RoleNPC2, RoleNPC3} {
		r := role
		registerRealmChar(RealmCharCode(0, r), func(s CharStyle) *CharMeshes { return mareCast(s, r) })
	}
}
