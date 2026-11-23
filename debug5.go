package main

import (
	"fmt"
	"os"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 5.0: verificarile automate noi (BATTLE_DEBUG=fun, pasii 47+;
// BATTLE_DEBUG_FROM=47 le ruleaza doar pe ele).
// ====================================================================

type debug5 struct {
	dummies  []int
	hp0      []float64
	gun      int
	saved    PlayerSave
	followT  float32
	lastWorl int
}

func (d *DebugRun) dummyEnemy(g *Game, pos rl.Vector3) int {
	g.Enemies = append(g.Enemies, Enemy{Name: "tinta", Pos: pos, Scale: 1.2, HP: 1e6, MaxHP: 1e6, Alive: true,
		Meshes: GetCharMeshes(ThemedEnemyStyle(0, 0)), Home: rl.NewVector2(pos.X, pos.Z), HomeR: 1, Stun: 1e6, AtkCD: 1e6})
	return len(g.Enemies) - 1
}

func (d *DebugRun) faceTo(g *Game, p rl.Vector3) {
	q := g.Player.Pos
	g.CamYaw = atan2f(-(p.X - q.X), -(p.Z - q.Z))
	g.CamPitch = 0.1
}

func (d *DebugRun) funStep5(g *Game) bool {
	p := &g.Player
	w := g.City.W
	v := &d.v5
	switch d.step {
	case 47: // lumile inchise: nivelul 1 = doar piata
		d.NoLocks = false
		g.Level = 1
		g.Prog = newProgState()
		g.Prog.GPS.World = -1
		g.syncFog()
		g.spawnEnemies(1)
		okAll := true
		for wi := 1; wi < NumWorlds; wi++ {
			c := Sectors[Worlds[wi].Sector].Center
			okAll = okAll && g.lockedAt(c.X, c.Y, 0.5)
		}
		d.check(okAll && !g.lockedAt(0, 30, 0.3), "lumi: la nivelul 1 sunt inchise toate insulele in afara de piata")
		in := 0
		for i := range g.Enemies {
			if SectorAt(g.Enemies[i].Pos.X, g.Enemies[i].Pos.Z) == 4 {
				in++
			}
		}
		d.check(in == len(g.Enemies) && g.bossIndex() >= 0, "lumi: toti cei %d monstri (cu Seful) sunt in piata", len(g.Enemies))
		// spre muntele NE pe podul scurt: perdeaua de nori il opreste
		m := &w.Mountains[1]
		x, z := m.World(CenterR-12-IsletD, 0)
		d.teleport(g, x, z)
		d.moveDir, d.moveOn = v3(m.Dir.X, 0, m.Dir.Y), true
		d.next()
	case 48:
		if d.stepT > 5 {
			m := &w.Mountains[1]
			u := p.Pos.X*m.Dir.X + p.Pos.Z*m.Dir.Y
			d.check(u < CenterR+6, "lumi: pe jos se opreste la norii de pe podul scurt (u=%.1f)", u)
			// inot spre insula NE din larg
			d.teleport(g, 60, -128)
			d.moveDir = rl.Vector3Normalize(v3(160-60, 0, -160+128))
			d.next()
		}
	case 49:
		if d.stepT > 12 {
			dd := rl.Vector2Distance(rl.NewVector2(p.Pos.X, p.Pos.Z), Sectors[1].Center)
			d.check(dd > 86, "lumi: inotand, nu intra sub nori (la %.0f m de centrul insulei NE)", dd)
			d.moveOn = false
			d.teleport(g, 0, 30)
			d.next()
		}
	case 50: // nivel trecut: Seful apare dupa ceilalti
		for i := range g.Enemies {
			if g.Enemies[i].Alive && !g.Enemies[i].Boss {
				g.defeatEnemy(i, 1)
			}
		}
		bi := g.bossIndex()
		d.check(bi >= 0 && g.Prog.BossOut && g.Enemies[bi].Alive && g.Enemies[bi].Scale > 2, "sef: dupa ultimul monstru a aparut Seful (mare, cu coroana)")
		d.check(g.Level == 1, "sef: nivelul nu trece pana nu e invins Seful")
		if bi >= 0 {
			// ca in joc: crucea e pe Sef cand moare, nivelul trece la sfarsitul cadrului, apoi se deseneaza
			g.AimEnemy = bi
			g.defeatEnemy(bi, 1)
			if g.pendingLevelUp {
				g.doLevelComplete()
			}
			g.Draw()
			d.check(g.AimEnemy < len(g.Enemies), "nivel: ecranul se deseneaza dupa Seful invins cu tinta pe el (jocul nu cade)")
		}
		d.next()
	case 51:
		if d.stepT > 0.2 && d.stepInt == 0 {
			d.stepInt = 1
			d.check(g.Level == 2 && g.Prog.Reveal.Active, "nivel: Seful invins -> nivelul 2 si harta de dezvaluire")
		}
		if d.stepT > 3.5 && d.stepInt == 1 {
			d.stepInt = 2
			d.check(g.revealMapAlpha() > 0.9, "nivel: harta e pe ecran (alfa %.2f)", g.revealMapAlpha())
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v5_reveal.png")
		}
		if d.stepT > revealTotal()+0.5 {
			gp := &g.Prog.GPS
			d.check(!g.Prog.Reveal.Active && g.worldOpen(1), "nivel: harta a disparut, lumea 2 e deschisa")
			d.check(gp.Active && gp.World == 1 && len(gp.Path) > 3, "gps: sagetile duc spre cabina lumii 2 (%d puncte)", len(gp.Path))
			in := 0
			for i := range g.Enemies {
				if SectorAt(g.Enemies[i].Pos.X, g.Enemies[i].Pos.Z) == 1 {
					in++
				}
			}
			d.check(in == len(g.Enemies) && in > 0, "nivel: monstrii nivelului 2 sunt pe insula noua (%d)", in)
			g.Enemies = g.Enemies[:0] // sa nu ne incurce la mers
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v5_gps.png")
			v.followT = 0
			d.next()
		}
	case 52: // mergem dupa sageti (cum ar merge Aventurierul)
		gp := &g.Prog.GPS
		v.followT += g.FrameTime
		if d.stepT > 2.5 && d.stepInt == 0 {
			d.stepInt = 1
			g.Draw()
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v5_gps_arrows.png")
		}
		if gp.Active && len(gp.Path) > 1 {
			// tinta: un punct cu ~5 m inainte pe drum
			tgt := gp.Path[len(gp.Path)-1]
			acc := float32(0)
			for i := 0; i+1 < len(gp.Path); i++ {
				acc += rl.Vector3Distance(gp.Path[i], gp.Path[i+1])
				if acc > 5 {
					tgt = gp.Path[i+1]
					break
				}
			}
			dx, dz := tgt.X-p.Pos.X, tgt.Z-p.Pos.Z
			l := sqrtf(dx*dx+dz*dz) + 0.001
			d.moveDir, d.moveOn = v3(dx/l, 0, dz/l), true
			d.faceTo(g, tgt)
		}
		if !gp.Active || d.stepT > 160 {
			d.moveOn = false
			d.check(!gp.Active && g.currentWorld() == 1, "gps: a ajuns pe insula noua dupa sageti (%.0f s), sagetile au disparut", d.stepT)
			d.check(g.Prog.Visited[1] && g.Prog.OwnOutfit[1] && g.Prog.Outfit == 1, "costum: pe prima insula noua a primit costumul Surferului si il poarta")
			d.next()
		}
	case 53:
		if d.stepT > 4 {
			d.check(g.Prog.Fog[1] < 0.01, "nori: norii de pe insula noua s-au risipit")
			b := &g.City.Booths[1]
			dp := b.DoorPos(1.4)
			d.teleport(g, dp.X, dp.Z)
			d.eQueued = true
			d.next()
		}
	case 54:
		if d.stepT > 0.6 && d.stepInt == 0 {
			d.stepInt = 1
			d.check(g.Scene == SceneBooth, "cabina: E langa cabina deschide ecranul cabinei")
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v5_booth_home.png")
			g.Booth.Tab = BoothMap
			g.Booth.Sel = 0
		}
		if d.stepT > 1.2 && d.stepInt == 1 {
			d.stepInt = 2
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v5_booth_map.png")
			g.Booth.Tab = BoothOutfits
		}
		if d.stepT > 1.8 && d.stepInt == 2 {
			d.stepInt = 3
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v5_booth_outfits.png")
			g.Booth.Tab = BoothGuns
		}
		if d.stepT > 2.4 && d.stepInt == 3 {
			d.stepInt = 4
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v5_booth_guns.png")
			d.check(g.boothWear(0) && g.Prog.Outfit == 0, "cabina: schimba costumul in cel de baza")
			d.check(!g.boothWear(5), "cabina: costumul inca necastigat (Spiridusul) nu se poate purta")
			d.check(!g.boothGun(3), "cabina: racheta e blocata (trebuie stele)")
			g.boothTeleport(0)
		}
		if d.stepT > 3.6 && d.stepInt == 4 {
			b := &g.City.Booths[0]
			dd := rl.Vector3Distance(p.Pos, b.Pos)
			d.check(g.Scene == SceneMapExplore && dd < 5, "cabina: teleport la cabina din piata (%.1f m)", dd)
			d.next()
		}
	case 55: // piesele aurii
		bad := 0
		for wi := range Worlds {
			for _, b := range goldBrickPos[wi] {
				gy, k := w.GroundAt(b.X, b.Z, b.Y+0.5)
				if k == SurfWater || absf(b.Y-0.9-gy) > 0.3 {
					bad++
					d.logf("  piesa gresita lume %d la %.1f %.1f %.1f (sol %.1f, tip %d)", wi, b.X, b.Y, b.Z, gy, k)
				}
			}
		}
		d.check(bad == 0, "piese aurii: toate cele %d stau pe ceva (sol, acoperis, tunel), %d gresite", NumWorlds*BricksPerWorld, bad)
		b := goldBrickPos[0][len(goldBrickPos[0])-1]
		d.teleport(g, b.X, b.Z)
		d.next()
	case 56:
		if d.stepT > 1 {
			d.check(g.bricksFound(0) == 1, "piese aurii: a luat piesa atingand-o (%d/5)", g.bricksFound(0))
			for i := range g.Prog.Bricks[0] {
				g.Prog.Bricks[0][i] = true
			}
			d.next()
		}
	case 57:
		if d.stepT > 0.5 {
			d.check(g.totalStars() >= 2 && g.gunUnlocked(GunRed) && g.Prog.Gun == GunRed, "stele: %d stele -> laserul rosu deblocat si pus in mana", g.totalStars())
			d.teleport(g, 0, 34)
			v.gun = 0
			d.next()
		}
	case 58: // fiecare arma loveste tinta
		if d.stepInt == 0 {
			g.Enemies = g.Enemies[:0]
			fwd := v3(1, 0, 0) // nu spre fantana (0, 20)
			pos := rl.Vector3Add(p.Pos, rl.Vector3Scale(fwd, 22))
			v.dummies = []int{d.dummyEnemy(g, pos), d.dummyEnemy(g, v3(pos.X+3, pos.Y, pos.Z))}
			v.hp0 = []float64{g.Enemies[v.dummies[0]].HP, g.Enemies[v.dummies[1]].HP}
			g.Prog.Gun = v.gun
			g.Player.Pose.Gun = v.gun
			d.stepInt = 1
			d.stepT = 0
		}
		e := &g.Enemies[v.dummies[0]]
		d.faceTo(g, v3(e.Pos.X, e.Pos.Y+1.5, e.Pos.Z))
		d.fireOn = d.stepT > 0.4
		if d.stepT > 3 {
			d.fireOn = false
			dmg := v.hp0[0] - e.HP
			d.check(dmg > 0, "arme: %s loveste tinta (-%.0f)", GunNames[v.gun], dmg)
			if v.gun == GunRocket {
				d2 := v.hp0[1] - g.Enemies[v.dummies[1]].HP
				d.check(d2 > 0, "arme: racheta loveste si tinta de alaturi (explozie, -%.0f)", d2)
			}
			if v.gun == GunBeam {
				g.Draw()
				rl.TakeScreenshot(d.OutDir + "/v5_beam.png")
			}
			v.gun++
			d.stepInt = 0
			if v.gun >= NumGuns {
				g.Enemies = g.Enemies[:0]
				d.next()
			}
		}
	case 59: // salvare si incarcare
		if d.stepInt == 0 {
			g.Prog.RodeWheel = true
			UpdatePlayerData(&v.saved, g)
			lvl, br, st := g.Level, g.bricksFound(0), g.totalStars()
			g.Level = 1
			g.Prog = newProgState()
			LoadPlayerIntoGame(&v.saved, g)
			d.check(g.Level == lvl && g.bricksFound(0) == br && g.totalStars() == st && g.Prog.RodeWheel && g.Prog.OwnOutfit[1], "salvare: nivelul, piesele, stelele, costumele se pastreaza (nivel %d, stele %d)", g.Level, g.totalStars())
			d.stepInt = 1
		}
		if d.stepT > 0.5 {
			d.NoLocks = true
			d.next()
		}
	case 60: // pod nou: din insula SV pe Insula Plajei cu masina
		d.NoLocks = true
		d.teleport(g, WestBridgeA.X+4, WestBridgeA.Y)
		k := -1
		for i := range g.City.Cars {
			if g.City.Cars[i].Hittable() {
				k = i
				break
			}
		}
		g.enterCar(k)
		c := &g.City.Cars[g.Drive.Car]
		dir := rl.Vector2Normalize(rl.Vector2Subtract(WestBridgeB, WestBridgeA))
		c.Pos = v3(WestBridgeA.X+dir.X*2, 0.4, WestBridgeA.Y+dir.Y*2)
		c.Yaw = atan2f(dir.X, dir.Y)
		c.Speed = 0
		d.driveTh, d.driveSt, d.driveOn = 1, 0, true
		d.next()
	case 61:
		c := &g.City.Cars[g.Drive.Car]
		// tine directia podului
		dir := rl.Vector2Normalize(rl.Vector2Subtract(WestBridgeB, WestBridgeA))
		want := atan2f(dir.X, dir.Y)
		dy := want - c.Yaw
		for dy > math32Pi {
			dy -= 2 * math32Pi
		}
		for dy < -math32Pi {
			dy += 2 * math32Pi
		}
		d.driveSt = clampf(dy*3, -1, 1)
		if SectorAt(c.Pos.X, c.Pos.Z) == SecWest || d.stepT > 25 {
			d.check(SectorAt(c.Pos.X, c.Pos.Z) == SecWest, "pod nou: cu masina din insula SV pe Insula Plajei (%.0f s)", d.stepT)
			d.driveOn = false
			g.exitCar()
			d.next()
		}
	default:
		return d.funStep6(g)
	}
	return false
}

// setupRealmShots: in lumile de uscat doar capturile date in BATTLE_DEBUG_VIEWS
// (sau un set implicit: centrul, de sus, fiecare zona, gara).
func (d *DebugRun) setupRealmShots(g *Game) {
	d.Views = nil
	if v := os.Getenv("BATTLE_DEBUG_VIEWS"); v != "" {
		for _, part := range strings.Split(v, ";") {
			nv := strings.SplitN(part, ":", 2)
			if len(nv) != 2 {
				continue
			}
			var x, y, z, yaw, pitch, dist float32
			fmt.Sscanf(nv[1], "%f,%f,%f,%f,%f,%f", &x, &y, &z, &yaw, &pitch, &dist)
			d.Views = append(d.Views, DebugView{Name: nv[0], Pos: v3(x, y, z), CamYaw: yaw, CamPitch: pitch, CamDist: dist, Wait: 10})
		}
	} else {
		id := Realms[curRealm].ID
		st := g.City.StationDoor()
		d.Views = append(d.Views,
			DebugView{Name: id + "_01_gara", Pos: st, CamYaw: 0.5, CamPitch: 0.25, CamDist: 16, Wait: 10},
			DebugView{Name: id + "_02_sus", Pos: v3(0, 0, 20), CamYaw: 0.6, CamPitch: 1.1, CamDist: 200, Wait: 10},
		)
		// trenul magic la gara (vine din cer si opreste la peron)
		_, toHub := stationAxes(g.City.Station)
		d.Views = append(d.Views, DebugView{Name: id + "_02b_tren", Pos: v3(st.X+toHub.X*14, 0, st.Z+toHub.Y*14), CamYaw: atan2f(toHub.X, toHub.Y), CamPitch: 0.3, CamDist: 22, Wait: 200, Setup: func(g *Game) { g.callTrain(curRealm) }})
		for i, z := range g.City.Zones {
			d.Views = append(d.Views,
				DebugView{Name: fmt.Sprintf("%s_%02d_zona%d", id, 3+i*2, i), Pos: v3(z.C.X+8, 0, z.C.Y+22), CamYaw: float32(i) * 1.3, CamPitch: 0.25, CamDist: 13, Wait: 10},
				DebugView{Name: fmt.Sprintf("%s_%02d_zona%d_sus", id, 4+i*2, i), Pos: v3(z.C.X, 0, z.C.Y), CamYaw: float32(i)*1.3 + 0.4, CamPitch: 0.75, CamDist: 70, Wait: 10},
			)
		}
	}
	d.applyView(g)
}
