package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

func capeColor(s CharStyle) rl.Color {
	if s.Costume == 5 {
		return rl.NewColor(22, 23, 33, 255)
	}
	return LegoRed
}
func decorateCharacter(b *MeshBuilder, s CharStyle) {
	switch s.Costume {
	case 1:
		b.Ellipsoid(v3(0, 2.65, 0), v3(.4, .22, .4), 5, 12, LegoYellow)
		b.Box(v3(0, 2.61, .34), v3(.72, .06, .2), LegoYellow)
		for _, z := range []float32{-.26, .26} {
			for _, x := range []float32{-.25, .25} {
				b.Box(v3(x, 1.53, z), v3(.08, .75, .035), LegoWhite)
			}
			b.Box(v3(0, 1.25, z), v3(.85, .08, .035), LegoWhite)
		}
		b.Box(v3(0, 1.08, 0), v3(1.05, .12, .57), LegoDarkGrey)
		for _, x := range []float32{-.38, .38} {
			b.BeveledBox(v3(x, 1.0, -.31), v3(.19, .26, .12), .03, shade(LegoOrange, .55))
			b.Cylinder(v3(x, 1.08, -.38), v3(x, 1.33, -.38), .035, .035, 6, LegoLightGrey, true)
		}
	case 2:
		// Web strands on the mask, following its curved surface.
		for _, x := range []float32{-.24, 0, .24} {
			b.Cylinder(v3(x, 2.12, .23), v3(x*.8, 2.55, .26), .012, .012, 5, LegoDarkGrey, false)
		}
		for _, x := range []float32{-.14, .14} {
			b.Ellipsoid(v3(x, 2.4, .315), v3(.12, .13, .035), 4, 8, LegoWhite)
		}
		for _, z := range []float32{-.26, .26} {
			for _, y := range []float32{1.27, 1.39, 1.51, 1.63, 1.75, 1.87} {
				b.Cylinder(v3(-.39, y, z), v3(.39, y, z), .009, .009, 5, LegoDarkGrey, false)
			}
			for _, x := range []float32{-.3, -.15, 0, .15, .3} {
				b.Cylinder(v3(x, 1.12, z), v3(x*.8, 1.94, z), .009, .009, 5, LegoDarkGrey, false)
			}
			b.Sphere(v3(0, 1.55, z), .09, LegoDarkGrey)
			for _, x := range []float32{-.18, .18} {
				for k := 0; k < 3; k++ {
					b.Cylinder(v3(0, 1.55, z), v3(x, 1.4+float32(k)*.15, z), .025, .018, 5, LegoDarkGrey, false)
				}
			}
		}
	case 3:
		b.Ellipsoid(v3(0, 2.55, -.10), v3(.38, .31, .32), 6, 14, LegoGreen)
		for k := 0; k < 5; k++ {
			y := 1.2 + float32(k)*.22
			b.Cylinder(v3(0, y, -.27), v3(0, y+.13, -.53), .13, 0, 5, LegoLime, true)
		}
	case 4:
		// Ten-point star on the chest.
		for k := 0; k < 10; k++ {
			a := float32(k) * math.Pi / 5
			next := float32(k+1) * math.Pi / 5
			r1, r2 := float32(.25), float32(.11)
			if k%2 != 0 {
				r1, r2 = r2, r1
			}
			b.Tri(v3(0, 1.58, .262), v3(r1*sinf(a), 1.58+r1*cosf(a), .262), v3(r2*sinf(next), 1.58+r2*cosf(next), .262), v3(0, 0, 1), LegoYellow)
		}
	case 0:
		b.BeveledBox(v3(0, 1.4, .27), v3(.65, .12, .04), .03, LegoGoldBright)
		b.Sphere(v3(0, 1.6, .29), .12, LegoGoldBright)
	case 5:
		b.Ellipsoid(v3(0, 2.58, -.08), v3(.37, .25, .32), 6, 14, LegoDarkGrey)
		// A friendly mask leaves the smile visible.
		b.Box(v3(0, 2.4, .32), v3(.62, .18, .07), LegoDarkGrey)
		for _, x := range []float32{-.14, .14} {
			b.Ellipsoid(v3(x, 2.4, .365), v3(.09, .045, .025), 4, 8, LegoWhite)
		}
		b.Ellipsoid(v3(0, 1.57, .27), v3(.31, .14, .03), 5, 12, LegoYellow)
		b.Ellipsoid(v3(0, 1.57, .305), v3(.19, .075, .02), 5, 12, LegoDarkGrey)
		for _, x := range []float32{-.16, .16} {
			b.Cylinder(v3(x, 1.56, .305), v3(x*1.55, 1.64, .305), .06, .015, 5, LegoDarkGrey, false)
		}
		for _, x := range []float32{-.28, .28} {
			b.Cylinder(v3(x, 2.65, 0), v3(x*1.3, 3.1, 0), .14, 0, 6, LegoDarkGrey, true)
		}
		b.Box(v3(0, 1.12, 0), v3(1.02, .14, .55), LegoYellow)
	case 6:
		for _, x := range []float32{-.14, .14} {
			b.Ellipsoid(v3(x, 2.4, .32), v3(.13, .08, .05), 4, 8, LegoDarkGrey)
		}
		for _, z := range []float32{-.26, .26} {
			for k := 0; k < 16; k++ {
				x := -.36 + float32(k)*.045
				y := float32(1.53) + .07*sinf(float32(k)*math.Pi/4)
				y2 := float32(1.53) + .07*sinf(float32(k+1)*math.Pi/4)
				b.Cylinder(v3(x, y, z), v3(x+.045, y2, z), .022, .022, 6, LegoWhite, false)
			}
		}
		b.Box(v3(0, 2.4, .36), v3(.12, .035, .04), LegoDarkGrey)
	case 7:
		b.Cylinder(v3(0, 2.57, 0), v3(0, 2.68, 0), .38, .38, 12, LegoWhite, true)
		b.Box(v3(0, 1.22, .27), v3(.86, .10, .04), LegoGreen)
		b.Box(v3(0, 1.42, .27), v3(.08, .5, .04), LegoWhite)
		b.Cylinder(v3(0, 2.62, 0), v3(.24, 3.17, 0), .36, 0, 10, LegoGreen, true)
		b.Sphere(v3(.24, 3.17, 0), .09, LegoGold)
	}
}
func buildExtraGuns() {
	for k := 1; k < NumGuns; k++ {
		b := NewMB()
		b.BeveledBox(v3(0, -.12, 0), v3(.18, .34, .20), .04, LegoDarkGrey)
		switch k {
		case 1:
			b.BeveledBox(v3(0, .12, .42), v3(.18, .19, .85), .04, LegoRed)
			b.Cylinder(v3(0, .12, .70), v3(0, .12, 1.30), .055, .045, 12, LegoLightGrey, true)
			b.Emis = 1
			b.Cylinder(v3(0, .12, 1.30), v3(0, .12, 1.37), .075, .075, 12, LegoRed, true)
			GunMuzzle[k] = v3(0, .12, 1.37)
		case 2:
			b.BeveledBox(v3(0, .12, .43), v3(.42, .38, .78), .08, LegoOrange)
			b.Cylinder(v3(0, .12, .65), v3(0, .12, 1.02), .20, .24, 16, LegoGold, true)
			b.Emis = 1
			b.Cylinder(v3(0, .12, 1.02), v3(0, .12, 1.08), .19, .19, 16, LegoYellow, true)
			GunMuzzle[k] = v3(0, .12, 1.08)
		case 3:
			b.Cylinder(v3(0, .12, -.18), v3(0, .12, 1.12), .23, .23, 16, LegoGreen, true)
			b.Cylinder(v3(0, .12, 1.12), v3(0, .12, 1.39), .18, 0, 16, LegoRed, true)
			b.Box(v3(0, .4, .34), v3(.12, .14, .45), LegoDarkGrey)
			b.Cylinder(v3(0, .48, .16), v3(0, .48, .52), .09, .09, 12, LegoDarkGrey, true)
			GunMuzzle[k] = v3(0, .12, 1.39)
		case 4:
			b.Cylinder(v3(0, .12, 0), v3(0, .12, 1.05), .11, .11, 16, LegoPurple, true)
			for j := 0; j < 4; j++ {
				z := .22 + float32(j)*.20
				b.Cylinder(v3(0, .12, z), v3(0, .12, z+.07), .18, .18, 16, LegoLightGrey, true)
			}
			b.Cylinder(v3(0, .12, 1.02), v3(0, .12, 1.17), .12, .29, 16, LegoPurple, true)
			b.Emis = 1
			b.Cylinder(v3(0, .12, 1.17), v3(0, .12, 1.20), .27, .27, 16, LegoPink, true)
			GunMuzzle[k] = v3(0, .12, 1.20)
		case 5:
			b.BeveledBox(v3(0, .12, .38), v3(.32, .30, .72), .07, LegoCyan)
			b.Ellipsoid(v3(0, .43, .34), v3(.25, .25, .29), 8, 16, rl.NewColor(145, 225, 240, 255))
			for j, c := range []rl.Color{LegoPink, LegoYellow, LegoCyan} {
				b.Sphere(v3(-.12+float32(j)*.12, .46, .59), .065, c)
			}
			b.Cylinder(v3(0, .12, .70), v3(0, .12, 1.02), .13, .30, 16, LegoOrange, true)
			b.Emis = 1
			b.Cylinder(v3(0, .12, 1.02), v3(0, .12, 1.06), .29, .29, 16, LegoCyan, true)
			GunMuzzle[k] = v3(0, .12, 1.06)
		}
		b.Emis = 0
		gunMeshes[k] = b.Build()
	}
}
