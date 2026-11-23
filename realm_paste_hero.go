package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() {
	registerRealmChar(RealmCharCode(5, RoleHero), func(s CharStyle) *CharMeshes { return pasteCharacter(s, false) })
	registerRealmChar(RealmCharCode(5, RoleBoss), func(s CharStyle) *CharMeshes { return pasteCharacter(s, true) })
	registerRealmVehicle("paste", pasteCar)
}
func pasteCharacter(s CharStyle, fox bool) *CharMeshes {
	cream := rl.NewColor(245, 226, 190, 255)
	pink := rl.NewColor(239, 146, 160, 255)
	blue := rl.NewColor(63, 137, 210, 255)
	dark := rl.NewColor(55, 43, 48, 255)
	white := rl.NewColor(255, 248, 230, 255)
	orange := rl.NewColor(225, 123, 56, 255)
	green := rl.NewColor(54, 138, 92, 255)
	gold := rl.NewColor(244, 194, 65, 255)
	fur := cream
	clothing := blue
	if fox {
		fur = orange
		clothing = green
	}
	b := NewMB()
	b.Ellipsoid(v3(0, 1.42, 0), v3(.49, .55, .31), 10, 16, fur)
	b.Ellipsoid(v3(0, 1.32, .06), v3(.48, .43, .31), 10, 16, clothing)
	if !fox {
		b.BeveledBox(v3(0, 1.52, .31), v3(.52, .42, .08), .05, blue)
		for _, x := range []float32{-.28, .28} {
			b.Cylinder(v3(x, 1.42, .3), v3(x*.82, 1.82, .23), .075, .075, 8, blue, true)
			b.Sphere(v3(x*.82, 1.72, .31), .045, gold)
		}
		b.BeveledBox(v3(0, 1.43, .37), v3(.34, .20, .06), .03, shade(blue, .8))
		b.Cylinder(v3(.04, 1.49, .41), v3(.04, 1.75, .41), .02, .075, 8, orange, true)
		b.Cylinder(v3(.04, 1.74, .41), v3(-.03, 1.84, .41), .025, .01, 6, green, true)
	} else {
		for _, x := range []float32{-.12, .12} {
			b.Cylinder(v3(x, 1.13, .34), v3(x, 1.75, .26), .018, .018, 6, shade(green, .65), true)
		}
		for _, y := range []float32{1.3, 1.48, 1.65} {
			b.Sphere(v3(0, y, .36), .035, gold)
		}
	}
	b.Ellipsoid(v3(0, 2.23, 0), v3(.52, .49, .39), 12, 18, fur)
	if fox {
		for _, x := range []float32{-.34, .34} {
			b.Cylinder(v3(x, 2.48, 0), v3(x*1.4, 2.94, 0), .19, .015, 10, fur, true)
			b.Cylinder(v3(x, 2.51, .08), v3(x*1.35, 2.83, .08), .09, .005, 8, pink, true)
		}
		b.Ellipsoid(v3(0, 2.08, .36), v3(.38, .20, .22), 10, 14, white)
	} else {
		b.Ellipsoid(v3(-.27, 2.88, 0), v3(.14, .50, .11), 10, 12, fur)
		b.Ellipsoid(v3(-.27, 2.89, .105), v3(.075, .38, .025), 8, 10, pink)
		b.Cylinder(v3(.25, 2.55, 0), v3(.31, 3.05, 0), .13, .11, 12, fur, true)
		b.Ellipsoid(v3(.43, 3.05, 0), v3(.23, .11, .11), 8, 12, fur)
		b.Ellipsoid(v3(.31, 2.87, .11), v3(.065, .24, .025), 8, 10, pink)
		b.Ellipsoid(v3(0, 2.06, .36), v3(.32, .18, .14), 10, 14, white)
	}
	for _, x := range []float32{-.19, .19} {
		b.Ellipsoid(v3(x, 2.30, .35), v3(.105, .14, .065), 8, 12, white)
		b.Sphere(v3(x, 2.30, .41), .06, dark)
		b.Sphere(v3(x-.02, 2.34, .46), .018, white)
	}
	b.Ellipsoid(v3(0, 2.12, .52), v3(.09, .06, .045), 8, 10, pink)
	for _, x := range []float32{-.09, .09} {
		b.Ellipsoid(v3(x, 2.02, .47), v3(.11, .025, .025), 6, 10, dark)
	}
	if fox {
		b.Ellipsoid(v3(0, 2.13, .59), v3(.09, .06, .05), 8, 10, dark)
		for _, x := range []float32{-.15, .15} {
			b.Ellipsoid(v3(x, 2.02, .56), v3(.13, .025, .025), 6, 10, dark)
		}
	}
	top := float32(3.15)
	if fox {
		b.Cylinder(v3(0, 2.68, 0), v3(0, 2.75, 0), .53, .53, 16, dark, true)
		b.Cylinder(v3(0, 2.74, 0), v3(0, 3.18, 0), .34, .31, 16, dark, true)
		b.Cylinder(v3(0, 2.78, 0), v3(0, 2.86, 0), .35, .35, 16, green, true)
		top = 3.18
		for i := 0; i < 16; i++ {
			a := float32(i) * 6.28318 / 16
			c := float32(i+1) * 6.28318 / 16
			b.Cylinder(v3(.19+.16*cosf(a), 2.30+.16*sinf(a), .43), v3(.19+.16*cosf(c), 2.30+.16*sinf(c), .43), .014, .014, 5, gold, true)
		}
	}
	if s.Crown {
		AddCrown(b, top)
	}
	cm := &CharMeshes{Body: b.Build(), HeadTop: top, HasTail: true}
	part := func(c, r rl.Vector3, col rl.Color) rl.Mesh {
		m := NewMB()
		m.Ellipsoid(c, r, 8, 12, col)
		return m.Build()
	}
	cm.Leg = part(v3(0, -.22, 0), v3(.2, .27, .21), clothing)
	cm.Shin = part(v3(0, -.25, .06), v3(.19, .26, .27), fur)
	cm.ArmL = part(v3(0, -.19, 0), v3(.17, .24, .17), fur)
	cm.ArmR = cm.ArmL
	cm.ForeL = part(v3(0, -.21, .02), v3(.16, .24, .17), fur)
	cm.ForeR = cm.ForeL
	t := NewMB()
	if fox {
		t.Ellipsoid(v3(0, .08, -.51), v3(.33, .32, .64), 10, 16, fur)
		t.Ellipsoid(v3(0, .10, -.98), v3(.25, .24, .22), 8, 12, white)
	} else {
		t.Sphere(v3(0, 0, -.25), .25, white)
		for _, x := range []float32{-.14, .14} {
			t.Sphere(v3(x, .04, -.28), .13, cream)
		}
	}
	cm.Tail = t.Build()
	return cm
}
func pasteCar(col rl.Color) rl.Mesh {
	b := NewMB()
	pink := rl.NewColor(246, 145, 190, 255)
	yellow := rl.NewColor(255, 218, 91, 255)
	white := rl.NewColor(255, 248, 226, 255)
	purple := rl.NewColor(148, 111, 200, 255)
	b.EllipsoidF(v3(0, 1.05, 0), v3(1.05, .78, 2.1), 14, 24, func(n rl.Vector3) rl.Color {
		if n.Y > .7 {
			return white
		}
		if n.Y > .25 {
			return pink
		}
		if n.Y > -.25 {
			return yellow
		}
		return col
	})
	b.Ellipsoid(v3(0, 1.75, 0), v3(.72, .08, 1.25), 8, 16, shade(purple, .7))
	b.Ellipsoid(v3(0, 1.8, -.4), v3(.48, .22, .28), 8, 12, purple)
	for _, x := range []float32{-1.02, 1.02} {
		for _, z := range []float32{-1.25, 1.25} {
			b.Cylinder(v3(x*.93, .47, z), v3(x*1.12, .47, z), .43, .43, 16, purple, true)
			for i := 0; i < 8; i++ {
				a := float32(i) * 6.28318 / 8
				b.Ellipsoid(v3(x*1.14, .47+.27*cosf(a), z+.27*sinf(a)), v3(.09, .14, .14), 6, 8, white)
			}
			b.Sphere(v3(x*1.18, .47, z), .15, yellow)
		}
		for _, z := range []float32{-.7, 0, .7} {
			b.Sphere(v3(x, 1.12, z), .12, yellow)
		}
	}
	return b.Build()
}
