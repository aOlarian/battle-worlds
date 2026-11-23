package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
)

func init() {
	registerRealmChar(RealmCharCode(9, RoleHero), explorerHero)
	if os.Getenv("BATTLE_TEST_EXPLORER") != "" {
		registerRealmChar(RealmCharCode(2, RoleNPC1), explorerHero)
	}
}

func explorerHero(s CharStyle) *CharMeshes {
	c := func(r, g, b uint8) rl.Color { return rl.NewColor(r, g, b, 255) }
	skin, hair := c(241, 185, 148), c(99, 55, 33)
	blue, red, gold := c(36, 112, 220), c(220, 45, 52), c(255, 199, 49)
	white, brown, dark := c(250, 248, 241), c(110, 62, 35), c(55, 32, 27)
	b := NewMB()
	b.Ellipsoid(v3(0, 1.47, 0), v3(.46, .53, .28), 10, 16, blue)
	b.Ellipsoid(v3(0, 1.02, 0), v3(.39, .20, .25), 8, 12, red)
	b.Cylinder(v3(0, 1.90, 0), v3(0, 2.03, 0), .15, .15, 10, skin, true)
	b.Ellipsoid(v3(0, 2.30, 0), v3(.42, .43, .36), 12, 18, skin)
	for _, x := range []float32{-.4, .4} {
		b.Ellipsoid(v3(x, 2.28, 0), v3(.09, .13, .10), 7, 10, skin)
	}
	b.Ellipsoid(v3(0, 2.60, -.06), v3(.41, .20, .34), 9, 14, hair)
	for _, x := range []float32{-.28, -.14, 0, .14, .28} {
		b.Ellipsoid(v3(x, 2.56, .27), v3(.095, .105, .09), 7, 10, hair)
	}
	for _, x := range []float32{-.17, .17} {
		b.Ellipsoid(v3(x, 2.34, .319), v3(.105, .135, .055), 8, 12, white)
		b.Ellipsoid(v3(x, 2.34, .37), v3(.060, .078, .025), 8, 10, brown)
		b.Ellipsoid(v3(x, 2.34, .39), v3(.029, .046, .014), 7, 10, dark)
		b.Sphere(v3(x-.018, 2.37, .405), .018, white)
		b.Ellipsoid(v3(x*1.7, 2.17, .29), v3(.09, .045, .025), 7, 10, c(235, 119, 112))
	}
	b.Ellipsoid(v3(0, 2.23, .36), v3(.06, .045, .045), 7, 10, skin)
	for i := 0; i < 6; i++ {
		x := float32(i-3) * .045
		y := float32(2.13) + .5*x*x
		b.Cylinder(v3(x, y, .337), v3(x+.045, 2.13+.5*(x+.045)*(x+.045), .337), .012, .012, 5, dark, true)
	}
	// Open-faced helmet leaves the hair and face visible.
	b.Ellipsoid(v3(0, 2.67, -.07), v3(.39, .16, .32), 9, 14, gold)
	b.Cylinder(v3(-.39, 2.60, .05), v3(.39, 2.60, .05), .035, .035, 8, gold, true)
	b.Ellipsoid(v3(0, 2.83, -.06), v3(.075, .16, .25), 8, 12, red)
	// Five-pointed star on the chest.
	for i := 0; i < 10; i++ {
		a := float32(i) * 3.14159265 / 5
		a2 := float32(i+1) * 3.14159265 / 5
		r, r2 := float32(.25), float32(.105)
		if i%2 == 1 {
			r, r2 = r2, r
		}
		b.Tri(v3(0, 1.52, .292), v3(sinf(a)*r, 1.52+cosf(a)*r, .285), v3(sinf(a2)*r2, 1.52+cosf(a2)*r2, .285), v3(0, 0, 1), gold)
	}
	if s.Crown {
		AddCrown(b, 3.0)
	}
	cm := &CharMeshes{Body: b.Build(), HeadTop: 3.0}
	leg := NewMB()
	leg.Ellipsoid(v3(0, -.21, 0), v3(.19, .25, .19), 8, 12, red)
	cm.Leg = leg.Build()
	shin := NewMB()
	shin.Ellipsoid(v3(0, -.17, 0), v3(.15, .20, .15), 8, 12, skin)
	shin.Ellipsoid(v3(0, -.39, .09), v3(.18, .12, .26), 8, 12, white)
	for _, x := range []float32{-.1, 0, .1} {
		shin.Cylinder(v3(x, -.29, .18), v3(x, -.32, .27), .018, .018, 5, red, true)
	}
	cm.Shin = shin.Build()
	arm := func() rl.Mesh {
		m := NewMB()
		m.Ellipsoid(v3(0, -.16, 0), v3(.17, .21, .17), 8, 12, blue)
		return m.Build()
	}
	cm.ArmL = arm()
	cm.ArmR = arm()
	fore := func() rl.Mesh {
		m := NewMB()
		m.Ellipsoid(v3(0, -.17, 0), v3(.13, .20, .13), 8, 12, skin)
		m.Sphere(v3(0, -.37, 0), .14, skin)
		return m.Build()
	}
	cm.ForeL = fore()
	cm.ForeR = fore()
	cape := NewMB()
	cape.Quad(v3(-.28, 0, 0), v3(.28, 0, 0), v3(.40, -.65, -.12), v3(-.40, -.65, -.12), v3(0, 0, -1), red)
	cape.Quad(v3(.28, 0, 0), v3(-.28, 0, 0), v3(-.40, -.65, -.12), v3(.40, -.65, -.12), v3(0, 0, 1), red)
	cm.Cape = cape.Build()
	cm.HasCape = true
	return cm
}
