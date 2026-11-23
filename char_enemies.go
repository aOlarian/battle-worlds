package main

func decorateEnemy(b *MeshBuilder, s CharStyle) {
	switch s.Costume {
	case 10: // brick golem: studs and mortar
		for _, z := range []float32{-.27, .27} {
			for row := 0; row < 3; row++ {
				y := float32(1.25) + float32(row)*.22
				b.Box(v3(0, y, z), v3(.85, .035, .03), shade(s.Shirt, .55))
				for col := -1; col <= 1; col++ {
					x := float32(col)*.25 + float32(row%2)*.12
					b.Box(v3(x, y+.1, z), v3(.025, .18, .03), shade(s.Shirt, .55))
				}
			}
		}
		for _, x := range []float32{-.22, .22} {
			b.Cylinder(v3(x, 2.64, 0), v3(x, 2.77, 0), .11, .11, 10, s.Shirt, true)
		}
	case 11: // friendly spider robot, mask and four side arms
		b.Box(v3(0, 2.4, .32), v3(.62, .19, .08), shade(s.Shirt, .5))
		for _, x := range []float32{-.15, .15} {
			b.Ellipsoid(v3(x, 2.4, .37), v3(.09, .065, .03), 4, 8, LegoWhite)
		}
		for _, side := range []float32{-1, 1} {
			for k := 0; k < 2; k++ {
				y := float32(1.4) + float32(k)*.32
				b.Cylinder(v3(side*.4, y, -.2), v3(side*.75, y+.25, -.25), .055, .055, 6, LegoLightGrey, true)
				b.Cylinder(v3(side*.75, y+.25, -.25), v3(side*.9, y-.08, -.15), .055, .035, 6, LegoLightGrey, true)
			}
		}
	case 12: // rounded dino snout and orange crest
		b.Ellipsoid(v3(0, 2.29, .32), v3(.31, .17, .27), 5, 10, s.Shirt)
		for k := 0; k < 4; k++ {
			z := float32(.18) - float32(k)*.17
			b.Cylinder(v3(0, 2.63, z), v3(0, 2.86, z-.06), .09, .015, 6, LegoOrange, true)
		}
	case 13: // drone antenna and rotor pods
		b.Cylinder(v3(0, 2.65, 0), v3(0, 2.98, 0), .035, .035, 6, LegoLightGrey, true)
		b.Sphere(v3(0, 3, 0), .09, LegoYellow)
		for _, x := range []float32{-.62, .62} {
			b.Cylinder(v3(x, 1.95, 0), v3(x, 2.04, 0), .25, .25, 12, LegoLightGrey, true)
			b.Box(v3(x, 2.08, 0), v3(.55, .04, .09), LegoYellow)
		}
		b.Box(v3(0, 1.5, .27), v3(.35, .2, .04), LegoCyan)
	case 14: // jester cap and round bells
		for _, x := range []float32{-.27, .27} {
			b.Cylinder(v3(x, 2.58, 0), v3(x*1.6, 2.98, 0), .18, .035, 8, s.Shirt, true)
			b.Sphere(v3(x*1.6, 2.98, 0), .085, LegoYellow)
		}
		b.Ellipsoid(v3(0, 2.22, .32), v3(.16, .07, .035), 4, 8, LegoRed)
		for _, x := range []float32{-.28, .28} {
			b.Sphere(v3(x, 1.85, .15), .13, LegoYellow)
		}
	case 15: // pirate bandana and friendly pincers
		b.Cylinder(v3(0, 2.56, 0), v3(0, 2.69, 0), .35, .35, 14, LegoRed, true)
		b.Box(v3(.31, 2.54, -.1), v3(.16, .23, .08), LegoRed)
		for _, x := range []float32{-.74, .74} {
			b.Ellipsoid(v3(x, 1.19, .08), v3(.23, .2, .16), 5, 10, s.Shirt)
			b.Sphere(v3(x*1.18, 1.3, .1), .1, LegoOrange)
		}
		b.Box(v3(0, 1.55, .27), v3(.55, .08, .04), LegoWhite)
	case 16: // green furry grump carrying a gift sack
		b.Ellipsoid(v3(0, 2.57, -.05), v3(.37, .2, .35), 5, 12, s.Shirt)
		for _, x := range []float32{-.17, .17} {
			b.Cylinder(v3(x-.09, 2.53, .32), v3(x+.09, 2.49, .32), .035, .035, 6, shade(s.Shirt, .45), true)
		}
		for k := 0; k < 5; k++ {
			x := float32(k-2) * .15
			b.Cylinder(v3(x, 2.7, 0), v3(x+.04, 2.86, 0), .07, .01, 5, s.Shirt, true)
		}
		b.Ellipsoid(v3(0, 1.45, -.43), v3(.39, .5, .27), 6, 12, LegoRed)
		b.Box(v3(0, 1.45, -.7), v3(.1, .7, .04), LegoYellow)
		b.Box(v3(0, 1.45, -.71), v3(.65, .1, .04), LegoYellow)
	}
}
