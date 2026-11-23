package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() {
	registerRealmChar(RealmCharCode(1, RoleHero), monstriHero)
	registerRealmChar(RealmCharCode(1, RoleBoss), monstriBoss)
	registerRealmVehicle("monstri", monstriCoach)
}

func monstriColor(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }

func monstriParts(skin, cloth, boots rl.Color) *CharMeshes {
	c := &CharMeshes{}
	b := NewMB()
	b.Ellipsoid(v3(0, -.23, 0), v3(.19, .27, .21), 8, 12, cloth)
	c.Leg = b.Build()
	b = NewMB()
	b.Ellipsoid(v3(0, -.22, 0), v3(.18, .25, .19), 8, 12, cloth)
	b.Ellipsoid(v3(0, -.42, .09), v3(.2, .1, .28), 8, 12, boots)
	c.Shin = b.Build()
	for i := 0; i < 2; i++ {
		b = NewMB()
		b.Ellipsoid(v3(0, -.18, 0), v3(.17, .23, .18), 8, 12, cloth)
		upper := b.Build()
		b = NewMB()
		b.Ellipsoid(v3(0, -.17, 0), v3(.15, .21, .16), 8, 12, cloth)
		b.Ellipsoid(v3(0, -.38, .04), v3(.14, .13, .14), 8, 12, skin)
		fore := b.Build()
		if i == 0 {
			c.ArmL = upper
			c.ForeL = fore
		} else {
			c.ArmR = upper
			c.ForeR = fore
		}
	}
	return c
}
func monstriFace(b *MeshBuilder, skin rl.Color, witch bool) {
	dark := monstriColor(38, 25, 49)
	white := monstriColor(255, 250, 240)
	b.Ellipsoid(v3(0, 2.38, 0), v3(.39, .43, .34), 12, 18, skin)
	for _, x := range []float32{-.16, .16} {
		b.Ellipsoid(v3(x, 2.43, .306), v3(.105, .12, .045), 8, 12, white)
		b.Ellipsoid(v3(x, 2.43, .348), v3(.045, .065, .025), 8, 10, dark)
		b.Sphere(v3(x-.015, 2.46, .369), .015, white)
	}
	b.Ellipsoid(v3(0, 2.23, .316), v3(.16, .045, .025), 8, 12, dark)
	if witch {
		b.Ellipsoid(v3(0, 2.32, .37), v3(.075, .08, .18), 8, 12, skin)
	} else {
		for _, x := range []float32{-.09, .09} {
			b.Cylinder(v3(x, 2.25, .344), v3(x, 2.19, .35), .035, 0, 6, white, true)
		}
	}
}
func monstriHero(s CharStyle) *CharMeshes {
	skin := monstriColor(208, 184, 221)
	black := monstriColor(30, 25, 43)
	purple := monstriColor(121, 62, 153)
	red := monstriColor(217, 45, 66)
	c := monstriParts(skin, black, black)
	b := NewMB()
	b.Ellipsoid(v3(0, 1.48, 0), v3(.48, .53, .27), 10, 16, black)
	b.Ellipsoid(v3(0, 1.48, .23), v3(.29, .41, .07), 10, 14, purple)
	for _, y := range []float32{1.35, 1.5, 1.65} {
		b.Sphere(v3(0, y, .3), .027, monstriColor(247, 205, 99))
	}
	b.Ellipsoid(v3(-.11, 1.84, .27), v3(.13, .08, .05), 8, 10, red)
	b.Ellipsoid(v3(.11, 1.84, .27), v3(.13, .08, .05), 8, 10, red)
	b.Sphere(v3(0, 1.84, .31), .055, red)
	b.Cylinder(v3(0, 1.9, 0), v3(0, 2.08, 0), .16, .16, 10, skin, true)
	monstriFace(b, skin, false)
	b.Ellipsoid(v3(0, 2.69, -.04), v3(.4, .18, .35), 10, 16, black)
	b.Tri(v3(-.2, 2.66, .29), v3(.2, 2.66, .29), v3(0, 2.48, .35), v3(0, 0, 1), black)
	for _, x := range []float32{-.34, .34} {
		b.Tri(v3(x, 1.92, -.15), v3(x*1.4, 2.27, -.16), v3(0, 2.0, -.27), v3(0, 0, 1), red)
	}
	if s.Crown {
		AddCrown(b, 2.86)
	}
	c.Body = b.Build()
	b = NewMB()
	b.Quad(v3(-.35, 0, 0), v3(.35, 0, 0), v3(.57, -1.25, -.12), v3(-.57, -1.25, -.12), v3(0, 0, -1), black)
	b.Quad(v3(-.35, 0, .01), v3(.35, 0, .01), v3(.57, -1.25, -.11), v3(-.57, -1.25, -.11), v3(0, 0, 1), red)
	c.Cape = b.Build()
	c.HasCape = true
	return c
}
func monstriBoss(s CharStyle) *CharMeshes {
	skin := monstriColor(206, 190, 153)
	purple := monstriColor(105, 49, 143)
	green := monstriColor(99, 180, 70)
	dark := monstriColor(38, 25, 49)
	gold := monstriColor(255, 214, 93)
	c := monstriParts(skin, purple, dark)
	b := NewMB()
	b.Ellipsoid(v3(0, 1.45, 0), v3(.48, .58, .28), 10, 16, purple)
	for _, p := range []rl.Vector3{v3(-.23, 1.35, .25), v3(.17, 1.65, .25), v3(.08, 1.18, .25)} {
		b.Sphere(p, .055, gold)
		b.Box(p, v3(.14, .025, .025), gold)
		b.Box(p, v3(.025, .14, .025), gold)
	}
	b.Cylinder(v3(0, 1.9, 0), v3(0, 2.08, 0), .16, .16, 10, skin, true)
	for _, x := range []float32{-.33, .33} {
		for j := 0; j < 3; j++ {
			b.Ellipsoid(v3(x, 2.3+float32(j)*.15, -.08), v3(.13, .2, .18), 7, 10, green)
		}
	}
	monstriFace(b, skin, true)
	b.Cylinder(v3(0, 2.73, 0), v3(0, 2.8, 0), .54, .54, 16, dark, true)
	b.Cylinder(v3(0, 2.8, 0), v3(.14, 3.02, 0), .34, .18, 16, purple, true)
	b.Cylinder(v3(.14, 3.02, 0), v3(.38, 3.12, 0), .18, 0, 12, purple, true)
	b.Cylinder(v3(0, 2.8, 0), v3(0, 2.87, 0), .34, .32, 16, gold, true)
	if s.Crown {
		AddCrown(b, 3.12)
	}
	c.Body = b.Build()
	return c
}
func monstriCoach(col rl.Color) rl.Mesh {
	b := NewMB()
	dark := monstriColor(34, 24, 43)
	purple := monstriColor(76, 38, 102)
	gold := monstriColor(211, 166, 83)
	green := monstriColor(126, 255, 129)
	b.BeveledBox(v3(0, 1.18, 0), v3(1.8, 1.2, 2.6), .16, purple)
	b.BeveledBox(v3(0, .57, 0), v3(2, .22, 3.1), .08, dark)
	for _, x := range []float32{-.91, .91} {
		b.Box(v3(x, 1.2, 0), v3(.04, .76, 1.7), col)
		for _, z := range []float32{-.8, .8} {
			b.Box(v3(x, 1.2, z), v3(.06, .8, .06), gold)
		}
	}
	b.BeveledBox(v3(0, 1.9, 0), v3(2, .25, 2.9), .12, dark)
	b.Ellipsoid(v3(0, 2.03, 0), v3(.85, .25, 1.28), 8, 14, purple)
	b.Sphere(v3(0, 2.29, 0), .13, gold)
	for _, x := range []float32{-1.04, 1.04} {
		for _, z := range []float32{-1.15, 1.15} {
			for j := 0; j < 20; j++ {
				a := float32(j) * 3.14159 / 10
				d := float32(j+1) * 3.14159 / 10
				b.Cylinder(v3(x, .59+.53*cosf(a), z+.53*sinf(a)), v3(x, .59+.53*cosf(d), z+.53*sinf(d)), .07, .07, 6, dark, true)
			}
			b.Cylinder(v3(x-.1, .59, z), v3(x+.1, .59, z), .13, .13, 12, gold, true)
			for k := 0; k < 8; k++ {
				a := float32(k) * 3.14159 / 4
				b.Cylinder(v3(x, .59, z), v3(x, .59+.5*cosf(a), z+.5*sinf(a)), .027, .027, 6, gold, true)
			}
		}
	}
	for _, x := range []float32{-.68, .68} {
		b.Cylinder(v3(x, 1.25, 1.35), v3(x, 1.57, 1.35), .12, .12, 8, gold, true)
		b.Emis = 1
		b.Ellipsoid(v3(x, 1.4, 1.43), v3(.13, .18, .13), 8, 12, green)
		b.Emis = 0
	}
	return b.Build()
}
