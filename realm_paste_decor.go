package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

func init() { registerRealmDecor("paste", pasteDecor) }

func pasteDecor(rb *RealmBuilder) {
	pink := rl.NewColor(250, 145, 185, 255)
	white := rl.NewColor(255, 246, 230, 255)
	yellow := rl.NewColor(255, 211, 54, 255)
	red := rl.NewColor(207, 55, 50, 255)
	green := rl.NewColor(74, 153, 65, 255)
	brown := rl.NewColor(135, 87, 47, 255)
	blue := rl.NewColor(120, 210, 230, 255)
	colors := []rl.Color{pink, yellow, red, white, rl.NewColor(177, 107, 211, 255)}
	for zi := 0; zi < 3 && zi < len(rb.Zones); zi++ {
		c := rb.Zones[zi].C
		mb := rb.Mesh()
		// Pick an open location on the perimeter for each landmark.
		var lx, lz float32
		found := false
		for k := 0; k < 32; k++ {
			a := float64(k) * math.Pi / 16
			x := c.X + float32(math.Cos(a))*43
			z := c.Y + float32(math.Sin(a))*43
			if rb.Free(x, z, 10) {
				lx = x
				lz = z
				found = true
				break
			}
		}
		if found {
			y := rb.H(lx, lz)
			rb.Reserve(lx, lz, 10)
			switch zi {
			case 0:
				mb.Cylinder(v3(lx, y, lz), v3(lx, y+1, lz), 8, 8, 20, green, true)
				mb.EllipsoidF(v3(lx, y+8.4, lz), v3(5.2, 7.4, 5.2), 16, 24, func(n rl.Vector3) rl.Color {
					if n.Y > .28 && n.Y < .44 || n.Y < -.27 && n.Y > -.42 {
						return yellow
					}
					if math.Sin(float64(n.X*18))*math.Sin(float64(n.Z*18)) > .55 {
						return white
					}
					return pink
				})
				rb.Pillar(lx, lz, 5.2, y+16)
				for k := 0; k < 18; k++ {
					a := float64(k) * math.Pi / 9
					x := lx + float32(math.Cos(a))*7
					z := lz + float32(math.Sin(a))*7
					mb.Sphere(v3(x, y+1.1, z), .65, colors[k%len(colors)])
				}
			case 1:
				mb.Box(v3(lx, y+5, lz), v3(17, 10, 13), red)
				mb.Box(v3(lx, y+10.5, lz), v3(18, 1, 14), white)
				mb.Ellipsoid(v3(lx, y+10, lz), v3(9, 4, 7), 8, 16, red)
				mb.Box(v3(lx, y+3, lz-6.6), v3(5, 6, .3), brown)
				mb.Box(v3(lx, y+3, lz-6.8), v3(.35, 6, .3), white)
				mb.Box(v3(lx, y+3, lz-6.8), v3(5, .35, .3), white)
				rb.Solid(lx, lz, 17, 13, 10, y)
			case 2:
				mb.Box(v3(lx, y+3.5, lz), v3(15, 7, 10), blue)
				mb.Box(v3(lx, y+7.2, lz), v3(16, .5, 11), white)
				for j := -3; j <= 3; j++ {
					mb.Box(v3(lx+float32(j)*2.4, y+3.5, lz-5.1), v3(.18, 7, .2), white)
				}
				for j := -2; j <= 2; j++ {
					mb.Box(v3(lx-7.6, y+3.5, lz+float32(j)*2.4), v3(.2, 7, .18), white)
				}
				rb.Solid(lx, lz, 15, 10, 7, y)
			}
		}
		// Themed objects around the battle area; leave center and roads clear.
		for k := 0; k < 100; k++ {
			a := rb.Rng.Float64() * 2 * math.Pi
			rad := 22 + rb.Rng.Float64()*39
			x := c.X + float32(math.Cos(a)*rad)
			z := c.Y + float32(math.Sin(a)*rad)
			if !rb.Free(x, z, 1.3) {
				continue
			}
			y := rb.H(x, z)
			if k%9 == 0 { // flowering tree
				if !rb.Free(x, z, 2.5) {
					continue
				}
				rb.Reserve(x, z, 2.5)
				mb.Cylinder(v3(x, y, z), v3(x, y+4, z), .45, .3, 7, brown, true)
				mb.Ellipsoid(v3(x, y+5, z), v3(2.5, 2.2, 2.5), 7, 10, pink)
				rb.Pillar(x, z, .5, y+6)
			} else if zi == 0 && k%13 == 0 { // grassy burrow
				if !rb.Free(x, z, 4) {
					continue
				}
				rb.Reserve(x, z, 4)
				mb.Ellipsoid(v3(x, y+1, z), v3(3.4, 2.5, 3), 8, 12, green)
				mb.Ellipsoid(v3(x, y+1, z-2.8), v3(1, 1.3, .2), 8, 10, brown)
				rb.Pillar(x, z, 3, y+3.5)
			} else if zi == 1 && k%8 == 0 {
				mb.Box(v3(x, y+.65, z), v3(2, 1.3, 1.2), yellow)
			} else { // flowers in clusters
				for j := 0; j < 5; j++ {
					dx := float32(j%3) * .42
					dz := float32(j/3) * .45
					h := float32(.65) + float32(j%2)*.2
					mb.Cylinder(v3(x+dx, y, z+dz), v3(x+dx, y+h, z+dz), .045, .04, 4, green, false)
					mb.Ellipsoid(v3(x+dx, y+h, z+dz), v3(.23, .28, .23), 4, 6, colors[(k+j)%len(colors)])
				}
			}
		}
		rb.Add(mb, true)
	}
	// Groves and flowers between the zones, split into small chunks.
	for gx := -2; gx <= 2; gx++ {
		for gz := -2; gz <= 2; gz++ {
			mb := rb.Mesh()
			for k := 0; k < 35; k++ {
				x := float32(gx*100) + float32(rb.Rng.Float64()*90-45)
				z := float32(gz*100) + float32(rb.Rng.Float64()*90-45)
				if rb.ZoneAt(x, z) >= 0 || !rb.Free(x, z, 3) {
					continue
				}
				y := rb.H(x, z)
				rb.Reserve(x, z, 3)
				mb.Cylinder(v3(x, y, z), v3(x, y+4, z), .5, .3, 7, brown, true)
				mb.Ellipsoid(v3(x, y+5, z), v3(2.7, 2.2, 2.7), 7, 10, pink)
				rb.Pillar(x, z, .5, y+7)
				for j := 0; j < 4; j++ {
					mb.Sphere(v3(x+float32(j)-1.5, y+.35, z+2), .3, colors[(k+j)%len(colors)])
				}
			}
			rb.Add(mb, true)
		}
	}
}
