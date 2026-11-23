package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ====================================================================
// V 6.0: verificarile noi (BATTLE_DEBUG=fun, pasii 62+;
// BATTLE_DEBUG_FROM=62 le ruleaza doar pe ele).
// ====================================================================

type debug6 struct {
	maxWind  int
	minOther float32
	maxY     float32
	baseApex float32
	saved    PlayerSave
}

func (d *DebugRun) waitTrainDone(g *Game) bool {
	return g.Train.Phase == TrainOff || (g.Train.Phase == TrainUnload && g.Train.T > 1.3)
}

func (d *DebugRun) funStep6(g *Game) bool {
	p := &g.Player
	v := &d.v6
	switch d.step {
	case 62: // monstrii ataca pe rand
		d.NoLocks = true
		g.Level = 1
		g.spawnEnemies(1)
		d.teleport(g, 0, 30)
		for i := range g.Enemies {
			e := &g.Enemies[i]
			if e.Boss {
				continue
			}
			a := 0.5 + float32(i%6)*0.4 // spre sud, pe partea libera a pietei
			e.Pos = v3(p.Pos.X+cosf(a)*12, p.Pos.Y, p.Pos.Z+sinf(a)*12)
			e.Home = rl.NewVector2(p.Pos.X, p.Pos.Z)
			e.HomeR = 40
			e.AtkCD = 0
		}
		v.maxWind, v.minOther = 0, 1e9
		g.Invuln = 0
		d.next()
	case 63:
		n := 0
		for i := range g.Enemies {
			e := &g.Enemies[i]
			if e.Alive && e.Windup > 0 {
				n++
			}
		}
		if n > v.maxWind {
			v.maxWind = n
		}
		if d.stepT > 4 {
			for i := range g.Enemies {
				e := &g.Enemies[i]
				if !e.Alive || i == g.Attacker || e.Boss {
					continue
				}
				dd := sqrtf(sq(e.Pos.X-p.Pos.X) + sq(e.Pos.Z-p.Pos.Z))
				if dd < v.minOther {
					v.minOther = dd
				}
			}
		}
		g.Player.HP = g.Player.MaxHP
		if d.stepT > 6.9 && d.stepT < 6.95 && g.Attacker >= 0 && g.Attacker < len(g.Enemies) {
			a := &g.Enemies[g.Attacker]
			fmt.Printf("atacator %d: dist %.1f ranged %v stun %.1f atkcd %.1f alive %v\n", g.Attacker, sqrtf(sq(a.Pos.X-p.Pos.X)+sq(a.Pos.Z-p.Pos.Z)), a.Ranged, a.Stun, a.AtkCD, a.Alive)
		}
		if d.stepT > 7 {
			d.check(v.maxWind <= 1, "pe rand: cel mult un monstru isi pregateste lovitura odata (maxim %d)", v.maxWind)
			d.check(v.minOther > 9, "pe rand: ceilalti asteapta la distanta, nu vin peste Aventurierul (cel mai aproape %.1f m)", v.minOther)
			g.Enemies = g.Enemies[:0]
			d.next()
		}
	case 64: // ultima insula a lumii -> trenul magic
		d.NoLocks = false
		g.Level = NumWorlds
		g.Prog.OwnOutfit[1] = true
		g.doLevelComplete()
		d.check(g.Train.Phase == TrainArrive && g.UnlockedRealm >= 1, "tren: dupa ultima insula vine trenul magic si se deschide lumea 2")
		d.check(g.Prog.GPS.Active, "tren: sagetile GPS duc la gara")
		d.next()
	case 65:
		if d.stepT > 2 && d.stepInt == 0 {
			d.stepInt = 1
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v6_tren_vine.png")
		}
		if g.Train.Phase == TrainWait {
			door := g.City.StationDoor()
			p.Pos = door
			d.next()
		} else if d.stepT > 10 {
			d.check(false, "tren: trenul a ajuns la gara (faza %d)", g.Train.Phase)
			d.next()
		}
	case 66:
		if g.Train.Phase == TrainArriveNew && d.stepInt == 0 && g.Train.T > 2.5 {
			d.stepInt = 1
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v6_tren_soseste.png")
		}
		if d.waitTrainDone(g) || d.stepT > 25 {
			oi := outfitOfRealm(1)
			d.check(curRealm == 1 && Realms[curRealm].ID == "monstri", "tren: Aventurierul a ajuns in Hotelul Monstrilor (lumea %d)", curRealm)
			d.check(g.Prog.Outfit == oi && g.ability() == AbilGlide, "tren: a primit si poarta Micul Vampir (planeaza)")
			d.check(NumWorlds == 5 && len(g.Enemies) > 0, "lume noua: 5 insule si monstrii primului nivel (%d)", len(g.Enemies))
			d.next()
		}
	case 67: // planare
		if d.stepInt == 0 {
			g.Enemies = g.Enemies[:0]
			d.teleport(g, 0, 30)
			p.Pos.Y += 30
			p.OnGround = false
			p.VelY = 0
			d.jumpOn = true
			d.stepInt = 1
		}
		if d.stepT > 1.5 {
			d.check(p.VelY > -2.6 && !p.OnGround, "Micul Vampir: tinand SPATIU planeaza (viteza de cadere %.1f)", p.VelY)
			d.jumpOn = false
			g.travelNow(2)
			d.next()
		}
	case 68: // calatorie din cabina spre Orasul Panzelor + panza
		if d.waitTrainDone(g) || d.stepT > 25 {
			d.check(curRealm == 2 && g.ability() == AbilWeb, "cabina: calatorie spre Orasul Panzelor, poarta Baiatul-Paianjen")
			g.Enemies = g.Enemies[:0]
			// cel mai inalt acoperis din apropiere
			w := g.City.W
			best := -1
			for i := range w.Boxes {
				b := &w.Boxes[i]
				cx, cz := (b.MinX+b.MaxX)/2, (b.MinZ+b.MaxZ)/2
				if b.Roof == RoofNone || b.Top < 15 || b.Top > 60 || sq(cx-p.Pos.X)+sq(cz-p.Pos.Z) > 70*70 || b.MaxX-b.MinX < 5 {
					continue
				}
				if best < 0 || b.Top > w.Boxes[best].Top {
					best = i
				}
			}
			if best >= 0 {
				b := &w.Boxes[best]
				t := v3((b.MinX+b.MaxX)/2, b.Top+0.3, (b.MinZ+b.MaxZ)/2)
				dir := rl.Vector3Normalize(v3(t.X-p.Pos.X, 0, t.Z-p.Pos.Z))
				p.Pos = v3(t.X-dir.X*30, 0, t.Z-dir.Z*30)
				gy, _ := w.GroundAt(p.Pos.X, p.Pos.Z, 5)
				p.Pos.Y = gy
				g.Zip = ZipState{Active: true, Target: t}
				v.maxY = b.Top
			} else {
				d.check(false, "panza: n-am gasit un acoperis inalt langa Aventurierul")
			}
			d.next()
		}
	case 69:
		if d.stepT > 4 {
			_, k := g.City.W.GroundAt(p.Pos.X, p.Pos.Z, p.Pos.Y+0.5)
			d.check(p.Pos.Y > v.maxY-1.5 && (k == SurfRoof || !g.Zip.Active), "panza: s-a tras pe acoperis (y=%.1f, acoperis %.1f)", p.Pos.Y, v.maxY)
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v6_panza.png")
			g.travelNow(3)
			d.next()
		}
	case 70: // zbor
		if d.stepInt == 0 && d.waitTrainDone(g) {
			d.check(curRealm == 3 && g.ability() == AbilFly, "calatorie spre Orasul Eroilor, poarta Eroul Zburator")
			g.Enemies = g.Enemies[:0]
			d.teleport(g, 0, 30)
			d.jumpOn = true
			d.stepInt = 1
			d.stepT = 0
		}
		if d.stepInt == 1 && d.stepT > 2.5 {
			d.check(p.Pos.Y > 12, "Eroul Zburator: tinand SPATIU zboara (y=%.1f)", p.Pos.Y)
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v6_zbor.png")
			d.jumpOn = false
			g.travelNow(5)
			d.next()
		}
		if d.stepT > 30 && d.stepInt == 0 {
			d.check(false, "calatorie spre Orasul Eroilor (timp depasit)")
			d.next()
		}
	case 71: // salt de iepuras
		if d.stepInt == 0 && d.waitTrainDone(g) {
			d.check(curRealm == 5 && g.ability() == AbilHop, "calatorie spre Pajistea de Paste, poarta Iepurasul")
			g.Enemies = g.Enemies[:0]
			d.teleport(g, 0, 30)
			v.maxY = p.Pos.Y
			v.baseApex = p.Pos.Y
			d.jumpOn = true
			d.stepInt = 1
			d.stepT = 0
		}
		if d.stepInt == 1 {
			if p.Pos.Y > v.maxY {
				v.maxY = p.Pos.Y
			}
			if d.stepT > 0.2 {
				d.jumpOn = false
			}
			if d.stepT > 2.5 {
				d.check(v.maxY-v.baseApex > 3.5, "Iepurasul: sare mult mai sus decat Aventurierul (%.1f m fata de ~1.7 m)", v.maxY-v.baseApex)
				g.travelNow(0)
				d.next()
			}
		}
	case 72: // inapoi la mare, progresul lumii 0 ramas
		if d.waitTrainDone(g) || d.stepT > 30 {
			d.check(curRealm == 0 && NumWorlds == 7 && g.City.Realm == 0, "inapoi in Lumea Marii (7 insule)")
			d.check(g.Level == 8, "lumea 0 si-a pastrat nivelul (%d)", g.Level)
			// surf: Surferul pe apa sta pe placa
			oi := outfitOfRealm(0)
			g.Prog.OwnOutfit[oi] = true
			g.wearOutfit(oi)
			d.teleport(g, 0, -110)
			g.enterWater()
			d.moveDir, d.moveOn = v3(0, 0, -1), true
			d.next()
		}
	case 73:
		if d.stepT > 2 {
			d.check(g.surfing() && p.Pos.Y > WaterY, "Surferul: pe apa sta pe placa (y=%.2f)", p.Pos.Y)
			g.Draw()
			rl.TakeScreenshot(d.OutDir + "/v6_surf.png")
			d.moveOn = false
			g.Swimming = false
			d.teleport(g, 0, 30)
			d.next()
		}
	case 74: // cheat = totul deblocat
		old := g.Difficulty
		g.Difficulty = "cheat"
		all := true
		for i := range Realms {
			all = all && g.realmOpen(i)
		}
		for i := range Outfits {
			all = all && g.ownOutfit(i)
		}
		for k := 0; k < NumGuns; k++ {
			all = all && g.gunUnlocked(k)
		}
		d.check(all && g.worldsOpen() == NumWorlds, "cheat: toate lumile, insulele, costumele si armele sunt deschise")
		g.Difficulty = old
		d.next()
	case 75: // salvare cu mai multe lumi
		g.SaveData.Players = []PlayerSave{CreatePlayer("Aventurierul")}
		g.PlayerIdx = 0
		UpdatePlayerData(&v.saved, g)
		st := g.RealmStates[1].Level
		un := g.UnlockedRealm
		g.RealmStates = nil
		g.UnlockedRealm = 0
		LoadPlayerIntoGame(&v.saved, g)
		d.check(curRealm == 0 && g.UnlockedRealm == un && g.RealmStates[1].Seen && g.RealmStates[1].Level == st, "salvare: lumile vizitate si cea deschisa se pastreaza (deschisa %d)", g.UnlockedRealm)
		fmt.Println("lumi:", len(g.RealmStates))
		d.next()
	default:
		return true
	}
	return false
}

// probeColliders: BATTLE_DEBUG_PROBE="x,z" scrie obstacolele din jurul punctului (depanare blocaje).
func probeColliders(g *Game, s string) {
	var x, z float32
	fmt.Sscanf(s, "%f,%f", &x, &z)
	w := g.City.W
	for i, c := range w.Circles {
		if sq(c.X-x)+sq(c.Z-z) < sq(c.R+6) {
			fmt.Printf("cerc %d: (%.1f,%.1f) r=%.1f top=%.1f\n", i, c.X, c.Z, c.R, c.Top)
		}
	}
	for i, b := range w.Boxes {
		if x > b.MinX-6 && x < b.MaxX+6 && z > b.MinZ-6 && z < b.MaxZ+6 {
			fmt.Printf("cutie %d: x %.1f..%.1f z %.1f..%.1f top=%.1f roof=%d\n", i, b.MinX, b.MaxX, b.MinZ, b.MaxZ, b.Top, b.Roof)
		}
	}
}
