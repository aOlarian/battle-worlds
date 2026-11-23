package main

import rl "github.com/gen2brain/raylib-go/raylib"

func init() { registerRealmDecor("monstri", monstriDecor) }

func monstriDecor(rb *RealmBuilder) {
	purple := rl.NewColor(75, 43, 100, 255)
	roof := rl.NewColor(48, 29, 71, 255)
	wood := rl.NewColor(105, 68, 59, 255)
	gold := rl.NewColor(255, 192, 75, 255)
	stone := rl.NewColor(139, 119, 150, 255)
	green := rl.NewColor(43, 83, 66, 255)
	orange := rl.NewColor(239, 121, 39, 255)
	// Locate landmarks on the outer ring, clear of roads and reserved plots.
	place := func(zi int, radius, size float32) (float32, float32, bool) {
		c := rb.Zones[zi].C
		for i := 0; i < 96; i++ {
			a := float32(i) * 6.2831853 / 96
			x, z := c.X+cosf(a)*radius, c.Y+sinf(a)*radius
			if rb.Free(x, z, size) {
				return x, z, true
			}
		}
		return 0, 0, false
	}
	for zi := 0; zi < 3; zi++ {
		mb := rb.Mesh()
		if zi == 0 {
			x, z, ok := place(0, 49, 7)
			if ok {
				y := rb.H(x, z)
				mb.Box(v3(x, y+7, z), v3(8, 14, 8), wood)
				mb.Cylinder(v3(x, y+14, z), v3(x, y+21, z), 6, 0, 4, roof, true)
				mb.Emis = 1
				mb.Sphere(v3(x, y+11, z-4.1), 2.2, gold)
				mb.Emis = 0
				mb.Box(v3(x, y+11, z-4.25), v3(.22, 3, .15), roof)
				mb.Box(v3(x+.6, y+11, z-4.25), v3(1.3, .2, .15), roof)
				rb.Solid(x, z, 8, 8, 14, y)
				rb.Reserve(x, z, 7)
			}
		}
		if zi == 1 {
			x, z, ok := place(1, 55, 16)
			if ok {
				y := rb.H(x, z)
				mb.BeveledBox(v3(x, y+10, z), v3(22, 20, 14), .5, stone)
				rb.Solid(x, z, 22, 14, 20, y)
				for _, dx := range []float32{-10, 10} {
					for _, dz := range []float32{-6, 6} {
						tx, tz := x+dx, z+dz
						mb.Cylinder(v3(tx, y, tz), v3(tx, y+25, tz), 3.3, 3, 10, purple, true)
						mb.Cylinder(v3(tx, y+25, tz), v3(tx, y+35, tz), 4, 0, 10, roof, true)
						mb.Cylinder(v3(tx, y+35, tz), v3(tx, y+39, tz), .12, .12, 5, gold, true)
						mb.Box(v3(tx+1, y+37.8, tz), v3(2, .9, .12), purple)
						rb.Pillar(tx, tz, 3.3, y+25)
					}
				}
				mb.Box(v3(x, y+3, z-7.1), v3(4, 6, .3), wood)
				for row := 0; row < 3; row++ {
					for col := -3; col <= 3; col++ {
						if row == 0 && col == 0 {
							continue
						}
						mb.Emis = 1
						mb.Box(v3(x+float32(col)*2.7, y+4+float32(row)*5, z-7.12), v3(1.4, 2.2, .2), gold)
						mb.Emis = 0
					}
				}
				for i := 0; i < 4; i++ {
					mb.Box(v3(x, y+.2+float32(i)*.25, z-9-float32(i)*.6), v3(6, .4, 1), stone)
				}
				rb.Reserve(x, z, 16)
			}
		}
		if zi == 2 {
			x, z, ok := place(2, 52, 9)
			if ok {
				y := rb.H(x, z)
				mb.Cylinder(v3(x, y, z), v3(x, y+24, z), 4.5, 3, 12, wood, true)
				mb.Ellipsoid(v3(x, y+25, z), v3(12, 8, 10), 8, 12, green)
				mb.Box(v3(x, y+12, z-3), v3(8, 7, 5), purple)
				mb.Cylinder(v3(x, y+15, z-3), v3(x, y+20, z-3), 6, 0, 4, roof, true)
				mb.Emis = 1
				mb.Box(v3(x, y+12, z-5.6), v3(2, 2, .2), gold)
				mb.Emis = 0
				rb.Pillar(x, z, 4.5, y+24)
				rb.Reserve(x, z, 9)
			}
		}
		c := rb.Zones[zi].C
		for i := 0; i < 160; i++ {
			a := rb.Rng.Float32() * 6.2831853
			r := 18 + rb.Rng.Float32()*49
			x, z := c.X+cosf(a)*r, c.Y+sinf(a)*r
			if !rb.Free(x, z, 1.4) {
				continue
			}
			y := rb.H(x, z)
			switch i % 4 {
			case 0: // warm lantern
				mb.Cylinder(v3(x, y, z), v3(x, y+4, z), .13, .1, 6, roof, true)
				mb.Emis = 1
				mb.Box(v3(x, y+4, z), v3(.8, 1, .8), gold)
				mb.Emis = 0
				mb.Cylinder(v3(x, y+4.5, z), v3(x, y+5, z), .7, 0, 4, roof, true)
			case 1: // pumpkin
				mb.Ellipsoid(v3(x, y+.65, z), v3(1, .65, .8), 6, 10, orange)
				mb.Cylinder(v3(x, y+1.2, z), v3(x, y+1.6, z), .14, .1, 5, green, true)
			case 2:
				if zi == 2 {
					mb.Cylinder(v3(x, y, z), v3(x, y+2.3, z), .4, .3, 7, stone, true)
					mb.Ellipsoid(v3(x, y+2.4, z), v3(1.6, .7, 1.6), 5, 9, purple)
					mb.Emis = 1
					mb.Sphere(v3(x+.6, y+3, z), .2, gold)
					mb.Emis = 0
				} else {
					mb.Ellipsoid(v3(x, y+1, z), v3(1.6, 1.1, 1.4), 5, 8, green)
				}
			case 3:
				if zi == 2 {
					mb.Emis = 1
					for j := 0; j < 3; j++ {
						mb.Sphere(v3(x+float32(j)*.5, y+1.5+float32(j)*.7, z), .16, gold)
					}
					mb.Emis = 0
				} else {
					mb.Box(v3(x, y+.8, z), v3(2, .3, .8), wood)
					for _, dx := range []float32{-.7, .7} {
						mb.Box(v3(x+dx, y+.4, z), v3(.18, .8, .6), roof)
					}
				}
			}
			rb.Reserve(x, z, 1.4)
		}
		rb.Add(mb, true)
	}
}
