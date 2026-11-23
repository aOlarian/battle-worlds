package main

import (
	"fmt"
	"math"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 5.0: CABINELE TELEFONICE (cate una in fiecare lume).
// In cabina: harta (navigare / teleport), personaje (costume), arme.
// ====================================================================

type Booth struct {
	Pos    rl.Vector3
	Yaw    float32 // usa priveste spre +Z rotit cu Yaw
	Sector int
	World  int
}

// DoorPos: unde sta Aventurierul cand iese / se teleporteaza (in fata usii).
func (b *Booth) DoorPos(d float32) rl.Vector3 {
	return v3(b.Pos.X+sinf(b.Yaw)*d, b.Pos.Y, b.Pos.Z+cosf(b.Yaw)*d)
}

// placeBooths: alege un loc liber langa centrul fiecarei lumi, inainte de cladiri.
func (cb *CityBuilder) placeBooths(cw *CityWorld) {
	cb.bridgeLandings()
	for wi, wd := range Worlds {
		c := Sectors[wd.Sector].Center
		var best rl.Vector2
		bestScore := float32(1e9)
		prefR := float32(20)
		pref := c
		switch wd.Sector {
		case 4:
			pref = rl.NewVector2(11, 31) // langa locul de pornire, in piata
			prefR = 0
		case SecWest, SecPark:
			// langa capatul podului nou, spre interiorul insulei
			land := WestBridgeB
			if wd.Sector == SecPark {
				land = EastBridgeB
			}
			in := rl.Vector2Normalize(rl.Vector2Subtract(c, land))
			pref = rl.Vector2Add(land, rl.Vector2Scale(in, 15))
			prefR = 0
		}
		for ri := 0; ri < 14; ri++ {
			r := 12 + float32(ri)*2.5
			for ai := 0; ai < 32; ai++ {
				a := float32(ai) / 32 * 2 * math.Pi
				x := c.X + cosf(a)*r
				z := c.Y + sinf(a)*r
				if prefR == 0 {
					x, z = pref.X+cosf(a)*(r-10), pref.Y+sinf(a)*(r-10)
				}
				if cb.nearRoad(x, z, 2.2) || cb.reserved(x, z, 2.5) || cb.inMountain(x, z, 3) {
					continue
				}
				if cb.W.LandSD(x, z) > -9 || cb.W.stairMouth(x, z, 3) || cb.W.nearOpenCut(x, z, 3) {
					continue
				}
				var sc float32
				if prefR == 0 {
					sc = rl.Vector2Distance(rl.NewVector2(x, z), pref)
				} else {
					sc = absf(r - prefR)
				}
				if sc < bestScore {
					bestScore, best = sc, rl.NewVector2(x, z)
				}
			}
		}
		yaw := atan2f(c.X-best.X, c.Y-best.Y) // usa spre centrul insulei
		if wd.Sector == 4 {
			yaw = atan2f(-best.X, -best.Y)
		}
		cw.Booths = append(cw.Booths, Booth{Pos: v3(best.X, 0, best.Y), Yaw: yaw, Sector: wd.Sector, World: wi})
		cb.Reserve = append(cb.Reserve, CircleCollider{X: best.X, Z: best.Y, R: 4.5})
	}
}

// coliziunea cabinei (dupa ce s-a stabilit inaltimea solului)
func (cw *CityWorld) finishBooths() {
	w := cw.W
	for i := range cw.Booths {
		b := &cw.Booths[i]
		gy, _ := w.GroundAt(b.Pos.X, b.Pos.Z, 5)
		b.Pos.Y = gy
		w.Circles = append(w.Circles, CircleCollider{X: b.Pos.X, Z: b.Pos.Z, R: 1.0, Top: gy + 2.9})
		if os.Getenv("BATTLE_DEBUG") != "" {
			fmt.Printf("cabina %d (%s): x=%.1f y=%.1f z=%.1f yaw=%.2f\n", i, Worlds[b.World].Name, b.Pos.X, b.Pos.Y, b.Pos.Z, b.Yaw)
		}
	}
}

func (cw *CityWorld) drawBooths(draw DrawFn, near func(rl.Vector3, float32) bool) {
	for i := range cw.Booths {
		b := &cw.Booths[i]
		if !near(b.Pos, 3) {
			continue
		}
		draw(boothMesh, rl.MatrixMultiply(rl.MatrixRotateY(b.Yaw), rl.MatrixTranslate(b.Pos.X, b.Pos.Y, b.Pos.Z)), rl.White)
	}
}

// bridgeLandings: platforma pavata la capetele podurilor noi (pe insula plajei si in parc).
func (cb *CityBuilder) bridgeLandings() {
	mb := NewMB()
	for _, bi := range []int{BridgeWestIdx, BridgeEastIdx} {
		b := &cb.W.Bridges[bi]
		land := b.B
		d := rl.Vector2Normalize(rl.Vector2Subtract(b.B, b.A))
		end := rl.Vector2Add(land, rl.Vector2Scale(d, 9))
		cb.roadArm(mb, land, end, 4.5, true, true)
		cb.Reserve = append(cb.Reserve, CircleCollider{X: land.X, Z: land.Y, R: 9}, CircleCollider{X: end.X, Z: end.Y, R: 6})
	}
	cb.addChunk(mb, false)
}
