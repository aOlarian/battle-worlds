package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() {
	registerRealmChar(RealmCharCode(6, RoleHero), halloweenHero)
	registerRealmChar(RealmCharCode(6, RoleBoss), halloweenBoss)
	registerRealmVehicle("halloween", halloweenCoach)
}
func hc(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
func halloweenParts(skin, cloth, shoe rl.Color) *CharMeshes {
	c := &CharMeshes{}
	b := NewMB()
	b.Ellipsoid(v3(0, -.23, 0), v3(.19, .27, .19), 8, 12, cloth)
	c.Leg = b.Build()
	b = NewMB()
	b.Ellipsoid(v3(0, -.21, 0), v3(.17, .24, .18), 8, 12, cloth)
	b.Ellipsoid(v3(0, -.42, .09), v3(.2, .1, .26), 8, 12, shoe)
	c.Shin = b.Build()
	b = NewMB()
	b.Ellipsoid(v3(0, -.19, 0), v3(.17, .23, .17), 8, 12, cloth)
	c.ArmL = b.Build()
	c.ArmR = c.ArmL
	b = NewMB()
	b.Ellipsoid(v3(0, -.18, 0), v3(.15, .21, .15), 8, 12, cloth)
	b.Sphere(v3(0, -.39, .02), .14, skin)
	c.ForeL = b.Build()
	c.ForeR = c.ForeL
	return c
}
func halloweenCape(c *CharMeshes, outer, inner rl.Color) {
	b := NewMB()
	b.Quad(v3(-.35, 0, 0), v3(.35, 0, 0), v3(.62, -1.1, -.15), v3(-.62, -1.1, -.15), v3(0, 0, -1), outer)
	b.Quad(v3(-.35, 0, .01), v3(.35, 0, .01), v3(.62, -1.1, -.14), v3(-.62, -1.1, -.14), v3(0, 0, 1), inner)
	c.Cape = b.Build()
	c.HasCape = true
}
func halloweenStar(b *MeshBuilder, x, y, z, r float32, col rl.Color) {
	for i := 0; i < 10; i++ {
		a := float32(i) * 3.14159265 / 5
		d := float32(i+1) * 3.14159265 / 5
		r1 := r
		r2 := r
		if i%2 == 1 {
			r1 *= .42
		}
		if (i+1)%2 == 1 {
			r2 *= .42
		}
		b.Tri(v3(x, y, z), v3(x+r1*sinf(a), y+r1*cosf(a), z), v3(x+r2*sinf(d), y+r2*cosf(d), z), v3(0, 0, 1), col)
	}
}
func halloweenHero(s CharStyle) *CharMeshes {
	skin := hc(242, 190, 151)
	purple := hc(98, 48, 145)
	dark := hc(48, 27, 75)
	orange := hc(244, 126, 34)
	gold := hc(255, 210, 66)
	hair := hc(105, 55, 34)
	black := hc(35, 27, 39)
	white := hc(255, 250, 234)
	c := halloweenParts(skin, purple, dark)
	b := NewMB()
	b.Ellipsoid(v3(0, 1.45, 0), v3(.47, .53, .29), 10, 16, purple)
	for _, y := range []float32{1.25, 1.43, 1.61} {
		b.Sphere(v3(0, y, .29), .035, gold)
	}
	b.Ellipsoid(v3(0, 2.28, 0), v3(.43, .46, .36), 12, 18, skin)
	b.Ellipsoid(v3(0, 2.57, -.09), v3(.44, .23, .33), 10, 16, hair)
	for _, x := range []float32{-.29, -.15, 0, .15, .29} {
		b.Sphere(v3(x, 2.56, .26), .12, hair)
	}
	for _, x := range []float32{-.17, .17} {
		b.Ellipsoid(v3(x, 2.32, .32), v3(.10, .13, .05), 8, 12, white)
		b.Ellipsoid(v3(x, 2.32, .37), v3(.045, .065, .025), 8, 10, black)
		b.Sphere(v3(x-.015, 2.36, .395), .017, white)
		b.Ellipsoid(v3(x*1.65, 2.16, .30), v3(.085, .045, .025), 8, 10, hc(228, 104, 99))
	}
	b.Ellipsoid(v3(0, 2.14, .365), v3(.13, .035, .025), 8, 12, hc(133, 59, 57))
	b.Cylinder(v3(0, 1.92, 0), v3(0, 2.05, 0), .29, .29, 16, orange, true)
	for i := 0; i < 3; i++ {
		y := float32(1.87) - float32(i)*.10
		color := orange
		if i == 1 {
			color = black
		}
		b.BeveledBox(v3(.18, y, .3), v3(.18, .105, .08), .02, color)
	}
	b.Cylinder(v3(0, 2.68, 0), v3(0, 2.75, 0), .57, .57, 20, dark, true)
	b.Cylinder(v3(0, 2.75, 0), v3(.08, 3.15, 0), .39, .18, 16, purple, true)
	b.Cylinder(v3(.08, 3.15, 0), v3(.42, 3.25, 0), .18, 0, 12, purple, true)
	for _, p := range []rl.Vector3{v3(-.19, 2.89, .32), v3(.12, 3.04, .24), v3(.17, 2.82, .34)} {
		halloweenStar(b, p.X, p.Y, p.Z, .09, gold)
	}
	if s.Crown {
		AddCrown(b, 3.25)
	}
	c.Body = b.Build()
	c.HeadTop = 3.25
	b = NewMB()
	b.Ellipsoid(v3(0, -.18, 0), v3(.15, .21, .15), 8, 12, purple)
	b.Sphere(v3(0, -.39, .02), .14, skin)
	b.Cylinder(v3(0, -.40, .04), v3(.12, .03, .22), .035, .025, 8, hair, true)
	b.Emis = 1
	halloweenStar(b, .12, .10, .23, .16, gold)
	b.Emis = 0
	c.ForeR = b.Build()
	halloweenCape(c, dark, orange)
	return c
}
func halloweenBoss(s CharStyle) *CharMeshes {
	green := hc(57, 118, 54)
	leaf := hc(97, 160, 54)
	orange := hc(230, 112, 24)
	glow := hc(255, 180, 42)
	dark := hc(113, 57, 19)
	c := halloweenParts(green, green, dark)
	b := NewMB()
	b.Ellipsoid(v3(0, 1.45, 0), v3(.48, .53, .30), 10, 16, green)
	for _, x := range []float32{-.27, 0, .27} {
		b.Cylinder(v3(x, 1.08, .24), v3(-x*.5, 1.82, .28), .035, .035, 7, leaf, true)
		b.Ellipsoid(v3(x, 1.42, .31), v3(.13, .22, .035), 8, 10, leaf)
	}
	b.Ellipsoid(v3(0, 2.30, 0), v3(.58, .48, .41), 12, 20, orange)
	for _, x := range []float32{-.38, -.2, .2, .38} {
		b.Ellipsoid(v3(x, 2.30, .16), v3(.16, .44, .27), 10, 12, hc(245, 133, 29))
	}
	b.Emis = 1
	for _, x := range []float32{-.22, .22} {
		b.Ellipsoid(v3(x, 2.38, .405), v3(.105, .13, .035), 8, 12, glow)
	}
	for i := 0; i < 8; i++ {
		a := 3.14159 + float32(i)*3.14159/8
		d := 3.14159 + float32(i+1)*3.14159/8
		b.Cylinder(v3(.32*cosf(a), 2.20+.15*sinf(a), .42), v3(.32*cosf(d), 2.20+.15*sinf(d), .42), .035, .035, 6, glow, true)
	}
	b.Emis = 0
	b.Cylinder(v3(0, 2.72, 0), v3(.08, 2.92, 0), .10, .07, 8, green, true)
	for _, x := range []float32{-.20, 0, .20} {
		b.Ellipsoid(v3(x, 2.78, .02), v3(.16, .08, .22), 8, 10, leaf)
	}
	if s.Crown {
		AddCrown(b, 2.92)
	}
	c.Body = b.Build()
	c.HeadTop = 2.92
	halloweenCape(c, green, leaf)
	return c
}
func halloweenCoach(_ rl.Color) rl.Mesh {
	b := NewMB()
	orange := hc(238, 121, 24)
	rib := hc(255, 145, 32)
	green := hc(56, 120, 48)
	gold := hc(255, 205, 84)
	dark := hc(111, 72, 30)
	b.Ellipsoid(v3(0, 1.35, 0), v3(.93, 1.05, 1.55), 14, 20, orange)
	for _, x := range []float32{-.55, 0, .55} {
		b.Ellipsoid(v3(x, 1.35, 0), v3(.35, 1.03, 1.52), 10, 16, rib)
	}
	b.Cylinder(v3(0, 2.28, 0), v3(.18, 2.62, 0), .13, .08, 10, green, true)
	for _, x := range []float32{-.89, .89} {
		for _, z := range []float32{-.65, .65} {
			b.Cylinder(v3(x*.99, 1.48, z), v3(x*1.06, 1.48, z), .31, .31, 16, dark, true)
			b.Emis = 1
			b.Cylinder(v3(x*1.06, 1.48, z), v3(x*1.08, 1.48, z), .25, .25, 16, gold, true)
			b.Emis = 0
		}
		for _, z := range []float32{-1.05, 1.05} {
			for i := 0; i < 24; i++ {
				a := float32(i) * 6.28318 / 24
				d := float32(i+1) * 6.28318 / 24
				r := .10 + .39*float32(i)/24
				r2 := .10 + .39*float32(i+1)/24
				b.Cylinder(v3(x, .53+r*cosf(a), z+r*sinf(a)), v3(x, .53+r2*cosf(d), z+r2*sinf(d)), .055, .055, 6, green, true)
			}
		}
	}
	b.Cylinder(v3(0, .65, 1.45), v3(0, 1.20, 1.85), .045, .045, 8, dark, true)
	b.Emis = 1
	b.Sphere(v3(0, 1.24, 1.85), .14, gold)
	b.Emis = 0
	return b.Build()
}
