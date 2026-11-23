package main

import "testing"

// Regresie V 4.3 („in tunel e apa”): fundul marii plat (y = SeabedY) se deseneaza
// doar in celulele de mare deschisa, adica unde terenul e la SeabedY in toate colturile.
// Oriunde tunelul (cu bolta) urca peste SeabedY, deasupra lui trebuie sa fie teren
// mai sus de SeabedY, altfel fundul marii ar taia tunelul si ar parea apa.
func TestSeabedNeverCrossesTunnels(t *testing.T) {
	w := NewWorldLayout()
	for ci := range w.Corridors {
		c := &w.Corridors[ci]
		for s := c.WallFrom; s <= c.Len; s += 0.5 {
			if c.IsOpen(s) {
				continue
			}
			for _, lat := range []float32{-c.HalfW - 1, 0, c.HalfW + 1} {
				top := c.CeilY(s, clampf(lat, -c.HalfW, c.HalfW))
				if top < SeabedY-0.2 {
					continue
				}
				x, z := c.Point(s, lat)
				// cele 4 colturi ale celulei de teren de 2 m in care cade punctul
				x0 := float32(int((x+432)/2))*2 - 432
				z0 := float32(int((z+252)/2))*2 - 252
				open := true
				for _, q := range [][2]float32{{x0, z0}, {x0 + 2, z0}, {x0, z0 + 2}, {x0 + 2, z0 + 2}} {
					if w.NaturalLand(q[0], q[1]) > SeabedY+0.05 {
						open = false
					}
				}
				if open {
					t.Fatalf("%s: la s=%.1f lat=%.1f (x=%.1f z=%.1f) bolta e la %.2f, peste fundul marii, sub mare deschisa", c.Name, s, lat, x, z, top)
				}
			}
		}
	}
}
